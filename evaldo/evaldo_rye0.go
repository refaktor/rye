// evaldo_rye0.go - A simplified interpreter for the Rye language
package evaldo

import (
	"fmt"
	"strconv"

	"github.com/refaktor/rye/env"
)

// Rye0_EvalBlockInj evaluates a block with an optional injected value.
// Parameters:
//   - ps: The program state
//   - inj: The value to inject (can be nil)
//   - injnow: Whether to inject the value now
//
// Returns:
//   - The updated program state
func Rye0_EvalBlockInj(ps *env.ProgramState, inj env.Object, injnow bool) *env.ProgramState {
	// Repeat until at the end of the block
	for ps.Ser.Pos() < ps.Ser.Len() {
		// Evaluate expression at the block cursor
		ps, injnow = Rye0_EvalExpressionInj(ps, inj, injnow)

		// If flags raised return program state
		if Rye0_checkFlagsAfterBlock(ps, 101) {
			return ps
		}
		if Rye0_checkErrorReturnFlag(ps) {
			return ps
		}
	}
	return ps
}

// Rye0_EvalExpression2 evaluates an expression with optional limitations.
// Parameters:
//   - ps: The program state
//   - limited: Whether evaluation is limited
//
// Returns:
//   - The updated program state
func Rye0_EvalExpression2(ps *env.ProgramState, limited bool) *env.ProgramState {
	Rye0_EvalExpressionConcrete(ps)
	return ps
}

// Rye0_EvalExpressionInj evaluates an expression with an optional injected value.
// Parameters:
//   - ps: The program state
//   - inj: The value to inject (can be nil)
//   - injnow: Whether to inject the value now
//
// Returns:
//   - The updated program state and updated injnow flag
func Rye0_EvalExpressionInj(ps *env.ProgramState, inj env.Object, injnow bool) (*env.ProgramState, bool) {
	var esleft *env.ProgramState
	if inj == nil || !injnow {
		// If there is no injected value just eval the concrete expression
		esleft = Rye0_EvalExpressionConcrete(ps)
		if ps.ReturnFlag {
			return ps, injnow
		}
	} else {
		// Otherwise set program state to specific one and injected value to result
		// Set injnow to false and if return flag return
		esleft = ps
		esleft.Res = inj
		injnow = false
	}
	return ps, injnow
}

// Rye0_EvalExpressionConcrete evaluates a concrete expression.
// This is the main part of the evaluator that handles all Rye value types.
// Parameters:
//   - ps: The program state
//
// Returns:
//   - The updated program state
func Rye0_EvalExpressionConcrete(ps *env.ProgramState) *env.ProgramState {
	object := ps.Ser.Pop()

	if object == nil {
		ps.ErrorFlag = true
		ps.Res = env.NewError("Expected Rye value but it's missing")
		return ps
	}

	switch object.Type() {
	// Literal values evaluate to themselves
	case env.IntegerType, env.DecimalType, env.StringType, env.VoidType, env.UriType, env.EmailType:
		if !ps.SkipFlag {
			ps.Res = object
		}

	// Block handling
	case env.BlockType:
		if !ps.SkipFlag {
			return Rye0_EvaluateBlock(ps, object.(env.Block))
		}

	// Word types
	case env.TagwordType:
		ps.Res = *env.NewWord(object.(env.Tagword).Index)
		return ps
	case env.WordType:
		return Rye0_EvalWord(ps, object.(env.Word), nil, false, false)
	case env.CPathType:
		return Rye0_EvalWord(ps, object, nil, false, false)
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

	// Error handling
	case env.CommaType:
		ps.ErrorFlag = true
		ps.Res = env.NewError("Expression guard inside expression")
	case env.ErrorType:
		ps.ErrorFlag = true
		ps.Res = env.NewError("Error object encountered")

	// Unknown type
	default:
		fmt.Println(object.Inspect(*ps.Idx))
		ps.ErrorFlag = true
		ps.Res = env.NewError("Unknown Rye value type: " + strconv.Itoa(int(object.Type())))
	}

	return ps
}

