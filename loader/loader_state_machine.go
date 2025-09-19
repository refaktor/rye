//go:build stm_loader

// loader_state_machine.go
package loader

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/refaktor/rye/env"
)

// State machine states
const (
	STATE_NORMAL = iota
	STATE_IN_WORD
	STATE_IN_NUMBER
	STATE_IN_DECIMAL
	STATE_IN_STRING
	STATE_IN_COMMENT
	STATE_IN_SETWORD
	STATE_IN_GETWORD
	STATE_IN_OPWORD
	STATE_IN_PIPEWORD
	STATE_IN_TAGWORD
	STATE_IN_GENWORD
	STATE_IN_KINDWORD
	STATE_IN_XWORD
	STATE_IN_EXWORD
	STATE_IN_FPATH
	STATE_IN_URI
	STATE_IN_CPATH
	STATE_IN_MODWORD
)

// StateMachineParser is a single-pass parser that builds Rye objects directly
type StateMachineParser struct {
	input     string
	pos       int
	ch        byte
	line      int
	col       int
	wordIndex *env.Idxs
	state     int

	// Current token being built
	tokenStart int
	tokenValue string

	// Stack for nested blocks
	blockStack [][]env.Object
	blockTypes []int
}

// NewStateMachineParser creates a new state machine parser
func NewStateMachineParser(input string, wordIndex *env.Idxs) *StateMachineParser {
	p := &StateMachineParser{
		input:      input,
		wordIndex:  wordIndex,
		line:       1,
		col:        0,
		state:      STATE_NORMAL,
		blockStack: make([][]env.Object, 0),
		blockTypes: make([]int, 0),
	}

	// Initialize by reading the first character
	if len(input) > 0 {
		p.ch = input[0]
		p.pos = 1
	}

	return p
}

// readChar reads the next character
func (p *StateMachineParser) readChar() {
	if p.pos >= len(p.input) {
		p.ch = 0
	} else {
		p.ch = p.input[p.pos]
	}
	p.pos++
	p.col++

	if p.ch == '\n' {
		p.line++
		p.col = 0
	}
}

// peekChar returns the next character without advancing
func (p *StateMachineParser) peekChar() byte {
	if p.pos >= len(p.input) {
		return 0
	}
	return p.input[p.pos]
}

// isWhitespace checks if a character is whitespace
func (p *StateMachineParser) isWhitespace(ch byte) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r'
}

