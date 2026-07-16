//go:build wasm
// +build wasm

package console

import (
	"github.com/refaktor/rye/env"
)

// OfferDebuggingOptions is a stub for WASM builds - debugging options are not available
func OfferDebuggingOptions(es *env.ProgramState, genv *env.Idxs, tag string) {
	// No-op for WASM build - debugging options not supported in browser environment
}
