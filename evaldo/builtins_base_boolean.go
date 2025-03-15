package evaldo

import (
	"github.com/refaktor/rye/env"
	"github.com/refaktor/rye/util"
	// JM 20230825	"github.com/refaktor/rye/term"
)

var builtins_boolean = map[string]*env.Builtin{
	//
	// ##### Boolean logic #####  "Functions that work with true and false values."
	//
	// Tests:
	// equal { true } true
	// equal { true |type? } 'boolean
	// Args:
	// * none
	// Returns:
	// * boolean true value
	"true": {
		Argsn: 0,
		Doc:   "Returns a boolean true value.",
		Pure:  true,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewBoolean(true)
		},
	},

	// Tests:
	// equal { false } false
	// equal { false |type? } 'boolean
	// Args:
	// * none
	// Returns:
	// * boolean false value
	"false": {
		Argsn: 0,
		Doc:   "Returns a boolean false value.",
		Pure:  true,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewBoolean(false)
		},
	},

	// Tests:
	// equal { not true } 0
	// equal { not false } 1
	// Args:
	// * value: Any value to be negated
	// Returns:
	// * the original value (used in pipeline operations)
	"_|": {
		Argsn: 1,
		Doc:   "Pipeline operator that passes the value through unchanged (used with 'not' and other operations).",
		Pure:  true,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return arg0
		},
	},

	// Tests:
	// equal { not true } false
	// equal { not false } true
	// error { not 0 }
	// error { not 5 }
	// Args:
	// * value: Boolean value to be negated
	// Returns:
	// * boolean false if the input is true, true if the input is false
	"not": {
		Argsn: 1,
		Doc:   "Performs logical negation on boolean values only.",
		Pure:  true,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch b := arg0.(type) {
			case env.Boolean:
				return *env.NewBoolean(!b.Value)
			default:
				return MakeArgError(env1, 1, []env.Type{env.BooleanType}, "not")
			}
		},
	},

	// Tests:
	// equal { true .and true } true
	// equal { false .and true } false
	// equal { true .and false } false
	// equal { false .and false } false
	// equal { 3 .and 5 } 1  ; bitwise 011 AND 101 = 001
	// Args:
	// * value1: First value (boolean or integer)
	// * value2: Second value (boolean or integer)
	// Returns:
	// * boolean result of logical AND operation if both inputs are booleans, otherwise integer result of bitwise AND
	"and": {
		Argsn: 2,
		Doc:   "Performs a logical AND operation between two boolean values, or a bitwise AND operation between two integer values.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Boolean:
				switch s2 := arg1.(type) {
				case env.Boolean:
					return *env.NewBoolean(s1.Value && s2.Value)
				default:
					return MakeArgError(ps, 2, []env.Type{env.BooleanType}, "and")
				}
			case env.Integer:
				switch s2 := arg1.(type) {
				case env.Integer:
					return *env.NewInteger(s1.Value & s2.Value)
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "and")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.BooleanType, env.IntegerType}, "and")
			}
		},
	},

	// Tests:
	// equal { true .or true } true
	// equal { false .or true } true
	// equal { true .or false } true
	// equal { false .or false } false
	// equal { 3 .or 5 } 7  ; bitwise 011 OR 101 = 111
	// Args:
	// * value1: First value (boolean or integer)
	// * value2: Second value (boolean or integer)
	// Returns:
	// * boolean result of logical OR operation if both inputs are booleans, otherwise integer result of bitwise OR
	"or": {
		Argsn: 2,
		Doc:   "Performs a logical OR operation between two boolean values, or a bitwise OR operation between two integer values.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Boolean:
				switch s2 := arg1.(type) {
				case env.Boolean:
					return *env.NewBoolean(s1.Value || s2.Value)
				default:
					return MakeArgError(ps, 2, []env.Type{env.BooleanType}, "or")
				}
			case env.Integer:
				switch s2 := arg1.(type) {
				case env.Integer:
					return *env.NewInteger(s1.Value | s2.Value)
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "or")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.BooleanType, env.IntegerType}, "or")
			}
		},
	},

	// Tests:
	// equal { true .xor true } false
	// equal { false .xor true } true
	// equal { true .xor false } true
	// equal { false .xor false } false
	// equal { 3 .xor 5 } 6  ; bitwise 011 XOR 101 = 110
	// Args:
	// * value1: First value (boolean or integer)
	// * value2: Second value (boolean or integer)
	// Returns:
	// * boolean result of logical XOR operation if both inputs are booleans, otherwise integer result of bitwise XOR
	"xor": {
		Argsn: 2,
		Doc:   "Performs a logical XOR (exclusive OR) operation between two boolean values, or a bitwise XOR operation between two integer values.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Boolean:
				switch s2 := arg1.(type) {
				case env.Boolean:
					return *env.NewBoolean((s1.Value || s2.Value) && !(s1.Value && s2.Value))
				default:
					return MakeArgError(ps, 2, []env.Type{env.BooleanType}, "xor")
				}
			case env.Integer:
				switch s2 := arg1.(type) {
				case env.Integer:
					return *env.NewInteger(s1.Value ^ s2.Value)
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "xor")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.BooleanType, env.IntegerType}, "xor")
			}
		},
	},

	// Tests:
	// equal { all { 1 2 3 } } 3
	// equal { all { 1 0 3 } } 0
	// equal { all { true true true } } 1
	// Args:
	// * block: Block of expressions to evaluate
	// Returns:
	// * the last value if all expressions are truthy, otherwise the first falsy value encountered
	"all": { // **
		Argsn: 1,
		Doc:   "Evaluates all expressions in a block and returns the last value if all are truthy, or the first falsy value encountered.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch bloc := arg0.(type) {
			case env.Block:
				ser := ps.Ser
				ps.Ser = bloc.Series
				for ps.Ser.Pos() < ps.Ser.Len() {
					EvalExpression2(ps, false)
					if !util.IsTruthy(ps.Res) {
						break
					}
				}
				ps.Ser = ser
				return ps.Res
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType}, "all")
			}
		},
	},

	// Tests:
	// equal { any { 1 2 3 } } 1
	// equal { any { 0 1 3 } } 1
	// equal { any { 0 0 3 } } 3
	// equal { any { 0 0 0 } } 0
	// Args:
	// * block: Block of expressions to evaluate
	// Returns:
	// * the first truthy value encountered, or the last value if none are truthy
	"any": { // **
		Argsn: 1,
		Doc:   "Evaluates expressions in a block until a truthy value is found and returns it, or returns the last value if none are truthy.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch bloc := arg0.(type) {
			case env.Block:
				ser := ps.Ser
				ps.Ser = bloc.Series
				for ps.Ser.Pos() < ps.Ser.Len() {
					EvalExpression2(ps, false)
					if util.IsTruthy(ps.Res) {
						break
					}
				}
				ps.Ser = ser
				return ps.Res
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType}, "any")
			}
		},
	},

	// Tests:
	// equal { any\with 10 { + 10 , * 10 } } 20
	// ; equal { any\with -10 { + 10 , * 10 } } -100  ; TODO -- fix halting issue
	// Args:
	// * value: Value to be used in the expressions
	// * block: Block of expressions to evaluate with the provided value
	// Returns:
	// * the first truthy result of applying an expression to the value, or the last result if none are truthy
	"any\\with": { // TODO-FIXME error handling, halts on multiple expressions
		Argsn: 2,
		Doc:   "Applies each expression in the block to the provided value until a truthy result is found and returns it.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch bloc := arg1.(type) {
			case env.Block:
				ser := ps.Ser
				ps.Ser = bloc.Series
				for ps.Ser.Pos() < ps.Ser.Len() {
					EvalExpressionInjLimited(ps, arg0, true)
					//					EvalExpression2(ps, false)
					if util.IsTruthy(ps.Res) {
						break
					}
				}
				ps.Ser = ser
				return ps.Res
			default:
				return MakeArgError(ps, 2, []env.Type{env.BlockType}, "any\\with")
			}
		},
	},
}
