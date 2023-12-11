package evaldo

import (
	"fmt"
	"rye/env"
	//"fmt"
	//"strconv"
)

// TODO NEXT -- figure out how to call a builtin function .. look at monkey
// TODO NEXT -- figure out how to map a builtin function ... directly by word index or by the value after the index?? can the value be Go-s compiled function?
/*
type ProgramState struct {
	Ser env.TSeries
	Res env.Object
	Env env.Env
	Idx env.Idxs
}

func NewProgramState(ser env.TSeries, idx env.Idxs) *ProgramState {
	ps := ProgramState{
		ser,
		nil,
		*env.NewEnv(),
		idx,
	}
	return &ps
}
*/
// Rejy0 DO dialect evaluator works like this:
// literal values (numbers) evaluate to itself
// blocks return itself, don't evaluate it's contents
// words are referenced in Env, and evaluate to it's value (which can be a literal, block, word, function or builtin)
//  word returns it's value and evaluates it
// functions evaluate by executing, taking objects according to spec setting local Env and evaluating body block
// builtins evaluate by executing, collecting arguments and calling a builtin function with them
// setwords take expression on the right and sets an environment reference to that word
//
// The basic goal of Rejy0 evaluator is to run fibonacci and (factorial 10000x) and make basic function calls fast enough
//  Goal is to reach speed similar to Rebol2 and Red
//
// Rejy1 and 2 will have a little more complex evaluators with strings, infix, postfix, ...
// We should keep separate evaluators possible at any time to test for regressions while adding those features
//

// DESCR: the most general EvalBlock
func EvalBlock(ps *env.ProgramState) *env.ProgramState {
	return EvalBlockInj(ps, nil, false)
}

// DESCR: eval a block in specific context
func EvalBlockInCtx(ps *env.ProgramState, ctx *env.RyeCtx) *env.ProgramState {
	ctx2 := ps.Ctx
	ps.Ctx = ctx
	res := EvalBlockInj(ps, nil, false)
	ps.Ctx = ctx2
	return res
}

// DESCR: eval a block in specific context
func EvalBlockInCtxInj(ps *env.ProgramState, ctx *env.RyeCtx, inj env.Object, injnow bool) *env.ProgramState {
	ctx2 := ps.Ctx
	ps.Ctx = ctx
	res := EvalBlockInj(ps, inj, injnow)
	ps.Ctx = ctx2
	return res
}

// DESCR: the main evaluator of block
func EvalBlockInj(ps *env.ProgramState, inj env.Object, injnow bool) *env.ProgramState {
	//fmt.Println("BEFORE BLOCK ***")
	// repeats until at the end of the block
	for ps.Ser.Pos() < ps.Ser.Len() {
		//fmt.Println("EVALBLOCK: " + strconv.FormatInt(int64(es.Ser.Pos()), 10))
		//fmt.Println("EVALBLOCK N: " + strconv.FormatInt(int64(es.Ser.Len()), 10))
		// TODO --- look at JS code for eval .. what state we carry around
		// TODO --- probably block, position, env ... pack all this into one struct
		//		--- that could be passed in and returned from eval functions (I think)
		// evaluate expression
		ps, injnow = EvalExpressionInj(ps, inj, injnow)
		// check and raise the flags if needed if true (error) return
		// --- 20201213: removed because require didn't really work :	if checkFlagsAfterBlock(ps, 101) {
		// we could add current error, block and position to the trace
		//		return ps
		//	}
		if checkFlagsAfterBlock(ps, 101) {
			return ps
		}
		// if return flag was raised return ( errorflag I think would return in previous if anyway)
		// --- 20201213 --
		if checkErrorReturnFlag(ps) {
			return ps
		}
		ps, injnow = MaybeAcceptComma(ps, inj, injnow)
		//es.Res.Trace("After eval expression")
	}
	// added here from above 20201213
	//if checkErrorReturnFlag(ps) {
	//	return ps
	//}
	//es.Inj = nil
	return ps
}

// comma (expression guard) can be present between block-level expressions, in case of injected block they
// reinject the value
func MaybeAcceptComma(ps *env.ProgramState, inj env.Object, injnow bool) (*env.ProgramState, bool) {
	obj := ps.Ser.Peek()
	switch obj.(type) {
	case env.Comma:
		ps.Ser.Next()
		if inj != nil {
			injnow = true
		}
	}
	return ps, injnow
}

