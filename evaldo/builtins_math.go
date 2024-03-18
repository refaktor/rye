package evaldo

import (
	"fmt"
	"math"

	"github.com/refaktor/rye/env"
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
	"math": {
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
