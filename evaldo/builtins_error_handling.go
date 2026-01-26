// builtins_error_handling.go
package evaldo

import (
	"fmt"
	"time"

	"github.com/refaktor/rye/env"
	"github.com/refaktor/rye/util"
)

// Error Creation Functions
var ErrorCreationBuiltins = map[string]*env.Builtin{
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
	"fail": {
		Argsn: 1,
		Doc:   "Creates an error and sets the failure flag, but continues execution (unlike ^fail).",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			ps.FailureFlag = true
			return MakeRyeError(ps, arg0, nil)
		},
	},

	// Tests:
	// equal { ff:: fn { } { ^fail "error message" } ff |disarm |type? } 'error
	// equal { ff:: fn { } { ^fail "error message" } ff |disarm |message? } "error message"
	// equal { ff:: fn { } { ^fail 404 } ff |disarm |status? } 404
	// equal { ff:: fn { } { ^fail 'user-error } ff |disarm  |kind? } 'user-error
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

	"refail": {
		Argsn: 2,
		Doc:   "Re-raises an existing error with additional context. TODO -- duplicate of check",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch er := arg0.(type) {
			case *env.Error:
				ps.FailureFlag = true
				return MakeRyeError(ps, arg1, er)
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
	"failure": {
		Argsn: 1,
		Doc:   "Creates an error object without setting any flags (unlike fail and ^fail).",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return MakeRyeError(ps, arg0, nil)
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
				return MakeRyeError(ps, arg0, er)
			default:
				return MakeArgError(ps, 2, []env.Type{env.ErrorType}, "wrap\\failure")
			}
		},
	},
}

// Error Inspection Functions
var ErrorInspectionBuiltins = map[string]*env.Builtin{
	// Tests:
	// equal { is-error failure "test" } true
	// equal { is-error 123 } false
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
	// equal { is-error-of-kind failure { 'syntax-error 404 "Syntax error" } 'syntax-error } true
	// equal { is-error-of-kind failure { 'syntax-error 404 "Syntax error" } 'runtime-error } false
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
				return MakeArgError(ps, 2, []env.Type{env.WordType, env.TagwordType}, "is-error-of-kind")
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

	// Tests:
	// equal { cause? failure\wrap "outer error" failure "inner error" |message? } "inner error"
	// equal { cause? failure "single error" |message? } "single error"
	// Args:
	// * error: Error object to extract the root cause from
	// Returns:
	// * the root cause error from an error chain
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
				return MakeArgError(ps, 1, []env.Type{env.ErrorType}, "cause?")
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
	"status?": {
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
				return MakeArgError(ps, 1, []env.Type{env.ErrorType}, "status?")
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
	"message?": {
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
				return MakeArgError(ps, 1, []env.Type{env.ErrorType}, "message?")
			}
		},
	},

	// Tests:
	// ; equal { failure { "code" 404 "info" "Not Found" } |details? |type? } 'dict
	// ; equal { failure { "code" 404 "info" "Not Found" } |details? .code } 404
	// error { "not an error" |details? }
	// Args:
	// * error: Error object to extract additional details from
	// Returns:
	// * dictionary containing any additional values stored in the error
	"details?": {
		AcceptFailure: true,
		Argsn:         1,
		Doc:           "Extracts additional details from an error object as a dictionary.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			var originalMap map[string]env.Object
			switch er := arg0.(type) {
			case env.Error:
				originalMap = er.Values
			case *env.Error:
				originalMap = er.Values
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.ErrorType}, "details?")
			}

			convertedMap := make(map[string]any)
			for key, value := range originalMap {
				convertedMap[key] = value
			}
			return env.NewDict(convertedMap)
		},
	},

	// Tests:
	// equal { has-failed failure "error" } true
	// equal { has-failed "not an error" } false
	// equal { has-failed 123 } false
	// equal { has-failed try { fail "error" } } true
	// Args:
	// * value: Any value to check
	// Returns:
	// * boolean true if the value is an error, false otherwise
	"has-failed": {
		AcceptFailure: true,
		Argsn:         1,
		Doc:           "Tests if a value is an error object, returning true for errors and false for non-errors.",
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
	// equal { try { fail "error" } |disarm |is-failure } true
	// equal { failure "error" |is-failure } true
	// equal { 123 |is-failure } false
	// equal { "hello" |is-failure } false
	// Args:
	// * value: Any value to check (must already be disarmed if it was a failure)
	// Returns:
	// * boolean true if the value is an error type, false otherwise
	"is-failure": {
		Argsn: 1,
		Doc:   "Checks if a value is a failure/error type. Unlike has-failed, this doesn't accept failures - the value must already be disarmed.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch arg0.(type) {
			case env.Error, *env.Error:
				return *env.NewBoolean(true)
			}
			return *env.NewBoolean(false)
		},
	},

	// Tests:
	// equal { 123 |is-success } true
	// equal { "hello" |is-success } true
	// equal { { 1 2 3 } |is-success } true
	// equal { failure "error" |is-success } false
	// equal { try { fail "error" } |disarm |is-success } false
	// Args:
	// * value: Any value to check (must already be disarmed if it was a failure)
	// Returns:
	// * boolean true if the value is not an error type, false if it is an error
	"is-success": {
		Argsn: 1,
		Doc:   "Returns true for any value that is not a failure/error type. The opposite of is-failure.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch arg0.(type) {
			case env.Error, *env.Error:
				return *env.NewBoolean(false)
			}
			return *env.NewBoolean(true)
		},
	},
}

