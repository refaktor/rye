package evaldo

import (
	"github.com/refaktor/rye/env"
	"github.com/refaktor/rye/util"
	// JM 20230825	"github.com/refaktor/rye/term"
)

var Builtins_failure = map[string]*env.Builtin{

	//
	// ##### Failure ###### "Error handling and failure management functions"
	//

	// Tests:
	// equal { fn { } { ^fail "error message" } |type? } 'error
	// equal { fn { } { ^fail "error message" } |message? } "error message"
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
			return MakeRyeError(ps, arg0, nil)
		},
	},

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
			return MakeRyeError(ps, arg0, nil)
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
			return MakeRyeError(ps, arg0, nil)
		},
	},

	// Tests:
	// equal { wrap\failure "outer error" failure "inner error" |message? } "outer error"
	// equal { wrap\failure "outer error" failure "inner error" |type? } 'error
	// Args:
	// * error_info: String message, Integer code, or block for multiple parameters
	// * error: Error object to wrap
	// Returns:
	// * new error object that wraps the provided error
	"wrap\\failure": {
		Argsn: 2,
		Doc:   "Creates a new error that wraps an existing error, allowing for error chaining.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch er := arg1.(type) {
			case *env.Error:
				return MakeRyeError(ps, arg0, er)
			default:
				return MakeArgError(ps, 2, []env.Type{env.ErrorType}, "wrap\\failure")
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
	// equal { fn { x } { x > 0 |^require "Must be positive" } |call 5 } 5
	// equal { fn { x } { x > 0 |^require "Must be positive" } |call -1 |message? } "Must be positive"
	// Args:
	// * condition: Value to test for truthiness
	// * error_info: Error information to use if condition is not truthy
	// Returns:
	// * condition value if truthy, or immediately returns from function with an error
	"^require": {
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
	// equal { 5 > 0 |require "Must be positive" } 1
	// equal { try { -1 > 0 |require "Must be positive" } |message? } "Must be positive"
	// Args:
	// * condition: Value to test for truthiness
	// * error_info: Error information to use if condition is not truthy
	// Returns:
	// * condition value if truthy, or creates an error with failure flag set
	"require": { // **
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
	// equal { assert-equal 5 5 } 1
	// equal { try { assert-equal 5 6 } |type? } 'error
	// equal { try { assert-equal "abc" "def" } |message? |contains "not equal" } 1
	// Args:
	// * value1: First value to compare
	// * value2: Second value to compare
	// Returns:
	// * integer 1 if values are equal, or creates an error if they are not equal
	"assert-equal": { // **
		Argsn: 2,
		Doc:   "Tests if two values are equal using the Equal method, returning 1 if equal or creating an error if not.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			if arg0.Equal(arg1) {
				return *env.NewInteger(1)
			} else {
				return makeError(ps, "Values are not equal: "+arg0.Inspect(*ps.Idx)+" "+arg1.Inspect(*ps.Idx))
			}
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

	// Tests:
	// equal { 5 |fix\either { "error handler" } { "success handler" } } "success handler"
	// equal { try { fail "error" |fix\either { "error handler" } { "success handler" } } } "error handler"
	// Args:
	// * value: Value to check for failure state
	// * error_handler: Block to execute if value is in failure state
	// * success_handler: Block to execute if value is not in failure state
	// Returns:
	// * result of executing the appropriate handler block
	"fix\\either": {
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
}
