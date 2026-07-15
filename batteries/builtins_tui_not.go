//go:build no_tui || wasm || js
// +build no_tui wasm js

package batteries

import (
	"github.com/refaktor/rye/env"
)

var Builtins_tui = map[string]*env.Builtin{}
