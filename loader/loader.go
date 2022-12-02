// loader.go
package loader

import (
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"rye/env"

	//. "github.com/yhirose/go-peg"
	//. "github.com/CWood1/go-peg"
	. "github.com/refaktor/go-peg"
)

func trace(x interface{}) {
	//fmt.Print("\x1b[56m")
	//fmt.Print(x)
	//fmt.Println("\x1b[0m")
}

var wordIndex *env.Idxs

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
	// fmt.Println(content)
	// fmt.Println(signature)
	// fmt.Println(strings.Index(signature, ";#codesig "))
	sig := strings.TrimSpace(signature)
	bsig, _ := hex.DecodeString(sig)
	// ba8eaa125ee3c8abfc98d8b2b7e5d900bfec0073b701e5c3ca9187a39508b2f1827ba5f0904227678bf33446abbca8bf6a3a5333815920741eb475582a4c31dd privk
	puk, _ := hex.DecodeString("827ba5f0904227678bf33446abbca8bf6a3a5333815920741eb475582a4c31dd") // pubk
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
	if sig {
		signed := checkCodeSignature(input)
		if signed == -1 {
			return *env.NewError(""), wordIndex
		} else if signed == -2 {
			return *env.NewError(""), wordIndex
		}
	}

	input = removeBangLine(input)

	inp1 := strings.TrimSpace(input)
	if strings.Index("{", inp1) != 0 {
		input = "{ " + input + " }"
	}

	parser := newParser()
	val, err := parser.ParseAndGetValue(input, nil)

	if err != nil {
		fmt.Print("\x1b[35;3m")
		fmt.Print(err.Error())
		fmt.Println("\x1b[0m")

		empty1 := make([]env.Object, 0)
		ser := env.NewTSeries(empty1)
		return *env.NewBlock(*ser), wordIndex
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
	return env.Block{*ser, 0}, nil
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
	return env.Block{*ser, 1}, nil
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
	return env.Block{*ser, 2}, nil
}

func parseNumber(v *Values, d Any) (Any, error) {
	val, er := strconv.ParseInt(v.Token(), 10, 64)
	return env.Integer{val}, er
}

func parseDecimal(v *Values, d Any) (Any, error) {
	val, er := strconv.ParseFloat(v.Token(), 64)
	return env.Decimal{val}, er
}

func parseString(v *Values, d Any) (Any, error) {
	str := v.Token()[1 : len(v.Token())-1]
	// turn \n to newlines
	str = strings.Replace(str, "\\n", "\n", -1)
	return env.String{str}, nil
}

func parseUri(v *Values, d Any) (Any, error) {
	// fmt.Println(v.Vs[0])
	return *env.NewUri(wordIndex, v.Vs[0].(env.Word), v.Token()), nil // ){v.Vs[0].(env.Word), v.Token()}, nil // TODO let the second part be it's own object that parser returns like path
}

func parseEmail(v *Values, d Any) (Any, error) {
	return env.Email{v.Token()}, nil
}

func parseFpath(v *Values, d Any) (Any, error) {
	idx := wordIndex.IndexWord("file")
	return *env.NewUri(wordIndex, env.Word{idx}, "file://"+v.Token()[1:]), nil // ){v.Vs[0].(env.Word), v.Token()}, nil // TODO let the second part be it's own object that parser returns like path
}

func parseCPath(v *Values, d Any) (Any, error) {
	switch len(v.Vs) {
	case 2:
		return *env.NewCPath2(v.Vs[0].(env.Word), v.Vs[1].(env.Word)), nil
	case 3:
		return *env.NewCPath3(v.Vs[0].(env.Word), v.Vs[1].(env.Word), v.Vs[2].(env.Word)), nil
	default:
		return *env.NewCPath3(v.Vs[0].(env.Word), v.Vs[1].(env.Word), v.Vs[2].(env.Word)), nil
	}
}

func parseWord(v *Values, d Any) (Any, error) {
	idx := wordIndex.IndexWord(v.Token())
	return env.Word{idx}, nil
}

func parseArgword(v *Values, d Any) (Any, error) {
	return env.Argword{v.Vs[0].(env.Word), v.Vs[1].(env.Word)}, nil
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
	return env.Setword{idx}, nil
}

func parseLSetword(v *Values, d Any) (Any, error) {
	//fmt.Println("SETWORD:" + v.Token())
	word := v.Token()
	idx := wordIndex.IndexWord(word[1:])
	return env.LSetword{idx}, nil
}

func parseOpword(v *Values, d Any) (Any, error) {
	//fmt.Println("OPWORD:" + v.Token())
	word := v.Token()
	force := 0
	var idx int
	if len(word) == 1 || word == "<<" || word == "<-" || word == "<--" {
		// onecharopwords < > + * ... their naming is equal to _< _> _* ...
		idx = wordIndex.IndexWord("_" + word)
	} else {
		if word[len(word)-1:] == "*" {
			force = 1
			word = word[:len(word)-1]
		}
		idx = wordIndex.IndexWord(word[1:])
	}
	return env.Opword{idx, force}, nil
}

