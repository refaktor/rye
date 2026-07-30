package evaldo

import "github.com/refaktor/rye/env"

// builtins_baseio_not.go - always-active stub.  OS / file / shell / stdin / args
// builtins have been moved to the baseio package (github.com/refaktor/rye/baseio).
// Call baseio.Register(ps) after RegisterBuiltins(ps) to get full I/O support.
// RegisterBaseIOBuiltins is now a no-op (kept for backward compatibility).

var builtins_baseio = map[string]*env.Builtin{}
