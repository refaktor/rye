//go:build no_io
// +build no_io

package evaldo

import (
	"github.com/refaktor/rye/env"
)

var Builtins_io = map[string]*env.Builtin{}
