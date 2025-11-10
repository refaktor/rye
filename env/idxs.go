package env

import (
	"fmt"
	"strconv"
	"strings"
)

type Idxs struct {
	words1 []string
	words2 map[string]int
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
	"Context",
	"Dict",
	"List",
	"Date",
	"CPath",
	"Xword",
	"EXword",
	"Table",
	"Email",
	"Kind",
	"Kindword",
	"Converter",
	"Time",
	"TableRowType",
	"Decimal",
	"Vector",
	"OpCPath",
	"PipeCPath",
	"Modword",
	"LModword",
	"Boolean",
	"VarBuiltin",
	"CurriedCaller",
	"Complex",
	"Markdown",
	"PersistentCtx",
	"LocationNode",
	"Flagword",
	"PersistentTable",
}

func (e *Idxs) IndexWord(w string) int {
	idx, ok := e.words2[w]
	if ok {
		return idx
	} else {
		e.words1 = append(e.words1, w)
		e.words2[w] = len(e.words1) - 1
		return len(e.words1) - 1
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
	for i := range e.words1 {
		fmt.Print(strconv.FormatInt(int64(i), 10) + ": " + e.words1[i] + " ")
	}
	fmt.Println(">")
}

func (e Idxs) GetWordCount() int {
	return len(e.words1)
}

func NewIdxs() *Idxs {
	var e Idxs
	e.words1 = []string{""}
	e.words2 = make(map[string]int)

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
