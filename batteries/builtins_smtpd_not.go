//go:build no_smtpd
// +build no_smtpd

package batteries

import (
	"github.com/refaktor/rye/env"
)

var Builtins_smtpd = map[string]*env.Builtin{}
