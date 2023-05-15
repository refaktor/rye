//go:build !b_bcrypt
// +build !b_bcrypt

package evaldo

import (
	"rye/env"
)

var Builtins_bcrypt = map[string]*env.Builtin{}
