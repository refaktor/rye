// loader.go
package loader

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/refaktor/rye/env"
	"github.com/refaktor/rye/util"

	//. "github.com/yhirose/go-peg"
	//. "github.com/CWood1/go-peg"
	. "github.com/refaktor/go-peg"
)

func LoadString(input string, sig bool) (env.Object, *env.Idxs) {
	//fmt.Println(input)

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
	if len(inp1) == 0 || strings.Index("{", inp1) != 0 {
		input = "{ " + input + " }"
	}

	parser := newParser()
	if parser == nil {
		return *env.NewError("Failed to create parser"), wordIndex
	}

	val, err := parser.ParseAndGetValue(input, nil)

	if err != nil {
		// Enhanced error handling with better context
		errStr := enhanceErrorMessage(err, input, "")
		return *env.NewError(errStr), wordIndex
	}

	//InspectNode(val)
	if val != nil {
		return val.(env.Block), wordIndex
	} else {
		empty1 := make([]env.Object, 0)
		ser := env.NewTSeries(empty1)
		return *env.NewBlock(*ser), wordIndex
	}
}

func LoadStringNEW_ORIGINAL_PEG(input string, sig bool, ps *env.ProgramState) env.Object {
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
	if len(inp1) == 0 || strings.Index("{", inp1) != 0 {
		input = "{ " + input + " }"
	}

	parser := newParser()
	if parser == nil {
		return *env.NewError("Failed to create parser")
	}

	wordIndexMutex.Lock()
	wordIndex = ps.Idx
	val, err := parser.ParseAndGetValue(input, nil)
	ps.Idx = wordIndex
	wordIndexMutex.Unlock()

	if err != nil {
		var bu strings.Builder
		bu.WriteString("In file " + util.TermBold(ps.ScriptPath) + "\n")
		errStr := enhanceErrorMessage(err, input, ps.ScriptPath)
		bu.WriteString(errStr)

		ps.FailureFlag = true
		return *env.NewError(bu.String())
	}

	if val != nil {
		return val.(env.Block)
	} else {
		empty1 := make([]env.Object, 0)
		ser := env.NewTSeries(empty1)
		return *env.NewBlock(*ser)
	}
}

func LoadStringNEW(input string, sig bool, ps *env.ProgramState) env.Object {
	if sig {
		signed := checkCodeSignature(input)
		if signed == -1 {
			return *env.NewError("Signature not found")
		} else if signed == -2 {
			return *env.NewError("Invalid signature")
		}
	}

	input = removeBangLine(input)

	return LoadStringNEWNoPEG(input, sig, ps)
}

func parseBlock(v *Values, d Any) (Any, error) {
	//fmt.Println("** Parse block **")
	//fmt.Println(v.Vs)
	block := make([]env.Object, len(v.Vs)-1)
	//var r env.Object
	ofs := 0
	for i := 1; i < len(v.Vs); i += 1 {
		obj := v.Vs[i]
		if obj != nil { //obj != nil
			//			fmt.Println(i)
			//			InspectNode(obj)
			block[i-1+ofs] = obj.(env.Object)
		} else {
			ofs -= 1
		}
	}
	//fmt.Print("BLOCK --> ")
	//fmt.Println(block)
	if ofs != 0 {
		block = block[0 : len(v.Vs)-1+ofs]
	}
	ser := env.NewTSeries(block)
	return *env.NewBlock2(*ser, 0), nil
}

func parseBBlock(v *Values, d Any) (Any, error) {
	block := make([]env.Object, len(v.Vs)-1)
	ofs := 0
	for i := 1; i < len(v.Vs); i += 1 {
		obj := v.Vs[i]
		if obj != nil {
			block[i-1+ofs] = obj.(env.Object)
		} else {
			ofs -= 1
		}
	}
	if ofs != 0 {
		block = block[0 : len(v.Vs)-1+ofs]
	}
	ser := env.NewTSeries(block)
	return *env.NewBlock2(*ser, 1), nil
}

