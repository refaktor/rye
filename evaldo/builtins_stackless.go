// builtins.go
package evaldo

import (
	"fmt"

	"github.com/refaktor/rye/env"
)

// definiraj frame <builtin nargs arg0 arg1>
// definiraj stack []evalframe
// callbui kreira trenuten frame, nastavi bui nargs in vrne
// while loop pogleda naslednji arg, če je literal nastavi arg in poveča argc če je argc nargs potem pokliče frame in iz stacka potegne naslednjega, če ni potem zalopa
// 									če je builtin potem pusha trenuten frame na stack in kreira novega

func Stck_CallBuiltin(bi env.Builtin, ps *env.ProgramState, arg0_ env.Object, toLeft bool) *env.ProgramState {
	evalExprFn := EvalExpression2
	arg0 := bi.Cur0 //env.Object(bi.Cur0)
	arg1 := bi.Cur1

	if bi.Argsn > 0 && bi.Cur0 == nil {
		//fmt.Println(" ARG 1 ")
		//fmt.Println(ps.Ser.GetPos())
		evalExprFn(ps, true)

		if checkFlagsBi(bi, ps, 0) {
			return ps
		}
		if ps.ErrorFlag || ps.ReturnFlag {
			return ps
		}
		arg0 = ps.Res
	}
	if bi.Argsn > 1 && bi.Cur1 == nil {
		evalExprFn(ps, true) // <---- THESE DETERMINE IF IT CONSUMES WHOLE EXPRESSION OR NOT IN CASE OF PIPEWORDS .. HM*... MAYBE WOULD COULD HAVE A WORD MODIFIER?? a: 2 |add 5 a:: 2 |add 5 print* --TODO

		if checkFlagsBi(bi, ps, 1) {
			return ps
		}
		if ps.ErrorFlag || ps.ReturnFlag {
			return ps
		}
		//fmt.Println(ps.Res)

		arg1 = ps.Res
	}
	ps.Res = bi.Fn(ps, arg0, arg1, nil, nil, nil)
	return ps
}

func Stck_EvalObject(ps *env.ProgramState, object env.Object, leftVal env.Object, toLeft bool, ctx *env.RyeCtx) *env.ProgramState {
	//fmt.Print("EVAL OBJECT")
	switch object.Type() {
	case env.BuiltinType:
		//fmt.Println(" BUIL**")
		//fmt.Println(ps.Ser.GetPos())
		//fmt.Println(" BUILTIN**")
		bu := object.(env.Builtin)

		// OBJECT INJECTION EXPERIMENT
		// ps.Ser.Put(bu)

		if checkFlagsBi(bu, ps, 333) {
			return ps
		}
		return Stck_CallBuiltin(bu, ps, leftVal, toLeft)

		//ps.Res.Trace("After builtin call")
		//return ps
	default:
		//d object.Trace("DEFAULT**")
		ps.Res = object
		//ps.Res.Trace("After object returned")
		return ps
	}
}

func Stck_EvalWord(ps *env.ProgramState, word env.Object, leftVal env.Object, toLeft bool) *env.ProgramState {
	// LOCAL FIRST
	found, object, ctx := findWordValue(ps, word)
	if found {
		trace("****33")
		return Stck_EvalObject(ps, object, leftVal, toLeft, ctx) //ww0128a *
		//ps.Res.Trace("After eval Object")
		//return ps
	} else {
		trace("****34")
		ps.ErrorFlag = true
		if !ps.FailureFlag {
			ps.Res = *env.NewError2(5, "Word not found: "+word.Inspect(*ps.Idx))
			fmt.Println("Error: Not known type")
		}
		return ps
	}
}

func Stck_EvalExpression(ps *env.ProgramState) *env.ProgramState {
	object := ps.Ser.Pop()
	trace("Before entering expression")
	if object != nil {
		switch object.Type() {
		case env.IntegerType:
			ps.Res = object
		case env.StringType:
			ps.Res = object
		case env.BlockType:
			ps.Res = object
		case env.WordType:
			rr := Stck_EvalWord(ps, object.(env.Word), nil, false)
			return rr
		default:
			ps.ErrorFlag = true
			ps.Res = env.NewError("Not known type")
			fmt.Println("Error: Not known type")
		}
	} else {
		ps.ErrorFlag = true
		ps.Res = env.NewError("Not known type")
		fmt.Println("Error: Not known type")
	}

	return ps
}

func Stck_EvalBlock(ps *env.ProgramState) *env.ProgramState {
	for ps.Ser.Pos() < ps.Ser.Len() {
		ps = Stck_EvalExpression(ps)
		if tryHandleFailure(ps) {
			return ps
		}
		if ps.ReturnFlag || ps.ErrorFlag {
			return ps
		}
	}
	return ps
}

var Builtins_stackless = map[string]*env.Builtin{

	"ry0": {
		Argsn: 1,
		Doc:   "TODODOC",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch bloc := arg0.(type) {
			case env.Block:
				ser := ps.Ser
				ps.Ser = bloc.Series
				Stck_EvalBlock(ps)
				ps.Ser = ser
				return ps.Res
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType}, "ry0")
			}
		},
	},

	"ry0-loop": {
		Argsn: 2,
		Doc:   "TODODOC",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch cond := arg0.(type) {
			case env.Integer:
				switch bloc := arg1.(type) {
				case env.Block:
					ser := ps.Ser
					ps.Ser = bloc.Series
					for i := 0; int64(i) < cond.Value; i++ {
						ps = Stck_EvalBlock(ps)
						ps.Ser.Reset()
					}
					ps.Ser = ser
					return ps.Res
				default:
					return MakeArgError(ps, 1, []env.Type{env.BlockType}, "ry0-loop")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType}, "ry0-loop")
			}
		},
	},
}
