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

func EvalBlock(ps *env.ProgramState) *env.ProgramState {
	return EvalBlockInj(ps, nil, false)
}

func EvalBlockInCtx(ps *env.ProgramState, ctx *env.RyeCtx) *env.ProgramState {
	ctx2 := ps.Ctx
	ps.Ctx = ctx
	res := EvalBlockInj(ps, nil, false)
	ps.Ctx = ctx2
	return res
}

func EvalBlockInj(ps *env.ProgramState, inj env.Object, injnow bool) *env.ProgramState {
	//fmt.Println("BEFORE BLOCK ***")
	for ps.Ser.Pos() < ps.Ser.Len() {
		//fmt.Println("EVALBLOCK: " + strconv.FormatInt(int64(es.Ser.Pos()), 10))
		//fmt.Println("EVALBLOCK N: " + strconv.FormatInt(int64(es.Ser.Len()), 10))
		// TODO --- look at JS code for eval .. what state we carry around
		// TODO --- probably block, position, env ... pack all this into one struct
		//		--- that could be passed in and returned from eval functions (I think)
		ps, injnow = EvalExpressionInj(ps, inj, injnow)
		if checkFlagsAfterBlock(ps, 101) {
			return ps
		}
		if ps.ReturnFlag || ps.ErrorFlag {
			return ps
		}
		ps, injnow = MaybeAcceptComma(ps, inj, injnow)
		//es.Res.Trace("After eval expression")
	}
	//es.Inj = nil
	return ps
}

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

func EvalExpression2(ps *env.ProgramState, limited bool) *env.ProgramState {
	esleft := EvalExpression_(ps)
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

func EvalExpression(ps *env.ProgramState) *env.ProgramState {
	es1, _ := EvalExpressionInj(ps, nil, false)
	return es1
}

func EvalExpressionInj(ps *env.ProgramState, inj env.Object, injnow bool) (*env.ProgramState, bool) {
	var esleft *env.ProgramState
	if inj == nil || injnow == false {
		esleft = EvalExpression_(ps)
		trace2(esleft)
		trace("EvalExpressionInj in first IF")
		if ps.ReturnFlag {
			return ps, injnow
		}
		/*if checkFlags2(es, 102) {
			return es, injnow
		}*/

	} else {
		trace("EvalExpressionInj in ELSE")
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
	trace2("Calling Maybe from EvalExp Inj")
	return MaybeEvalOpwordOnRight(esleft.Ser.Peek(), esleft, false), injnow
	//return esleft
}

/* OPWORDWWW */
func MaybeEvalOpwordOnRight(nextObj env.Object, ps *env.ProgramState, limited bool) *env.ProgramState {
	trace2("MaybeEvalWord -----------======--------> 1")
	if ps.ReturnFlag {
		return ps
	}
	switch opword := nextObj.(type) {
	case env.Opword:
		//ProcOpword(nextObj, es)
		ps.Ser.Next()
		//fmt.Println("MaybeEvalOpword..1")
		ps = EvalWord(ps, opword.ToWord(), ps.Res, false)
		//fmt.Println("MaybeEvalOpword..2")
		return MaybeEvalOpwordOnRight(ps.Ser.Peek(), ps, limited)
	case env.Pipeword:
		//ProcOpword(nextObj, es)
		if limited {
			trace("LIMITED JUMP")
			return ps
		}
		ps.Ser.Next()
		trace2(ps.Res)
		ps = EvalWord(ps, opword.ToWord(), ps.Res, false)
		trace2("MaybeEvalPipeword ---------> 2")
		if ps.ReturnFlag {
			trace2("RETURN ES")
			return ps //... not sure if we need this
		}
		if ps.FailureFlag { // uncommented 202008017
			trace2("FAILURE FLAG DETECTED !")
			ps.FailureFlag = false
			ps.ErrorFlag = true
			ps.ReturnFlag = true
			return ps
		}
		trace2("MaybeEval --------------------------------------> looping around")
		return MaybeEvalOpwordOnRight(ps.Ser.Peek(), ps, limited)
	case env.LSetword:
		if limited {
			return ps
		}
		//ProcOpword(nextObj, es)
		idx := opword.Index
		ps.Ctx.Set(idx, ps.Res)
		ps.Ser.Next()
		return MaybeEvalOpwordOnRight(ps.Ser.Peek(), ps, limited)
	}
	return ps
}

/*func ProcOpword(obj env.Object, left env.Object, es *env.ProgramState) *env.ProgramState {
	//collect next args if there are more than 1
	// call function as normal with it
	// maybe we could just call previous function, split it to first and rest args or not for performances?
}*/

/* */

func EvalExpression_(ps *env.ProgramState) *env.ProgramState {
	trace2("<<<EvalExpression_")
	defer trace2("EvalExpression_>>>")
	object := ps.Ser.Pop()
	//es.Idx.Probe()
	trace2("Before entering expression")
	if object != nil {
		switch object.Type() {
		case env.IntegerType:
			ps.Res = object
		case env.StringType:
			ps.Res = object
		case env.BlockType:
			ps.Res = object
		case env.VoidType:
			ps.Res = object
		case env.TagwordType:
			ps.Res = object
		case env.UriType:
			ps.Res = object
		case env.EmailType:
			ps.Res = object
		case env.WordType:
			rr := EvalWord(ps, object.(env.Word), nil, false)
			return rr
		case env.CPathType:
			rr := EvalWord(ps, object, nil, false)
			return rr
		case env.BuiltinType:
			return CallBuiltin(object.(env.Builtin), ps, nil, false)
		case env.GenwordType:
			return EvalGenword(ps, object.(env.Genword), nil, false)
		case env.SetwordType:
			return EvalSetword(ps, object.(env.Setword))
		case env.GetwordType:
			return EvalGetword(ps, object.(env.Getword), nil, false)
		case env.CommaType:
			ps.ErrorFlag = true
			ps.Res = env.NewError("ERROR: expression guard inside expression!")
		default:
			ps.ErrorFlag = true
			ps.Res = env.NewError("Not known type")
		}
	} else {
		ps.ErrorFlag = true
		ps.Res = env.NewError("Not known type")
	}

	return ps
}

/* func EvalExpression(es *env.ProgramState) *env.ProgramState {
	switch object := es.Ser.Pop().(type) {
	case env.Integer:
		es.Res = object
	case env.Block:
		es.Res = object
	case env.Word:
		return EvalWord(es, object)
	case env.Setword:
		return EvalSetword(es, object)
	default:
		es.Res = env.NewError("Not know type")
	}
	return es
}*/

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
			}
		}
		return found, object, currCtx
		/*
			if found {
				//var ctx env.RyeCtx
				ctx := ctx1.(env.RyeCtx)
				object, found = ctx.Get()
				if object.Type() == env.CtxType {
					ctx3 := object.(env.RyeCtx)
					currCtx = &ctx3
					i += 1
					goto gogo1
				}
				return found, object, &ctx
			}*/
		return false, object, nil // TODO -- should trigger error
	default:
		return false, nil, nil
	}
}

