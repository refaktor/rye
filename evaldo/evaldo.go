package evaldo

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/refaktor/rye/env"
)

// init initializes the evaldo package by setting up the observer executor callback.
// Called: Automatically by Go runtime during package initialization
// Purpose: Registers EvalBlockInj as the callback for executing observer blocks, avoiding circular import issues
func init() {
	env.ObserverExecutor = EvalBlockInj
}

// NoInspectMode controls whether to exit immediately on error without showing debugging options
var NoInspectMode bool

// Flag to control whether to use the fast evaluator
var useFastEvaluator = false

// EnableFastEvaluator enables the fast evaluator for Rye0 dialect.
// Called from: Command-line flags or configuration code
// Purpose: Switches to optimized evaluation mode for Rye0 dialect code
func EnableFastEvaluator() {
	useFastEvaluator = true
}

// DisableFastEvaluator disables the fast evaluator for Rye0 dialect.
// Called from: Command-line flags or configuration code
// Purpose: Switches back to standard evaluation mode for Rye0 dialect code
func DisableFastEvaluator() {
	useFastEvaluator = false
}

// EvalBlock is the main entry point for evaluating a block of code.
// Called from: Throughout the codebase - main evaluation loops, builtins, function calls
// Purpose: Dispatches to the appropriate dialect-specific evaluator (Rye2, Eyr, Rye0, Rye00)
// HOTCODE: Performance-critical function called frequently during execution
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

// EvalBlockInjMultiDialect evaluates a block with value injection support across multiple dialects.
// Called from: EvalExpression_DispatchType (for OPGROUP/OPBLOCK modes), observer execution
// Purpose: Provides multi-dialect block evaluation with the ability to inject values into the first expression
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

// MaybeAcceptComma handles comma expression guards between block-level expressions.
// Called from: EvalBlockInj, EvalExpression_DispatchType (OPBBLOCK mode)
// Purpose: Checks for and consumes comma separators; re-enables injection if comma found
// HOTPATH: Called frequently during block evaluation
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

// EvalBlockInj evaluates a block of code with optional value injection for Rye2 dialect.
// Called from: EvalBlock, CallFunction_CollectArgs, ExecuteDeferredBlocks, observer execution, init()
// Purpose: Core Rye2 evaluator - loops through expressions, handling injection, commas, and error/failure flags
func EvalBlockInj(ps *env.ProgramState, inj env.Object, injnow bool) {
	//fmt.Println("--------------------BLOCK------------------->")
	// fmt.Println(ps.Ser)
	// fmt.Println(ps.BlockFile)
	// fmt.Println(ps.BlockLine)
	// fmt.Println("---------------------------------------------")
	// // repeats until at the end of the block
	for ps.Ser.Pos() < ps.Ser.Len() {
		injnow = EvalExpressionInj(ps, inj, injnow)
		// Check for both failure and error flags immediately after expression evaluation
		if ps.ErrorFlag || ps.ReturnFlag {
			return
		}
		if tryHandleFailure(ps) {
			return
		}
		injnow = MaybeAcceptComma(ps, inj, injnow)
	}
}

// EvalBlockInCtxInj evaluates a block in a specific context with value injection.
// Called from: Builtins that need to evaluate code in a different context
// Purpose: Temporarily switches to specified context, evaluates block, then restores original context
func EvalBlockInCtxInj(ps *env.ProgramState, ctx *env.RyeCtx, inj env.Object, injnow bool) {
	ctx2 := ps.Ctx
	ps.Ctx = ctx
	EvalBlockInj(ps, inj, injnow)
	// Note: We restore context even on error to maintain consistent state
	// The error flag will be checked by the caller
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

// EvalExpression is the consolidated expression evaluator handling both regular and injected evaluation.
// Called from: EvalExpressionInj, EvalExpression_CollectArg, EvalExpressionInjLimited (wrapper functions)
// Purpose: Evaluates left side via DispatchType, then optionally evaluates right side (opwords/pipewords)
func EvalExpression(ps *env.ProgramState, inj env.Object, injnow bool, limited bool) bool {
	if inj == nil || !injnow {
		EvalExpression_DispatchType(ps)
		if ps.ReturnFlag || ps.ErrorFlag {
			return injnow
		}
	} else {
		ps.Res = inj
		injnow = false
		if ps.ReturnFlag {
			return injnow
		}
	}
	OptionallyEvalExpressionRight(ps.Ser.Peek(), ps, limited)
	return injnow
}

// EvalExpression_CollectArg evaluates an expression without injection (for collecting function arguments).
// Called from: CallFunction_CollectArgs, CallBuiltin_CollectArgs, CallVarBuiltin, EvalWord
// Purpose: Wrapper for EvalExpression used when collecting arguments from code
func EvalExpression_CollectArg(ps *env.ProgramState, limited bool) {
	EvalExpression(ps, nil, false, limited)
}

// EvalExpressionInj evaluates an expression with optional value injection.
// Called from: EvalBlockInj, EvalExpression_DispatchType (OPBBLOCK mode), EvalSetword, EvalModword
// Purpose: Wrapper for EvalExpression that allows injecting a value into the expression
func EvalExpressionInj(ps *env.ProgramState, inj env.Object, injnow bool) bool {
	return EvalExpression(ps, inj, injnow, false)
}

// EvalExpressionInjLimited evaluates an expression with injection in limited mode (stops at setwords/pipewords).
// Called from: Various places where expression evaluation should not consume setwords or pipewords
// Purpose: Limited expression evaluation that prevents consuming certain right-side constructs
func EvalExpressionInjLimited(ps *env.ProgramState, inj env.Object, injnow bool) bool {
	return EvalExpression(ps, inj, injnow, true)
}

// OptionallyEvalExpressionRight evaluates right-side constructs like opwords, pipewords, and setwords.
// Called from: EvalExpression, recursively from itself
// Purpose: Handles operator precedence by evaluating opwords/pipewords/setwords/modwords to the right
func OptionallyEvalExpressionRight(nextObj env.Object, ps *env.ProgramState, limited bool) {
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
		OptionallyEvalExpressionRight(ps.Ser.Peek(), ps, limited)
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
		OptionallyEvalExpressionRight(ps.Ser.Peek(), ps, limited)
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
		OptionallyEvalExpressionRight(ps.Ser.Peek(), ps, limited)
		return
	case env.LModword:
		if limited {
			return
		}
		idx := opword.Index

		// Get old value for observer notification
		oldValue, exists := ps.Ctx.GetCurrent(idx)

		ok := ps.Ctx.Mod(idx, ps.Res)
		if !ok {
			ps.Res = env.NewError("Cannot modify constant " + ps.Idx.GetWord(idx) + ", use 'var' to declare it as a variable")
			ps.FailureFlag = true
			ps.ErrorFlag = true
			return
		} else {
			// Trigger observers if the variable was successfully modified
			if exists && ps.Ctx.IsVariable(idx) {
				// Only trigger if the value actually changed
				if oldValue == nil || !oldValue.Equal(ps.Res) {
					TriggerObservers(ps, ps.Ctx, idx, oldValue, ps.Res)
				}
			}
		}
		ps.Ser.Next()
		OptionallyEvalExpressionRight(ps.Ser.Peek(), ps, limited)
		return
	case env.CPath:
		if opword.Mode == 1 {
			ps.Ser.Next()
			EvalWord(ps, opword, ps.Res, false, false)
			// when calling cpath
			OptionallyEvalExpressionRight(ps.Ser.Peek(), ps, limited)
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
			OptionallyEvalExpressionRight(ps.Ser.Peek(), ps, limited)
			return
		}
	}
	return
}

