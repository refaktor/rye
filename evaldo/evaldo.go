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

func EvalBlock(es *env.ProgramState) *env.ProgramState {
	return EvalBlockInj(es, nil, false)
}

func EvalBlockInCtx(es *env.ProgramState, ctx *env.RyeCtx) *env.ProgramState {
	ctx2 := es.Ctx
	es.Ctx = ctx
	res := EvalBlockInj(es, nil, false)
	es.Ctx = ctx2
	return res
}

func EvalBlockInj(es *env.ProgramState, inj env.Object, injnow bool) *env.ProgramState {
	//fmt.Println("BEFORE BLOCK ***")
	for es.Ser.Pos() < es.Ser.Len() {
		//fmt.Println("EVALBLOCK: " + strconv.FormatInt(int64(es.Ser.Pos()), 10))
		//fmt.Println("EVALBLOCK N: " + strconv.FormatInt(int64(es.Ser.Len()), 10))
		// TODO --- look at JS code for eval .. what state we carry around
		// TODO --- probably block, position, env ... pack all this into one struct
		//		--- that could be passed in and returned from eval functions (I think)
		es, injnow = EvalExpressionInj(es, inj, injnow)
		if checkFlags2(es, 101) {
			return es
		}
		if es.ReturnFlag || es.ErrorFlag {
			return es
		}
		es, injnow = MaybeAcceptComma(es, inj, injnow)
		//es.Res.Trace("After eval expression")
	}
	//es.Inj = nil
	return es
}

