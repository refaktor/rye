//go:build no_ssh
// +build no_ssh

package evaldo

import (
	"github.com/refaktor/rye/env"
)

var Builtins_ssh = map[string]*env.Builtin{}
