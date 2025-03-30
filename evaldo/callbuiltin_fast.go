package evaldo

import (
	"github.com/refaktor/rye/env"
)

// FastCallBuiltin is an optimized version of CallBuiltin for the common case
// where all arguments are already available. It skips the argument evaluation,
// error checking, and currying logic that makes CallBuiltin slower.
//
// This function should be used when:
// 1. All arguments are already available (no need to evaluate from program state)
// 2. No currying is needed
// 3. The builtin function has a fixed number of arguments
func FastCallBuiltin(ps *env.ProgramState, bi env.Builtin, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
	// Skip all the argument evaluation, error checking, and currying logic
	// and directly call the builtin function
	return bi.Fn(ps, arg0, arg1, arg2, arg3, arg4)
}
