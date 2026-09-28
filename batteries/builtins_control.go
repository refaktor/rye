package batteries

import (
	"github.com/refaktor/rye/env"
	"github.com/refaktor/rye/evaldo"
)

// Builtins_control contains control flow helpers.
var Builtins_control = map[string]*env.Builtin{
	"case": {
		Argsn: 1,
		Doc:   "Evaluates condition expressions followed by action blocks. Returns the result of the first block whose condition is boolean true, or void if none match. Non-boolean conditions fail.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			pairs, ok := arg0.(env.Block)
			if !ok {
				return evaldo.MakeArgError(ps, 1, []env.Type{env.BlockType}, "case")
			}

			ser := ps.Ser
			defer func() { ps.Ser = ser }()
			ps.Ser = pairs.Series
			for ps.Ser.Pos() < ps.Ser.Len() {
				// Let the evaluator consume the whole expression, including arguments.
				evaldo.EvalExpression_CollectArg(ps, false, false)
				if ps.ErrorFlag || ps.FailureFlag || ps.ReturnFlag {
					return ps.Res
				}
				condition, ok := ps.Res.(env.Boolean)
				if !ok {
					return evaldo.MakeBuiltinError(ps, "Condition must evaluate to boolean", "case")
				}
				if ps.Ser.Pos() >= ps.Ser.Len() {
					return evaldo.MakeBuiltinError(ps, "Expected block after condition", "case")
				}
				action, ok := ps.Ser.Pop().(env.Block)
				if !ok {
					return evaldo.MakeBuiltinError(ps, "Expected block after condition", "case")
				}
				if condition.Value {
					ps.Ser = action.Series
					ps.Res = *env.NewVoid()
					evaldo.EvalBlockInj(ps, nil, false)
					return ps.Res
				}
				evaldo.MaybeAcceptComma(ps, nil, false)
			}
			return *env.NewVoid()
		},
	},
}
