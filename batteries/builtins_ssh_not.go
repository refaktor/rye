//go:build !add_ssh
// +build !add_ssh

package batteries

import (
	"github.com/refaktor/rye/env"
)

var Builtins_ssh = map[string]*env.Builtin{}
