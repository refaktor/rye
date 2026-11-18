package env

// ForkForGoroutine creates execution-local state, not a snapshot of the heap.
// Contexts, symbols, generic dispatch and runtime services remain shared. Their
// synchronization and the ownership of mutable language values are separate
// concerns; this constructor does not make concurrent context mutation safe.
// Call only from the goroutine that owns ps, before publishing the child.
func (ps *ProgramState) ForkForGoroutine() *ProgramState {
	return &ProgramState{
		Ctx:          ps.Ctx,
		PCtx:         ps.PCtx,
		Idx:          ps.Idx,
		Gen:          ps.Gen,
		Args:         make([]int, 6),
		ScriptPath:   ps.ScriptPath,
		WorkingPath:  ps.WorkingPath,
		AllowMod:     ps.AllowMod,
		LiveObj:      ps.LiveObj,
		Dialect:      ps.Dialect,
		Embedded:     ps.Embedded,
		EmbeddedFS:   ps.EmbeddedFS,
		EffectMode:   ps.EffectMode,
		MaxCallDepth: ps.MaxCallDepth,
		MaxOps:       ps.MaxOps,
		// Preserve consumed budget at spawn, but subsequent accounting is local.
		// MaxOps is not an aggregate budget across all descendants.
		OpsCount:       ps.OpsCount,
		GetHistoryLast: ps.GetHistoryLast,
		BlockLine:      -1,
	}
}
