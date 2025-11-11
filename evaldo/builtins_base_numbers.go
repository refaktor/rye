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
	"addnums": {
		Argsn: 2,
		Doc:   "Optimized version of + that adds two numbers, working with both integers and decimals.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			// Fast path for the most common case: Integer + Integer
			if i1, ok1 := arg0.(env.Integer); ok1 {
				if i2, ok2 := arg1.(env.Integer); ok2 {
					// Direct integer addition without creating a new object until the end
					i1.Value = i1.Value + i2.Value
					return i1 // we don't have to create new Value as it's already copied by value
					// return *env.NewInteger(i1.Value + i2.Value)
				}

				// Handle Integer + Decimal case
				if d2, ok2 := arg1.(env.Decimal); ok2 {
					return *env.NewDecimal(float64(i1.Value) + d2.Value)
				}

				// Type error for second argument
				return MakeArgError(ps, 2, []env.Type{env.IntegerType, env.DecimalType}, "addnums")
			}

			// Handle Decimal + (Integer or Decimal) case
			if d1, ok1 := arg0.(env.Decimal); ok1 {
				if i2, ok2 := arg1.(env.Integer); ok2 {
					return *env.NewDecimal(d1.Value + float64(i2.Value))
				}

				if d2, ok2 := arg1.(env.Decimal); ok2 {
					return *env.NewDecimal(d1.Value + d2.Value)
				}

				// Type error for second argument
				return MakeArgError(ps, 2, []env.Type{env.IntegerType, env.DecimalType}, "addnums")
			}

			// Handle Time + Integer case
			if t1, ok1 := arg0.(env.Time); ok1 {
				if i2, ok2 := arg1.(env.Integer); ok2 {
					return *env.NewTime(t1.Value.Add(time.Duration(i2.Value * 1000000)))
				}

				// Error for invalid first argument type
				return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "addnums")
			}

			// Type error for first argument
			return MakeArgError(ps, 1, []env.Type{env.IntegerType, env.DecimalType, env.TimeType}, "addnums")
		},
	},

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
	// equal { decr 124 } 123
	// equal { decr 1 } 0
	// equal { decr -4 } -5
	// error { decr "123" }
	// Args:
	// * value: Integer to decrement
	// Returns:
	// * integer value decremented by 1
	"decr": { // ***
		Argsn: 1,
		Doc:   "Decrements an integer value by 1.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch arg := arg0.(type) {
			case env.Integer:
				return *env.NewInteger(arg.Value - 1)
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType}, "decr")
			}
		},
	},

	// Tests:
	// equal { negate 123 } -123
	// equal { negate -123 } 123
	// equal { negate 0 } 0
	// equal { negate 5.5 } -5.5
	// equal { negate -2.3 } 2.3
	// error { negate "123" }
	// Args:
	// * value: Number (integer, decimal, or complex) to negate
	// Returns:
	// * negated number of the same type
	"negate": { // ***
		Argsn: 1,
		Doc:   "Negates a number by multiplying it by -1, works with integers, decimals, and complex numbers.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch arg := arg0.(type) {
			case env.Integer:
				return *env.NewInteger(-arg.Value)
			case env.Decimal:
				return *env.NewDecimal(-arg.Value)
			case env.Complex:
				return *env.NewComplex(-arg.Value)
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType, env.DecimalType, env.ComplexType}, "negate")
			}
		},
	},

	// Tests:
	// equal { invert 2 } 0.5
	// equal { invert 4 } 0.25
	// equal { invert 0.5 } 2.0
	// equal { invert -2 } -0.5
	// equal { invert 1 } 1.0
	// error { invert 0 }
	// error { invert "123" }
	// Args:
	// * value: Number (integer, decimal, or complex) to invert (must not be zero)
	// Returns:
	// * reciprocal (1/value) as decimal or complex number
	"invert": { // ***
		Argsn: 1,
		Doc:   "Calculates the reciprocal (1/x) of a number, works with integers, decimals, and complex numbers.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch arg := arg0.(type) {
			case env.Integer:
				if arg.Value == 0 {
					ps.FailureFlag = true
					return MakeBuiltinError(ps, "Can't invert zero.", "invert")
				}
				return *env.NewDecimal(1.0 / float64(arg.Value))
			case env.Decimal:
				if arg.Value == 0.0 {
					ps.FailureFlag = true
					return MakeBuiltinError(ps, "Can't invert zero.", "invert")
				}
				return *env.NewDecimal(1.0 / arg.Value)
			case env.Complex:
				if arg.Value == complex(0, 0) {
					ps.FailureFlag = true
					return MakeBuiltinError(ps, "Can't invert zero complex number.", "invert")
				}
				return *env.NewComplex(complex(1.0, 0.0) / arg.Value)
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType, env.DecimalType, env.ComplexType}, "invert")
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
	// equal { random\integer 1 |< 2 } true
	// equal { random\integer 100 | >= 0 } true
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
	// equal { random\decimal 2.0 |type? } 'decimal
	// equal { random\decimal 1.0 |< 1.0 } true
	// equal { random\decimal 100.0 | >= 0.0 } true
	// Args:
	// * max: Upper bound (exclusive) for the random number
	// Returns:
	// * random decimal in the range [0.0, max)
	"random\\decimal": {
		Argsn: 1,
		Doc:   "Generates a cryptographically secure random decimal between 0.0 (inclusive) and the specified maximum (exclusive).",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch arg := arg0.(type) {
			case env.Decimal:
				// Generate a random integer in a large range to get good precision
				maxInt := int64(1000000000000) // 10^12 for good precision
				val, err := rand.Int(rand.Reader, big.NewInt(maxInt))
				if err != nil {
					return MakeBuiltinError(ps, err.Error(), "random-decimal")
				}
				// Convert to float64 in range [0, 1) and scale by max
				randomFloat := float64(val.Int64()) / float64(maxInt)
				return *env.NewDecimal(randomFloat * arg.Value)
			case env.Integer:
				// Allow integer input, convert to decimal
				maxInt := int64(1000000000000) // 10^12 for good precision
				val, err := rand.Int(rand.Reader, big.NewInt(maxInt))
				if err != nil {
					return MakeBuiltinError(ps, err.Error(), "random-decimal")
				}
				// Convert to float64 in range [0, 1) and scale by max
				randomFloat := float64(val.Int64()) / float64(maxInt)
				return *env.NewDecimal(randomFloat * float64(arg.Value))
			default:
				return MakeArgError(ps, 1, []env.Type{env.DecimalType, env.IntegerType}, "random-decimal")
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

						// Attempt to modify the word
						ret := *env.NewInteger(1 + iintval.Value)

						if ok := ctx.Mod(arg.Index, ret); !ok {
							ps.FailureFlag = true
							return env.NewError("Cannot modify constant '" + ps.Idx.GetWord(arg.Index) + "', use 'var' to declare it as a variable")
						}

						return ret

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
	// equal { a:: 123 decr! 'a a } 122
	// equal { counter:: 1 decr! 'counter counter } 0
	// error { decr! 123 }
	// Args:
	// * word: Word referring to an integer value to decrement
	// Returns:
	// * the new decremented integer value
	"decr!": { // ***
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
						// Attempt to modify the word
						ret := *env.NewInteger(iintval.Value - 1)

						if ok := ctx.Mod(arg.Index, ret); !ok {
							ps.FailureFlag = true
							return env.NewError("Cannot modify constant '" + ps.Idx.GetWord(arg.Index) + "', use 'var' to declare it as a variable")
						}

						return ret
					default:
						return MakeBuiltinError(ps, "Value in word is not integer.", "decr!")
					}
				}
				return MakeBuiltinError(ps, "Word not found in context.", "decr!")

			default:
				return MakeArgError(ps, 1, []env.Type{env.WordType}, "decr!")
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
	// error { "A" + "b" }
	// error { "A" + 1 }
	// error { { 1 2 } + { 3 4 } } { 1 2 3 4 }
	// error { dict { "a" 1 } |+ { "b" 2 } }
	// error { dict { "a" 1 } |+ dict { "b" 2 } }
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
				case env.Complex:
					return *env.NewComplex(complex(float64(s1.Value), 0) + s2.Value)
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType, env.DecimalType, env.ComplexType}, "_+")
				}
			case env.Decimal:
				switch s2 := arg1.(type) {
				case env.Integer:
					return *env.NewDecimal(s1.Value + float64(s2.Value))
				case env.Decimal:
					return *env.NewDecimal(s1.Value + s2.Value)
				case env.Complex:
					return *env.NewComplex(complex(s1.Value, 0) + s2.Value)
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType, env.DecimalType, env.ComplexType}, "_+")
				}
			case env.Complex:
				switch s2 := arg1.(type) {
				case env.Integer:
					return *env.NewComplex(s1.Value + complex(float64(s2.Value), 0))
				case env.Decimal:
					return *env.NewComplex(s1.Value + complex(s2.Value, 0))
				case env.Complex:
					return *env.NewComplex(s1.Value + s2.Value)
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType, env.DecimalType, env.ComplexType}, "_+")
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
				return MakeArgError(ps, 1, []env.Type{env.IntegerType, env.DecimalType, env.ComplexType, env.TimeType}, "_+")
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
		Doc:   "Subtracts the second number from the first, working with integers, decimals, and complex numbers.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch a := arg0.(type) {
			case env.Integer:
				switch b := arg1.(type) {
				case env.Integer:
					return *env.NewInteger(a.Value - b.Value)
				case env.Decimal:
					return *env.NewDecimal(float64(a.Value) - b.Value)
				case env.Complex:
					return *env.NewComplex(complex(float64(a.Value), 0) - b.Value)
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType, env.DecimalType, env.ComplexType}, "_-")
				}
			case env.Decimal:
				switch b := arg1.(type) {
				case env.Integer:
					return *env.NewDecimal(a.Value - float64(b.Value))
				case env.Decimal:
					return *env.NewDecimal(a.Value - b.Value)
				case env.Complex:
					return *env.NewComplex(complex(a.Value, 0) - b.Value)
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType, env.DecimalType, env.ComplexType}, "_-")
				}
			case env.Complex:
				switch b := arg1.(type) {
				case env.Integer:
					return *env.NewComplex(a.Value - complex(float64(b.Value), 0))
				case env.Decimal:
					return *env.NewComplex(a.Value - complex(b.Value, 0))
				case env.Complex:
					return *env.NewComplex(a.Value - b.Value)
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType, env.DecimalType, env.ComplexType}, "_-")
				}
			case env.Time:
				switch b2 := arg1.(type) {
				case env.Integer:
					v := a.Value.Add(time.Duration(-1000000 * b2.Value))
					return *env.NewTime(v)
				case env.Time:
					v1 := a.Value.Sub(b2.Value)
					return *env.NewInteger(int64(v1) / 1000000)
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType, env.TimeType}, "_-")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType, env.DecimalType, env.ComplexType, env.TimeType}, "_-")
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
		Doc:   "Multiplies two numbers, working with integers, decimals, and complex numbers.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch a := arg0.(type) {
			case env.Integer:
				switch b := arg1.(type) {
				case env.Integer:
					return *env.NewInteger(a.Value * b.Value)
				case env.Decimal:
					return *env.NewDecimal(float64(a.Value) * b.Value)
				case env.Complex:
					return *env.NewComplex(complex(float64(a.Value), 0) * b.Value)
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType, env.DecimalType, env.ComplexType}, "_*")
				}
			case env.Decimal:
				switch b := arg1.(type) {
				case env.Integer:
					return *env.NewDecimal(a.Value * float64(b.Value))
				case env.Decimal:
					return *env.NewDecimal(a.Value * b.Value)
				case env.Complex:
					return *env.NewComplex(complex(a.Value, 0) * b.Value)
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType, env.DecimalType, env.ComplexType}, "_*")
				}
			case env.Complex:
				switch b := arg1.(type) {
				case env.Integer:
					return *env.NewComplex(a.Value * complex(float64(b.Value), 0))
				case env.Decimal:
					return *env.NewComplex(a.Value * complex(b.Value, 0))
				case env.Complex:
					return *env.NewComplex(a.Value * b.Value)
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType, env.DecimalType, env.ComplexType}, "_*")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType, env.DecimalType, env.ComplexType}, "_*")
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
		Doc:   "Divides the first number by the second and returns a result, with error checking for division by zero.",
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
				case env.Complex:
					if b.Value == complex(0, 0) {
						ps.FailureFlag = true
						return MakeBuiltinError(ps, "Can't divide by Zero complex number.", "_/")
					}
					return *env.NewComplex(complex(float64(a.Value), 0) / b.Value)
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType, env.DecimalType, env.ComplexType}, "_/")
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
				case env.Complex:
					if b.Value == complex(0, 0) {
						ps.FailureFlag = true
						return MakeBuiltinError(ps, "Can't divide by Zero complex number.", "_/")
					}
					return *env.NewComplex(complex(a.Value, 0) / b.Value)
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType, env.DecimalType, env.ComplexType}, "_/")
				}
			case env.Complex:
				switch b := arg1.(type) {
				case env.Integer:
					if b.Value == 0 {
						ps.FailureFlag = true
						return MakeBuiltinError(ps, "Can't divide by Zero.", "_/")
					}
					return *env.NewComplex(a.Value / complex(float64(b.Value), 0))
				case env.Decimal:
					if b.Value == 0.0 {
						ps.FailureFlag = true
						return MakeBuiltinError(ps, "Can't divide by Zero.", "_/")
					}
					return *env.NewComplex(a.Value / complex(b.Value, 0))
				case env.Complex:
					if b.Value == complex(0, 0) {
						ps.FailureFlag = true
						return MakeBuiltinError(ps, "Can't divide by Zero complex number.", "_/")
					}
					return *env.NewComplex(a.Value / b.Value)
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType, env.DecimalType, env.ComplexType}, "_/")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType, env.DecimalType, env.ComplexType}, "_/")
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

	// Tests:
	// equal { clamp 5 0 10 } 5
	// equal { clamp -5 0 10 } 0
	// equal { clamp 15 0 10 } 10
	// equal { clamp 5.5 0.0 10.0 } 5.5
	// equal { clamp -2.3 0 10 } 0.0
	// equal { clamp 12.7 0 10 } 10.0
	// error { clamp "5" 0 10 }
	// Args:
	// * value: Number (integer or decimal) to clamp
	// * min: Minimum value (integer or decimal)
	// * max: Maximum value (integer or decimal)
	// Returns:
	// * clamped number, ensuring value is between min and max (inclusive)
	"clamp": { // ***
		Argsn: 3,
		Doc:   "Clamps a number between a minimum and maximum value, ensuring the result stays within the specified bounds.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			// Handle different combinations of Integer and Decimal for all three arguments
			switch val := arg0.(type) {
			case env.Integer:
				switch minVal := arg1.(type) {
				case env.Integer:
					switch maxVal := arg2.(type) {
					case env.Integer:
						// All integers
						result := val.Value
						if result < minVal.Value {
							result = minVal.Value
						}
						if result > maxVal.Value {
							result = maxVal.Value
						}
						return *env.NewInteger(result)
					case env.Decimal:
						// Integer, Integer, Decimal -> Decimal result
						result := float64(val.Value)
						minFloat := float64(minVal.Value)
						maxFloat := maxVal.Value
						if result < minFloat {
							result = minFloat
						}
						if result > maxFloat {
							result = maxFloat
						}
						return *env.NewDecimal(result)
					default:
						return MakeArgError(ps, 3, []env.Type{env.IntegerType, env.DecimalType}, "clamp")
					}
				case env.Decimal:
					switch maxVal := arg2.(type) {
					case env.Integer:
						// Integer, Decimal, Integer -> Decimal result
						result := float64(val.Value)
						minFloat := minVal.Value
						maxFloat := float64(maxVal.Value)
						if result < minFloat {
							result = minFloat
						}
						if result > maxFloat {
							result = maxFloat
						}
						return *env.NewDecimal(result)
					case env.Decimal:
						// Integer, Decimal, Decimal -> Decimal result
						result := float64(val.Value)
						minFloat := minVal.Value
						maxFloat := maxVal.Value
						if result < minFloat {
							result = minFloat
						}
						if result > maxFloat {
							result = maxFloat
						}
						return *env.NewDecimal(result)
					default:
						return MakeArgError(ps, 3, []env.Type{env.IntegerType, env.DecimalType}, "clamp")
					}
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType, env.DecimalType}, "clamp")
				}
			case env.Decimal:
				switch minVal := arg1.(type) {
				case env.Integer:
					switch maxVal := arg2.(type) {
					case env.Integer:
						// Decimal, Integer, Integer -> Decimal result
						result := val.Value
						minFloat := float64(minVal.Value)
						maxFloat := float64(maxVal.Value)
						if result < minFloat {
							result = minFloat
						}
						if result > maxFloat {
							result = maxFloat
						}
						return *env.NewDecimal(result)
					case env.Decimal:
						// Decimal, Integer, Decimal -> Decimal result
						result := val.Value
						minFloat := float64(minVal.Value)
						maxFloat := maxVal.Value
						if result < minFloat {
							result = minFloat
						}
						if result > maxFloat {
							result = maxFloat
						}
						return *env.NewDecimal(result)
					default:
						return MakeArgError(ps, 3, []env.Type{env.IntegerType, env.DecimalType}, "clamp")
					}
				case env.Decimal:
					switch maxVal := arg2.(type) {
					case env.Integer:
						// Decimal, Decimal, Integer -> Decimal result
						result := val.Value
						minFloat := minVal.Value
						maxFloat := float64(maxVal.Value)
						if result < minFloat {
							result = minFloat
						}
						if result > maxFloat {
							result = maxFloat
						}
						return *env.NewDecimal(result)
					case env.Decimal:
						// All decimals
						result := val.Value
						if result < minVal.Value {
							result = minVal.Value
						}
						if result > maxVal.Value {
							result = maxVal.Value
						}
						return *env.NewDecimal(result)
					default:
						return MakeArgError(ps, 3, []env.Type{env.IntegerType, env.DecimalType}, "clamp")
					}
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType, env.DecimalType}, "clamp")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType, env.DecimalType}, "clamp")
			}
		},
	},
}
