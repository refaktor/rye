//go:build no_mail
// +build no_mail

package evaldo

import "github.com/refaktor/rye/env"

var Builtins_mail = map[string]*env.Builtin{}
