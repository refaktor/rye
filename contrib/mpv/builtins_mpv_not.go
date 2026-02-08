//go:build !b_mpv
// +build !b_mpv

package mpv

import (
	"github.com/refaktor/rye/env"
)

var Builtins_mpv = map[string]*env.Builtin{}
