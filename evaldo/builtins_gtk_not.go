//go:build !b_gtk
// +build !b_gtk

package evaldo

import (
	"rye/env"
)

var Builtins_gtk = map[string]*env.Builtin{}
