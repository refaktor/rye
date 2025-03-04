package evaldo

import (
	"fmt"
	"strconv"

	"github.com/refaktor/rye/env"
)

// Descr: Main evaluator of a block that injects a value
func Rye0_EvalBlockInj(ps *env.ProgramState, inj env.Object, injnow bool) *env.ProgramState {
	// repeats until at the end of the block
	for ps.Ser.Pos() < ps.Ser.Len() {
		// evaluate expression at the block cursor
		ps, injnow = Rye0_EvalExpressionInj(ps, inj, injnow)
		// If flags raised return program state
		if checkFlagsAfterBlock(ps, 101) {
			return ps
		}
		if checkErrorReturnFlag(ps) {
			return ps
		}
		//ps, injnow = MaybeAcceptComma(ps, inj, injnow)
	}
	return ps
}

func Rye0_EvalExpression2(ps *env.ProgramState, limited bool) *env.ProgramState {
	Rye0_EvalExpressionConcrete(ps)
	if ps.ReturnFlag {
		return ps
	}
	return ps
}

func Rye0_EvalExpressionInj(ps *env.ProgramState, inj env.Object, injnow bool) (*env.ProgramState, bool) {
	var esleft *env.ProgramState
	if inj == nil || !injnow {
		// if there is no injected value just eval the concrete expression
		esleft = Rye0_EvalExpressionConcrete(ps)
		if ps.ReturnFlag {
			return ps, injnow
		}
	} else {
		// otherwise set program state to specific one and injected value to result
		// set injnow to false and if return flag return
		esleft = ps
		esleft.Res = inj
		injnow = false
	}
	return ps, injnow
}

// the main part of evaluator, if it were a polish only we would need almost only this
// switches over all rye values and acts on them
func Rye0_EvalExpressionConcrete(ps *env.ProgramState) *env.ProgramState {
	//defer trace2("EvalExpression_>>>")
	object := ps.Ser.Pop()
	//trace2("Before entering expression")
	if object != nil {
		switch object.Type() {
		case env.IntegerType, env.DecimalType, env.StringType, env.VoidType, env.UriType, env.EmailType: // env.TagwordType, JM 20230126
			if !ps.SkipFlag {
				ps.Res = object
			}
		case env.BlockType:
			if !ps.SkipFlag {
				block := object.(env.Block)
				// block mode 1 is for eval blocks
				if block.Mode == 1 {
					ser := ps.Ser
					ps.Ser = block.Series
					res := make([]env.Object, 0)
					for ps.Ser.Pos() < ps.Ser.Len() {
						Rye0_EvalExpression2(ps, false)
						if checkErrorReturnFlag(ps) {
							return ps
						}
						res = append(res, ps.Res)
					}
					ps.Ser = ser
					ps.Res = *env.NewBlock(*env.NewTSeries(res))
				} else if block.Mode == 2 {
					ser := ps.Ser
					ps.Ser = block.Series
					EvalBlock(ps)
					ps.Ser = ser
					// return ps.Res
				} else {
					ps.Res = object
				}
			}
		case env.TagwordType:
			ps.Res = *env.NewWord(object.(env.Tagword).Index)
			return ps
		case env.WordType:
			rr := Rye0_EvalWord(ps, object.(env.Word), nil, false, false)
			return rr
		case env.CPathType:
			rr := Rye0_EvalWord(ps, object, nil, false, false)
			return rr
		case env.BuiltinType:
			return Rye0_CallBuiltin(object.(env.Builtin), ps, nil, false, false, nil)
		case env.GenwordType:
			return Rye0_EvalGenword(ps, object.(env.Genword), nil, false)
		case env.SetwordType:
			return Rye0_EvalSetword(ps, object.(env.Setword))
		case env.ModwordType:
			return Rye0_EvalModword(ps, object.(env.Modword))
		case env.GetwordType:
			return Rye0_EvalGetword(ps, object.(env.Getword), nil, false)
		case env.CommaType:
			ps.ErrorFlag = true
			ps.Res = env.NewError("expression guard inside expression")
		case env.ErrorType:
			ps.ErrorFlag = true
			ps.Res = env.NewError("Error ??")
		default:
			fmt.Println(object.Inspect(*ps.Idx))
			ps.ErrorFlag = true
			ps.Res = env.NewError("unknown rye value")
		}
	} else {
		ps.ErrorFlag = true
		ps.Res = env.NewError("expected rye value but it's missing")
	}

	return ps
}