// isLetter checks if a character is a letter
func (p *StateMachineParser) isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_' || ch == '^' || ch == '`'
}

// isDigit checks if a character is a digit
func (p *StateMachineParser) isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

// startToken marks the beginning of a token
func (p *StateMachineParser) startToken() {
	p.tokenStart = p.pos - 1
	p.tokenValue = ""
}

// finishToken completes the current token and returns its value
func (p *StateMachineParser) finishToken() string {
	return p.input[p.tokenStart : p.pos-1]
}

// readUntilWhitespace reads characters until whitespace or a delimiter is encountered
func (p *StateMachineParser) readUntilWhitespace() string {
	start := p.pos - 1
	for p.ch != 0 && !p.isWhitespace(p.ch) && p.ch != '{' && p.ch != '}' && p.ch != '[' && p.ch != ']' && p.ch != '(' && p.ch != ')' {
		p.readChar()
	}
	return p.input[start : p.pos-1]
}

// pushBlock starts a new block
func (p *StateMachineParser) pushBlock(blockType int) {
	p.blockStack = append(p.blockStack, make([]env.Object, 0))
	p.blockTypes = append(p.blockTypes, blockType)
}

// popBlock finishes the current block and returns it
func (p *StateMachineParser) popBlock() (env.Object, error) {
	if len(p.blockStack) == 0 {
		return nil, fmt.Errorf("unmatched block end")
	}

	objects := p.blockStack[len(p.blockStack)-1]
	blockType := p.blockTypes[len(p.blockTypes)-1]

	p.blockStack = p.blockStack[:len(p.blockStack)-1]
	p.blockTypes = p.blockTypes[:len(p.blockTypes)-1]

	ser := env.NewTSeries(objects)
	return *env.NewBlock2(*ser, blockType), nil
}

// addToCurrentBlock adds an object to the current block
func (p *StateMachineParser) addToCurrentBlock(obj env.Object) {
	if len(p.blockStack) > 0 {
		p.blockStack[len(p.blockStack)-1] = append(p.blockStack[len(p.blockStack)-1], obj)
	}
}

// Parse parses the input and returns a Rye block
func (p *StateMachineParser) Parse() (env.Object, error) {
	// Start with an outer block
	p.pushBlock(0)

	wordBuffer := ""
	stringDelimiter := byte(0)

	for p.ch != 0 {
		switch p.state {
		case STATE_NORMAL:
			if p.isWhitespace(p.ch) {
				p.readChar()
				continue
			}

			switch p.ch {
			case '{':
				p.pushBlock(0)
				p.readChar()
			case '}':
				block, err := p.popBlock()
				if err != nil {
					return nil, err
				}
				if len(p.blockStack) > 0 {
					p.addToCurrentBlock(block)
				} else {
					// This is the outermost block
					return block, nil
				}
				p.readChar()
			case '[':
				p.pushBlock(1)
				p.readChar()
			case ']':
				block, err := p.popBlock()
				if err != nil {
					return nil, err
				}
				p.addToCurrentBlock(block)
				p.readChar()
			case '(':
				p.pushBlock(2)
				p.readChar()
			case ')':
				block, err := p.popBlock()
				if err != nil {
					return nil, err
				}
				p.addToCurrentBlock(block)
				p.readChar()
			case ',':
				p.addToCurrentBlock(env.Comma{})
				p.readChar()
			case '_':
				p.addToCurrentBlock(env.Void{})
				p.readChar()
			case '"', '`':
				stringDelimiter = p.ch
				p.state = STATE_IN_STRING
				p.startToken()
				p.readChar() // Skip opening quote
			case ':':
				if p.peekChar() == ':' {
					// Handle left mod-word (::word)
					p.state = STATE_IN_MODWORD
					p.startToken()
					p.readChar() // Skip first colon
					p.readChar() // Skip second colon
				} else {
					// Handle left set-word (:word)
					p.state = STATE_IN_SETWORD
					p.startToken()
					p.readChar() // Skip colon
				}
			case '?':
				p.state = STATE_IN_GETWORD
				p.startToken()
				p.readChar() // Skip question mark
			case '.':
				p.state = STATE_IN_OPWORD
				p.startToken()
				p.readChar() // Skip dot
			case '\\':
				p.state = STATE_IN_PIPEWORD
				p.startToken()
				p.readChar() // Skip backslash
			case '|':
				p.state = STATE_IN_PIPEWORD
				p.startToken()
				p.readChar() // Skip pipe
			case '\'':
				p.state = STATE_IN_TAGWORD
				p.startToken()
				p.readChar() // Skip single quote
			case '~':
				if p.peekChar() == '(' {
					p.state = STATE_IN_KINDWORD
					p.startToken()
					p.readChar() // Skip tilde
					p.readChar() // Skip opening parenthesis
				} else {
					p.state = STATE_IN_GENWORD
					p.startToken()
					p.readChar() // Skip tilde
				}
			case '<':
				if p.peekChar() == '/' {
					p.state = STATE_IN_EXWORD
					p.startToken()
					p.readChar() // Skip opening angle bracket
					p.readChar() // Skip slash
				} else {
					p.state = STATE_IN_XWORD
					p.startToken()
					p.readChar() // Skip opening angle bracket
				}
			case '%':
				p.state = STATE_IN_FPATH
				p.startToken()
				p.readChar() // Skip percent sign
			case ';':
				p.state = STATE_IN_COMMENT
				p.readChar() // Skip semicolon
			case '-':
				if p.isDigit(p.peekChar()) {
					p.state = STATE_IN_NUMBER
					p.startToken()
					p.readChar() // Include the minus sign
				} else {
					// Handle as an op-word
					idx := p.wordIndex.IndexWord("_-")
					p.addToCurrentBlock(*env.NewOpword(idx, 0))
					p.readChar()
				}
			default:
				if p.isDigit(p.ch) {
					p.state = STATE_IN_NUMBER
					p.startToken()
				} else if p.isLetter(p.ch) {
					p.state = STATE_IN_WORD
					p.startToken()
				} else {
					// Skip unknown character
					p.readChar()
				}
			}

		case STATE_IN_WORD:
			if p.isLetter(p.ch) || p.isDigit(p.ch) || p.ch == '-' || p.ch == '?' || p.ch == '=' || p.ch == '.' {
				p.readChar()
			} else if p.ch == ':' {
				if p.peekChar() == '/' && p.pos+1 < len(p.input) && p.input[p.pos+1] == '/' {
					// Handle URI (scheme://path)
					p.state = STATE_IN_URI
					p.readChar() // Skip colon
					p.readChar() // Skip first slash
					p.readChar() // Skip second slash
				} else if p.peekChar() == ':' {
					// Handle mod-word (word::)
					word := p.finishToken()
					p.readChar() // Skip colon
					p.readChar() // Skip second colon
					idx := p.wordIndex.IndexWord(word)
					p.addToCurrentBlock(*env.NewModword(idx))
					p.state = STATE_NORMAL
				} else {
					// Handle set-word (word:)
					word := p.finishToken()
					p.readChar() // Skip colon
					idx := p.wordIndex.IndexWord(word)
					p.addToCurrentBlock(*env.NewSetword(idx))
					p.state = STATE_NORMAL
				}
			} else if p.ch == '/' {
				// Handle context path (word/word)
				p.state = STATE_IN_CPATH
				p.readChar() // Skip slash
			} else if p.ch == '@' {
				// Handle email (user@domain)
				word := p.finishToken()
				p.readChar() // Skip @
				p.addToCurrentBlock(*env.NewEmail(word + "@" + p.readUntilWhitespace()))
				p.state = STATE_NORMAL
			} else {
				// Regular word
				word := p.finishToken()
				idx := p.wordIndex.IndexWord(word)
				p.addToCurrentBlock(*env.NewWord(idx))
				p.state = STATE_NORMAL
			}

		case STATE_IN_NUMBER:
			if p.isDigit(p.ch) {
				p.readChar()
			} else if p.ch == '.' && p.isDigit(p.peekChar()) {
				p.state = STATE_IN_DECIMAL
				p.readChar() // Skip decimal point
			} else {
				// Complete the number
				numStr := p.finishToken()
				val, err := strconv.ParseInt(numStr, 10, 64)
				if err != nil {
					return nil, fmt.Errorf("invalid number format: %s", err.Error())
				}
				p.addToCurrentBlock(*env.NewInteger(val))
				p.state = STATE_NORMAL
			}

		case STATE_IN_DECIMAL:
			if p.isDigit(p.ch) {
				p.readChar()
			} else {
				// Complete the decimal
				decStr := p.finishToken()
				val, err := strconv.ParseFloat(decStr, 64)
				if err != nil {
					return nil, fmt.Errorf("invalid decimal format: %s", err.Error())
				}
				p.addToCurrentBlock(*env.NewDecimal(val))
				p.state = STATE_NORMAL
			}

		case STATE_IN_STRING:
			if p.ch == stringDelimiter {
				// Complete the string
				str := p.input[p.tokenStart+1 : p.pos-1]
				// Process escape sequences
				str = strings.Replace(str, "\\n", "\n", -1)
				str = strings.Replace(str, "\\r", "\r", -1)
				str = strings.Replace(str, "\\t", "\t", -1)
				str = strings.Replace(str, "\\\"", "\"", -1)
				str = strings.Replace(str, "\\\\", "\\", -1)
				p.addToCurrentBlock(*env.NewString(str))
				p.readChar() // Skip closing quote
				p.state = STATE_NORMAL
			} else if p.ch == '\\' && p.peekChar() == stringDelimiter {
				p.readChar() // Skip backslash
				p.readChar() // Include the escaped quote
			} else {
				p.readChar()
			}

		case STATE_IN_COMMENT:
			if p.ch == '\n' || p.ch == 0 {
				p.state = STATE_NORMAL
			}
			p.readChar()

		case STATE_IN_SETWORD:
			if p.isLetter(p.ch) || p.isDigit(p.ch) || p.ch == '-' || p.ch == '?' || p.ch == '=' || p.ch == '.' {
				p.readChar()
			} else {
				// Complete the left set-word
				word := p.finishToken()
				// Remove the leading colon for left set-words
				idx := p.wordIndex.IndexWord(word[1:])
				p.addToCurrentBlock(*env.NewLSetword(idx))
				p.state = STATE_NORMAL
			}

		case STATE_IN_MODWORD:
			if p.isLetter(p.ch) || p.isDigit(p.ch) || p.ch == '-' || p.ch == '?' || p.ch == '=' || p.ch == '.' {
				p.readChar()
			} else {
				// Complete the left mod-word
				word := p.finishToken()
				// Remove the leading double colon for left mod-words
				idx := p.wordIndex.IndexWord(word[2:])
				p.addToCurrentBlock(*env.NewLModword(idx))
				p.state = STATE_NORMAL
			}

		case STATE_IN_GETWORD:
			if p.isLetter(p.ch) || p.isDigit(p.ch) || p.ch == '-' || p.ch == '?' || p.ch == '=' || p.ch == '.' {
				p.readChar()
			} else {
				// Complete the get-word
				word := p.finishToken()
				// Remove the leading question mark for get-words
				idx := p.wordIndex.IndexWord(word[1:])
				p.addToCurrentBlock(*env.NewGetword(idx))
				p.state = STATE_NORMAL
			}

		case STATE_IN_OPWORD:
			if p.isLetter(p.ch) || p.isDigit(p.ch) || p.ch == '-' || p.ch == '?' || p.ch == '=' || p.ch == '.' {
				p.readChar()
			} else if p.ch == '/' {
				// Handle op context path (.word/word)
				p.state = STATE_IN_CPATH
				wordBuffer = "." + p.finishToken()
				p.readChar() // Skip slash
			} else {
				// Complete the op-word
				word := p.finishToken()
				var idx int
				if len(word) == 0 || word == "<" || word == "-" || word == "<~" || word == ">=" || word == "<=" || word == "//" || word == ".." || word == "++" || word == "." || word == "|" {
					idx = p.wordIndex.IndexWord("_" + word)
				} else {
					// Remove the leading dot for op-words
					idx = p.wordIndex.IndexWord(word[1:])
				}
				p.addToCurrentBlock(*env.NewOpword(idx, 0))
				p.state = STATE_NORMAL
			}

		case STATE_IN_PIPEWORD:
			if p.isLetter(p.ch) || p.isDigit(p.ch) || p.ch == '-' || p.ch == '?' || p.ch == '=' || p.ch == '.' {
				p.readChar()
			} else if p.ch == '/' {
				// Handle pipe context path (\word/word)
				p.state = STATE_IN_CPATH
				wordBuffer = "\\" + p.finishToken()
				p.readChar() // Skip slash
			} else {
				// Complete the pipe-word
				word := p.finishToken()
				var idx int
				if word == ">>" || word == "->" || word == "~>" || word == "-->" || word == ".." || word == "|" {
					idx = p.wordIndex.IndexWord("_" + word)
				} else if p.input[p.tokenStart] == '|' && p.tokenStart > 0 && p.input[p.tokenStart-1] == '_' {
					// Special case for _|one-char-pipe
					idx = p.wordIndex.IndexWord("_" + word)
				} else {
					// Remove the leading backslash or pipe for pipe-words
					idx = p.wordIndex.IndexWord(word[1:])
				}
				p.addToCurrentBlock(*env.NewPipeword(idx, 0))
				p.state = STATE_NORMAL
			}

		case STATE_IN_TAGWORD:
			if p.isLetter(p.ch) || p.isDigit(p.ch) || p.ch == '-' || p.ch == '?' || p.ch == '=' || p.ch == '.' {
				p.readChar()
			} else {
				// Complete the tag-word
				word := p.finishToken()
				// Remove the leading single quote for tag-words
				idx := p.wordIndex.IndexWord(word[1:])
				p.addToCurrentBlock(*env.NewTagword(idx))
				p.state = STATE_NORMAL
			}

		case STATE_IN_GENWORD:
			if p.isLetter(p.ch) || p.isDigit(p.ch) || p.ch == '-' || p.ch == '?' || p.ch == '=' || p.ch == '.' {
				p.readChar()
			} else {
				// Complete the gen-word
				word := p.finishToken()
				idx := p.wordIndex.IndexWord(strings.ToLower(word))
				p.addToCurrentBlock(*env.NewGenword(idx))
				p.state = STATE_NORMAL
			}

		case STATE_IN_KINDWORD:
			if p.ch == ')' {
				// Complete the kind-word
				word := p.input[p.tokenStart+2 : p.pos-1] // Skip ~( and )
				idx := p.wordIndex.IndexWord(word)
				p.readChar() // Skip closing parenthesis

				// Optional trailing tilde
				if p.ch == '~' {
					p.readChar()
				}

				p.addToCurrentBlock(*env.NewKindword(idx))
				p.state = STATE_NORMAL
			} else {
				p.readChar()
			}

		case STATE_IN_XWORD:
			if p.ch == '>' {
				// Complete the x-word
				content := p.input[p.tokenStart+1 : p.pos-1] // Skip < and >
				parts := strings.SplitN(content, " ", 2)
				idx := p.wordIndex.IndexWord(parts[0])
				args := ""
				if len(parts) > 1 {
					args = parts[1]
				}
				p.readChar() // Skip closing angle bracket
				p.addToCurrentBlock(*env.NewXword(idx, args))
				p.state = STATE_NORMAL
			} else {
				p.readChar()
			}

		case STATE_IN_EXWORD:
			if p.ch == '>' {
				// Complete the ex-word
				word := p.input[p.tokenStart+2 : p.pos-1] // Skip </ and >
				idx := p.wordIndex.IndexWord(word)
				p.readChar() // Skip closing angle bracket
				p.addToCurrentBlock(*env.NewEXword(idx))
				p.state = STATE_NORMAL
			} else {
				p.readChar()
			}

		case STATE_IN_FPATH:
			if !p.isWhitespace(p.ch) && p.ch != '{' && p.ch != '}' && p.ch != '[' && p.ch != ']' && p.ch != 0 {
				p.readChar()
			} else {
				// Complete the file path
				path := p.finishToken()
				idx := p.wordIndex.IndexWord("file")
				p.addToCurrentBlock(*env.NewUri(p.wordIndex, *env.NewWord(idx), "file://"+path))
				p.state = STATE_NORMAL
			}

		case STATE_IN_URI:
			if !p.isWhitespace(p.ch) && p.ch != '{' && p.ch != '}' && p.ch != '[' && p.ch != ']' && p.ch != 0 {
				p.readChar()
			} else {
				// Complete the URI
				uri := p.finishToken()
				parts := strings.SplitN(uri, "://", 2)
				idx := p.wordIndex.IndexWord(parts[0])
				p.addToCurrentBlock(*env.NewUri(p.wordIndex, *env.NewWord(idx), uri))
				p.state = STATE_NORMAL
			}

		case STATE_IN_CPATH:
			if p.isLetter(p.ch) || p.isDigit(p.ch) || p.ch == '/' {
				p.readChar()
			} else {
				// Complete the context path
				path := p.finishToken()
				if wordBuffer != "" {
					// Handle op or pipe context path
					if strings.HasPrefix(wordBuffer, ".") {
						// Op context path
						parts := strings.Split(wordBuffer[1:]+"/"+path, "/")
						if len(parts) == 2 {
							idx1 := p.wordIndex.IndexWord(parts[0])
							idx2 := p.wordIndex.IndexWord(parts[1])
							p.addToCurrentBlock(*env.NewCPath2(1, *env.NewWord(idx1), *env.NewWord(idx2)))
						} else if len(parts) >= 3 {
							idx1 := p.wordIndex.IndexWord(parts[0])
							idx2 := p.wordIndex.IndexWord(parts[1])
							idx3 := p.wordIndex.IndexWord(parts[2])
							p.addToCurrentBlock(*env.NewCPath3(1, *env.NewWord(idx1), *env.NewWord(idx2), *env.NewWord(idx3)))
						}
					} else if strings.HasPrefix(wordBuffer, "\\") {
						// Pipe context path
						parts := strings.Split(wordBuffer[1:]+"/"+path, "/")
						if len(parts) == 2 {
							idx1 := p.wordIndex.IndexWord(parts[0])
							idx2 := p.wordIndex.IndexWord(parts[1])
							p.addToCurrentBlock(*env.NewCPath2(2, *env.NewWord(idx1), *env.NewWord(idx2)))
						} else if len(parts) >= 3 {
							idx1 := p.wordIndex.IndexWord(parts[0])
							idx2 := p.wordIndex.IndexWord(parts[1])
							idx3 := p.wordIndex.IndexWord(parts[2])
							p.addToCurrentBlock(*env.NewCPath3(2, *env.NewWord(idx1), *env.NewWord(idx2), *env.NewWord(idx3)))
						}
					}
					wordBuffer = ""
				} else {
					// Regular context path
					parts := strings.Split(path, "/")
					if len(parts) == 2 {
						idx1 := p.wordIndex.IndexWord(parts[0])
						idx2 := p.wordIndex.IndexWord(parts[1])
						p.addToCurrentBlock(*env.NewCPath2(0, *env.NewWord(idx1), *env.NewWord(idx2)))
					} else if len(parts) >= 3 {
						idx1 := p.wordIndex.IndexWord(parts[0])
						idx2 := p.wordIndex.IndexWord(parts[1])
						idx3 := p.wordIndex.IndexWord(parts[2])
						p.addToCurrentBlock(*env.NewCPath3(0, *env.NewWord(idx1), *env.NewWord(idx2), *env.NewWord(idx3)))
					}
				}
				p.state = STATE_NORMAL
			}
		}
	}

	// Handle any unclosed blocks
	if len(p.blockStack) > 1 {
		return nil, fmt.Errorf("unclosed block")
	}

	// Return the outermost block
	if len(p.blockStack) == 1 {
		objects := p.blockStack[0]
		// If the first object is a block and it's the only object, return it directly
		if len(objects) == 1 {
			if block, ok := objects[0].(env.Block); ok {
				return block, nil
			}
		}
		ser := env.NewTSeries(objects)
		return *env.NewBlock(*ser), nil
	}

	return nil, fmt.Errorf("no blocks parsed")
}