func EvalWord(ps *env.ProgramState, word env.Object, leftVal env.Object, toLeft bool) *env.ProgramState {
	//fmt.Println("*EVAL WORD*")
	//es.Ctx.Probe(*es.Idx)
	/* WE MUST PROCESS FUNCTIONS THAT DON'T ACCEPT ANY ARGS IN THIS CASE .. TODO SOON
	if leftVal == nil {
		EvalExpression_(es)
		leftVal = es.Res
	}
	//es.Res.Trace("EvalGenword")
	object, found := es.Gen.Get(leftVal.GetKind(), word.Index)
	if !found {
		//object.Trace("OBJECT RETURNED: ")
		object, found = es.Ctx.Get(word.Index)
	}*/
	// LOCAL FIRST
	found, object, ctx := findWordValue(ps, word)
	if !found { // look at Generic words, but first check type
		if leftVal == nil {
			trace("****31")
			trace(ps.Ser.Pos())
			trace(ps.Ser.Len())
			if !ps.Ser.AtLast() {
				trace("****32")
				EvalExpression_(ps)
				trace("****32")
				if ps.ReturnFlag {
					return ps
				}
				leftVal = ps.Res
			}
		}
		trace("****21")
		//es.Res.Trace("EvalGenword")
		if leftVal != nil {
			object, found = ps.Gen.Get(leftVal.GetKind(), word.(env.Word).Index)
		}
		//object.Trace("OBJECT RETURNED: ")
	}
	trace("****33")
	//object, found := es.Ctx.Get(word.Index)
	if found {
		trace("****33")
		return EvalObject(ps, object, leftVal, toLeft, ctx) //ww0128a *
		//es.Res.Trace("After eval Object")
		//return es
	} else {
		trace("****34")
		ps.ErrorFlag = true
		if ps.FailureFlag == false {
			ps.Res = *env.NewError2(5, "Word not found: "+word.Inspect(*ps.Idx))
		}
		return ps
	}
}

