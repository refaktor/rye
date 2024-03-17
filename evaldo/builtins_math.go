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

var Builtins_math = map[string]*env.Builtin{

	"mod": {
		Argsn: 1,
		Doc:   "Return a decimal remainder.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch a := arg0.(type) {
			case env.Integer:
				switch b := arg0.(type) {
				case env.Integer:
					return *env.NewDecimal(float64(math.Mod(float64(a.Value), float64(b.Value))))
				}
			case env.Decimal:
				switch b := arg0.(type) {
				case env.Decimal:
					return *env.NewDecimal(float64(math.Mod(float64(a.Value), float64(b.Value))))
				}
			}
			return nil // TODO
		},
	},
	"log2": {
		Argsn: 1,
		Doc:   "Return binary logarithm of x",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch val := arg0.(type) {
			case env.Integer:
				return *env.NewDecimal(float64(math.Log2(float64(val.Value))))
			}
			return nil // TODO
		},
	},
	"sin": {
		Argsn: 1,
		Doc:   "Return the sine of the radian argument.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch val := arg0.(type) {
			case env.Integer:
				return *env.NewDecimal(float64(math.Sin(float64(val.Value))))
			}
			return nil // TODO
		},
	},
	"cos": {
		Argsn: 1,
		Doc:   "Return the cosine of the radian argument.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch val := arg0.(type) {
			case env.Integer:
				return *env.NewDecimal(float64(math.Cos(float64(val.Value))))
			}
			return nil // TODO
		},
	},
	"sqrt": {
		Argsn: 1,
		Doc:   "Return the square root of x.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch val := arg0.(type) {
			case env.Integer:
				return *env.NewDecimal(float64(math.Sqrt(float64(val.Value))))
			case env.Decimal:
				return *env.NewDecimal(float64(math.Sqrt(float64(val.Value))))
			}
			return nil // TODO
		},
	},
	"abs": {
		Argsn: 1,
		Doc:   "Return absolute value.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch val := arg0.(type) {
			case env.Integer:
				return *env.NewDecimal(float64(math.Abs(float64(val.Value))))
			}
			return nil // TODO
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
