//go:build no_bcrypt
// +build no_bcrypt

package batteries

import (
	"github.com/refaktor/rye/env"
)

var Builtins_bcrypt = map[string]*env.Builtin{}