func EvalGenword(ps *env.ProgramState, word env.Genword, leftVal env.Object, toLeft bool) *env.ProgramState {
	//fmt.Println("*EVAL GENWORD*")
	//es.Ctx.Probe(*es.Idx)
	//es.Ser.Next()
	EvalExpression_(ps)
	trace("****4")

	var arg0 = ps.Res
	//es.Res.Trace("EvalGenword")
	object, found := ps.Gen.Get(arg0.GetKind(), word.Index)
	//object.Trace("OBJECT RETURNED: ")
	if found {
		return EvalObject(ps, object, arg0, toLeft, nil) //ww0128a *
		//es.Res.Trace("After eval Object")
		//return es
	} else {
		ps.Res = env.NewError("Generic word not found: " + word.Inspect(*ps.Idx))
		return ps
	}
}

func EvalGetword(ps *env.ProgramState, word env.Getword, leftVal env.Object, toLeft bool) *env.ProgramState {
	object, found := ps.Ctx.Get(word.Index)
	if found {
		ps.Res = object
		return ps
	} else {
		ps.Res = env.NewError("Generic word not found: " + word.Inspect(*ps.Idx))
		return ps
	}
}

/* func EvalObject2(es *env.ProgramState, object env.Object) *env.ProgramState {
	switch object.(type) {
	case env.Function:
		return CallFunction(object.(env.Function), es)
	case env.Builtin:
		bu := object.(env.Builtin)
		return CallBuiltin(bu, es)
	default:
		es.Res = object
		return es
	}
	return es
} */

func EvalObject(ps *env.ProgramState, object env.Object, leftVal env.Object, toLeft bool, ctx *env.RyeCtx) *env.ProgramState {
	//fmt.Print("EVAL OBJECT")
	switch object.Type() {
	case env.FunctionType:
		//d fmt.Println(" FUN**")
		fn := object.(env.Function)
		return CallFunction(fn, ps, leftVal, toLeft, ctx)
		//return es
	case env.CPathType: // RMME
		//fmt.Println(" CPATH **************")
		fn := object.(env.Function)
		return CallFunction(fn, ps, leftVal, toLeft, ctx)
		//return es
	case env.BuiltinType:
		//fmt.Println(" BUIL**")
		//fmt.Println(es.Ser.GetPos())
		//fmt.Println(" BUILTIN**")
		bu := object.(env.Builtin)

		// OBJECT INJECTION EXPERIMENT
		// es.Ser.Put(bu)

		if checkFlagsBi(bu, ps, 333) {
			return ps
		}
		return CallBuiltin(bu, ps, leftVal, toLeft)

		//es.Res.Trace("After builtin call")
		//return es
	default:
		//d object.Trace("DEFAULT**")
		ps.Res = object
		//es.Res.Trace("After object returned")
		return ps
	}
	return ps
}

func EvalSetword(es *env.ProgramState, word env.Setword) *env.ProgramState {
	es1 := EvalExpression(es)
	idx := word.Index
	es1.Ctx.Set(idx, es1.Res)
	return es1
}

/*func CallRFunction2(fn env.Function, es *env.ProgramState, arg0 env.Object, arg1 env.Object) *env.ProgramState {

}*/

