//go:build !b_email
// +build !b_email

package evaldo

import (
	"github.com/refaktor/rye/env"
)

var Builtins_email = map[string]*env.Builtin{}
