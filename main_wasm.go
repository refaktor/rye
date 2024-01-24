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
	"github.com/refaktor/rye/util"
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

var ES *env.ProgramState

var CODE []any

var prevResult env.Object

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

var (
	jsCallback  js.Value
	jsCallback2 js.Value
)

func sendMessageToJS(message string) {
	jsCallback.Invoke(message)
}

func sendLineToJS(line string) {
	jsCallback2.Invoke(line)
}

func main() {

	c := make(chan util.KeyEvent)

	ml := util.NewMicroLiner(c, sendMessageToJS, sendLineToJS)

	js.Global().Set("RyeEvalString", js.FuncOf(RyeEvalString))

	js.Global().Set("RyeEvalString2", js.FuncOf(RyeEvalString))

	js.Global().Set("RyeEvalShellLine", js.FuncOf(RyeEvalShellLine))

	js.Global().Set("InitRyeShell", js.FuncOf(InitRyeShell))

	js.Global().Set("SendKeypress", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			cc := util.KeyEvent{args[0].String(), args[1].Int(), args[2].Bool(), args[3].Bool(), args[4].Bool()}
			c <- cc
		}
		return nil
	}))

	// Get the JavaScript function to call back
	jsCallback = js.Global().Get("receiveMessageFromGo")
	jsCallback2 = js.Global().Get("receiveLineFromGo")

	ml.MicroPrompt("x> ", "", 0)

	/* for {
		key := <-c
		// Process the keypress and then send a message back to JavaScript
		response := key
		if key == "A" {
			response = "\x1B[1;3;31mA\x1B[0m"
		}
		sendMessageToJS(response)
	} */

}

func InitRyeShell(this js.Value, args []js.Value) any {
	subc := false
	block, genv := loader.LoadString(" ", false)
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
		evaldo.MaybeDisplayFailureOrErrorWASM(es, genv, sendMessageToJS)

		ES = es

		prevResult = env.Void{}

	case env.Error:
		fmt.Println(val.Message)
		return "Error"
	}
	return "Other"
}

func RyeEvalShellLine(this js.Value, args []js.Value) any {
	sig := false
	subc := false

	code := args[0].String()
	//fmt.Println("RYE EVAL STRING")
	//fmt.Println(code)

	//util.PrintHeader()
	//defer profile.Start(profile.CPUProfile).Stop()
	if ES == nil {
		return "Not initialized"
	}

	es := ES
	fmt.Print("code")
	fmt.Print(code)
	fmt.Print("code")
	block, genv := loader.LoadString(" "+code+" ", sig)
	switch val := block.(type) {
	case env.Block:
		es = env.AddToProgramState(es, val.Series, genv)
		evaldo.RegisterBuiltins(es)
		contrib.RegisterBuiltins(es, &evaldo.BuiltinNames)

		if subc {
			ctx := es.Ctx
			es.Ctx = env.NewEnv(ctx)
		}

		evaldo.EvalBlockInj(es, prevResult, true)
		evaldo.MaybeDisplayFailureOrErrorWASM(es, genv, sendMessageToJS)

		prevResult = es.Res

		if !es.ErrorFlag && es.Res != nil {
			// prevResult = es.Res
			// TEMP - make conditional
			// print the result
			sendMessageToJS("\033[38;5;37m" + es.Res.Inspect(*genv) + "\x1b[0m")
		}

		es.ReturnFlag = false
		es.ErrorFlag = false
		es.FailureFlag = false
		return ""
	case env.Error:
		fmt.Println(val.Message)
		return "Error"
	}
	return "Other"
}

func RyeEvalString(this js.Value, args []js.Value) any {
	sig := false
	subc := false

	code := args[0].String()
	//fmt.Println("RYE EVAL STRING")
	// fmt.Println(code)

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
