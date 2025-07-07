package evaldo

import (
	"fmt"
	"math"
	"math/cmplx"

	"github.com/fxtlabs/primes"
	"github.com/refaktor/rye/env"
	"github.com/refaktor/rye/util"
)

// Integer represents an integer.

func Math_EvalBlock(es *env.ProgramState) []env.Object {
	opplus, _ := es.Idx.GetIndex("_+")
	opminus, _ := es.Idx.GetIndex("_-")
	opmul, _ := es.Idx.GetIndex("_*")
	opdiv, _ := es.Idx.GetIndex("_/")

	precedence := map[int]int{opplus: 1, opminus: 1, opmul: 2, opdiv: 2}
	var output []env.Object
	var operators []env.Opword
	for es.Ser.Pos() < es.Ser.Len() {
		object := es.Ser.Pop()
		switch obj := object.(type) {
		case env.Integer:
			output = append(output, obj)
		case env.Decimal:
			output = append(output, obj)
		case env.Block:
			ser1 := es.Ser
			es.Ser = obj.Series
			val := Math_EvalBlock(es)
			es.Ser = ser1
			output = append(output, val...)
		case env.Opword:
			_, found := precedence[obj.Index]
			if found {
				for len(operators) > 0 && precedence[operators[len(operators)-1].Index] >= precedence[obj.Index] {
					output = append(output, operators[len(operators)-1])
					operators = operators[:len(operators)-1]
				}
				operators = append(operators, obj)
			} else { // regular functions
				return nil // TODO return error
			}
		case env.Word:
			args := es.Ser.Pop()
			switch blk := args.(type) {
			case env.Block:
				ser1 := es.Ser
				es.Ser = blk.Series
				val := Math_EvalBlock(es)
				es.Ser = ser1
				output = append(output, val...)
			}
			output = append(output, obj)
		default:
			fmt.Println("Type is not matching - Validation_EvalBlock.")
		}
	}

	for len(operators) > 0 {
		output = append(output, operators[len(operators)-1])
		operators = operators[:len(operators)-1]
	}

	return output
}

func DialectMath(env1 *env.ProgramState, arg0 env.Object) env.Object {
	switch blk := arg0.(type) {
	case env.Block:
		ser1 := env1.Ser
		env1.Ser = blk.Series
		val := Math_EvalBlock(env1)
		env1.Ser = ser1
		return *env.NewBlock(*env.NewTSeries(val))
	default:
		return *env.NewError("arg 1 should be block")
	}
}

func assureFloats(aa env.Object, bb env.Object) (float64, float64, int) {
	var fa, fb float64
	switch a := aa.(type) {
	case env.Integer:
		fa = float64(a.Value)
	case env.Decimal:
		fa = a.Value
	default:
		return 0.0, 0.0, 1 // MakeArgError(ps, 1, []env.Type{env.IntegerType, env.BlockType}, "mod")
	}
	switch b := bb.(type) {
	case env.Integer:
		fb = float64(b.Value)
	case env.Decimal:
		fb = b.Value
	default:
		return 0.0, 0.0, 2 // MakeArgError(ps, 2, []env.Type{env.IntegerType, env.BlockType}, "mod")
	}
	return fa, fb, 0
}

