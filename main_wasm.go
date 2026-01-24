//go:build wasm
// +build wasm

package main

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"
	"syscall/js"

	"github.com/refaktor/rye/contrib"
	"github.com/refaktor/rye/env"
	"github.com/refaktor/rye/evaldo"
	"github.com/refaktor/rye/loader"
	"github.com/refaktor/rye/term"
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

var PREV_LINES string

var prevResult env.Object

var ml *term.MLState

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
	jsCallback        js.Value
	jsCallback2       js.Value
	jsCallBack_noterm js.Value
)

func sendMessageToJS(message string) {
	jsCallback.Invoke(message)
}
func sendMessageToJSNL(message string) {
	jsCallback.Invoke(message + "\n")
}

func sendMessageToJS_NOTERM(message string) {
	jsCallBack_noterm.Invoke(message)
}

// browserConsoleLog sends a message to the browser's console.log
func browserConsoleLog(message string) {
	js.Global().Get("console").Call("log", message)
}

func sendLineToJS(line string) string {
	ret := jsCallback2.Invoke(line)
	return ret.String()
}

// BrowserConsoleWriter is a custom io.Writer that writes to the browser console
type BrowserConsoleWriter struct{}

// Write implements the io.Writer interface
func (w *BrowserConsoleWriter) Write(p []byte) (n int, err error) {
	// Convert bytes to string and log to browser console
	message := string(p)
	// Remove trailing newlines for cleaner console output
	message = strings.TrimSuffix(message, "\n")
	// Send to browser console without prefix for cleaner output
	browserConsoleLog(message)
	return len(p), nil
}

