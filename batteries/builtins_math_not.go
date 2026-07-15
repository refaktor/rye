//go:build no_vector
// +build no_vector

package batteries

import (
	"github.com/refaktor/rye/env"
	"github.com/refaktor/rye/evaldo"
)

// Builtins_math is empty when the no_vector build tag is active.
// This removes the govector and primes dependencies.
var Builtins_math = map[string]*env.Builtin{}

// DialectMath is a no-op stub used by repl.go when no_vector is active.
func DialectMath(ps *env.ProgramState, arg0 env.Object) env.Object {
	return evaldo.MakeBuiltinError(ps, "math dialect unavailable (built with no_vector)", "DialectMath")
}