// Rye0_EvaluateBlock handles the evaluation of a block object.
// Extracted from Rye0_EvalExpressionConcrete to reduce function size.
// Parameters:
//   - ps: The program state
//   - block: The block to evaluate
//
// Returns:
//   - The updated program state
func Rye0_EvaluateBlock(ps *env.ProgramState, block env.Block) *env.ProgramState {
	// Block mode 1 is for eval blocks
	if block.Mode == 1 {
		ser := ps.Ser
		ps.Ser = block.Series
		res := make([]env.Object, 0)
		for ps.Ser.Pos() < ps.Ser.Len() {
			Rye0_EvalExpression2(ps, false)
			if Rye0_checkErrorReturnFlag(ps) {
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
	} else {
		ps.Res = block
	}
	return ps
}

// Rye0_findWordValue returns the value associated with a word in the current context.
// Parameters:
//   - ps: The program state
//   - word1: The word to look up
//
// Returns:
//   - found: Whether the word was found
//   - object: The value associated with the word
//   - ctx: The context where the word was found
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
		return Rye0_findCPathValue(ps, word)
	default:
		return false, nil, nil
	}
}

// Rye0_findCPathValue handles the lookup of context path values.
// Extracted from Rye0_findWordValue to improve readability.
// Parameters:
//   - ps: The program state
//   - word: The context path to look up
//
// Returns:
//   - found: Whether the path was found
//   - object: The value at the path
//   - ctx: The context where the value was found
func Rye0_findCPathValue(ps *env.ProgramState, word env.CPath) (bool, env.Object, *env.RyeCtx) {
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
}

// Rye0_EvalWord evaluates a word in the current context.
// Parameters:
//   - ps: The program state
//   - word: The word to evaluate
//   - leftVal: The left value (if any)
//   - toLeft: Whether the evaluation is to the left
//   - pipeSecond: Whether this is a pipe second evaluation
//
// Returns:
//   - The updated program state
func Rye0_EvalWord(ps *env.ProgramState, word env.Object, leftVal env.Object, toLeft bool, pipeSecond bool) *env.ProgramState {
	var firstVal env.Object
	found, object, session := Rye0_findWordValue(ps, word)
	pos := ps.Ser.GetPos()

	if !found {
		// Look at Generic words, but first check type
		kind := 0
		if leftVal != nil {
			kind = leftVal.GetKind()
		}

		// Handle different evaluation scenarios
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
			}
		}

		// Try to find a generic word
		if rword, ok := word.(env.Word); ok && leftVal != nil && ps.Ctx.Kind.Index != -1 {
			object, found = ps.Gen.Get(kind, rword.Index)
		}
	}

	if found {
		return Rye0_EvalObject(ps, object, leftVal, toLeft, session, pipeSecond, firstVal)
	} else {
		ps.ErrorFlag = true
		if !ps.FailureFlag {
			ps.Ser.SetPos(pos)
			ps.Res = env.NewError2(5, "Word not found: "+word.Print(*ps.Idx))
		}
		return ps
	}
}

// Rye0_EvalGenword evaluates a generic word.
// Parameters:
//   - ps: The program state
//   - word: The generic word to evaluate
//   - leftVal: The left value (if any)
//   - toLeft: Whether the evaluation is to the left
//
// Returns:
//   - The updated program state
func Rye0_EvalGenword(ps *env.ProgramState, word env.Genword, leftVal env.Object, toLeft bool) *env.ProgramState {
	Rye0_EvalExpressionConcrete(ps)

	var arg0 = ps.Res
	object, found := ps.Gen.Get(arg0.GetKind(), word.Index)
	if found {
		return Rye0_EvalObject(ps, object, arg0, toLeft, nil, false, nil)
	} else {
		ps.ErrorFlag = true
		ps.Res = env.NewError("Generic word not found: " + word.Print(*ps.Idx))
		return ps
	}
}

// Rye0_EvalGetword evaluates a get-word.
// Parameters:
//   - ps: The program state
//   - word: The get-word to evaluate
//   - leftVal: The left value (if any)
//   - toLeft: Whether the evaluation is to the left
//
// Returns:
//   - The updated program state
func Rye0_EvalGetword(ps *env.ProgramState, word env.Getword, leftVal env.Object, toLeft bool) *env.ProgramState {
	object, found := ps.Ctx.Get(word.Index)
	if found {
		ps.Res = object
		return ps
	} else {
		ps.ErrorFlag = true
		ps.Res = env.NewError("Word not found: " + word.Print(*ps.Idx))
		return ps
	}
}