// this basicalls returns a rye value behind a word or cpath (context path)
// for words it just looks in to current context and with it to parent contexts
func Rye0_findWordValue(ps *env.ProgramState, word1 env.Object) (bool, env.Object, *env.RyeCtx) {
	switch word := word1.(type) {
	case env.Word:
		object, found := ps.Ctx.Get(word.Index)
		return found, object, nil
	case env.Opword:
		object, found := ps.Ctx.Get(word.Index)
		return found, object, nil
	case env.Pipeword:
		object, found := ps.Ctx.Get(word.Index)
		return found, object, nil
	case env.CPath:
		currCtx := ps.Ctx
		i := 1
	gogo1:
		currWord := word.GetWordNumber(i)
		object, found := currCtx.Get(currWord.Index)
		if found && word.Cnt > i {
			switch swObj := object.(type) {
			case env.RyeCtx:
				currCtx = &swObj
				i += 1
				goto gogo1
			case env.Dict:
				return found, *env.NewString("asdsad"), currCtx
			}
		}
		return found, object, currCtx
	default:
		return false, nil, nil
	}
}

// Evaluates a word
// first tries to find a value in normal context. If there were no generic words this would be mostly it
// if word is not found then it tries to get the value of next expression
// and find a generic word based
// on that, it here is leftval already present it can dispatc on it otherwise
func Rye0_EvalWord(ps *env.ProgramState, word env.Object, leftVal env.Object, toLeft bool, pipeSecond bool) *env.ProgramState {
	// LOCAL FIRST
	var firstVal env.Object
	found, object, session := Rye0_findWordValue(ps, word)
	pos := ps.Ser.GetPos()
	if !found { // look at Generic words, but first check type
		// fmt.Println(pipeSecond)
		kind := 0
		if leftVal != nil {
			kind = leftVal.GetKind()
		}
		if leftVal == nil && !pipeSecond {
			if !ps.Ser.AtLast() {
				Rye0_EvalExpressionConcrete(ps)
				if ps.ReturnFlag {
					return ps
				}
				leftVal = ps.Res
				kind = leftVal.GetKind()
			}
		}
		if pipeSecond {
			if !ps.Ser.AtLast() {
				Rye0_EvalExpressionConcrete(ps)
				if ps.ReturnFlag {
					return ps
				}
				firstVal = ps.Res
				kind = firstVal.GetKind()
				// fmt.Println("pipeSecond kind")
			}
		}
		// fmt.Println(kind)
		rword, ok := word.(env.Word)
		if ok && leftVal != nil && ps.Ctx.Kind.Index != -1 { // don't use generic words if context kind is -1 --- TODO temporary solution to isolates, think about it more
			object, found = ps.Gen.Get(kind, rword.Index)
		}
	}
	if found {
		return Rye0_EvalObject(ps, object, leftVal, toLeft, session, pipeSecond, firstVal) //ww0128a *
	} else {
		ps.ErrorFlag = true
		if !ps.FailureFlag {
			ps.Ser.SetPos(pos)
			ps.Res = env.NewError2(5, "word not found: "+word.Print(*ps.Idx))
		}
		return ps
	}
}

// if word is defined to be generic ... I am not sure we will keep this ... we will decide with more use
// then if explicitly treats it as generic word
func Rye0_EvalGenword(ps *env.ProgramState, word env.Genword, leftVal env.Object, toLeft bool) *env.ProgramState {
	Rye0_EvalExpressionConcrete(ps)

	var arg0 = ps.Res
	object, found := ps.Gen.Get(arg0.GetKind(), word.Index)
	if found {
		return Rye0_EvalObject(ps, object, arg0, toLeft, nil, false, nil) //ww0128a *
	} else {
		ps.ErrorFlag = true
		ps.Res = env.NewError("generic word not found: " + word.Print(*ps.Idx))
		return ps
	}
}

// evaluates a get-word . it retrieves rye value behid it w/o evaluation
func Rye0_EvalGetword(ps *env.ProgramState, word env.Getword, leftVal env.Object, toLeft bool) *env.ProgramState {
	object, found := ps.Ctx.Get(word.Index)
	if found {
		ps.Res = object
		return ps
	} else {
		ps.ErrorFlag = true
		ps.Res = env.NewError("word not found: " + word.Print(*ps.Idx))
		return ps
	}
}