//
// How opwords are processed:
// after an expression is evaluated we would usually just return the result, but now we have to check
// one spot to the right if there is an opword there.
// opword references a function (user or builtin) that takes N number of arguments. First argument
// is provided by the left side of expression (the current value of expression) all other, if there are any are provided
// by expressions to the right, which themselves can use opwords
//
// " 1 + 2 " + is a function taking two args. After 1 is returned it checks and sees an opword + on right. Since it has 2
// args it collects 1 expressions to the right (2) and calls a function.
//
// " 1 + 2 + 3 " this would work on it's own after 1 it would see + opword seek another val from the right, recurse lower
// see 2 check and see another opword accept 3 and call + 2 3 first get 5 and then call + 1 2. But this way
// we would have right to left execution, which looses whole point of opwords where we want to have left to right "stream"
//
// we need to after seeing + collect first expression without including further opwords. Return the expressions as args to
// first opword and calling it + 1 2 then looking if there is another opword on the right, recursing and doing the same.
//
// later we should add processing of parenthesis / groups to this
//
// " 1 + ( 2 + 3 ) "
// just quick speculation ... () will also have to work with general evaluator, not just op-words like (add 1 2) it would be best
// if it didn't slow things down, but would just be some limit (on a stack?) of how much further current expression can go.
// ( would add it to stack ) would stop processing another expr and throw error if not all were provided and remove from stack.
//
// do we need to recurse in all these cases or can we flatten it to some while loop? which could maybe be faster?
// - while loop + stack should in general be faster .. we should try it

// this functions is used to evaluate expression in the middle of block
// currently it's called to collect arguments for builtins and functions
func EvalExpression2(ps *env.ProgramState, limited bool) *env.ProgramState {
	esleft := EvalExpressionConcrete(ps)
	if ps.ReturnFlag {
		return ps
	}
	////// OPWORDWWW
	// IF WE COMMENT IN NEXT LINE IT WORKS WITHOUT OPWORDS PROCESSING
	//fmt.Println("EvalExpression")
	//fmt.Println(es.Ser.GetPos())
	return MaybeEvalOpwordOnRight(esleft.Ser.Peek(), esleft, limited)
	//return esleft
}

// this only seems to be used for evalserword ... refactored ... DELETE later
/* func EvalExpression(ps *env.ProgramState) *env.ProgramState {
	es1, _ := EvalExpressionInj(ps, nil, false)
	return es1
} */

// I don't fully get this function in this review ... it's this way so it handles op and pipe words
// mainly, but I need to get deeper again to write a proper explanation
// TODO -- return to this and explain
func EvalExpressionInj(ps *env.ProgramState, inj env.Object, injnow bool) (*env.ProgramState, bool) {
	var esleft *env.ProgramState
	if inj == nil || !injnow {
		// if there is no injected value just eval the concrete expression
		esleft = EvalExpressionConcrete(ps)
		if ps.ReturnFlag {
			return ps, injnow
		}
	} else {
		// otherwise set program state to specific one and injected value to result
		// set injnow to false and if return flag return
		esleft = ps
		esleft.Res = inj
		injnow = false
		if ps.ReturnFlag { //20200817
			return ps, injnow
		}
		//esleft.Inj = nil
	}
	////// OPWORDWWW
	// IF WE COMMENT IN NEXT LINE IT WORKS WITHOUT OPWORDS PROCESSING
	//fmt.Println("EvalExpression")
	//fmt.Println(es.Ser.GetPos())
	// trace2("Calling Maybe from EvalExp Inj")
	return MaybeEvalOpwordOnRight(esleft.Ser.Peek(), esleft, false), injnow
	//return esleft
}

