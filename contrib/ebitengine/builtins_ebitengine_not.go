//go:build !b_ebitengine
// +build !b_ebitengine

package ebitengine

import (
	"github.com/refaktor/rye/env"
)

var Builtins_ebitengine = map[string]*env.Builtin{}
