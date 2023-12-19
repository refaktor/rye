//go:build !b_contrib
// +build !b_contrib

package contrib

import (
	"github.com/refaktor/rye/env"
)

var Builtins_contrib = map[string]*env.Builtin{}

func RegisterBuiltins(ps *env.ProgramState, builtinNames *map[string]int) {
}
