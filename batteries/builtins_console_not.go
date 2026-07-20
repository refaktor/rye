//go:build b_norepl || wasm || js

package batteries

import "github.com/refaktor/rye/env"

var Builtins_console = map[string]*env.Builtin{}
