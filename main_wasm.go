//go:build wasm
// +build wasm

package main

import (
	"fmt"
	"syscall/js"

	"github.com/refaktor/rye/contrib"
	"github.com/refaktor/rye/env"
	"github.com/refaktor/rye/evaldo"
	"github.com/refaktor/rye/loader"
)

type TagType int
type RjType int
type Series []any

type anyword struct {
	kind RjType
	idx  int
}

type node struct {
	kind  RjType
	value any
}

var CODE []any

//
// main function. Dispatches to appropriate mode function
//

func main1() {
	evaldo.ShowResults = true
	// main_rye_string("print $Hello world$", false, false)
}

//
// main for awk like functionality with rye language
//

func main() {
	c := make(chan struct{}, 0)
	js.Global().Set("RyeEvalString", js.FuncOf(RyeEvalString))
	<-c
}

func RyeEvalString(this js.Value, args []js.Value) any {
	sig := false
	subc := true

	code := args[0].String()

	//util.PrintHeader()
	//defer profile.Start(profile.CPUProfile).Stop()

	block, genv := loader.LoadString(code, sig)
	switch val := block.(type) {
	case env.Block:
		es := env.NewProgramState(block.(env.Block).Series, genv)
		evaldo.RegisterBuiltins(es)
		contrib.RegisterBuiltins(es, &evaldo.BuiltinNames)

		if subc {
			ctx := es.Ctx
			es.Ctx = env.NewEnv(ctx)
		}

		evaldo.EvalBlock(es)
		evaldo.MaybeDisplayFailureOrError(es, genv)
		return es.Res.Probe(*es.Idx)
	case env.Error:
		fmt.Println(val.Message)
		return "Error"
	}
	return "Other"
}
