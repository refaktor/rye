// evaldo_rye0.go - A simplified interpreter for the Rye language
package evaldo

import (
	"fmt"
	"strconv"

	"github.com/refaktor/rye/env"
)

// Rye0_EvalBlockInj evaluates a block with an optional injected value.
func Rye0_EvalBlockInj(ps *env.ProgramState, inj env.Object, injnow bool) *env.ProgramState {
	for ps.Ser.Pos() < ps.Ser.Len() {
		ps, injnow = Rye0_EvalExpressionInj(ps, inj, injnow)

		if Rye0_checkFlagsAfterBlock(ps, 101) || Rye0_checkErrorReturnFlag(ps) {
			return ps
		}
	}
	return ps
}

// Rye0_EvalExpression2 evaluates an expression with optional limitations.
func Rye0_EvalExpression2(ps *env.ProgramState, limited bool) *env.ProgramState {
	Rye0_EvalExpressionConcrete(ps)
	return ps
}

// Rye0_EvalExpressionInj evaluates an expression with an optional injected value.
func Rye0_EvalExpressionInj(ps *env.ProgramState, inj env.Object, injnow bool) (*env.ProgramState, bool) {
	if inj == nil || !injnow {
		// If there is no injected value just eval the concrete expression
		Rye0_EvalExpressionConcrete(ps)
		if ps.ReturnFlag {
			return ps, injnow
		}
	} else {
		// Otherwise set injected value to result and reset injnow flag
		ps.Res = inj
		injnow = false
	}
	return ps, injnow
}

// Rye0_EvalExpressionConcrete evaluates a concrete expression.
// This is the main part of the evaluator that handles all Rye value types.
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
		setError(ps, "Expression guard inside expression")
	case env.ErrorType:
		setError(ps, "Error object encountered")

	// Unknown type
	default:
		fmt.Println(object.Inspect(*ps.Idx))
		setError(ps, "Unknown Rye value type: "+strconv.Itoa(int(object.Type())))
	}

	return ps
}

// Helper function to set error state
func setError(ps *env.ProgramState, message string) {
	ps.ErrorFlag = true
	ps.Res = env.NewError(message)
}

// Rye0_EvaluateBlock handles the evaluation of a block object.
func Rye0_EvaluateBlock(ps *env.ProgramState, block env.Block) *env.ProgramState {
	// Save original series to restore later
	ser := ps.Ser

	switch block.Mode {
	case 1: // Eval blocks
		ps.Ser = block.Series
		res := make([]env.Object, 0)
		for ps.Ser.Pos() < ps.Ser.Len() {
			Rye0_EvalExpression2(ps, false)
			if Rye0_checkErrorReturnFlag(ps) {
				ps.Ser = ser // Restore original series
				return ps
			}
			res = append(res, ps.Res)
		}
		ps.Ser = ser // Restore original series
		ps.Res = *env.NewBlock(*env.NewTSeries(res))
	case 2:
		ps.Ser = block.Series
		EvalBlock(ps)
		ps.Ser = ser // Restore original series
	default:
		ps.Res = block
	}
	return ps
}

// Rye0_findWordValue returns the value associated with a word in the current context.
func Rye0_findWordValue(ps *env.ProgramState, word1 env.Object) (bool, env.Object, *env.RyeCtx) {
	switch word := word1.(type) {
	case env.Word, env.Opword, env.Pipeword:
		// Handle all simple word types the same way
		var index int
		switch w := word.(type) {
		case env.Word:
			index = w.Index
		case env.Opword:
			index = w.Index
		case env.Pipeword:
			index = w.Index
		}
		object, found := ps.Ctx.Get(index)
		return found, object, nil
	case env.CPath:
		return Rye0_findCPathValue(ps, word)
	default:
		return false, nil, nil
	}
}

