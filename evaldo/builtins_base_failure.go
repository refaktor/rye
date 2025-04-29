package evaldo

import (
	"fmt"
	"time"

	"github.com/refaktor/rye/env"
	"github.com/refaktor/rye/util"
)

var Builtins_failure = map[string]*env.Builtin{

	//
	// ##### Failure ###### "Handling failures"
	//

	// Tests:
	// equal { try { fail "error message" } |type? } 'error
	// equal { try { fail "error message" } |message? } "error message"
	// equal { try { fail 404 } |status? } 404
	// Args:
	// * error_info: String message, Integer code, or block for multiple parameters
	// Returns:
	// * error object and sets the failure flag
	"fail": { // **
		Argsn: 1,
		Doc:   "Creates an error and sets the failure flag, but continues execution (unlike ^fail).",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			ps.FailureFlag = true
			return *MakeRyeError(ps, arg0, nil)
		},
	},

	// Tests:
	// equal { fn { } { ^fail "error message" } |type? } 'error
	// equal { fn { } { ^fail "error message" } |message? } "error message"
	// equal { fn { } { ^fail 404 } |status? } 404
	// equal { fn { } { ^fail 'user-error } |kind? } 'user-error
	// Args:
	// * error_info: String message, Integer code, or block for multiple parameters
	// Returns:
	// * error object and sets both failure and return flags
	"^fail": {
		Argsn: 1,
		Doc:   "Creates an error and immediately returns from the current function with failure state.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			ps.FailureFlag = true
			ps.ReturnFlag = true
			return *MakeRyeError(ps, arg0, nil)
		},
	},

	"refail": {
		Argsn: 2,
		Doc:   "Re-raises an existing error with additional context.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch er := arg0.(type) {
			case *env.Error:
				ps.FailureFlag = true
				return *MakeRyeError(ps, arg1, er)
			default:
				return MakeArgError(ps, 1, []env.Type{env.ErrorType}, "reraise")
			}
		},
	},

	// Tests:
	// equal { failure "error message" |type? } 'error
	// equal { failure "error message" |message? } "error message"
	// equal { failure 404 |status? } 404
	// Args:
	// * error_info: String message, Integer code, or block for multiple parameters
	// Returns:
	// * error object without setting any flags
	"failure": { // **
		Argsn: 1,
		Doc:   "Creates an error object without setting any flags (unlike fail and ^fail).",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *MakeRyeError(ps, arg0, nil)
		},
	},

	// Tests:
	// equal { failure\wrap "outer error" failure "inner error" |message? } "outer error"
	// equal { failure\wrap "outer error" failure "inner error" |type? } 'error
	// Args:
	// * error_info: String message, Integer code, or block for multiple parameters
	// * error: Error object to wrap
	// Returns:
	// * new error object that wraps the provided error
	"failure\\wrap": {
		Argsn: 2,
		Doc:   "Creates a new error that wraps an existing error, allowing for error chaining.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch er := arg1.(type) {
			case *env.Error:
				return *MakeRyeError(ps, arg0, er)
			default:
				return MakeArgError(ps, 2, []env.Type{env.ErrorType}, "wrap\\failure")
			}
		},
	},

	// TODOC -- add documentation like other builtins have
	"cause?": {
		AcceptFailure: true,
		Argsn:         1,
		Doc:           "Extracts the root cause from an error chain by traversing the Parent references.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch er := arg0.(type) {
			case env.Error:
				// Find the deepest error in the chain
				current := &er
				for current.Parent != nil {
					current = current.Parent
				}
				return *current
			case *env.Error:
				// Find the deepest error in the chain
				current := er
				for current.Parent != nil {
					current = current.Parent
				}
				return *current
			default:
				ps.FailureFlag = true
				return env.NewError("arg 0 not error")
			}
		},
	},

	// Tests:
	// equal { failure 404 |status? } 404
	// equal { failure "message" |status? } 0
	// error { "not an error" |status? }
	// Args:
	// * error: Error object to extract status code from
	// Returns:
	// * integer status code of the error
	"status?": { // **
		AcceptFailure: true,
		Argsn:         1,
		Doc:           "Extracts the numeric status code from an error object.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch er := arg0.(type) {
			case env.Error:
				return *env.NewInteger(int64(er.Status))
			case *env.Error:
				return *env.NewInteger(int64(er.Status))
			default:
				ps.FailureFlag = true
				return env.NewError("arg 0 not error")
			}
		},
	},

	// Tests:
	// equal { failure "error message" |message? } "error message"
	// equal { failure 404 |message? } ""
	// error { "not an error" |message? }
	// Args:
	// * error: Error object to extract message from
	// Returns:
	// * string message of the error
	"message?": { // **
		AcceptFailure: true,
		Argsn:         1,
		Doc:           "Extracts the message string from an error object.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch er := arg0.(type) {
			case env.Error:
				return *env.NewString(er.Message)
			case *env.Error:
				return *env.NewString(er.Message)
			default:
				ps.FailureFlag = true
				return env.NewError("arg 0 not error")
			}
		},
	},

	// Tests:
	// equal { failure { "code" 404 "info" "Not Found" } |details? |type? } 'dict
	// equal { failure { "code" 404 "info" "Not Found" } |details? .code } 404
	// error { "not an error" |details? }
	// Args:
	// * error: Error object to extract additional details from
	// Returns:
	// * dictionary containing any additional values stored in the error
	"details?": { // **
		AcceptFailure: true,
		Argsn:         1,
		Doc:           "Extracts additional details from an error object as a dictionary.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			var originalMap map[string]env.Object
			switch er := arg0.(type) {
			case env.Error:
			case *env.Error:
				originalMap = er.Values
			default:
				ps.FailureFlag = true
				return env.NewError("arg 0 not error")
			}
			// originalMap := er.Values // map[string]env.Object
			convertedMap := make(map[string]any)
			for key, value := range originalMap {
				convertedMap[key] = value
			}
			return env.NewDict(convertedMap)
		},
	},

	// Tests:
	// equal { try { fail "error" |disarm } |type? } 'error
	// equal { try { fail "error" |disarm |message? } } "error"
	// equal { try { fail "error" } |failed? } 1
	// equal { try { fail "error" |disarm } |failed? } 0
	// Args:
	// * error: Error object to disarm
	// Returns:
	// * the original error object, but clears the failure flag
	"disarm": { // **
		AcceptFailure: true,
		Argsn:         1,
		Doc:           "Clears the failure flag while preserving the error object, allowing error inspection without propagation.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			ps.FailureFlag = false
			return arg0
		},
	},

	// Tests:
	// equal { failure "error" |failed? } 1
	// equal { "not an error" |failed? } 0
	// equal { 123 |failed? } 0
	// equal { try { fail "error" } |failed? } 1
	// Args:
	// * value: Any value to check
	// Returns:
	// * integer 1 if the value is an error, 0 otherwise
	"failed?": { // **
		AcceptFailure: true,
		Argsn:         1,
		Doc:           "Tests if a value is an error object, returning 1 for errors and 0 for non-errors.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			ps.FailureFlag = false
			switch arg0.(type) {
			case env.Error:
				return *env.NewInteger(int64(1))
			case *env.Error:
				return *env.NewInteger(int64(1))
			}
			return *env.NewInteger(int64(0))
		},
	},

	// ##### Failure combinators ##### "manage flow  when failure happens"

	// Tests:
	// equal { 5 |check "Value must be positive" } 5
	// equal { try { fail "Original error" |check "Wrapped error" } |message? } "Wrapped error"
	// equal { try { fail "Original error" |check "Wrapped error" } |details? |type? } 'dict
	// Args:
	// * value: Value to check for failure state
	// * error_info: Error information to use if value is in failure state
	// Returns:
	// * original value if not in failure state, or a new error wrapping the original error
	"check": { // **
		AcceptFailure: true,
		Argsn:         2,
		Doc:           "Checks if a value is in failure state and wraps it with a new error if so, otherwise returns the original value.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			if ps.FailureFlag || arg0.Type() == env.ErrorType {
				ps.FailureFlag = false
				switch er := arg0.(type) {
				case *env.Error: // todo .. make Error type .. make error construction micro dialect, return the error wrapping error that caused it
					if er.Status == 0 && er.Message == "" {
						er = nil
					}
					return MakeRyeError(ps, arg1, er)
				}
			}
			return arg0
		},
	},

	// Tests:
	// equal { fn { x } { x |^check "Error in function" } |call 5 } 5
	// equal { fn { x } { fail "Original" |^check "Wrapped" } |call 5 |message? } "Wrapped"
	// Args:
	// * value: Value to check for failure state
	// * error_info: Error information to use if value is in failure state
	// Returns:
	// * original value if not in failure state, or immediately returns from function with a new error
	"^check": {
		AcceptFailure: true,
		Argsn:         2,
		Doc:           "Like 'check' but also sets the return flag to immediately exit the current function.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			if ps.FailureFlag {
				ps.ReturnFlag = true
				switch er := arg0.(type) {
				case *env.Error: // todo .. make Error type .. make error construction micro dialect, return the error wrapping error that caused it
					if er.Status == 0 && er.Message == "" {
						er = nil
					}
					return MakeRyeError(ps, arg1, er)
				}
				return env.NewError("error 1")
			}
			return arg0
		},
	},

	// Tests:
	// equal { fn { x } { x > 0 |^ensure "Must be positive" } |call 5 } 5
	// equal { fn { x } { x > 0 |^ensure "Must be positive" } |call -1 |message? } "Must be positive"
	// Args:
	// * condition: Value to test for truthiness
	// * error_info: Error information to use if condition is not truthy
	// Returns:
	// * condition value if truthy, or immediately returns from function with an error
	"^ensure": {
		AcceptFailure: true,
		Argsn:         2,
		Doc:           "Checks if a value is truthy and returns it if so, otherwise creates an error and immediately returns from the function.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch cond := arg0.(type) {
			case env.Object:
				if !util.IsTruthy(cond) {
					ps.FailureFlag = true
					ps.ReturnFlag = true
					return MakeRyeError(ps, arg1, nil)
				} else {
					return arg0
				}
			}
			return arg0
		},
	},

	// Tests:
	// equal { 5 > 0 |ensure "Must be positive" } 1
	// equal { try { -1 > 0 |ensure "Must be positive" } |message? } "Must be positive"
	// Args:
	// * condition: Value to test for truthiness
	// * error_info: Error information to use if condition is not truthy
	// Returns:
	// * condition value if truthy, or creates an error with failure flag set
	"ensure": { // **
		AcceptFailure: true,
		Argsn:         2,
		Doc:           "Checks if a value is truthy and returns it if so, otherwise creates an error with the failure flag set.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch cond := arg0.(type) {
			case env.Object:
				if !util.IsTruthy(cond) {
					ps.FailureFlag = true
					// ps.ReturnFlag = true
					return MakeRyeError(ps, arg1, nil)
				} else {
					return arg0
				}
			}
			return arg0
		},
	},

	// Tests:
	// equal { 5 |fix { + 10 } } 5
	// equal { try { fail "error" |fix { "fixed" } } } "fixed"
	// equal { try { fail "error" |fix { fail "new error" } } |message? } "new error"
	// Args:
	// * value: Value to check for failure state
	// * handler: Block to execute if value is in failure state
	// Returns:
	// * original value if not in failure state, or result of executing the handler block
	"fix": { // **
		AcceptFailure: true,
		Argsn:         2,
		Doc:           "Handles errors by executing a block if the value is in failure state, clearing the failure flag.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			if ps.FailureFlag || arg0.Type() == env.ErrorType {
				ps.FailureFlag = false
				// TODO -- create function do_block and call in all cases
				switch bloc := arg1.(type) {
				case env.Block:
					ser := ps.Ser
					ps.Ser = bloc.Series
					EvalBlockInjMultiDialect(ps, arg0, true)
					ps.Ser = ser
					return ps.Res
				default:
					ps.FailureFlag = true
					return env.NewError("expecting block")
				}
			} else {
				return arg0
			}
		},
	},

	// Tests:
	// equal { fn { x } { x |^fix { "fixed" } } |call 5 } 5
	// equal { fn { x } { fail "error" |^fix { "fixed" } } |call 5 } "fixed"
	// Args:
	// * value: Value to check for failure state
	// * handler: Block to execute if value is in failure state
	// Returns:
	// * original value if not in failure state, or immediately returns from function with handler result
	"^fix": {
		AcceptFailure: true,
		Argsn:         2,
		Doc:           "Like 'fix' but also sets the return flag to immediately exit the current function with the handler result.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			if ps.FailureFlag || arg0.Type() == env.ErrorType {
				ps.FailureFlag = false
				// TODO -- create function do_block and call in all cases
				switch bloc := arg1.(type) {
				case env.Block:
					ser := ps.Ser
					ps.Ser = bloc.Series
					EvalBlockInjMultiDialect(ps, arg0, true)
					ps.Ser = ser
					ps.ReturnFlag = true
					return ps.Res
				default:
					ps.FailureFlag = true
					return env.NewError("expecting block")
				}
			} else {
				return arg0
			}
		},
	},

	"fix\\either": {
		AcceptFailure: true,
		Argsn:         3,
		Doc:           "Fix also with else block.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			if ps.FailureFlag || arg0.Type() == env.ErrorType {
				ps.FailureFlag = false
				// TODO -- create function do_block and call in all cases
				switch bloc := arg1.(type) {
				case env.Block:
					ser := ps.Ser
					ps.Ser = bloc.Series
					EvalBlockInjMultiDialect(ps, arg0, true)
					ps.Ser = ser
					return ps.Res
				default:
					ps.FailureFlag = true
					return env.NewError("expecting block")
				}
			} else {
				switch bloc := arg2.(type) {
				case env.Block:
					ser := ps.Ser
					ps.Ser = bloc.Series
					EvalBlockInjMultiDialect(ps, arg0, true)
					ps.Ser = ser
					return ps.Res
				default:
					ps.FailureFlag = true
					return env.NewError("expecting block")
				}
			}
		},
	},

	"fix\\else": {
		AcceptFailure: true,
		Argsn:         2,
		Doc:           "Do a block of code if Arg 1 is not a failure.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			if !(ps.FailureFlag || arg0.Type() == env.ErrorType) {
				ps.FailureFlag = false
				// TODO -- create function do_block and call in all cases
				switch bloc := arg1.(type) {
				case env.Block:
					ser := ps.Ser
					ps.Ser = bloc.Series
					EvalBlockInjMultiDialect(ps, arg0, true)
					ps.Ser = ser
					return ps.Res
				default:
					ps.FailureFlag = true
					return env.NewError("expecting block")
				}
			} else {
				return arg0
			}
		},
	},

	/* "fix\\when": {
		Argsn: 3,
		Doc:   "Recovers from an error if a condition is met.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			if ps.FailureFlag || arg0.Type() == env.ErrorType {
				ps.FailureFlag = false
				ps.Ser = arg1.Series
				EvalBlockInjMultiDialect(ps, arg0, true)
				if util.IsTruthy(ps.Res) {
					ps.Ser = arg2.Series
					EvalBlockInjMultiDialect(ps, arg0, true)
					return ps.Res
				}
				ps.FailureFlag = true
				return arg0
			}
			return arg0
		},
	}, */

	// Tests:
	// equal  { try { 123 + 123 } } 246
	// equal  { try { 123 + "asd" } \type? } 'error
	// equal  { try { 123 + } \type? } 'error
	"try": { // **
		Argsn: 1,
		Doc:   "Takes a block of code and does (runs) it.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch bloc := arg0.(type) {
			case env.Block:
				ser := ps.Ser
				ps.Ser = bloc.Series
				EvalBlock(ps)

				// TODO -- probably shouldn't just display error ... but we return it and then handle it / display it
				// MaybeDisplayFailureOrError(ps, ps.Idx)

				ps.ReturnFlag = false
				ps.ErrorFlag = false
				ps.FailureFlag = false

				ps.Ser = ser
				return ps.Res
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType}, "try")
			}
		},
	},

	// Tests:
	// equal  { c: context { x: 100 } try\in c { x * 9.99 } } 999.0
	// equal  { c: context { x: 100 } try\in c { inc! 'x } } 101
	// equal  { c: context { x: 100 } try\in c { x:: 200 , x } } 200
	// equal  { c: context { x: 100 } try\in c { x:: 200 } c/x } 200
	// equal  { c: context { x: 100 } try\in c { inc! 'y } |type? } 'error
	"try\\in": { // **
		Argsn: 2,
		Doc:   "Takes a Context and a Block. It Does a block inside a given Context.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch ctx := arg0.(type) {
			case env.RyeCtx:
				switch bloc := arg1.(type) {
				case env.Block:
					ser := ps.Ser
					ps.Ser = bloc.Series
					EvalBlockInCtxInj(ps, &ctx, nil, false)

					// TODO -- probably shouldn't just display error ... but we return it and then handle it / display it
					// MaybeDisplayFailureOrError(ps, ps.Idx)

					ps.ReturnFlag = false
					ps.ErrorFlag = false
					ps.FailureFlag = false

					ps.Ser = ser
					return ps.Res
				default:
					ps.ErrorFlag = true
					return MakeArgError(ps, 2, []env.Type{env.BlockType}, "try\\in")
				}
			default:
				ps.ErrorFlag = true
				return MakeArgError(ps, 1, []env.Type{env.CtxType}, "try\\in")
			}

		},
	},

	/** ` skip words aren't block level and are one too many ... maybe block escape words
	** woul make more sense for loops and other block level operations
	// Tests:
	// equal { for { 1 2 fail "error" 4 } { i } { i |`fix { continue } } } { 1 2 4 }
	// Args:
	// * value: Value to check for failure state
	// * handler: Block to execute if value is in failure state
	// Returns:
	// * original value if not in failure state, or sets skip flag and returns handler result
	"`fix": {
		AcceptFailure: true,
		Argsn:         2,
		Doc:           "Like 'fix' but also sets the skip flag, useful in loops to continue to the next iteration.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			if ps.FailureFlag || arg0.Type() == env.ErrorType {
				ps.FailureFlag = false
				// TODO -- create function do_block and call in all cases
				switch bloc := arg1.(type) {
				case env.Block:
					ser := ps.Ser
					ps.Ser = bloc.Series
					EvalBlockInjMultiDialect(ps, arg0, true)
					ps.Ser = ser
					ps.SkipFlag = true
					return ps.Res
				default:
					ps.FailureFlag = true
					return env.NewError("expecting block")
				}
			} else {
				return arg0
			}
		},
	},
	*/

	// Tests:
	// equal { 5 |fix\continue { "error handler" } { "success handler" } } "success handler"
	// equal { try { fail "error" |fix\continue { "error handler" } { "success handler" } } } "error handler"
	// Args:
	// * value: Value to check for failure state
	// * error_handler: Block to execute if value is in failure state
	// * success_handler: Block to execute if value is not in failure state
	// Returns:
	// * result of executing the appropriate handler block
	"fix\\continue": {
		AcceptFailure: true,
		Argsn:         3,
		Doc:           "Executes one of two blocks depending on whether the value is in failure state, like an error-handling if/else.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			if ps.FailureFlag || arg0.Type() == env.ErrorType {
				ps.FailureFlag = false
				// TODO -- create function do_block and call in all cases
				switch bloc := arg1.(type) {
				case env.Block:
					ser := ps.Ser
					ps.Ser = bloc.Series
					EvalBlockInjMultiDialect(ps, arg0, true)
					ps.Ser = ser
					return ps.Res
				default:
					ps.FailureFlag = true
					return env.NewError("expecting block")
				}
			} else {
				switch bloc := arg2.(type) {
				case env.Block:
					ser := ps.Ser
					ps.Ser = bloc.Series
					EvalBlockInjMultiDialect(ps, arg0, true)
					ps.Ser = ser
					return ps.Res
				default:
					ps.FailureFlag = true
					return env.NewError("expecting block")
				}
			}
		},
	},

	// Tests:
	// equal { 5 |continue { + 10 } } 15
	// equal { try { fail "error" |continue { + 10 } } } failure "error"
	// Args:
	// * value: Value to check for failure state
	// * block: Block to execute if value is not in failure state
	// Returns:
	// * result of executing the block if value is not in failure state, or the original value
	"continue": {
		AcceptFailure: true,
		Argsn:         2,
		Doc:           "Executes a block only if the value is not in failure state, opposite of 'fix'.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			if !(ps.FailureFlag || arg0.Type() == env.ErrorType) {
				ps.FailureFlag = false
				// TODO -- create function do_block and call in all cases
				switch bloc := arg1.(type) {
				case env.Block:
					ser := ps.Ser
					ps.Ser = bloc.Series
					EvalBlockInjMultiDialect(ps, arg0, true)
					ps.Ser = ser
					return ps.Res
				default:
					ps.FailureFlag = true
					return env.NewError("expecting block")
				}
			} else {
				ps.FailureFlag = false
				return arg0
			}
		},
	},

	// Tests:
	// ; equal  { fail 404 |^fix\match { 404 { "ER1" } 305 { "ER2" } } } "ER1"
	// Args:
	// * error: Error object to match against
	// * cases: Block containing error codes and corresponding handler blocks
	// Returns:
	// * result of executing the matching handler block, or the original error if no match
	"^fix\\match": {
		Argsn:         2,
		Doc:           "Error handling switch that matches error codes with handler blocks and sets the return flag.",
		AcceptFailure: true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			fmt.Println("FLAGS")

			ps.FailureFlag = false

			switch er := arg0.(type) {
			case env.Error:
				fmt.Println("ERR")

				switch bloc := arg1.(type) {
				case env.Block:

					var code env.Object

					any_found := false
					fmt.Println("BLOCK")

					for i := 0; i < bloc.Series.Len(); i += 2 {
						fmt.Println("LOOP")

						if i > bloc.Series.Len()-2 {
							return MakeBuiltinError(ps, "Switch block malformed.", "^tidy-switch")
						}

						switch ev := bloc.Series.Get(i).(type) {
						case env.Integer:
							if er.Status == int(ev.Value) {
								any_found = true
								code = bloc.Series.Get(i + 1)
							}
						case env.Void:
							fmt.Println("VOID")
							if !any_found {
								code = bloc.Series.Get(i + 1)
								any_found = false
							}
						default:
							return MakeBuiltinError(ps, "Invalid type in block series.", "^tidy-switch")
						}
					}
					switch cc := code.(type) {
					case env.Block:
						fmt.Println(code.Print(*ps.Idx))
						// we store current series (block of code with position we are at) to temp 'ser'
						ser := ps.Ser
						// we set ProgramStates series to series ob the block
						ps.Ser = cc.Series
						// we eval the block (current context / scope stays the same as it was in parent block)
						// Inj means we inject the condition value into the block, because it costs us very little. we could do "if name { .print }"
						EvalBlockInjMultiDialect(ps, arg0, true)
						// we set temporary series back to current program state
						ps.Ser = ser
						// we return the last return value (the return value of executing the block) "a: if 1 { 100 }" a becomes 100,
						// in future we will also handle the "else" case, but we have to decide
						//						ps.ReturnFlag = true

						ps.ReturnFlag = true
						ps.FailureFlag = true
						return arg0
					default:
						// if it's not a block we return error for now
						ps.FailureFlag = true
						return MakeBuiltinError(ps, "Malformed switch block.", "^tidy-switch")
					}
				default:
					// if it's not a block we return error for now
					ps.FailureFlag = true
					return MakeArgError(ps, 2, []env.Type{env.BlockType}, "^tidy-switch")
				}
			default:
				return arg0
			}
		},
	},

	// Tests:
	// equal { retry 3 { fail 101 } |type? } 'error
	// equal { retry 3 { fail 101 } |status? } 101
	// equal { retry 3 { 10 + 1 } } 11
	// Args:
	// * retries: Integer number of retries to attempt
	// * block: Block of code to execute and potentially retry
	// Returns:
	// * result of the block if successful, or the last failure if all retries fail
	"retry": {
		Argsn:         2,
		Doc:           "Executes a block and retries it up to N times if it results in a failure.",
		AcceptFailure: true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch retries := arg0.(type) {
			case env.Integer:
				switch bloc := arg1.(type) {
				case env.Block:
					// Store current series
					ser := ps.Ser

					// Try the block initially
					ps.Ser = bloc.Series
					EvalBlock(ps)

					// If it succeeded (no failure flag), return the result immediately
					if !ps.FailureFlag {
						ps.Ser = ser
						return ps.Res
					}

					// Store the initial failure result
					result := ps.Res

					// Retry up to N-1 more times (we already tried once)
					for i := int64(1); i < retries.Value; i++ {
						// Reset the series and failure flag
						ps.Ser.Reset()
						ps.FailureFlag = false
						ps.ErrorFlag = false

						// Execute the block again
						EvalBlock(ps)

						// If it succeeded, return the result
						if !ps.FailureFlag {
							ps.Ser = ser
							return ps.Res
						}

						// Update the result to the latest failure
						result = ps.Res
					}

					// Restore the original series and return the last failure result
					ps.Ser = ser
					ps.FailureFlag = true
					return result
				default:
					return MakeArgError(ps, 2, []env.Type{env.BlockType}, "retry")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType}, "retry")
			}
		},
	},

	// Tests:
	// equal { timeout 5000 { "ok" } } "ok"
	// equal { try { timeout 100 { sleep 1000 , "ok" } } |message? |contains "timeout" } 1
	// Args:
	// * ms: Integer timeout duration in milliseconds
	// * block: Block of code to execute with a timeout
	// Returns:
	// * result of the block if it completes within the timeout, or a timeout error
	"timeout": {
		AcceptFailure: true,
		Argsn:         2,
		Doc:           "Executes a block of code with a timeout, failing if execution exceeds the specified duration in milliseconds.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch ms := arg0.(type) {
			case env.Integer:
				switch bloc := arg1.(type) {
				case env.Block:
					// Store current series
					ser := ps.Ser

					// Create channels for result and completion
					resultChan := make(chan env.Object, 1)
					doneChan := make(chan bool, 1)

					// Create a new program state for the goroutine
					psCopy := env.NewProgramState(bloc.Series, ps.Idx)
					psCopy.Ctx = ps.Ctx
					psCopy.PCtx = ps.PCtx
					psCopy.Gen = ps.Gen

					// Execute the block in a goroutine
					go func() {
						EvalBlock(psCopy)
						resultChan <- psCopy.Res
						doneChan <- psCopy.FailureFlag
					}()

					// Set up timeout
					timeoutDuration := time.Duration(ms.Value) * time.Millisecond

					// Wait for either completion or timeout
					select {
					case result := <-resultChan:
						ps.FailureFlag = <-doneChan
						ps.Ser = ser
						return result

					case <-time.After(timeoutDuration):
						ps.FailureFlag = true
						ps.Ser = ser
						return MakeRyeError(ps, *env.NewString(fmt.Sprintf("Execution timed out after %d ms", ms.Value)), nil)
					}
				default:
					return MakeArgError(ps, 2, []env.Type{env.BlockType}, "timeout")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType}, "timeout")
			}
		},
	},
}