func parseTagword(v *Values, d Any) (Any, error) {
	//fmt.Println("TAGWORD:" + v.Token())
	word := v.Token()
	idx := wordIndex.IndexWord(word[1:])
	return env.Tagword{idx}, nil
}

func parseXword(v *Values, d Any) (Any, error) {
	//fmt.Println("TAGWORD:" + v.Token())
	word := v.Token()
	idx := wordIndex.IndexWord(word[1 : len(word)-1])
	return env.Xword{idx}, nil
}

func parseKindword(v *Values, d Any) (Any, error) {
	word := v.Token()
	idx := wordIndex.IndexWord(word[1 : len(word)-1])
	return env.Kindword{idx}, nil
}

func parseEXword(v *Values, d Any) (Any, error) {
	//fmt.Println("TAGWORD:" + v.Token())
	word := v.Token()
	idx := wordIndex.IndexWord(word[2 : len(word)-1])
	return env.EXword{idx}, nil
}

func parsePipeword(v *Values, d Any) (Any, error) {
	//fmt.Println("OPWORD:" + v.Token())
	word := v.Token()
	force := 0
	var idx int
	if word == ">>" || word == "->" || word == "-->" {
		idx = wordIndex.IndexWord("_" + word)
	} else {
		if word[len(word)-1:] == "*" {
			force = 1
			word = word[:len(word)-1]
		}
		idx = wordIndex.IndexWord(word[1:])
	}
	return env.Pipeword{idx, force}, nil
}

func parseOnecharpipe(v *Values, d Any) (Any, error) {
	//fmt.Println("OPWORD:" + v.Token())
	word := v.Token()
	var idx int
	idx = wordIndex.IndexWord("_" + word[1:])
	return env.Pipeword{idx, 0}, nil
}

func parseGenword(v *Values, d Any) (Any, error) {
	trace("GENWORD:" + v.Token())
	word := v.Token()
	idx := wordIndex.IndexWord(strings.ToLower(word))
	return env.Genword{idx}, nil
}

func parseGetword(v *Values, d Any) (Any, error) {
	trace("GETWORD:" + v.Token())
	word := v.Token()
	idx := wordIndex.IndexWord(word[1:])
	return env.Getword{idx}, nil
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
    SERIES     		<-  (COMMENT / URI / EMAIL / STRING / DECIMAL / NUMBER / COMMA / SETWORD / LSETWORD / ONECHARPIPE / PIPEWORD / XWORD / OPWORD / TAGWORD / EXWORD / CPATH / FPATH / KINDWORD / GENWORD / GETWORD / VOID / WORD / BLOCK / GROUP / BBLOCK / ARGBLOCK ) SPACES
    ARGBLOCK       	<-  "{" WORD ":" WORD "}"
    WORD           	<-  LETTER LETTERORNUM*
	GENWORD 		<-  UCLETTER LCLETTERORNUM* 
	SETWORD    		<-  LETTER LETTERORNUM* ":"
	LSETWORD    	<-  ":" LETTER LETTERORNUM*
	GETWORD   		<-  "?" LETTER LETTERORNUM*
	PIPEWORD   		<-  "|" LETTER LETTERORNUM* / PIPEARROWS
	ONECHARPIPE    <-  "|" ONECHARWORDS
	OPWORD    		<-  "." LETTER LETTERORNUM* / OPARROWS / ONECHARWORDS / "[*" LETTERORNUM* 
	TAGWORD    		<-  "'" LETTER LETTERORNUM*
	KINDWORD    		<-  "_(" LETTER LETTERORNUM* ")_"?
	XWORD    		<-  "<" LETTER LETTERORNUM* ">"?
	EXWORD    		<-  "</" LETTER LETTERORNUM* ">"?
	STRING			<-  ('"' STRINGCHAR* '"') / ("$" STRINGCHAR1* "$")
	SPACES			<-  SPACE+
	URI    			<-  WORD "://" URIPATH*
	EMAIL			<-  EMAILPART "@" EMAILPART 
	EMAILPART		<-  < ([a-zA-Z0-9._]+) >
	FPATH 	   		<-  "%" URIPATH*
	CPATH    		<-  WORD ( "/" WORD )+
	ONECHARWORDS	    <-  < [<>*+-=/] >
	PIPEARROWS      <-  ">>" / "-->" / "->"
	OPARROWS        <-  "<<" / "<--" / "<-"
	LETTER  	       	<-  < [a-zA-Z=^(` + "`" + `_] >
	LETTERORNUM		<-  < [a-zA-Z0-9-?=.\\!_+<>\]*()] >
	URIPATH			<-  < [a-zA-Z0-9-?=.:@/\\!_>	()] >
	UCLETTER  		<-  < [A-Z] >
	LCLETTERORNUM  	        <-  < [a-z0-9] >
        NUMBER          	<-  < [0-9]+ >
        DECIMAL         	<-  < [0-9]+.[0-9]+ >
	SPACE			<-  < [ \t\r\n] >
	STRINGCHAR		<-  < !'"' . >
	STRINGCHAR1		<-  < !"$" . >
	COMMA			<-  ","
	VOID				<-  "_"
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