// Rye0_findCPathValue handles the lookup of context path values.
func Rye0_findCPathValue(ps *env.ProgramState, word env.CPath) (bool, env.Object, *env.RyeCtx) {
	currCtx := ps.Ctx
	i := 1

	for i <= word.Cnt {
		currWord := word.GetWordNumber(i)
		object, found := currCtx.Get(currWord.Index)

		if !found {
			return false, nil, nil
		}

		if word.Cnt > i {
			switch swObj := object.(type) {
			case env.RyeCtx:
				currCtx = &swObj
				i++
			case env.Dict:
				return found, *env.NewString("asdsad"), currCtx
			default:
				return false, nil, nil
			}
		} else {
			return found, object, currCtx
		}
	}

	return false, nil, nil
}

// Rye0_EvalWord evaluates a word in the current context.
func Rye0_EvalWord(ps *env.ProgramState, word env.Object, leftVal env.Object, toLeft bool, pipeSecond bool) *env.ProgramState {
	var firstVal env.Object
	found, object, session := Rye0_findWordValue(ps, word)
	pos := ps.Ser.GetPos()

	if !found {
		// Determine the kind for generic word lookup
		kind := 0
		if leftVal != nil {
			kind = leftVal.GetKind()
		}

		// Evaluate next expression if needed
		if (leftVal == nil && !pipeSecond) || pipeSecond {
			if !ps.Ser.AtLast() {
				Rye0_EvalExpressionConcrete(ps)
				if ps.ReturnFlag {
					return ps
				}

				if pipeSecond {
					firstVal = ps.Res
					kind = firstVal.GetKind()
				} else {
					leftVal = ps.Res
					kind = leftVal.GetKind()
				}
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
		if !ps.FailureFlag {
			ps.Ser.SetPos(pos)
			setError(ps, "Word not found: "+word.Print(*ps.Idx))
		} else {
			ps.ErrorFlag = true
		}
		return ps
	}
}

// Rye0_EvalGenword evaluates a generic word.
func Rye0_EvalGenword(ps *env.ProgramState, word env.Genword, leftVal env.Object, toLeft bool) *env.ProgramState {
	Rye0_EvalExpressionConcrete(ps)

	var arg0 = ps.Res
	object, found := ps.Gen.Get(arg0.GetKind(), word.Index)
	if found {
		return Rye0_EvalObject(ps, object, arg0, toLeft, nil, false, nil)
	} else {
		setError(ps, "Generic word not found: "+word.Print(*ps.Idx))
		return ps
	}
}

// Rye0_EvalGetword evaluates a get-word.
func Rye0_EvalGetword(ps *env.ProgramState, word env.Getword, leftVal env.Object, toLeft bool) *env.ProgramState {
	object, found := ps.Ctx.Get(word.Index)
	if found {
		ps.Res = object
		return ps
	} else {
		setError(ps, "Word not found: "+word.Print(*ps.Idx))
		return ps
	}
}

// Rye0_EvalObject evaluates a Rye object.
func Rye0_EvalObject(ps *env.ProgramState, object env.Object, leftVal env.Object, toLeft bool, ctx *env.RyeCtx, pipeSecond bool, firstVal env.Object) *env.ProgramState {
	switch object.Type() {
	case env.FunctionType, env.CPathType: // Handle both function types the same way
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
	// Fast path: If all arguments are already available (curried), call directly
	if (bi.Argsn == 0) ||
		(bi.Argsn == 1 && bi.Cur0 != nil) ||
		(bi.Argsn == 2 && bi.Cur0 != nil && bi.Cur1 != nil) {
		ps.Res = bi.Fn(ps, bi.Cur0, bi.Cur1, bi.Cur2, bi.Cur3, bi.Cur4)
		return ps
	}

	// Initialize arguments with curried values
	arg0 := bi.Cur0
	arg1 := bi.Cur1
	arg2 := bi.Cur2
	arg3 := bi.Cur3
	arg4 := bi.Cur4

	// Process first argument if needed
	if bi.Argsn > 0 && bi.Cur0 == nil {
		// Direct call to avoid function pointer indirection
		Rye0_EvalExpression2(ps, true)

		// Inline error checking for speed
		if ps.FailureFlag {
			if !bi.AcceptFailure {
				updateErrorCodeContext(ps)
				ps.ErrorFlag = true
				return ps
			}
		}

		if ps.ErrorFlag || ps.ReturnFlag {
			ps.Res = env.NewError4(0, "Argument 1 missing for builtin", ps.Res.(*env.Error), nil)
			return ps
		}

		arg0 = ps.Res
	}

	// Process second argument if needed
	if bi.Argsn > 1 && bi.Cur1 == nil {
		Rye0_EvalExpression2(ps, true)

		// Inline error checking for speed
		if ps.FailureFlag {
			if !bi.AcceptFailure {
				updateErrorCodeContext(ps)
				ps.ErrorFlag = true
				return ps
			}
		}

		if ps.ErrorFlag || ps.ReturnFlag {
			ps.Res = env.NewError4(0, "Argument 2 missing for builtin", ps.Res.(*env.Error), nil)
			return ps
		}

		arg1 = ps.Res
	}

	// Process third argument if needed
	if bi.Argsn > 2 && bi.Cur2 == nil {
		Rye0_EvalExpression2(ps, true)

		// Inline error checking for speed
		if ps.FailureFlag {
			if !bi.AcceptFailure {
				updateErrorCodeContext(ps)
				ps.ErrorFlag = true
				return ps
			}
		}

		if ps.ErrorFlag || ps.ReturnFlag {
			ps.Res = env.NewError4(0, "Argument 3 missing for builtin", ps.Res.(*env.Error), nil)
			return ps
		}

		arg2 = ps.Res
	}

	// Process remaining arguments with minimal error checking
	if bi.Argsn > 3 && bi.Cur3 == nil {
		Rye0_EvalExpression2(ps, true)
		arg3 = ps.Res
	}

	if bi.Argsn > 4 && bi.Cur4 == nil {
		Rye0_EvalExpression2(ps, true)
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

// updateErrorCodeContext updates the error's code context if it's not already set
func updateErrorCodeContext(ps *env.ProgramState) {
	if err, ok := ps.Res.(*env.Error); ok {
		if err.CodeBlock.Len() == 0 {
			err.CodeBlock = ps.Ser
			err.CodeContext = ps.Ctx
		}
	} else if err, ok := ps.Res.(env.Error); ok {
		if err.CodeBlock.Len() == 0 {
			err.CodeBlock = ps.Ser
			err.CodeContext = ps.Ctx
		}
	}
}

// Rye0_checkFlagsBi checks if there are failure flags and handles them appropriately.
func Rye0_checkFlagsBi(bi env.Builtin, ps *env.ProgramState, n int) bool {
	trace("CHECK FLAGS BI")

	if ps.FailureFlag {
		trace("------ > FailureFlag")
		if bi.AcceptFailure {
			trace2("----- > Accept Failure")
		} else {
			trace2("Fail ------->  Error.")
			updateErrorCodeContext(ps)
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
func Rye0_checkFlagsAfterBlock(ps *env.ProgramState, n int) bool {
	trace2("CHECK FLAGS AFTER BLOCKS")
	trace2(n)

	if ps.FailureFlag && !ps.ReturnFlag {
		trace2("FailureFlag")
		trace2("Fail->Error.")

		if !ps.InErrHandler && Rye0_checkContextErrorHandler(ps) {
			return false // Error should be picked up in the handler block
		}

		updateErrorCodeContext(ps)
		trace2("FAIL -> ERROR blk")
		ps.ErrorFlag = true
		return true
	} else {
		trace2("NOT FailureFlag")
	}
	return false
}

// Rye0_checkErrorReturnFlag checks if there are error or return flags.
func Rye0_checkErrorReturnFlag(ps *env.ProgramState) bool {
	if ps.ErrorFlag {
		updateErrorCodeContext(ps)
		return true
	}
	return ps.ReturnFlag
}
