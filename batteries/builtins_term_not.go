//go:build b_no_term
// +build b_no_term

package batteries

import (
	"github.com/refaktor/rye/env"
)

var Builtins_term = map[string]*env.Builtin{}