// REFATOR THIS WITH CODE ABOVE
// when seeing bigger picture, just adding fow eval-with
func EvalExpressionInjLimited(ps *env.ProgramState, inj env.Object, injnow bool) (*env.ProgramState, bool) { // TODO -- doesn't work .. would be nice - eval-with
	var esleft *env.ProgramState
	if inj == nil || !injnow {
		// if there is no injected value just eval the concrete expression
		esleft = EvalExpressionConcrete(ps)
		if ps.ReturnFlag {
			return ps, injnow
		}
	} else {
		// otherwise set program state to specific one and injected value to result
		// set injnow to false and if return flag return
		esleft = ps
		esleft.Res = inj
		injnow = false
		if ps.ReturnFlag { //20200817
			return ps, injnow
		}
		//esleft.Inj = nil
	}
	////// OPWORDWWW
	// IF WE COMMENT IN NEXT LINE IT WORKS WITHOUT OPWORDS PROCESSING
	//fmt.Println("EvalExpression")
	//fmt.Println(es.Ser.GetPos())
	// trace2("Calling Maybe from EvalExp Inj")
	return MaybeEvalOpwordOnRight(esleft.Ser.Peek(), esleft, false), injnow
	//return esleft
}

// this function get's the next object (unevaluated), progra state, limited bool (op or pipe)
// first if there is return flag it returns (not sure if this is necesarry here) TODO -- figure out
// if next object is opword it steps to next and evaluates the word then recurse  to maybe again
// if next object is pipeword
//
//	on limited return (what is limited exactly ? TODO)
//	step to next word and evaluate it
//	again check for return flag
//	check for failure flag and cwitch to error ... doesn't one of checkFlags do this or similar? .TODO
//	recurse again
//
// if next is lsetword
//
//	set the value to word and recurse
func MaybeEvalOpwordOnRight(nextObj env.Object, ps *env.ProgramState, limited bool) *env.ProgramState {
	//trace2("MaybeEvalWord -----------======--------> 1")
	if ps.ReturnFlag || ps.ErrorFlag {
		return ps
	}
	switch opword := nextObj.(type) {
	case env.Opword:
		ps.Ser.Next()
		ps = EvalWord(ps, opword.ToWord(), ps.Res, false, opword.Force > 0)
		return MaybeEvalOpwordOnRight(ps.Ser.Peek(), ps, limited)
	case env.Pipeword:
		if limited {
			return ps
		}
		ps.Ser.Next()
		ps = EvalWord(ps, opword.ToWord(), ps.Res, false, opword.Force > 0)
		if ps.ReturnFlag {
			return ps //... not sure if we need this
		}
		// checkFlagsBi()
		/*if ps.FailureFlag { // uncommented 202008017
			ps.FailureFlag = false
			ps.ErrorFlag = true
			ps.ReturnFlag = true
			return ps
		}*/
		return MaybeEvalOpwordOnRight(ps.Ser.Peek(), ps, limited)
	case env.LSetword:
		if limited {
			return ps
		}
		//ProcOpword(nextObj, es)
		idx := opword.Index
		ps.Ctx.Set(idx, ps.Res)
		ps.Ser.Next()
		ps.SkipFlag = false
		return MaybeEvalOpwordOnRight(ps.Ser.Peek(), ps, limited)
	default:
		ps.SkipFlag = false
	}
	return ps
}