// LoadStringStateMachine loads a string using the state machine parser
func LoadStringStateMachine(input string, sig bool) (env.Object, *env.Idxs) {
	InitIndex()
	wordIndexMutex.Lock()
	defer wordIndexMutex.Unlock()

	if sig {
		signed := checkCodeSignature(input)
		if signed == -1 {
			return *env.NewError("Signature not found"), wordIndex
		} else if signed == -2 {
			return *env.NewError("Invalid signature"), wordIndex
		}
	}

	input = removeBangLine(input)

	inp1 := strings.TrimSpace(input)
	if len(inp1) == 0 || strings.Index(inp1, "{") != 0 {
		input = "{ " + input + " }"
	}

	parser := NewStateMachineParser(input, wordIndex)
	if parser == nil {
		return *env.NewError("Failed to create parser"), wordIndex
	}

	val, err := parser.Parse()
	if err != nil {
		return *env.NewError(fmt.Sprintf("Parse error: %s", err.Error())), wordIndex
	}

	return val, wordIndex
}

// LoadStringNEWStateMachine loads a string using the state machine parser with a program state
func LoadStringNEWStateMachine(input string, sig bool, ps *env.ProgramState) env.Object {
	if sig {
		signed := checkCodeSignature(input)
		if signed == -1 {
			return *env.NewError("Signature not found")
		} else if signed == -2 {
			return *env.NewError("Invalid signature")
		}
	}

	input = removeBangLine(input)

	inp1 := strings.TrimSpace(input)
	if len(inp1) == 0 || strings.Index(inp1, "{") != 0 {
		input = "{ " + input + " }"
	}

	wordIndexMutex.Lock()
	wordIndex = ps.Idx
	parser := NewStateMachineParser(input, wordIndex)
	if parser == nil {
		wordIndexMutex.Unlock()
		return *env.NewError("Failed to create parser")
	}

	val, err := parser.Parse()
	ps.Idx = wordIndex
	wordIndexMutex.Unlock()

	if err != nil {
		ps.FailureFlag = true
		return *env.NewError(fmt.Sprintf("Parse error: %s", err.Error()))
	}

	return val
}
