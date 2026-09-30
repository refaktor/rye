//go:build no_os || b_wasm
// +build no_os b_wasm

package baseio

import (
	"github.com/refaktor/rye/env"
)

var Builtins_os = map[string]*env.Builtin{}
