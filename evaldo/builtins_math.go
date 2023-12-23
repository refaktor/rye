package evaldo

import (
	"fmt"

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
		case env.Block:
			ser1 := es.Ser
			es.Ser = obj.Series
			val := Math_EvalBlock(es)
			es.Ser = ser1
			output = append(output, val...)
		case env.Opword:
			for len(operators) > 0 && precedence[operators[len(operators)-1].Index] >= precedence[obj.Index] {
				output = append(output, operators[len(operators)-1])
				operators = operators[:len(operators)-1]
			}
			operators = append(operators, obj)
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

	"math": {
		Argsn: 1,
		Doc:   "Math dialect",
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return DialectMath(env1, arg0)
		},
	},
}
