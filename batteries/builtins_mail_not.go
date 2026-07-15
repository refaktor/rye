//go:build no_mail
// +build no_mail

package batteries

import "github.com/refaktor/rye/env"

var Builtins_mail = map[string]*env.Builtin{}
