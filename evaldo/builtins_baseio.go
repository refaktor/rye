//go:build !no_baseio
// +build !no_baseio

package evaldo

// builtins_baseio.go — OS / file / shell / stdin / args builtins.
//
// These are kept separate from the pure-computation base builtins so that
// embedding use-cases (embed.New / RegisterBaseBuiltins) can opt-out of OS
// access entirely.  The full runner registers both sets via RegisterBuiltins.

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/refaktor/rye/env"
	"github.com/refaktor/rye/loader"
	"github.com/refaktor/rye/util"

	"golang.org/x/sync/errgroup"
	goterm "golang.org/x/term"
)

// builtins_baseio groups all builtins that touch the operating-system
// boundary: file I/O, shell commands, stdin, os.Exit, os.Args, and
// capture-stdout.  They are registered by RegisterBaseIOBuiltins which is
// called from the full RegisterBuiltins but NOT from RegisterBaseBuiltins.
var builtins_baseio = map[string]*env.Builtin{

	// -------------------------------------------------------------------------
	// Save / persist state
	// -------------------------------------------------------------------------

	// Tests:
	// equal  { save\current |type? } 'integer
	// Args:
	// * None
	// Returns:
	// * Integer 1 on success
	"save\\current": {
		Argsn: 0,
		Doc:   "Saves current state of the program to a file.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) (res env.Object) {
			s := ps.Dump()
			fileName := fmt.Sprintf("console_%s.rye", time.Now().Format("060102_150405"))

			err := os.WriteFile(fileName, []byte(s), 0600)
			if err != nil {
				ps.FailureFlag = true
				return MakeBuiltinError(ps, fmt.Sprintf("error writing state: %s", err.Error()), "save\\state")
			}
			fmt.Println("State current context to \033[1m" + fileName + "\033[0m.")
			return *env.NewInteger(1)
		},
	},

	// Tests:
	// ; equal  { save\current\secure |type? } 'integer
	// Args:
	// * None
	// Returns:
	// * Integer 1 on success
	"save\\current\\secure": {
		Argsn: 0,
		Doc:   "Saves current state of the program to a file with password protection.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) (res env.Object) {
			s := ps.Dump()
			fileName := fmt.Sprintf("console_%s.rye.enc", time.Now().Format("060102_150405"))

			fmt.Print("Enter Password: ")
			bytePassword, err := goterm.ReadPassword(int(os.Stdin.Fd()))
			if err != nil {
				panic(err)
			}
			password := string(bytePassword)

			util.SaveSecure(s, fileName, password)
			fmt.Println("State current context to \033[1m" + fileName + "\033[0m.")
			return *env.NewInteger(1)
		},
	},

	// -------------------------------------------------------------------------
	// File import / load (URI-based)
	// -------------------------------------------------------------------------

	// Tests:
	// ; import file://test.rye  ; imports and executes test.rye
	// Args:
	// * uri: URI of the file to import and execute
	// Returns:
	// * result of executing the imported file
	"file-uri//Import": { // **
		Argsn: 1,
		Doc:   "Imports a file, loads and does it from script local path.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Uri:
				block_, script_ := LoadScriptLocalFile(ps, s1)
				ps.Res = EvaluateLoadedValue(ps, block_, script_, false)
				ps.ScriptPath = script_
				return ps.Res
			default:
				return MakeArgError(ps, 1, []env.Type{env.UriType}, "import")
			}
		},
	},

	// Tests:
	// ; import\live file://test.rye  ; imports, executes, and watches test.rye for changes
	// Args:
	// * uri: URI of the file to import, execute, and watch for changes
	// Returns:
	// * result of executing the imported file
	"file-uri//Import\\live": { // **
		Argsn: 1,
		Doc:   "Imports a file, loads and does it from script local path.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Uri:
				block_, script_ := LoadScriptLocalFile(ps, s1)
				ps.Res = EvaluateLoadedValue(ps, block_, script_, false)
				ps.LiveObj.Add(s1.GetPath()) // add to watcher
				ps.ScriptPath = script_
				return ps.Res
			default:
				return MakeArgError(ps, 1, []env.Type{env.UriType}, "import\\live")
			}
		},
	},

	// Tests:
	// ; equal  { load " 1 2 3 " |third } 3
	// ; equal  { load "{ 1 2 3 }" |first |third } 3
	// Args:
	// * source: String containing Rye code or URI of file to load
	// Returns:
	// * Block containing the parsed Rye values
	"file-uri//Load": { // **
		Argsn: 1,
		Doc:   "Loads a file URI into Rye values.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Uri:
				var str string
				fileIdx, _ := ps.Idx.GetIndex("file")
				if s1.Scheme.Index == fileIdx {
					b, err := os.ReadFile(s1.GetPath())
					if err != nil {
						return makeError(ps, err.Error())
					}
					str = string(b)
				}
				scrip := ps.ScriptPath
				ps.ScriptPath = s1.GetPath()
				block := loader.LoadString(str, false, ps)
				ps.ScriptPath = scrip
				return block
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.UriType}, "file-uri//Load")
			}
		},
	},

	// TODO -- refactor load variants so they use common function LoadString and LoadFile

	// Tests:
	// ; load\mod file://modifiable.rye  ; loads file with word modification allowed
	// Args:
	// * source: URI of file to load with modification allowed
	// Returns:
	// * Block containing the parsed Rye values
	"load\\mod\\file": { // **
		Argsn: 1,
		Doc:   "Loads a file URI into Rye values, allowing modification of words during load.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Uri:
				var str string
				fileIdx, _ := ps.Idx.GetIndex("file")
				if s1.Scheme.Index == fileIdx {
					b, err := os.ReadFile(s1.GetPath())
					if err != nil {
						return makeError(ps, err.Error())
					}
					str = string(b)
				}
				scrip := ps.ScriptPath
				ps.AllowMod = true
				ps.ScriptPath = s1.GetPath()
				block := loader.LoadString(str, false, ps)
				ps.AllowMod = false
				ps.ScriptPath = scrip
				return block
			default:
				ps.FailureFlag = true
				return env.NewError("Must be a file URI")
			}
		},
	},

	// Tests:
	// ; load\live file://watched.rye  ; loads and watches file for changes
	// Args:
	// * source: URI of file to load with modification allowed and file watching
	// Returns:
	// * Block containing the parsed Rye values
	"load\\live": { // **
		Argsn: 1,
		Doc:   "Loads a file URI into Rye values, allows modification of words, and watches for changes.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Uri:
				var str string
				fileIdx, _ := ps.Idx.GetIndex("file")
				if s1.Scheme.Index == fileIdx {
					b, err := os.ReadFile(s1.GetPath())
					ps.LiveObj.Add(s1.GetPath()) // add to watcher
					if err != nil {
						return makeError(ps, err.Error())
					}
					str = string(b)
				}
				scrip := ps.ScriptPath
				ps.AllowMod = true
				ps.ScriptPath = s1.GetPath()
				block := loader.LoadString(str, false, ps)
				ps.AllowMod = false
				ps.ScriptPath = scrip
				return block
			default:
				ps.FailureFlag = true
				return env.NewError("Must be a file URI")
			}
		},
	},

	// -------------------------------------------------------------------------
	// Process / shell
	// -------------------------------------------------------------------------

	// Tests:
	// equal { scmd `echo "hello"` } 0
	// equal { scmd `exit 1` } 1
	// equal { scmd `exit 42` } 42
	"scmd": {
		Argsn: 1,
		Doc:   "Execute a shell command and return its exit status code.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s0 := arg0.(type) {
			case env.String:
				r := exec.Command("sh", "-c", s0.Value) //nolint: gosec
				r.Stdout = os.Stdout
				r.Stderr = os.Stderr

				err := r.Run()
				if err != nil {
					if exitError, ok := err.(*exec.ExitError); ok {
						return *env.NewInteger(int64(exitError.ExitCode()))
					}
					fmt.Println(err)
					return *env.NewInteger(-1)
				}
				return *env.NewInteger(0)
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "scmd")
			}
		},
	},

	// Tests:
	// equal { scmd\capture `echo "hello"` } "hello\n"
	"scmd\\capture": {
		Argsn: 1,
		Doc:   "Execute a shell command and capture the output, return it as string",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s0 := arg0.(type) {
			case env.String:
				r := exec.Command("sh", "-c", s0.Value) //nolint: gosec
				var outb, errb bytes.Buffer
				r.Stdout = &outb
				r.Stderr = &errb

				err := r.Run()
				if err != nil {
					if errb.Len() > 0 {
						ps.FailureFlag = true
						return env.NewError("Command failed: " + errb.String())
					}
					if _, ok := err.(*exec.ExitError); ok && outb.Len() > 0 {
						return *env.NewString(outb.String())
					}
					ps.FailureFlag = true
					return env.NewError("Command failed: " + err.Error())
				}
				return *env.NewString(outb.String())
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "scmd\\capture")
			}
		},
	},

	// Tests:
	// ; equal { exit 0 } ...
	"exit": { // **
		Argsn: 1,
		Doc:   "Exits the process with the given integer status code (or 0 for any non-integer).",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			util.BeforeExit()
			switch code := arg0.(type) {
			case env.Integer:
				os.Exit(int(code.Value))
				return nil
			default:
				fmt.Println(code.Inspect(*ps.Idx))
				os.Exit(0)
				return nil
			}
		},
	},

	// -------------------------------------------------------------------------
	// Rye-itself — args / history (requires os.Args / process context)
	// -------------------------------------------------------------------------

	// Deprecated
	"Rye-itself//args?": {
		Argsn: 0,
		Doc:   "Returns command line arguments as a block of parsed values. Each argument is converted to appropriate type (integer, float, or string).",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return ryeItselfArgsParsed(ps)
		},
	},

	"Rye-itself//Args?": {
		Argsn: 0,
		Doc:   "Returns command line arguments as a block of parsed values. Each argument is converted to appropriate type (integer, float, or string).",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return ryeItselfArgsParsed(ps)
		},
	},

	"Rye-itself//Args\\raw?": {
		Argsn: 1,
		Doc:   "Returns raw command line arguments joined as a single string.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			firstArg := ryeFirstArgIdx(ps)
			if len(os.Args) > firstArg {
				return *env.NewString(strings.Join(os.Args[firstArg:], " "))
			}
			return *env.NewString("")
		},
	},

	// Tests:
	// ; equal { rye .history 5 |length? } 5
	// Args:
	// * n: Integer specifying how many history lines to return
	// Returns:
	// * Block of strings containing the last N lines of REPL history
	"Rye-itself//History?": {
		Argsn: 2,
		Doc:   "Returns a block of the last N lines from REPL history.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch n := arg1.(type) {
			case env.Integer:
				if ps.GetHistoryLast == nil {
					return MakeBuiltinError(ps, "History not available (not running in REPL)", "Rye-itself//history")
				}
				lines := ps.GetHistoryLast(int(n.Value))
				objs := make([]env.Object, len(lines))
				for i, line := range lines {
					objs[i] = *env.NewString(line)
				}
				return *env.NewBlock(*env.NewTSeries(objs))
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType}, "Rye-itself//history")
			}
		},
	},

	// -------------------------------------------------------------------------
	// stdout capture
	// -------------------------------------------------------------------------

	// Tests:
	// equal { capture-stdout { print "hello" } } "hello\n"
	"capture-stdout": { // **
		Argsn: 1,
		Doc:   "Executes a block of code while capturing all output to stdout, returning the captured output as a string.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch bloc := arg0.(type) {
			case env.Block:
				old := os.Stdout // keep backup of the real stdout
				r, w, _ := os.Pipe()
				os.Stdout = w

				outC := make(chan string, 1000)
				g := errgroup.Group{}
				g.Go(func() error {
					var buf bytes.Buffer
					_, err := io.Copy(&buf, r)
					if err != nil {
						w.Close()
						os.Stdout = old
						fmt.Println(err.Error())
						return err
					}
					outC <- buf.String()
					return nil
				})

				ser := ps.Ser
				ps.Ser = bloc.Series
				ps.BlockFile = bloc.FileName
				ps.BlockLine = bloc.Line
				Eval(ps)
				ps.Ser = ser

				w.Close()
				os.Stdout = old

				if err := g.Wait(); err != nil {
					return MakeBuiltinError(ps, fmt.Sprintf("Error reading stdout: %v", err), "capture-stdout")
				}
				out := <-outC

				ps.SkipFlag = false
				MaybeDisplayFailureOrError(ps, ps.Idx, "capture-stdout")

				if ps.ErrorFlag {
					return ps.Res
				}
				return *env.NewString(out)
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType}, "capture-stdout")
			}
		},
	},
}