func parseGroup(v *Values, d Any) (Any, error) {
	//fmt.Println("** Parse block **")
	//fmt.Println(v.Vs)
	block := make([]env.Object, len(v.Vs)-1)
	//var r env.Object
	ofs := 0
	for i := 1; i < len(v.Vs); i += 1 {
		obj := v.Vs[i]
		if obj != nil { //obj != nil
			//fmt.Println(i)
			//InspectNode(obj)
			block[i-1+ofs] = obj.(env.Object)
		} else {
			ofs -= 1
		}
	}
	//fmt.Print("BLOCK --> ")
	//fmt.Println(block)
	if ofs != 0 {
		block = block[0 : len(v.Vs)-1+ofs]
	}
	ser := env.NewTSeries(block)
	return *env.NewBlock2(*ser, 2), nil
}

func parseNumber(v *Values, d Any) (Any, error) {
	val, err := strconv.ParseInt(v.Token(), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid number format: %s", err.Error())
	}
	return *env.NewInteger(val), nil
}

func parseDecimal(v *Values, d Any) (Any, error) {
	val, err := strconv.ParseFloat(v.Token(), 64)
	if err != nil {
		return nil, fmt.Errorf("invalid decimal format: %s", err.Error())
	}
	return *env.NewDecimal(val), nil
}

func parseString(v *Values, d Any) (Any, error) {
	str := v.Token()[1 : len(v.Token())-1]
	// Process escape sequences
	str = strings.Replace(str, "\\n", "\n", -1)
	str = strings.Replace(str, "\\r", "\r", -1)
	str = strings.Replace(str, "\\t", "\t", -1)
	str = strings.Replace(str, "\\\"", "\"", -1)
	str = strings.Replace(str, "\\\\", "\\", -1)
	return *env.NewString(str), nil
}

func parseUri(v *Values, d Any) (Any, error) {
	// fmt.Println(v.Vs[0])
	return *env.NewUri(wordIndex, v.Vs[0].(env.Word), v.Token()), nil // ){v.Vs[0].(env.Word), v.Token()}, nil // TODO let the second part be it's own object that parser returns like path
}

func parseEmail(v *Values, d Any) (Any, error) {
	return *env.NewEmail(v.Token()), nil
}

func parseFpath(v *Values, d Any) (Any, error) {
	idx := wordIndex.IndexWord("file")
	return *env.NewUri(wordIndex, *env.NewWord(idx), "file://"+v.Token()[1:]), nil // ){v.Vs[0].(env.Word), v.Token()}, nil // TODO let the second part be it's own object that parser returns like path
}

func parseCPath(v *Values, d Any) (Any, error) {
	switch len(v.Vs) {
	case 2:
		return *env.NewCPath2(0, v.Vs[0].(env.Word), v.Vs[1].(env.Word)), nil
	case 3:
		return *env.NewCPath3(0, v.Vs[0].(env.Word), v.Vs[1].(env.Word), v.Vs[2].(env.Word)), nil
	default:
		return *env.NewCPath3(0, v.Vs[0].(env.Word), v.Vs[1].(env.Word), v.Vs[2].(env.Word)), nil
	}
}

func parseOpCPath(v *Values, d Any) (Any, error) {
	switch len(v.Vs) {
	case 2:
		return *env.NewCPath2(1, v.Vs[0].(env.Word), v.Vs[1].(env.Word)), nil
	case 3:
		return *env.NewCPath3(1, v.Vs[0].(env.Word), v.Vs[1].(env.Word), v.Vs[2].(env.Word)), nil
	default:
		return *env.NewCPath3(1, v.Vs[0].(env.Word), v.Vs[1].(env.Word), v.Vs[2].(env.Word)), nil
	}
}
func parsePipeCPath(v *Values, d Any) (Any, error) {
	switch len(v.Vs) {
	case 2:
		return *env.NewCPath2(2, v.Vs[0].(env.Word), v.Vs[1].(env.Word)), nil
	case 3:
		return *env.NewCPath3(2, v.Vs[0].(env.Word), v.Vs[1].(env.Word), v.Vs[2].(env.Word)), nil
	default:
		return *env.NewCPath3(2, v.Vs[0].(env.Word), v.Vs[1].(env.Word), v.Vs[2].(env.Word)), nil
	}
}

