package evaldo

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/refaktor/rye/env"
)

// Flag to control whether to use the fast evaluator
var useFastEvaluator = false

// EnableFastEvaluator enables the fast evaluator for Rye0 dialect
func EnableFastEvaluator() {
	useFastEvaluator = true
}

// DisableFastEvaluator disables the fast evaluator for Rye0 dialect
func DisableFastEvaluator() {
	useFastEvaluator = false
}

// EVALUATE BLOCK
// HOTCODE
// DESCR: the most general EvalBlock
func EvalBlock(ps *env.ProgramState) {
	switch ps.Dialect {
	case env.Rye2Dialect:
		EvalBlockInj(ps, nil, false)
	case env.EyrDialect:
		Eyr_EvalBlockInside(ps, nil, false) // TODO ps.Stack is already in ps ... refactor
	case env.Rye0Dialect:
		// Check if we should use the fast evaluator
		if useFastEvaluator {
			Rye0_FastEvalBlock(ps)
		}
		Rye0_EvalBlockInj(ps, nil, false) // TODO ps.Stack is already in ps ... refactor
	case env.Rye00Dialect:
		Rye00_EvalBlockInj(ps, nil, false) // Simplified dialect for builtins and integers
	default:
		// TODO fail
	}
}

// This is the evaluator we use for general code, because it can be multidialect
func EvalBlockInjMultiDialect(ps *env.ProgramState, inj env.Object, injnow bool) { // TODO temp name -- refactor
	switch ps.Dialect {
	case env.Rye2Dialect:
		EvalBlockInj(ps, inj, injnow)
	case env.EyrDialect:
		Eyr_EvalBlockInside(ps, inj, injnow) // TODO ps.Stack is already in ps ... refactor
	case env.Rye0Dialect:
		Rye0_EvalBlockInj(ps, inj, injnow) // TODO ps.Stack is already in ps ... refactor
		// return Rye0_EvaluateBlock(ps) // TODO ps.Stack is already in ps ... refactor
	case env.Rye00Dialect:
		Rye00_EvalBlockInj(ps, inj, injnow) // Simplified dialect for builtins and integers
	default:
		//
	}
}

// HOTPATH
// comma (expression guard) can be present between block-level expressions, in case of injected block they
// reinject the value
func MaybeAcceptComma(ps *env.ProgramState, inj env.Object, injnow bool) bool {
	obj := ps.Ser.Peek()
	if _, ok := obj.(env.Comma); ok {
		ps.Ser.Next()
		if inj != nil {
			injnow = true
		}
	}
	return injnow
}

func EvalBlockInj(ps *env.ProgramState, inj env.Object, injnow bool) {
	// repeats until at the end of the block
	for ps.Ser.Pos() < ps.Ser.Len() {
		injnow = EvalExpressionInj(ps, inj, injnow)
		if tryHandleFailure(ps) {
			return
		}
		// if return flag was raised return ( errorflag I think would return in previous if anyway)
		if ps.ReturnFlag || ps.ErrorFlag {
			// Execute deferred blocks before returning
			if len(ps.DeferBlocks) > 0 {
				fmt.Println("TEMP: EvalBlockInj DeferBlocks triggered")
				// ExecuteDeferredBlocks(ps)
			}
			return
		}
		injnow = MaybeAcceptComma(ps, inj, injnow)
	}
}

