// builtins.go
package evaldo

import (
	// "fmt"

	"fmt"

	"github.com/refaktor/rye/env"
)

// definiraj frame <builtin nargs arg0 arg1>
// definiraj stack []evalframe
// callbui kreira trenuten frame, nastavi bui nargs in vrne
// while loop pogleda naslednji arg, če je literal nastavi arg in poveča argc če je argc nargs potem pokliče frame in iz stacka potegne naslednjega, če ni potem zalopa
// 									če je builtin potem pusha trenuten frame na stack in kreira novega

func Eyr_CallBuiltinPipe(bi env.Builtin, ps *env.ProgramState, arg0_ env.Object) *env.ProgramState {
	//arg0 := bi.Cur0     //env.Object(bi.Cur0)
	//var arg1 env.Object // := bi.Cur1
	//var arg2 env.Object

	// for now works just with functions taht accept block as a first and only argument ... will have to conceptualize other options first

	if bi.Argsn > 0 && bi.Cur0 == nil {
		if checkFlagsBi(bi, ps, 0) {
			return ps
		}
		if ps.ErrorFlag || ps.ReturnFlag {
			return ps
		}
		if ps.ErrorFlag {
			return ps
		}
		block := stackToBlock(ps.Stack)
		if bi.Argsn == 1 {
			fmt.Println("** CALL BI")
			ps.Res = bi.Fn(ps, block, nil, nil, nil, nil)
			// stack.Push(ps.Res)
		}
	}
	return ps
}

func Eyr_CallBuiltin(bi env.Builtin, ps *env.ProgramState, arg0_ env.Object, toLeft bool) *env.ProgramState {
	arg0 := bi.Cur0     //env.Object(bi.Cur0)
	var arg1 env.Object // := bi.Cur1
	var arg2 env.Object

	if bi.Argsn > 0 && bi.Cur0 == nil {
		if checkFlagsBi(bi, ps, 0) {
			return ps
		}
		if ps.ErrorFlag || ps.ReturnFlag {
			return ps
		}
		arg0 = ps.Stack.Pop(ps)
		if ps.ErrorFlag {
			return ps
		}
		if bi.Argsn == 1 {
			fmt.Println("** CALL BI")
			ps.Res = bi.Fn(ps, arg0, nil, nil, nil, nil)
			// stack.Push(ps.Res)
		}
	}
	if bi.Argsn > 1 && bi.Cur1 == nil {
		if checkFlagsBi(bi, ps, 1) {
			return ps
		}
		if ps.ErrorFlag || ps.ReturnFlag {
			return ps
		}

		arg1 = ps.Stack.Pop(ps)
		if ps.ErrorFlag {
			return ps
		}
		if bi.Argsn == 2 {
			ps.Res = bi.Fn(ps, arg1, arg0, nil, nil, nil)
			// stack.Push(ps.Res)
		}
	}
	if bi.Argsn > 2 && bi.Cur2 == nil {
		if checkFlagsBi(bi, ps, 0) {
			return ps
		}
		if ps.ErrorFlag || ps.ReturnFlag {
			return ps
		}

		arg2 = ps.Stack.Pop(ps)
		if ps.ErrorFlag {
			return ps
		}
		if bi.Argsn == 3 {
			ps.Res = bi.Fn(ps, arg2, arg1, arg0, nil, nil)
			//stack.Push(ps.Res)
		}
	}
	return ps
}

// This is separate from CallFuncitonArgsN so it can manage pulling args directly off of the eyr stack
func Eyr_CallFunction(fn env.Function, es *env.ProgramState, leftVal env.Object, toLeft bool, session *env.RyeCtx) *env.ProgramState {
	var fnCtx = DetermineContext(fn, es, session)
	if checkErrorReturnFlag(es) {
		return es
	}

	var arg0 env.Object = nil
	for i := fn.Argsn - 1; i >= 0; i-- {
		var stackElem = es.Stack.Pop(es)
		// TODO: Consider doing check once outside of loop once this version is ready as a correctness comparison point
		if es.ErrorFlag {
			return es
		}
		if arg0 == nil {
			arg0 = stackElem
		}
		fnCtx.Set(fn.Spec.Series.Get(i).(env.Word).Index, stackElem)
	}

	tempCtx := es.Ctx
	tempSer := es.Ser

	fn.Body.Series.Reset()

	es.Ctx = fnCtx
	es.Ser = fn.Body.Series
	// setup
	/* psX := env.NewProgramState(fn.Body.Series, es.Idx)
	psX.Ctx = fnCtx
	psX.PCtx = es.PCtx
	psX.Gen = es.Gen
	psX.Dialect = es.Dialect
	psX.Stack = es.Stack

	var result *env.ProgramState */
	// es.Ser.SetPos(0)
	// fmt.Println("***")
	if fn.Argsn > 0 {
		EvalBlock(es)
		// EvalBlockInjMultiDialect(es, arg0, arg0 != nil)
	} else {
		EvalBlock(es)
	}
	// MaybeDisplayFailureOrError(result, result.Idx)

	/* if result.ForcedResult != nil {
		es.Res = result.ForcedResult
		result.ForcedResult = nil
	} else {
		es.Res = result.Res
	}
	es.Stack.Push(es, es.Res) */
	es.Ser = tempSer
	es.Ctx = tempCtx
	es.ReturnFlag = false
	return es
}

