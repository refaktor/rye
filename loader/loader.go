// loader.go
package loader

import (
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/refaktor/rye/env"
	"github.com/refaktor/rye/util"

	//. "github.com/yhirose/go-peg"
	//. "github.com/CWood1/go-peg"
	. "github.com/refaktor/go-peg"
)

func trace(x any) {
	//fmt.Print("\x1b[56m")
	//fmt.Print(x)
	//fmt.Println("\x1b[0m")
}

var wordIndex *env.Idxs
var wordIndexMutex sync.Mutex

func InitIndex() {
	if wordIndex == nil {
		wordIndex = env.NewIdxs()
	}
}

func GetIdxs() *env.Idxs {
	if wordIndex == nil {
		wordIndex = env.NewIdxs()
	}
	return wordIndex
}

func removeBangLine(content string) string {
	if strings.Index(content, "#!") == 0 {
		content = content[strings.Index(content, "\n")+1:]
	}
	return content
}

func checkCodeSignature(content string) int {
	parts := strings.SplitN(content, ";ryesig ", 2)
	content = strings.TrimSpace(parts[0])
	if len(parts) != 2 {
		fmt.Println("\x1b[33m" + "No rye signature found. Exiting." + "\x1b[0m")
		return -1
	}

	signature := parts[1]
	sig := strings.TrimSpace(signature)
	bsig, err := hex.DecodeString(sig)
	if err != nil {
		fmt.Println("\x1b[33m" + "Invalid signature format: " + err.Error() + "\x1b[0m")
		return -2
	}

	// ba8eaa125ee3c8abfc98d8b2b7e5d900bfec0073b701e5c3ca9187a39508b2f1827ba5f0904227678bf33446abbca8bf6a3a5333815920741eb475582a4c31dd privk
	puk, err := hex.DecodeString("827ba5f0904227678bf33446abbca8bf6a3a5333815920741eb475582a4c31dd") // pubk
	if err != nil {
		fmt.Println("\x1b[33m" + "Invalid public key format: " + err.Error() + "\x1b[0m")
		return -2
	}

	bbpuk := ed25519.PublicKey(puk)
	if !ed25519.Verify(bbpuk, []byte(content), bsig) {
		fmt.Println("\x1b[33m" + "Rye signature is not valid! Exiting." + "\x1b[0m")
		return -2
	}
	return 1
}

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
	val, err := parser.ParseAndGetValue(input, nil)

	if err != nil {
		// Create a proper error object with the error message instead of just printing
		errStr := err.Error()
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

	inp1 := strings.TrimSpace(input)
	if len(inp1) == 0 || strings.Index("{", inp1) != 0 {
		input = "{ " + input + " }"
	}

	parser := newParser()

	wordIndexMutex.Lock()
	wordIndex = ps.Idx
	val, err := parser.ParseAndGetValue(input, nil)
	ps.Idx = wordIndex
	wordIndexMutex.Unlock()

	if err != nil {
		var bu strings.Builder
		bu.WriteString("In file " + util.TermBold(ps.ScriptPath) + "\n")
		errStr := err.Error()
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
	if len(word) == 1 || word == "<<" || word == "<-" || word == "<~" || word == ">=" || word == "<=" || word == "//" || word == ".." || word == "." || word == "|" {
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

func newParser() *Parser { // TODO -- add string eaddress path url time
	// Create a PEG parser
	parser, _ := NewParser(`
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
	OPARROWS        <-  "<<" / "<~" / "<-" / ">=" / "<=" / "//"
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
	return parser
}

func InspectNode(v Any) {
	if v != nil {
		fmt.Println(v.(env.Object).Inspect(*wordIndex))
	}
}