// Eval block in specific context and inject a value
func EvalBlockInCtxInj(ps *env.ProgramState, ctx *env.RyeCtx, inj env.Object, injnow bool) {
	ctx2 := ps.Ctx
	ps.Ctx = ctx
	EvalBlockInj(ps, inj, injnow)
	ps.Ctx = ctx2
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
// just quick speculation ... () will also have to work with general evaluator, not just op-words like (add 1 2) it would be best
// if it didn't slow things down, but would just be some limit (on a stack?) of how much further current expression can go.
// ( would add it to stack ) would stop processing another expr and throw error if not all were provided and remove from stack.
//

// EVAL EXPRESSION

// Consolidated evaluation function that handles both regular and injected evaluation
func EvalExpression(ps *env.ProgramState, inj env.Object, injnow bool, limited bool) bool {
	if inj == nil || !injnow {
		EvalExpressionConcrete(ps)
		if ps.ReturnFlag {
			return injnow
		}
	} else {
		ps.Res = inj
		injnow = false
		if ps.ReturnFlag {
			return injnow
		}
	}
	MaybeEvalOpwordOnRight(ps.Ser.Peek(), ps, limited)
	return injnow
}

// Replace EvalExpression2 with a call to EvalExpression
func EvalExpression2(ps *env.ProgramState, limited bool) {
	EvalExpression(ps, nil, false, limited)
}

// Replace EvalExpressionInj with a call to EvalExpression
func EvalExpressionInj(ps *env.ProgramState, inj env.Object, injnow bool) bool {
	return EvalExpression(ps, inj, injnow, false)
}

// Replace EvalExpressionInjLimited with a call to EvalExpression
func EvalExpressionInjLimited(ps *env.ProgramState, inj env.Object, injnow bool) bool {
	return EvalExpression(ps, inj, injnow, true)
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
func MaybeEvalOpwordOnRight(nextObj env.Object, ps *env.ProgramState, limited bool) {
	if nextObj == nil || ps.ReturnFlag || ps.ErrorFlag {
		return
	}
	// exit quickly for most common value types
	objType := nextObj.Type()
	if objType == env.StringType ||
		objType == env.IntegerType ||
		objType == env.BlockType ||
		objType == env.WordType {
		return
	}
	switch opword := nextObj.(type) {
	case env.Opword:
		// val := ps.Ser.Pop()
		ps.Ser.Next()
		EvalWord(ps, opword.ToWord(), ps.Res, false, opword.Force > 0)
		MaybeEvalOpwordOnRight(ps.Ser.Peek(), ps, limited)
		return
	case env.Pipeword:
		if limited {
			return
		}
		ps.Ser.Next()
		EvalWord(ps, opword.ToWord(), ps.Res, false, opword.Force > 0)
		if ps.ReturnFlag {
			return //... not sure if we need this
		}
		MaybeEvalOpwordOnRight(ps.Ser.Peek(), ps, limited)
		return
	case env.LSetword:
		if limited {
			return
		}
		idx := opword.Index
		if ps.AllowMod {
			ok := ps.Ctx.Mod(idx, ps.Res)
			if !ok {
				ps.Res = env.NewError("Cannot modify constant " + ps.Idx.GetWord(idx) + ", use 'var' to declare it as a variable")
				ps.FailureFlag = true
				ps.ErrorFlag = true
				return
			}
		} else {
			ok := ps.Ctx.SetNew(idx, ps.Res, ps.Idx)
			if !ok {
				ps.Res = env.NewError("Can't set already set word " + ps.Idx.GetWord(idx) + ", try using modword (1)")
				ps.FailureFlag = true
				ps.ErrorFlag = true
				return
			}
		}
		ps.Ser.Next()
		MaybeEvalOpwordOnRight(ps.Ser.Peek(), ps, limited)
		return
	case env.LModword:
		if limited {
			return
		}
		idx := opword.Index
		ok := ps.Ctx.Mod(idx, ps.Res)
		if !ok {
			ps.Res = env.NewError("Cannot modify constant " + ps.Idx.GetWord(idx) + ", use 'var' to declare it as a variable")
			ps.FailureFlag = true
			ps.ErrorFlag = true
			return
		}
		ps.Ser.Next()
		MaybeEvalOpwordOnRight(ps.Ser.Peek(), ps, limited)
		return
	case env.CPath:
		if opword.Mode == 1 {
			ps.Ser.Next()
			EvalWord(ps, opword, ps.Res, false, false)
			// when calling cpath
			MaybeEvalOpwordOnRight(ps.Ser.Peek(), ps, limited)
			return
		} else if opword.Mode == 2 {
			if limited {
				return
			}
			ps.Ser.Next()
			EvalWord(ps, opword, ps.Res, false, false) // TODO .. check opword force
			if ps.ReturnFlag {
				return //... not sure if we need this
			}
			MaybeEvalOpwordOnRight(ps.Ser.Peek(), ps, limited)
			return
		}
	}
	return
}

// the main part of evaluator, if it were a polish only we would need almost only this
// switches over all rye values and acts on them
func EvalExpressionConcrete(ps *env.ProgramState) {
	object := ps.Ser.Pop()
	if object == nil {
		ps.ErrorFlag = true
		ps.Res = env.NewError("expected rye value but got to the end of the block")
		return
	}
	objType := object.Type()
	if objType == env.StringType ||
		objType == env.IntegerType ||
		objType == env.DecimalType ||
		objType == env.VoidType ||
		objType == env.UriType ||
		objType == env.EmailType {
		ps.Res = object
		return
	}
	switch object.Type() {
	case env.BlockType:
		block := object.(env.Block)
		// block mode 1 is for eval blocks
		if block.Mode == 0 {
			ps.Res = object
		} else if block.Mode == 1 {
			ser := ps.Ser
			ps.Ser = block.Series
			res := make([]env.Object, 0)
			for ps.Ser.Pos() < ps.Ser.Len() {
				EvalExpression2(ps, false)
				if ps.ReturnFlag || ps.ErrorFlag {
					return
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
		}
	case env.TagwordType:
		ps.Res = *env.NewWord(object.(env.Tagword).Index)
		return
	case env.WordType:
		EvalWord(ps, object.(env.Word), nil, false, false)
		return
	case env.CPathType:
		EvalWord(ps, object, nil, false, false)
		return
	case env.BuiltinType:
		CallBuiltin(object.(env.Builtin), ps, nil, false, false, nil)
		return
	case env.VarBuiltinType:
		CallVarBuiltin(object.(env.VarBuiltin), ps, nil, false, false, nil)
		return
	case env.CurriedCallerType:
		CallCurriedCaller(object.(env.CurriedCaller), ps, nil, false, false, nil)
		return
	// case env.FunctionType: // works just for regular words ... as function
	// 	CallFunction(object.(env.Function), ps, nil, false, nil)
	case env.GenwordType:
		EvalGenword(ps, object.(env.Genword), nil, false)
		return
	case env.SetwordType:
		EvalSetword(ps, object.(env.Setword))
		return
	case env.ModwordType:
		EvalModword(ps, object.(env.Modword))
		return
	case env.GetwordType:
		EvalGetword(ps, object.(env.Getword), nil, false)
		return
	case env.CommaType:
		ps.ErrorFlag = true
		ps.Res = env.NewError("expression guard inside expression")
		return
	case env.ErrorType:
		ps.ErrorFlag = true
		ps.Res = env.NewError("Error Type in code block")
		return
	default:
		fmt.Println(object.Inspect(*ps.Idx))
		ps.ErrorFlag = true
		ps.Res = env.NewError("Unknown rye value in block")
		return
	}
}

// this basicalls returns a rye value behind a word or cpath (context path)
// for words it just looks in to current context and with it to parent contexts
func findWordValue(ps *env.ProgramState, word1 env.Object) (bool, env.Object, *env.RyeCtx) {
	switch word := word1.(type) {
	case env.Word:
		object, found := ps.Ctx.Get(word.Index)
		// if is constant ... stamp it in
		// TODO ... just stamp constants
		// fmt.Println("*")
		// ps.Ser.Put(object)
		// }
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

// EVALUATOR FUNCTIONS FOR SPECIFIC VALUE TYPES

// Evaluates a word
// first tries to find a value in normal context. If there were no generic words this would be mostly it
// if word is not found then it tries to get the value of next expression
// and find a generic word based
// on that, it here is leftval already present it can dispatc on it otherwise
func EvalWord(ps *env.ProgramState, word env.Object, leftVal env.Object, toLeft bool, pipeSecond bool) {
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
					return
				}
				leftVal = ps.Res
				kind = leftVal.GetKind()
			}
		}
		if pipeSecond {
			if !ps.Ser.AtLast() {
				EvalExpressionConcrete(ps)
				if ps.ReturnFlag {
					return
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
		EvalObject(ps, object, leftVal, toLeft, session, pipeSecond, firstVal) //ww0128a *
		return
	} else {
		ps.ErrorFlag = true
		if !ps.FailureFlag {
			ps.Ser.SetPos(pos)
			ps.Res = env.NewError2(5, "word not found: "+word.Print(*ps.Idx))
		}
		return
	}
}

// if word is defined to be generic ... I am not sure we will keep this ... we will decide with more use
// then if explicitly treats it as generic word
func EvalGenword(ps *env.ProgramState, word env.Genword, leftVal env.Object, toLeft bool) {
	EvalExpressionConcrete(ps)

	var arg0 = ps.Res
	object, found := ps.Gen.Get(arg0.GetKind(), word.Index)
	if found {
		EvalObject(ps, object, arg0, toLeft, nil, false, nil) //ww0128a *
		return
	} else {
		ps.ErrorFlag = true
		ps.Res = env.NewError("generic word not found: " + word.Print(*ps.Idx))
		return
	}
}

// evaluates a get-word . it retrieves rye value behid it w/o evaluation
func EvalGetword(ps *env.ProgramState, word env.Getword, leftVal env.Object, toLeft bool) {
	object, found := ps.Ctx.Get(word.Index)
	if found {
		ps.Res = object
		return
	} else {
		ps.ErrorFlag = true
		ps.Res = env.NewError("word not found: " + word.Print(*ps.Idx))
		return
	}
}

// evaluates a rye value, most of them just get returned, except builtins, functions and context paths
func EvalObject(ps *env.ProgramState, object env.Object, leftVal env.Object, toLeft bool, ctx *env.RyeCtx, pipeSecond bool, firstVal env.Object) {
	switch object.Type() {
	case env.BuiltinType:
		bu := object.(env.Builtin)
		if checkForFailureWithBuiltin(bu, ps, 333) {
			return
		}
		CallBuiltin(bu, ps, leftVal, toLeft, pipeSecond, firstVal)
		return
	case env.FunctionType:
		fn := object.(env.Function)
		CallFunction(fn, ps, leftVal, toLeft, ctx)
		return
	case env.CPathType: // RMME
		fn := object.(env.Function)
		CallFunction(fn, ps, leftVal, toLeft, ctx)
		return
	case env.VarBuiltinType:
		bu := object.(env.VarBuiltin)
		if checkForFailureWithVarBuiltin(bu, ps, 333) {
			return
		}
		CallVarBuiltin(bu, ps, leftVal, toLeft, pipeSecond, firstVal)
		return
	case env.CurriedCallerType:
		cc := object.(env.CurriedCaller)
		CallCurriedCaller(cc, ps, leftVal, toLeft, pipeSecond, firstVal)
		return
	default:
		ps.Res = object
	}
}

// evaluates expression to the right and sets the result of it to a word in current context
func EvalSetword(ps *env.ProgramState, word env.Setword) {
	// es1 := EvalExpression(es)
	EvalExpressionInj(ps, nil, false)
	idx := word.Index
	if ps.AllowMod {
		ok := ps.Ctx.Mod(idx, ps.Res)
		if !ok {
			ps.Res = env.NewError("Cannot modify constant " + ps.Idx.GetWord(idx) + ", use 'var' to declare it as a variable")
			ps.FailureFlag = true
			ps.ErrorFlag = true
		}
	} else {
		ok := ps.Ctx.SetNew(idx, ps.Res, ps.Idx)
		if !ok {
			ps.Res = env.NewError("Can't set already set word " + ps.Idx.GetWord(idx) + ", try using modword (2)")
			ps.FailureFlag = true
			ps.ErrorFlag = true
		}
	}
}

// evaluates expression to the right and sets the result of it to a word in current context
func EvalModword(ps *env.ProgramState, word env.Modword) {
	// es1 := EvalExpression(es)
	EvalExpressionInj(ps, nil, false)
	idx := word.Index
	ok := ps.Ctx.Mod(idx, ps.Res)
	if !ok {
		ps.Res = env.NewError("Cannot modify constant " + ps.Idx.GetWord(idx) + ", use 'var' to declare it as a variable")
		ps.FailureFlag = true
		ps.ErrorFlag = true
	}
}

//
// CALLING FUNCTIONS
//

// Consolidated function calling
func CallFunctionWithArgs(fn env.Function, ps *env.ProgramState, ctx *env.RyeCtx, args ...env.Object) {
	ctx = DetermineContext(fn, ps, ctx)
	if ctx == nil {
		return
	}

	switch len(args) {
	case 0:
		CallFunction(fn, ps, nil, false, ctx)
		return
	case 1:
		CallFunction(fn, ps, args[0], false, ctx)
		return
	case 2:
		CallFunctionArgs2(fn, ps, args[0], args[1], ctx)
		return
	case 4:
		CallFunctionArgs4(fn, ps, args[0], args[1], args[2], args[3], ctx)
		return
	default:
		CallFunctionArgsN(fn, ps, ctx, args...)
		return
	}
}

// functionCallPool is a sync.Pool for reusing ProgramState objects specifically for function calls
var envPool = sync.Pool{
	New: func() interface{} {
		return env.NewEnv(nil)
	},
}

// This method is used in the evaluator and takes arguments from code if needed
func CallFunction(fn env.Function, ps *env.ProgramState, arg0 env.Object, toLeft bool, ctx *env.RyeCtx) {
	// fmt.Println(1)

	env0 := ps.Ctx // store reference to current env in local
	var fnCtx *env.RyeCtx
	if ctx != nil { // called via contextpath and this is the context
		//		fmt.Println("if 111")
		if fn.Pure {
			//			fmt.Println("calling pure function")
			//		fmt.Println(es.PCtx)
			fnCtx = envPool.Get().(*env.RyeCtx)
			fnCtx.Clear()
			fnCtx.Parent = ps.PCtx
			// fnCtx = env.NewEnv(ps.PCtx)
		} else {
			if fn.Ctx != nil { // if context was defined at definition time, pass it as parent.
				if fn.InCtx {
					fnCtx = fn.Ctx
				} else {
					fn.Ctx.Parent = ctx
					fnCtx = envPool.Get().(*env.RyeCtx)
					fnCtx.Clear()
					fnCtx.Parent = fn.Ctx
					// fnCtx = env.NewEnv(fn.Ctx)
				}
			} else {
				fnCtx = envPool.Get().(*env.RyeCtx)
				fnCtx.Clear()
				fnCtx.Parent = ctx
				// fnCtx = env.NewEnv(ctx)
			}
		}
	} else {
		//fmt.Println("else1")
		if fn.Pure {
			//		fmt.Println("calling pure function")
			//	fmt.Println(es.PCtx)
			fnCtx = envPool.Get().(*env.RyeCtx)
			fnCtx.Clear()
			fnCtx.Parent = ps.Ctx
			// fnCtx = env.NewEnv(ps.PCtx)
		} else {
			if fn.Ctx != nil { // if context was defined at definition time, pass it as parent.
				// Q: Would we want to pass it directly at any point?
				//    Maybe to remove need of creating new contexts, for reuse, of to be able to modify it?
				fnCtx = envPool.Get().(*env.RyeCtx)
				fnCtx.Clear()
				fnCtx.Parent = fn.Ctx
				// fnCtx = env.NewEnv(fn.Ctx)
			} else {
				fnCtx = envPool.Get().(*env.RyeCtx)
				fnCtx.Clear()
				fnCtx.Parent = env0
				// fnCtx = env.NewEnv(env0)
			}
		}
	}

	// fmt.Println(fnCtx)

	ii := 0
	// evalExprFn := EvalExpression // 2020-01-12 .. changed to ion2
	evalExprFn := EvalExpression2
	if arg0 != nil {
		if fn.Spec.Series.Len() > 0 {
			index := fn.Spec.Series.Get(ii).(env.Word).Index
			fnCtx.Set(index, arg0)
			ps.Args[ii] = index
			ii = 1
			if !toLeft {
				//evalExprFn = EvalExpression_ // 2020-01-12 .. changed to ion2
				evalExprFn = EvalExpression2
			}
		}
	}

	defer func() {
		if len(ps.DeferBlocks) > 0 {
			ExecuteDeferredBlocks(ps)
		}
	}()

	// collect arguments
	for i := ii; i < fn.Argsn; i += 1 {
		evalExprFn(ps, true)
		if ps.ReturnFlag || ps.ErrorFlag {
			return
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

	//	if ctx != nil {
	//		result = EvalBlockInCtx(es, ctx)
	//	} else {
	if arg0 != nil {
		EvalBlockInj(ps, arg0, true)
	} else {
		EvalBlock(ps)
	}
	//	}
	// MaybeDisplayFailureOrError(result, result.Idx, "call function")
	if ps.ForcedResult != nil {
		ps.Res = ps.ForcedResult
		ps.ForcedResult = nil
	}
	ps.Ctx = env0
	ps.Ser = ser0
	ps.ReturnFlag = false
	envPool.Put(fnCtx)

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

// This is used in builtins and works specifically for functions with two arguments
func CallFunctionArgs2(fn env.Function, ps *env.ProgramState, arg0 env.Object, arg1 env.Object, ctx *env.RyeCtx) {
	// fmt.Println(2)
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
	if ps.ReturnFlag || ps.ErrorFlag {
		return
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
	defer func() {
		if len(psX.DeferBlocks) > 0 {
			ExecuteDeferredBlocks(ps)
		}
	}()

	var result *env.ProgramState
	psX.Ser.SetPos(0)
	EvalBlockInj(psX, arg0, true)
	// fmt.Println(result)
	// fmt.Println(result.Res)
	MaybeDisplayFailureOrError(psX, psX.Idx, "call func args 2")
	if psX.ForcedResult != nil {
		ps.Res = result.ForcedResult
		result.ForcedResult = nil
	}
	ps.ReturnFlag = false
}

// This one is called from builtins and calls functions with 4 arguments
func CallFunctionArgs4(fn env.Function, ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, ctx *env.RyeCtx) {
	fmt.Println(3)
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
	if ps.ReturnFlag || ps.ErrorFlag {
		return
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
	defer func() {
		if len(psX.DeferBlocks) > 0 {
			ExecuteDeferredBlocks(ps)
		}
	}()

	EvalBlockInj(psX, arg0, true)
	MaybeDisplayFailureOrError(result, result.Idx, "call func args 4")
	if psX.ForcedResult != nil {
		ps.Res = psX.ForcedResult
		result.ForcedResult = nil
	}
	ps.ReturnFlag = false
}

// Used in builtins ... for variable number of arguments
func CallFunctionArgsN(fn env.Function, ps *env.ProgramState, ctx *env.RyeCtx, args ...env.Object) {
	// fmt.Println(6)
	// ctx = nil
	var fnCtx = DetermineContext(fn, ps, ctx)
	if ps.ReturnFlag || ps.ErrorFlag {
		return
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
	psX.Ser.SetPos(0)
	defer func() {
		if len(psX.DeferBlocks) > 0 {
			ExecuteDeferredBlocks(ps)
		}
	}()

	if len(args) > 0 {
		EvalBlockInj(psX, args[0], true)
	} else {
		EvalBlock(psX)
	}
	MaybeDisplayFailureOrError(ps, ps.Idx, "call func args N")
	if psX.ForcedResult != nil {
		ps.Res = ps.ForcedResult
		ps.ForcedResult = nil
	} else {
		ps.Res = psX.Res
	}
	ps.ReturnFlag = false
}

// Determine the context for CallFunctionArgsVar
func DetermineContext(fn env.Function, ps *env.ProgramState, ctx *env.RyeCtx) *env.RyeCtx {
	// fmt.Println(55)
	var fnCtx *env.RyeCtx
	env0 := ps.Ctx  // store reference to current env in local
	if ctx != nil { // called via contextpath and this is the context
		// fmt.Println("DIREXT CTX 0")
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
		// fmt.Println("DIREXT CTX 1")
		if fn.Pure {
			// fmt.Println("DIREXT CTX 2")
			fnCtx = env.NewEnv(ps.PCtx)
		} else {
			// fmt.Println("DIREXT CTX 3")

			if fn.Ctx != nil { // if context was defined at definition time, pass it as parent.
				// Q: Would we want to pass it directly at any point?
				//    Maybe to remove need of creating new contexts, for reuse, of to be able to modify it?
				if fn.InCtx {
					// fmt.Println("DIREXT CTX 10")
					fnCtx = fn.Ctx
				} else {
					// fn.Ctx.Parent = ctx // 20250225 ... trying to make buttons example work
					fnCtx = env.NewEnv(fn.Ctx)
				}
				// fnCtx = env.NewEnv(fn.Ctx)
			} else {
				fnCtx = env.NewEnv(env0)
			}
		}
	}
	return fnCtx
}

// CALLING BUILTINS

func CallBuiltin(bi env.Builtin, ps *env.ProgramState, arg0_ env.Object, toLeft bool, pipeSecond bool, firstVal env.Object) {
	////args := make([]env.Object, bi.Argsn)
	/*pospos := ps.Ser.GetPos()
	for i := 0; i < bi.Argsn; i += 1 {
		EvalExpression(ps)
		args[i] = ps.Res
	}
	ps.Ser.SetPos(pospos)*/

	// let's try to make it without array allocation and without variadic arguments that also maybe actualizes splice
	var arg0 env.Object
	var arg1 env.Object
	var arg2 env.Object
	var arg3 env.Object
	var arg4 env.Object

	evalExprFn := EvalExpression2
	curry := false

	trace("*** BUILTIN ***")

	if arg0_ != nil && !pipeSecond {
		//fmt.Println("ARG0 = LEFT")
		arg0 = arg0_
		//if !toLeft {
		//fmt.Println("L TO R *** ")
		//evalExprFn = EvalExpression_
		// }
	} else if firstVal != nil && pipeSecond {
		arg0 = firstVal
	} else if bi.Argsn > 0 {
		//fmt.Println(" ARG 1 ")
		//fmt.Println(ps.Ser.GetPos())
		evalExprFn(ps, true)

		if checkForFailureWithBuiltin(bi, ps, 0) {
			return
		}
		if ps.ReturnFlag || ps.ErrorFlag {
			ps.Res = env.NewError4(0, "argument 1 of "+strconv.Itoa(bi.Argsn)+" missing of builtin: '"+bi.Doc+"'", ps.Res.(*env.Error), nil)
			return
		}
		if ps.Res.Type() == env.VoidType {
			curry = true
		} else {
			arg0 = ps.Res
		}
	}

	if arg0_ != nil && pipeSecond {
		arg1 = arg0_
	} else if bi.Argsn > 1 {
		evalExprFn(ps, true) // <---- THESE DETERMINE IF IT CONSUMES WHOLE EXPRESSION OR NOT IN CASE OF PIPEWORDS .. HM*... MAYBE WOULD COULD HAVE A WORD MODIFIER?? a: 2 |add 5 a:: 2 |add 5 print* --TODO

		if checkForFailureWithBuiltin(bi, ps, 1) {
			return
		}
		if ps.ReturnFlag || ps.ErrorFlag {
			ps.Res = env.NewError4(0, "argument 2 of "+strconv.Itoa(bi.Argsn)+" missing of builtin: '"+bi.Doc+"'", ps.Res.(*env.Error), nil)
			return
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

		if checkForFailureWithBuiltin(bi, ps, 2) {
			return
		}
		if ps.ReturnFlag || ps.ErrorFlag {
			ps.Res = env.NewError4(0, "argument 3 missing", ps.Res.(*env.Error), nil)
			return
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
		ps.Res = *env.NewCurriedCallerFromBuiltin(bi, arg0, arg1, arg2, arg3, arg4)
	} else {
		ps.Res = bi.Fn(ps, arg0, arg1, arg2, arg3, arg4)
	}
}

func CallVarBuiltin(bi env.VarBuiltin, ps *env.ProgramState, arg0_ env.Object, toLeft bool, pipeSecond bool, firstVal env.Object) {

	args := make([]env.Object, bi.Argsn)
	ii := 0

	if bi.Argsn > 0 {
		if arg0_ != nil && !pipeSecond {
			args[ii] = arg0_
			ii++
		} else if firstVal != nil && pipeSecond {
			args[ii] = firstVal
			ii++
		} else if bi.Argsn > 0 {
			EvalExpression2(ps, true)
			args[ii] = ps.Res
			ii++
		}

		if arg0_ != nil && pipeSecond {
			args[ii] = arg0_
			ii++
		} else if bi.Argsn > 1 {
			EvalExpression2(ps, true)
			args[ii] = ps.Res
			ii++
		}
		//variadic version
		for i := 2; i < bi.Argsn; i += 1 {
			EvalExpression2(ps, true)
			args[ii] = ps.Res
			ii++

		}
	}

	ps.Res = bi.Fn(ps, args...)
}

func DirectlyCallBuiltin(ps *env.ProgramState, bi env.Builtin, a0 env.Object, a1 env.Object) env.Object {
	// Direct call without currying
	return bi.Fn(ps, a0, a1, nil, nil, nil)
}

// CallCurriedCaller handles calling a CurriedCaller object
func CallCurriedCaller(cc env.CurriedCaller, ps *env.ProgramState, arg0_ env.Object, toLeft bool, pipeSecond bool, firstVal env.Object) {
	var arg0 env.Object = cc.Cur0
	var arg1 env.Object = cc.Cur1
	var arg2 env.Object = cc.Cur2
	var arg3 env.Object = cc.Cur3
	var arg4 env.Object = cc.Cur4

	evalExprFn := EvalExpression2
	curry := false

	// Process arguments based on what's already curried
	if arg0_ != nil && !pipeSecond {
		if arg0 == nil {
			arg0 = arg0_
		} else if arg1 == nil {
			arg1 = arg0_
		} else if arg2 == nil {
			arg2 = arg0_
		} else if arg3 == nil {
			arg3 = arg0_
		} else if arg4 == nil {
			arg4 = arg0_
		}
	} else if firstVal != nil && pipeSecond {
		if arg0 == nil {
			arg0 = firstVal
		} else if arg1 == nil {
			arg1 = firstVal
		} else if arg2 == nil {
			arg2 = firstVal
		} else if arg3 == nil {
			arg3 = firstVal
		} else if arg4 == nil {
			arg4 = firstVal
		}
	}

	// Collect any remaining arguments needed
	if cc.Argsn > 0 && arg0 == nil {
		evalExprFn(ps, true)
		if ps.ReturnFlag || ps.ErrorFlag {
			return
		}
		if ps.Res.Type() == env.VoidType {
			curry = true
		} else {
			arg0 = ps.Res
		}
	}

	if arg0_ != nil && pipeSecond {
		if arg1 == nil {
			arg1 = arg0_
		} else if arg2 == nil {
			arg2 = arg0_
		} else if arg3 == nil {
			arg3 = arg0_
		} else if arg4 == nil {
			arg4 = arg0_
		}
	} else if cc.Argsn > 1 && arg1 == nil {
		evalExprFn(ps, true)
		if ps.ReturnFlag || ps.ErrorFlag {
			return
		}
		if ps.Res.Type() == env.VoidType {
			curry = true
		} else {
			arg1 = ps.Res
		}
	}

	if cc.Argsn > 2 && arg2 == nil {
		evalExprFn(ps, true)
		if ps.ReturnFlag || ps.ErrorFlag {
			return
		}
		if ps.Res.Type() == env.VoidType {
			curry = true
		} else {
			arg2 = ps.Res
		}
	}

	if cc.Argsn > 3 && arg3 == nil {
		evalExprFn(ps, true)
		if ps.Res.Type() == env.VoidType {
			curry = true
		} else {
			arg3 = ps.Res
		}
	}

	if cc.Argsn > 4 && arg4 == nil {
		evalExprFn(ps, true)
		if ps.Res.Type() == env.VoidType {
			curry = true
		} else {
			arg4 = ps.Res
		}
	}

	if curry {
		// Create a new CurriedCaller with updated arguments
		if cc.CallerType == 0 {
			ps.Res = *env.NewCurriedCallerFromBuiltin(*cc.Builtin, arg0, arg1, arg2, arg3, arg4)
		} else {
			ps.Res = *env.NewCurriedCallerFromFunction(*cc.Function, arg0, arg1, arg2, arg3, arg4)
		}
	} else {
		// Execute the function with all arguments
		if cc.CallerType == 0 {
			ps.Res = cc.Builtin.Fn(ps, arg0, arg1, arg2, arg3, arg4)
		} else {
			// Call the function with the arguments
			CallFunctionArgsN(*cc.Function, ps, nil, arg0, arg1, arg2, arg3, arg4)
		}
	}
}

// DISPLAYING FAILURE OR ERRROR

func MaybeDisplayFailureOrError(es *env.ProgramState, genv *env.Idxs, tag string) {
	if es.FailureFlag {
		fmt.Println("\x1b[33m" + "Failure" + "\x1b[0m")
		// DEBUG: fmt.Println(tag)
	}
	if es.ErrorFlag {
		fmt.Println("\x1b[31m" + es.Res.Print(*genv))
		switch es.Res.(type) {
		case env.Error:
			fmt.Println(es.Ser.PositionAndSurroundingElements(*genv))
			fmt.Println("Error not pointer so bug. #temp")
		case *env.Error:
			fmt.Println("At location::")
			fmt.Print(es.Ser.PositionAndSurroundingElements(*genv))
		}
		fmt.Println("\x1b[0m")
		// fmt.Println(tag)
		// ENTER CONSOLE ON ERROR
		// es.ErrorFlag = false
		// es.FailureFlag = false
		// DoRyeRepl(es, "do", true)
	}
	// cebelca2659- vklopi kontne skupine
}

func MaybeDisplayFailureOrErrorWASM(es *env.ProgramState, genv *env.Idxs, printfn func(string), tag string) {
	if es.FailureFlag {
		printfn("\x1b[33m" + "Failure" + "\x1b[0m")
		printfn(tag)
	}
	if es.ErrorFlag {
		printfn("\x1b[31;3m" + es.Res.Print(*genv))
		switch es.Res.(type) {
		case env.Error:
			printfn(es.Ser.PositionAndSurroundingElements(*genv))
			printfn("Error not pointer so bug. #temp")
		case *env.Error:
			printfn("At location:")
			printfn(es.Ser.PositionAndSurroundingElements(*genv))
		}
		printfn("\x1b[0m")
		printfn(tag)
	}
}

//  CHECKING VARIOUS FLAGS

// Replace individual flag checking functions with calls to checkFlags
func checkForFailureWithBuiltin(bi env.Builtin, ps *env.ProgramState, n int) bool {
	if ps.FailureFlag && !bi.AcceptFailure {
		ps.ErrorFlag = true
		return true
	}
	return false
}

func checkForFailureWithVarBuiltin(bi env.VarBuiltin, ps *env.ProgramState, n int) bool {
	if ps.FailureFlag && !bi.AcceptFailure {
		ps.ErrorFlag = true
		return true
	}
	return false
}

func trace(s string) {

}

func tryHandleFailure(ps *env.ProgramState) bool {
	if ps.FailureFlag && !ps.ReturnFlag && !ps.InErrHandler {
		if checkContextErrorHandler(ps) {
			return false // Successfully handled
		}
		ps.ErrorFlag = true
		return true // Unhandled failure
	}
	return false // No failure
}

// ExecuteDeferredBlocks executes all deferred blocks in LIFO order (last in, first out)
// and clears the deferred blocks list
func ExecuteDeferredBlocks(ps *env.ProgramState) {
	// TODO: Implement deferred block execution
}

// Remove unused debugging functions
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
}

/* func tryHandleFailure_OLD(ps *env.ProgramState, n int) bool {
	if ps.FailureFlag && !ps.ReturnFlag {
		if !ps.InErrHandler {
			if checkContextErrorHandler(ps) {
				return false
			}
		}
		ps.ErrorFlag = true
		return true
	}
	return false
}
*/

/* // Consolidated flag checking function
func checkFlags(ps *env.ProgramState, n int, flags ...bool) bool {
	if ps.ReturnFlag || ps.ErrorFlag {
		return true
	}
	if ps.FailureFlag {
		ps.ErrorFlag = true
		return true
	}
	for _, flag := range flags {
		if flag {
			return true
		}
	}
	return false
} */
