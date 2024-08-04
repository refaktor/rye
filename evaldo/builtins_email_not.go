//go:build no_email
// +build no_email

package evaldo

import (
	"github.com/refaktor/rye/env"
)

var Builtins_email = map[string]*env.Builtin{}