func CallFunction(fn env.Function, ps *env.ProgramState, arg0 env.Object, toLeft bool, ctx *env.RyeCtx) *env.ProgramState {
	//fmt.Println("Call Function")
	//fmt.Println(es.Ser.GetPos())
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
		if ps.ErrorFlag || ps.ReturnFlag {
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
	if ps.ErrorFlag || ps.ReturnFlag {
		return ps
	}
	i := 0
	index := fn.Spec.Series.Get(i).(env.Word).Index
	fnCtx.Set(index, arg0)
	i = 1
	index = fn.Spec.Series.Get(i).(env.Word).Index
	fnCtx.Set(index, arg1)
	ser0 := ps.Ser
	ps.Ser = fn.Body.Series
	env0 = ps.Ctx
	ps.Ctx = fnCtx
	var result *env.ProgramState
	result = EvalBlockInj(ps, arg0, true)
	if result.ForcedResult != nil {
		ps.Res = result.ForcedResult
		result.ForcedResult = nil
	} else {
		ps.Res = result.Res
	}
	ps.Ctx = env0
	ps.Ser = ser0
	ps.ReturnFlag = false
	return ps
}

func fmt1() { fmt.Print(1) }

func trace(x interface{}) {
	// fmt.Print("\x1b[36m")
	// fmt.Print(x)
	// fmt.Println("\x1b[0m")
}
func trace2(x interface{}) {
	// fmt.Print("\x1b[56m")
	// fmt.Print(x)
	// fmt.Println("\x1b[0m")
}

// if there is failure flag and given builtin doesn't accept failure
// then error flag is raised and true returned
// otherwise false
// USED -- before evaluationg a builtin
// TODO -- once we know it works in all situations remove all debug lines
// 		and rewrite
func checkFlagsBi(bi env.Builtin, ps *env.ProgramState, n int) bool {
	//trace("CHECK FLAGS")
	//trace(n)
	//trace(ps.Res)
	//	trace(bi)
	if ps.FailureFlag {
		//trace("------ > FailureFlag")
		if bi.AcceptFailure {
			//trace("----- > Accept Failure")
		} else {
			//trace("Fail ------->  Error.")
			ps.ErrorFlag = true
			return true
		}
	} else {
		//trace("NOT FailuteFlag")
	}
	return false
}

// if failure flag is raised and return flag is not up
// then raise the error flag and return true
// USED -- on returns from block
func checkFlagsAfterBlock(ps *env.ProgramState, n int) bool {
	//trace("CHECK FLAGS 2")
	//trace(n)
	//trace(ps.Res)
	if ps.FailureFlag && !ps.ReturnFlag {
		//trace("FailureFlag")
		//trace("Fail->Error.")
		ps.ErrorFlag = true
		return true
	} else {
		//trace("NOT FailureFlag")
	}
	return false
}

func CallBuiltin(bi env.Builtin, ps *env.ProgramState, arg0_ env.Object, toLeft bool) *env.ProgramState {
	////args := make([]env.Object, bi.Argsn)
	/*pospos := ps.Ser.GetPos()
	for i := 0; i < bi.Argsn; i += 1 {
		EvalExpression(ps)
		args[i] = ps.Res
	}
	ps.Ser.SetPos(pospos)*/

	// let's try to make it without array allocation and without variadic arguments that also maybe actualizes splice
	arg0 := env.Object(bi.Cur0) //env.Object(bi.Cur0)
	arg1 := env.Object(bi.Cur1)
	arg2 := env.Object(bi.Cur2)
	arg3 := env.Object(bi.Cur3)
	arg4 := env.Object(bi.Cur4)

	// This is just experiment if we could at currying provide ?fn or ?builtin and
	// with arity of 0 and it would get executed at calltime. So closure would become
	// closure: fnc _ ?current-context _
	// this is maybe only usefull to provide sort of dynamic constant to a curried
	// probably not worthe the special case but here for exploration for now just
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

	if arg0_ != nil {
		//fmt.Println("ARG0 = LEFT")
		arg0 = arg0_
		if !toLeft {
			//fmt.Println("L TO R *** ")
			//evalExprFn = EvalExpression_
		} else {
			//fmt.Println("TO THE *** LEFT")
		}
	} else if bi.Argsn > 0 && bi.Cur0 == nil {
		//fmt.Println(" ARG 1 ")
		//fmt.Println(ps.Ser.GetPos())
		evalExprFn(ps, true)

		if checkFlagsBi(bi, ps, 0) {
			return ps
		}
		if ps.ErrorFlag || ps.ReturnFlag {
			return ps
		}
		if ps.Res.Type() == env.VoidType {
			curry = true
		} else {
			arg0 = ps.Res
		}
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
		if ps.ErrorFlag || ps.ReturnFlag {
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
	if curry {
		bi.Cur0 = arg0
		bi.Cur1 = arg1
		bi.Cur2 = arg2
		bi.Cur3 = arg3
		bi.Cur4 = arg4
		ps.Res = bi
	} else {
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
		arg0 = env.Object(bi.Cur0)
		if bi.Cur1 != nil {
			arg1 = env.Object(bi.Cur1)
		} else {
			arg1 = a0
		}
	} else {
		arg0 = a0
		if bi.Cur1 != nil {
			arg1 = env.Object(bi.Cur1)
		} else {
			arg1 = a1
		}
	}
	arg2 := env.Object(bi.Cur2)
	arg3 := env.Object(bi.Cur3)
	arg4 := env.Object(bi.Cur4)
	return bi.Fn(ps, arg0, arg1, arg2, arg3, arg4)
}