// evaluates a rye value, most of them just get returned, except builtins, functions and context paths
func Rye0_EvalObject(ps *env.ProgramState, object env.Object, leftVal env.Object, toLeft bool, ctx *env.RyeCtx, pipeSecond bool, firstVal env.Object) *env.ProgramState {
	switch object.Type() {
	case env.FunctionType:
		fn := object.(env.Function)
		return Rye0_CallFunction(fn, ps, leftVal, toLeft, ctx)
	case env.CPathType: // RMME
		fn := object.(env.Function)
		return Rye0_CallFunction(fn, ps, leftVal, toLeft, ctx)
	case env.BuiltinType:
		bu := object.(env.Builtin)

		if checkFlagsBi(bu, ps, 333) {
			return ps
		}
		return Rye0_CallBuiltin(bu, ps, leftVal, toLeft, pipeSecond, firstVal)
	default:
		if !ps.SkipFlag {
			ps.Res = object
		}
		return ps
	}
}

// evaluates expression to the right and sets the result of it to a word in current context
func Rye0_EvalSetword(ps *env.ProgramState, word env.Setword) *env.ProgramState {
	// es1 := EvalExpression(es)
	ps1, _ := Rye0_EvalExpressionInj(ps, nil, false)
	idx := word.Index
	if ps.AllowMod {
		ps1.Ctx.Mod(idx, ps.Res)
	} else {
		ok := ps1.Ctx.SetNew(idx, ps1.Res, ps.Idx)
		if !ok {
			ps.Res = env.NewError("Can't set already set word " + ps.Idx.GetWord(idx) + ", try using modword (2)")
			ps.FailureFlag = true
			ps.ErrorFlag = true
		}
	}
	return ps
}

// evaluates expression to the right and sets the result of it to a word in current context
func Rye0_EvalModword(ps *env.ProgramState, word env.Modword) *env.ProgramState {
	// es1 := EvalExpression(es)
	ps1, _ := Rye0_EvalExpressionInj(ps, nil, false)
	idx := word.Index
	ps1.Ctx.Mod(idx, ps1.Res)
	return ps1
}

func Rye0_CallFunction(fn env.Function, ps *env.ProgramState, arg0 env.Object, toLeft bool, ctx *env.RyeCtx) *env.ProgramState {
	env0 := ps.Ctx // store reference to current env in local
	var fnCtx *env.RyeCtx
	if ctx != nil { // called via contextpath and this is the context
		//		fmt.Println("if 111")
		if fn.Pure {
			//			fmt.Println("calling pure function")
			//		fmt.Println(es.PCtx)
			fnCtx = env.NewEnv(ps.PCtx)
		} else {
			if fn.Ctx != nil { // if context was defined at definition time, pass it as parent.
				if fn.InCtx {
					fnCtx = fn.Ctx
				} else {
					fn.Ctx.Parent = ctx
					fnCtx = env.NewEnv(fn.Ctx)
				}
			} else {
				fnCtx = env.NewEnv(ctx)
			}
		}
	} else {
		//fmt.Println("else1")
		if fn.Pure {
			//		fmt.Println("calling pure function")
			//	fmt.Println(es.PCtx)
			fnCtx = env.NewEnv(ps.PCtx)
		} else {
			if fn.Ctx != nil { // if context was defined at definition time, pass it as parent.
				// Q: Would we want to pass it directly at any point?
				//    Maybe to remove need of creating new contexts, for reuse, of to be able to modify it?
				fnCtx = env.NewEnv(fn.Ctx)
			} else {
				fnCtx = env.NewEnv(env0)
			}
		}
	}

	ii := 0
	// evalExprFn := EvalExpression // 2020-01-12 .. changed to ion2
	evalExprFn := Rye0_EvalExpression2
	if arg0 != nil {
		if fn.Spec.Series.Len() > 0 {
			index := fn.Spec.Series.Get(ii).(env.Word).Index
			fnCtx.Set(index, arg0)
			ps.Args[ii] = index
			ii = 1
			if !toLeft {
				//evalExprFn = EvalExpression_ // 2020-01-12 .. changed to ion2
				evalExprFn = Rye0_EvalExpression2
			}
		}
	}
	// collect arguments
	for i := ii; i < fn.Argsn; i += 1 {
		ps = evalExprFn(ps, true)
		if checkErrorReturnFlag(ps) {
			return ps
		}
		index := fn.Spec.Series.Get(i).(env.Word).Index
		fnCtx.Set(index, ps.Res)
		if i == 0 {
			arg0 = ps.Res
		}
		ps.Args[i] = index
	}
	ser0 := ps.Ser // only after we process the arguments and get new position
	ps.Ser = fn.Body.Series

	// *******
	env0 = ps.Ctx // store reference to current env in local
	ps.Ctx = fnCtx

	var result *env.ProgramState
	//	if ctx != nil {
	//		result = EvalBlockInCtx(es, ctx)
	//	} else {
	if arg0 != nil {
		result = Rye0_EvalBlockInj(ps, arg0, true)
	} else {
		result = EvalBlock(ps)
	}
	//	}
	// MaybeDisplayFailureOrError(result, result.Idx, "call function")
	if result.ForcedResult != nil {
		ps.Res = result.ForcedResult
		result.ForcedResult = nil
	} else {
		ps.Res = result.Res
	}
	ps.Ctx = env0
	ps.Ser = ser0
	ps.ReturnFlag = false
	trace2("Before user function returns")
	return ps
	/*         for (var i=0;i<h.length;i+=1) {
	    var e = this.evalExpr(block,pos,state,depth+1);
	    pos = e[1];
	    state = e[2];
	    var idx = this.indexWord(h[i][1]);
	    lctx[idx] = e[0];
	}
	// evaluate code block of function
	r = this.evalBlock(b,0,lctx,depth+1);
	return [r.length>4?r:r[0],pos,state,depth];
	*/
}