// Rye0_EvalObject evaluates a Rye object.
// Parameters:
//   - ps: The program state
//   - object: The object to evaluate
//   - leftVal: The left value (if any)
//   - toLeft: Whether the evaluation is to the left
//   - ctx: The context (if any)
//   - pipeSecond: Whether this is a pipe second evaluation
//   - firstVal: The first value (if any)
//
// Returns:
//   - The updated program state
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

		if Rye0_checkFlagsBi(bu, ps, 333) {
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

// Rye0_EvalSetword evaluates a set-word.
// Parameters:
//   - ps: The program state
//   - word: The set-word to evaluate
//
// Returns:
//   - The updated program state
func Rye0_EvalSetword(ps *env.ProgramState, word env.Setword) *env.ProgramState {
	ps1, _ := Rye0_EvalExpressionInj(ps, nil, false)
	idx := word.Index
	if ps.AllowMod {
		ps1.Ctx.Mod(idx, ps.Res)
	} else {
		ok := ps1.Ctx.SetNew(idx, ps1.Res, ps.Idx)
		if !ok {
			ps.Res = env.NewError("Can't set already set word " + ps.Idx.GetWord(idx) + ", try using modword (::)")
			ps.FailureFlag = true
			ps.ErrorFlag = true
		}
	}
	return ps
}

// Rye0_EvalModword evaluates a mod-word.
// Parameters:
//   - ps: The program state
//   - word: The mod-word to evaluate
//
// Returns:
//   - The updated program state
func Rye0_EvalModword(ps *env.ProgramState, word env.Modword) *env.ProgramState {
	ps1, _ := Rye0_EvalExpressionInj(ps, nil, false)
	idx := word.Index
	ps1.Ctx.Mod(idx, ps1.Res)
	return ps1
}

// Rye0_DetermineContext determines the appropriate context for a function call.
// Parameters:
//   - fn: The function
//   - ps: The program state
//   - ctx: The context (if any)
//
// Returns:
//   - The determined context
func Rye0_DetermineContext(fn env.Function, ps *env.ProgramState, ctx *env.RyeCtx) *env.RyeCtx {
	var fnCtx *env.RyeCtx
	env0 := ps.Ctx

	if ctx != nil { // Called via context path
		if fn.Pure {
			fnCtx = env.NewEnv(ps.PCtx)
		} else if fn.Ctx != nil { // If context was defined at definition time
			if fn.InCtx {
				fnCtx = fn.Ctx
			} else {
				fn.Ctx.Parent = ctx
				fnCtx = env.NewEnv(fn.Ctx)
			}
		} else {
			fnCtx = env.NewEnv(ctx)
		}
	} else {
		if fn.Pure {
			fnCtx = env.NewEnv(ps.PCtx)
		} else if fn.Ctx != nil { // If context was defined at definition time
			if fn.InCtx {
				fnCtx = fn.Ctx
			} else {
				fnCtx = env.NewEnv(fn.Ctx)
			}
		} else {
			fnCtx = env.NewEnv(env0)
		}
	}

	return fnCtx
}

// Rye0_CallFunction calls a function.
// Parameters:
//   - fn: The function to call
//   - ps: The program state
//   - arg0: The first argument (if any)
//   - toLeft: Whether the call is to the left
//   - ctx: The context (if any)
//
// Returns:
//   - The updated program state
func Rye0_CallFunction(fn env.Function, ps *env.ProgramState, arg0 env.Object, toLeft bool, ctx *env.RyeCtx) *env.ProgramState {
	// Determine the function context
	fnCtx := Rye0_DetermineContext(fn, ps, ctx)
	env0 := ps.Ctx

	// Set up arguments
	ii := 0
	evalExprFn := Rye0_EvalExpression2

	if arg0 != nil && fn.Spec.Series.Len() > 0 {
		index := fn.Spec.Series.Get(ii).(env.Word).Index
		fnCtx.Set(index, arg0)
		ps.Args[ii] = index
		ii = 1
	}

	// Collect arguments
	for i := ii; i < fn.Argsn; i++ {
		ps = evalExprFn(ps, true)
		if Rye0_checkErrorReturnFlag(ps) {
			return ps
		}

		// Safely get the word index
		specObj := fn.Spec.Series.Get(i)
		if word, ok := specObj.(env.Word); ok {
			index := word.Index
			fnCtx.Set(index, ps.Res)
			if i == 0 {
				arg0 = ps.Res
			}
			ps.Args[i] = index
		} else {
			ps.ErrorFlag = true
			ps.Res = env.NewError("Expected Word in function spec but got: " + specObj.Inspect(*ps.Idx))
			return ps
		}
	}

	// Save current state
	ser0 := ps.Ser
	ps.Ser = fn.Body.Series
	ps.Ctx = fnCtx

	// Evaluate the function body
	var result *env.ProgramState
	if arg0 != nil {
		result = Rye0_EvalBlockInj(ps, arg0, true)
	} else {
		result = EvalBlock(ps)
	}

	// Process the result
	if result.ForcedResult != nil {
		ps.Res = result.ForcedResult
		result.ForcedResult = nil
	} else {
		ps.Res = result.Res
	}

	// Restore state
	ps.Ctx = env0
	ps.Ser = ser0
	ps.ReturnFlag = false

	return ps
}

// Rye0_CallFunctionWithArgs calls a function with the given arguments.
// This consolidates the functionality of Rye0_CallFunctionArgs2, Rye0_CallFunctionArgs4, and Rye0_CallFunctionArgsN.
// Parameters:
//   - fn: The function to call
//   - ps: The program state
//   - ctx: The context (if any)
//   - args: The arguments to pass to the function
//
// Returns:
//   - The updated program state
func Rye0_CallFunctionWithArgs(fn env.Function, ps *env.ProgramState, ctx *env.RyeCtx, args ...env.Object) *env.ProgramState {
	// Determine the function context
	fnCtx := Rye0_DetermineContext(fn, ps, ctx)

	// Check for errors
	if Rye0_checkErrorReturnFlag(ps) {
		return ps
	}

	// Set arguments
	for i, arg := range args {
		if i < fn.Spec.Series.Len() {
			specObj := fn.Spec.Series.Get(i)
			if word, ok := specObj.(env.Word); ok {
				fnCtx.Set(word.Index, arg)
			} else {
				ps.ErrorFlag = true
				ps.Res = env.NewError("Expected Word in function spec but got: " + specObj.Inspect(*ps.Idx))
				return ps
			}
		} else {
			ps.ErrorFlag = true
			ps.Res = env.NewError("Too many arguments provided to function")
			return ps
		}
	}

	// Create a new program state for evaluation
	psX := env.NewProgramState(fn.Body.Series, ps.Idx)
	psX.Ctx = fnCtx
	psX.PCtx = ps.PCtx
	psX.Gen = ps.Gen

	// Evaluate the function body
	var result *env.ProgramState
	psX.Ser.SetPos(0)
	if len(args) > 0 {
		result = Rye0_EvalBlockInj(psX, args[0], true)
	} else {
		result = EvalBlock(psX)
	}

	// Process the result
	Rye0_MaybeDisplayFailureOrError(result, result.Idx, "call func with args")
	if result.ForcedResult != nil {
		ps.Res = result.ForcedResult
		result.ForcedResult = nil
	} else {
		ps.Res = result.Res
	}

	ps.ReturnFlag = false
	return ps
}

// Rye0_CallBuiltin calls a builtin function.
// Parameters:
//   - bi: The builtin to call
//   - ps: The program state
//   - arg0_: The first argument (if any)
//   - toLeft: Whether the call is to the left
//   - pipeSecond: Whether this is a pipe second call
//   - firstVal: The first value (if any)
//
// Returns:
//   - The updated program state
func Rye0_CallBuiltin(bi env.Builtin, ps *env.ProgramState, arg0_ env.Object, toLeft bool, pipeSecond bool, firstVal env.Object) *env.ProgramState {
	// Initialize arguments
	arg0 := bi.Cur0
	arg1 := bi.Cur1
	arg2 := bi.Cur2
	arg3 := bi.Cur3
	arg4 := bi.Cur4

	evalExprFn := Rye0_EvalExpression2

	// Process arguments
	if bi.Argsn > 0 && bi.Cur0 == nil {
		evalExprFn(ps, true)

		if Rye0_checkFlagsBi(bi, ps, 0) {
			return ps
		}
		if Rye0_checkErrorReturnFlag(ps) {
			ps.Res = env.NewError4(0, "Argument 1 of "+strconv.Itoa(bi.Argsn)+" missing for builtin: '"+bi.Doc+"'", ps.Res.(*env.Error), nil)
			return ps
		}
		arg0 = ps.Res
	}

	if bi.Argsn > 1 && bi.Cur1 == nil {
		evalExprFn(ps, true)

		if Rye0_checkFlagsBi(bi, ps, 1) {
			return ps
		}
		if Rye0_checkErrorReturnFlag(ps) {
			ps.Res = env.NewError4(0, "Argument 2 of "+strconv.Itoa(bi.Argsn)+" missing for builtin: '"+bi.Doc+"'", ps.Res.(*env.Error), nil)
			return ps
		}
		arg1 = ps.Res
	}

	if bi.Argsn > 2 {
		evalExprFn(ps, true)

		if Rye0_checkFlagsBi(bi, ps, 2) {
			return ps
		}
		if Rye0_checkErrorReturnFlag(ps) {
			ps.Res = env.NewError4(0, "Argument 3 missing for builtin: '"+bi.Doc+"'", ps.Res.(*env.Error), nil)
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

	// Call the builtin function
	ps.Res = bi.Fn(ps, arg0, arg1, arg2, arg3, arg4)
	return ps
}

// Rye0_DirectlyCallBuiltin directly calls a builtin function with the given arguments.
// Parameters:
//   - ps: The program state
//   - bi: The builtin to call
//   - a0: The first argument
//   - a1: The second argument
//
// Returns:
//   - The result of the builtin function
func Rye0_DirectlyCallBuiltin(ps *env.ProgramState, bi env.Builtin, a0 env.Object, a1 env.Object) env.Object {
	// Determine arguments based on curried values
	var arg0, arg1 env.Object

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

	// Call the builtin function
	return bi.Fn(ps, arg0, arg1, bi.Cur2, bi.Cur3, bi.Cur4)
}

// Rye0_MaybeDisplayFailureOrError displays failure or error information if present.
// Parameters:
//   - es: The program state
//   - genv: The index environment
//   - tag: A tag for the error message
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
	}
}

// Rye0_MaybeDisplayFailureOrErrorWASM displays failure or error information for WASM environment.
// Parameters:
//   - es: The program state
//   - genv: The index environment
//   - printfn: The print function to use
//   - tag: A tag for the error message
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

// Rye0_checkFlagsBi checks if there are failure flags and handles them appropriately.
// Parameters:
//   - bi: The builtin function
//   - ps: The program state
//   - n: The argument number
//
// Returns:
//   - Whether an error was found and handled
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

// Rye0_checkContextErrorHandler checks if there is an error handler defined in the context.
// Parameters:
//   - ps: The program state
//
// Returns:
//   - Whether an error handler was found and executed
func Rye0_checkContextErrorHandler(ps *env.ProgramState) bool {
	// Check if there is error-handler word defined in context (or parent)
	erh, w_exists := ps.Idx.GetIndex("error-handler")
	if !w_exists {
		return false
	}

	// If it exists, get the block
	handler, exists := ps.Ctx.Get(erh)
	if !exists {
		return false
	}

	// Execute the error handler
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

// Rye0_checkFlagsAfterBlock checks if there are failure flags after evaluating a block.
// Parameters:
//   - ps: The program state
//   - n: The block number
//
// Returns:
//   - Whether an error was found and handled
func Rye0_checkFlagsAfterBlock(ps *env.ProgramState, n int) bool {
	trace2("CHECK FLAGS AFTER BLOCKS")
	trace2(n)

	if ps.FailureFlag && !ps.ReturnFlag {
		trace2("FailureFlag")
		trace2("Fail->Error.")

		if !ps.InErrHandler {
			if Rye0_checkContextErrorHandler(ps) {
				return false // Error should be picked up in the handler block
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

// Rye0_checkErrorReturnFlag checks if there are error or return flags.
// Parameters:
//   - ps: The program state
//
// Returns:
//   - Whether an error or return flag was found
func Rye0_checkErrorReturnFlag(ps *env.ProgramState) bool {
	if ps.ErrorFlag {
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
