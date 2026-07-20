//go:build no_vector
// +build no_vector

package batteries

import "github.com/refaktor/rye/env"

// Builtins_vector is empty when the no_vector build tag is active.
// This removes the govector dependency.
var Builtins_vector = map[string]*env.Builtin{}