func Eyr_EvalObject(es *env.ProgramState, object env.Object, leftVal env.Object, pipeWord bool, session *env.RyeCtx, bakein bool) *env.ProgramState {
	//fmt.Print("EVAL OBJECT")
	switch object.Type() {
	case env.BuiltinType:
		bu := object.(env.Builtin)
		if bakein {
			es.Ser.Put(bu)
		} //es.Ser.SetPos(es.Ser.Pos() - 1)
		if checkFlagsBi(bu, es, 333) {
			return es
		}
		if pipeWord {
			es = Eyr_CallBuiltinPipe(bu, es, leftVal)
		} else {
			es = Eyr_CallBuiltin(bu, es, leftVal, false)
		}
		if es.Res != nil && es.Res.Type() != env.VoidType {
			es.Stack.Push(es, es.Res)
		}
		return es
	case env.FunctionType:
		fn := object.(env.Function)
		return Eyr_CallFunction(fn, es, leftVal, pipeWord, session)
	default:
		es.Res = object
		es.Stack.Push(es, es.Res)
		return es
	}
}

func Eyr_EvalWord(es *env.ProgramState, word env.Object, leftVal env.Object, pipeWord bool) *env.ProgramState {
	// LOCAL FIRST
	found, object, ctx := findWordValue(es, word)
	if found {
		es = Eyr_EvalObject(es, object, leftVal, pipeWord, ctx, true) //ww0128a *
		// es.Stack.Push(es, ¸.Res)
		return es
	} else {
		es.ErrorFlag = true
		if !es.FailureFlag {
			es.Res = *env.NewError2(5, "Word not found: "+word.Inspect(*es.Idx))
		}
		return es
	}
}

func Eyr_EvalLSetword(ps *env.ProgramState, word env.LSetword, leftVal env.Object, toLeft bool) *env.ProgramState {
	idx := word.Index
	val := ps.Stack.Pop(ps)
	if ps.ErrorFlag {
		return ps
	}
	ps.Ctx.Mod(idx, val)
	return ps
}

func Eyr_EvalExpression(ps *env.ProgramState) *env.ProgramState {
	object := ps.Ser.Pop()
	trace2("Before entering expression")
	if object != nil {
		switch object.Type() {
		case env.IntegerType:
			// fmt.Println("** INTEGER")
			ps.Stack.Push(ps, object)
		case env.DecimalType:
			ps.Stack.Push(ps, object)
		case env.StringType:
			ps.Stack.Push(ps, object)
		case env.BlockType:
			ps.Stack.Push(ps, object)
		case env.UriType:
			ps.Stack.Push(ps, object)
		case env.EmailType:
			ps.Stack.Push(ps, object)
		case env.WordType:
			// fmt.Println("** WORD")
			rr := Eyr_EvalWord(ps, object.(env.Word), nil, false)
			return rr
		case env.OpwordType: // + and other operators are basically opwords too
			// fmt.Println("** OPWORD")
			rr := Eyr_EvalWord(ps, object.(env.Opword), nil, false)
			return rr
		case env.PipewordType: // + and other operators are basically opwords too
			// fmt.Println("** OPWORD")
			rr := Eyr_EvalWord(ps, object.(env.Pipeword), nil, true)
			return rr
		case env.CPathType:
			rr := Eyr_EvalWord(ps, object.(env.CPath), nil, false)
			return rr
		case env.LSetwordType:
			// print(stack)
			rr := Eyr_EvalLSetword(ps, object.(env.LSetword), nil, false)
			return rr
		case env.BuiltinType:
			// fmt.Println("** BUILTIN")
			return Eyr_EvalObject(ps, object, nil, false, nil, false) //ww0128a *
		default:
			ps.ErrorFlag = true
			ps.Res = env.NewError("Not known type for Eyr")
		}
	} else {
		ps.ErrorFlag = true
		ps.Res = env.NewError("Not known type (nil)")
	}

	return ps
}

