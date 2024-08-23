//go:build !add_psutil
// +build !add_psutil

package evaldo

import (
	"github.com/refaktor/rye/env"
)

var Builtins_devops = map[string]*env.Builtin{}