func main() {
	// Override standard log output to redirect to browser console
	log.SetOutput(&BrowserConsoleWriter{})
	// Set log flags to not include date/time prefix
	log.SetFlags(0)

	c := make(chan term.KeyEvent)

	ml = term.NewMicroLiner(c, sendMessageToJS, sendLineToJS)
	ml.SetWasmMode(true) // Enable WASM mode to avoid xterm.js scrolling issues

	// Initialize the key event channel in the term package
	// term.SetSB(sendMessageToJSNL)
	term.InitKeyEventChannel(c)

	// Set up the completer function for tab completion
	ml.SetCompleter(func(line string, mode int) (c []string) {
		// Get suggestions based on the current context
		suggestions := make([]string, 0)
		var wordpart string
		spacePos := strings.LastIndex(line, " ")
		var prefix string

		if spacePos < 0 {
			wordpart = line
			prefix = ""
		} else {
			wordpart = strings.TrimSpace(line[spacePos:])
			prefix = line[0:spacePos] + " "
			if wordpart == "" { // we are probably 1 space after last word
				return
			}
		}

		// Mode 0: Get all words from the index
		// Mode 1: Get words from the current context
		if ES != nil {
			switch mode {
			case 0:
				for i := 0; i < ES.Idx.GetWordCount(); i++ {
					if strings.HasPrefix(ES.Idx.GetWord(i), strings.ToLower(wordpart)) {
						c = append(c, prefix+ES.Idx.GetWord(i))
						suggestions = append(suggestions, ES.Idx.GetWord(i))
					} else if strings.HasPrefix("."+ES.Idx.GetWord(i), strings.ToLower(wordpart)) {
						c = append(c, prefix+"."+ES.Idx.GetWord(i))
						suggestions = append(suggestions, ES.Idx.GetWord(i))
					} else if strings.HasPrefix("|"+ES.Idx.GetWord(i), strings.ToLower(wordpart)) {
						c = append(c, prefix+"|"+ES.Idx.GetWord(i))
						suggestions = append(suggestions, ES.Idx.GetWord(i))
					}
				}
			case 1:
				if ES.Ctx != nil {
					for key := range ES.Ctx.GetState() {
						if strings.HasPrefix(ES.Idx.GetWord(key), strings.ToLower(wordpart)) {
							c = append(c, prefix+ES.Idx.GetWord(key))
							suggestions = append(suggestions, ES.Idx.GetWord(key))
						} else if strings.HasPrefix("."+ES.Idx.GetWord(key), strings.ToLower(wordpart)) {
							c = append(c, prefix+"."+ES.Idx.GetWord(key))
							suggestions = append(suggestions, ES.Idx.GetWord(key))
						} else if strings.HasPrefix("|"+ES.Idx.GetWord(key), strings.ToLower(wordpart)) {
							c = append(c, prefix+"|"+ES.Idx.GetWord(key))
							suggestions = append(suggestions, ES.Idx.GetWord(key))
						}
					}
				}
			}
		}

		// Display suggestions below the current line
		if len(suggestions) > 0 {
			// Log suggestions for debugging
			browserConsoleLog("Tab completion suggestions: " + strings.Join(suggestions, ", "))

			// Display suggestions in the terminal
			// First, send a newline
			sendMessageToJS("\n")

			// Display the suggestions in magenta using ANSI color codes directly
			// \033[35m is the ANSI code for magenta, \033[0m resets the color

			sendMessageToJSNL("\033[35m" + strings.Join(suggestions, " ") + "\033[0m")

			// Move back up to the input line using direct ANSI escape sequences
			// \033[1A moves cursor up 1 line, \033[2K clears the line
			sendMessageToJS("\033[1A\033[1A\033[2K")
			term.CurUp(1)
		}

		return
	})

	js.Global().Set("RyeEvalString", js.FuncOf(RyeEvalString))

	js.Global().Set("RyeEvalStringNoTerm", js.FuncOf(RyeEvalStringNoTerm))

	js.Global().Set("RyeEvalShellLine", js.FuncOf(RyeEvalShellLine))

	js.Global().Set("InitRyeShell", js.FuncOf(InitRyeShell))

	js.Global().Set("SetTerminalSize", js.FuncOf(SetTerminalSize))

	// Add a function to explicitly log to browser console
	js.Global().Set("LogToBrowserConsole", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			browserConsoleLog(args[0].String())
		}
		return nil
	}))

	js.Global().Set("SendKeypress", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			cc := term.NewKeyEvent(args[0].String(), args[1].Int(), args[2].Bool(), args[3].Bool(), args[4].Bool())
			c <- cc
		}
		return nil
	}))

	// Get the JavaScript function to call back
	jsCallback = js.Global().Get("receiveMessageFromGo")
	jsCallback2 = js.Global().Get("receiveLineFromGo")
	jsCallBack_noterm = js.Global().Get("receiveLineFromGo_noterm")

	// Redirect fmt.Println and log.Println output to browser console
	// This is done by creating custom writers that write to both xterm.js and browser console
	/* r, w, _ := os.Pipe()
	os.Stdout = w

	go func() {
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			text := scanner.Text()
			// Send to xterm.js console
			sendMessageToJSNL(text)
			// Also send to browser console
			browserConsoleLog(text)
		}
	}() */

	ctx := context.Background()

	ml.MicroPrompt("x> ", "", 0, ctx)

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

func SetTerminalSize(this js.Value, args []js.Value) any {
	term.SetTerminalColumns(args[0].Int())
	ml.SetColumns(term.GetTerminalColumns())
	return "Ok"
}

func InitRyeShell(this js.Value, args []js.Value) any {
	// subc := false
	// fmt.Println("INITIALIZATION")
	ps := env.NewProgramStateNEW()
	evaldo.RegisterBuiltins(ps)
	contrib.RegisterBuiltins(ps, &evaldo.BuiltinNames)
	ctx := ps.Ctx
	ps.Ctx = env.NewEnv(ctx)
	ES = ps
	evaldo.ShowResults = true
	/* bloc	k := loader.LoadString(" ", false)
	switch val := block.(type) {
	case env.Block:

		if subc {
		}

		evaldo.EvalBlock(es)
		evaldo.MaybeDisplayFailureOrErrorWASM(es, genv, sendMessageToJS)

		ES = es

		prevResult = env.Void{}

	case env.Error:
		fmt.Println(val.Message)
		return "Error"
	}
	return "Other"*/
	return "Initalized"
}

