package main

import (
	"rye/env"
	"rye/evaldo"
	//	"rye/loader"
	//	"syscall/js"
)

type TagType int
type RjType int
type Series []interface{}

type anyword struct {
	kind RjType
	idx  int
}

type node struct {
	kind  RjType
	value interface{}
}

var ps *env.ProgramState
/* 
func doBlock(code string) env.Object {
	block, genv := loader.LoadString(code, false)
	env.AddToProgramState(ps, block.Series, genv)
	evaldo.EvalBlock(ps)
	return ps.Res
    }*/

/* func doBlockWrap() js.Func {
	function := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) != 1 {
			return "Invalid no of arguments passed"
		}
		code := args[0].String()
		res := doBlock(code)
		println(res.Inspect(*ps.Idx))
		return js.ValueOf(res.Inspect(*ps.Idx))
	})
	return function
} */ 

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


func main() {
	//println("Hello from Rye in webasm!")
	// create the initial interpreter and program state
	// block, genv := loader.LoadString("{ }", false)
	body := []env.Object{env.Integer{234}}
	ps = env.NewProgramState(*env.NewTSeries(body), GetIdxs())
	evaldo.RegisterBuiltins(ps)
	evaldo.EvalBlock(ps)

	///// js.Global().Set("doBlock", doBlockWrap())

	// wait so that the wasm go program doesn't exit
	<-make(chan bool)
}
