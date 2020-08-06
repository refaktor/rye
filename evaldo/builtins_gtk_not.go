// +build !b_gtk

package evaldo

import "C"

import (
	"rye/env"
)

var Builtins_gtk = map[string]*env.Builtin{}
