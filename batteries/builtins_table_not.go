//go:build no_table
// +build no_table

package batteries

import (
	"github.com/refaktor/rye/env"
)

// Builtins_table is empty when the no_table build tag is active.
// This removes the excelize (xlsx) dependency.
var Builtins_table = map[string]*env.Builtin{}