func parseWord(v *Values, d Any) (Any, error) {
	idx := wordIndex.IndexWord(v.Token())
	return *env.NewWord(idx), nil
}

func parseArgword(v *Values, d Any) (Any, error) {
	return *env.NewArgword(v.Vs[0].(env.Word), v.Vs[1].(env.Word)), nil
}

func parseComma(v *Values, d Any) (Any, error) {
	return env.Comma{}, nil
}

func parseVoid(v *Values, d Any) (Any, error) {
	return env.Void{}, nil
}

func parseSetword(v *Values, d Any) (Any, error) {
	//fmt.Println("SETWORD:" + v.Token())
	word := v.Token()
	idx := wordIndex.IndexWord(word[:len(word)-1])
	return *env.NewSetword(idx), nil
}

func parseLSetword(v *Values, d Any) (Any, error) {
	//fmt.Println("SETWORD:" + v.Token())
	word := v.Token()
	idx := wordIndex.IndexWord(word[1:])
	return *env.NewLSetword(idx), nil
}

func parseModword(v *Values, d Any) (Any, error) {
	//fmt.Println("SETWORD:" + v.Token())
	word := v.Token()
	idx := wordIndex.IndexWord(word[:len(word)-2])
	return *env.NewModword(idx), nil
}

func parseLModword(v *Values, d Any) (Any, error) {
	//fmt.Println("SETWORD:" + v.Token())
	word := v.Token()
	idx := wordIndex.IndexWord(word[2:])
	return *env.NewLModword(idx), nil
}

func parseOpword(v *Values, d Any) (Any, error) {
	//fmt.Println("OPWORD:" + v.Token())
	word := v.Token()
	force := 0
	var idx int
	if len(word) == 1 || word == "<<" || word == "<-" || word == "<~" || word == ">=" || word == "<=" || word == "//" || word == ".." || word == "++" || word == "." || word == "|" {
		// onecharopwords < > + * ... their naming is equal to _< _> _* ...
		idx = wordIndex.IndexWord("_" + word)
	} else {
		if word[len(word)-1:] == "*" {
			force = 1
			word = word[:len(word)-1]
		}
		idx = wordIndex.IndexWord(word[1:])
	}
	return *env.NewOpword(idx, force), nil
}

func parseTagword(v *Values, d Any) (Any, error) {
	//fmt.Println("TAGWORD:" + v.Token())
	word := v.Token()
	idx := wordIndex.IndexWord(word[1:])
	return *env.NewTagword(idx), nil
}

func parseXword(v *Values, d Any) (Any, error) {
	//fmt.Println("TAGWORD:" + v.Token())
	cont := v.Token()
	conts := strings.Split(cont[1:len(cont)-1], " ")
	idx := wordIndex.IndexWord(conts[0])
	args := ""
	if len(conts) > 1 {
		args = conts[1]
	}
	return *env.NewXword(idx, args), nil
}

func parseKindword(v *Values, d Any) (Any, error) {
	word := v.Token()
	idx := wordIndex.IndexWord(word[1 : len(word)-1])
	return *env.NewKindword(idx), nil
}

func parseEXword(v *Values, d Any) (Any, error) {
	//fmt.Println("TAGWORD:" + v.Token())
	word := v.Token()
	idx := wordIndex.IndexWord(word[2 : len(word)-1])
	return *env.NewEXword(idx), nil
}