// ---------------------------------------------------------------------------
// Private helpers shared by the args builtins
// ---------------------------------------------------------------------------

func ryeFirstArgIdx(ps *env.ProgramState) int {
	if ps.Embedded {
		return 1
	}
	return 2
}

func ryeItselfArgsParsed(ps *env.ProgramState) env.Object {
	firstArg := ryeFirstArgIdx(ps)
	if firstArg >= len(os.Args) {
		return *env.NewBlock(*env.NewTSeries([]env.Object{}))
	}

	args := os.Args[firstArg:]
	lst := make([]env.Object, len(args))

	intRe := regexp.MustCompile("^[+-]?[0-9]+$")
	floatRe := regexp.MustCompile("^[+-]?[0-9]*\\.[0-9]+$")

	for i, arg := range args {
		if intRe.MatchString(arg) {
			if num, err := strconv.ParseInt(arg, 10, 64); err == nil {
				lst[i] = *env.NewInteger(num)
				continue
			}
		}
		if floatRe.MatchString(arg) {
			if num, err := strconv.ParseFloat(arg, 64); err == nil {
				lst[i] = *env.NewDecimal(num)
				continue
			}
		}
		lst[i] = *env.NewString(arg)
	}
	return *env.NewBlock(*env.NewTSeries(lst))
}