// the main part of evaluator, if it were a polish only we would need almost only this
// switches over all rye values and acts on them
func EvalExpressionConcrete(ps *env.ProgramState) *env.ProgramState {
	//defer trace2("EvalExpression_>>>")
	object := ps.Ser.Pop()
	//trace2("Before entering expression")
	if object != nil {
		switch object.Type() {
		case env.IntegerType, env.DecimalType, env.StringType, env.BlockType, env.VoidType, env.UriType, env.EmailType: // env.TagwordType, JM 20230126
			if !ps.SkipFlag {
				ps.Res = object
			}
		case env.TagwordType:
			ps.Res = *env.NewWord(object.(env.Tagword).Index)
			return ps
		case env.WordType:
			rr := EvalWord(ps, object.(env.Word), nil, false, false)
			return rr
		case env.CPathType:
			rr := EvalWord(ps, object, nil, false, false)
			return rr
		case env.BuiltinType:
			return CallBuiltin(object.(env.Builtin), ps, nil, false, false, nil)
		case env.GenwordType:
			return EvalGenword(ps, object.(env.Genword), nil, false)
		case env.SetwordType:
			return EvalSetword(ps, object.(env.Setword))
		case env.GetwordType:
			return EvalGetword(ps, object.(env.Getword), nil, false)
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
		ps.Res = env.NewError("expected rye value but got nothing")
	}

	return ps
}

// this basicalls returns a rye value behind a word or cpath (context path)
// for words it just looks in to current context and with it to parent contexts
func findWordValue(ps *env.ProgramState, word1 env.Object) (bool, env.Object, *env.RyeCtx) {
	switch word := word1.(type) {
	case env.Word:
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
func EvalWord(ps *env.ProgramState, word env.Object, leftVal env.Object, toLeft bool, pipeSecond bool) *env.ProgramState {
	// LOCAL FIRST
	var firstVal env.Object
	found, object, session := findWordValue(ps, word)
	pos := ps.Ser.GetPos()
	if !found { // look at Generic words, but first check type
		// fmt.Println(pipeSecond)
		kind := 0
		if leftVal != nil {
			kind = leftVal.GetKind()
		}
		if leftVal == nil && !pipeSecond {
			if !ps.Ser.AtLast() {
				EvalExpressionConcrete(ps)
				if ps.ReturnFlag {
					return ps
				}
				leftVal = ps.Res
				kind = leftVal.GetKind()
			}
		}
		if pipeSecond {
			if !ps.Ser.AtLast() {
				EvalExpressionConcrete(ps)
				if ps.ReturnFlag {
					return ps
				}
				firstVal = ps.Res
				kind = firstVal.GetKind()
				// fmt.Println("pipeSecond kind")
			}
		}
		// fmt.Println(kind)
		if leftVal != nil && ps.Ctx.Kind.Index != -1 { // don't use generic words if context kind is -1 --- TODO temporary solution to isolates, think about it more
			object, found = ps.Gen.Get(kind, word.(env.Word).Index)
		}
	}
	if found {
		return EvalObject(ps, object, leftVal, toLeft, session, pipeSecond, firstVal) //ww0128a *
	} else {
		ps.ErrorFlag = true
		if !ps.FailureFlag {
			ps.Ser.SetPos(pos)
			ps.Res = env.NewError2(5, "word not found: "+word.Probe(*ps.Idx))
		}
		return ps
	}
}

// if word is defined to be generic ... I am not sure we will keep this ... we will decide with more use
// then if explicitly treats it as generic word
func EvalGenword(ps *env.ProgramState, word env.Genword, leftVal env.Object, toLeft bool) *env.ProgramState {
	EvalExpressionConcrete(ps)

	var arg0 = ps.Res
	object, found := ps.Gen.Get(arg0.GetKind(), word.Index)
	if found {
		return EvalObject(ps, object, arg0, toLeft, nil, false, nil) //ww0128a *
	} else {
		ps.ErrorFlag = true
		ps.Res = env.NewError("generic word not found: " + word.Probe(*ps.Idx))
		return ps
	}
}

// evaluates a get-word . it retrieves rye value behid it w/o evaluation
func EvalGetword(ps *env.ProgramState, word env.Getword, leftVal env.Object, toLeft bool) *env.ProgramState {
	object, found := ps.Ctx.Get(word.Index)
	if found {
		ps.Res = object
		return ps
	} else {
		ps.ErrorFlag = true
		ps.Res = env.NewError("word not found: " + word.Probe(*ps.Idx))
		return ps
	}
}

// evaluates a rye value, most of them just get returned, except builtins, functions and context paths
func EvalObject(ps *env.ProgramState, object env.Object, leftVal env.Object, toLeft bool, ctx *env.RyeCtx, pipeSecond bool, firstVal env.Object) *env.ProgramState {
	switch object.Type() {
	case env.FunctionType:
		fn := object.(env.Function)
		return CallFunction(fn, ps, leftVal, toLeft, ctx)
	case env.CPathType: // RMME
		fn := object.(env.Function)
		return CallFunction(fn, ps, leftVal, toLeft, ctx)
	case env.BuiltinType:
		bu := object.(env.Builtin)

		if checkFlagsBi(bu, ps, 333) {
			return ps
		}
		return CallBuiltin(bu, ps, leftVal, toLeft, pipeSecond, firstVal)
	default:
		if !ps.SkipFlag {
			ps.Res = object
		}
		return ps
	}
}

// evaluates expression to the right and sets the result of it to a word in current context
func EvalSetword(ps *env.ProgramState, word env.Setword) *env.ProgramState {
	// es1 := EvalExpression(es)
	ps1, _ := EvalExpressionInj(ps, nil, false)
	idx := word.Index
	ps1.Ctx.Set(idx, ps1.Res)
	return ps1
}

func CallFunction(fn env.Function, ps *env.ProgramState, arg0 env.Object, toLeft bool, ctx *env.RyeCtx) *env.ProgramState {
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

	ii := 0
	// evalExprFn := EvalExpression // 2020-01-12 .. changed to ion2
	evalExprFn := EvalExpression2
	if arg0 != nil {
		index := fn.Spec.Series.Get(ii).(env.Word).Index
		fnCtx.Set(index, arg0)
		ps.Args[ii] = index
		ii = 1
		if !toLeft {
			//evalExprFn = EvalExpression_ // 2020-01-12 .. changed to ion2
			evalExprFn = EvalExpression2
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
	//es.Idx.Probe()
	//es.Ctx.Probe(*es.Idx)

	// *******
	env0 = ps.Ctx // store reference to current env in local
	ps.Ctx = fnCtx

	var result *env.ProgramState
	//	if ctx != nil {
	//		result = EvalBlockInCtx(es, ctx)
	//	} else {
	if arg0 != nil {
		result = EvalBlockInj(ps, arg0, true)
	} else {
		result = EvalBlock(ps)
	}
	//	}
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

func CallFunctionArgs2(fn env.Function, ps *env.ProgramState, arg0 env.Object, arg1 env.Object, ctx *env.RyeCtx) *env.ProgramState {
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
	ps.Ser.SetPos(0)
	result = EvalBlockInj(psX, arg0, true)
	// fmt.Println(result)
	// fmt.Println(result.Res)
	MaybeDisplayFailureOrError(result, result.Idx)
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

func CallFunctionArgs4(fn env.Function, ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, ctx *env.RyeCtx) *env.ProgramState {
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
	ps.Ser.SetPos(0)
	result = EvalBlockInj(psX, arg0, true)
	MaybeDisplayFailureOrError(result, result.Idx)
	if result.ForcedResult != nil {
		ps.Res = result.ForcedResult
		result.ForcedResult = nil
	} else {
		ps.Res = result.Res
	}
	ps.ReturnFlag = false
	return ps
}

func CallBuiltin(bi env.Builtin, ps *env.ProgramState, arg0_ env.Object, toLeft bool, pipeSecond bool, firstVal env.Object) *env.ProgramState {
	////args := make([]env.Object, bi.Argsn)
	/*pospos := ps.Ser.GetPos()
	for i := 0; i < bi.Argsn; i += 1 {
		EvalExpression(ps)
		args[i] = ps.Res
	}
	ps.Ser.SetPos(pospos)*/

	// let's try to make it without array allocation and without variadic arguments that also maybe actualizes splice
	arg0 := bi.Cur0 //env.Object(bi.Cur0)
	arg1 := bi.Cur1
	arg2 := bi.Cur2
	arg3 := bi.Cur3
	arg4 := bi.Cur4

	// This is just experiment if we could at currying provide ?fn or ?builtin and
	// with arity of 0 and it would get executed at call time. So closure would become
	// closure: fnc _ ?current-context _
	// this is maybe only useful to provide sort of dynamic constant to a curried
	// probably not worth the special case but here for exploration for now just
	// on arg1 . In case of arg being function this would not bind curry to static
	// value but to a result of a function, which would let us inject some context
	// bound dynamic value
	// ... we will see ...
	if bi.Cur1 != nil && bi.Cur1.Type() == env.BuiltinType {
		if bi.Cur1.(env.Builtin).Argsn == 0 {
			arg1 = DirectlyCallBuiltin(ps, bi.Cur1.(env.Builtin), nil, nil)
		}
	}
	// end of experiment

	evalExprFn := EvalExpression2
	curry := false

	trace("*** BUILTIN ***")
	trace(bi)

	if arg0_ != nil && !pipeSecond {
		//fmt.Println("ARG0 = LEFT")
		arg0 = arg0_
		//if !toLeft {
		//fmt.Println("L TO R *** ")
		//evalExprFn = EvalExpression_
		// }
	} else if firstVal != nil && pipeSecond {
		arg0 = firstVal
	} else if bi.Argsn > 0 && bi.Cur0 == nil {
		//fmt.Println(" ARG 1 ")
		//fmt.Println(ps.Ser.GetPos())
		evalExprFn(ps, true)

		if checkFlagsBi(bi, ps, 0) {
			return ps
		}
		if checkErrorReturnFlag(ps) {
			return ps
		}
		if ps.Res.Type() == env.VoidType {
			curry = true
		} else {
			arg0 = ps.Res
		}
	}

	if arg0_ != nil && pipeSecond {
		arg1 = arg0_
	} else if bi.Argsn > 1 && bi.Cur1 == nil {
		evalExprFn(ps, true) // <---- THESE DETERMINE IF IT CONSUMES WHOLE EXPRESSION OR NOT IN CASE OF PIPEWORDS .. HM*... MAYBE WOULD COULD HAVE A WORD MODIFIER?? a: 2 |add 5 a:: 2 |add 5 print* --TODO

		if checkFlagsBi(bi, ps, 1) {
			return ps
		}
		if checkErrorReturnFlag(ps) {
			return ps
		}
		//fmt.Println(ps.Res)
		if ps.Res.Type() == env.VoidType {
			curry = true
		} else {
			arg1 = ps.Res
		}
	}
	if bi.Argsn > 2 {
		evalExprFn(ps, true)

		if checkFlagsBi(bi, ps, 2) {
			return ps
		}
		if checkErrorReturnFlag(ps) {
			return ps
		}
		if ps.Res.Type() == env.VoidType {
			curry = true
		} else {
			arg2 = ps.Res
		}
	}
	if bi.Argsn > 3 {
		evalExprFn(ps, true)
		if ps.Res.Type() == env.VoidType {
			curry = true
		} else {
			arg3 = ps.Res
		}
	}
	if bi.Argsn > 4 {
		evalExprFn(ps, true)
		if ps.Res.Type() == env.VoidType {
			curry = true
		} else {
			arg4 = ps.Res
		}
	}
	/*
		variadic version
		for i := 0; i < bi.Argsn; i += 1 {
			EvalExpression(ps)
			args[i] = ps.Res
		}
		ps.Res = bi.Fn(ps, args...)
	*/
	trace("YOYOYOYOYOYOYOYOYOYO ---")
	if curry {
		bi.Cur0 = arg0
		bi.Cur1 = arg1
		bi.Cur2 = arg2
		bi.Cur3 = arg3
		bi.Cur4 = arg4
		ps.Res = bi
	} else {
		if ps.SkipFlag {
			trace2("SKIPPING ....")
			//if arg0_ != nil {
			trace2("PIPE ....")
			return ps
			//} else {
			//	trace2("RESETING ....")
			//ps.SkipFlag = false
			//}
		}
		ps.Res = bi.Fn(ps, arg0, arg1, arg2, arg3, arg4)
	}
	trace2(" ------------- Before builtin returns")
	return ps
}

func DirectlyCallBuiltin(ps *env.ProgramState, bi env.Builtin, a0 env.Object, a1 env.Object) env.Object {
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

// if there is failure flag and given builtin doesn't accept failure
// then error flag is raised and true returned
// otherwise false
// USED -- before evaluating a builtin
// TODO -- once we know it works in all situations remove all debug lines
//
//	and rewrite
func checkFlagsBi(bi env.Builtin, ps *env.ProgramState, n int) bool {
	trace("CHECK FLAGS BI")
	//trace(n)
	//trace(ps.Res)
	//	trace(bi)
	if ps.FailureFlag {
		trace("------ > FailureFlag")
		if bi.AcceptFailure {
			trace2("----- > Accept Failure")
		} else {
			fmt.Println("checkFlagsBi***")
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

func checkContextErrorHandler(ps *env.ProgramState) bool {
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
func checkFlagsAfterBlock(ps *env.ProgramState, n int) bool {
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

func checkErrorReturnFlag(ps *env.ProgramState) bool {
	// trace3("---- > return flags")
	if ps.ErrorFlag {
		/// fmt.Println("***checkErrorReturnFlags***")

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
		return true
	}
	return ps.ReturnFlag
}

func fmt1() { fmt.Print(1) }

func trace(x any) {
	//fmt.Print("\x1b[36m")
	//fmt.Print(x)
	//fmt.Println("\x1b[0m")
}
func trace2(x any) {
	//fmt.Print("\x1b[56m")
	//fmt.Print(x)
	//fmt.Println("\x1b[0m")
}

func trace3(x any) {
	fmt.Print("\x1b[56m")
	fmt.Print(x)
	fmt.Println("\x1b[0m")
}
