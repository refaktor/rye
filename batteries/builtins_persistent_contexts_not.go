//go:build no_persistent
// +build no_persistent

package batteries

import (
	"github.com/refaktor/rye/env"
)

// PersistentCtx is a stub for when the no_persistent build tag is active.
// Without badger, persistent contexts are unavailable.
type PersistentCtx struct {
	env.RyeCtx
}

func (pc PersistentCtx) Type() env.Type {
	return env.PersistentContextType
}

// EvalBlockInPersistentCtx is a no-op stub.
func EvalBlockInPersistentCtx(ps *env.ProgramState, pctx *PersistentCtx) {}

var builtins_persistent_contexts = map[string]*env.Builtin{}
