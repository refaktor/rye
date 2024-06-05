package env

import (
	"fmt"
	"strconv"
	"strings"
)

type Idxs struct {
	words1 [3000]string
	words2 map[string]int
	wordsn int
}

var NativeTypes = [...]string{ // Todo change to BuiltinTypes
	"Block",
	"Integer",
	"Word",
	"Setword",
	"Opword",
	"Pipeword",
	"Builtin",
	"Function",
	"Error",
	"Comma",
	"Void",
	"String",
	"Tagword",
	"Genword",
	"Getword",
	"Argword",
	"Native",
	"Uri",
	"LSetword",
	"Ctx",
	"Dict",
	"List",
	"Date",
	"CPath",
	"Xword",
	"EXword",
	"Spreadsheet",
	"Email",
	"Kind",
	"Kindword",
	"Converter",
	"Time",
	"SpreadsheetRowType",
	"Decimal",
	"Vector",
	"OpCPath",
	"PipeCPath",
	"Modword",
	"LModword",
}

func (e *Idxs) IndexWord(w string) int {
	idx, ok := e.words2[w]
	if ok {
		return idx
	} else {
		e.words1[e.wordsn] = w
		e.words2[w] = e.wordsn
		e.wordsn += 1
		return e.wordsn - 1
	}
}

func (e *Idxs) GetIndex(w string) (int, bool) {
	idx, ok := e.words2[w]
	if ok {
		return idx, true
	}
	return 0, false
}

func (e Idxs) GetWord(i int) string {
	if i < 0 {
		return "isolate!" // TODO -- behaviour aroung isolates ... so that it prevents generic function lookup
	}
	return e.words1[i]
}

func (e Idxs) Print() {
	fmt.Print("<IDXS: ")
	for i := 0; i < e.wordsn; i++ {
		fmt.Print(strconv.FormatInt(int64(i), 10) + ": " + e.words1[i] + " ")
	}
	fmt.Println(">")
}

func (e Idxs) GetWordCount() int {
	return e.wordsn
}

func NewIdxs() *Idxs {
	var e Idxs
	e.words2 = make(map[string]int)
	e.wordsn = 1

	/*
		BlockType    Type = 1
		IntegerType  Type = 2
		WordType     Type = 3
		SetwordType  Type = 4
		OpwordType   Type = 5
		PipewordType Type = 6
		BuiltinType  Type = 7
		FunctionType Type = 8
		ErrorType    Type = 9
		CommaType    Type = 10
		VoidType     Type = 11
		StringType   Type = 12
		TagwordType  Type = 13
	*/

	// register words for builtin kinds, which the value objects should return on GetKind()

	for _, value := range NativeTypes {
		e.IndexWord(strings.ToLower(value))
	}
	return &e
}
