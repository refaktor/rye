package evaldo

import (
	"crypto/rand"
	"math/big"
	"time"

	"github.com/refaktor/rye/env"
	// JM 20230825	"github.com/refaktor/rye/term"
)

var builtins_numbers = map[string]*env.Builtin{

	//
	// ##### Numbers ##### "Working with numbers, integers and decimals."
	//
	// Tests:
	// equal { inc 123 } 124
	// equal { inc 0 } 1
	// equal { inc -5 } -4
	// error { inc "123" }
	// Args:
	// * value: Integer to increment
	// Returns:
	// * integer value incremented by 1
	"inc": { // ***
		Argsn: 1,
		Doc:   "Increments an integer value by 1.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch arg := arg0.(type) {
			case env.Integer:
				return *env.NewInteger(1 + arg.Value)
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType}, "inc")
			}
		},
	},

	// Tests:
	// equal { is-positive 123 } true
	// equal { is-positive -123 } false
	// equal { is-positive 0 } false
	// equal { is-positive 5.5 } true
	// error { is-positive "123" }
	// Args:
	// * value: Integer or decimal to check
	// Returns:
	// * boolean true if the value is positive, false otherwise
	"is-positive": { // ***
		Argsn: 1,
		Doc:   "Checks if a number is positive (greater than zero).",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch arg := arg0.(type) {
			case env.Integer:
				if arg.Value > 0 {
					return *env.NewBoolean(true)
				} else {
					return *env.NewBoolean(false)
				}
			case env.Decimal:
				if arg.Value > 0 {
					return *env.NewBoolean(true)
				} else {
					return *env.NewBoolean(false)
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType, env.DecimalType}, "is-positive")
			}
		},
	},

	// Tests:
	// equal { is-zero 0 } true
	// equal { is-zero 123 } false
	// equal { is-zero 0.0 } true
	// error { is-zero "123" }
	// Args:
	// * value: Integer or decimal to check
	// Returns:
	// * boolean true if the value is zero, false otherwise
	"is-zero": { // ***
		Argsn: 1,
		Doc:   "Checks if a number is exactly zero.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) (res env.Object) {
			switch arg := arg0.(type) {
			case env.Integer:
				if arg.Value == 0 {
					return *env.NewBoolean(true)
				} else {
					return *env.NewBoolean(false)
				}
			case env.Decimal:
				if arg.Value == 0 {
					return *env.NewBoolean(true)
				} else {
					return *env.NewBoolean(false)
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType, env.DecimalType}, "is-zero")
			}
		},
	},

	// Tests:
	// equal { 10 .is-multiple-of 2 } true
	// equal { 10 .is-multiple-of 3 } false
	// equal { 15 .is-multiple-of 5 } true
	// equal { 0 .is-multiple-of 5 } true
	// Args:
	// * value: Integer to check
	// * divisor: Integer divisor to check against
	// Returns:
	// * boolean true if value is divisible by divisor with no remainder, false otherwise
	"is-multiple-of": { // ***
		Argsn: 2,
		Doc:   "Checks if the first integer is evenly divisible by the second integer.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch a := arg0.(type) {
			case env.Integer:
				switch b := arg1.(type) {
				case env.Integer:
					if a.Value%b.Value == 0 {
						return *env.NewBoolean(true)
					} else {
						return *env.NewBoolean(false)
					}
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "multiple-of")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType}, "multiple-of")
			}
		},
	},
	// Tests:
	// equal { 3 .is-odd } true
	// equal { 2 .is-odd } false
	// equal { 0 .is-odd } false
	// equal { -5 .is-odd } true
	// Args:
	// * value: Integer to check
	// Returns:
	// * boolean true if the value is odd, false if even
	"is-odd": { // ***
		Argsn: 1,
		Doc:   "Checks if an integer is odd (not divisible by 2).",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch a := arg0.(type) {
			case env.Integer:
				if a.Value%2 != 0 {
					return *env.NewBoolean(true)
				} else {
					return *env.NewBoolean(false)
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType}, "odd")
			}
		},
	},
	// Tests:
	// equal { 3 .is-even } false
	// equal { 2 .is-even } true
	// equal { 0 .is-even } true
	// equal { -4 .is-even } true
	// Args:
	// * value: Integer to check
	// Returns:
	// * boolean true if the value is even, false if odd
	"is-even": { // ***
		Argsn: 1,
		Doc:   "Checks if an integer is even (divisible by 2).",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch a := arg0.(type) {
			case env.Integer:
				if a.Value%2 == 0 {
					return *env.NewBoolean(true)
				} else {
					return *env.NewBoolean(false)
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType}, "even")
			}
		},
	},

	// Tests:
	// equal { 4 .mod 2 } 0
	// equal { 5 .mod 2 } 1
	// equal { 5 .mod 3 } 2
	// equal { -5 .mod 3 } -2
	// Args:
	// * value: Integer dividend
	// * divisor: Integer divisor
	// Returns:
	// * integer remainder after division
	"mod": { // ***
		Argsn: 2,
		Doc:   "Calculates the modulo (remainder) when dividing the first integer by the second.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch a := arg0.(type) {
			case env.Integer:
				switch b := arg1.(type) {
				case env.Integer:
					return *env.NewInteger(a.Value % b.Value)
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "mod")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType}, "mod")
			}
		},
	},

	// Tests:
	// equal { 4 % 2 } 0
	// equal { 5 % 2 } 1
	// equal { 5 % 3 } 2
	// Args:
	// * value: Integer dividend
	// * divisor: Integer divisor
	// Returns:
	// * integer remainder after division
	"_%": { // ***
		Argsn: 2,
		Doc:   "Alias for mod - calculates the modulo (remainder) when dividing the first integer by the second.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch a := arg0.(type) {
			case env.Integer:
				switch b := arg1.(type) {
				case env.Integer:
					return *env.NewInteger(a.Value % b.Value)
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "mod")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType}, "mod")
			}
		},
	},
	// Tests:
	// equal { random\integer 2 |type? } 'integer
	// equal { random\integer 1 |< 2 } 1
	// equal { random\integer 100 | >= 0 } 1
	// Args:
	// * max: Upper bound (exclusive) for the random number
	// Returns:
	// * random integer in the range [0, max)
	"random\\integer": {
		Argsn: 1,
		Doc:   "Generates a cryptographically secure random integer between 0 (inclusive) and the specified maximum (exclusive).",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch arg := arg0.(type) {
			case env.Integer:
				val, err := rand.Int(rand.Reader, big.NewInt(arg.Value))
				if err != nil {
					return MakeBuiltinError(ps, err.Error(), "random-integer")
				}
				return *env.NewInteger(val.Int64())
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType}, "random-integer")
			}
		},
	},

	// Tests:
	// equal { a:: 123 inc! 'a a } 124
	// equal { counter:: 0 inc! 'counter counter } 1
	// error { inc! 123 }
	// Args:
	// * word: Word referring to an integer value to increment
	// Returns:
	// * the new incremented integer value
	"inc!": { // ***
		Argsn: 1,
		Doc:   "Increments an integer value stored in a variable (word) by 1 and updates the variable in-place.",
		Pure:  false,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch arg := arg0.(type) {
			case env.Word:
				intval, found, ctx := ps.Ctx.Get2(arg.Index)
				if found {
					switch iintval := intval.(type) {
					case env.Integer:
						ctx.Mod(arg.Index, *env.NewInteger(1 + iintval.Value))
						return *env.NewInteger(1 + iintval.Value)
					default:
						return MakeBuiltinError(ps, "Value in word is not integer.", "inc!")
					}
				}
				return MakeBuiltinError(ps, "Word not found in context.", "inc!")

			default:
				return MakeArgError(ps, 1, []env.Type{env.WordType}, "inc!")
			}
		},
	},

	// Tests:
	// equal { a:: 123 dec! 'a a } 122
	// equal { counter:: 1 dec! 'counter counter } 0
	// error { dec! 123 }
	// Args:
	// * word: Word referring to an integer value to decrement
	// Returns:
	// * the new decremented integer value
	"dec!": { // ***
		Argsn: 1,
		Doc:   "Decrements an integer value stored in a variable (word) by 1 and updates the variable in-place.",
		Pure:  false,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch arg := arg0.(type) {
			case env.Word:
				intval, found, ctx := ps.Ctx.Get2(arg.Index)
				if found {
					switch iintval := intval.(type) {
					case env.Integer:
						ctx.Mod(arg.Index, *env.NewInteger(iintval.Value - 1))
						return *env.NewInteger(1 + iintval.Value)
					default:
						return MakeBuiltinError(ps, "Value in word is not integer.", "dec!")
					}
				}
				return MakeBuiltinError(ps, "Word not found in context.", "dec!")

			default:
				return MakeArgError(ps, 1, []env.Type{env.WordType}, "dec!")
			}
		},
	},

	// Tests:
	// equal { 4 . .type? } 'void
	// equal { "hello" . .type? } 'void
	// Args:
	// * value: Any value to discard
	// Returns:
	// * void value (used to discard values)
	"_.": { // ***
		Argsn: 1,
		Doc:   "Discards the input value and returns a void value, useful for ignoring unwanted results in a pipeline.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewVoid()
		},
	},

	// Tests:
	// equal { 1 + 1 } 2
	// equal { 3 + 4 } 7
	// equal { 5.6 + 7.8 } 13.400000
	// equal { "A" + "b" } "Ab"
	// equal { "A" + 1 } "A1"
	// equal { { 1 2 } + { 3 4 } } { 1 2 3 4 }
	// equal { dict { "a" 1 } |+ { "b" 2 } } dict { "a" 1 "b" 2 }
	// equal { dict { "a" 1 } |+ dict { "b" 2 } } dict { "a" 1 "b" 2 }
	// Args:
	// * value1: First value (number, string, block, dict, etc.)
	// * value2: Second value to add or join
	// Returns:
	// * result of adding or joining the values, type depends on input types
	"_+": { // **
		Argsn: 2,
		Doc:   "Adds or joins two values together, with behavior depending on types: adds numbers, concatenates strings/blocks, merges dictionaries, etc.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Integer:
				switch s2 := arg1.(type) {
				case env.Integer:
					return *env.NewInteger(s1.Value + s2.Value)
				case env.Decimal:
					return *env.NewDecimal(float64(s1.Value) + s2.Value)
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType, env.DecimalType}, "_+")
				}
			case env.Decimal:
				switch s2 := arg1.(type) {
				case env.Integer:
					return *env.NewDecimal(s1.Value + float64(s2.Value))
				case env.Decimal:
					return *env.NewDecimal(s1.Value + s2.Value)
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType, env.DecimalType}, "_+")
				}
			case env.Time:
				switch b2 := arg1.(type) {
				case env.Integer:
					v := s1.Value.Add(time.Duration(b2.Value * 1000000))
					return *env.NewTime(v)
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "_+")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType, env.DecimalType, env.TimeType}, "_+")
			}
		},
	},

	// Tests:
	// equal { 2 - 1 } 1
	// equal { 5 - 6 } -1
	// equal { 5.5 - 2.2 } 3.3
	// equal { 5 - 2.5 } 2.5
	// Args:
	// * value1: First number (integer or decimal)
	// * value2: Second number to subtract from the first
	// Returns:
	// * result of subtracting value2 from value1
	"_-": { // **
		Argsn: 2,
		Doc:   "Subtracts the second number from the first, working with both integers and decimals.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch a := arg0.(type) {
			case env.Integer:
				switch b := arg1.(type) {
				case env.Integer:
					return *env.NewInteger(a.Value - b.Value)
				case env.Decimal:
					return *env.NewDecimal(float64(a.Value) - b.Value)
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType, env.DecimalType}, "_-")
				}
			case env.Decimal:
				switch b := arg1.(type) {
				case env.Integer:
					return *env.NewDecimal(a.Value - float64(b.Value))
				case env.Decimal:
					return *env.NewDecimal(a.Value - b.Value)
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType, env.DecimalType}, "_-")
				}
			case env.Time:
				switch b2 := arg1.(type) {
				case env.Integer:
					v := a.Value.Add(time.Duration(-1000000 * b2.Value))
					return *env.NewTime(v)
				case env.Time:
					v1 := a.Value.Sub(b2.Value)
					return *env.NewInteger(int64(v1))
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "_+")
				}

			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType, env.DecimalType}, "_-")
			}
		},
	},

	// Tests:
	// equal { 4 * 2 } 8
	// equal { 2.5 * -2 } -5.0
	// equal { 0 * 5 } 0
	// equal { 1.5 * 2.5 } 3.75
	// Args:
	// * value1: First number (integer or decimal)
	// * value2: Second number to multiply by
	// Returns:
	// * product of the two numbers
	"_*": { // **
		Argsn: 2,
		Doc:   "Multiplies two numbers, working with both integers and decimals.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch a := arg0.(type) {
			case env.Integer:
				switch b := arg1.(type) {
				case env.Integer:
					return *env.NewInteger(a.Value * b.Value)
				case env.Decimal:
					return *env.NewDecimal(float64(a.Value) * b.Value)
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType, env.DecimalType}, "_*")
				}
			case env.Decimal:
				switch b := arg1.(type) {
				case env.Integer:
					return *env.NewDecimal(a.Value * float64(b.Value))
				case env.Decimal:
					return *env.NewDecimal(a.Value * b.Value)
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType, env.DecimalType}, "_*")
				}
			default:
				return MakeArgError(ps, 2, []env.Type{env.IntegerType, env.DecimalType}, "_*")
			}
		},
	},

	// Tests:
	// equal { 4 / 2 } 2.000
	// equal { 102.0 / 2.0 } 51.000
	// equal { 5 / 2 } 2.5
	// error { 5 / 0 }
	// Args:
	// * value1: Dividend (integer or decimal)
	// * value2: Divisor (integer or decimal, must not be zero)
	// Returns:
	// * decimal result of dividing value1 by value2
	"_/": { // **
		Argsn: 2,
		Doc:   "Divides the first number by the second and returns a decimal result, with error checking for division by zero.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch a := arg0.(type) {
			case env.Integer:
				switch b := arg1.(type) {
				case env.Integer:
					if b.Value == 0 {
						ps.FailureFlag = true
						return MakeBuiltinError(ps, "Can't divide by Zero.", "_/")
					}
					return *env.NewDecimal(float64(a.Value) / float64(b.Value))
				case env.Decimal:
					if b.Value == 0.0 {
						ps.FailureFlag = true
						return MakeBuiltinError(ps, "Can't divide by Zero.", "_/")
					}
					return *env.NewDecimal(float64(a.Value) / b.Value)
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType, env.DecimalType}, "_/")
				}
			case env.Decimal:
				switch b := arg1.(type) {
				case env.Integer:
					if b.Value == 0 {
						ps.FailureFlag = true
						return MakeBuiltinError(ps, "Can't divide by Zero.", "_/")
					}
					return *env.NewDecimal(a.Value / float64(b.Value))
				case env.Decimal:
					if b.Value == 0.0 {
						ps.FailureFlag = true
						return MakeBuiltinError(ps, "Can't divide by Zero.", "_/")
					}
					return *env.NewDecimal(a.Value / b.Value)
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType, env.DecimalType}, "_/")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType, env.DecimalType}, "_/")
			}
		},
	},

	// Tests:
	// equal { 5 // 2 } 2
	// equal { 102 // 5 } 20
	// equal { 7.99 // 2 } 3
	// equal { -5 // 2 } -2
	// error { 5 // 0 }
	// Args:
	// * value1: Dividend (integer or decimal)
	// * value2: Divisor (integer or decimal, must not be zero)
	// Returns:
	// * integer result of dividing value1 by value2 (truncated)
	"_//": { // **
		Argsn: 2,
		Doc:   "Performs integer division, dividing the first number by the second and truncating to an integer result.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch a := arg0.(type) {
			case env.Integer:
				switch b := arg1.(type) {
				case env.Integer:
					if b.Value == 0 {
						ps.FailureFlag = true
						return MakeBuiltinError(ps, "Can't divide by Zero.", "_//")
					}
					return *env.NewInteger(a.Value / b.Value)
				case env.Decimal:
					if b.Value == 0.0 {
						ps.FailureFlag = true
						return MakeBuiltinError(ps, "Can't divide by Zero.", "_//")
					}
					return *env.NewInteger(a.Value / int64(b.Value))
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType, env.DecimalType}, "_//")
				}
			case env.Decimal:
				switch b := arg1.(type) {
				case env.Integer:
					if b.Value == 0 {
						ps.FailureFlag = true
						return MakeBuiltinError(ps, "Can't divide by Zero.", "_//")
					}
					return *env.NewInteger(int64(a.Value) / b.Value)
				case env.Decimal:
					if b.Value == 0.0 {
						ps.FailureFlag = true
						return MakeBuiltinError(ps, "Can't divide by Zero.", "_//")
					}
					return *env.NewInteger(int64(a.Value) / int64(b.Value))
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType, env.DecimalType}, "_//")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType, env.DecimalType}, "_//")
			}
		},
	},

	// Tests:
	// equal { 5 = 5 } true
	// equal { 5 = 4 } false
	// equal { "abc" = "abc" } true
	// equal { { 1 2 } = { 1 2 } } true
	// equal { { 1 2 } = { 2 1 } } false
	// Args:
	// * value1: First value to compare
	// * value2: Second value to compare
	// Returns:
	// * boolean true if values are equal, false otherwise
	"_=": { // ***
		Argsn: 2,
		Doc:   "Compares two values for equality, returning 1 if equal or 0 if not equal.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			if arg0.Equal(arg1) {
				return *env.NewBoolean(true)
			} else {
				return *env.NewBoolean(false)
			}
		},
	},

	// Tests:
	// equal { 6 > 5 } true
	// equal { 5 > 5 } false
	// equal { 4 > 5 } false
	// equal { 5.5 > 5 } true
	// equal { "b" > "a" } true
	// Args:
	// * value1: First value to compare
	// * value2: Second value to compare
	// Returns:
	// * boolean true if value1 is greater than value2, false otherwise
	"_>": { // ***
		Argsn: 2,
		Doc:   "Compares if the first value is greater than the second value.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			if greaterThanNew(arg0, arg1) {
				return *env.NewBoolean(true)
			} else {
				return *env.NewBoolean(false)
			}
		},
	},

	// Tests:
	// equal { 5 >= 6 } false
	// equal { 5 >= 5 } true
	// equal { 6.0 >= 5 } true
	// equal { 4 >= 5 } false
	// equal { "b" >= "a" } true
	// equal { "a" >= "a" } true
	// Args:
	// * value1: First value to compare
	// * value2: Second value to compare
	// Returns:
	// * boolean true if value1 is greater than or equal to value2, false otherwise
	"_>=": { // * *
		Argsn: 2,
		Doc:   "Compares if the first value is greater than or equal to the second value.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			if arg0.Equal(arg1) || greaterThanNew(arg0, arg1) {
				return *env.NewBoolean(true)
			} else {
				return *env.NewBoolean(false)
			}
		},
	},

	// Tests:
	// equal { 5 < 6 } true
	// equal { 5 < 5 } false
	// equal { 6 < 5 } false
	// equal { 4.5 < 5 } true
	// equal { "a" < "b" } true
	// Args:
	// * value1: First value to compare
	// * value2: Second value to compare
	// Returns:
	// * boolean true if value1 is less than value2, false otherwise
	"_<": { // **
		Argsn: 2,
		Doc:   "Compares if the first value is less than the second value.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			if lesserThanNew(arg0, arg1) {
				return *env.NewBoolean(true)
			} else {
				return *env.NewBoolean(false)
			}
		},
	},

	// Tests:
	// equal { 5 <= 6 } true
	// equal { 5 <= 5 } true
	// equal { 6 <= 5 } false
	// equal { 4.5 <= 5 } true
	// equal { "a" <= "b" } true
	// equal { "a" <= "a" } true
	// Args:
	// * value1: First value to compare
	// * value2: Second value to compare
	// Returns:
	// * boolean true if value1 is less than or equal to value2, false otherwise
	"_<=": {
		Argsn: 2,
		Doc:   "Compares if the first value is less than or equal to the second value.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			if arg0.Equal(arg1) || lesserThanNew(arg0, arg1) {
				return *env.NewBoolean(true)
			} else {
				return *env.NewBoolean(false)
			}
		},
	},

	"recur-if": { //recur1-if
		Argsn: 2,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch cond := arg0.(type) {
			case env.Integer:
				if cond.Value > 0 {
					ps.Ser.Reset()
					return nil
				} else {
					return ps.Res
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType}, "recur-if")
			}
		},
	},
	//test if we can do recur similar to clojure one. Since functions in rejy are of fixed arity we would need recur1 recur2 recur3 and recur [ ] which is less optimal
	//otherwise word recur could somehow be bound to correct version or args depending on number of args of func. Try this at first.
	"recur-if\\1": { //recur1-if
		Argsn: 2,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch cond := arg0.(type) {
			case env.Integer:
				if cond.Value > 0 {
					switch arg := arg1.(type) {
					case env.Integer:
						ps.Ctx.Mod(ps.Args[0], arg)
						ps.Ser.Reset()
						return nil
					default:
						return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "recur-if\\1")
					}
				} else {
					return ps.Res
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType}, "recur-if\\1")
			}
		},
	},

	"recur-if\\2": { //recur1-if
		Argsn: 3,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			//arg0.Trace("a0")
			//arg1.Trace("a1")
			//arg2.Trace("a2")
			switch cond := arg0.(type) {
			case env.Integer:
				if cond.Value > 0 {
					switch argi1 := arg1.(type) {
					case env.Integer:
						switch argi2 := arg2.(type) {
						case env.Integer:
							ps.Ctx.Set(ps.Args[0], argi1)
							ps.Ctx.Set(ps.Args[1], argi2)
							ps.Ser.Reset()
							return ps.Res
						default:
							return MakeArgError(ps, 3, []env.Type{env.IntegerType}, "recur-if\\2")
						}
					default:
						return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "recur-if\\2")
					}
				} else {
					return ps.Res
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType}, "recur-if\\2")
			}
		},
	},

	"recur-if\\3": { //recur1-if
		Argsn: 4,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			//arg0.Trace("a0")
			//arg1.Trace("a1")
			//arg2.Trace("a2")
			switch cond := arg0.(type) {
			case env.Integer:
				if cond.Value > 0 {
					switch argi1 := arg1.(type) {
					case env.Integer:
						switch argi2 := arg2.(type) {
						case env.Integer:
							switch argi3 := arg3.(type) {
							case env.Integer:
								ps.Ctx.Set(ps.Args[0], argi1)
								ps.Ctx.Set(ps.Args[1], argi2)
								ps.Ctx.Set(ps.Args[2], argi3)
								ps.Ser.Reset()
								return ps.Res
							}
						default:
							return MakeArgError(ps, 3, []env.Type{env.IntegerType}, "recur-if\\3")
						}
					default:
						return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "recur-if\\3")
					}
				} else {
					return ps.Res
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType}, "recur-if\\3")
			}
			return nil
		},
	},
}