func Rye0_CallFunctionArgs2(fn env.Function, ps *env.ProgramState, arg0 env.Object, arg1 env.Object, ctx *env.RyeCtx) *env.ProgramState {
	var fnCtx *env.RyeCtx
	env0 := ps.Ctx  // store reference to current env in local
	if ctx != nil { // called via contextpath and this is the context
		//		fmt.Println("if 111")
		if fn.Pure {
			//			fmt.Println("calling pure function")
			//		fmt.Println(es.PCtx)
			fnCtx = env.NewEnv(ps.PCtx)
		} else {
			if fn.Ctx != nil { // if context was defined at definition time, pass it as parent.
				fn.Ctx.Parent = ctx
				fnCtx = env.NewEnv(fn.Ctx)
			} else {
				fnCtx = env.NewEnv(ctx)
			}
		}
	} else {
		//fmt.Println("else1")
		if fn.Pure {
			//		fmt.Println("calling pure function")
			//	fmt.Println(es.PCtx)
			fnCtx = env.NewEnv(ps.PCtx)
		} else {
			if fn.Ctx != nil { // if context was defined at definition time, pass it as parent.
				// Q: Would we want to pass it directly at any point?
				//    Maybe to remove need of creating new contexts, for reuse, of to be able to modify it?
				fnCtx = env.NewEnv(fn.Ctx)
			} else {
				fnCtx = env.NewEnv(env0)
			}
		}
	}
	if checkErrorReturnFlag(ps) {
		return ps
	}
	i := 0
	index := fn.Spec.Series.Get(i).(env.Word).Index
	fnCtx.Set(index, arg0)
	i = 1
	index = fn.Spec.Series.Get(i).(env.Word).Index
	fnCtx.Set(index, arg1)
	// TRY
	psX := env.NewProgramState(fn.Body.Series, ps.Idx)
	psX.Ctx = fnCtx
	psX.PCtx = ps.PCtx
	psX.Gen = ps.Gen

	// END TRY

	/// ser0 := ps.Ser
	/// ps.Ser = fn.Body.Series
	/// env0 = ps.Ctx
	/// ps.Ctx = fnCtx

	var result *env.ProgramState
	psX.Ser.SetPos(0)
	result = Rye0_EvalBlockInj(psX, arg0, true)
	// fmt.Println(result)
	// fmt.Println(result.Res)
	MaybeDisplayFailureOrError(result, result.Idx, "call func args 2")
	if result.ForcedResult != nil {
		ps.Res = result.ForcedResult
		result.ForcedResult = nil
	} else {
		ps.Res = result.Res
	}
	/// ps.Ctx = env0
	/// ps.Ser = ser0
	ps.ReturnFlag = false
	return ps
}

