//go:build b_no_io
// +build b_no_io

package evaldo

import (
	"rye/env"
)

var Builtins_io = map[string]*env.Builtin{}
