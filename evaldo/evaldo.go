package evaldo

import (
	"Rejy_go_v1/env"
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
		//es.Res.Trace("After eval expression")
	}
	return es
}

func EvalExpression(es *env.ProgramState) *env.ProgramState {
	object := es.Ser.Pop()
	//es.Idx.Probe()
	//object.Trace("Before entering expression")
	switch object.Type() {
	case env.IntegerType:
		es.Res = object
	case env.BlockType:
		es.Res = object
	case env.WordType:
		return EvalWord(es, object.(env.Word))
		//es1.Res.Trace("After eval word")
		//return es1
	case env.SetwordType:
		return EvalSetword(es, object.(env.Setword))
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

func EvalWord(es *env.ProgramState, word env.Word) *env.ProgramState {
	//fmt.Println("*EVAL WORD*")
	//es.Env.Probe(*es.Idx)
	object, found := es.Env.Get(word.Index)
	if found {
		return EvalObject(es, object) //ww0128a *
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

func EvalObject(es *env.ProgramState, object env.Object) *env.ProgramState {
	//fmt.Print(object.Inspect(*es.Idx))
	switch object.Type() {
	case env.FunctionType:
		//fmt.Println(" FUN**")
		fn := object.(env.Function)
		return CallFunction(fn, es)
		//es.Res.Trace("After user function call")
		//return es
	case env.BuiltinType:
		//fmt.Println(" BUILTIN**")
		bu := object.(env.Builtin)
		return CallBuiltin(bu, es)
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

func CallFunction(fn env.Function, es *env.ProgramState) *env.ProgramState {
	env0 := es.Env // store reference to current env in local
	es.Env = env.NewEnv(env0)

	// collect arguments
	for i := 0; i < fn.Argsn; i += 1 {
		es = EvalExpression(es)
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

func CallBuiltin(bi env.Builtin, ps *env.ProgramState) *env.ProgramState {
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
	if bi.Argsn > 0 {
		EvalExpression(ps)
		arg0 = ps.Res
	}
	if bi.Argsn > 1 {
		EvalExpression(ps)
		arg1 = ps.Res
	}
	if bi.Argsn > 2 {
		EvalExpression(ps)
		arg2 = ps.Res
	}
	if bi.Argsn > 3 {
		EvalExpression(ps)
		arg3 = ps.Res
	}
	if bi.Argsn > 4 {
		EvalExpression(ps)
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
