//go:build no_baseio
// +build no_baseio

package evaldo

import "github.com/refaktor/rye/env"

// builtins_baseio_not.go — stub used when the no_baseio build tag is active.
// The OS / file / shell / stdin / args builtins are omitted so that embedding
// use-cases (embed.New) can drop the heavy terminal / OS dependencies entirely.
// RegisterBaseIOBuiltins is defined in builtins.go and calls RegisterBuiltins2
// with this empty map, making it a no-op automatically.

var builtins_baseio = map[string]*env.Builtin{}
