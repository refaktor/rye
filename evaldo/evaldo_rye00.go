// evaldo_rye00.go - A highly simplified interpreter for the Rye language
// This version only handles builtins and integers for performance optimization experiments
package evaldo

import (
	"fmt"
	"strconv"

	"github.com/refaktor/rye/env"
)

// Rye00_EvalBlockInj evaluates a block with an optional injected value.
func Rye00_EvalBlockInj(ps *env.ProgramState, inj env.Object, injnow bool) *env.ProgramState {
	for ps.Ser.Pos() < ps.Ser.Len() {

		Rye00_EvalExpressionConcrete(ps)

		if Rye00_checkFlagsAfterExpression(ps) {
			return ps
		}
	}
	return ps
}

// Rye00_EvalExpressionConcrete evaluates a concrete expression.
// This is the main part of the evaluator that handles only integers and builtins.
func Rye00_EvalExpressionConcrete(ps *env.ProgramState) {
	object := ps.Ser.Pop()

	if object == nil {
		ps.ErrorFlag = true
		ps.Res = errMissingValue
	}

	switch object.Type() {
	case env.IntegerType:
		ps.Res = object
	case env.BlockType:
		ps.Res = object
	case env.WordType:
		Rye00_EvalWord(ps, object.(env.Word), nil, false, false)
	case env.BuiltinType:
		Rye00_CallBuiltin(object.(env.Builtin), ps, nil, false, false, nil)
	case env.ErrorType:
		setError00(ps, "Error object encountered")
	default:
		setError00(ps, "Unsupported type in simplified interpreter: "+strconv.Itoa(int(object.Type())))
	}
}

// Use pre-allocated error messages from evaldo_rye0.go
var (
	errUnsupportedType = env.NewError("Unsupported type in simplified interpreter")
)

// Helper function to set error state - uses the shared error variables
func setError00(ps *env.ProgramState, message string) {
	ps.ErrorFlag = true

	// Use pre-allocated errors for common messages
	switch message {
	case "Expected Rye value but it's missing":
		ps.Res = errMissingValue
	case "Expression guard inside expression":
		ps.Res = errExpressionGuard
	case "Error object encountered":
		ps.Res = errErrorObject
	case "Unsupported type in simplified interpreter":
		ps.Res = errUnsupportedType
	default:
		ps.Res = env.NewError(message)
	}
}

// Rye00_findWordValue returns the value associated with a word in the current context.
// Simplified version that only looks for builtins.
func Rye00_findWordValue(ps *env.ProgramState, word1 env.Object) (bool, env.Object, *env.RyeCtx) {
	// Extract the word index
	var index int
	if w, ok := word1.(env.Word); ok {
		index = w.Index
	} else {
		return false, nil, nil
	}

	// First try to get the value from the current context
	object, found := ps.Ctx.Get(index)
	if found {
		// Enable word replacement optimization for builtins
		if object.Type() == env.BuiltinType && ps.Ser.Pos() > 0 {
			ps.Ser.Put(object)
		}
		return found, object, nil
	}

	// If not found in the current context and there's no parent, return not found
	if ps.Ctx.Parent == nil {
		return false, nil, nil
	}

	// Try to get the value from parent contexts
	object, found, foundCtx := ps.Ctx.Get2(index)
	return found, object, foundCtx
}

// Rye00_EvalWord evaluates a word in the current context.
// Simplified version that only handles builtins.
func Rye00_EvalWord(ps *env.ProgramState, word env.Object, leftVal env.Object, toLeft bool, pipeSecond bool) {
	found, object, session := Rye00_findWordValue(ps, word)
	// pos := ps.Ser.GetPos()

	if found {
		Rye00_EvalObject(ps, object, leftVal, toLeft, session, pipeSecond, nil)
	} else {
		//ps.Ser.SetPos(pos)
		setError00(ps, "Word not found: "+word.Print(*ps.Idx))
	}
}

