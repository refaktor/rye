//go:build !wasm
// +build !sm

package evaldo

import (
	"github.com/refaktor/rye/env"
)

// JavaScript interop functions for Rye WASM
var Builtins_js_interop = map[string]*env.Builtin{
}
