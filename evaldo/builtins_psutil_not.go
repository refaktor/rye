//go:build !b_devops
// +build !b_devops

package evaldo

import (
	"github.com/refaktor/rye/env"
)

var Builtins_devops = map[string]*env.Builtin{}