// EvalExpression_DispatchType is the core type dispatcher that evaluates individual Rye values.
// Called from: EvalExpression, EvalWord, EvalGenword
// Purpose: Main type switch - handles all Rye value types and dispatches to appropriate handlers
// Note: This is the heart of the evaluator - if Rye were fully Polish notation, this would be most of it
func EvalExpression_DispatchType(ps *env.ProgramState) {
	object := ps.Ser.Pop()
	if object == nil {
		ps.ErrorFlag = true
		ps.Res = env.NewError("Expected Rye value but reached the end of the block. Check for missing values or incomplete expressions.")
		return
	}

	objType := object.Type()

	// Skip LocationNodes - they're only used for error reporting
	if objType == env.LocationNodeType {
		// Skip this node and try the next one
		if !ps.Ser.AtLast() {
			EvalExpression_DispatchType(ps)
		}
		return
	}

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
				EvalExpression_CollectArg(ps, false)
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
			if ps.ErrorFlag || ps.FailureFlag {
				return
			}
			ps.Ser = ser
			// return ps.Res
		} else if block.Mode == 3 {
			// OPBBLOCK - behaves like vals\with with a block argument
			// For now, inject nil - this might need to be the previous result or context value
			ser := ps.Ser
			ps.Ser = block.Series
			res := make([]env.Object, 0)
			injnow := true
			injVal := ps.Res // Use current result as injection value
			for ps.Ser.Pos() < ps.Ser.Len() {
				injnow = EvalExpressionInj(ps, injVal, injnow)
				if ps.ReturnFlag || ps.ErrorFlag {
					return
				}
				res = append(res, ps.Res)
				injnow = MaybeAcceptComma(ps, injVal, injnow)
			}
			ps.Ser = ser
			ps.Res = *env.NewBlock(*env.NewTSeries(res))
		} else if block.Mode == 4 {
			// OPGROUP - behaves like with function
			ser := ps.Ser
			ps.Ser = block.Series
			injVal := ps.Res // Use current result as injection value
			EvalBlockInjMultiDialect(ps, injVal, true)
			if ps.ErrorFlag || ps.FailureFlag {
				return
			}
			ps.Ser = ser
			// return ps.Res
		} else if block.Mode == 5 {
			// OPBLOCK - behaves like fn1 function call
			// Create a function with one anonymous argument and call it with current result
			spec := []env.Object{*env.NewWord(1)}
			ps.Res = *env.NewFunction(*env.NewBlock(*env.NewTSeries(spec)), block, false)
			// injVal := ps.Res // Use current result as argument
			// CallFunctionWithArgs(fn, ps, nil, injVal)
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
		CallBuiltin_CollectArgs(object.(env.Builtin), ps, nil, false, false, nil)
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
		ps.Res = env.NewError("Expression guard (comma) found inside an expression. Commas can only be used between block-level expressions, not within expressions.")
		return
	case env.ErrorType:
		ps.ErrorFlag = true
		ps.Res = env.NewError("Error object encountered in code block. This usually indicates a previous error that wasn't properly handled.")
		return
		ps.ErrorFlag = true
		ps.Res = env.NewError("Error object encountered in code block. This usually indicates a previous error that wasn't properly handled.")
		return
	default:
		ps.Res = object
		return
		// fmt.Println(object.Inspect(*ps.Idx))
		// ps.ErrorFlag = true
		// ps.Res = env.NewError("Unknown rye value in block")
		// return
	}
}

// findWordValue retrieves the value associated with a word or context path.
// Called from: EvalWord internally (now replaced by findWordValueWithFailureInfo in most places)
// Purpose: Looks up words in context hierarchy or traverses context paths to find values
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
				return found, *env.NewString("No word value!!"), currCtx
			}
		}
		return found, object, currCtx
	default:
		return false, nil, nil
	}
}

// findWordValueWithFailureInfo is an extended version of findWordValue that includes failure diagnostics.
// Called from: EvalWord
// Purpose: Like findWordValue but returns detailed info about which word failed in a context path
func findWordValueWithFailureInfo(ps *env.ProgramState, word1 env.Object) (bool, env.Object, *env.RyeCtx, string) {
	switch word := word1.(type) {
	case env.Word:
		object, found := ps.Ctx.Get(word.Index)
		if !found {
			return found, object, nil, ps.Idx.GetWord(word.Index)
		}
		return found, object, nil, ""
	case env.Opword:
		object, found := ps.Ctx.Get(word.Index)
		if !found {
			return found, object, nil, ps.Idx.GetWord(word.Index)
		}
		return found, object, nil, ""
	case env.Pipeword:
		object, found := ps.Ctx.Get(word.Index)
		if !found {
			return found, object, nil, ps.Idx.GetWord(word.Index)
		}
		return found, object, nil, ""
	case env.CPath:
		currCtx := ps.Ctx
		var contextPath strings.Builder
		i := 1
	gogo1:
		currWord := word.GetWordNumber(i)
		wordName := ps.Idx.GetWord(currWord.Index)
		if i == 1 {
			contextPath.WriteString(wordName)
		} else {
			contextPath.WriteString("/" + wordName)
		}

		object, found := currCtx.Get(currWord.Index)
		if !found {
			// Word not found - report which word and in which context
			if i == 1 {
				return false, object, currCtx, wordName
			} else {
				// Build context name from previous parts of the path
				var ctxName strings.Builder
				for j := 1; j < i; j++ {
					if j > 1 {
						ctxName.WriteString("/")
					}
					ctxName.WriteString(ps.Idx.GetWord(word.GetWordNumber(j).Index))
				}
				return false, object, currCtx, wordName + " (in context " + ctxName.String() + ")"
			}
		}
		if found && word.Cnt > i {
			switch swObj := object.(type) {
			case env.RyeCtx:
				currCtx = &swObj
				i += 1
				goto gogo1
			case env.Dict:
				return found, *env.NewString("Now word value 2!!"), currCtx, ""
			}
		}
		return found, object, currCtx, ""
	default:
		return false, nil, nil, "unknown word type"
	}
}

