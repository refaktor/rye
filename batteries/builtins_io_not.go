//go:build no_io
// +build no_io

package batteries

import (
	"github.com/refaktor/rye/env"
)

var Builtins_io = map[string]*env.Builtin{}