// Error Handling Functions
var ErrorHandlingBuiltins = map[string]*env.Builtin{
	// Tests:
	// equal { try { fail "error" |disarm } |type? } 'error
	// equal { try { fail "error" |disarm |message? } } "error"
	// equal { try { fail "error" } |has-failed } true
	// equal { try { fail "error" |disarm } |has-failed } false
	// Args:
	// * error: Error object to disarm
	// Returns:
	// * the original error object, but clears the failure flag
	"disarm": {
		AcceptFailure: true,
		Argsn:         1,
		Doc:           "Clears the failure flag while preserving the error object, allowing error inspection without propagation.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			ps.FailureFlag = false
			return arg0
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
	"check": {
		AcceptFailure: true,
		Argsn:         2,
		Doc:           "Checks if a value is in failure state and wraps it with a new error if so, otherwise returns the original value.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			if ps.FailureFlag || arg0.Type() == env.ErrorType {
				// ps.FailureFlag = false
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
	// equal { fn { x } { x |^check "Error in function" } |apply [ 5 ] } 5
	// equal { ff: fn { x } { fail "Original" |^check "Wrapped" } ff 5 |disarm |message? } "Wrapped"
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
				case env.Error: // todo .. make Error type .. make error construction micro dialect, return the error wrapping error that caused it
					// if er.Status == 0 && er.Message == "" {
					// er = nil
					// }
					return MakeRyeError(ps, arg1, &er)
				}
				return env.NewError("error 12221")
			}
			return arg0
		},
	},

	// Tests:
	// equal { fn { x } { x > 0 |^ensure "Must be positive" } |apply [ 5 ] } true
	// equal { ff:: fn { x } { x > 0 |^ensure "Must be positive" } ff -1 |disarm |message? } "Must be positive"
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
	// equal { 5 > 0 |ensure "Must be positive" } true
	// equal { try { -1 > 0 |ensure "Must be positive" } |message? } "Must be positive"
	// Args:
	// * condition: Value to test for truthiness
	// * error_info: Error information to use if condition is not truthy
	// Returns:
	// * condition value if truthy, or creates an error with failure flag set
	"ensure": {
		AcceptFailure: true,
		Argsn:         2,
		Doc:           "Checks if a value is truthy and returns it if so, otherwise creates an error with the failure flag set.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch cond := arg0.(type) {
			case env.Object:
				if !util.IsTruthy(cond) {
					ps.FailureFlag = true
					return MakeRyeError(ps, arg1, nil)
				} else {
					return arg0
				}
			}
			return arg0
		},
	},

	// Tests:
	// equal { "a" |requires-one-of { "a" "b" "c" } } "a"
	// equal { "b" |requires-one-of { "a" "b" "c" } } "b"
	// equal { try { "x" |requires-one-of { "a" "b" "c" } } |message? |contains "must be one of" } 1
	// equal { try { "x" |requires-one-of { "a" "b" "c" } } |message? |contains "\"x\"" } 1
	// equal { try { 5 |requires-one-of { 1 2 3 } } |message? |contains "5" } 1
	// Args:
	// * value: Value to check against valid options
	// * options: Block containing valid values
	// Returns:
	// * original value if it matches one of the options, or creates an error with failure flag set
	"requires-one-of": {
		AcceptFailure: true,
		Argsn:         2,
		Doc:           "Validates that a value matches one of the provided options, failing with a descriptive error if not.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch optionsBlock := arg1.(type) {
			case env.Block:
				// Check if value matches any option
				options := optionsBlock.Series.GetAll()
				for _, option := range options {
					// Compare both kind and inspect output for equality
					if arg0.GetKind() == option.GetKind() && arg0.Inspect(*ps.Idx) == option.Inspect(*ps.Idx) {
						return arg0
					}
				}

				// Value doesn't match any option - create descriptive error
				ps.FailureFlag = true

				// Format the value for display
				valueStr := arg0.Inspect(*ps.Idx)

				// Build error message with options
				var errorMsg string
				if len(options) > 0 {
					optionsStr := "{ "
					for i, opt := range options {
						if i > 0 {
							optionsStr += " "
						}
						optionsStr += opt.Inspect(*ps.Idx)
					}
					optionsStr += " }"
					errorMsg = fmt.Sprintf("Value %s must be one of %s", valueStr, optionsStr)
				} else {
					errorMsg = fmt.Sprintf("Value %s does not match any valid option", valueStr)
				}

				return MakeRyeError(ps, *env.NewString(errorMsg), nil)
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 2, []env.Type{env.BlockType}, "requires-one-of")
			}
		},
	},

	// Tests:
	// equal { "a" |requires-one-of { "a" "b" "c" } } "a"
	// equal { "b" |requires-one-of { "a" "b" "c" } } "b"
	// equal { try { "x" |requires-one-of { "a" "b" "c" } } |message? |contains "must be one of" } 1
	// equal { try { "x" |requires-one-of { "a" "b" "c" } } |message? |contains "\"x\"" } 1
	// equal { try { 5 |requires-one-of { 1 2 3 } } |message? |contains "5" } 1
	// Args:
	// * value: Value to check against valid options
	// * options: Block containing valid values
	// Returns:
	// * original value if it matches one of the options, or creates an error with failure flag set
	"^requires-one-of": {
		AcceptFailure: true,
		Argsn:         2,
		Doc:           "Validates that a value matches one of the provided options, failing with a descriptive error if not.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch optionsBlock := arg1.(type) {
			case env.Block:
				// Check if value matches any option
				options := optionsBlock.Series.GetAll()
				for _, option := range options {
					// Compare both kind and inspect output for equality
					if arg0.GetKind() == option.GetKind() && arg0.Inspect(*ps.Idx) == option.Inspect(*ps.Idx) {
						return arg0
					}
				}

				// Value doesn't match any option - create descriptive error
				ps.FailureFlag = true

				// Format the value for display
				valueStr := arg0.Inspect(*ps.Idx)

				// Build error message with options
				var errorMsg string
				if len(options) > 0 {
					optionsStr := "{ "
					for i, opt := range options {
						if i > 0 {
							optionsStr += " "
						}
						optionsStr += opt.Inspect(*ps.Idx)
					}
					optionsStr += " }"
					errorMsg = fmt.Sprintf("Value %s must be one of %s", valueStr, optionsStr)
				} else {
					errorMsg = fmt.Sprintf("Value %s does not match any valid option", valueStr)
				}
				ps.ReturnFlag = true
				return MakeRyeError(ps, *env.NewString(errorMsg), nil)
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 2, []env.Type{env.BlockType}, "requires-one-of")
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
	"fix": {
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
					EvalBlockInj(ps, arg0, true)
					MaybeDisplayFailureOrError(ps, ps.Idx, "fix")
					if ps.ErrorFlag {
						ps.Ser = ser
						return ps.Res
					}
					ps.Ser = ser
					return ps.Res
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 2, []env.Type{env.BlockType}, "fix")
				}
			} else {
				return arg0
			}
		},
	},

	// Tests:
	// equal { fn { x } { x |^fix { "fixed" } } |apply [ 5 ] } 5
	// equal { ff:: fn { x } { fail "error" |^fix { "fixed" } } ff 5 } "fixed"
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
					EvalBlockInj(ps, arg0, true)
					MaybeDisplayFailureOrError(ps, ps.Idx, "^fix")
					ps.Ser = ser
					ps.ReturnFlag = true
					return ps.Res
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 2, []env.Type{env.BlockType}, "^fix")
				}
			} else {
				return arg0
			}
		},
	},

	// Tests:
	// equal { fix\either failure "error" { "fixed" } { "not fixed" } } "fixed"
	// equal { fix\either 5 { "fixed" } { "not fixed" } } "not fixed"
	// Args:
	// * value: Value to check for failure state
	// * error_handler: Block to execute if value is in failure state
	// * success_handler: Block to execute if value is not in failure state
	// Returns:
	// * result of executing the appropriate handler block
	"fix\\either": {
		AcceptFailure: true,
		Argsn:         3,
		Doc:           "Executes one of two blocks depending on whether the value is in failure state.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			if ps.FailureFlag || arg0.Type() == env.ErrorType {
				ps.FailureFlag = false
				// TODO -- create function do_block and call in all cases
				switch bloc := arg1.(type) {
				case env.Block:
					ser := ps.Ser
					ps.Ser = bloc.Series
					EvalBlockInj(ps, arg0, true)
					MaybeDisplayFailureOrError(ps, ps.Idx, "fix\\either")
					ps.Ser = ser
					return ps.Res
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 2, []env.Type{env.BlockType}, "fix\\either")
				}
			} else {
				switch bloc := arg2.(type) {
				case env.Block:
					ser := ps.Ser
					ps.Ser = bloc.Series
					EvalBlockInj(ps, arg0, true)
					MaybeDisplayFailureOrError(ps, ps.Idx, "fix\\either")
					ps.Ser = ser
					return ps.Res
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 3, []env.Type{env.BlockType}, "fix\\either")
				}
			}
		},
	},

	// Tests:
	// equal { 5 |fix\else { "not fixed" } } "not fixed"
	// equal { try { fail "error" |fix\else { "not fixed" } } |message? } "error"
	// Args:
	// * value: Value to check for failure state
	// * success_handler: Block to execute if value is not in failure state
	// Returns:
	// * result of executing the success handler if value is not in failure state, or the original value
	"fix\\else": {
		AcceptFailure: true,
		Argsn:         2,
		Doc:           "Executes a block if the value is not in failure state, otherwise returns the original value.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			if !(ps.FailureFlag || arg0.Type() == env.ErrorType) {
				ps.FailureFlag = false
				// TODO -- create function do_block and call in all cases
				switch bloc := arg1.(type) {
				case env.Block:
					ser := ps.Ser
					ps.Ser = bloc.Series
					EvalBlockInj(ps, arg0, true)
					MaybeDisplayFailureOrError(ps, ps.Idx, "fix\\else")
					ps.Ser = ser
					return ps.Res
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 2, []env.Type{env.BlockType}, "fix\\else")
				}
			} else {
				return arg0
			}
		},
	},

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
					EvalBlockInj(ps, arg0, true)
					ps.Ser = ser
					return ps.Res
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 2, []env.Type{env.BlockType}, "fix\\continue")
				}
			} else {
				switch bloc := arg2.(type) {
				case env.Block:
					ser := ps.Ser
					ps.Ser = bloc.Series
					EvalBlockInj(ps, arg0, true)
					ps.Ser = ser
					return ps.Res
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 3, []env.Type{env.BlockType}, "fix\\continue")
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
					EvalBlockInj(ps, arg0, true)
					MaybeDisplayFailureOrError(ps, ps.Idx, "continue")
					ps.Ser = ser
					return ps.Res
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 2, []env.Type{env.BlockType}, "continue")
				}
			} else {
				ps.FailureFlag = false
				return arg0
			}
		},
	},

	// Tests:
	// equal { ^fix\match failure 404 { 404 { "Not Found" } 500 { "Server Error" } } } "Not Found"
	// equal { ^fix\match failure 500 { 404 { "Not Found" } 500 { "Server Error" } } } "Server Error"
	// equal { ^fix\match failure 403 { 404 { "Not Found" } 500 { "Server Error" } _ { "Unknown Error" } } } "Unknown Error"
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
			ps.FailureFlag = false

			switch er := arg0.(type) {
			case env.Error, *env.Error:
				switch bloc := arg1.(type) {
				case env.Block:
					var code env.Object
					any_found := false

					for i := 0; i < bloc.Series.Len(); i += 2 {
						if i > bloc.Series.Len()-2 {
							return MakeBuiltinError(ps, "Switch block malformed.", "^fix\\match")
						}

						switch ev := bloc.Series.Get(i).(type) {
						case env.Integer:
							var status int
							if err, ok := er.(env.Error); ok {
								status = err.Status
							} else if errPtr, ok := er.(*env.Error); ok {
								status = errPtr.Status
							}

							if status == int(ev.Value) {
								any_found = true
								code = bloc.Series.Get(i + 1)
							}
						case env.Void:
							if !any_found {
								code = bloc.Series.Get(i + 1)
								any_found = true
							}
						default:
							return MakeBuiltinError(ps, "Invalid type in block series.", "^fix\\match")
						}
					}

					switch cc := code.(type) {
					case env.Block:
						// we store current series (block of code with position we are at) to temp 'ser'
						ser := ps.Ser
						// we set ProgramStates series to series ob the block
						ps.Ser = cc.Series
						// we eval the block (current context / scope stays the same as it was in parent block)
						EvalBlockInj(ps, arg0, true)
						// we set temporary series back to current program state
						MaybeDisplayFailureOrError(ps, ps.Idx, "fix\\match")

						ps.Ser = ser
						// we return the last return value (the return value of executing the block)
						ps.ReturnFlag = true
						return ps.Res
					default:
						// if it's not a block we return error for now
						ps.FailureFlag = true
						return MakeBuiltinError(ps, "Malformed switch block.", "^fix\\match")
					}
				default:
					// if it's not a block we return error for now
					ps.FailureFlag = true
					return MakeArgError(ps, 2, []env.Type{env.BlockType}, "^fix\\match")
				}
			default:
				return arg0
			}
		},
	},

	// Tests:
	// equal  { try { 123 + 123 } } 246
	// equal  { try { 123 + "asd" } |type? } 'error
	// equal  { try { 123 + } |type? } 'error
	"try": {
		Argsn: 1,
		Doc:   "Takes a block of code and does (runs) it.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch bloc := arg0.(type) {
			case env.Block:
				ser := ps.Ser
				ps.Ser = bloc.Series
				ps.InErrHandler = true
				EvalBlock(ps)
				// We don't display it in try function, that is the point
				// MaybeDisplayFailureOrError(ps, ps.Idx, "try")
				ps.InErrHandler = false
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
	// equal { try-all { 1 + 2 } } [ true 3 ]
	// equal { try-all { fail "custom error" } |first } false
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
				ps.InErrHandler = true
				// Execute the block
				ps.Ser = bloc.Series
				EvalBlock(ps)
				MaybeDisplayFailureOrError(ps, ps.Idx, "try-all")

				// Create result tuple
				result := ps.Res
				success := !ps.FailureFlag && !ps.ErrorFlag
				ps.InErrHandler = false
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
	// equal  { c: context { x: 100 } try\in c { x * 9.99 } } 999.0
	// equal  { c: context { x: 100 } try\in c { inc! 'x } } 101
	// equal  { c: context { x: 100 } try\in c { x:: 200 , x } } 200
	// equal  { c: context { x: 100 } try\in c { x:: 200 } c/x } 200
	// equal  { c: context { x: 100 } try\in c { inc! 'y } |type? } 'error
	"try\\in": {
		Argsn: 2,
		Doc:   "Takes a Context and a Block. It Does a block inside a given Context.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch ctx := arg0.(type) {
			case env.RyeCtx:
				switch bloc := arg1.(type) {
				case env.Block:
					ser := ps.Ser
					ps.Ser = bloc.Series
					ps.InErrHandler = true
					EvalBlockInCtxInj(ps, &ctx, nil, false)
					MaybeDisplayFailureOrError(ps, ps.Idx, "try\\in")
					ps.InErrHandler = false
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
				return MakeArgError(ps, 1, []env.Type{env.ContextType}, "try\\in")
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
					MaybeDisplayFailureOrError(ps, ps.Idx, "finally")

					// Save result and flags
					result := ps.Res
					hasFailure := ps.FailureFlag
					hasError := ps.ErrorFlag

					// Execute the finally block
					ps.Ser = finallyBlock.Series
					ps.FailureFlag = false
					ps.ErrorFlag = false
					EvalBlock(ps)
					MaybeDisplayFailureOrError(ps, ps.Idx, "finally")

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

	// Tests:
	// ; equal { retry 3 { fail 101 } |disarm |type? } 'error
	// ; equal { retry 3 { fail 101 } |disarm |status? } 101
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
					MaybeDisplayFailureOrError(ps, ps.Idx, "retry")

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
						MaybeDisplayFailureOrError(ps, ps.Idx, "retry")

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
	// equal { persist { 10 + 1 } } 11
	// equal { counter:: 0 persist { counter:: counter + 1 , if counter > 3 { counter } { fail "not ready" } } } 4
	// error { persist { fail "always fails" } }
	// Args:
	// * block: Block of code to execute repeatedly until it succeeds
	// Returns:
	// * result of the block when it finally succeeds (no failure), or error if 1000 attempts exceeded
	"persist": {
		Argsn:         1,
		Doc:           "Executes a block repeatedly until it succeeds (no failure), then returns the successful result. Gives up after 1000 attempts.",
		AcceptFailure: true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch bloc := arg0.(type) {
			case env.Block:
				// Store current series
				ser := ps.Ser

				// Keep trying up to 1000 attempts
				for attempt := 1; attempt <= 1000; attempt++ {
					// Reset the series and flags for each attempt
					ps.Ser = bloc.Series
					ps.Ser.Reset()
					ps.FailureFlag = false
					ps.ErrorFlag = false

					// Execute the block
					EvalBlock(ps)
					MaybeDisplayFailureOrError(ps, ps.Idx, "persist")

					// If it succeeded (no failure flag), return the result
					if !ps.FailureFlag {
						ps.Ser = ser
						return ps.Res
					}

					// If it failed, continue to next attempt (unless we've reached the limit)
				}

				// If we reach here, we've exceeded 1000 attempts
				ps.Ser = ser
				ps.FailureFlag = true
				return MakeRyeError(ps, *env.NewString("Persist failed: Block did not succeed after 1000 attempts"), nil)
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType}, "persist")
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
						MaybeDisplayFailureOrError(ps, ps.Idx, "timeout")

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

// Export the error handling builtins for registration
var Builtins_error_creation = ErrorCreationBuiltins
var Builtins_error_inspection = ErrorInspectionBuiltins
var Builtins_error_handling = ErrorHandlingBuiltins