func Rye0_CallFunctionArgs4(fn env.Function, ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, ctx *env.RyeCtx) *env.ProgramState {
	var fnCtx *env.RyeCtx
	env0 := ps.Ctx  // store reference to current env in local
	if ctx != nil { // called via contextpath and this is the context
		if fn.Pure {
			fnCtx = env.NewEnv(ps.PCtx)
		} else {
			if fn.Ctx != nil { // if context was defined at definition time, pass it as parent.
				fn.Ctx.Parent = ctx
				fnCtx = env.NewEnv(fn.Ctx)
			} else {
				fnCtx = env.NewEnv(ctx)
			}
		}
	} else {
		if fn.Pure {
			fnCtx = env.NewEnv(ps.PCtx)
		} else {
			if fn.Ctx != nil { // if context was defined at definition time, pass it as parent.
				// Q: Would we want to pass it directly at any point?
				//    Maybe to remove need of creating new contexts, for reuse, of to be able to modify it?
				fnCtx = env.NewEnv(fn.Ctx)
			} else {
				fnCtx = env.NewEnv(env0)
			}
		}
	}
	if checkErrorReturnFlag(ps) {
		return ps
	}
	i := 0
	index := fn.Spec.Series.Get(i).(env.Word).Index
	fnCtx.Set(index, arg0)
	i = 1
	index = fn.Spec.Series.Get(i).(env.Word).Index
	fnCtx.Set(index, arg1)
	i = 2
	index = fn.Spec.Series.Get(i).(env.Word).Index
	fnCtx.Set(index, arg2)
	i = 3
	index = fn.Spec.Series.Get(i).(env.Word).Index
	fnCtx.Set(index, arg3)
	// TRY
	psX := env.NewProgramState(fn.Body.Series, ps.Idx)
	psX.Ctx = fnCtx
	psX.PCtx = ps.PCtx
	psX.Gen = ps.Gen

	// END TRY
	var result *env.ProgramState
	psX.Ser.SetPos(0)
	result = Rye0_EvalBlockInj(psX, arg0, true)
	MaybeDisplayFailureOrError(result, result.Idx, "call func args 4")
	if result.ForcedResult != nil {
		ps.Res = result.ForcedResult
		result.ForcedResult = nil
	} else {
		ps.Res = result.Res
	}
	ps.ReturnFlag = false
	return ps
}

func Rye0_DetermineContext(fn env.Function, ps *env.ProgramState, ctx *env.RyeCtx) *env.RyeCtx {
	var fnCtx *env.RyeCtx
	env0 := ps.Ctx  // store reference to current env in local
	if ctx != nil { // called via contextpath and this is the context
		if fn.Pure {
			fnCtx = env.NewEnv(ps.PCtx)
		} else {
			if fn.Ctx != nil { // if context was defined at definition time, pass it as parent.
				fn.Ctx.Parent = ctx
				fnCtx = env.NewEnv(fn.Ctx)
			} else {
				fnCtx = env.NewEnv(ctx)
			}
		}
	} else {
		if fn.Pure {
			fnCtx = env.NewEnv(ps.PCtx)
		} else {
			if fn.Ctx != nil { // if context was defined at definition time, pass it as parent.
				// Q: Would we want to pass it directly at any point?
				//    Maybe to remove need of creating new contexts, for reuse, of to be able to modify it?
				fnCtx = env.NewEnv(fn.Ctx)
			} else {
				fnCtx = env.NewEnv(env0)
			}
		}
	}
	return fnCtx
}

func Rye0_CallFunctionArgsN(fn env.Function, ps *env.ProgramState, ctx *env.RyeCtx, args ...env.Object) *env.ProgramState {
	var fnCtx = DetermineContext(fn, ps, ctx)
	if checkErrorReturnFlag(ps) {
		return ps
	}
	for i, arg := range args {
		index := fn.Spec.Series.Get(i).(env.Word).Index
		fnCtx.Set(index, arg)
	}
	// TRY
	psX := env.NewProgramState(fn.Body.Series, ps.Idx)
	psX.Ctx = fnCtx
	psX.PCtx = ps.PCtx
	psX.Gen = ps.Gen

	// END TRY
	var result *env.ProgramState
	psX.Ser.SetPos(0)
	if len(args) > 0 {
		result = Rye0_EvalBlockInj(psX, args[0], true)
	} else {
		result = EvalBlock(psX)
	}
	MaybeDisplayFailureOrError(result, result.Idx, "call func args N")
	if result.ForcedResult != nil {
		ps.Res = result.ForcedResult
		result.ForcedResult = nil
	} else {
		ps.Res = result.Res
	}
	ps.ReturnFlag = false
	return ps
}