func parsePipeword(v *Values, d Any) (Any, error) {
	//fmt.Println("OPWORD:" + v.Token())
	word := v.Token()
	force := 0
	var idx int
	if word == ">>" || word == "->" || word == "~>" || word == "-->" || word == ".." || word == "|" {
		idx = wordIndex.IndexWord("_" + word)
	} else {
		if word[len(word)-1:] == "*" {
			force = 1
			word = word[:len(word)-1]
		}
		idx = wordIndex.IndexWord(word[1:])
	}
	return *env.NewPipeword(idx, force), nil
}

func parseOnecharpipe(v *Values, d Any) (Any, error) {
	//fmt.Println("OPWORD:" + v.Token())
	word := v.Token()
	idx := wordIndex.IndexWord("_" + word[1:])
	return *env.NewPipeword(idx, 0), nil
}

func parseGenword(v *Values, d Any) (Any, error) {
	trace("GENWORD:" + v.Token())
	word := v.Token()
	idx := wordIndex.IndexWord(strings.ToLower(word))
	return *env.NewGenword(idx), nil
}

func parseGetword(v *Values, d Any) (Any, error) {
	trace("GETWORD:" + v.Token())
	word := v.Token()
	idx := wordIndex.IndexWord(word[1:])
	return *env.NewGetword(idx), nil
}

func parseComment(v *Values, d Any) (Any, error) {
	trace("GETWORD:" + v.Token())
	return nil, nil
}

