//go:build no_git
// +build no_git

package evaldo

import (
	"github.com/refaktor/rye/env"
)

var Builtins_git = map[string]*env.Builtin{}