func Eyr_EvalBlockInside(ps *env.ProgramState) *env.ProgramState {
	fmt.Println("** EVALB INSIDE")
	for ps.Ser.Pos() < ps.Ser.Len() {
		fmt.Println(ps.Ser.Pos())
		ps = Eyr_EvalExpression(ps)
		if checkFlagsAfterBlock(ps, 101) {
			fmt.Println("yy")
			return ps
		}
		if ps.ReturnFlag || ps.ErrorFlag {
			fmt.Println(ps.ReturnFlag)
			fmt.Println(ps.ErrorFlag)
			fmt.Println("xx")
			return ps
		}
	}
	fmt.Println("** EVAL BLOCK PS RES")
	fmt.Println(ps.Res)
	ps.Res = env.NewVoid()
	return ps
}

func Eyr_EvalBlock(ps *env.ProgramState, full bool) *env.ProgramState {
	fmt.Println("** EVALB")
	for ps.Ser.Pos() < ps.Ser.Len() {
		fmt.Println(ps.Ser.Pos())
		ps = Eyr_EvalExpression(ps)
		if checkFlagsAfterBlock(ps, 101) {
			fmt.Println("yy")
			return ps
		}
		if ps.ReturnFlag || ps.ErrorFlag {
			fmt.Println(ps.ReturnFlag)
			fmt.Println(ps.ErrorFlag)
			fmt.Println("xx")
			return ps
		}
	}
	if full {
		ps.Res = stackToBlock(ps.Stack)
	} else {
		ps.Res = ps.Stack.Peek(ps, 0)
	}
	fmt.Println("** EVAL BLOCK PS RES")
	fmt.Println(ps.Res)
	return ps
}

func stackToBlock(stack *env.EyrStack) env.Block {
	return *env.NewBlock(*env.NewTSeries(stack.D[0:stack.I]))
}

var Builtins_eyr = map[string]*env.Builtin{

	"eyr": {
		Argsn: 1,
		Doc:   "Evaluates Rye block as Eyr (postfix) stack based code.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch bloc := arg0.(type) {
			case env.Block:
				ser := ps.Ser
				ps.Ser = bloc.Series
				ps.Dialect = env.EyrDialect
				Eyr_EvalBlock(ps, false)
				ps.Ser = ser
				return ps.Res
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType}, "eyr")
			}
		},
	},

	"eyr\\full": {
		Argsn: 1,
		Doc:   "Evaluates Rye block as Eyr (postfix) stack based code.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch bloc := arg0.(type) {
			case env.Block:
				ser := ps.Ser
				ps.Ser = bloc.Series
				ps.Dialect = env.EyrDialect
				Eyr_EvalBlock(ps, true)
				ps.Ser = ser
				return ps.Res
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType}, "eyr\\full")
			}
		},
	},

	"eyr\\loop": {
		Argsn: 2,
		Doc:   "Evaluates Rye block in loop as Eyr code (postfix stack based) N times.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch cond := arg0.(type) {
			case env.Integer:
				switch bloc := arg1.(type) {
				case env.Block:
					ps.Dialect = env.EyrDialect
					ser := ps.Ser
					ps.Ser = bloc.Series
					for i := 0; int64(i) < cond.Value; i++ {
						ps = Eyr_EvalBlock(ps, false)
						ps.Ser.Reset()
					}
					ps.Ser = ser
					return ps.Res
				default:
					return MakeArgError(ps, 2, []env.Type{env.BlockType}, "eyr\\loop")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType}, "eyr\\loop")
			}
		},
	},
	"to-eyr": {
		Argsn: 1,
		Doc:   "Evaluates Rye block as Eyr (postfix) stack based code.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch bloc := arg0.(type) {
			case env.Block:
				eBlock := env.NewBlock(*env.NewTSeries(make([]env.Object, 0)))
				CompileRyeToEyr(&bloc, ps, eBlock)
				return *eBlock
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType}, "eyr")
			}
		},
	},
}