func MaybeAcceptComma(es *env.ProgramState, inj env.Object, injnow bool) (*env.ProgramState, bool) {
	obj := es.Ser.Peek()
	switch obj.(type) {
	case env.Comma:
		es.Ser.Next()
		if inj != nil {
			//fmt.Println("INJNOW")
			injnow = true
		}
	}
	return es, injnow
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

func EvalExpression2(es *env.ProgramState, limited bool) *env.ProgramState {
	esleft := EvalExpression_(es)
	trace("EvalExpression2 beflre check of return flag")
	if es.ReturnFlag {
		return es
	}
	////// OPWORDWWW
	// IF WE COMMENT IN NEXT LINE IT WORKS WITHOUT OPWORDS PROCESSING
	//fmt.Println("EvalExpression")
	//fmt.Println(es.Ser.GetPos())
	return MaybeEvalOpwordOnRight(esleft.Ser.Peek(), esleft, limited)
	//return esleft
}

func EvalExpression(es *env.ProgramState) *env.ProgramState {
	es1, _ := EvalExpressionInj(es, nil, false)
	return es1
}

func EvalExpressionInj(es *env.ProgramState, inj env.Object, injnow bool) (*env.ProgramState, bool) {
	var esleft *env.ProgramState
	if inj == nil || injnow == false {
		esleft = EvalExpression_(es)
		trace2(esleft)
		trace("EvalExpressionInj in first IF")
		if es.ReturnFlag {
			return es, injnow
		}
		/*if checkFlags2(es, 102) {
			return es, injnow
		}*/

	} else {
		trace("EvalExpressionInj in ELSE")
		esleft = es
		esleft.Res = inj
		injnow = false
		if es.ReturnFlag { //20200817
			return es, injnow
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
func MaybeEvalOpwordOnRight(nextObj env.Object, es *env.ProgramState, limited bool) *env.ProgramState {
	trace2("MaybeEvalWord -----------======--------> 1")
	if es.ReturnFlag {
		return es
	}
	switch opword := nextObj.(type) {
	case env.Opword:
		//ProcOpword(nextObj, es)
		es.Ser.Next()
		//fmt.Println("MaybeEvalOpword..1")
		es = EvalWord(es, opword.ToWord(), es.Res, false)
		//fmt.Println("MaybeEvalOpword..2")
		return MaybeEvalOpwordOnRight(es.Ser.Peek(), es, limited)
	case env.Pipeword:
		//ProcOpword(nextObj, es)
		if limited {
			trace("LIMITED JUMP")
			return es
		}
		es.Ser.Next()
		trace2(es.Res)
		es = EvalWord(es, opword.ToWord(), es.Res, false)
		trace2("MaybeEvalPipeword ---------> 2")
		if es.ReturnFlag {
			trace2("RETURN ES")
			return es //... not sure if we need this
		}
		if es.FailureFlag { // uncommented 202008017
			trace2("FAILURE FLAG DETECTED !")
			es.FailureFlag = false
			es.ErrorFlag = true
			es.ReturnFlag = true
			return es
		}
		trace2("MaybeEval --------------------------------------> looping around")
		return MaybeEvalOpwordOnRight(es.Ser.Peek(), es, limited)
	case env.LSetword:
		if limited {
			return es
		}
		//ProcOpword(nextObj, es)
		idx := opword.Index
		es.Ctx.Set(idx, es.Res)
		es.Ser.Next()
		return MaybeEvalOpwordOnRight(es.Ser.Peek(), es, limited)
	}
	return es
}

/*func ProcOpword(obj env.Object, left env.Object, es *env.ProgramState) *env.ProgramState {
	//collect next args if there are more than 1
	// call function as normal with it
	// maybe we could just call previous function, split it to first and rest args or not for performances?
}*/

/* */

func EvalExpression_(es *env.ProgramState) *env.ProgramState {
	trace2("<<<EvalExpression_")
	defer trace2("EvalExpression_>>>")
	object := es.Ser.Pop()
	//es.Idx.Probe()
	trace2("Before entering expression")
	switch object.Type() {
	case env.IntegerType:
		es.Res = object
	case env.StringType:
		es.Res = object
	case env.BlockType:
		es.Res = object
	case env.VoidType:
		es.Res = object
	case env.TagwordType:
		es.Res = object
	case env.UriType:
		es.Res = object
	case env.WordType:
		rr := EvalWord(es, object.(env.Word), nil, false)
		return rr
	case env.CPathType:
		rr := EvalWord(es, object, nil, false)
		return rr
	case env.BuiltinType:
		return CallBuiltin(object.(env.Builtin), es, nil, false)
	case env.GenwordType:
		return EvalGenword(es, object.(env.Genword), nil, false)
	case env.SetwordType:
		return EvalSetword(es, object.(env.Setword))
	case env.GetwordType:
		return EvalGetword(es, object.(env.Getword), nil, false)
	case env.CommaType:
		es.ErrorFlag = true
		es.Res = env.NewError("ERROR: expression guard inside expression!")
	default:
		es.ErrorFlag = true
		es.Res = env.NewError("Not known type")
	}
	return es
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

func findWordValue(es *env.ProgramState, word1 env.Object) (bool, env.Object, *env.RyeCtx) {
	switch word := word1.(type) {
	case env.Word:
		object, found := es.Ctx.Get(word.Index)
		return found, object, nil

	case env.CPath:
		currCtx := es.Ctx
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

func EvalWord(es *env.ProgramState, word env.Object, leftVal env.Object, toLeft bool) *env.ProgramState {
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
	found, object, ctx := findWordValue(es, word)
	if !found { // look at Generic words, but first check type
		if leftVal == nil {
			trace("****31")
			trace(es.Ser.Pos())
			trace(es.Ser.Len())
			if !es.Ser.AtLast() {
				trace("****32")
				EvalExpression_(es)
				trace("****32")
				if es.ReturnFlag {
					return es
				}
				leftVal = es.Res
			}
		}
		trace("****21")
		//es.Res.Trace("EvalGenword")
		if leftVal != nil {
			object, found = es.Gen.Get(leftVal.GetKind(), word.(env.Word).Index)
		}
		//object.Trace("OBJECT RETURNED: ")
	}
	trace("****33")
	//object, found := es.Ctx.Get(word.Index)
	if found {
		trace("****33")
		return EvalObject(es, object, leftVal, toLeft, ctx) //ww0128a *
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

func EvalGenword(es *env.ProgramState, word env.Genword, leftVal env.Object, toLeft bool) *env.ProgramState {
	//fmt.Println("*EVAL GENWORD*")
	//es.Ctx.Probe(*es.Idx)
	//es.Ser.Next()
	EvalExpression_(es)
	trace("****4")

	var arg0 = es.Res
	//es.Res.Trace("EvalGenword")
	object, found := es.Gen.Get(arg0.GetKind(), word.Index)
	//object.Trace("OBJECT RETURNED: ")
	if found {
		return EvalObject(es, object, arg0, toLeft, nil) //ww0128a *
		//es.Res.Trace("After eval Object")
		//return es
	} else {
		es.Res = env.NewError("Generic word not found: " + word.Inspect(*es.Idx))
		return es
	}
}

func EvalGetword(es *env.ProgramState, word env.Getword, leftVal env.Object, toLeft bool) *env.ProgramState {
	object, found := es.Ctx.Get(word.Index)
	if found {
		es.Res = object
		return es
	} else {
		es.Res = env.NewError("Generic word not found: " + word.Inspect(*es.Idx))
		return es
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

func EvalObject(es *env.ProgramState, object env.Object, leftVal env.Object, toLeft bool, ctx *env.RyeCtx) *env.ProgramState {
	//fmt.Print("EVAL OBJECT")
	switch object.Type() {
	case env.FunctionType:
		//d fmt.Println(" FUN**")
		fn := object.(env.Function)
		return CallFunction(fn, es, leftVal, toLeft, ctx)
		//return es
	case env.CPathType: // RMME
		fmt.Println(" CPATH **************")
		fn := object.(env.Function)
		return CallFunction(fn, es, leftVal, toLeft, ctx)
		//return es
	case env.BuiltinType:
		//fmt.Println(" BUIL**")
		//fmt.Println(es.Ser.GetPos())
		//fmt.Println(" BUILTIN**")
		bu := object.(env.Builtin)

		// OBJECT INJECTION EXPERIMENT
		// es.Ser.Put(bu)

		if checkFlags(bu, es, 333) {
			return es
		}
		return CallBuiltin(bu, es, leftVal, toLeft)

		//es.Res.Trace("After builtin call")
		//return es
	default:
		//d object.Trace("DEFAULT**")
		es.Res = object
		//es.Res.Trace("After object returned")
		return es
	}
	return es
}

func EvalSetword(es *env.ProgramState, word env.Setword) *env.ProgramState {
	es1 := EvalExpression(es)
	idx := word.Index
	es1.Ctx.Set(idx, es1.Res)
	return es1
}

/*func CallRFunction2(fn env.Function, es *env.ProgramState, arg0 env.Object, arg1 env.Object) *env.ProgramState {

}*/

func CallFunction(fn env.Function, es *env.ProgramState, arg0 env.Object, toLeft bool, ctx *env.RyeCtx) *env.ProgramState {
	//fmt.Println("Call Function")
	//fmt.Println(es.Ser.GetPos())
	env0 := es.Ctx // store reference to current env in local
	var fnCtx *env.RyeCtx
	if ctx != nil { // called via contextpath and this is the context
		if fn.Ctx != nil { // if context was defined at definition time, pass it as parent.
			fn.Ctx.Parent = ctx
			fnCtx = env.NewEnv(fn.Ctx)
		} else {
			fnCtx = env.NewEnv(ctx)
		}
	} else {
		if fn.Ctx != nil { // if context was defined at definition time, pass it as parent.
			// Q: Would we want to pass it directly at any point?
			//    Maybe to remove need of creating new contexts, for reuse, of to be able to modify it?
			fnCtx = env.NewEnv(fn.Ctx)
		} else {
			fnCtx = env.NewEnv(env0)
		}
	}

	ii := 0
	// evalExprFn := EvalExpression // 2020-01-12 .. changed to ion2
	evalExprFn := EvalExpression2
	if arg0 != nil {
		index := fn.Spec.Series.Get(ii).(env.Word).Index
		fnCtx.Set(index, arg0)
		es.Args[ii] = index
		ii = 1
		if !toLeft {
			//evalExprFn = EvalExpression_ // 2020-01-12 .. changed to ion2
			evalExprFn = EvalExpression2
		}
	}
	// collect arguments
	for i := ii; i < fn.Argsn; i += 1 {
		es = evalExprFn(es, true)
		if es.ErrorFlag || es.ReturnFlag {
			return es
		}
		index := fn.Spec.Series.Get(i).(env.Word).Index
		fnCtx.Set(index, es.Res)
		es.Args[i] = index
	}
	ser0 := es.Ser // only after we process the arguments and get new position
	es.Ser = fn.Body.Series
	//es.Idx.Probe()
	//es.Ctx.Probe(*es.Idx)

	// *******
	env0 = es.Ctx // store reference to current env in local
	es.Ctx = fnCtx

	var result *env.ProgramState
	//	if ctx != nil {
	//		result = EvalBlockInCtx(es, ctx)
	//	} else {
	result = EvalBlock(es)
	//	}
	es.Res = result.Res
	es.Ctx = env0
	es.Ser = ser0
	es.ReturnFlag = false
	trace2("Before user function returns")
	return es
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

func checkFlags(bi env.Builtin, ps *env.ProgramState, n int) bool {
	trace("CHECK FLAGS")
	trace(n)
	trace(ps.Res)
	//	trace(bi)
	if ps.FailureFlag {
		trace("------ > FailureFlag")
		if bi.AcceptFailure {
			trace("----- > Accept Failure")
		} else {
			trace("Fail ------->  Error.")
			ps.ErrorFlag = true
			return true
		}
	} else {
		trace("NOT FailuteFlag")
	}
	return false
}
func checkFlags2(ps *env.ProgramState, n int) bool {
	trace("CHECK FLAGS 2")
	trace(n)
	trace(ps.Res)
	if ps.FailureFlag && !ps.ReturnFlag {
		trace("FailureFlag")
		trace("Fail->Error.")
		ps.ErrorFlag = true
		return true
	} else {
		trace("NOT FailuteFlag")
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

		if checkFlags(bi, ps, 0) {
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

		if checkFlags(bi, ps, 1) {
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

		if checkFlags(bi, ps, 2) {
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
