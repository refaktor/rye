//go:build tinygo
// +build tinygo

package batteries

import (
	"github.com/refaktor/rye/env"
	"github.com/refaktor/rye/evaldo"
)

// RegisterBuiltinsTinyGo is a simplified version of RegisterBuiltins for TinyGo
func RegisterBuiltinsTinyGo(ps *env.ProgramState) {
	evaldo.BuiltinNames = make(map[string]int)

	// Register only the base builtins
	evaldo.RegisterBuiltins2(builtins, ps, "base")
	evaldo.RegisterBuiltins2(builtins_boolean, ps, "base")
	evaldo.RegisterBuiltins2(builtins_numbers, ps, "base")
	evaldo.RegisterBuiltins2(builtins_string, ps, "base")
	evaldo.RegisterBuiltins2(builtins_collection, ps, "base")
	evaldo.RegisterBuiltins2(builtins_conditionals, ps, "base")
	evaldo.RegisterBuiltins2(builtins_printing, ps, "base")
	evaldo.RegisterBuiltins2(builtins_types, ps, "base")
	evaldo.RegisterBuiltins2(builtins_functions, ps, "base")
	evaldo.RegisterBuiltins2(builtins_apply, ps, "base")
	evaldo.RegisterBuiltins2(builtins_iteration, ps, "base")
	evaldo.RegisterBuiltins2(builtins_contexts, ps, "base")
	evaldo.RegisterBuiltins2(builtins_time, ps, "base")
}
