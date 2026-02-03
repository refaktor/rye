//go:build no_tui
// +build no_tui

package evaldo

import (
	"github.com/refaktor/rye/env"
)

var Builtins_tui = map[string]*env.Builtin{}
