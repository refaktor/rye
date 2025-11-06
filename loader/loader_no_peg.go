//go:build !pico_loader

// loader_no_peg.go
package loader

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/refaktor/rye/env"
	"github.com/refaktor/rye/util"
)

// Token types for the non-PEG parser
const (
	NPEG_TOKEN_NONE = iota
	NPEG_TOKEN_WORD
	NPEG_TOKEN_SETWORD
	NPEG_TOKEN_LSETWORD
	NPEG_TOKEN_MODWORD
	NPEG_TOKEN_LMODWORD
	NPEG_TOKEN_GETWORD
	NPEG_TOKEN_OPWORD
	NPEG_TOKEN_PIPEWORD
	NPEG_TOKEN_ONECHARPIPE
	NPEG_TOKEN_TAGWORD
	NPEG_TOKEN_KINDWORD
	NPEG_TOKEN_XWORD
	NPEG_TOKEN_EXWORD
	NPEG_TOKEN_GENWORD
	NPEG_TOKEN_NUMBER
	NPEG_TOKEN_DECIMAL
	NPEG_TOKEN_STRING
	NPEG_TOKEN_URI
	NPEG_TOKEN_EMAIL
	NPEG_TOKEN_FPATH
	NPEG_TOKEN_CPATH
	NPEG_TOKEN_OPCPATH
	NPEG_TOKEN_PIPECPATH
	NPEG_TOKEN_GETCPATH
	NPEG_TOKEN_BLOCK_START
	NPEG_TOKEN_BLOCK_END
	NPEG_TOKEN_BBLOCK_START
	NPEG_TOKEN_BBLOCK_END
	NPEG_TOKEN_OPBBLOCK_START
	NPEG_TOKEN_GROUP_START
	NPEG_TOKEN_GROUP_END
	NPEG_TOKEN_OPGROUP_START
	NPEG_TOKEN_OPBLOCK_START
	NPEG_TOKEN_COMMA
	NPEG_TOKEN_VOID
	NPEG_TOKEN_COMMENT
	NPEG_TOKEN_SPACE
	NPEG_TOKEN_LOCATION_NODE
	NPEG_TOKEN_FLAGWORD
	NPEG_TOKEN_EOF
	NPEG_TOKEN_ERROR
)

const (
	ERR_NONE = iota
	ERR_UNKNOWN
	ERR_SPACING_OP
	ERR_SPACING_BLK
	ERR_SPACING_OTHR
)

// NoPEGToken represents a lexical token
type NoPEGToken struct {
	Type  int
	Value string
	Line  int
	Col   int
	Err   int
}

// Lexer tokenizes input string character by character
type Lexer struct {
	input      string
	pos        int
	readPos    int
	ch         byte
	line       int
	col        int
	startLine  int
	startCol   int
	tokenStart int
	// For LocationNode injection
	injectLocationNodes bool
	scriptPath          string
	lastNewlinePos      int
	pendingLocationNode *NoPEGToken
}

// NewLexer creates a new lexer
func NewLexer(input string) *Lexer {
	l := &Lexer{
		input: input,
		line:  1,
		col:   0,
	}
	// Initialize by reading the first character
	if len(input) > 0 {
		l.ch = input[0]
		l.readPos = 1
	}
	return l
}

// NewLexerWithLocationNodes creates a new lexer that injects LocationNodes at newlines
func NewLexerWithLocationNodes(input string, scriptPath string) *Lexer {
	l := &Lexer{
		input:               input,
		line:                1,
		col:                 0,
		injectLocationNodes: true,
		scriptPath:          scriptPath,
		lastNewlinePos:      0,
	}
	// Initialize by reading the first character
	if len(input) > 0 {
		l.ch = input[0]
		l.readPos = 1
	}
	return l
}

// NoPEGParser parses tokens into Rye values
type NoPEGParser struct {
	l            *Lexer
	currentToken NoPEGToken
	peekToken    NoPEGToken
	errors       []string
	wordIndex    *env.Idxs
}

// readChar reads the next character
func (l *Lexer) readChar() {
	if l.readPos >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPos]
	}
	l.pos = l.readPos
	l.readPos++
	l.col++

	if l.ch == '\n' {
		l.line++
		l.col = 0
	}
}

// peekChar returns the next character without advancing
func (l *Lexer) peekChar() byte {
	if l.readPos >= len(l.input) {
		return 0
	}
	return l.input[l.readPos]
}

func (l *Lexer) peekCharOffs(offs int) byte {
	if l.readPos >= len(l.input) {
		return 0
	}
	return l.input[l.readPos+offs]
}

// skipWhitespace skips whitespace characters
func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		// If we're injecting location nodes and we hit a newline, prepare a location node
		if l.injectLocationNodes && l.ch == '\n' && l.pendingLocationNode == nil {
			// Extract the source line that just ended
			sourceLine := l.extractSourceLine(l.lastNewlinePos, l.pos)
			l.pendingLocationNode = &NoPEGToken{
				Type:  NPEG_TOKEN_LOCATION_NODE,
				Value: sourceLine,
				Line:  l.line,
				Col:   1, // Location nodes are placed at the beginning of lines
			}
			l.lastNewlinePos = l.pos + 1 // Update position for next line
		}
		l.readChar()
	}
}

// makeToken creates a token with the current position information
func (l *Lexer) makeToken(tokenType int, value string) NoPEGToken {
	return NoPEGToken{
		Type:  tokenType,
		Value: value,
		Line:  l.startLine,
		Col:   l.startCol,
	}
}

func (l *Lexer) makeTokenErr(tokenType int, value string, err int) NoPEGToken {
	return NoPEGToken{
		Type:  tokenType,
		Value: value,
		Line:  l.startLine,
		Col:   l.startCol,
		Err:   err,
	}
}

// startToken marks the beginning of a token
func (l *Lexer) startToken() {
	l.startLine = l.line
	l.startCol = l.col
	l.tokenStart = l.pos
}

// extractSourceLine extracts the source line between two positions
func (l *Lexer) extractSourceLine(start, end int) string {
	if start < 0 {
		start = 0
	}
	if end > len(l.input) {
		end = len(l.input)
	}
	if start >= end {
		return ""
	}
	line := l.input[start:end]
	// Remove the newline character if present
	if len(line) > 0 && (line[len(line)-1] == '\n' || line[len(line)-1] == '\r') {
		line = line[:len(line)-1]
	}
	return line
}

