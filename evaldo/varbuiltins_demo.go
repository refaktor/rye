package evaldo

import (
	"github.com/refaktor/rye/env"
	// JM 20230825	"github.com/refaktor/rye/term"
)

var VarBuiltins_demo = map[string]*env.VarBuiltin{
	// Tests:
	// equal { var-not true } false
	// equal { var-not false } true
	// error { var-not 0 }
	// error { var-not 5 }
	// Args:
	// * value: Boolean value to be negated
	// Returns:
	// * boolean false if the input is true, true if the input is false
	"var-not": {
		Argsn: 1,
		Doc:   "Performs logical negation on boolean values only.",
		Pure:  true,
		Fn: func(env1 *env.ProgramState, args ...env.Object) env.Object {
			switch b := args[0].(type) {
			case env.Boolean:
				return *env.NewBoolean(!b.Value)
			default:
				return MakeArgError(env1, 1, []env.Type{env.BooleanType}, "not")
			}
		},
	},

	// Tests:
	// equal { true .var-and true } true
	// equal { false .var-and true } false
	// equal { true .var-and false } false
	// equal { false .var-and false } false
	// equal { 3 .var-and 5 } 1  ; bitwise 011 AND 101 = 001
	// Args:
	// * value1: First value (boolean or integer)
	// * value2: Second value (boolean or integer)
	// Returns:
	// * boolean result of logical AND operation if both inputs are booleans, otherwise integer result of bitwise AND
	"var-and": {
		Argsn: 2,
		Doc:   "Performs a logical AND operation between two boolean values, or a bitwise AND operation between two integer values.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, args ...env.Object) env.Object {
			switch s1 := args[0].(type) {
			case env.Boolean:
				switch s2 := args[1].(type) {
				case env.Boolean:
					return *env.NewBoolean(s1.Value && s2.Value)
				default:
					return MakeArgError(ps, 2, []env.Type{env.BooleanType}, "and")
				}
			case env.Integer:
				switch s2 := args[1].(type) {
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
}
