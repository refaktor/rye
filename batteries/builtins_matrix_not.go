//go:build no_vector
// +build no_vector

package batteries

import "github.com/refaktor/rye/env"

// Builtins_matrix is empty when the no_vector build tag is active.
// This removes the govector dependency.
var Builtins_matrix = map[string]*env.Builtin{}