// NextToken returns the next token from the input
func (l *Lexer) NextToken() NoPEGToken {

	// fmt.Println("NEXT TOKEN ENTER -->")
	// fmt.Println(l.pos)

	// Check if we have a pending LocationNode to return
	if l.pendingLocationNode != nil {
		token := *l.pendingLocationNode
		l.pendingLocationNode = nil
		return token
	}

	l.skipWhitespace()

	if l.ch == 0 {
		// Before EOF, check if we need to inject a final LocationNode for the last line
		if l.injectLocationNodes && l.lastNewlinePos < len(l.input) {
			sourceLine := l.extractSourceLine(l.lastNewlinePos, len(l.input))
			if sourceLine != "" {
				l.lastNewlinePos = len(l.input) // Prevent duplicate
				return l.makeToken(NPEG_TOKEN_LOCATION_NODE, sourceLine)
			}
		}
		return l.makeToken(NPEG_TOKEN_EOF, "")
	}

	l.startToken()

	switch l.ch {
	case '{':
		return l.readOneCharToken(NPEG_TOKEN_BLOCK_START, ERR_SPACING_BLK)
	case '}':
		return l.readOneCharToken(NPEG_TOKEN_BLOCK_END, ERR_SPACING_BLK)
	case '[':
		return l.readOneCharToken(NPEG_TOKEN_BBLOCK_START, ERR_SPACING_BLK)
	case ']':
		return l.readOneCharToken(NPEG_TOKEN_BBLOCK_END, ERR_SPACING_BLK)
	case '(':
		return l.readOneCharToken(NPEG_TOKEN_GROUP_START, ERR_SPACING_BLK)
	case ')':
		return l.readOneCharToken(NPEG_TOKEN_GROUP_END, ERR_SPACING_BLK)
	case ',':
		return l.readOneCharToken(NPEG_TOKEN_COMMA, ERR_SPACING_OTHR)
	case '_':
		if isWhitespaceCh(l.peekChar()) {
			l.readChar()
			return l.makeToken(NPEG_TOKEN_VOID, "_")
		}
		return l.readWord()
	case '"', '`':
		return l.readString()
	case ':':
		nch := l.peekChar()
		if nch == ':' {
			l.readChar()
			return l.readLModWord()
		}
		return l.readLSetWord()
	case '?':
		return l.readGetWord()
	case '.', '+', '*', '/', '>', '=': // taken for other tokens also - <
		// Special handling for ".[ ]" (OPBBLOCK) pattern
		if l.ch == '.' && l.peekChar() == '[' {
			l.readChar() // Skip '.'
			l.readChar() // Skip '['
			if isWhitespaceCh(l.ch) {
				return l.makeToken(NPEG_TOKEN_OPBBLOCK_START, ".[")
			}
			// If not followed by whitespace, reset and treat as regular opword
			l.pos -= 2
			l.readPos -= 2
			l.col -= 2
			l.ch = '.'
		}
		// Special handling for ".( )" (OPGROUP) pattern
		if l.ch == '.' && l.peekChar() == '(' {
			l.readChar() // Skip '.'
			l.readChar() // Skip '('
			if isWhitespaceCh(l.ch) {
				return l.makeToken(NPEG_TOKEN_OPGROUP_START, ".(")
			}
			// If not followed by whitespace, reset and treat as regular opword
			l.pos -= 2
			l.readPos -= 2
			l.col -= 2
			l.ch = '.'
		}
		// Special handling for ".{ }" (OPBLOCK) pattern
		if l.ch == '.' && l.peekChar() == '{' {
			l.readChar() // Skip '.'
			l.readChar() // Skip '{'
			if isWhitespaceCh(l.ch) {
				return l.makeToken(NPEG_TOKEN_OPBLOCK_START, ".{")
			}
			// If not followed by whitespace, reset and treat as regular opword
			l.pos -= 2
			l.readPos -= 2
			l.col -= 2
			l.ch = '.'
		}
		// Special handling for "//" operator
		if l.ch == '/' && l.peekChar() == '/' {
			l.readChar() // Skip first '/'
			l.readChar() // Skip second '/'
			if isWhitespaceCh(l.ch) {
				return l.makeToken(NPEG_TOKEN_OPWORD, "//")
			}
			// If not followed by whitespace, continue with normal opword parsing
			// Reset position to handle as regular opword
			l.pos -= 2
			l.readPos -= 2
			l.col -= 2
			l.ch = '/'
		}
		return l.readOpWord()
	case '|':
		return l.readPipeWord()
	case '\'':
		return l.readTagWord()
	/*case '~':
	if l.peekChar() == '(' {
		return l.readKindWord()
	} else {
		return l.readGenWord()
	}*/
	case '<':
		pch := l.peekChar()
		if isWhitespace(pch) || pch == '-' || pch == '~' || pch == '=' || pch == '<' || pch == '>' {
			return l.readOpWord()
			// l.readChar()
			// return l.makeToken(NPEG_TOKEN_OPWORD, "<")
		}
		return l.readXWord()
	case '%':
		if isWhitespace(l.peekChar()) {
			l.readChar()
			return l.makeToken(NPEG_TOKEN_OPWORD, "%")
		}
		return l.readFPath()
	case ';':
		return l.readComment()
	case '~':
		pch := l.peekChar()
		// fmt.Println("***1")
		if isWhitespace(pch) {
			l.readChar()
			// fmt.Println("***2")
			return l.makeToken(NPEG_TOKEN_OPWORD, "~")
		} else {
			return l.readPipeWord()
		}
	//if l.peekChar() == '(' {
	//	return l.readKindWord()
	// } else {
	// 	return l.readGenWord()
	// }*/
	case '-':
		if isDigit(l.peekChar()) {
			// fmt.Println("***0")
			return l.readNumber()
		} else {
			pch := l.peekChar()
			// fmt.Println("***1")
			if pch == '-' {
				// Could be a long flag (--verbose)
				return l.readFlagword()
			} else if isLetter(pch) {
				// Could be a short flag (-v)
				return l.readFlagword()
			} else if isWhitespace(pch) {
				l.readChar()
				// fmt.Println("***2")
				return l.makeToken(NPEG_TOKEN_OPWORD, "-")
			} else {
				// fmt.Println("***3")
				// return l.readOpWord()
				return l.readPipeWord()
			}
		}
	default:
		if isDigit(l.ch) {
			return l.readNumber()
		} else if isLetter(l.ch) {
			// Try to parse as word or special type
			word := l.readWord()

			// Check if it's a set-word (word:)
			if l.ch == ':' {
				if l.peekChar() == ':' { // word::
					if isWhitespace(l.peekCharOffs(1)) {
						l.readChar()
						l.readChar()
						return l.makeToken(NPEG_TOKEN_MODWORD, l.input[l.tokenStart:l.pos])
					}
				}
				if isWhitespace(l.peekChar()) {
					l.readChar()
					return l.makeToken(NPEG_TOKEN_SETWORD, l.input[l.tokenStart:l.pos])
				}
			}

			// Check if it's a URI (word://)
			if l.ch == ':' && l.peekChar() == '/' && l.peekChar() == '/' {
				l.readChar() // :
				l.readChar() // /
				l.readChar() // /

				// Read the rest of the URI
				for l.ch != 0 && !isWhitespaceCh(l.ch) && l.ch != '{' && l.ch != '}' && l.ch != '[' && l.ch != ']' {
					l.readChar()
				}

				return l.makeToken(NPEG_TOKEN_URI, l.input[l.tokenStart:l.pos])
			}

			if l.ch == '@' {
				// Read the rest of the URI
				for l.ch != 0 && !isWhitespaceCh(l.ch) && l.ch != '{' && l.ch != '}' && l.ch != '[' && l.ch != ']' {
					l.readChar()
				}

				return l.makeToken(NPEG_TOKEN_EMAIL, l.input[l.tokenStart:l.pos])

			}

			// Check if it's a context path (word/word)
			if l.ch == '/' {
				l.readChar()

				// Read the rest of the path
				for isWordCharacter(l.ch) {
					l.readChar()
				}

				return l.makeToken(NPEG_TOKEN_CPATH, l.input[l.tokenStart:l.pos])
			}

			return word
		}

		// Unknown character
		l.readChar()
		return l.makeToken(NPEG_TOKEN_NONE, string(l.input[l.pos-1]))
	}
}