// 20240217 - reduced the currying stuff, 30% speedup time-it { rye0 { loop 1000000 { _+ 1 1 } } }

func Rye0_CallBuiltin(bi env.Builtin, ps *env.ProgramState, arg0_ env.Object, toLeft bool, pipeSecond bool, firstVal env.Object) *env.ProgramState {
	arg0 := bi.Cur0 //env.Object(bi.Cur0)
	arg1 := bi.Cur1
	arg2 := bi.Cur2
	arg3 := bi.Cur3
	arg4 := bi.Cur4

	// fmt.Println("+**")

	evalExprFn := Rye0_EvalExpression2

	if bi.Argsn > 0 && bi.Cur0 == nil {
		evalExprFn(ps, true)

		if checkFlagsBi(bi, ps, 0) {
			return ps
		}
		if checkErrorReturnFlag(ps) {
			ps.Res = env.NewError4(0, "argument 1 of "+strconv.Itoa(bi.Argsn)+" missing of builtin: '"+bi.Doc+"'", ps.Res.(*env.Error), nil)
			return ps
		}
		arg0 = ps.Res
	}

	if bi.Argsn > 1 && bi.Cur1 == nil {
		evalExprFn(ps, true) // <---- THESE DETERMINE IF IT CONSUMES WHOLE EXPRESSION OR NOT IN CASE OF PIPEWORDS .. HM*... MAYBE WOULD COULD HAVE A WORD MODIFIER?? a: 2 |add 5 a:: 2 |add 5 print* --TODO

		if checkFlagsBi(bi, ps, 1) {
			return ps
		}
		if checkErrorReturnFlag(ps) {
			ps.Res = env.NewError4(0, "argument 2 of "+strconv.Itoa(bi.Argsn)+" missing of builtin: '"+bi.Doc+"'", ps.Res.(*env.Error), nil)
			return ps
		}
		arg1 = ps.Res
	}
	if bi.Argsn > 2 {
		evalExprFn(ps, true)

		if checkFlagsBi(bi, ps, 2) {
			return ps
		}
		if checkErrorReturnFlag(ps) {
			ps.Res = env.NewError4(0, "argument 3 missing", ps.Res.(*env.Error), nil)
			return ps
		}
		arg2 = ps.Res
	}
	if bi.Argsn > 3 {
		evalExprFn(ps, true)
		arg3 = ps.Res

	}
	if bi.Argsn > 4 {
		evalExprFn(ps, true)
		arg4 = ps.Res
	}
	/*
		variadic version
		for i := 0; i < bi.Argsn; i += 1 {
			EvalExpression(ps)
			args[i] = ps.Res
		}
		ps.Res = bi.Fn(ps, args...)
	*/
	ps.Res = bi.Fn(ps, arg0, arg1, arg2, arg3, arg4)
	return ps
}

func Rye0_DirectlyCallBuiltin(ps *env.ProgramState, bi env.Builtin, a0 env.Object, a1 env.Object) env.Object {
	// let's try to make it without array allocation and without variadic arguments that also maybe actualizes splice
	// up to 2 curried variables and 2 in caller
	// examples:
	// 	map { 1 2 3 } add _ 10
	var arg0 env.Object
	var arg1 env.Object

	if bi.Cur0 != nil {
		arg0 = bi.Cur0
		if bi.Cur1 != nil {
			arg1 = bi.Cur1
		} else {
			arg1 = a0
		}
	} else {
		arg0 = a0
		if bi.Cur1 != nil {
			arg1 = bi.Cur1
		} else {
			arg1 = a1
		}
	}
	arg2 := bi.Cur2
	arg3 := bi.Cur3
	arg4 := bi.Cur4
	return bi.Fn(ps, arg0, arg1, arg2, arg3, arg4)
}

func Rye0_MaybeDisplayFailureOrError(es *env.ProgramState, genv *env.Idxs, tag string) {
	if es.FailureFlag {
		fmt.Println("\x1b[33m" + "Failure" + "\x1b[0m")
		fmt.Println(tag)
	}
	if es.ErrorFlag {
		fmt.Println("\x1b[31m" + es.Res.Print(*genv))
		switch err := es.Res.(type) {
		case env.Error:
			fmt.Println(err.CodeBlock.PositionAndSurroundingElements(*genv))
			fmt.Println("Error not pointer so bug. #temp")
		case *env.Error:
			fmt.Println("At location:")
			fmt.Print(err.CodeBlock.PositionAndSurroundingElements(*genv))
		}
		fmt.Println("\x1b[0m")
		fmt.Println(tag)
		// ENTER CONSOLE ON ERROR
		// es.ErrorFlag = false
		// es.FailureFlag = false
		// DoRyeRepl(es, "do", true)
	}
	// cebelca2659- vklopi kontne skupine
}

