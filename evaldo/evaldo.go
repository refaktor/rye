package evaldo

import (
	"Rejy_go_v1/env"
	"fmt"
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
	//fmt.Println("BEFORE BLOCK ***")
	for es.Ser.Pos() < es.Ser.Len() {
		//fmt.Println("EVALBLOCK: " + strconv.FormatInt(int64(es.Ser.Pos()), 10))
		//fmt.Println("EVALBLOCK N: " + strconv.FormatInt(int64(es.Ser.Len()), 10))
		// TODO --- look at JS code for eval .. what state we carry around
		// TODO --- probably block, position, env ... pack all this into one struct
		//		--- that could be passed in and returned from eval functions (I think)
		es = EvalExpression(es)
		MaybeAcceptComma(es)
		//es.Res.Trace("After eval expression")
	}
	return es
}

func MaybeAcceptComma(es *env.ProgramState) *env.ProgramState {
	obj := es.Ser.Peek()
	switch obj.(type) {
	case env.Comma:
		es.Ser.Next()
	}
	return es
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
	////// OPWORDWWW
	// IF WE COMMENT IN NEXT LINE IT WORKS WITHOUT OPWORDS PROCESSING
	fmt.Println("EvalExpression")
	fmt.Println(es.Ser.GetPos())
	return MaybeEvalOpwordOnRight(esleft.Ser.Peek(), esleft, limited)
	return esleft
}

func EvalExpression(es *env.ProgramState) *env.ProgramState {
	esleft := EvalExpression_(es)
	////// OPWORDWWW
	// IF WE COMMENT IN NEXT LINE IT WORKS WITHOUT OPWORDS PROCESSING
	fmt.Println("EvalExpression")
	fmt.Println(es.Ser.GetPos())
	return MaybeEvalOpwordOnRight(esleft.Ser.Peek(), esleft, false)
	return esleft
}

/* OPWORDWWW */
func MaybeEvalOpwordOnRight(nextObj env.Object, es *env.ProgramState, limited bool) *env.ProgramState {
	switch opword := nextObj.(type) {
	case env.Opword:
		//ProcOpword(nextObj, es)
		es.Ser.Next()
		fmt.Println("MaybeEvalOpword..1")
		es = EvalWord(es, opword.ToWord(), es.Res, false)
		fmt.Println("MaybeEvalOpword..2")
		return MaybeEvalOpwordOnRight(es.Ser.Peek(), es, limited)
	case env.Pipeword:
		//ProcOpword(nextObj, es)
		if limited {
			return es
		}
		es.Ser.Next()
		fmt.Println("MaybeEvalPipeword..1")
		es = EvalWord(es, opword.ToWord(), es.Res, false)
		fmt.Println("MaybeEvalPipeword..2")
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
	fmt.Println("EvalExpression___")
	object := es.Ser.Pop()
	//es.Idx.Probe()
	//object.Trace("Before entering expression")
	switch object.Type() {
	case env.IntegerType:
		es.Res = object
	case env.BlockType:
		es.Res = object
	case env.WordType:
		return EvalWord(es, object.(env.Word), nil, false)
		//es1.Res.Trace("After eval word")
		//return es1
	case env.SetwordType:
		return EvalSetword(es, object.(env.Setword))
	case env.CommaType:
		es.Res = env.NewError("ERROR: expression guard inside expression!")
	default:
		es.Res = env.NewError("Not know type")
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

func EvalWord(es *env.ProgramState, word env.Word, leftVal env.Object, toLeft bool) *env.ProgramState {
	fmt.Println("*EVAL WORD*")
	//es.Env.Probe(*es.Idx)
	object, found := es.Env.Get(word.Index)
	if found {
		return EvalObject(es, object, leftVal, toLeft) //ww0128a *
		//es.Res.Trace("After eval Object")
		//return es
	} else {
		es.Res = env.NewError("Word not found: " + word.Inspect(*es.Idx))
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

func EvalObject(es *env.ProgramState, object env.Object, leftVal env.Object, toLeft bool) *env.ProgramState {
	fmt.Print("EVAL OBJECT")
	switch object.Type() {
	case env.FunctionType:
		fmt.Println(" FUN**")
		fn := object.(env.Function)
		return CallFunction(fn, es, leftVal, toLeft)
		//es.Res.Trace("After user function call")
		//return es
	case env.BuiltinType:
		fmt.Println(" BUIL**")
		fmt.Println(es.Ser.GetPos())
		//fmt.Println(" BUILTIN**")
		bu := object.(env.Builtin)
		return CallBuiltin(bu, es, leftVal, toLeft)
		//es.Res.Trace("After builtin call")
		//return es
	default:
		es.Res = object
		//es.Res.Trace("After object returned")
		return es
	}
	return es
}

func EvalSetword(es *env.ProgramState, word env.Setword) *env.ProgramState {
	es1 := EvalExpression(es)
	idx := word.Index
	es1.Env.Set(idx, es1.Res)
	return es1
}

func CallFunction(fn env.Function, es *env.ProgramState, arg0 env.Object, toLeft bool) *env.ProgramState {
	fmt.Println("Call Function")
	fmt.Println(es.Ser.GetPos())
	env0 := es.Env // store reference to current env in local
	es.Env = env.NewEnv(env0)

	ii := 0
	evalExprFn := EvalExpression
	if arg0 != nil {
		index := fn.Spec.Series.Get(ii).(env.Word).Index
		es.Env.Set(index, arg0)
		es.Args[ii] = index
		ii = 1
		if !toLeft {
			evalExprFn = EvalExpression_
		}
	}
	// collect arguments
	for i := ii; i < fn.Argsn; i += 1 {
		es = evalExprFn(es)
		index := fn.Spec.Series.Get(i).(env.Word).Index
		es.Env.Set(index, es.Res)
		es.Args[i] = index
	}
	ser0 := es.Ser // only after we process the arguments and get new position
	es.Ser = fn.Body.Series
	//es.Idx.Probe()
	//es.Env.Probe(*es.Idx)

	result := EvalBlock(es)

	es.Res = result.Res
	es.Env = env0
	es.Ser = ser0
	//es.Res.Trace("Before user function returns")
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

func CallBuiltin(bi env.Builtin, ps *env.ProgramState, arg0_ env.Object, toLeft bool) *env.ProgramState {
	////args := make([]env.Object, bi.Argsn)
	/*pospos := ps.Ser.GetPos()
	for i := 0; i < bi.Argsn; i += 1 {
		EvalExpression(ps)
		args[i] = ps.Res
	}
	ps.Ser.SetPos(pospos)*/
	// let's try to make it without array allocation and without variadic arguments that also maybe actualizes splice
	arg0 := env.Object(nil)
	arg1 := env.Object(nil)
	arg2 := env.Object(nil)
	arg3 := env.Object(nil)
	arg4 := env.Object(nil)
	evalExprFn := EvalExpression2
	if arg0_ != nil {
		fmt.Println("ARG0 = LEFT")
		arg0 = arg0_
		if !toLeft {
			//fmt.Println("L TO R *** ")
			//evalExprFn = EvalExpression_
		} else {
			//fmt.Println("TO THE *** LEFT")
		}
	} else if bi.Argsn > 0 {
		fmt.Println(" ARG 1 ")
		fmt.Println(ps.Ser.GetPos())
		evalExprFn(ps, true)
		arg0 = ps.Res
	}
	if bi.Argsn > 1 {
		evalExprFn(ps, true)
		arg1 = ps.Res
	}
	if bi.Argsn > 2 {
		evalExprFn(ps, true)
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
	//ps.Res.Trace("Before builtin returns")
	return ps
}