var Builtins_math = map[string]*env.Builtin{

	//
	// ##### Math ##### "Mathematical functions"
	//
	// Tests:
	// equal { mod 10 3 } 1.0
	// Args:
	// * x: integer or decimal value
	// * y: integer or decimal value
	// Returns:
	// * decimal remainder of x/y
	"mod": {
		Argsn: 2,
		Doc:   "Returns the remainder of dividing x by y.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			fa, fb, errPos := assureFloats(arg0, arg1)
			if errPos > 0 {
				return MakeArgError(ps, errPos, []env.Type{env.IntegerType, env.BlockType}, "mod")
			}
			return *env.NewDecimal(math.Mod(fa, fb))
		},
	},
	// Tests:
	// equal { pow 2 3 } 8.0
	// equal { pow complex 2 0 complex 3 0 |print } "8.000000+0.000000i"
	// equal { pow complex 0 1 complex 2 0 |print } "-1.000000+0.000000i"
	// Args:
	// * base: integer, decimal, or complex value
	// * exponent: integer, decimal, or complex value
	// Returns:
	// * decimal result of base raised to the power of exponent for integer/decimal inputs, complex result for complex inputs
	"pow": {
		Argsn: 2,
		Doc:   "Returns base raised to the power of exponent.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch base := arg0.(type) {
			case env.Complex:
				switch exp := arg1.(type) {
				case env.Complex:
					return *env.NewComplex(cmplx.Pow(base.Value, exp.Value))
				case env.Integer:
					return *env.NewComplex(cmplx.Pow(base.Value, complex(float64(exp.Value), 0)))
				case env.Decimal:
					return *env.NewComplex(cmplx.Pow(base.Value, complex(exp.Value, 0)))
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType, env.DecimalType, env.ComplexType}, "pow")
				}
			case env.Integer:
				switch exp := arg1.(type) {
				case env.Complex:
					return *env.NewComplex(cmplx.Pow(complex(float64(base.Value), 0), exp.Value))
				case env.Integer:
					return *env.NewDecimal(math.Pow(float64(base.Value), float64(exp.Value)))
				case env.Decimal:
					return *env.NewDecimal(math.Pow(float64(base.Value), exp.Value))
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType, env.DecimalType, env.ComplexType}, "pow")
				}
			case env.Decimal:
				switch exp := arg1.(type) {
				case env.Complex:
					return *env.NewComplex(cmplx.Pow(complex(base.Value, 0), exp.Value))
				case env.Integer:
					return *env.NewDecimal(math.Pow(base.Value, float64(exp.Value)))
				case env.Decimal:
					return *env.NewDecimal(math.Pow(base.Value, exp.Value))
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType, env.DecimalType, env.ComplexType}, "pow")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType, env.DecimalType, env.ComplexType}, "pow")
			}
		},
	},
	// Tests:
	// equal { log2 8 } 3.0
	// Args:
	// * x: integer or decimal value
	// Returns:
	// * decimal binary logarithm of x
	"log2": {
		Argsn: 1,
		Doc:   "Returns the binary logarithm (base-2) of x.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch val := arg0.(type) {
			case env.Integer:
				return *env.NewDecimal(math.Log2(float64(val.Value)))
			case env.Decimal:
				return *env.NewDecimal(math.Log2(val.Value))
			default:
				return MakeArgError(ps, 2, []env.Type{env.IntegerType, env.BlockType}, "log2")
			}
		},
	},
	// Tests:
	// equal { log10 100 } 2.0
	// Args:
	// * x: integer or decimal value
	// Returns:
	// * decimal base-10 logarithm of x
	"log10": {
		Argsn: 1,
		Doc:   "Returns the decimal logarithm (base-10) of x.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch val := arg0.(type) {
			case env.Integer:
				return *env.NewDecimal(math.Log10(float64(val.Value)))
			case env.Decimal:
				return *env.NewDecimal(math.Log10(val.Value))
			default:
				return MakeArgError(ps, 2, []env.Type{env.IntegerType, env.DecimalType}, "log10")
			}
		},
	},
	// Tests:
	// equal { log 1 } 0.0
	// equal { log 2.718281828459045 } 1.0
	// equal { log complex 1 0 |print } "0.000000+0.000000i"
	// equal { log complex 2.718281828459045 0 |print } "1.000000+0.000000i"
	// error { log 0 }
	// error { log complex 0 0 }
	// Args:
	// * x: integer, decimal, or complex value (must not be zero)
	// Returns:
	// * decimal natural logarithm of x for integer/decimal input, complex logarithm for complex input
	"log": {
		Argsn: 1,
		Doc:   "Returns the natural logarithm of x.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch val := arg0.(type) {
			case env.Integer:
				if val.Value <= 0 {
					ps.FailureFlag = true
					return MakeBuiltinError(ps, "Can't compute logarithm of zero or negative number.", "log")
				}
				return *env.NewDecimal(math.Log(float64(val.Value)))
			case env.Decimal:
				if val.Value <= 0 {
					ps.FailureFlag = true
					return MakeBuiltinError(ps, "Can't compute logarithm of zero or negative number.", "log")
				}
				return *env.NewDecimal(math.Log(val.Value))
			case env.Complex:
				if val.Value == complex(0, 0) {
					ps.FailureFlag = true
					return MakeBuiltinError(ps, "Can't compute logarithm of zero.", "log")
				}
				return *env.NewComplex(cmplx.Log(val.Value))
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType, env.DecimalType, env.ComplexType}, "log")
			}
		},
	},
	// Tests:
	// equal { log1p 0 } 0.0
	// Args:
	// * x: integer or decimal value
	// Returns:
	// * decimal natural logarithm of (1 + x)
	"log1p": {
		Argsn: 1,
		Doc:   "Returns the natural logarithm of 1 plus its argument x.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch val := arg0.(type) {
			case env.Integer:
				return *env.NewDecimal(math.Log1p(float64(val.Value)))
			case env.Decimal:
				return *env.NewDecimal(math.Log1p(val.Value))
			default:
				return MakeArgError(ps, 2, []env.Type{env.IntegerType, env.DecimalType}, "log1p")
			}
		},
	},
	// Tests:
	// equal { logb 8 } 3.0
	// Args:
	// * x: integer or decimal value
	// Returns:
	// * decimal binary exponent of x
	"logb": {
		Argsn: 1,
		Doc:   "Returns the binary exponent of x.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch val := arg0.(type) {
			case env.Integer:
				return *env.NewDecimal(math.Logb(float64(val.Value)))
			case env.Decimal:
				return *env.NewDecimal(math.Logb(val.Value))
			default:
				return MakeArgError(ps, 2, []env.Type{env.IntegerType, env.DecimalType}, "logb")
			}
		},
	},
	// Tests:
	// equal { sq 4 } 16.0
	// Args:
	// * x: integer or decimal value
	// Returns:
	// * decimal square of x
	"sq": {
		Argsn: 1,
		Doc:   "Returns the square of x.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch val := arg0.(type) {
			case env.Integer:
				return *env.NewDecimal(math.Pow(float64(val.Value), 2.0))
			case env.Decimal:
				return *env.NewDecimal(math.Pow(val.Value, 2.0))
			default:
				return MakeArgError(ps, 2, []env.Type{env.IntegerType, env.BlockType}, "sq")
			}
		},
	},
	// Tests:
	// equal { sin 0 } 0.0
	// equal { round\to sin pi 10 } 0.0
	// equal { sin complex 0 0 |print } "0.000000+0.000000i"
	// equal { sin complex 1.570796326794897 0 |print } "1.000000+0.000000i"
	// Args:
	// * x: integer, decimal, or complex value in radians
	// Returns:
	// * decimal sine of x for integer/decimal input, complex sine for complex input
	"sin": {
		Argsn: 1,
		Doc:   "Returns the sine of the radian argument.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch val := arg0.(type) {
			case env.Integer:
				return *env.NewDecimal(math.Sin(float64(val.Value)))
			case env.Decimal:
				return *env.NewDecimal(math.Sin(val.Value))
			case env.Complex:
				return *env.NewComplex(cmplx.Sin(val.Value))
			default:
				return MakeArgError(ps, 2, []env.Type{env.IntegerType, env.DecimalType, env.ComplexType}, "sin")
			}
		},
	},
	// Tests:
	// equal { cos 0 } 1.0
	// equal { cos complex 0 0 |print } "1.000000+0.000000i"
	// equal { cos complex 3.141592653589793 0 |print } "-1.000000+0.000000i"
	// Args:
	// * x: integer, decimal, or complex value in radians
	// Returns:
	// * decimal cosine of x for integer/decimal input, complex cosine for complex input
	"cos": {
		Argsn: 1,
		Doc:   "Returns the cosine of the radian argument.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch val := arg0.(type) {
			case env.Integer:
				return *env.NewDecimal(math.Cos(float64(val.Value)))
			case env.Decimal:
				return *env.NewDecimal(math.Cos(val.Value))
			case env.Complex:
				return *env.NewComplex(cmplx.Cos(val.Value))
			default:
				return MakeArgError(ps, 2, []env.Type{env.IntegerType, env.DecimalType, env.ComplexType}, "cos")
			}
		},
	},
	// Tests:
	// equal { tan 0 } 0.0
	// equal { tan complex 0 0 |print } "0.000000+0.000000i"
	// equal { tan complex 0.7853981633974483 0 |print } "1.000000+0.000000i"
	// Args:
	// * x: integer, decimal, or complex value in radians
	// Returns:
	// * decimal tangent of x for integer/decimal input, complex tangent for complex input
	"tan": {
		Argsn: 1,
		Doc:   "Returns the tangent of the radian argument.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch val := arg0.(type) {
			case env.Integer:
				return *env.NewDecimal(math.Tan(float64(val.Value)))
			case env.Decimal:
				return *env.NewDecimal(math.Tan(val.Value))
			case env.Complex:
				return *env.NewComplex(cmplx.Tan(val.Value))
			default:
				return MakeArgError(ps, 2, []env.Type{env.IntegerType, env.DecimalType, env.ComplexType}, "tan")
			}
		},
	},
	// Tests:
	// equal { sqrt 9 } 3.0
	// equal { sqrt complex -1 0 |print } "0.000000+1.000000i"
	// Args:
	// * x: integer, decimal, or complex value
	// Returns:
	// * decimal square root of x for integer/decimal input, complex square root for complex input
	"sqrt": {
		Argsn: 1,
		Doc:   "Returns the square root of x.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch val := arg0.(type) {
			case env.Integer:
				return *env.NewDecimal(math.Sqrt(float64(val.Value)))
			case env.Decimal:
				return *env.NewDecimal(math.Sqrt(val.Value))
			case env.Complex:
				return *env.NewComplex(cmplx.Sqrt(val.Value))
			default:
				return MakeArgError(ps, 2, []env.Type{env.IntegerType, env.DecimalType, env.ComplexType}, "sqrt")
			}
		},
	},
	// Tests:
	// equal { abs -5 } 5
	// equal { abs 5 } 5
	// equal { abs complex 3 4 } 5.0
	// Args:
	// * x: integer, decimal, or complex value
	// Returns:
	// * absolute value of x (same type as input for integer/decimal, decimal for complex)
	"abs": {
		Argsn: 1,
		Doc:   "Returns the absolute value of a number.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch val := arg0.(type) {
			case env.Integer:
				return *env.NewInteger(int64(math.Abs(float64(val.Value))))
			case env.Decimal:
				return *env.NewDecimal(math.Abs(val.Value))
			case env.Complex:
				return *env.NewDecimal(cmplx.Abs(val.Value))
			default:
				return MakeArgError(ps, 2, []env.Type{env.IntegerType, env.DecimalType, env.ComplexType}, "abs")
			}
		},
	},
	// Tests:
	// equal { acos 1 } 0.0
	// Args:
	// * x: integer or decimal value between -1 and 1
	// Returns:
	// * decimal arccosine of x in radians
	"acos": {
		Argsn: 1,
		Doc:   "Returns the arccosine (inverse cosine) in radians.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch val := arg0.(type) {
			case env.Integer:
				if val.Value < -1.0 || val.Value > 1.0 {
					return MakeBuiltinError(ps, "Invalid input: Acos is only defined for -1 <= x <= 1.", "acos")
				}
				return *env.NewDecimal(math.Acos(float64(val.Value)))
			case env.Decimal:
				if val.Value < -1.0 || val.Value > 1.0 {
					return MakeBuiltinError(ps, "Invalid input: Acos is only defined for -1 <= x <= 1.", "acos")
				}
				return *env.NewDecimal(math.Acos(val.Value))
			default:
				return MakeArgError(ps, 2, []env.Type{env.IntegerType, env.DecimalType}, "acos")
			}
		},
	},
	// Tests:
	// equal { acosh 1 } 0.0
	// Args:
	// * x: integer or decimal value >= 1
	// Returns:
	// * decimal inverse hyperbolic cosine of x
	"acosh": {
		Argsn: 1,
		Doc:   "Returns the inverse hyperbolic cosine.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch val := arg0.(type) {
			case env.Integer:
				if val.Value < 1.0 {
					return MakeBuiltinError(ps, " Acosh is only defined for x >= 1.", "acosh")
				}
				return *env.NewDecimal(math.Log(float64(val.Value) + math.Sqrt(float64(val.Value)*float64(val.Value)-1)))
			case env.Decimal:
				if val.Value < 1.0 {
					return MakeBuiltinError(ps, " Acosh is only defined for x >= 1.", "acosh")
				}
				return *env.NewDecimal(math.Log(val.Value + math.Sqrt(val.Value*val.Value-1)))
			default:
				return MakeArgError(ps, 2, []env.Type{env.IntegerType, env.DecimalType}, "acosh")
			}
		},
	},
	// Tests:
	// equal { asin 0 } 0.0
	// Args:
	// * x: integer or decimal value between -1 and 1
	// Returns:
	// * decimal arcsine of x in radians
	"asin": {
		Argsn: 1,
		Doc:   "Returns the arcsine (inverse sine) in radians.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch val := arg0.(type) {
			case env.Integer:
				if val.Value < -1.0 || val.Value > 1.0 {
					return MakeBuiltinError(ps, "Invalid input: Asin is only defined for -1 <= x <= 1.", "asin")
				}
				return *env.NewDecimal(math.Asin(float64(val.Value)))
			case env.Decimal:
				if val.Value < -1.0 || val.Value > 1.0 {
					return MakeBuiltinError(ps, "Invalid input: Asin is only defined for -1 <= x <= 1.", "asin")
				}
				return *env.NewDecimal(math.Asin(val.Value))
			default:
				return MakeArgError(ps, 2, []env.Type{env.IntegerType, env.DecimalType}, "asin")
			}
		},
	},
	// Tests:
	// equal { asinh 0 } 0.0
	// Args:
	// * x: integer or decimal value
	// Returns:
	// * decimal inverse hyperbolic sine of x
	"asinh": {
		Argsn: 1,
		Doc:   "Returns the inverse hyperbolic sine.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch val := arg0.(type) {
			case env.Integer:
				return *env.NewDecimal(math.Log(float64(val.Value) + math.Sqrt(float64(val.Value)*float64(val.Value)+1)))
			case env.Decimal:
				return *env.NewDecimal(math.Log(val.Value + math.Sqrt(val.Value*val.Value+1)))
			default:
				return MakeArgError(ps, 2, []env.Type{env.IntegerType, env.DecimalType}, "asinh")
			}
		},
	},
	// Tests:
	// equal { atan 0 } 0.0
	// Args:
	// * x: integer or decimal value
	// Returns:
	// * decimal arctangent of x in radians
	"atan": {
		Argsn: 1,
		Doc:   "Returns the arctangent (inverse tangent) in radians.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch val := arg0.(type) {
			case env.Integer:
				return *env.NewDecimal(math.Atan(float64(val.Value)))
			case env.Decimal:
				return *env.NewDecimal(math.Atan(val.Value))
			default:
				return MakeArgError(ps, 2, []env.Type{env.IntegerType, env.DecimalType}, "atan")
			}
		},
	},
	// Tests:
	// equal { atan2 0 1 } 0.0
	// Args:
	// * y: integer or decimal value
	// * x: integer or decimal value
	// Returns:
	// * decimal arctangent of y/x in radians, using signs to determine quadrant
	"atan2": {
		Argsn: 2,
		Doc:   "Returns the arc tangent of y/x, using the signs of the two to determine the quadrant of the return value.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch val := arg0.(type) {
			case env.Integer:
				switch val2 := arg1.(type) {
				case env.Integer:
					return *env.NewDecimal(math.Atan2(float64(val.Value), float64(val2.Value)))
				case env.Decimal:
					return *env.NewDecimal(math.Atan2(float64(val.Value), val2.Value))
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType, env.DecimalType}, "atan2")
				}
			case env.Decimal:
				switch val2 := arg1.(type) {
				case env.Integer:
					return *env.NewDecimal(math.Atan2(val.Value, float64(val2.Value)))
				case env.Decimal:
					return *env.NewDecimal(math.Atan2(val.Value, val2.Value))
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType, env.DecimalType}, "atan2")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType, env.DecimalType}, "atan2")
			}
		},
	},
	// Tests:
	// equal { atanh 0 } 0.0
	// Args:
	// * x: integer or decimal value between -1 and 1
	// Returns:
	// * decimal inverse hyperbolic tangent of x
	"atanh": {
		Argsn: 1,
		Doc:   "Returns the inverse hyperbolic tangent.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch val := arg0.(type) {
			case env.Integer:
				return *env.NewDecimal(math.Atanh(float64(val.Value)))
			case env.Decimal:
				return *env.NewDecimal(math.Atanh(val.Value))
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType, env.DecimalType}, "atanh")
			}
		},
	},
	// Tests:
	// equal { ceil 3.1 } 4.0
	// Args:
	// * x: integer or decimal value
	// Returns:
	// * decimal ceiling of x (smallest integer >= x)
	"ceil": {
		Argsn: 1,
		Doc:   "Returns the least integer value greater than or equal to x.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch val := arg0.(type) {
			case env.Integer:
				return *env.NewDecimal(float64(val.Value))
			case env.Decimal:
				return *env.NewDecimal(math.Ceil(val.Value))
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType, env.DecimalType}, "ceil")
			}
		},
	},
	// Tests:
	// equal { cbrt 8 } 2.0
	// Args:
	// * x: integer or decimal value
	// Returns:
	// * decimal cube root of x
	"cbrt": {
		Argsn: 1,
		Doc:   "Returns the cube root of x.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch val := arg0.(type) {
			case env.Integer:
				return *env.NewDecimal(math.Cbrt(float64(val.Value)))
			case env.Decimal:
				return *env.NewDecimal(math.Cbrt(val.Value))
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType, env.DecimalType}, "cbrt")
			}
		},
	},
	// Tests:
	// equal { copysign 1.0 -2.0 } -1.0
	// Args:
	// * x: integer or decimal value
	// * y: integer or decimal value
	// Returns:
	// * decimal with magnitude of x and sign of y
	"copysign": {
		Argsn: 2,
		Doc:   "Returns a value with the magnitude of x and the sign of y.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch val := arg0.(type) {
			case env.Integer:
				switch val2 := arg1.(type) {
				case env.Integer:
					return *env.NewDecimal(math.Copysign(float64(val.Value), float64(val2.Value)))
				case env.Decimal:
					return *env.NewDecimal(math.Copysign(float64(val.Value), val2.Value))
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType, env.DecimalType}, "copysign")
				}
			case env.Decimal:
				switch val2 := arg1.(type) {
				case env.Integer:
					return *env.NewDecimal(math.Copysign(val.Value, float64(val2.Value)))
				case env.Decimal:
					return *env.NewDecimal(math.Copysign(val.Value, val2.Value))
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType, env.DecimalType}, "copysign")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType, env.DecimalType}, "copysign")
			}
		},
	},
	// Tests:
	// equal { dim 5 3 } 2.0
	// equal { dim 3 5 } 0.0
	// Args:
	// * x: integer or decimal value
	// * y: integer or decimal value
	// Returns:
	// * decimal max(x-y, 0)
	"dim": {
		Argsn: 2,
		Doc:   "Returns the maximum of x-y or 0.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch val := arg0.(type) {
			case env.Integer:
				switch val2 := arg1.(type) {
				case env.Integer:
					return *env.NewDecimal(util.GetDimValue(float64(val.Value), float64(val2.Value)))
				case env.Decimal:
					return *env.NewDecimal(util.GetDimValue(float64(val.Value), val2.Value))
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType, env.DecimalType}, "dim")
				}
			case env.Decimal:
				switch val2 := arg1.(type) {
				case env.Integer:
					return *env.NewDecimal(util.GetDimValue(val.Value, float64(val2.Value)))
				case env.Decimal:
					return *env.NewDecimal(util.GetDimValue(val.Value, val2.Value))
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType, env.DecimalType}, "dim")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType, env.DecimalType}, "dim")
			}
		},
	},
	// Tests:
	// equal { round\to 3.14159 2 } 3.14
	// Args:
	// * x: decimal value to round
	// * digits: integer number of decimal places to round to
	// Returns:
	// * decimal rounded to specified number of decimal places
	"round\\to": {
		Argsn: 2,
		Doc:   "Rounds a decimal to the specified number of decimal places.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch val := arg0.(type) {
			case env.Decimal:
				switch precision := arg1.(type) {
				case env.Integer:
					ratio := math.Pow(10, float64(precision.Value))
					return *env.NewDecimal(math.Round(val.Value*ratio) / ratio)
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType, env.DecimalType}, "round\\to")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.DecimalType}, "round\\to")
			}
		},
	},
	// Tests:
	// equal { round 3.7 } 4.0
	// equal { round 3.2 } 3.0
	// Args:
	// * x: decimal value
	// Returns:
	// * decimal rounded to nearest integer
	"round": {
		Argsn: 1,
		Doc:   "Rounds a decimal to the nearest integer.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch val := arg0.(type) {
			case env.Decimal:
				return *env.NewDecimal(math.Round(val.Value))
			default:
				return MakeArgError(ps, 1, []env.Type{env.DecimalType}, "round\\to")
			}
		},
	},
	// Tests:
	// equal { roundtoeven 3.5 } 4.0
	// equal { roundtoeven 2.5 } 2.0
	// Args:
	// * x: integer or decimal value
	// Returns:
	// * decimal rounded to nearest integer, with ties rounded to even
	"roundtoeven": {
		Argsn: 1,
		Doc:   "Returns the nearest integer, rounding ties to even.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch val := arg0.(type) {
			case env.Decimal:
				return *env.NewDecimal(math.RoundToEven(val.Value))
			case env.Integer:
				return *env.NewDecimal(math.RoundToEven(float64(val.Value)))
			default:
				return MakeArgError(ps, 1, []env.Type{env.DecimalType, env.IntegerType}, "roundtoeven")
			}
		},
	},
	// Tests:
	// equal { erf 0 } 0.0
	// Args:
	// * x: integer or decimal value
	// Returns:
	// * decimal error function value of x
	"erf": {
		Argsn: 1,
		Doc:   "Returns the error function of value.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch val := arg0.(type) {
			case env.Integer:
				return *env.NewDecimal(math.Erf(float64(val.Value)))
			case env.Decimal:
				return *env.NewDecimal(math.Erf(val.Value))
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType, env.DecimalType}, "erf")
			}
		},
	},
	// Tests:
	// equal { erfc 0 } 1.0
	// Args:
	// * x: integer or decimal value
	// Returns:
	// * decimal complementary error function value of x
	"erfc": {
		Argsn: 1,
		Doc:   "Returns the complementary error function of value.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch val := arg0.(type) {
			case env.Integer:
				return *env.NewDecimal(math.Erfc(float64(val.Value)))
			case env.Decimal:
				return *env.NewDecimal(math.Erfc(val.Value))
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType, env.DecimalType}, "erfc")
			}
		},
	},
	// Tests:
	// equal { erfcinv 1.0 } 0.0
	// Args:
	// * x: integer or decimal value between 0 and 2
	// Returns:
	// * decimal inverse complementary error function value of x
	"erfcinv": {
		Argsn: 1,
		Doc:   "Returns the inverse of erfc(x) function.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch val := arg0.(type) {
			case env.Integer:
				return *env.NewDecimal(math.Erfcinv(float64(val.Value)))
			case env.Decimal:
				return *env.NewDecimal(math.Erfcinv(val.Value))
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType, env.DecimalType}, "erfcinv")
			}
		},
	},
	// Tests:
	// equal { erfinv 0 } 0.0
	// Args:
	// * x: integer or decimal value between -1 and 1
	// Returns:
	// * decimal inverse error function value of x
	"erfinv": {
		Argsn: 1,
		Doc:   "Returns the inverse error function of value.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch val := arg0.(type) {
			case env.Integer:
				return *env.NewDecimal(math.Erfinv(float64(val.Value)))
			case env.Decimal:
				return *env.NewDecimal(math.Erfinv(val.Value))
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType, env.DecimalType}, "erfinv")
			}
		},
	},
	// Tests:
	// equal { exp 0 } 1.0
	// equal { exp 1 } 2.718281828459045
	// equal { exp complex 0 0 |print } "1.000000+0.000000i"
	// equal { exp complex 0 3.141592653589793 |print } "-1.000000+0.000000i"
	// Args:
	// * x: integer, decimal, or complex value
	// Returns:
	// * decimal e^x for integer/decimal input, complex e^z for complex input
	"exp": {
		Argsn: 1,
		Doc:   "Returns e**x, the base-e exponential of x.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch val := arg0.(type) {
			case env.Integer:
				return *env.NewDecimal(math.Exp(float64(val.Value)))
			case env.Decimal:
				return *env.NewDecimal(math.Exp(val.Value))
			case env.Complex:
				return *env.NewComplex(cmplx.Exp(val.Value))
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType, env.DecimalType, env.ComplexType}, "exp")
			}
		},
	},
	// Tests:
	// equal { exp2 3 } 8.0
	// Args:
	// * x: integer or decimal value
	// Returns:
	// * decimal 2^x
	"exp2": {
		Argsn: 1,
		Doc:   "Returns 2**x, the base-2 exponential of x.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch val := arg0.(type) {
			case env.Integer:
				return *env.NewDecimal(math.Exp2(float64(val.Value)))
			case env.Decimal:
				return *env.NewDecimal(math.Exp2(val.Value))
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType, env.DecimalType}, "exp2")
			}
		},
	},
	// Tests:
	// equal { expm1 0 } 0.0
	// Args:
	// * x: integer or decimal value
	// Returns:
	// * decimal e^x - 1
	"expm1": {
		Argsn: 1,
		Doc:   "Returns e**x - 1, the base-e exponential of x minus 1. It is more accurate than exp(x) - 1 when x is near zero.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch val := arg0.(type) {
			case env.Integer:
				return *env.NewDecimal(math.Expm1(float64(val.Value)))
			case env.Decimal:
				return *env.NewDecimal(math.Expm1(val.Value))
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType, env.DecimalType}, "expm1")
			}
		},
	},
	// Tests:
	// equal { fma 2 3 4 } 10.0
	// Args:
	// * x: integer or decimal value
	// * y: integer or decimal value
	// * z: integer or decimal value
	// Returns:
	// * decimal (x * y) + z computed with only one rounding
	"fma": {
		Argsn: 3,
		Doc:   "Returns x * y + z, computed with only one rounding. (That is, FMA returns the fused multiply-add of x, y, and z.)",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch val1 := arg0.(type) {
			case env.Integer:
				switch val2 := arg1.(type) {
				case env.Integer:
					switch val3 := arg2.(type) {
					case env.Integer:
						return *env.NewDecimal(math.FMA(float64(val1.Value), float64(val2.Value), float64(val3.Value)))
					case env.Decimal:
						return *env.NewDecimal(math.FMA(float64(val1.Value), float64(val2.Value), val3.Value))
					default:
						return MakeArgError(ps, 3, []env.Type{env.IntegerType, env.DecimalType}, "fma")
					}
				case env.Decimal:
					switch val3 := arg2.(type) {
					case env.Integer:
						return *env.NewDecimal(math.FMA(float64(val1.Value), val2.Value, float64(val3.Value)))
					case env.Decimal:
						return *env.NewDecimal(math.FMA(float64(val1.Value), val2.Value, val3.Value))
					default:
						return MakeArgError(ps, 3, []env.Type{env.IntegerType, env.DecimalType}, "fma")
					}
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType, env.DecimalType}, "fma")
				}
			case env.Decimal:
				switch val2 := arg1.(type) {
				case env.Integer:
					switch val3 := arg2.(type) {
					case env.Integer:
						return *env.NewDecimal(math.FMA(val1.Value, float64(val2.Value), float64(val3.Value)))
					case env.Decimal:
						return *env.NewDecimal(math.FMA(val1.Value, float64(val2.Value), val3.Value))
					default:
						return MakeArgError(ps, 3, []env.Type{env.IntegerType, env.DecimalType}, "fma")
					}
				case env.Decimal:
					switch val3 := arg2.(type) {
					case env.Integer:
						return *env.NewDecimal(math.FMA(val1.Value, val2.Value, float64(val3.Value)))
					case env.Decimal:
						return *env.NewDecimal(math.FMA(val1.Value, val2.Value, val3.Value))
					default:
						return MakeArgError(ps, 3, []env.Type{env.IntegerType, env.DecimalType}, "fma")
					}
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType, env.DecimalType}, "fma")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType, env.DecimalType}, "fma")
			}
		},
	},
	// Tests:
	// equal { j0 0 } 1.0
	// Args:
	// * x: integer or decimal value
	// Returns:
	// * decimal order-zero Bessel function of the first kind
	"j0": {
		Argsn: 1,
		Doc:   "Returns the order-zero Bessel function of the first kind.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch val := arg0.(type) {
			case env.Integer:
				return *env.NewDecimal(math.J0(float64(val.Value)))
			case env.Decimal:
				return *env.NewDecimal(math.J0(val.Value))
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType, env.DecimalType}, "j0")
			}
		},
	},
	// Tests:
	// equal { j1 0 } 0.0
	// Args:
	// * x: integer or decimal value
	// Returns:
	// * decimal order-one Bessel function of the first kind
	"j1": {
		Argsn: 1,
		Doc:   "Returns the order-one Bessel function of the first kind.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch val := arg0.(type) {
			case env.Integer:
				return *env.NewDecimal(math.J1(float64(val.Value)))
			case env.Decimal:
				return *env.NewDecimal(math.J1(val.Value))
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType, env.DecimalType}, "j1")
			}
		},
	},
	// Tests:
	// equal { y0 1 } 0.08825696421567698
	// Args:
	// * x: integer or decimal value
	// Returns:
	// * decimal order-zero Bessel function of the second kind
	"y0": {
		Argsn: 1,
		Doc:   "Returns the order-zero Bessel function of the second kind.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch val := arg0.(type) {
			case env.Integer:
				return *env.NewDecimal(math.Y0(float64(val.Value)))
			case env.Decimal:
				return *env.NewDecimal(math.Y0(val.Value))
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType, env.DecimalType}, "y0")
			}
		},
	},
	// Tests:
	// equal { y1 1 } -0.7812128213002887
	// Args:
	// * x: integer or decimal value
	// Returns:
	// * decimal order-one Bessel function of the second kind
	"y1": {
		Argsn: 1,
		Doc:   "Returns the order-one Bessel function of the second kind.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch val := arg0.(type) {
			case env.Integer:
				return *env.NewDecimal(math.Y1(float64(val.Value)))
			case env.Decimal:
				return *env.NewDecimal(math.Y1(val.Value))
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType, env.DecimalType}, "y1")
			}
		},
	},
	// Tests:
	// equal { yn 2 1 } -1.6506826068162546
	// Args:
	// * n: integer order
	// * x: integer or decimal value
	// Returns:
	// * decimal order-n Bessel function of the second kind
	"yn": {
		Argsn: 2,
		Doc:   "Returns the order-n Bessel function of the second kind.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch v1 := arg0.(type) {
			case env.Integer:
				switch v2 := arg1.(type) {
				case env.Integer:
					return *env.NewDecimal(math.Yn(int(v1.Value), float64(v2.Value)))
				case env.Decimal:
					return *env.NewDecimal(math.Yn(int(v1.Value), v2.Value))
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType, env.DecimalType}, "yn")
				}
			case env.Decimal:
				switch v2 := arg1.(type) {
				case env.Integer:
					return *env.NewDecimal(math.Yn(int(v1.Value), float64(v2.Value)))
				case env.Decimal:
					return *env.NewDecimal(math.Yn(int(v1.Value), v2.Value))
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType, env.DecimalType}, "yn")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType, env.DecimalType}, "yn")
			}
		},
	},
	// Tests:
	// equal { jn 2 1 } 0.1149034849319005
	// Args:
	// * n: integer order
	// * x: integer or decimal value
	// Returns:
	// * decimal order-n Bessel function of the first kind
	"jn": {
		Argsn: 2,
		Doc:   "Returns the order-n Bessel function of the first kind.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch v1 := arg0.(type) {
			case env.Integer:
				switch v2 := arg1.(type) {
				case env.Integer:
					return *env.NewDecimal(math.Jn(int(v1.Value), float64(v2.Value)))
				case env.Decimal:
					return *env.NewDecimal(math.Jn(int(v1.Value), v2.Value))
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType, env.DecimalType}, "jn")
				}
			case env.Decimal:
				switch v2 := arg1.(type) {
				case env.Integer:
					return *env.NewDecimal(math.Jn(int(v1.Value), float64(v2.Value)))
				case env.Decimal:
					return *env.NewDecimal(math.Jn(int(v1.Value), v2.Value))
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType, env.DecimalType}, "jn")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType, env.DecimalType}, "jn")
			}
		},
	},
	// Tests:
	// equal { trunc 3.7 } 3.0
	// Args:
	// * x: integer or decimal value
	// Returns:
	// * decimal integer value of x (truncated toward zero)
	"trunc": {
		Argsn: 1,
		Doc:   "Returns the integer value of input.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch val := arg0.(type) {
			case env.Integer:
				return *env.NewDecimal(float64(val.Value))
			case env.Decimal:
				return *env.NewDecimal(math.Trunc(val.Value))
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType, env.DecimalType}, "trunc")
			}
		},
	},
	// Tests:
	// equal { round\to pi 5 } 3.14159
	// Args:
	// * none
	// Returns:
	// * decimal value of π (pi)
	"pi": {
		Argsn: 0,
		Doc:   "Returns the mathematical constant π (pi).",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewDecimal(float64(math.Pi))
		},
	},
	// Tests:
	// equal { deg->rad 180 } 3.141592653589793
	// Args:
	// * degrees: integer or decimal value in degrees
	// Returns:
	// * decimal value in radians
	"deg->rad": {
		Argsn: 1,
		Doc:   "Converts degrees to radians.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			var fa float64
			switch a := arg0.(type) {
			case env.Decimal:
				fa = a.Value
			case env.Integer:
				fa = float64(a.Value)
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType, env.DecimalType}, "deg->rad")
			}
			return *env.NewDecimal(fa * float64(math.Pi) / 180.0)
		},
	},
	// Tests:
	// equal { is-near 0.0000000000001 0 } 1
	// equal { is-near 0.1 0 } 0
	// Args:
	// * x: integer or decimal value
	// * y: integer or decimal value
	// Returns:
	// * integer 1 if values are very close, 0 otherwise
	"is-near": {
		Argsn: 2,
		Doc:   "Returns true if two decimals are close.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			fa, fb, errPos := assureFloats(arg0, arg1)
			if errPos > 0 {
				return MakeArgError(ps, errPos, []env.Type{env.IntegerType, env.BlockType}, "is-near")
			}
			const epsilon = 0.0000000000001 // math.SmallestNonzeroFloat64
			if math.Abs(fa-fb) <= (epsilon) {
				return env.NewInteger(1)
			} else {
				return env.NewInteger(0)
			}
		},
	},
	// Tests:
	// equal { near-zero 0.0000000000001 } 1
	// equal { near-zero 0.1 } 0
	// Args:
	// * x: integer or decimal value
	// Returns:
	// * integer 1 if value is very close to zero, 0 otherwise
	"near-zero": {
		Argsn: 1,
		Doc:   "Returns true if a decimal is close to zero.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			var fa float64
			switch a := arg0.(type) {
			case env.Decimal:
				fa = a.Value
			case env.Integer:
				fa = float64(a.Value)
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType, env.BlockType}, "near-zero")
			}
			// const epsilon = math.SmallestNonzeroFloat64
			const epsilon = 0.0000000000001 // math.SmallestNonzeroFloat64
			if math.Abs(fa) <= epsilon {
				return env.NewInteger(1)
			} else {
				return env.NewInteger(0)
			}
		},
	},
	// Tests:
	// equal { is-prime 7 } 1
	// equal { is-prime 10 } 0
	// Args:
	// * n: integer value to check
	// Returns:
	// * integer 1 if n is prime, 0 otherwise
	"is-prime": {
		Argsn: 1,
		Doc:   "Returns true if the integer is a prime number.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch val := arg0.(type) {
			case env.Integer:
				if val.Value <= 1 {
					return *env.NewBoolean(false) // 0 and 1 are not prime numbers
				}
				if primes.IsPrime(int(val.Value)) {
					return *env.NewBoolean(true)
				} else {
					return *env.NewBoolean(false)
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType}, "is-prime")
			}
		},
	},
	// Tests:
	// equal { to-eyr { 1 + 2 * 3 } |type? } 'block
	// Args:
	// * block: block containing math expressions
	// Returns:
	// * block converted to Eyr dialect
	"to-eyr": {
		Argsn: 1,
		Doc:   "Converts math dialect to Eyr dialect.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return DialectMath(ps, arg0)
		},
	},
	// Tests:
	// equal { calc { 1 + 2 * 3 } } 7
	// Args:
	// * block: block containing math expressions
	// Returns:
	// * result of evaluating the math expressions
	"calc": {
		Argsn: 1,
		Doc:   "Evaluates expressions in math dialect.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			res := DialectMath(ps, arg0)
			switch block := res.(type) {
			case env.Block:
				// stack := env.NewEyrStack() // TODO -- stack moved to PS ... look it up if it requires changes here
				ps.ResetStack()
				ser := ps.Ser
				ps.Ser = block.Series
				Eyr_EvalBlock(ps, false)
				ps.Ser = ser
				return ps.Res
			default:
				return res
			}
		},
	},
}
