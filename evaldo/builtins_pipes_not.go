//go:build no_pipes
// +build no_pipes

package evaldo

import (
	"github.com/refaktor/rye/env"
)

var Builtins_os = map[string]*env.Builtin{}
