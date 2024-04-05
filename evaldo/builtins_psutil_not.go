//go:build !b_psutil
// +build !b_psutil

package evaldo

import (
	"github.com/refaktor/rye/env"
)

var Builtins_ps = map[string]*env.Builtin{}