// isLetter checks if a character is a letter
func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_' || ch == '^' || ch == '`'
}

// isDigit checks if a character is a digit
func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

// isWordCharacter checks if a character can be part of a word
func isWordCharacter(ch byte) bool {
	return isLetter(ch) || isDigit(ch) || ch == '-' || ch == '+' || ch == '.' ||
		ch == '!' || ch == '*' || ch == '>' || ch == '<' || ch == '\\' || ch == '?' || ch == '=' || ch == '_'
}

// isWhitespaceCh checks if a character is whitespace
func isWhitespaceCh(ch byte) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r'
}

// isTokenDelimiter checks if a character is a token delimiter (whitespace or end of input)
func isTokenDelimiter(ch byte) bool {
	return isWhitespaceCh(ch) || ch == 0
}

// isSpecialChar checks if a character is a special character that should be treated as a separate token
func isSpecialChar(ch byte) bool {
	return ch == '{' || ch == '}' || ch == '[' || ch == ']' || ch == '(' || ch == ')' ||
		ch == ',' || ch == '+' || ch == '-' || ch == '*' || ch == '/' || ch == '.' ||
		ch == '>' || ch == '<' || ch == '|' || ch == '\\' || ch == ':' || ch == '?' ||
		ch == '~' || ch == '%' || ch == ';' || ch == '@'
}

func (l *Lexer) readOneCharToken(tokenType int, errType int) NoPEGToken {
	l.readChar()
	// fmt.Println("**")
	c := l.ch
	// fmt.Println(c)
	if c != 0 && !isWhitespaceCh(c) {
		l.readChar()
		return l.makeTokenErr(NPEG_TOKEN_ERROR, "501", errType)
	}
	return l.makeToken(tokenType, "")
}

func (l *Lexer) readOneCharToken2(tokenType int, errType int) NoPEGToken {
	l.readChar()
	if !isWhitespace(l.peekChar()) {
		l.readChar()
		return l.makeTokenErr(NPEG_TOKEN_ERROR, "", errType)
	}
	return l.makeToken(tokenType, "{}")
}

// readWord reads a word token
func (l *Lexer) readWord() NoPEGToken {
	for isWordCharacter(l.ch) {
		l.readChar()
	}

	// Ensure the word is followed by a token delimiter or a valid word terminator
	if !isTokenDelimiter(l.ch) && l.ch != ':' && l.ch != '@' && l.ch != '/' {
		// If not, this is an invalid token
		// Continue reading until we hit a delimiter to report the full invalid token
		for !isTokenDelimiter(l.ch) {
			l.readChar()
		}
		return l.makeToken(NPEG_TOKEN_NONE, l.input[l.tokenStart:l.pos])
	}

	return l.makeToken(NPEG_TOKEN_WORD, l.input[l.tokenStart:l.pos])
}

// readString reads a string token
func (l *Lexer) readString() NoPEGToken {
	delimiter := l.ch
	l.readChar() // Skip opening quote

	for l.ch != 0 && l.ch != delimiter {
		// Handle escape sequences properly
		if l.ch == '\\' {
			// Skip the backslash
			l.readChar()
			// Skip the escaped character (if any)
			if l.ch != 0 {
				l.readChar()
			}
		} else {
			l.readChar()
		}
	}

	if l.ch == delimiter {
		l.readChar() // Skip closing quote
	}

	// Ensure the string is followed by a token delimiter
	if !isWhitespace(l.ch) {
		// fmt.Println("--NOSPACING INVOKED*->")
		// fmt.Println(l.pos)
		// fmt.Println(l.input[l.tokenStart:l.pos])
		return l.makeTokenErr(NPEG_TOKEN_ERROR, l.input[l.tokenStart:l.pos], determineLexerError(l.ch))
	}

	return l.makeToken(NPEG_TOKEN_STRING, l.input[l.tokenStart:l.pos])
}

// readNumber reads a number or decimal token
func (l *Lexer) readNumber() NoPEGToken {
	// Handle negative sign
	if l.ch == '-' {
		l.readChar()
	}

	// Read integer part
	for isDigit(l.ch) {
		l.readChar()
	}

	// fmt.Println(l.pos)
	// fmt.Println(l.col)

	// Check for decimal point
	if l.ch == '.' && isDigit(l.peekChar()) {
		l.readChar() // Skip decimal point

		// Read decimal part
		for isDigit(l.ch) {
			l.readChar()
		}

		// Ensure the decimal is followed by a token delimiter
		if !isWhitespace(l.ch) {
			// Check if there should be spacing between tokens
			// fmt.Println("--->")
			// fmt.Println(l.pos)
			// fmt.Println(l.input[l.tokenStart:l.pos])
			return l.makeTokenErr(NPEG_TOKEN_ERROR, l.input[l.tokenStart:l.pos], determineLexerError(l.ch))
		}

		return l.makeToken(NPEG_TOKEN_DECIMAL, l.input[l.tokenStart:l.pos])
	}

	// Ensure the number is followed by a token delimiter
	if !isWhitespace(l.ch) {
		// fmt.Println("--NOSPACING INVOKED*->")
		// fmt.Println(l.pos)
		// fmt.Println(l.input[l.tokenStart:l.pos])
		return l.makeTokenErr(NPEG_TOKEN_ERROR, l.input[l.tokenStart:l.pos], determineLexerError(l.ch))
	}

	return l.makeToken(NPEG_TOKEN_NUMBER, l.input[l.tokenStart:l.pos])
}

func determineLexerError(ch byte) int {
	switch ch {
	case '+', '-', '/', '*', '%':
		return ERR_SPACING_OP
	case '}', ']', ')', '{', '[', '(':
		return ERR_SPACING_BLK
	}
	return ERR_SPACING_OTHR
}

// readComment reads a comment token
func (l *Lexer) readComment() NoPEGToken {
	l.readChar() // Skip semicolon

	// Read until end of line or end of input
	for l.ch != 0 && l.ch != '\n' {
		l.readChar()
	}

	return l.makeToken(NPEG_TOKEN_COMMENT, l.input[l.tokenStart:l.pos])
}

// Placeholder implementations for other token types
func (l *Lexer) readLSetWord() NoPEGToken {
	l.readChar() // Skip colon

	// Read the word part
	for isWordCharacter(l.ch) {
		l.readChar()
	}

	// Ensure the token is followed by whitespace
	if !isWhitespace(l.ch) {
		return l.makeTokenErr(NPEG_TOKEN_ERROR, l.input[l.tokenStart:l.pos], determineLexerError(l.ch))
	}

	return l.makeToken(NPEG_TOKEN_LSETWORD, l.input[l.tokenStart:l.pos])
}

// Placeholder implementations for other token types
func (l *Lexer) readLModWord() NoPEGToken {
	l.readChar() // Skip colon

	// Read the word part
	for isWordCharacter(l.ch) {
		l.readChar()
	}

	// Ensure the token is followed by whitespace
	if !isWhitespace(l.ch) {
		return l.makeTokenErr(NPEG_TOKEN_ERROR, l.input[l.tokenStart:l.pos], determineLexerError(l.ch))
	}

	return l.makeToken(NPEG_TOKEN_LMODWORD, l.input[l.tokenStart:l.pos])
}

func (l *Lexer) readGetWord() NoPEGToken {
	l.readChar() // Skip question mark

	cpath := false
	slashPos := -1

	// Read the word part
	for isWordCharacter(l.ch) || l.ch == '/' {
		// Check if it's a context path (word/word)
		if l.ch == '/' {
			slashPos = l.pos
		}

		l.readChar()
	}

	// Only treat as context path if there's content after the slash
	if slashPos >= 0 && slashPos < l.pos-1 {
		cpath = true
	}

	// Ensure the token is followed by whitespace
	if !isWhitespace(l.ch) {
		return l.makeTokenErr(NPEG_TOKEN_ERROR, l.input[l.tokenStart:l.pos], determineLexerError(l.ch))
	}

	if cpath {
		return l.makeToken(NPEG_TOKEN_GETCPATH, l.input[l.tokenStart:l.pos])
	}

	return l.makeToken(NPEG_TOKEN_GETWORD, l.input[l.tokenStart:l.pos])
}

func (l *Lexer) readOpWord() NoPEGToken {
	l.readChar() // Skip first character

	cpath := false

	// Read the word part
	for isWordCharacter(l.ch) || l.ch == '/' || l.ch == '<' || l.ch == '~' {
		// Check if it's a context path (word/word)
		if l.ch == '/' {
			cpath = true
		}

		l.readChar()
	}

	// Ensure the token is followed by whitespace
	if !isWhitespace(l.ch) {
		return l.makeTokenErr(NPEG_TOKEN_ERROR, l.input[l.tokenStart:l.pos], determineLexerError(l.ch))
	}

	if cpath {
		return l.makeToken(NPEG_TOKEN_OPCPATH, l.input[l.tokenStart:l.pos])
	}

	return l.makeToken(NPEG_TOKEN_OPWORD, l.input[l.tokenStart:l.pos])
}

func (l *Lexer) readPipeWord() NoPEGToken {
	ch := l.ch
	l.readChar() // Skip backslash or pipe

	cpath := false

	// Read the word part
	for isWordCharacter(l.ch) || l.ch == '/' || l.ch == '<' {
		if l.ch == '/' {
			cpath = true
		}

		l.readChar()
	}

	// Ensure the token is followed by whitespace
	if !isWhitespace(l.ch) {
		return l.makeTokenErr(NPEG_TOKEN_ERROR, l.input[l.tokenStart:l.pos], determineLexerError(l.ch))
	}

	if cpath {
		return l.makeToken(NPEG_TOKEN_PIPECPATH, l.input[l.tokenStart:l.pos])
	}

	if ch == '|' {
		return l.makeToken(NPEG_TOKEN_PIPEWORD, l.input[l.tokenStart:l.pos])
	} else {
		return l.makeToken(NPEG_TOKEN_PIPEWORD, string(ch)+l.input[l.tokenStart:l.pos])
	}

}

func (l *Lexer) readTagWord() NoPEGToken {
	l.readChar() // Skip single quote

	// Read the word part
	for isWordCharacter(l.ch) {
		l.readChar()
	}

	// Ensure the token is followed by whitespace
	if !isWhitespace(l.ch) {
		return l.makeTokenErr(NPEG_TOKEN_ERROR, l.input[l.tokenStart:l.pos], determineLexerError(l.ch))
	}

	return l.makeToken(NPEG_TOKEN_TAGWORD, l.input[l.tokenStart:l.pos])
}

func (l *Lexer) readGenWord() NoPEGToken {
	l.readChar() // Skip tilde

	// Read the word part
	for isWordCharacter(l.ch) {
		l.readChar()
	}

	// Ensure the token is followed by whitespace
	if !isWhitespace(l.ch) {
		return l.makeTokenErr(NPEG_TOKEN_ERROR, l.input[l.tokenStart:l.pos], determineLexerError(l.ch))
	}

	return l.makeToken(NPEG_TOKEN_GENWORD, l.input[l.tokenStart:l.pos])
}

func (l *Lexer) readKindWord() NoPEGToken {
	l.readChar() // Skip tilde
	l.readChar() // Skip opening parenthesis

	// Read the word part
	for l.ch != 0 && l.ch != ')' {
		l.readChar()
	}

	if l.ch == ')' {
		l.readChar() // Skip closing parenthesis

		// Optional trailing tilde
		if l.ch == '~' {
			l.readChar()
		}
	}

	// Ensure the token is followed by whitespace
	if !isWhitespace(l.ch) {
		return l.makeTokenErr(NPEG_TOKEN_ERROR, l.input[l.tokenStart:l.pos], determineLexerError(l.ch))
	}

	return l.makeToken(NPEG_TOKEN_KINDWORD, l.input[l.tokenStart:l.pos])
}

func (l *Lexer) readXWord() NoPEGToken {
	l.readChar() // Skip opening angle bracket

	// Check if it's an EXWord (</word>)
	if l.ch == '/' {
		l.readChar() // Skip slash

		// Read the word part
		for l.ch != 0 && l.ch != '>' {
			l.readChar()
		}

		if l.ch == '>' {
			l.readChar() // Skip closing angle bracket
		}

		// Ensure the token is followed by whitespace
		if !isWhitespace(l.ch) {
			return l.makeTokenErr(NPEG_TOKEN_ERROR, l.input[l.tokenStart:l.pos], determineLexerError(l.ch))
		}

		return l.makeToken(NPEG_TOKEN_EXWORD, l.input[l.tokenStart:l.pos])
	}

	// Regular XWord
	// Read until closing angle bracket
	for l.ch != 0 && l.ch != '>' {
		l.readChar()
	}

	if l.ch == '>' {
		l.readChar() // Skip closing angle bracket
	}

	// Ensure the token is followed by whitespace
	if !isWhitespace(l.ch) {
		return l.makeTokenErr(NPEG_TOKEN_ERROR, l.input[l.tokenStart:l.pos], determineLexerError(l.ch))
	}

	return l.makeToken(NPEG_TOKEN_XWORD, l.input[l.tokenStart:l.pos])
}

func (l *Lexer) readFPath() NoPEGToken {
	l.readChar() // Skip percent sign

	// Read the path part
	for l.ch != 0 && !isWhitespaceCh(l.ch) && l.ch != '{' && l.ch != '}' && l.ch != '[' && l.ch != ']' {
		l.readChar()
	}

	// Ensure the token is followed by whitespace
	if !isWhitespace(l.ch) {
		return l.makeTokenErr(NPEG_TOKEN_ERROR, l.input[l.tokenStart:l.pos], determineLexerError(l.ch))
	}

	return l.makeToken(NPEG_TOKEN_FPATH, l.input[l.tokenStart:l.pos])
}

func (l *Lexer) readOnePipeWord() NoPEGToken {
	l.readChar() // Skip pipe character

	// Read the word part
	for isWordCharacter(l.ch) {
		l.readChar()
	}

	// Ensure the token is followed by whitespace
	if !isWhitespace(l.ch) {
		return l.makeTokenErr(NPEG_TOKEN_ERROR, l.input[l.tokenStart:l.pos], determineLexerError(l.ch))
	}

	// Extract the word part (without the pipe character)
	word := l.input[l.tokenStart+1 : l.pos]
	return l.makeToken(NPEG_TOKEN_ONECHARPIPE, word)
}

// readFlagword reads a flag token (-v, --verbose, or -v|verbose)
func (l *Lexer) readFlagword() NoPEGToken {
	// possible inputs:
	// -v
	// --verbose
	// -v|verbose

	l.readChar() // Skip the first '-'

	shortForm := true

	// if second charater is also '-'
	if l.ch == '-' {
		shortForm = false
		// This is a long flag (--word)
		l.readChar() // Skip the second '-'
	}

	// Read the flag name
	if !isLetter(l.ch) {
		// Invalid flag - should start with a letter
		return l.makeTokenErr(NPEG_TOKEN_ERROR, l.input[l.tokenStart:l.pos], ERR_SPACING_OTHR)
	}

	// Read the single letter
	l.readChar()

	// Check if there's a pipe for combined short|long format (and it started with one -)
	if shortForm && l.ch == '|' {
		shortForm = false
		l.readChar() // Skip the '|'
	}

	if !shortForm {

		// Read the rest of long form part
		for isWordCharacter(l.ch) {
			l.readChar()
		}

	}

	// Ensure the token is followed by whitespace
	if !isWhitespace(l.ch) {
		return l.makeTokenErr(NPEG_TOKEN_ERROR, l.input[l.tokenStart:l.pos], determineLexerError(l.ch))
	}

	return l.makeToken(NPEG_TOKEN_FLAGWORD, l.input[l.tokenStart:l.pos])
}

// NewParserNoPEG creates a new parser
func NewParserNoPEG(input string, wordIndex *env.Idxs) *NoPEGParser {
	l := NewLexer(input)
	p := &NoPEGParser{
		l:         l,
		wordIndex: wordIndex,
	}
	return p
}

// NewParserNoPEGWithLocationNodes creates a new parser that injects LocationNodes
func NewParserNoPEGWithLocationNodes(input string, wordIndex *env.Idxs, scriptPath string) *NoPEGParser {
	l := NewLexerWithLocationNodes(input, scriptPath)
	p := &NoPEGParser{
		l:         l,
		wordIndex: wordIndex,
	}
	return p
}

// nextToken advances to the next token
func (p *NoPEGParser) nextToken() {
	p.currentToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

// initTokens initializes the current and peek tokens
func (p *NoPEGParser) initTokens() {
	p.nextToken() // Set currentToken
	p.nextToken() // Set peekToken
}

// parseBlock parses a block of tokens
func (p *NoPEGParser) parseBlock(blockType int) (env.Object, error) {
	// Capture location information before skipping the opening token
	blockLine := p.currentToken.Line
	blockCol := p.currentToken.Col

	// Skip the opening token
	if p.peekToken.Type == NPEG_TOKEN_ERROR {
		return nil, fmt.Errorf("%s", p.currentToken.Value)
	}
	p.nextToken()
	if p.peekToken.Type == NPEG_TOKEN_ERROR {
		return nil, fmt.Errorf("%s", p.currentToken.Value)
	}

	// Collect all tokens until the closing token
	var objects []env.Object

	for {
		// Check for end of block
		if p.currentToken.Type == NPEG_TOKEN_BLOCK_END && blockType == 0 ||
			p.currentToken.Type == NPEG_TOKEN_BBLOCK_END && blockType == 1 ||
			p.currentToken.Type == NPEG_TOKEN_GROUP_END && blockType == 2 ||
			p.currentToken.Type == NPEG_TOKEN_BBLOCK_END && blockType == 3 ||
			p.currentToken.Type == NPEG_TOKEN_GROUP_END && blockType == 4 ||
			p.currentToken.Type == NPEG_TOKEN_BLOCK_END && blockType == 5 {
			break
		}

		// Check for unexpected end of input
		if p.currentToken.Type == NPEG_TOKEN_EOF {
			blockTypeName := "block"
			closingDelimiter := "}"
			switch blockType {
			case 0:
				blockTypeName = "block"
				closingDelimiter = "}"
			case 1:
				blockTypeName = "bblock"
				closingDelimiter = "]"
			case 2:
				blockTypeName = "group"
				closingDelimiter = ")"
			case 3:
				blockTypeName = "opbblock"
				closingDelimiter = "]"
			case 4:
				blockTypeName = "opgroup"
				closingDelimiter = ")"
			case 5:
				blockTypeName = "opblock"
				closingDelimiter = "}"
			}
			return nil, fmt.Errorf("unexpected end of input while parsing %s. Missing closing delimiter '%s'", blockTypeName, closingDelimiter)
		}

		// fmt.Println("BEFORE CALLING PARSE TOKEN ON PARSER")
		if p.peekToken.Type == NPEG_TOKEN_ERROR {
			// fmt.Println("= PARSE BLOCK ERROR DETECTED IN LOOP====>")
			// fmt.Println(p.l.pos)
			return nil, fmt.Errorf("%s", p.currentToken.Value)
		}

		// Parse the current token
		obj, err := p.parseToken()
		if err != nil {
			// fmt.Println("= PARSE BLOCK ERROR DETECTED IN LOOP====>")
			// fmt.Println(p.l.pos)
			return nil, err
		}

		// Skip comments and location nodes (we no longer add them to blocks)
		if obj != nil {
			objects = append(objects, obj)
		}

		p.nextToken()
		// fmt.Println("==parseBlock END OF LOOP===>")
		// fmt.Println(p.l.pos)

	}

	// Create the block with location information
	ser := env.NewTSeries(objects)
	return *env.NewBlockWithLocation(*ser, blockType, p.l.scriptPath, blockLine, blockCol), nil
}

// parseToken parses a single token and returns the corresponding Rye object
func (p *NoPEGParser) parseToken() (env.Object, error) {
	// fmt.Println("PARSE TOKEN ENTER (((()))))")
	switch p.currentToken.Type {
	case NPEG_TOKEN_LOCATION_NODE:
		// Location nodes are no longer used - skip them
		return nil, nil
	case NPEG_TOKEN_BLOCK_START:
		return p.parseBlock(0)
	case NPEG_TOKEN_BBLOCK_START:
		return p.parseBlock(1)
	case NPEG_TOKEN_GROUP_START:
		return p.parseBlock(2)
	case NPEG_TOKEN_OPBBLOCK_START:
		return p.parseBlock(3)
	case NPEG_TOKEN_OPGROUP_START:
		return p.parseBlock(4)
	case NPEG_TOKEN_OPBLOCK_START:
		return p.parseBlock(5)
	case NPEG_TOKEN_WORD:
		idx := p.wordIndex.IndexWord(p.currentToken.Value)
		return *env.NewWord(idx), nil
	case NPEG_TOKEN_SETWORD:
		word := p.currentToken.Value
		idx := p.wordIndex.IndexWord(word[:len(word)-1])
		return *env.NewSetword(idx), nil
	case NPEG_TOKEN_LMODWORD:
		word := p.currentToken.Value
		idx := p.wordIndex.IndexWord(word[2:])
		return *env.NewLModword(idx), nil
	case NPEG_TOKEN_LSETWORD:
		word := p.currentToken.Value
		idx := p.wordIndex.IndexWord(word[1:])
		return *env.NewLSetword(idx), nil
	case NPEG_TOKEN_MODWORD:
		word := p.currentToken.Value
		idx := p.wordIndex.IndexWord(word[:len(word)-2])
		return *env.NewModword(idx), nil
	case NPEG_TOKEN_GETWORD:
		word := p.currentToken.Value
		idx := p.wordIndex.IndexWord(word[1:])
		return *env.NewGetword(idx), nil
	case NPEG_TOKEN_OPWORD:
		var word string
		if p.currentToken.Value[0] == '.' && len(p.currentToken.Value) > 1 {
			word = p.currentToken.Value[1:]
		} else {
			word = p.currentToken.Value
		}
		var idx int
		force := 0
		if len(word) == 1 || word == "<<" || word == "<-" || word == "<~" || word == ">=" || word == "<=" || word == "//" || word == ".." || word == "++" || word == "." || word == "|" {
			idx = p.wordIndex.IndexWord("_" + word)
		} else {
			if word[len(word)-1:] == "*" {
				force = 1
				word = word[:len(word)-1]
			}
			idx = p.wordIndex.IndexWord(word)
		}
		return *env.NewOpword(idx, force), nil
	case NPEG_TOKEN_PIPEWORD:
		var word string
		if p.currentToken.Value != "|" {
			word = p.currentToken.Value[1:]
		} else {
			word = p.currentToken.Value
		}
		var idx int
		force := 0
		if len(word) == 1 || word == ">>" || word == "->" || word == "~>" || word == "-->" || word == ".." || word == "|" {
			idx = p.wordIndex.IndexWord("_" + word)
		} else {
			if word[len(word)-1:] == "*" {
				force = 1
				word = word[:len(word)-1]
			}
			idx = p.wordIndex.IndexWord(word)
		}
		return *env.NewPipeword(idx, force), nil
	case NPEG_TOKEN_ONECHARPIPE:
		word := p.currentToken.Value
		idx := p.wordIndex.IndexWord("_" + word)
		return *env.NewPipeword(idx, 0), nil
	case NPEG_TOKEN_TAGWORD:
		word := p.currentToken.Value
		idx := p.wordIndex.IndexWord(word[1:])
		return *env.NewTagword(idx), nil
	case NPEG_TOKEN_KINDWORD:
		word := p.currentToken.Value
		idx := p.wordIndex.IndexWord(word[2 : len(word)-1])
		return *env.NewKindword(idx), nil
	case NPEG_TOKEN_XWORD:
		word := p.currentToken.Value
		parts := strings.SplitN(word[1:len(word)-1], " ", 2)
		idx := p.wordIndex.IndexWord(parts[0])
		args := ""
		if len(parts) > 1 {
			args = parts[1]
		}
		return *env.NewXword(idx, args), nil
	case NPEG_TOKEN_EXWORD:
		word := p.currentToken.Value
		idx := p.wordIndex.IndexWord(word[2 : len(word)-1])
		return *env.NewEXword(idx), nil
	case NPEG_TOKEN_GENWORD:
		word := p.currentToken.Value
		idx := p.wordIndex.IndexWord(strings.ToLower(word))
		return *env.NewGenword(idx), nil
	case NPEG_TOKEN_FLAGWORD:
		word := p.currentToken.Value
		if strings.HasPrefix(word, "--") {
			// Long flag (--verbose)
			flagName := word[2:] // Remove "--"
			idx := p.wordIndex.IndexWord(flagName)
			return *env.NewFlagword(-1, idx), nil
		} else if strings.HasPrefix(word, "-") {
			// Short flag (-v) or combined flag (-v|verbose)
			flagPart := word[1:] // Remove "-"

			// Check for combined format (short|long)
			if strings.Contains(flagPart, "|") {
				parts := strings.SplitN(flagPart, "|", 2)
				shortName := parts[0]
				longName := parts[1]
				shortIdx := p.wordIndex.IndexWord(shortName)
				longIdx := p.wordIndex.IndexWord(longName)
				return *env.NewFlagword(shortIdx, longIdx), nil
			} else {
				// Regular short flag
				idx := p.wordIndex.IndexWord(flagPart)
				return *env.NewFlagword(idx, -1), nil
			}
		}
		// This shouldn't happen due to lexer validation, but handle gracefully
		idx := p.wordIndex.IndexWord(word)
		return *env.NewFlagword(-1, idx), nil
	case NPEG_TOKEN_NUMBER:
		val, err := strconv.ParseInt(p.currentToken.Value, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid number format '%s': %s. Numbers must be integers like 42, -123, or 0", p.currentToken.Value, err.Error())
		}
		return *env.NewInteger(val), nil
	case NPEG_TOKEN_DECIMAL:
		val, err := strconv.ParseFloat(p.currentToken.Value, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid decimal format '%s': %s. Decimals must be numbers with decimal points like 3.14, -0.5, or 123.0", p.currentToken.Value, err.Error())
		}
		return *env.NewDecimal(val), nil
	case NPEG_TOKEN_STRING:
		str := p.currentToken.Value[1 : len(p.currentToken.Value)-1]
		// Process escape sequences
		str = strings.Replace(str, "\\n", "\n", -1)
		str = strings.Replace(str, "\\r", "\r", -1)
		str = strings.Replace(str, "\\t", "\t", -1)
		str = strings.Replace(str, "\\\\", "\\", -1)
		str = strings.Replace(str, "\\\"", "\"", -1)
		// str = strings.Replace(str, "\\\\", "\\", -1)
		return *env.NewString(str), nil
	case NPEG_TOKEN_URI:
		parts := strings.SplitN(p.currentToken.Value, "://", 2)
		idx := p.wordIndex.IndexWord(parts[0])
		return *env.NewUri(p.wordIndex, *env.NewWord(idx), p.currentToken.Value), nil
	case NPEG_TOKEN_EMAIL:
		return *env.NewEmail(p.currentToken.Value), nil
	case NPEG_TOKEN_FPATH:
		idx := p.wordIndex.IndexWord("file")
		return *env.NewUri(p.wordIndex, *env.NewWord(idx), "file://"+p.currentToken.Value[1:]), nil
	case NPEG_TOKEN_CPATH:
		parts := strings.Split(p.currentToken.Value, "/")
		if len(parts) == 2 {
			idx1 := p.wordIndex.IndexWord(parts[0])
			idx2 := p.wordIndex.IndexWord(parts[1])
			return *env.NewCPath2(0, *env.NewWord(idx1), *env.NewWord(idx2)), nil
		} else if len(parts) >= 3 {
			idx1 := p.wordIndex.IndexWord(parts[0])
			idx2 := p.wordIndex.IndexWord(parts[1])
			idx3 := p.wordIndex.IndexWord(parts[2])
			return *env.NewCPath3(0, *env.NewWord(idx1), *env.NewWord(idx2), *env.NewWord(idx3)), nil
		}
		return nil, fmt.Errorf("invalid context path: %s", p.currentToken.Value)
	case NPEG_TOKEN_OPCPATH:
		parts := strings.Split(p.currentToken.Value[1:], "/")
		if len(parts) == 2 {
			idx1 := p.wordIndex.IndexWord(parts[0])
			idx2 := p.wordIndex.IndexWord(parts[1])
			return *env.NewCPath2(1, *env.NewWord(idx1), *env.NewWord(idx2)), nil
		} else if len(parts) >= 3 {
			idx1 := p.wordIndex.IndexWord(parts[0])
			idx2 := p.wordIndex.IndexWord(parts[1])
			idx3 := p.wordIndex.IndexWord(parts[2])
			return *env.NewCPath3(1, *env.NewWord(idx1), *env.NewWord(idx2), *env.NewWord(idx3)), nil
		}
		return nil, fmt.Errorf("invalid op context path: %s", p.currentToken.Value)
	case NPEG_TOKEN_PIPECPATH:
		parts := strings.Split(p.currentToken.Value[1:], "/")
		if len(parts) == 2 {
			idx1 := p.wordIndex.IndexWord(parts[0])
			idx2 := p.wordIndex.IndexWord(parts[1])
			return *env.NewCPath2(2, *env.NewWord(idx1), *env.NewWord(idx2)), nil
		} else if len(parts) >= 3 {
			idx1 := p.wordIndex.IndexWord(parts[0])
			idx2 := p.wordIndex.IndexWord(parts[1])
			idx3 := p.wordIndex.IndexWord(parts[2])
			return *env.NewCPath3(2, *env.NewWord(idx1), *env.NewWord(idx2), *env.NewWord(idx3)), nil
		}
		return nil, fmt.Errorf("invalid pipe context path: %s", p.currentToken.Value)
	case NPEG_TOKEN_GETCPATH:
		parts := strings.Split(p.currentToken.Value[1:], "/")
		if len(parts) == 2 {
			idx1 := p.wordIndex.IndexWord(parts[0])
			idx2 := p.wordIndex.IndexWord(parts[1])
			return *env.NewCPath2(3, *env.NewWord(idx1), *env.NewWord(idx2)), nil
		} else if len(parts) >= 3 {
			idx1 := p.wordIndex.IndexWord(parts[0])
			idx2 := p.wordIndex.IndexWord(parts[1])
			idx3 := p.wordIndex.IndexWord(parts[2])
			return *env.NewCPath3(3, *env.NewWord(idx1), *env.NewWord(idx2), *env.NewWord(idx3)), nil
		}
		return nil, fmt.Errorf("invalid get context path: %s", p.currentToken.Value)
	case NPEG_TOKEN_COMMA:
		return env.Comma{}, nil
	case NPEG_TOKEN_VOID:
		return env.Void{}, nil
	case NPEG_TOKEN_COMMENT:
		return nil, nil // Skip comments
	case NPEG_TOKEN_NONE:
		// Check if this is a spacing error
		if p.currentToken.Value == "there should be spacing between tokens" {
			return nil, fmt.Errorf("%s", p.currentToken.Value)
		}
		// Skip other unknown tokens
		return nil, nil
	case NPEG_TOKEN_ERROR:
		// fmt.Println("parseToken NOSPACING CASE-->")
		// fmt.Println(p.l.pos)
		return nil, fmt.Errorf("%s", p.currentToken.Value)
	default:
		return nil, fmt.Errorf("unknown token type: %d", p.currentToken.Type)
	}
}

// Parse parses the input and returns a Rye block

// Parse parses the input and returns a Rye block
func (p *NoPEGParser) Parse() (env.Object, error) {
	// Initialize tokens
	p.initTokens()

	// Expect a block start
	if p.currentToken.Type != NPEG_TOKEN_BLOCK_START {
		return nil, fmt.Errorf("expected block start, got %s", p.currentToken.Value)
	}

	// Parse the block
	return p.parseBlock(0)
}

// formatErrorLocationNoPEG creates a visual representation of the error location
func formatErrorLocationNoPEG(line string, col int) string {
	var bu strings.Builder

	// Add the line with error in bright white
	bu.WriteString("\x1b[1;37m" + line + "\x1b[0m\n")

	// Add pointer to error position with better visibility
	if col > 0 && col <= len(line)+1 {
		// Create a more visible error pointer in bold red
		bu.WriteString("\x1b[1;31m" + strings.Repeat(" ", col-1) + "^\n")
		bu.WriteString(strings.Repeat(" ", col-1) + "|\x1b[0m\n")
	}

	return bu.String()
}

// inferErrorContextNoPEG tries to provide helpful context about what might be wrong
func inferErrorContextNoPEG(tok NoPEGToken, err error, line string, col int, fullInput string, lineNum int) string {
	switch tok.Err {
	case ERR_SPACING_BLK:
		if col > 0 && col <= len(line) {
			prevChar := "?"
			nextChar := "?"
			if col-1 < len(line) {
				prevChar = string(line[col-1])
			}
			if col < len(line) {
				nextChar = string(line[col])
			}
			return fmt.Sprintf("Missing space between '%s' and block delimiter '%s'. Block delimiters ({ } [ ] ( )) must be separated from adjacent tokens by spaces.", prevChar, nextChar)
		}
		return "Missing space around block delimiters. Block delimiters ({ } [ ] ( )) must be separated from adjacent tokens by spaces."
	case ERR_SPACING_OP:
		if col > 0 && col <= len(line) {
			prevChar := "?"
			nextChar := "?"
			if col-1 < len(line) {
				prevChar = string(line[col-1])
			}
			if col < len(line) {
				nextChar = string(line[col])
			}
			return fmt.Sprintf("Missing space between value '%s' and operator '%s'. Operators must be separated from values by spaces.", prevChar, nextChar)
		}
		return "Missing space around operator. Operators must be separated from values by spaces."
	case ERR_SPACING_OTHR:
		if col > 0 && col <= len(line) {
			prevChar := "?"
			nextChar := "?"
			if col-1 < len(line) {
				prevChar = string(line[col-1])
			}
			if col < len(line) {
				nextChar = string(line[col])
			}
			return fmt.Sprintf("Missing space between tokens '%s' and '%s'. All tokens must be separated by whitespace.", prevChar, nextChar)
		}
		return "Missing space between tokens. All tokens must be separated by whitespace."
	}

	// Provide more specific error messages based on the error string
	errMsg := err.Error()
	if strings.Contains(errMsg, "unexpected end of input") {
		return "Unexpected end of input. Check for missing closing delimiters like }, ], or )."
	}
	if strings.Contains(errMsg, "expected block start") {
		return "Expected opening block delimiter '{' at the beginning of the input."
	}
	if strings.Contains(errMsg, "invalid number format") {
		return fmt.Sprintf("Invalid number format in token '%s'. Check for malformed numbers.", tok.Value)
	}
	if strings.Contains(errMsg, "invalid decimal format") {
		return fmt.Sprintf("Invalid decimal format in token '%s'. Check for malformed decimal numbers.", tok.Value)
	}
	if strings.Contains(errMsg, "invalid context path") {
		return fmt.Sprintf("Invalid context path '%s'. Context paths must follow the pattern 'word/word' or 'word/word/word'.", tok.Value)
	}
	if strings.Contains(errMsg, "unknown token type") {
		return fmt.Sprintf("Unrecognized token type at position. Check for invalid characters or malformed syntax.")
	}

	return fmt.Sprintf("Syntax error with token '%s'. %s", tok.Value, errMsg)
}

/* func inferErrorContextNoPEG_TOREMOVE(line string, col int, fullInput string, lineNum int) string {
	// Check for common syntax errors
	if col > len(line) {
		return "Possible issue: Unexpected end of line. Check for missing closing delimiter."
	}

	// First, check for specific syntax patterns at the error position
	if col > 0 && col <= len(line) {
		// Get the character at the error position
		errorChar := ' '
		if col-1 < len(line) {
			errorChar = rune(line[col-1])
		}

		// Check for operator-related errors
		if errorChar == '+' || errorChar == '-' || errorChar == '*' || errorChar == '/' {
			// Check if operator is at the beginning of an expression or line
			if col == 1 || (col > 1 && (isWhitespaceCh(line[col-2]) || isSpecialChar(line[col-2]))) {
				return "Possible issue: Operator at beginning of expression. Operators must be between values."
			}

			// Check if operator is at the end of an expression
			if col == len(line) || (col < len(line) && (isWhitespaceCh(line[col]) || isSpecialChar(line[col]))) {
				return "Possible issue: Operator at end of expression. Operators must be between values."
			}

			// Check for consecutive operators
			if col < len(line) && (line[col] == '+' || line[col] == '-' || line[col] == '*' || line[col] == '/') {
				return "Possible issue: Invalid consecutive operators. Operators must be separated by values."
			}
		}

		// Check for invalid characters
		if !isLetter(byte(errorChar)) && !isDigit(byte(errorChar)) && !isSpecialChar(byte(errorChar)) && !isWhitespaceCh(byte(errorChar)) {
			return fmt.Sprintf("Possible issue: Invalid character '%c' in input.", errorChar)
		}
	}

	// Check for specific patterns in the entire line
	if strings.Contains(line, "+") && !strings.Contains(line, " + ") {
		return "Possible issue: Invalid syntax. Operators must be between values with spaces."
	}

	if strings.Contains(line, ",") && !strings.Contains(line, " , ") {
		return "Possible issue: Invalid syntax. Commas must be surrounded by spaces."
	}

	if strings.Contains(line, "\"") && strings.Contains(line, "+") &&
		strings.Index(line, "+") > strings.LastIndex(line, "\"") {
		return "Possible issue: Invalid operation with string. Check string concatenation syntax."
	}

	// Check for nested blocks without proper spacing
	if strings.Contains(line, "}") {
		for i := 0; i < len(line)-1; i++ {
			if line[i] == '}' && i+1 < len(line) && !isWhitespaceCh(line[i+1]) && line[i+1] != '}' {
				return "Possible issue: Missing space after nested block. Add spaces between blocks and other elements."
			}
		}
	}

	// Check for unclosed string
	if strings.Count(line, "\"")%2 != 0 {
		return "Possible issue: Unclosed string. Add a closing double quote."
	}

	// Check for mismatched delimiters in the entire line
	if strings.Count(line, "{") != strings.Count(line, "}") {
		// Don't return this as it's often a false positive due to the automatic wrapping
		// return "Possible issue: Mismatched braces in line."
	}

	if strings.Count(line, "[") != strings.Count(line, "]") {
		return "Possible issue: Mismatched brackets in line."
	}

	if strings.Count(line, "(") != strings.Count(line, ")") {
		return "Possible issue: Mismatched parentheses in line."
	}

	// Generic message if we can't infer anything specific
	return "Syntax error at this position. Check for invalid tokens or incorrect operators."
} */

// suggestFixNoPEG provides suggestions for fixing common syntax errors
/* func suggestFixNoPEG(line string, col int, fullInput string, lineNum int) string {
	// Check for specific patterns in the line
	if strings.Contains(line, "+") && !strings.Contains(line, " + ") {
		return "Add spaces between numbers and operators, e.g., '123 + 456'."
	}

	if strings.Contains(line, ",") && !strings.Contains(line, " , ") {
		return "Add spaces around commas, e.g., '123 , 456'."
	}

	if strings.Contains(line, "\"") && strings.Contains(line, "+") &&
		strings.Index(line, "+") > strings.LastIndex(line, "\"") {
		return "For string concatenation, use proper syntax like 'join \"string\" value'."
	}

	// Check for nested blocks without proper spacing
	if strings.Contains(line, "}") {
		for i := 0; i < len(line)-1; i++ {
			if line[i] == '}' && i+1 < len(line) && !isWhitespaceCh(line[i+1]) && line[i+1] != '}' {
				return "Add spaces between nested blocks and other elements, e.g., '{ ... } value' instead of '{ ... }value'."
			}
		}
	}

	// Check for operator at beginning or end of expression
	if col > 0 && col <= len(line) {
		char := line[col-1]
		if char == '+' || char == '-' || char == '*' || char == '/' {
			// At beginning of expression
			if col == 1 || (col > 1 && (isWhitespaceCh(line[col-2]) || isSpecialChar(line[col-2]))) {
				return "Operators must be between values, not at the beginning of an expression."
			}

			// At end of expression
			if col == len(line) || (col < len(line) && (isWhitespaceCh(line[col]) || isSpecialChar(line[col]))) {
				return "Operators must be between values, not at the end of an expression."
			}
		}
	}

	// Check for unclosed string
	if strings.Count(line, "\"")%2 != 0 {
		return "Add a closing double quote (\") to complete the string."
	}

	// Check for mismatched brackets
	if strings.Count(line, "[") != strings.Count(line, "]") {
		openBrackets := strings.Count(line, "[") - strings.Count(line, "]")
		if openBrackets > 0 {
			return "Add " + strings.Repeat("]", openBrackets) + " to close open brackets."
		} else {
			return "Remove extra closing brackets or add matching opening brackets."
		}
	}

	// Check for mismatched parentheses
	if strings.Count(line, "(") != strings.Count(line, ")") {
		openParens := strings.Count(line, "(") - strings.Count(line, ")")
		if openParens > 0 {
			return "Add " + strings.Repeat(")", openParens) + " to close open parentheses."
		} else {
			return "Remove extra closing parentheses or add matching opening parentheses."
		}
	}

	// Check for invalid characters
	if col > 0 && col <= len(line) {
		char := line[col-1]
		if !isLetter(char) && !isDigit(char) && !isSpecialChar(char) && !isWhitespaceCh(char) {
			return fmt.Sprintf("Remove or replace the invalid character '%c'.", char)
		}
	}

	// Generic suggestion
	return "Check your syntax. Ensure operators are between values and expressions are properly formed."
} */

// enhanceErrorMessageNoPEG improves the error message with better context and suggestions
func enhanceErrorMessageNoPEG(tok NoPEGToken, err error, input string, filePath string, parser *NoPEGParser) string {

	// Get line and column from the parser's lexer
	lineNum := parser.l.line
	colNum := parser.l.col

	// Get the line with the error
	lines := strings.Split(input, "\n")
	if lineNum <= 0 || lineNum > len(lines) {
		return err.Error() // Fallback if line number is invalid
	}

	line := lines[lineNum-1]

	// Build enhanced error message
	var bu strings.Builder

	// Red background banner for syntax errors
	bu.WriteString("\x1b[41m\x1b[30m SYNTAX ERROR \x1b[0m\n") // Red background, black text

	// Add error location with colors
	if filePath != "" {
		bu.WriteString("\x1b[1;31mSyntax error\x1b[0m in \x1b[1;34m" + filePath +
			"\x1b[0m at line \x1b[1;33m" + fmt.Sprintf("%d", lineNum) +
			"\x1b[0m, column \x1b[1;33m" + fmt.Sprintf("%d", colNum) + "\x1b[0m\n")
	} else {
		bu.WriteString("\x1b[1;31mSyntax error\x1b[0m at line \x1b[1;33m" +
			fmt.Sprintf("%d", lineNum) + "\x1b[0m, column \x1b[1;33m" +
			fmt.Sprintf("%d", colNum) + "\x1b[0m\n")
	}

	// Add the error location visualization
	bu.WriteString(formatErrorLocationNoPEG(line, colNum))

	// Add context about what might be wrong
	errorContext := inferErrorContextNoPEG(tok, err, line, colNum, input, lineNum)
	if errorContext != "" {
		bu.WriteString("\x1b[33m" + errorContext + "\x1b[0m\n")
	}

	// Add suggestions for fixing the error
	// suggestion := suggestFixNoPEG(line, colNum, input, lineNum)
	// if suggestion != "" {
	//	bu.WriteString("Suggestion: " + suggestion + "\n")
	// }

	// Add a red separator line after error for better visibility
	bu.WriteString("\x1b[1;31m" + strings.Repeat("─", 50) + "\x1b[0m\n")

	return bu.String()
}

// LoadStringNoPEG loads a string using the non-PEG parser
func LoadStringNoPEG(input string, sig bool) (env.Object, *env.Idxs) {
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

	input = "{ " + input + " } "

	//fmt.Println(input)

	parser := NewParserNoPEG(input, wordIndex)
	if parser == nil {
		return *env.NewError("Failed to create parser"), wordIndex
	}

	val, err := parser.Parse()
	// fmt.Println("parser.l.col")
	// fmt.Println(parser.l.col)
	if err != nil {
		// fmt.Println(err.Error())
		// return *env.NewError(err.Error()), wordIndex
		// Enhanced error handling with better context
		errStr := enhanceErrorMessageNoPEG(parser.peekToken, err, input, "", parser)
		return *env.NewError(errStr), wordIndex
	}

	return val.(env.Block), wordIndex
}

// LoadStringNEWNoPEG loads a string using the non-PEG parser with a program state
// This version injects LocationNodes during parsing for better error reporting
func LoadStringNEWNoPEG(input string, sig bool, ps *env.ProgramState) env.Object {
	if sig {
		signed := checkCodeSignature(input)
		if signed == -1 {
			return *env.NewError("Signature not found")
		} else if signed == -2 {
			return *env.NewError("Invalid signature")
		}
	}

	input = removeBangLine(input)
	input = "{ " + input + " }"

	wordIndexMutex.Lock()
	wordIndex = ps.Idx

	parser := NewParserNoPEGWithLocationNodes(input, wordIndex, ps.ScriptPath)
	if parser == nil {
		wordIndexMutex.Unlock()
		return *env.NewError("Failed to create parser")
	}

	val, err := parser.Parse()
	if err != nil {
		wordIndexMutex.Unlock()
		var bu strings.Builder
		bu.WriteString("In file " + util.TermBold(ps.ScriptPath) + " at line " + strconv.Itoa(parser.l.line) + "\n")
		errStr := enhanceErrorMessageNoPEG(parser.peekToken, err, input, ps.ScriptPath, parser)
		bu.WriteString(errStr)

		ps.FailureFlag = true
		return *env.NewError(bu.String())
	}

	ps.Idx = wordIndex
	wordIndexMutex.Unlock()

	return val.(env.Block)
}
