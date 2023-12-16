//go:build !b_mail
// +build !b_mail

package evaldo

import "github.com/refaktor/rye/env"

var Builtins_mail = map[string]*env.Builtin{}
