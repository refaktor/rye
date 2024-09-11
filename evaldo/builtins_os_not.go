//go:build no_os
// +build no_os

package evaldo

import (
	"github.com/refaktor/rye/env"
)

var Builtins_os = map[string]*env.Builtin{}
