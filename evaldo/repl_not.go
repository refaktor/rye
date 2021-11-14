// +build b_norepl

package evaldo

import (
	"fmt"
	//	"log"
	"os"
	"path/filepath"

	//	"regexp"
	"rye/env"
	//	"rye/loader"
	//	"strings"
	//	"github.com/peterh/liner"
)

var (
	history_fn = filepath.Join(os.TempDir(), ".rye_repl_history")
	names      = []string{"add", "join", "return", "fn", "fail", "if"}
)

type ShellEd struct {
	CurrObj env.Function
	Pause   bool
	Askfor  []string
	Mode    string
	Return  env.Object
}

func genPrompt(shellEd *ShellEd, line string) (string, string) {
	return "{ Rye } ", ""
}

func maybeDoShedCommands(line string, es *env.ProgramState, shellEd *ShellEd) {
}

func maybeDoShedCommandsBlk(line string, es *env.ProgramState, block *env.Block, shed_pause *bool) {
}

// terminal commands from buger/goterm

// Clear screen
func Clear() {
	fmt.Print("\033[2J")
}

// Move cursor to given position
func MoveCursor(x int, y int) {
	fmt.Printf("\033[%d;%dH", y, x)
}

// Move cursor up relative the current position
func MoveCursorUp(bias int) {
	fmt.Printf("\033[%dA", bias)
}

// Move cursor down relative the current position
func MoveCursorDown(bias int) {
	fmt.Printf("\033[%dB", bias)
}

// Move cursor forward relative the current position
func MoveCursorForward(bias int) {
	fmt.Printf("\033[%dC", bias)
}

// Move cursor backward relative the current position
func MoveCursorBackward(bias int) {
	fmt.Printf("\033[%dD", bias)
}

//

func DoRyeRepl(es *env.ProgramState) {

}

func MaybeDisplayFailureOrError(es *env.ProgramState, genv *env.Idxs) {
	if es.FailureFlag {
		fmt.Println("\x1b[33m" + "Failure" + "\x1b[0m")
	}

	if es.ErrorFlag {
		fmt.Println("\x1b[35;3m" + es.Res.Probe(*genv))
		switch err := es.Res.(type) {
		case env.Error:

			fmt.Println(err.CodeBlock.Probe(*genv))
			fmt.Println("Error not pointer so bug. #temp")
		case *env.Error:
			fmt.Println("At location:")
			fmt.Println(err.CodeBlock.Probe(*genv))
		}
		fmt.Println("\x1b[0m")
	}
}
