package evaldo

import (
	"sync"

	"github.com/refaktor/rye/env"
)

// fastFunctionContextPool is a sync.Pool for reusing RyeCtx objects specifically for function calls
var fastFunctionContextPool = sync.Pool{
	New: func() interface{} {
		return env.NewEnv(nil)
	},
}

// FastCallFunction is an optimized version of Rye0_CallFunction for the common case
// where all arguments are already available. It skips the complex context determination,
// argument evaluation, and error checking that makes Rye0_CallFunction slower.
//
// This function should be used when:
// 1. All arguments are already available (no need to evaluate from program state)
// 2. The function has a fixed number of arguments (typically small, like 0-3)
// 3. No complex context path resolution is needed
func FastCallFunction(fn env.Function, ps *env.ProgramState, args []env.Object, ctx *env.RyeCtx) *env.ProgramState {
	// Fast path for determining context
	var fnCtx *env.RyeCtx

	// Get a context from the pool instead of creating a new one
	fnCtx = fastFunctionContextPool.Get().(*env.RyeCtx)

	// Set up the context based on function properties
	if fn.Pure {
		fnCtx.Parent = ps.PCtx
	} else if fn.Ctx != nil {
		if fn.InCtx {
			// Put the pooled context back since we're using the function's context directly
			fastFunctionContextPool.Put(fnCtx)
			fnCtx = fn.Ctx
		} else {
			fnCtx.Parent = fn.Ctx
		}
	} else {
		// Use the current context as parent
		fnCtx.Parent = ps.Ctx
	}

	// Set arguments directly without evaluation
	for i, arg := range args {
		if i < fn.Spec.Series.Len() {
			// Get the word index directly
			index := fn.Spec.Series.Get(i).(env.Word).Index
			fnCtx.Set(index, arg)
		}
	}

	// Get a program state from the pool
	psX := functionCallPool.Get().(*env.ProgramState)
	resetProgramState(psX, fn.Body.Series, ps.Idx)

	// Set up the program state
	psX.Ctx = fnCtx
	psX.PCtx = ps.PCtx
	psX.Gen = ps.Gen
	psX.Dialect = ps.Dialect
	psX.WorkingPath = ps.WorkingPath
	psX.ScriptPath = ps.ScriptPath
	psX.LiveObj = ps.LiveObj
	psX.Embedded = ps.Embedded

	// Evaluate the function body
	if len(args) > 0 {
		Rye0_EvalBlockInj(psX, args[0], true)
	} else {
		EvalBlock(psX)
	}

	// Process the result
	if psX.ForcedResult != nil {
		ps.Res = psX.ForcedResult
		psX.ForcedResult = nil
	}

	// Put the program state back in the pool
	functionCallPool.Put(psX)

	// Put the context back in the pool if it's not the function's own context
	if !fn.InCtx {
		// Clear the context before returning it to the pool
		fnCtx.Parent = nil
		for k := range fnCtx.GetState() {
			delete(fnCtx.GetState(), k)
		}
		fastFunctionContextPool.Put(fnCtx)
	}

	ps.ReturnFlag = false
	return ps
}

// FastCallFunctionWithArgs is an optimized version of Rye0_CallFunctionWithArgs
// It's similar to FastCallFunction but follows the signature of Rye0_CallFunctionWithArgs
func FastCallFunctionWithArgs(fn env.Function, ps *env.ProgramState, ctx *env.RyeCtx, args ...env.Object) *env.ProgramState {
	return FastCallFunction(fn, ps, args, ctx)
}

// Modified version of Rye0_CallFunction that uses FastCallFunction for common cases
func Rye0_CallFunction_Optimized(fn env.Function, ps *env.ProgramState, arg0 env.Object, toLeft bool, ctx *env.RyeCtx) *env.ProgramState {
	// Fast path: If we have all arguments already and no special handling is needed
	if arg0 != nil && fn.Argsn == 1 {
		return FastCallFunction(fn, ps, []env.Object{arg0}, ctx)
	}

	// Fast path: If no arguments are needed
	if fn.Argsn == 0 {
		return FastCallFunction(fn, ps, nil, ctx)
	}

	// Fallback to the original implementation for complex cases
	return Rye0_CallFunction(fn, ps, arg0, toLeft, ctx)
}
