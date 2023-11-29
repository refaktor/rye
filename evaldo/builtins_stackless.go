// builtins.go
package evaldo

import (
	"rye/env"
)

// definiraj frame <builtin nargs arg0 arg1>
// definiraj stack []evalframe
// callbui kreira trenuten frame, nastavi bui nargs in vrne
// while loop pogleda naslednji arg, če je literal nastavi arg in poveča argc če je argc nargs potem pokliče frame in iz stacka potegne naslednjega, če ni potem zalopa
// 									če je builtin potem pusha trenuten frame na stack in kreira novega

func Stck_CallBuiltin(bi env.Builtin, ps *env.ProgramState, arg0_ env.Object, toLeft bool) *env.ProgramState {

	evalExprFn := EvalExpression2
	arg0 := env.Object(bi.Cur0) //env.Object(bi.Cur0)
	arg1 := env.Object(bi.Cur1)

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

func Stck_EvalObject(es *env.ProgramState, object env.Object, leftVal env.Object, toLeft bool, ctx *env.RyeCtx) *env.ProgramState {
	//fmt.Print("EVAL OBJECT")
	switch object.Type() {
	case env.BuiltinType:
		//fmt.Println(" BUIL**")
		//fmt.Println(es.Ser.GetPos())
		//fmt.Println(" BUILTIN**")
		bu := object.(env.Builtin)

		// OBJECT INJECTION EXPERIMENT
		// es.Ser.Put(bu)

		if checkFlagsBi(bu, es, 333) {
			return es
		}
		return Stck_CallBuiltin(bu, es, leftVal, toLeft)

		//es.Res.Trace("After builtin call")
		//return es
	default:
		//d object.Trace("DEFAULT**")
		es.Res = object
		//es.Res.Trace("After object returned")
		return es
	}
}

func Stck_EvalWord(es *env.ProgramState, word env.Object, leftVal env.Object, toLeft bool) *env.ProgramState {
	// LOCAL FIRST
	found, object, ctx := findWordValue(es, word)
	if found {
		trace("****33")
		return Stck_EvalObject(es, object, leftVal, toLeft, ctx) //ww0128a *
		//es.Res.Trace("After eval Object")
		//return es
	} else {
		trace("****34")
		es.ErrorFlag = true
		if es.FailureFlag == false {
			es.Res = *env.NewError2(5, "Word not found: "+word.Inspect(*es.Idx))
		}
		return es
	}
}

func Stck_EvalExpression(es *env.ProgramState) *env.ProgramState {
	object := es.Ser.Pop()
	//es.Idx.Probe()
	trace2("Before entering expression")
	if object != nil {
		switch object.Type() {
		case env.IntegerType:
			es.Res = object
		case env.StringType:
			es.Res = object
		case env.BlockType:
			es.Res = object
		case env.WordType:
			rr := Stck_EvalWord(es, object.(env.Word), nil, false)
			return rr
		default:
			es.ErrorFlag = true
			es.Res = env.NewError("Not known type")
		}
	} else {
		es.ErrorFlag = true
		es.Res = env.NewError("Not known type")
	}

	return es
}

func Stck_EvalBlock(es *env.ProgramState) *env.ProgramState {
	for es.Ser.Pos() < es.Ser.Len() {
		es = Stck_EvalExpression(es)
		if checkFlagsAfterBlock(es, 101) {
			return es
		}
		if es.ReturnFlag || es.ErrorFlag {
			return es
		}
	}
	return es
}

var Builtins_stackless = map[string]*env.Builtin{

	"ry0": {
		Argsn: 1,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch bloc := arg0.(type) {
			case env.Block:
				ser := ps.Ser
				ps.Ser = bloc.Series
				Stck_EvalBlock(ps)
				ps.Ser = ser
				return ps.Res
			}
			return nil
		},
	},

	"ry0-loop": {
		Argsn: 2,
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
				}
			}
			return nil
		},
	},
}
