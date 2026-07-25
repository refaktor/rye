//go:build no_cli
// +build no_cli

package batteries

import (
	"github.com/refaktor/rye/env"
)

var Builtins_cli = map[string]*env.Builtin{}
