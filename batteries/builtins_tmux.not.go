//go:build no_os || b_wasm

package batteries

import (
	"github.com/refaktor/rye/env"
)

var Builtins_tmux = map[string]*env.Builtin{}
