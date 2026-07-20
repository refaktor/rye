//go:build no_pipes
// +build no_pipes

package batteries

import (
	"github.com/refaktor/rye/env"
)

var Builtins_pipes = map[string]*env.Builtin{}
