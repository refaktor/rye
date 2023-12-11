//go:build !b_raylib
// +build !b_raylib

package evaldo

import (
	"rye/env"
)

var Builtins_raylib = map[string]*env.Builtin{}
