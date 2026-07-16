//go:build b_norepl || wasm || js

package console

import (
	"fmt"

	"github.com/refaktor/rye/env"
)

func DoRyeRepl(es *env.ProgramState, dialect string, showResults bool, localHist bool, histFile string) {
	fmt.Println("REPL not available in this build.")
}

func DoRyeDualRepl(leftPs *env.ProgramState, rightPs *env.ProgramState, dialect string, showResults bool) {
	fmt.Println("Dual REPL not available in this build.")
}