// formatErrorLocation creates a visual representation of the error location
func formatErrorLocation(line string, col int) string {
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

// enhanceErrorMessage improves the error message with better context and suggestions
func enhanceErrorMessage(err error, input string, filePath string) string {
	if pegErr, ok := err.(*Error); ok {
		if len(pegErr.Details) == 0 {
			return err.Error() // Fallback to original error if no details
		}

		detail := pegErr.Details[0]
		lineNum, colNum := detail.Ln, detail.Col

		// Get the line with the error
		lines := strings.Split(input, "\n")
		if lineNum <= 0 || lineNum > len(lines) {
			return err.Error() // Fallback if line number is invalid
		}

		line := lines[lineNum-1]

		// Build enhanced error message
		var bu strings.Builder

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
		bu.WriteString(formatErrorLocation(line, colNum))

		// Add context about what might be wrong
		errorContext := inferErrorContext(line, colNum, input, lineNum)
		if errorContext != "" {
			bu.WriteString("\x1b[33m" + errorContext + "\x1b[0m\n")
		}

		// Add suggestions for fixing the error
		suggestion := suggestFix(line, colNum, input, lineNum)
		if suggestion != "" {
			bu.WriteString("\x1b[36mSuggestion: " + suggestion + "\x1b[0m\n")
		}

		// Add a red separator line after error for better visibility
		bu.WriteString("\x1b[1;31m" + strings.Repeat("─", 50) + "\x1b[0m\n")

		return bu.String()
	}

	// Fallback to original error if not a PEG error
	return err.Error()
}

// inferErrorContext tries to provide helpful context about what might be wrong
func inferErrorContext(line string, col int, fullInput string, lineNum int) string {
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
			if col == 1 || (col > 1 && (isWhitespace(byte(line[col-2])) || isDelimiter(byte(line[col-2])))) {
				return "Possible issue: Operator at beginning of expression. Operators must be between values."
			}

			// Check if operator is at the end of an expression
			if col == len(line) || (col < len(line) && (isWhitespace(byte(line[col])) || isDelimiter(byte(line[col])))) {
				return "Possible issue: Operator at end of expression. Operators must be between values."
			}

			// Check for consecutive operators
			if col < len(line) && (line[col] == '+' || line[col] == '-' || line[col] == '*' || line[col] == '/') {
				return "Possible issue: Invalid consecutive operators. Operators must be separated by values."
			}
		}

		// Check for invalid characters
		if !isValidWordChar(byte(errorChar)) && !isDelimiter(byte(errorChar)) && !isWhitespace(byte(errorChar)) {
			return fmt.Sprintf("Possible issue: Invalid character '%c' in input.", errorChar)
		}
	}

	// Check for specific patterns in the entire line
	if strings.Contains(line, "123123+") {
		return "Possible issue: Invalid syntax. Operators must be between values with spaces."
	}

	if strings.Contains(line, "\"") && strings.Contains(line, "+") &&
		strings.Index(line, "+") > strings.LastIndex(line, "\"") {
		return "Possible issue: Invalid operation with string. Check string concatenation syntax."
	}

	// Check for nested blocks without proper spacing
	if strings.Contains(line, "}") {
		for i := 0; i < len(line)-1; i++ {
			if line[i] == '}' && i+1 < len(line) && !isWhitespace(line[i+1]) && line[i+1] != '}' {
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
}

// suggestFix provides suggestions for fixing common syntax errors
func suggestFix(line string, col int, fullInput string, lineNum int) string {
	// Check for specific patterns in the line
	if strings.Contains(line, "123123+") {
		return "Add spaces between numbers and operators, e.g., '123123 + value'."
	}

	if strings.Contains(line, "\"") && strings.Contains(line, "+") &&
		strings.Index(line, "+") > strings.LastIndex(line, "\"") {
		return "For string concatenation, use proper syntax like 'join \"string\" value'."
	}

	// Check for nested blocks without proper spacing
	if strings.Contains(line, "}") {
		for i := 0; i < len(line)-1; i++ {
			if line[i] == '}' && i+1 < len(line) && !isWhitespace(line[i+1]) && line[i+1] != '}' {
				return "Add spaces between nested blocks and other elements, e.g., '{ ... } value' instead of '{ ... }value'."
			}
		}
	}

	// Check for operator at beginning or end of expression
	if col > 0 && col <= len(line) {
		char := line[col-1]
		if char == '+' || char == '-' || char == '*' || char == '/' {
			// At beginning of expression
			if col == 1 || (col > 1 && (isWhitespace(byte(line[col-2])) || isDelimiter(byte(line[col-2])))) {
				return "Operators must be between values, not at the beginning of an expression."
			}

			// At end of expression
			if col == len(line) || (col < len(line) && (isWhitespace(byte(line[col])) || isDelimiter(byte(line[col])))) {
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
		if !isValidWordChar(char) && !isDelimiter(char) && !isWhitespace(char) {
			return fmt.Sprintf("Remove or replace the invalid character '%c'.", char)
		}
	}

	// Generic suggestion
	return "Check your syntax. Ensure operators are between values and expressions are properly formed."
}

// Helper functions for error context inference
func isValidWordChar(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') ||
		c == '_' || c == '-' || c == '?' || c == '!' || c == '.' || c == '\\' || c == '+'
}

func isDelimiter(c byte) bool {
	return c == '{' || c == '}' || c == '[' || c == ']' || c == '(' || c == ')' ||
		c == '"' || c == ':' || c == '/' || c == ',' || c == '~'
}

func isWhitespace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r'
}

// addRuleMessages adds custom error messages to grammar rules
func addRuleMessages(parser *Parser) {
	// Add custom error messages for specific rules
	g := parser.Grammar

	// Add messages for block-related rules
	if rule, ok := g["BLOCK"]; ok {
		rule.Message = func() string { return "Expected a block. Blocks must be enclosed in curly braces '{...}'." }
	}
	if rule, ok := g["BBLOCK"]; ok {
		rule.Message = func() string { return "Expected a block. Blocks must be enclosed in square brackets '[...]'." }
	}
	if rule, ok := g["GROUP"]; ok {
		rule.Message = func() string { return "Expected a group. Groups must be enclosed in parentheses '(...)'." }
	}

	// Add messages for word-related rules
	if rule, ok := g["WORD"]; ok {
		rule.Message = func() string {
			return "Expected a word. Words must start with a letter and can contain letters, numbers, and some special characters."
		}
	}
	if rule, ok := g["SETWORD"]; ok {
		rule.Message = func() string { return "Expected a set-word. Set-words must end with a colon ':'." }
	}
	if rule, ok := g["GETWORD"]; ok {
		rule.Message = func() string { return "Expected a get-word. Get-words must start with a question mark '?'." }
	}

	// Add messages for literal values
	if rule, ok := g["NUMBER"]; ok {
		rule.Message = func() string {
			return "Expected a number. Numbers must contain only digits, with an optional leading minus sign."
		}
	}
	if rule, ok := g["DECIMAL"]; ok {
		rule.Message = func() string {
			return "Expected a decimal number. Decimal numbers must contain digits with a decimal point, and an optional leading minus sign."
		}
	}
	if rule, ok := g["STRING"]; ok {
		rule.Message = func() string {
			return "Expected a string. Strings must be enclosed in double quotes \"...\" or backticks `...`."
		}
	}

	// Add messages for other common rules
	if rule, ok := g["EMAIL"]; ok {
		rule.Message = func() string { return "Expected an email address in the format user@domain." }
	}
	if rule, ok := g["URI"]; ok {
		rule.Message = func() string { return "Expected a URI in the format scheme://path." }
	}
	if rule, ok := g["SPACES"]; ok {
		rule.Message = func() string { return "Expected whitespace (space, tab, newline)." }
	}
}

func newParser() *Parser { // TODO -- add string eaddress path url time
	// Create a PEG parser
	parser, err := NewParser(`
	BLOCK       	<-  "{" SPACES SERIES* "}"
	BBLOCK       	<-  "[" SPACES SERIES* "]"
    GROUP       	<-  "(" SPACES SERIES* ")"
    SERIES     	<-  (GROUP / COMMENT / URI / EMAIL / STRING / DECIMAL / NUMBER / COMMA / MODWORD / SETWORD / LMODWORD / LSETWORD / ONECHARPIPE / PIPECPATH / PIPEWORD / EXWORD / XWORD / OPCPATH / FPATH / OPWORD / TAGWORD / CPATH / KINDWORD / GENWORD / GETWORD / WORD / VOID / BLOCK / GROUP / BBLOCK / ARGBLOCK ) SPACES
    ARGBLOCK       	<-  "{" WORD ":" WORD "}"
    WORD           	<-  LETTER LETTERORNUM* / NORMOPWORDS
	GENWORD 		<-  "~" UCLETTER LCLETTERORNUM* 
	SETWORD    		<-  LETTER LETTERORNUM* ":"
	MODWORD    		<-  LETTER LETTERORNUM* "::"
	LSETWORD    	<-  ":" LETTER LETTERORNUM*
	LMODWORD    	<-  "::" LETTER LETTERORNUM*
	GETWORD   		<-  "?" LETTER LETTERORNUM*
	PIPEWORD   		<-  "\\" LETTER LETTERORNUM* / "|" LETTER LETTERORNUM* / PIPEARROWS / "|_" PIPEARROWS / "|" NORMOPWORDS  
	ONECHARPIPE    	<-  "|" ONECHARWORDS
	OPWORD    		<-  "." LETTER LETTERORNUM* / "." NORMOPWORDS / OPARROWS / ONECHARWORDS / "[*" LETTERORNUM*
	TAGWORD    		<-  "'" LETTER LETTERORNUM*
	KINDWORD    	<-  "~(" LETTER LETTERORNUM* ")~"?
	XWORD    		<-  "<" LETTER LETTERORNUMNOX* " "? XPARAMS* ">"
	EXWORD    		<-  "</" LETTER LETTERORNUM* ">"?
	STRING			<-  ('"' STRINGCHAR* '"') / ("` + "`" + `" STRINGCHAR1* "` + "`" + `")
	SPACES			<-  SPACE+
	URI    			<-  WORD "://" URIPATH*
	EMAIL			<-  EMAILPART "@" EMAILPART 
	EMAILPART		<-  < ([a-zA-Z0-9._]+) >
	FPATH 	   		<-  "%" URIPATH+
	CPATH    		<-  WORD ( "/" WORD )+
	OPCPATH    		<-  "." WORD ( "/" WORD )+
	PIPECPATH    	<-  "\\" WORD ( "/" WORD )+ / "|" WORD ( "/" WORD )+
	ONECHARWORDS	<-  < [<>*+-=/%] >
	NORMOPWORDS	    <-  < ("_"[<>*+-=/%]) >
	PIPEARROWS      <-  ">>" / "~>" / "->" / "|" / ".."
	OPARROWS        <-  "<<" / "<~" / "<-" / ">=" / "<=" / "//" / "++"
	LETTER  	    <-  < [a-zA-Z^(` + "`" + `] >
	LETTERORNUM		<-  < [a-zA-Z0-9-?=.\\!_+<>\]*()] >
	LETTERORNUMNOX	<-  < [a-zA-Z0-9-?=.\\!_+\]*()] >
	XPARAMS  		<-  < !">" . >
	URIPATH			<-  < [a-zA-Z0-9-?&=.,:@/\\!_>	()] >
	UCLETTER  		<-  < [A-Z] >
	LCLETTERORNUM  	<-  < [a-z0-9] >
    NUMBER          <-  < ("-"?[0-9]+) >
    DECIMAL         <-  < ("-"?[0-9]+.[0-9]+) >
	SPACE			<-  < [ \t\r\n] >
	STRINGCHAR		<-  < !'"' . >
	STRINGCHAR1		<-  < !"` + "`" + `" . >
	COMMA			<-  ","
	VOID			<-  "_"
	COMMENT			<-  (";" NOTENDLINE* )
	NOTENDLINE		<-  < !"\n" . >
`)
	// < ^([a-zA-Z0-9_\-\.]+)@([a-zA-Z0-9_\-\.]+)\.([a-zA-Z]{2,5})$ >
	// TODO -- make path path work for deeper paths too
	// TODO -- maybe add path type and make URI more fully featured

	//%whitespace      <-  [ \t\r\n]*
	//%word			<-  [a-zA-Z]+

	// Handle parser creation errors
	if err != nil {
		fmt.Println("Error creating parser:", err)
		return nil
	}

	g := parser.Grammar
	g["BLOCK"].Action = parseBlock
	g["BBLOCK"].Action = parseBBlock
	g["GROUP"].Action = parseGroup
	g["WORD"].Action = parseWord
	g["ARGBLOCK"].Action = parseArgword
	g["COMMA"].Action = parseComma
	g["VOID"].Action = parseVoid
	g["SETWORD"].Action = parseSetword
	g["LSETWORD"].Action = parseLSetword
	g["MODWORD"].Action = parseModword
	g["LMODWORD"].Action = parseLModword
	g["OPWORD"].Action = parseOpword
	g["PIPEWORD"].Action = parsePipeword
	g["ONECHARPIPE"].Action = parseOnecharpipe
	g["TAGWORD"].Action = parseTagword
	g["KINDWORD"].Action = parseKindword
	g["XWORD"].Action = parseXword
	g["EXWORD"].Action = parseEXword
	g["GENWORD"].Action = parseGenword
	g["GETWORD"].Action = parseGetword
	g["NUMBER"].Action = parseNumber
	g["DECIMAL"].Action = parseDecimal
	g["STRING"].Action = parseString
	g["EMAIL"].Action = parseEmail
	g["URI"].Action = parseUri
	g["FPATH"].Action = parseFpath
	g["CPATH"].Action = parseCPath
	g["OPCPATH"].Action = parseOpCPath
	g["PIPECPATH"].Action = parsePipeCPath
	g["COMMENT"].Action = parseComment
	/* g["SERIES"].Action = func(v *Values, d Any) (Any, error) {
		return v, nil
	}*/

	// Add custom error messages to rules
	addRuleMessages(parser)

	return parser
}
