// loader.go
package loader

import (
	"fmt"
	"strconv"

	"rye/env"

	. "github.com/yhirose/go-peg"
)

func trace(x interface{}) {
	//fmt.Print("\x1b[56m")
	//fmt.Print(x)
	//fmt.Println("\x1b[0m")
}

var wordIndex = *env.NewIdxs()

func GetIdxs() *env.Idxs {
	return &wordIndex
}

func LoadString(input string) (env.Block, *env.Idxs) {
	//fmt.Println(input)
	parser := newParser()
	val, _ := parser.ParseAndGetValue(input, nil)
	//InspectNode(val)
	if val != nil {
		return val.(env.Block), &wordIndex
	} else {
		empty1 := make([]env.Object, 0)
		ser := env.NewTSeries(empty1)
		return *env.NewBlock(*ser), &wordIndex
	}
}

func parseBlock(v *Values, d Any) (Any, error) {
	//fmt.Println("** Parse block **")
	//fmt.Println(v.Vs)
	block := make([]env.Object, len(v.Vs)-1)
	//var r env.Object
	for i := 1; i < len(v.Vs); i += 1 {
		obj := v.Vs[i]
		if true { //obj != nil
			//fmt.Println(i)
			//InspectNode(obj)
			block[i-1] = obj.(env.Object)
		}
	}
	//fmt.Print("BLOCK --> ")
	//fmt.Println(block)
	ser := env.NewTSeries(block)
	return env.Block{*ser, 0}, nil
}

func parseBBlock(v *Values, d Any) (Any, error) {
	block := make([]env.Object, len(v.Vs)-1)
	for i := 1; i < len(v.Vs); i += 1 {
		obj := v.Vs[i]
		if true {
			block[i-1] = obj.(env.Object)
		}
	}
	ser := env.NewTSeries(block)
	return env.Block{*ser, 1}, nil
}

func parseGroup(v *Values, d Any) (Any, error) {
	//fmt.Println("** Parse block **")
	//fmt.Println(v.Vs)
	block := make([]env.Object, len(v.Vs)-1)
	//var r env.Object
	for i := 1; i < len(v.Vs); i += 1 {
		obj := v.Vs[i]
		if true { //obj != nil
			//fmt.Println(i)
			//InspectNode(obj)
			block[i-1] = obj.(env.Object)
		}
	}
	//fmt.Print("BLOCK --> ")
	//fmt.Println(block)
	ser := env.NewTSeries(block)
	return env.Block{*ser, 2}, nil
}

func parseNumber(v *Values, d Any) (Any, error) {
	val, er := strconv.ParseInt(v.Token(), 10, 64)
	return env.Integer{val}, er
}

func parseString(v *Values, d Any) (Any, error) {
	return env.String{v.Token()[1 : len(v.Token())-1]}, nil
}

func parseUri(v *Values, d Any) (Any, error) {
	return *env.NewUri(&wordIndex, v.Vs[0].(env.Word), v.Token()), nil // ){v.Vs[0].(env.Word), v.Token()}, nil // TODO let the second part be it's own object that parser returns like path
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
	//d fmt.Println("OPWORD:" + v.Token())
	word := v.Token()
	idx := wordIndex.IndexWord(word[1:])
	return env.Opword{idx}, nil
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

func parseEXword(v *Values, d Any) (Any, error) {
	//fmt.Println("TAGWORD:" + v.Token())
	word := v.Token()
	idx := wordIndex.IndexWord(word[2 : len(word)-1])
	return env.EXword{idx}, nil
}

func parsePipeword(v *Values, d Any) (Any, error) {
	//fmt.Println("OPWORD:" + v.Token())
	word := v.Token()
	idx := wordIndex.IndexWord(word[1:])
	return env.Pipeword{idx}, nil
}

func parseGenword(v *Values, d Any) (Any, error) {
	trace("GENWORD:" + v.Token())
	word := v.Token()
	idx := wordIndex.IndexWord(word)
	return env.Genword{idx}, nil
}

func parseGetword(v *Values, d Any) (Any, error) {
	trace("GETWORD:" + v.Token())
	word := v.Token()
	idx := wordIndex.IndexWord(word[1:])
	return env.Getword{idx}, nil
}

func newParser() *Parser {
	// TODO -- add string eaddress path url time
	// Create a PEG parser
	parser, _ := NewParser(`
	BLOCK       	<-  "{" SPACES SERIES* "}"
	BBLOCK       	<-  "[" SPACES SERIES* "]"
    GROUP       	<-  "(" SPACES SERIES* ")"
    SERIES          <-  (URI / STRING / NUMBER / COMMA / VOID / SETWORD / LSETWORD / OPWORD / PIPEWORD / TAGWORD / EXWORD / CPATH / XWORD / GENWORD / GETWORD / WORD / BLOCK / GROUP / BBLOCK / ARGBLOCK ) SPACES
    ARGBLOCK       	<-  "{" WORD ":" WORD "}"
    WORD           	<-  LETTER LETTERORNUM* / ONECHARWORDS
	GENWORD           	<-  UCLETTER LCLETTERORNUM* 
	SETWORD    		<-  LETTER LETTERORNUM* ":"
	LSETWORD    		<-  ":" LETTER LETTERORNUM*
	GETWORD   		<-  "?" LETTER LETTERORNUM*
	PIPEWORD   		<-  "|" LETTER LETTERORNUM*
	OPWORD    		<-  "." LETTER LETTERORNUM*
	TAGWORD    		<-  "'" LETTER LETTERORNUM*
	XWORD    		<-  "<" LETTER LETTERORNUM* ">"?
	EXWORD    		<-  "</" LETTER LETTERORNUM* ">"?
	STRING			<-  '"' STRINGCHAR* '"'
	SPACES			<-  SPACE+
	URI    		<-  WORD "://" URIPATH*
	CPATH    		<-  WORD ( "/" WORD )+
	COMMA			<-  ","
	VOID				<-  "_"
	ONECHARWORDS		<-  < [<>*+] >
	LETTERORNUM		<-  < [a-zA-Z0-9-?=.\\!_>] >
	URIPATH		<-  < [a-zA-Z0-9-?=.:@/\\!_>] >
	LETTER  			<-  < [a-zA-Z?=^_] >
	UCLETTER  			<-  < [A-Z] >
	LCLETTERORNUM		<-  < [a-z0-9] >
    NUMBER           <-  < [0-9]+ >
	SPACE			<-  < [ \t\r\n] >
	STRINGCHAR		<-  < !'"' . >
`)

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
	g["TAGWORD"].Action = parseTagword
	g["XWORD"].Action = parseXword
	g["EXWORD"].Action = parseEXword
	g["GENWORD"].Action = parseGenword
	g["GETWORD"].Action = parseGetword
	g["NUMBER"].Action = parseNumber
	g["STRING"].Action = parseString
	g["URI"].Action = parseUri
	g["CPATH"].Action = parseCPath
	/* g["SERIES"].Action = func(v *Values, d Any) (Any, error) {
		return v, nil
	}*/
	return parser
}

func InspectNode(v Any) {
	if v != nil {
		fmt.Println(v.(env.Object).Inspect(wordIndex))
	}
}
