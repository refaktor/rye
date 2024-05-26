package evaldo

import (
	"fmt"
	"math"

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

	"mod": {
		Argsn: 2,
		Doc:   "Return a decimal remainder",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			fa, fb, errPos := assureFloats(arg0, arg1)
			if errPos > 0 {
				return MakeArgError(ps, errPos, []env.Type{env.IntegerType, env.BlockType}, "mod")
			}
			return *env.NewDecimal(math.Mod(fa, fb))
		},
	},
	"log2": {
		Argsn: 1,
		Doc:   "Return binary logarithm of x",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch val := arg0.(type) {
			case env.Integer:
				return *env.NewDecimal(math.Log2(float64(val.Value)))
			case env.Decimal:
				return *env.NewDecimal(math.Log2(val.Value))
			default:
				return MakeArgError(ps, 2, []env.Type{env.IntegerType, env.BlockType}, "mod")
			}
		},
	},
	"sin": {
		Argsn: 1,
		Doc:   "Return the sine of the radian argument.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch val := arg0.(type) {
			case env.Integer:
				return *env.NewDecimal(math.Sin(float64(val.Value)))
			case env.Decimal:
				return *env.NewDecimal(math.Sin(val.Value))
			default:
				return MakeArgError(ps, 2, []env.Type{env.IntegerType, env.BlockType}, "mod")
			}
		},
	},
	"cos": {
		Argsn: 1,
		Doc:   "Return the cosine of the radian argument.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch val := arg0.(type) {
			case env.Integer:
				return *env.NewDecimal(math.Cos(float64(val.Value)))
			case env.Decimal:
				return *env.NewDecimal(math.Cos(val.Value))
			default:
				return MakeArgError(ps, 2, []env.Type{env.IntegerType, env.BlockType}, "mod")
			}
		},
	},
	"sqrt": {
		Argsn: 1,
		Doc:   "Return the square root of x.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch val := arg0.(type) {
			case env.Integer:
				return *env.NewDecimal(math.Sqrt(float64(val.Value)))
			case env.Decimal:
				return *env.NewDecimal(math.Sqrt(val.Value))
			default:
				return MakeArgError(ps, 2, []env.Type{env.IntegerType, env.BlockType}, "mod")
			}
		},
	},
	"abs": {
		Argsn: 1,
		Doc:   "Return absolute value.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch val := arg0.(type) {
			case env.Integer:
				return *env.NewDecimal(math.Abs(float64(val.Value)))
			case env.Decimal:
				return *env.NewDecimal(math.Abs(val.Value))
			default:
				return MakeArgError(ps, 2, []env.Type{env.IntegerType, env.BlockType}, "mod")
			}
		},
	},
	"acos": {
		Argsn: 1,
		Doc:   "Returns the arccosine.",
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
	"asin": {
		Argsn: 1,
		Doc:   "Returns the arcsine (inverse sine).",
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
	"atan": {
		Argsn: 1,
		Doc:   "Returns the arctangent (inverse tangent).",
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
	"cbrt": {
		Argsn: 1,
		Doc:   "Returns returns the cube root of x.",
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
	"copysign": {
		Argsn: 2,
		Doc:   "Copysign returns a value with the magnitude of arg1 and the sign of arg2.",
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
	"dim": {
		Argsn: 2,
		Doc:   "Dim returns the maximum of arg1-arg2 or 0.",
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
	"pi": {
		Argsn: 0,
		Doc:   "Return Pi constant.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewDecimal(float64(math.Pi))
		},
	},
	"is-near": {
		Argsn: 2,
		Doc:   "Returns true if two decimals are close.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			fa, fb, errPos := assureFloats(arg0, arg1)
			if errPos > 0 {
				return MakeArgError(ps, errPos, []env.Type{env.IntegerType, env.BlockType}, "equals")
			}
			const epsilon = 0.0000000000001 // math.SmallestNonzeroFloat64
			if math.Abs(fa-fb) <= (epsilon) {
				return env.NewInteger(1)
			} else {
				return env.NewInteger(0)
			}
		},
	},
	"near-zero": {
		Argsn: 1,
		Doc:   "Returns true if two decimals are close.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			var fa float64
			switch a := arg0.(type) {
			case env.Decimal:
				fa = a.Value
			case env.Integer:
				fa = float64(a.Value)
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType, env.BlockType}, "is-zero")
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
	"to-eyr": {
		Argsn: 1,
		Doc:   "Math dialect to Eyr dialect",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return DialectMath(ps, arg0)
		},
	},
	"calc": {
		Argsn: 1,
		Doc:   "Do math dialect",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			res := DialectMath(ps, arg0)
			switch block := res.(type) {
			case env.Block:
				stack := NewEyrStack()
				ser := ps.Ser
				ps.Ser = block.Series
				Eyr_EvalBlock(ps, stack)
				ps.Ser = ser
				return ps.Res
			default:
				return res
			}
		},
	},
}
