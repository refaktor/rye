//go:build no_devops
// +build no_devops

package evaldo

import (
	"github.com/refaktor/rye/env"
)

var Builtins_devops = map[string]*env.Builtin{}
