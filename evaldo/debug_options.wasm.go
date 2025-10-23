//go:build wasm
// +build wasm

package evaldo

import (
	"github.com/refaktor/rye/env"
)

// OfferDebuggingOptions is a stub for WASM builds - debugging options are not available
func OfferDebuggingOptions(es *env.ProgramState, genv *env.Idxs, tag string) {
	// No-op for WASM build - debugging options not supported
	// In WASM, we can't offer interactive debugging options like REPL, VSCode, etc.
}

// HandleEnterConsoleOption is a stub for WASM builds
func HandleEnterConsoleOption(es *env.ProgramState, genv *env.Idxs) {
	// No-op for WASM build - REPL console not available
}

// HandleListContextOption is a stub for WASM builds
func HandleListContextOption(es *env.ProgramState, genv *env.Idxs) {
	// No-op for WASM build - interactive context listing not available
}

// HandleVSCodeOption is a stub for WASM builds
func HandleVSCodeOption(fileName string, lineNumber int) {
	// No-op for WASM build - VSCode integration not available
}

// HandleNvimOption is a stub for WASM builds
func HandleNvimOption(fileName string, lineNumber int) {
	// No-op for WASM build - Neovim integration not available
}

// DoRyeRepl is a stub for WASM builds - REPL functionality is not available
func DoRyeRepl(es *env.ProgramState, prompt string, debug bool) {
	// No-op for WASM build - REPL not supported in browser environment
}
