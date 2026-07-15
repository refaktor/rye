//go:build no_email
// +build no_email

package batteries

import (
	"github.com/refaktor/rye/env"
)

var Builtins_email = map[string]*env.Builtin{}
