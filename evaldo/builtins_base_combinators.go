package evaldo

import (
	"github.com/refaktor/rye/env"
)

var builtins_combinators = map[string]*env.Builtin{

	// COMBINATORS

	// Tests:
	// equal  { 101 .pass { 202 } } 101
	// equal  { 101 .pass { 202 + 303 } } 101
	// Args:
	// * value: Any value that will be passed to the block and returned
	// * block: Block of code to execute with the value injected
	// Returns:
	// * The original value, regardless of what the block returns
	"pass": { // **
		Argsn: 2,
		Doc:   "Accepts a value and a block. It does the block, with value injected, and returns (passes on) the initial value.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch bloc := arg1.(type) {
			case env.Block:
				res := arg0
				ser := ps.Ser
				ps.Ser = bloc.Series
				EvalBlockInjMultiDialect(ps, arg0, true)
				MaybeDisplayFailureOrError(ps, ps.Idx, "pass")
				if ps.ErrorFlag || ps.ReturnFlag {
					ps.Ser = ser
					return ps.Res
				}
				ps.Ser = ser
				return res
			default:
				return MakeArgError(ps, 2, []env.Type{env.BlockType}, "pass")
			}
		},
	},

	// Tests:
	// stdout { wrap { prn "*" } { prn "x" } } "*x*"
	// Args:
	// * wrapper: Block of code to execute before and after the main block
	// * block: Main block of code to execute between wrapper executions
	// Returns:
	// * The result of the main block execution
	"wrap": { // **
		Argsn: 2,
		Doc:   "Executes a wrapper block before and after executing a main block, returning the result of the main block.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch wrap := arg0.(type) {
			case env.Block:
				switch bloc := arg1.(type) {
				case env.Block:
					ser := ps.Ser
					ps.Ser = wrap.Series
					EvalBlockInjMultiDialect(ps, arg0, true)
					MaybeDisplayFailureOrError(ps, ps.Idx, "wrap")
					if ps.ErrorFlag || ps.ReturnFlag {
						ps.Ser = ser
						return ps.Res
					}

					ps.Ser = bloc.Series
					EvalBlockInjMultiDialect(ps, arg0, true)
					MaybeDisplayFailureOrError(ps, ps.Idx, "wrap")
					if ps.ErrorFlag || ps.ReturnFlag {
						ps.Ser = ser
						return ps.Res
					}
					res := ps.Res

					ps.Ser = wrap.Series
					EvalBlockInjMultiDialect(ps, arg0, true)
					MaybeDisplayFailureOrError(ps, ps.Idx, "wrap")
					if ps.ErrorFlag || ps.ReturnFlag {
						ps.Ser = ser
						return ps.Res
					}
					ps.Ser = ser
					return res
				default:
					return MakeArgError(ps, 2, []env.Type{env.BlockType}, "wrap")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType}, "wrap")
			}
		},
	},

	// Tests:
	// equal  { 20 .keep { + 202 } { + 101 } } 222
	// Args:
	// * value: Value to be injected into both blocks
	// * block1: First block whose result will be returned
	// * block2: Second block to execute after the first one
	// Returns:
	// * The result of the first block, ignoring the result of the second block
	"keep": { // **
		Argsn: 3,
		Doc:   "Do the first block, then the second one but return the result of the first one.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch b1 := arg1.(type) {
			case env.Block:
				switch b2 := arg2.(type) {
				case env.Block:
					ser := ps.Ser
					ps.Ser = b1.Series
					EvalBlockInjMultiDialect(ps, arg0, true)
					MaybeDisplayFailureOrError(ps, ps.Idx, "keep")
					if ps.ErrorFlag || ps.ReturnFlag {
						ps.Ser = ser
						return ps.Res
					}
					res := ps.Res
					ps.Ser = b2.Series
					EvalBlockInjMultiDialect(ps, arg0, true)
					MaybeDisplayFailureOrError(ps, ps.Idx, "keep")
					if ps.ErrorFlag || ps.ReturnFlag {
						ps.Ser = ser
						return ps.Res
					}
					ps.Ser = ser
					return res
				default:
					return MakeArgError(ps, 3, []env.Type{env.BlockType}, "keep")
				}
			default:
				return MakeArgError(ps, 2, []env.Type{env.BlockType}, "keep")
			}
		},
	},
}