// Rye00_EvalObject evaluates a Rye object.
// Simplified version that only handles builtins.
func Rye00_EvalObject(ps *env.ProgramState, object env.Object, leftVal env.Object, toLeft bool, ctx *env.RyeCtx, pipeSecond bool, firstVal env.Object) {
	switch object.Type() {
	case env.BuiltinType:
		bu := object.(env.Builtin)

		if Rye00_checkForFailureWithBuiltin(bu, ps, 333) {
			return
		}
		Rye00_CallBuiltin(bu, ps, leftVal, toLeft, pipeSecond, firstVal)
		return
	default:
		ps.Res = object
	}
}

// Rye00_CallBuiltin calls a builtin function.
// Optimized version that focuses on performance.
func Rye00_CallBuiltin(bi env.Builtin, ps *env.ProgramState, arg0_ env.Object, toLeft bool, pipeSecond bool, firstVal env.Object) {
	// Fast path: If all arguments are already available (curried), call directly
	if (bi.Argsn == 0) ||
		(bi.Argsn == 1 && bi.Cur0 != nil) ||
		(bi.Argsn == 2 && bi.Cur0 != nil && bi.Cur1 != nil) {
		ps.Res = bi.Fn(ps, bi.Cur0, bi.Cur1, bi.Cur2, bi.Cur3, bi.Cur4)
		return
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
		Rye00_EvalExpressionConcrete(ps)

		// Inline error checking for speed
		if ps.FailureFlag {
			if !bi.AcceptFailure {
				updateErrorCodeContext(ps)
				ps.ErrorFlag = true
				return
			}
		}

		if ps.ErrorFlag || ps.ReturnFlag {
			ps.Res = errArg1Missing
			return
		}

		arg0 = ps.Res
	}

	// Process second argument if needed
	if bi.Argsn > 1 && bi.Cur1 == nil {
		Rye00_EvalExpressionConcrete(ps)

		// Inline error checking for speed
		if ps.FailureFlag {
			if !bi.AcceptFailure {
				updateErrorCodeContext(ps)
				ps.ErrorFlag = true
				return
			}
		}

		if ps.ErrorFlag || ps.ReturnFlag {
			ps.Res = errArg2Missing
			return
		}

		arg1 = ps.Res
	}

	// Process third argument if needed
	if bi.Argsn > 2 && bi.Cur2 == nil {
		Rye00_EvalExpressionConcrete(ps)

		// Inline error checking for speed
		if ps.FailureFlag {
			if !bi.AcceptFailure {
				updateErrorCodeContext(ps)
				ps.ErrorFlag = true
				return
			}
		}

		if ps.ErrorFlag || ps.ReturnFlag {
			ps.Res = errArg3Missing
			return
		}

		arg2 = ps.Res
	}

	// Process remaining arguments with minimal error checking
	if bi.Argsn > 3 && bi.Cur3 == nil {
		Rye00_EvalExpressionConcrete(ps)
		arg3 = ps.Res
	}

	if bi.Argsn > 4 && bi.Cur4 == nil {
		Rye00_EvalExpressionConcrete(ps)
		arg4 = ps.Res
	}

	// Call the builtin function
	ps.Res = bi.Fn(ps, arg0, arg1, arg2, arg3, arg4)
}

// Use the shared updateErrorCodeContext function from evaldo_rye0.go

// Rye00_checkForFailureWithBuiltin checks if there are failure flags and handles them appropriately.
func Rye00_checkForFailureWithBuiltin(bi env.Builtin, ps *env.ProgramState, n int) bool {
	if ps.FailureFlag {
		if bi.AcceptFailure {
			// Accept failure
		} else {
			updateErrorCodeContext(ps)
			ps.ErrorFlag = true
			return true
		}
	}
	return false
}

// Rye00_checkFlagsAfterExpression checks if there are failure flags after evaluating a block.
func Rye00_checkFlagsAfterExpression(ps *env.ProgramState) bool {
	if (ps.FailureFlag && !ps.ReturnFlag) || ps.ErrorFlag {
		updateErrorCodeContext(ps)
		ps.ErrorFlag = true
		return true
	}
	return false
}

// Rye00_MaybeDisplayFailureOrError displays failure or error information if present.
func Rye00_MaybeDisplayFailureOrError(es *env.ProgramState, genv *env.Idxs, tag string) {
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
