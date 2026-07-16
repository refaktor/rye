package evaldo

import "github.com/refaktor/rye/env"

// CheckForFailureWithBuiltin is an exported wrapper around checkForFailureWithBuiltin,
// exposed for use by the batteries package.
func CheckForFailureWithBuiltin(bi env.Builtin, ps *env.ProgramState, n int) bool {
	return checkForFailureWithBuiltin(bi, ps, n)
}

// TryHandleFailure is an exported wrapper around tryHandleFailure,
// exposed for use by the batteries package.
func TryHandleFailure(ps *env.ProgramState) bool {
	return tryHandleFailure(ps)
}

// ReturnContextToPool is an exported wrapper around returnContextToPool,
// exposed for use by the batteries package.
func ReturnContextToPool(fnCtx *env.RyeCtx, fromPool bool) {
	returnContextToPool(fnCtx, fromPool)
}

// FindWordValue is an exported wrapper around findWordValue,
// exposed for use by the batteries package.
func FindWordValue(ps *env.ProgramState, word1 env.Object) (bool, env.Object, *env.RyeCtx) {
	return findWordValue(ps, word1)
}

// Trace is an exported wrapper around trace,
// exposed for use by the batteries package.
func Trace(s string) { trace(s) }

// GreaterThanNew is an exported wrapper around greaterThanNew,
// exposed for use by the batteries package.
func GreaterThanNew(arg0 env.Object, arg1 env.Object) bool {
	return greaterThanNew(arg0, arg1)
}

// LesserThanNew is an exported wrapper around lesserThanNew,
// exposed for use by the batteries package.
func LesserThanNew(arg0 env.Object, arg1 env.Object) bool {
	return lesserThanNew(arg0, arg1)
}
