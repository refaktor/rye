//go:build no_smtpd
// +build no_smtpd

package evaldo

import (
	"github.com/refaktor/rye/env"
)

var Builtins_smtpd = map[string]*env.Builtin{}