func RyeEvalShellLine(this js.Value, args []js.Value) any {
	sig := false
	subc := false

	evaldo.ShowResults = false
	code := args[0].String()
	multiline := len(code) > 1 && code[len(code)-1:] == " "

	comment := regexp.MustCompile(`\s*;`)
	codes := comment.Split(code, 2) //--- just very temporary solution for some comments in repl. Later should probably be part of loader ... maybe?
	code1 := strings.Trim(codes[0], "\t")
	if multiline {
		PREV_LINES += code1
		return "next line"
	}

	code1 = PREV_LINES + "\n" + code1

	PREV_LINES = ""

	if ES == nil {
		return "Error: Rye is not initialized"
	}

	ps := ES
	block := loader.LoadStringNEW(" "+code1+" ", sig, ps)
	switch val := block.(type) {
	case env.Block:

		ps = env.AddToProgramState(ps, val.Series, ps.Idx)

		if subc {
			ctx := ps.Ctx
			ps.Ctx = env.NewEnv(ctx)
		}

		evaldo.EvalBlockInj(ps, prevResult, true)
		evaldo.MaybeDisplayFailureOrErrorWASM(ps, ps.Idx, sendMessageToJSNL, "(Invoked by: Eval console line)")

		prevResult = ps.Res

		if !ps.ErrorFlag && ps.Res != nil {
			sendMessageToJS("\033[38;5;37m" + ps.Res.Inspect(*ps.Idx) + "\x1b[0m\n")
		}

		ps.ReturnFlag = false
		ps.ErrorFlag = false
		ps.FailureFlag = false

		ml.AppendHistory(code)

		return ""

	case env.Error:
		fmt.Println("\033[31mParsing error: " + val.Message + "\033[0m")
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

	block, genv := loader.LoadStringNoPEG(code, sig)
	switch val := block.(type) {
	case env.Block:
		es := env.NewProgramState(val.Series, genv)
		evaldo.RegisterBuiltins(es)
		contrib.RegisterBuiltins(es, &evaldo.BuiltinNames)

		if subc {
			ctx := es.Ctx
			es.Ctx = env.NewEnv(ctx)
		}

		evaldo.EvalBlock(es)
		evaldo.MaybeDisplayFailureOrError(es, genv, "rye eval string wasm")
		return es.Res.Print(*es.Idx)
	case env.Error:
		// Check if the result is an error
		fmt.Println("\033[31mParsing error: " + val.Message + "\033[0m")
		// r.fullCode = ""
		//return ""
		//fmt.Println(val.Message)
		return "Error"
	}
	return "Other"
}

func RyeEvalStringNoTerm(this js.Value, args []js.Value) any {
	sig := false
	subc := false

	code := args[0].String()

	if ES == nil {
		return "Error: Rye is not initialized"
	}

	ps := ES
	block := loader.LoadStringNEW(" "+code+" ", sig, ps)
	switch val := block.(type) {
	case env.Block:

		ps = env.AddToProgramState(ps, val.Series, ps.Idx)

		if subc {
			ctx := ps.Ctx
			ps.Ctx = env.NewEnv(ctx)
		}

		evaldo.EvalBlockInj(ps, prevResult, true)
		evaldo.MaybeDisplayFailureOrErrorWASM(ps, ps.Idx, sendMessageToJSNL, "(Invoked by: Eval console line)")

		prevResult = ps.Res

		//if !ps.ErrorFlag && ps.Res != nil {
		// 	sendMessageToJS("\033[38;5;37m" + ps.Res.Inspect(*ps.Idx) + "\x1b[0m\n")
		// }

		ps.ReturnFlag = false
		ps.ErrorFlag = false
		ps.FailureFlag = false

		// ml.AppendHistory(code)

		return ps.Res.Inspect(*ps.Idx)

	case env.Error:
		fmt.Println("\033[31mParsing error: " + val.Message + "\033[0m")
		return "Error"
	}
	return "Other"
}
