// builtins_base_error_utils.go
package evaldo

import (
	"github.com/refaktor/rye/env"
)

// Additional error handling utilities to complement the existing ones in builtins_base_failure.go. Will get merged once tested and decided which ones should stay.

var Builtins_error_utils = map[string]*env.Builtin{
	// Tests:
	// equal { error? failure "test" } true
	// equal { error? 123 } false
	// Args:
	// * value: Any value to check
	// Returns:
	// * boolean true if the value is an error, false otherwise
	"is-error": {
		AcceptFailure: true,
		Argsn:         1,
		Doc:           "Returns true if the value is an error, false otherwise.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			ps.FailureFlag = false
			switch arg0.(type) {
			case env.Error, *env.Error:
				return *env.NewBoolean(true)
			}
			return *env.NewBoolean(false)
		},
	},

	// Tests:
	// equal { error-kind? failure { 'syntax-error 404 "Syntax error" } } 'syntax-error
	// equal { error-kind? 123 } _
	// Args:
	// * value: Any value to check
	// Returns:
	// * the kind of the error as a word, or void if not an error
	"error-kind?": {
		AcceptFailure: true,
		Argsn:         1,
		Doc:           "Returns the kind of an error, or void if not an error.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			ps.FailureFlag = false
			switch err := arg0.(type) {
			case env.Error:
				return *env.NewWord(err.Kind.Index)
			case *env.Error:
				return *env.NewWord(err.Kind.Index)
			}
			return *env.NewVoid()
		},
	},

	// Tests:
	// equal { error-of-kind? failure { 'syntax-error 404 "Syntax error" } 'syntax-error } true
	// equal { error-of-kind? failure { 'syntax-error 404 "Syntax error" } 'runtime-error } false
	// Args:
	// * error: Error object to check
	// * kind: Word or tagword representing the error kind to check against
	// Returns:
	// * boolean true if the error is of the specified kind, false otherwise
	"is-error-of-kind": {
		AcceptFailure: true,
		Argsn:         2,
		Doc:           "Returns true if the value is an error of the specified kind, false otherwise.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			ps.FailureFlag = false

			// Get the kind to check
			var kindIndex int
			switch kind := arg1.(type) {
			case env.Word:
				kindIndex = kind.Index
			case env.Tagword:
				kindIndex = kind.Index
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 2, []env.Type{env.WordType, env.TagwordType}, "error-of-kind?")
			}

			// Check if the error is of the specified kind
			switch err := arg0.(type) {
			case env.Error:
				return *env.NewBoolean(err.Kind.Index == kindIndex)
			case *env.Error:
				return *env.NewBoolean(err.Kind.Index == kindIndex)
			}

			return *env.NewBoolean(false)
		},
	},

	// Comment: Proposed but not sure it's really needed, might get removed
	// Tests:
	// equal { try-all { 1 + 2 } } [true 3]
	// equal { try-all { fail "error" } |first } false
	// Args:
	// * block: Block of code to execute
	// Returns:
	// * a block containing [success, result], where success is true if no error occurred
	"try-all": {
		Argsn: 1,
		Doc:   "Executes a block and returns a result tuple [success, result], where success is true if no error occurred.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch bloc := arg0.(type) {
			case env.Block:
				// Save current state
				oldSer := ps.Ser
				oldFailureFlag := ps.FailureFlag
				oldErrorFlag := ps.ErrorFlag

				// Execute the block
				ps.Ser = bloc.Series
				EvalBlock(ps)

				// Create result tuple
				result := ps.Res
				success := !ps.FailureFlag && !ps.ErrorFlag

				// Restore state
				ps.Ser = oldSer
				ps.FailureFlag = oldFailureFlag
				ps.ErrorFlag = oldErrorFlag

				// Return [success, result]
				return *env.NewBlock(*env.NewTSeries([]env.Object{
					*env.NewBoolean(success),
					result,
				}))
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.BlockType}, "try-all")
			}
		},
	},

	// Tests:
	// equal { finally { 1 + 2 } { print "cleanup" } } 3
	// equal { try { finally { fail "error" } { print "cleanup" } } |message? } "error"
	// Args:
	// * main-block: Block of code to execute
	// * finally-block: Block to execute afterward, regardless of errors
	// Returns:
	// * result of the main block, preserving any failure state
	"finally": {
		Argsn: 2,
		Doc:   "Executes a block and ensures another block is executed afterward, regardless of errors.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch mainBlock := arg0.(type) {
			case env.Block:
				switch finallyBlock := arg1.(type) {
				case env.Block:
					// Save current state
					oldSer := ps.Ser

					// Execute the main block
					ps.Ser = mainBlock.Series
					EvalBlock(ps)

					// Save result and flags
					result := ps.Res
					hasFailure := ps.FailureFlag
					hasError := ps.ErrorFlag

					// Execute the finally block
					ps.Ser = finallyBlock.Series
					ps.FailureFlag = false
					ps.ErrorFlag = false
					EvalBlock(ps)

					// Restore original result and flags
					ps.Res = result
					ps.FailureFlag = hasFailure
					ps.ErrorFlag = hasError

					// Restore series
					ps.Ser = oldSer

					return ps.Res
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 2, []env.Type{env.BlockType}, "finally")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.BlockType}, "finally")
			}
		},
	},
}

// RegisterErrorUtilsBuiltins registers the error utility builtins
func RegisterErrorUtilsBuiltins(ps *env.ProgramState) {
	// Register the error utility builtins
	for k, v := range Builtins_error_utils {
		RegisterBuiltins2(map[string]*env.Builtin{k: v}, ps, "error-utils")
	}
}