// EVALUATOR FUNCTIONS FOR SPECIFIC VALUE TYPES

// EvalWord evaluates a word by looking up its value and potentially dispatching to generic words.
// Called from: EvalExpression_DispatchType, OptionallyEvalExpressionRight, EvalObject
// Purpose: Main word evaluator - looks up in context, tries generic words if not found, handles getcpath mode
func EvalWord(ps *env.ProgramState, word env.Object, leftVal env.Object, toLeft bool, pipeSecond bool) {
	// Special handling for getcpath (mode 3) - behave like get-word
	if cpath, ok := word.(env.CPath); ok && cpath.Mode == 3 {
		found, object, _, failureInfo := findWordValueWithFailureInfo(ps, word)
		if found {
			// For getcpath, just return the value without calling it (like get-word behavior)
			ps.Res = object
			return
		} else {
			ps.ErrorFlag = true
			ps.Res = env.NewError2(5, "Word not found: "+failureInfo+". Check spelling or ensure the word is defined in the current context.")
			return
		}
	}

	// LOCAL FIRST
	var firstVal env.Object
	found, object, session, failureInfo := findWordValueWithFailureInfo(ps, word)
	pos := ps.Ser.GetPos()
	if !found { // look at Generic words, but first check type
		// fmt.Println(pipeSecond)
		kind := 0
		if leftVal != nil {
			kind = leftVal.GetKind()
		}
		if leftVal == nil && !pipeSecond {
			if !ps.Ser.AtLast() {
				EvalExpression_DispatchType(ps)
				if ps.ReturnFlag || ps.ErrorFlag {
					return
				}
				leftVal = ps.Res
				kind = leftVal.GetKind()
			}
		}
		if pipeSecond {
			if !ps.Ser.AtLast() {
				EvalExpression_DispatchType(ps)
				if ps.ReturnFlag || ps.ErrorFlag {
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
			ps.Res = env.NewError2(5, "Word not found: "+failureInfo+". Check spelling or ensure the word is defined in the current context.")
		}
		return
	}
}

// EvalGenword evaluates a generic word (explicitly declared generic).
// Called from: EvalExpression_DispatchType
// Purpose: Handles words explicitly marked as generic - evaluates next expression and dispatches on its type
func EvalGenword(ps *env.ProgramState, word env.Genword, leftVal env.Object, toLeft bool) {
	EvalExpression_DispatchType(ps)

	if ps.ReturnFlag || ps.ErrorFlag {
		return
	}

	var arg0 = ps.Res
	object, found := ps.Gen.Get(arg0.GetKind(), word.Index)
	if found {
		EvalObject(ps, object, arg0, toLeft, nil, false, nil) //ww0128a *
		return
	} else {
		ps.ErrorFlag = true
		ps.Res = env.NewError("Generic word not found: " + word.Print(*ps.Idx) + ". No implementation found for the given argument type.")
		return
	}
}

// EvalGetword evaluates a get-word (prefixed with ?) which retrieves a value without calling it.
// Called from: EvalExpression_DispatchType
// Purpose: Returns the raw value associated with a word without evaluating it (like quoting)
func EvalGetword(ps *env.ProgramState, word env.Getword, leftVal env.Object, toLeft bool) {
	object, found := ps.Ctx.Get(word.Index)
	if found {
		ps.Res = object
		return
	} else {
		ps.ErrorFlag = true
		ps.Res = env.NewError("Word not found: " + word.Print(*ps.Idx) + ". Get-word (?) requires the word to be defined in the current context.")
		return
	}
}

// EvalObject evaluates a Rye object, particularly handling callable types (builtins, functions).
// Called from: EvalWord, EvalGenword
// Purpose: Evaluates found objects - dispatches builtins/functions/cpaths to their callers, returns other types
func EvalObject(ps *env.ProgramState, object env.Object, leftVal env.Object, toLeft bool, ctx *env.RyeCtx, pipeSecond bool, firstVal env.Object) {
	switch object.Type() {
	case env.BuiltinType:
		bu := object.(env.Builtin)
		if checkForFailureWithBuiltin(bu, ps, 333) {
			return
		}
		CallBuiltin_CollectArgs(bu, ps, leftVal, toLeft, pipeSecond, firstVal)
		return
	case env.FunctionType:
		fn := object.(env.Function)
		CallFunction_CollectArgs(fn, ps, leftVal, toLeft, ctx, pipeSecond, firstVal)
		return
	case env.CPathType: // RMME
		// Check if this is a getcpath (mode 3) - behave like get-word
		if cpath, ok := object.(env.CPath); ok && cpath.Mode == 3 {
			// For getcpath, just return the object without calling it (like get-word behavior)
			ps.Res = object
			return
		}
		// For other CPath modes (opcpath, pipecpath), treat as function
		fn := object.(env.Function)
		CallFunction_CollectArgs(fn, ps, leftVal, toLeft, ctx)
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

// EvalSetword evaluates a set-word (word:) which sets a new word in the context.
// Called from: EvalExpression_DispatchType
// Purpose: Evaluates the expression to the right and binds the result to a word (creating new binding)
func EvalSetword(ps *env.ProgramState, word env.Setword) {
	// es1 := EvalExpression(es)
	EvalExpressionInj(ps, nil, false)
	if ps.ErrorFlag || ps.FailureFlag {
		return
	}
	idx := word.Index
	if ps.AllowMod {
		ok := ps.Ctx.Mod(idx, ps.Res)
		if !ok {
			ps.Res = env.NewError("Cannot modify constant '" + ps.Idx.GetWord(idx) + "'. Use 'var' to declare it as a variable, or use modword (::) if it's already a variable.")
			ps.FailureFlag = true
			ps.ErrorFlag = true
		}
	} else {
		ok := ps.Ctx.SetNew(idx, ps.Res, ps.Idx)
		if !ok {
			ps.Res = env.NewError("Cannot set word '" + ps.Idx.GetWord(idx) + "' because it's already set. Use modword (::) to modify an existing word, or use a different name.")
			ps.FailureFlag = true
			ps.ErrorFlag = true
		}
	}
}

// EvalModword evaluates a mod-word (word::) which modifies an existing word in the context.
// Called from: EvalExpression_DispatchType
// Purpose: Evaluates the expression to the right and modifies an existing word's value, triggers observers
func EvalModword(ps *env.ProgramState, word env.Modword) {
	// es1 := EvalExpression(es)
	EvalExpressionInj(ps, nil, false)
	if ps.ErrorFlag || ps.FailureFlag {
		return
	}
	idx := word.Index

	// Get old value for observer notification
	oldValue, exists := ps.Ctx.GetCurrent(idx)

	ok := ps.Ctx.Mod(idx, ps.Res)
	if !ok {
		ps.Res = env.NewError("Cannot modify constant '" + ps.Idx.GetWord(idx) + "'. Use 'var' to declare it as a variable before modifying it.")
		ps.FailureFlag = true
		ps.ErrorFlag = true
	} else {
		// Trigger observers if the variable was successfully modified
		if exists && ps.Ctx.IsVariable(idx) {
			// Only trigger if the value actually changed
			if oldValue == nil || !oldValue.Equal(ps.Res) {
				TriggerObservers(ps, ps.Ctx, idx, oldValue, ps.Res)
			}
		}
	}
}

//
// CALLING FUNCTIONS
//

// CallFunctionWithArgs is a consolidated function caller that dispatches based on argument count.
// Called from: Various builtins that need to call user functions
// Purpose: Dispatcher that routes to specialized function callers based on number of arguments (0,1,2,4,N)
func CallFunctionWithArgs(fn env.Function, ps *env.ProgramState, ctx *env.RyeCtx, args ...env.Object) {
	ctx = DetermineContext(fn, ps, ctx)
	if ctx == nil {
		return
	}

	switch len(args) {
	case 0:
		CallFunction_CollectArgs(fn, ps, nil, false, ctx)
		return
	case 1:
		CallFunction_CollectArgs(fn, ps, args[0], false, ctx)
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

// CallFunction_CollectArgs calls a function by collecting arguments from the code stream.
// Called from: EvalObject, CallFunctionWithArgs (0 or 1 arg case)
// Purpose: Main function caller in evaluator - collects args from code, sets up context, executes function body
func CallFunction_CollectArgs(fn env.Function, ps *env.ProgramState, arg0_ env.Object, toLeft bool, ctx *env.RyeCtx, pipeSecond ...interface{}) {
	// fmt.Println(1)

	// Handle optional pipeSecond and firstVal parameters
	var pipeSecondFlag bool
	var firstVal env.Object
	if len(pipeSecond) >= 1 {
		if ps, ok := pipeSecond[0].(bool); ok {
			pipeSecondFlag = ps
		}
	}
	if len(pipeSecond) >= 2 {
		if fv, ok := pipeSecond[1].(env.Object); ok {
			firstVal = fv
		}
	}

	// Determine arg0 based on pipeSecond flag (same logic as CallBuiltin_CollectArgs)
	var arg0 env.Object
	if arg0_ != nil && !pipeSecondFlag {
		arg0 = arg0_
	} else if firstVal != nil && pipeSecondFlag {
		arg0 = firstVal
	} else if pipeSecondFlag && fn.Argsn > 0 {
		// When pipeSecond is true but firstVal is nil (non-generic word),
		// evaluate the next expression to get arg0
		EvalExpression_CollectArg(ps, true)
		if ps.ReturnFlag || ps.ErrorFlag {
			return
		}
		arg0 = ps.Res
	}

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
	evalExprFn := EvalExpression_CollectArg
	if arg0 != nil {
		if fn.Spec.Series.Len() > 0 {
			index := fn.Spec.Series.Get(ii).(env.Word).Index
			fnCtx.Set(index, arg0)
			ps.Args[ii] = index
			ii = 1
			if !toLeft {
				//evalExprFn = EvalExpression_ // 2020-01-12 .. changed to ion2
				evalExprFn = EvalExpression_CollectArg
			}
		}
	}

	// Handle arg1 when pipeSecond is true (same logic as CallBuiltin_CollectArgs)
	// When pipeSecond is true and arg0_ is provided, arg0_ should become arg1 (the second argument)
	if arg0_ != nil && pipeSecondFlag && fn.Argsn > 1 && ii == 1 {
		if fn.Spec.Series.Len() > 1 {
			index := fn.Spec.Series.Get(1).(env.Word).Index
			fnCtx.Set(index, arg0_)
			ps.Args[1] = index
			ii = 2 // Skip collecting the second argument from code stream
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
		// The createcurriedcaller is now created explicitly with partial builtin function
		index := fn.Spec.Series.Get(i).(env.Word).Index
		fnCtx.Set(index, ps.Res)
		if i == 0 {
			arg0 = ps.Res
		}
		ps.Args[i] = index
	}
	ser0 := ps.Ser // only after we process the arguments and get new position
	ps.Ser = fn.Body.Series
	blockFile := ps.BlockFile
	blockLine := ps.BlockLine

	ps.BlockFile = fn.Body.FileName
	ps.BlockLine = fn.Body.Line

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
	MaybeDisplayFailureOrError(ps, ps.Idx, "Call func collect args")
	if ps.ErrorFlag || ps.FailureFlag {
		ps.Ctx = env0
		ps.Ser = ser0
		// Don't restore state on error - let error handler deal with it
		return
	}

	//	}
	// MaybeDisplayFailureOrError(result, result.Idx, "call function")
	if ps.ForcedResult != nil {
		ps.Res = ps.ForcedResult
		ps.ForcedResult = nil
	}
	ps.Ctx = env0
	ps.Ser = ser0
	ps.BlockFile = blockFile
	ps.BlockLine = blockLine
	ps.ReturnFlag = false

	// Don't return closure contexts to the pool to prevent context reuse issues
	if !fnCtx.IsClosure {
		// Observers are now automatically cleaned up with the context
		envPool.Put(fnCtx)
	}

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

// CallFunctionArgs2 calls a function with exactly 2 arguments provided.
// Called from: CallFunctionWithArgs, builtins needing to call 2-arg functions
// Purpose: Optimized path for 2-argument function calls from builtins
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
			ExecuteDeferredBlocks(psX)
		}
	}()

	psX.Ser.SetPos(0)
	EvalBlockInj(psX, arg0, true)
	if psX.ErrorFlag || psX.FailureFlag {
		ps.Res = psX.Res
		ps.ErrorFlag = psX.ErrorFlag
		ps.FailureFlag = psX.FailureFlag
		return
	}
	// fmt.Println(psX.Res)
	MaybeDisplayFailureOrError(psX, psX.Idx, "call func args 2")
	if psX.ForcedResult != nil {
		ps.Res = psX.ForcedResult
		psX.ForcedResult = nil
	} else {
		ps.Res = psX.Res
	}
	ps.ReturnFlag = false
}

// CallFunctionArgs4 calls a function with exactly 4 arguments provided.
// Called from: CallFunctionWithArgs, builtins needing to call 4-arg functions
// Purpose: Optimized path for 4-argument function calls from builtins
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
	psX.Ser.SetPos(0)
	defer func() {
		if len(psX.DeferBlocks) > 0 {
			ExecuteDeferredBlocks(psX)
		}
	}()

	EvalBlockInj(psX, arg0, true)
	if psX.ErrorFlag || psX.FailureFlag {
		ps.Res = psX.Res
		ps.ErrorFlag = psX.ErrorFlag
		ps.FailureFlag = psX.FailureFlag
		return
	}
	MaybeDisplayFailureOrError(psX, psX.Idx, "call func args 4")
	if psX.ForcedResult != nil {
		ps.Res = psX.ForcedResult
		psX.ForcedResult = nil
	} else {
		ps.Res = psX.Res
	}
	ps.ReturnFlag = false
}

// CallFunctionArgsN calls a function with a variable number of arguments (N arguments).
// Called from: CallFunctionWithArgs, CallCurriedCaller, builtins with variable args
// Purpose: Generic function caller for any number of arguments provided as a slice
func CallFunctionArgsN(fn env.Function, ps *env.ProgramState, ctx *env.RyeCtx, args ...env.Object) {
	// fmt.Println(6)
	// ctx = nil
	var fnCtx = DetermineContext(fn, ps, ctx)
	if ps.ReturnFlag || ps.ErrorFlag {
		return
	}

	for i, argWord := range fn.Spec.Series.S {
		index := argWord.(env.Word).Index
		arg := args[i]
		fnCtx.Set(index, arg)
	}
	/* for i, arg := range args {
		index := fn.Spec.Series.Get(i).(env.Word).Index
		fnCtx.Set(index, arg)
	}*/

	// TRY
	psX := env.NewProgramState(fn.Body.Series, ps.Idx)
	psX.Ctx = fnCtx
	psX.PCtx = ps.PCtx
	psX.Gen = ps.Gen

	// END TRY
	psX.Ser.SetPos(0)
	defer func() {
		if len(psX.DeferBlocks) > 0 {
			ExecuteDeferredBlocks(psX)
		}
	}()

	if len(args) > 0 {
		EvalBlockInj(psX, args[0], true)
	} else {
		EvalBlock(psX)
	}
	if psX.ErrorFlag || psX.FailureFlag {
		ps.Res = psX.Res
		ps.ErrorFlag = psX.ErrorFlag
		ps.FailureFlag = psX.FailureFlag
		return
	}
	MaybeDisplayFailureOrError(psX, psX.Idx, "call func args N")
	if psX.ForcedResult != nil {
		ps.Res = psX.ForcedResult
		psX.ForcedResult = nil
	} else {
		ps.Res = psX.Res
	}
	ps.ReturnFlag = false
}

// DetermineContext determines the appropriate context for a function call.
// Called from: CallFunctionWithArgs, CallFunctionArgsN
// Purpose: Sets up function execution context based on pure/impure, defined context, and parent context
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

// CallCurriedCaller handles calling a curried caller (partially applied function or builtin).
// Called from: EvalExpression_DispatchType, EvalObject
// Purpose: Executes curried callers by filling in remaining arguments and calling the underlying builtin/function
func CallCurriedCaller(cc env.CurriedCaller, ps *env.ProgramState, arg0_ env.Object, toLeft bool, pipeSecond bool, firstVal env.Object) {
	// Initialize arguments with curried values if available
	var arg0 env.Object = cc.Cur0
	var arg1 env.Object = cc.Cur1
	var arg2 env.Object = cc.Cur2
	var arg3 env.Object = cc.Cur3
	var arg4 env.Object = cc.Cur4

	// Determine the number of arguments needed
	var argsn int
	if cc.CallerType == 0 { // Builtin
		argsn = cc.Builtin.Argsn
	} else { // Function
		argsn = cc.Function.Argsn
	}

	evalExprFn := EvalExpression_CollectArg

	// Handle arg0 - override with provided arg if available
	if arg0_ != nil && !pipeSecond {
		arg0 = arg0_
	} else if firstVal != nil && pipeSecond {
		arg0 = firstVal
	} else if arg0 == nil && argsn > 0 {
		// Only evaluate if we don't have a curried value
		evalExprFn(ps, true)
		if ps.ReturnFlag || ps.ErrorFlag {
			return
		}
		arg0 = ps.Res
	}

	// Handle arg1
	if arg0_ != nil && pipeSecond {
		arg1 = arg0_
	} else if arg1 == nil && argsn > 1 {
		// Only evaluate if we don't have a curried value
		evalExprFn(ps, true)
		if ps.ReturnFlag || ps.ErrorFlag {
			return
		}
		arg1 = ps.Res
	}

	// Handle remaining arguments - only evaluate if not curried
	if arg2 == nil && argsn > 2 {
		evalExprFn(ps, true)
		if ps.ReturnFlag || ps.ErrorFlag {
			return
		}
		arg2 = ps.Res
	}

	if arg3 == nil && argsn > 3 {
		evalExprFn(ps, true)
		if ps.ReturnFlag || ps.ErrorFlag {
			return
		}
		arg3 = ps.Res
	}

	if arg4 == nil && argsn > 4 {
		evalExprFn(ps, true)
		if ps.ReturnFlag || ps.ErrorFlag {
			return
		}
		arg4 = ps.Res
	}

	// Call the appropriate function based on caller type
	if cc.CallerType == 0 { // Builtin
		bi := *cc.Builtin
		ps.Res = bi.Fn(ps, arg0, arg1, arg2, arg3, arg4)
	} else { // Function
		fn := *cc.Function
		CallFunctionArgsN(fn, ps, nil, arg0, arg1, arg2, arg3, arg4)
	}
}

// CALLING BUILTINS

// CallBuiltin_CollectArgs calls a builtin by collecting up to 5 arguments from the code stream.
// Called from: EvalExpression_DispatchType, EvalObject
// Purpose: Main builtin caller - collects arguments, handles failure flags, calls builtin function
func CallBuiltin_CollectArgs(bi env.Builtin, ps *env.ProgramState, arg0_ env.Object, toLeft bool, pipeSecond bool, firstVal env.Object) {
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

	// Removed experiment with currying since Cur fields were removed from Builtin type
	// end of experiment

	evalExprFn := EvalExpression_CollectArg
	curry := false

	//fmt.Println("*** BUILTIN ***")

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
		if ps.ErrorFlag {
			// ps.Res = env.NewError4(0, "argument 1 of "+strconv.Itoa(bi.Argsn)+" missing of builtin: '"+bi.Doc+"'", ps.Res.(*env.Error), nil)
			return
		}
		if ps.ReturnFlag {
			return
		}
		// The CallCurriedCaller is now created explicitly with partial builtin function
		arg0 = ps.Res
	}

	if arg0_ != nil && pipeSecond {
		arg1 = arg0_
	} else if bi.Argsn > 1 {
		evalExprFn(ps, true) // <---- THESE DETERMINE IF IT CONSUMES WHOLE EXPRESSION OR NOT IN CASE OF PIPEWORDS .. HM*... MAYBE WOULD COULD HAVE A WORD MODIFIER?? a: 2 |add 5 a:: 2 |add 5 print* --TODO

		if checkForFailureWithBuiltin(bi, ps, 1) {
			return
		}
		if ps.ReturnFlag || ps.ErrorFlag {
			ps.Res = env.NewError4(0, "Argument 2 of "+strconv.Itoa(bi.Argsn)+" missing for builtin "+FormatBuiltinReference(bi.Doc)+". Check that all required arguments are provided.", ps.Res.(*env.Error), nil)
			return
		}
		//fmt.Println(ps.Res)
		// The CallCurriedCaller is now created explicitly with partial builtin function
		arg1 = ps.Res
	}
	if bi.Argsn > 2 {
		evalExprFn(ps, true)

		if checkForFailureWithBuiltin(bi, ps, 2) {
			return
		}
		if ps.ReturnFlag || ps.ErrorFlag {
			ps.Res = env.NewError4(0, "Argument 3 missing. Check that all required arguments are provided for the builtin function.", ps.Res.(*env.Error), nil)
			return
		}
		// The CallCurriedCaller is now created explicitly with partial builtin function
		arg2 = ps.Res
	}
	if bi.Argsn > 3 {
		evalExprFn(ps, true)
		// The CallCurriedCaller is now created explicitly with partial builtin function
		arg3 = ps.Res
	}
	if bi.Argsn > 4 {
		evalExprFn(ps, true)
		// The CallCurriedCaller is now created explicitly with partial builtin function
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
	if curry {
		// Currying is no longer supported since Cur fields were removed
		ps.Res = bi.Fn(ps, arg0, arg1, arg2, arg3, arg4)
	} else {
		ps.Res = bi.Fn(ps, arg0, arg1, arg2, arg3, arg4)
	}
	if ps.Res == nil {
		ps.Res = env.NewError4(0, "Builtin returned a invalid value (nil)", nil, nil)
		ps.ErrorFlag = true
	}
}

// CallVarBuiltin calls a variadic builtin by collecting all required arguments into a slice.
// Called from: EvalExpression_DispatchType, EvalObject
// Purpose: Handles builtins with variable number of arguments, collecting them into a slice
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
			EvalExpression_CollectArg(ps, true)
			if ps.ReturnFlag || ps.ErrorFlag {
				return
			}

			args[ii] = ps.Res
			ii++
		}

		if arg0_ != nil && pipeSecond {
			args[ii] = arg0_
			ii++
		} else if bi.Argsn > 1 {
			EvalExpression_CollectArg(ps, true)
			if ps.ReturnFlag || ps.ErrorFlag {
				return
			}

			args[ii] = ps.Res
			ii++
		}
		//variadic version
		for i := 2; i < bi.Argsn; i += 1 {
			EvalExpression_CollectArg(ps, true)
			if ps.ReturnFlag || ps.ErrorFlag {
				return
			}

			args[ii] = ps.Res
			ii++

		}
	}

	ps.Res = bi.Fn(ps, args...)
	if ps.Res == nil {
		ps.Res = env.NewError4(0, "Builtin returned a invalid value (nil)", nil, nil)
		ps.ErrorFlag = true
	}

}

// DirectlyCallBuiltin directly calls a builtin with provided arguments (no collection from code).
// Called from: Builtins that need to call other builtins directly
// Purpose: Direct builtin invocation helper for inter-builtin calls
func DirectlyCallBuiltin(ps *env.ProgramState, bi env.Builtin, a0 env.Object, a1 env.Object) env.Object {
	// Since Cur fields were removed from Builtin type, we just use the provided arguments
	var arg2 env.Object
	var arg3 env.Object
	var arg4 env.Object
	res := bi.Fn(ps, a0, a1, arg2, arg3, arg4)
	if res == nil {
		ps.ErrorFlag = true
		return env.NewError4(0, "Builtin returned a invalid value (nil)", nil, nil)
	}
	return res
}

// DISPLAYING FAILURE OR ERRROR

// findNearestLocationNode searches for the nearest LocationNode for error reporting.
// Called from: Error display functions
// Purpose: Finds source location information by searching backward then forward from error position
func findNearestLocationNode(ps *env.ProgramState) *env.LocationNode {
	pos := ps.Ser.GetPos()

	// Search backwards from current position
	for i := pos - 1; i >= 0; i-- {
		obj := ps.Ser.Get(i)
		if obj != nil && obj.Type() == env.LocationNodeType {
			if locNode, ok := obj.(env.LocationNode); ok {
				return &locNode
			}
		}
	}

	// Search forward if nothing found backwards
	for i := pos; i < ps.Ser.Len(); i++ {
		obj := ps.Ser.Get(i)
		if obj != nil && obj.Type() == env.LocationNodeType {
			if locNode, ok := obj.(env.LocationNode); ok {
				return &locNode
			}
		}
	}

	return nil
}

// calculateErrorPosition calculates the character position of an error in a source line.
// Called from: Error display functions (currently unused but available)
// Purpose: Determines exact column position of error by counting characters from LocationNode
func calculateErrorPosition(ps *env.ProgramState, locationNode *env.LocationNode) int {
	// Get the current series position (similar to how PositionAndSurroundingElements works)
	errorPos := ps.Ser.Pos() - 1 // The error is typically at pos-1 like in the original (here) logic

	if errorPos < 0 {
		return 0
	}

	// Count characters from the beginning of the line to find the error position
	charCount := 0
	foundLocationNode := false

	// Go through the series and count characters from the LocationNode for this line
	for i := 0; i < ps.Ser.Len(); i++ {
		obj := ps.Ser.Get(i)
		if obj == nil {
			continue
		}

		// Check if this is a LocationNode for the current line
		if obj.Type() == env.LocationNodeType {
			if locNode, ok := obj.(env.LocationNode); ok && locNode.Line == locationNode.Line {
				foundLocationNode = true
				charCount = 0 // Reset count from this LocationNode
				continue
			} else if foundLocationNode {
				// We've moved to a different line, stop counting
				break
			}
			continue
		}

		// Only start counting after we've found the LocationNode for this line
		if !foundLocationNode {
			continue
		}

		// If this is the error position, return current character count
		if i == errorPos {
			return charCount
		}

		// Add the length of this object's string representation plus space
		objStr := obj.Print(*ps.Idx)
		charCount += len(objStr)

		// Add space separator (except for the last token before error)
		if i < errorPos {
			charCount += 1
		}
	}

	return charCount
}

// DisplayEnhancedError displays a formatted error with source location and block context.
// Called from: MaybeDisplayFailureOrError2
// Purpose: Main error display - shows red banner, error message, block location, and <here> marker
func DisplayEnhancedError(es *env.ProgramState, genv *env.Idxs, tag string, topLevel bool) {
	// Red background banner for runtime errors
	if !es.SkipFlag {
		fmt.Print("\x1b[41m\x1b[30m RUNTIME ERROR \x1b[0m\n") // Red background, black text

		// Bold red for error message
		fmt.Print("\x1b[1;31m") // Bold red
		fmt.Println(es.Res.Print(*genv))
		fmt.Print("\x1b[0m") // Reset
	}
	// Get location information from the current block
	displayBlockWithErrorPosition(es, genv)
	if topLevel {
		es.SkipFlag = false
	}
}

// displayBlockWithErrorPosition shows the current block with a <here> marker at the error position.
// Called from: DisplayEnhancedError
// Purpose: Displays block content with visual indicator showing exactly where the error occurred
func displayBlockWithErrorPosition(es *env.ProgramState, genv *env.Idxs) {

	// Bold cyan for location information
	fmt.Print("\x1b[1;36mCode block starting at	\x1b[1;34m") // Bold cyan "At", bold blue for location
	if es.BlockFile != "" {
		fmt.Printf("%s:%d", es.BlockFile, es.BlockLine)
	} else {
		fmt.Printf("line %d", es.BlockLine)
	}
	fmt.Print("\x1b[0m\n") // Reset

	// Show the current block content with <here> marker
	fmt.Print("\x1b[37mBlock:\x1b[0m\n")
	fmt.Print("\x1b[1;37m  ")

	// Get current position in the block
	errorPos := es.Ser.Pos() - 1
	if errorPos < 0 {
		errorPos = 0
	}

	// Build the block representation with <here> marker
	blockStr := buildBlockStringWithMarker(es.Ser.S, errorPos, genv)
	fmt.Print(blockStr)
	fmt.Print("\x1b[0m\n") // Reset
}

// DELETE_ME_getCurrentBlock is deprecated and marked for deletion.
// Called from: Nowhere (deprecated)
// Purpose: Was used to get current executing block, now replaced by better location tracking
func DELETE_ME_getCurrentBlock(es *env.ProgramState) *env.Block {
	// First, check if we have block location information from the program state
	// This would be the case when we're executing inside a block with location data

	// Try to find the current executing block by looking for any block with location information in the series
	for i := 0; i < es.Ser.Len(); i++ {
		obj := es.Ser.Get(i)
		if obj != nil && obj.Type() == env.BlockType {
			if block, ok := obj.(env.Block); ok {
				// Return the first block that has location information
				if block.FileName != "" || block.Line > 0 {
					return &block
				}
			}
		}
	}

	// If no block found in series, create a synthetic block with available location info
	// This handles cases where we're in the root execution context
	if es.ScriptPath != "" {
		// Create a block representing the current execution context
		syntheticBlock := &env.Block{
			Series:   es.Ser,
			Mode:     0,
			FileName: es.ScriptPath,
			Line:     1, // Default to line 1 if we don't have more specific info
			Column:   1, // Default to column 1
		}
		return syntheticBlock
	}

	return nil
}

// buildBlockStringWithMarker creates a string representation of a block with <here> marker at error position.
// Called from: displayBlockWithErrorPosition
// Purpose: Builds block display string showing 8 nodes before/after error with <here> marker and ellipses
func buildBlockStringWithMarker(currSer []env.Object, errorPos int, genv *env.Idxs) string {
	var result strings.Builder
	result.WriteString("{ ")

	// Calculate the range to display: 8 nodes before and 8 nodes after
	startPos := errorPos - 8
	endPos := errorPos + 8

	// Adjust boundaries to stay within the block
	if startPos < 0 {
		startPos = 0
	}
	if endPos >= len(currSer) {
		endPos = len(currSer) - 1
	}

	// Show ellipsis if we're not starting from the beginning
	if startPos > 0 {
		result.WriteString("... ")
	}

	// Display the selected range
	for i := startPos; i <= endPos && i < len(currSer); i++ {
		if i == errorPos {
			result.WriteString("\x1b[1;31m<here>\x1b[0m ")
		}

		obj := currSer[i]
		if obj != nil {
			result.WriteString(obj.Dump(*genv))
			result.WriteString(" ")
		}
	}

	// If error position is at the end and within our display range
	if errorPos >= len(currSer) && errorPos <= endPos {
		result.WriteString("\x1b[1;31m<here>\x1b[0m ")
	}

	// Show ellipsis if we're not ending at the end of the block
	if endPos < len(currSer)-1 {
		result.WriteString("... ")
	}

	result.WriteString("}")
	return result.String()
}

// displayLocationFromSeries extracts and displays location information from LocationNodes in the series.
// Called from: Potentially error display code (legacy)
// Purpose: Displays source location by finding LocationNodes and showing source lines with position markers
func displayLocationFromSeries(es *env.ProgramState, genv *env.Idxs) {
	// First try to find LocationNodes in the series and display them nicely
	foundLocation := false

	// Look through the series for LocationNodes
	for i := 0; i < es.Ser.Len(); i++ {
		obj := es.Ser.Get(i)
		if obj != nil && obj.Type() == env.LocationNodeType {
			if locNode, ok := obj.(env.LocationNode); ok {
				if foundLocation {
					fmt.Print(" -> ")
				}
				fmt.Print(locNode.String())
				foundLocation = true
			}
		}
	}

	if foundLocation {
		fmt.Println()

		// Try to show the first LocationNode's source line for context
		for i := 0; i < es.Ser.Len(); i++ {
			obj := es.Ser.Get(i)
			if obj != nil && obj.Type() == env.LocationNodeType {
				if locNode, ok := obj.(env.LocationNode); ok && locNode.SourceLine != "" {
					fmt.Println("Source:")
					fmt.Printf("  %s\n", locNode.SourceLine)
					if locNode.Column > 0 {
						pointer := strings.Repeat(" ", locNode.Column+1) + "^"
						fmt.Printf("  %s\n", pointer)
					}
					break
				}
			}
		}
	} else {
		// Fallback to the original position reporting
		fmt.Print(es.Ser.PositionAndSurroundingElements(*genv))
	}
}

// MaybeDisplayFailureOrError displays errors/failures if error flag is set (wrapper).
// Called from: Throughout the codebase after evaluation operations
// Purpose: Wrapper that calls MaybeDisplayFailureOrError2 with default parameters
func MaybeDisplayFailureOrError(es *env.ProgramState, genv *env.Idxs, tag string) {
	MaybeDisplayFailureOrError2(es, genv, tag, false, false)
}

// MaybeDisplayFailureOrError2 displays errors/failures and optionally offers debugging options.
// Called from: MaybeDisplayFailureOrError, main REPL/file execution code
// Purpose: Main error display coordinator - shows enhanced errors and offers debugging in file mode
func MaybeDisplayFailureOrError2(es *env.ProgramState, genv *env.Idxs, tag string, topLevel bool, fileMode bool) {
	if !es.InErrHandler && (es.ErrorFlag || (es.FailureFlag && topLevel)) {
		// Use the enhanced error reporting with source location
		DisplayEnhancedError(es, genv, tag, topLevel)

		// Offer debugging options to the user
		if fileMode {
			OfferDebuggingOptions(es, genv, tag)
		}

		es.SkipFlag = false
	}
}

// MaybeDisplayFailureOrErrorWASM displays errors/failures in WASM environment using custom print function.
// Called from: WASM build code (main_wasm.go)
// Purpose: WASM-specific error display that uses provided print function instead of fmt.Println
func MaybeDisplayFailureOrErrorWASM(es *env.ProgramState, genv *env.Idxs, printfn func(string), tag string) {
	// if es.FailureFlag {
	// 	fmt.Print("\x1b[43m\x1b[33m FAILURE \x1b[0m\n") // Red background, black text
	//	printfn(tag)
	// }
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
// checkForFailureWithBuiltin checks if a failure should stop builtin execution.
// Called from: CallBuiltin_CollectArgs (between argument collection)
// Purpose: Converts failure to error if builtin doesn't accept failures
func checkForFailureWithBuiltin(bi env.Builtin, ps *env.ProgramState, n int) bool {
	if ps.FailureFlag && !bi.AcceptFailure {
		ps.ErrorFlag = true
		return true
	}
	return false
}

// checkForFailureWithVarBuiltin checks if a failure should stop variadic builtin execution.
// Called from: CallVarBuiltin, EvalObject
// Purpose: Converts failure to error if variadic builtin doesn't accept failures
func checkForFailureWithVarBuiltin(bi env.VarBuiltin, ps *env.ProgramState, n int) bool {
	if ps.FailureFlag && !bi.AcceptFailure {
		ps.ErrorFlag = true
		return true
	}
	return false
}

// trace is an empty trace function placeholder.
// Called from: Nowhere currently (can be used for debugging)
// Purpose: Placeholder for debug tracing - currently unused
func trace(s string) {

}

// tryHandleFailure attempts to handle a failure by calling context error handlers.
// Called from: EvalBlockInj
// Purpose: Checks for failure flag and tries to invoke error-handler word from context
func tryHandleFailure(ps *env.ProgramState) bool {
	if ps.FailureFlag && !ps.ReturnFlag && !ps.InErrHandler {
		if checkContextErrorHandler(ps) {
			return false // Successfully handled
		}
		// notify user that there was first state of failure, so it can be handeled
		fmt.Print("\x1b[43m\x1b[31m FAILURE \x1b[0m\n")
		ps.ErrorFlag = true
		return true // Unhandled failure
	}
	return false // No failure
}

// TriggerObservers triggers all observers watching a variable that has changed.
// Called from: EvalModword, OptionallyEvalExpressionRight (LModword case)
// Purpose: Notifies observers when a variable's value changes, executing their observer blocks
func TriggerObservers(ps *env.ProgramState, ctx *env.RyeCtx, wordIndex int, oldValue, newValue env.Object) {
	// Use the new context-level observer system
	env.TriggerObserversInChain(ps, ctx, wordIndex, oldValue, newValue)
}

// ExecuteDeferredBlocks executes all deferred blocks in LIFO order (last in, first out).
// Called from: CallFunction_CollectArgs, CallFunctionArgs2, CallFunctionArgs4, CallFunctionArgsN (via defer)
// Purpose: Executes cleanup blocks registered with defer, similar to Go's defer statement
func ExecuteDeferredBlocks(ps *env.ProgramState) {
	if len(ps.DeferBlocks) == 0 {
		return
	}

	// Save current state
	originalSer := ps.Ser
	originalFailureFlag := ps.FailureFlag
	originalErrorFlag := ps.ErrorFlag

	// Execute blocks in LIFO order (last in, first out)
	for i := len(ps.DeferBlocks) - 1; i >= 0; i-- {
		block := ps.DeferBlocks[i]

		// Reset failure/error flags for each deferred block
		ps.FailureFlag = false
		ps.ErrorFlag = false

		// Execute the deferred block
		ps.Ser = block.Series
		EvalBlock(ps)

		// If there was an error in a deferred block, we should still continue
		// executing other deferred blocks but preserve the error state
		if ps.ErrorFlag || ps.FailureFlag {
			fmt.Println("Error or failure in deferrer block")
			fmt.Println(ps.Res.Inspect(*ps.Idx))
			// Log or handle deferred block errors if needed
			// For now, continue with other deferred blocks
		}
	}

	// Clear the deferred blocks list
	ps.DeferBlocks = ps.DeferBlocks[:0]

	// Restore original state
	ps.Ser = originalSer
	ps.FailureFlag = originalFailureFlag
	ps.ErrorFlag = originalErrorFlag
}

// checkContextErrorHandler checks for and executes an error-handler word in the context.
// Called from: tryHandleFailure
// Purpose: Looks up and executes error-handler block from context to handle failures
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
		// If error handler itself had an error, log it but don't fail silently
		// The original error is still in ps.Res
		if ps.ErrorFlag {
			// Error in error handler - this is a serious issue
			// The original error remains, but we note that handler failed
			// Could potentially set a flag or create a compound error
		}
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
