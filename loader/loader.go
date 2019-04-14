// loader.go
package loader

import (
	"fmt"
	"strconv"

	"Rejy_go_v1/env"

	. "github.com/yhirose/go-peg"
)

var genv = *env.NewIdxs()

func GetIdxs() *env.Idxs {
	return &genv
}

func LoadString(input string) (env.Block, *env.Idxs) {
	parser := newParser()
	val, _ := parser.ParseAndGetValue(input, nil)
	//InspectNode(val)
	return val.(env.Block), &genv
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
	return env.Block{*ser}, nil
}

func parseNumber(v *Values, d Any) (Any, error) {
	val, er := strconv.ParseInt(v.Token(), 10, 64)
	return env.Integer{val}, er
}

func parseString(v *Values, d Any) (Any, error) {
	return env.String{v.Token()[1 : len(v.Token())-1]}, nil
}

func parseWord(v *Values, d Any) (Any, error) {
	idx := genv.IndexWord(v.Token())
	return env.Word{idx}, nil
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
	idx := genv.IndexWord(word[:len(word)-1])
	return env.Setword{idx}, nil
}

func parseOpword(v *Values, d Any) (Any, error) {
	//fmt.Println("OPWORD:" + v.Token())
	word := v.Token()
	idx := genv.IndexWord(word[1:])
	return env.Opword{idx}, nil
}

func parsePipeword(v *Values, d Any) (Any, error) {
	//fmt.Println("OPWORD:" + v.Token())
	word := v.Token()
	idx := genv.IndexWord(word[1:])
	return env.Pipeword{idx}, nil
}

func newParser() *Parser {
	// TODO -- add string eaddress path url time
	// Create a PEG parser
	parser, _ := NewParser(`
    BLOCK       	<-  "{" SPACES SERIES* "}"
    SERIES          <-  (STRING / NUMBER / COMMA / VOID / SETWORD / OPWORD / PIPEWORD / WORD / BLOCK) SPACES
    WORD           	<-  LETTER LETTERORNUM* 
	SETWORD    		<-  LETTER LETTERORNUM* ":"
	PIPEWORD   		<-  "|" LETTER LETTERORNUM*
	OPWORD    		<-  "." LETTER LETTERORNUM*
	STRING			<-  '"' STRINGCHAR* '"'
	SPACES			<-  SPACE+
	COMMA			<-  ","
	VOID				<-  "_"
	LETTERORNUM		<-  < [a-zA-Z0-9] >
	LETTER  			<-  < [a-zA-Z] >
    NUMBER           <-  < [0-9]+ >
	SPACE			<-  < [ \t\r\n] >
	STRINGCHAR		<-  < !'"' . >
`)

	//%whitespace      <-  [ \t\r\n]*
	//%word			<-  [a-zA-Z]+
	g := parser.Grammar
	g["BLOCK"].Action = parseBlock
	g["WORD"].Action = parseWord
	g["COMMA"].Action = parseComma
	g["VOID"].Action = parseVoid
	g["SETWORD"].Action = parseSetword
	g["OPWORD"].Action = parseOpword
	g["PIPEWORD"].Action = parsePipeword
	g["NUMBER"].Action = parseNumber
	g["STRING"].Action = parseString
	/* g["SERIES"].Action = func(v *Values, d Any) (Any, error) {
		return v, nil
	}*/
	return parser
}

func InspectNode(v Any) {
	if v != nil {
		fmt.Println(v.(env.Object).Inspect(genv))
	}
}
