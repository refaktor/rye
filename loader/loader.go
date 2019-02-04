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

func parseWord(v *Values, d Any) (Any, error) {
	idx := genv.IndexWord(v.Token())
	return env.Word{idx}, nil
}

func parseSetword(v *Values, d Any) (Any, error) {
	//fmt.Println("SETWORD:" + v.Token())
	word := v.Token()
	idx := genv.IndexWord(word[:len(word)-1])
	return env.Setword{idx}, nil
}

func newParser() *Parser {
	// Create a PEG parser
	parser, _ := NewParser(`
    BLOCK       		<-  "{" SPACES SERIES* "}"
    SERIES           <-  (NUMBER / SETWORD / WORD / BLOCK) SPACES
    WORD           	<-  LETTER LETTERORNUM+ 
	SETWORD    		<-  LETTER LETTERORNUM+ ":"
	SPACES			<-  SPACE+
	LETTERORNUM		<-  < [a-zA-Z0-9] >
	LETTER  			<-  < [a-zA-Z] >
    NUMBER           <-  < [0-9]+ >
	SPACE			<-  < [ \t\r\n] >
`)

	//%whitespace      <-  [ \t\r\n]*
	//%word			<-  [a-zA-Z]+
	g := parser.Grammar
	g["BLOCK"].Action = parseBlock
	g["WORD"].Action = parseWord
	g["SETWORD"].Action = parseSetword
	g["NUMBER"].Action = parseNumber
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
