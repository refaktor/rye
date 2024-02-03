package env

import (
	"fmt"
	"strconv"
)

// Gen -- generic functions dictionary

type Gen struct {
	dict map[int]map[int]Object
}

func NewGen() *Gen {
	var e Gen
	e.dict = make(map[int]map[int]Object)
	return &e
}

func (e *Gen) Print(idxs Idxs) {
	fmt.Print("<Gen Dict: ")
	for k, v := range e.dict {
		fmt.Print(" <Kind: " + strconv.FormatInt(int64(k), 10))
		for k1, v1 := range v {
			fmt.Print(" " + strconv.FormatInt(int64(k1), 10) + ": " + v1.Inspect(idxs) + " ")
		}
		fmt.Print(" >")
	}
	fmt.Println(">")
}

func (e *Gen) Get(kind int, word int) (Object, bool) {
	obj, exists := e.dict[kind][word]
	// since here is no parent ... this lookup could be faster maybe ... that parent lookup was taking a lot of time
	// in Env
	return obj, exists
}

func (e *Gen) Set(kind int, word int, val Object) Object {
	if e.dict[kind] == nil {
		e.dict[kind] = make(map[int]Object)
	}
	e.dict[kind][word] = val
	return val
}