func Rye0_MaybeDisplayFailureOrErrorWASM(es *env.ProgramState, genv *env.Idxs, printfn func(string), tag string) {
	if es.FailureFlag {
		printfn("\x1b[33m" + "Failure" + "\x1b[0m")
		printfn(tag)
	}
	if es.ErrorFlag {
		printfn("\x1b[31;3m" + es.Res.Print(*genv))
		switch err := es.Res.(type) {
		case env.Error:
			printfn(err.CodeBlock.PositionAndSurroundingElements(*genv))
			printfn("Error not pointer so bug. #temp")
		case *env.Error:
			printfn("At location:")
			printfn(err.CodeBlock.PositionAndSurroundingElements(*genv))
		}
		printfn("\x1b[0m")
		printfn(tag)
	}
}

// if there is failure flag and given builtin doesn't accept failure
// then error flag is raised and true returned
// otherwise false
// USED -- before evaluating a builtin
// TODO -- once we know it works in all situations remove all debug lines
//
//	and rewrite
func Rye0_checkFlagsBi(bi env.Builtin, ps *env.ProgramState, n int) bool {
	trace("CHECK FLAGS BI")
	//trace(n)
	//trace(ps.Res)
	//	trace(bi)
	if ps.FailureFlag {
		trace("------ > FailureFlag")
		if bi.AcceptFailure {
			trace2("----- > Accept Failure")
		} else {
			// fmt.Println("checkFlagsBi***")
			trace2("Fail ------->  Error.")
			switch err := ps.Res.(type) {
			case env.Error:
				if err.CodeBlock.Len() == 0 {
					err.CodeBlock = ps.Ser
					err.CodeContext = ps.Ctx
				}
			case *env.Error:
				if err.CodeBlock.Len() == 0 {
					err.CodeBlock = ps.Ser
					err.CodeContext = ps.Ctx
				}
			}
			ps.ErrorFlag = true
			return true
		}
	} else {
		trace2("NOT FailuteFlag")
	}
	return false
}

func Rye0_checkContextErrorHandler(ps *env.ProgramState) bool {
	// check if there is error-handler word defined in context (or parent).
	erh, w_exists := ps.Idx.GetIndex("error-handler")
	if !w_exists {
		return false
	}
	handler, exists := ps.Ctx.Get(erh)
	// if it is get the block
	if !exists {
		return false
	}
	// ps.FailureFlag = false
	ps.InErrHandler = true
	switch bloc := handler.(type) {
	case env.Block:
		ser := ps.Ser
		ps.Ser = bloc.Series
		EvalBlockInj(ps, ps.Res, true)
		ps.Ser = ser
	}
	ps.InErrHandler = false
	return true
	// evaluate the block, where injected value is error

	// NOT SURE YET, if we proceed with failure based on return, always or what
	// need more practical-situations to figure this out
}

// if failure flag is raised and return flag is not up
// then raise the error flag and return true
// USED -- on returns from block
func Rye0_checkFlagsAfterBlock(ps *env.ProgramState, n int) bool {
	trace2("CHECK FLAGS AFTER BLOCKS")
	trace2(n)
	/// fmt.Println("checkFlagsAfterBlock***")

	//trace(ps.Res)
	if ps.FailureFlag && !ps.ReturnFlag {
		trace2("FailureFlag")
		trace2("Fail->Error.")

		if !ps.InErrHandler {
			if checkContextErrorHandler(ps) {
				return false // error should be picked up in the handler block if not handeled -- TODO -- hopefully
			}
		}

		switch err := ps.Res.(type) {
		case env.Error:
			if err.CodeBlock.Len() == 0 {
				err.CodeBlock = ps.Ser
				err.CodeContext = ps.Ctx
			}
		case *env.Error:
			if err.CodeBlock.Len() == 0 {
				err.CodeBlock = ps.Ser
				err.CodeContext = ps.Ctx
			}
		}
		trace2("FAIL -> ERROR blk")
		ps.ErrorFlag = true
		return true
	} else {
		trace2("NOT FailureFlag")
	}
	return false
}
