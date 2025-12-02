package evaldo

import (
	"github.com/refaktor/rye/env"
	"github.com/refaktor/rye/util"
)

var builtins_conditionals = map[string]*env.Builtin{

	//
	// ##### Flow control ##### "Functions for conditional execution and branching logic"
	//
	// Tests:
	// equal  { if true { 222 } } 222
	// equal  { if false { 333 } } false
	// error  { if 1 { 222 } }
	// error  { if 0 { 333 } }
	// Args:
	// * condition: Boolean value determining whether to execute the block
	// * block: Block of code to execute if condition is true
	// Returns:
	// * result of the block if condition is true, false otherwise
	"if": { // **
		Argsn: 2,
		Doc:   "Executes a block of code only if the condition is true, returning the result of the block or false.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			// function accepts 2 args. arg0 is a boolean value, arg1 is a block of code

			// Check if the first argument is a boolean
			switch cond := arg0.(type) {
			case env.Boolean:
				// we switch on the type of second argument, so far it should be block (later we could accept native and function)
				switch bloc := arg1.(type) {
				case env.Block:
					// if cond.Value is true, execute the block
					if cond.Value {
						// we store current series (block of code with position we are at) to temp 'ser'
						ser := ps.Ser
						// we set ProgramStates series to series ob the block
						ps.Ser = bloc.Series
						// we eval the block (current context / scope stays the same as it was in parent block)
						// Inj means we inject the condition value into the block, because it costs us very little. we could do "if name { .print }"
						EvalBlockInjMultiDialect(ps, arg0, true)
						if ps.ErrorFlag {
							ps.Ser = ser
							return ps.Res
						}
						// we set temporary series back to current program state
						ps.Ser = ser
						// we return the last return value (the return value of executing the block)
						return ps.Res
					}
					return *env.NewBoolean(false)
				default:
					// if it's not a block we return error
					ps.FailureFlag = true
					return MakeArgError(ps, 2, []env.Type{env.BlockType}, "if")
				}
			default:
				// if it's not a boolean we return error
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.BooleanType}, "if")
			}
		},
	},

	// Tests:
	// equal  { 10 .when { > 5 } { + 3 } } 13
	// equal  { 10 .when { < 5 } { + 3 } } 10
	// Args:
	// * value: Value to inject into both condition and action blocks
	// * condition: Block that evaluates to a truthy/falsy value with the injected value
	// * action: Block to execute if condition is truthy, with the injected value
	// Returns:
	// * the value of the evaluated action block if condition is truthy or the original injected value if condition is falsy
	"when": { // **
		Argsn: 3,
		Doc:   "Conditionally executes an action block if a condition block evaluates to true, injecting the same value into both blocks and returning the value of the action block if the condition is truthy or the original value otherwise.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			// function accepts 3 args: arg0 is the value to inject, arg1 is the condition block, arg2 is the action block

			// Check if the second argument is a block (condition block)
			switch condBlock := arg1.(type) {
			case env.Block:
				// Check if the third argument is a block (action block)
				switch actionBlock := arg2.(type) {
				case env.Block:
					// Store current series
					ser := ps.Ser

					// Set series to condition block and evaluate it with the value injected
					ps.Ser = condBlock.Series
					EvalBlockInjMultiDialect(ps, arg0, true)
					if ps.ErrorFlag {
						ps.Ser = ser
						return ps.Res
					}

					// Check if the result is truthy
					if util.IsTruthy(ps.Res) {
						// Set series to action block and evaluate it with the value injected
						ps.Ser = actionBlock.Series
						EvalBlockInjMultiDialect(ps, arg0, true)
						if ps.ErrorFlag {
							ps.Ser = ser
							return ps.Res
						}
					} else {
						ps.Res = arg0
					}

					// Restore original series
					ps.Ser = ser

					// Return the original value regardless of the evaluation result
					return ps.Res
				default:
					// If the third argument is not a block, return an error
					ps.FailureFlag = true
					return MakeArgError(ps, 3, []env.Type{env.BlockType}, "when")
				}
			default:
				// If the second argument is not a block, return an error
				ps.FailureFlag = true
				return MakeArgError(ps, 2, []env.Type{env.BlockType}, "when")
			}
		},
	},

	// Tests:
	// equal  { x: does { ^if true { 222 } 555 } x } 222
	// equal  { x: does { ^if false { 333 } 444 } x } 333
	// Args:
	// * condition: Value to evaluate for truthiness
	// * block: Block of code to execute and return from if condition is truthy
	// Returns:
	// * result of the block if condition is truthy (with return flag set), 0 otherwise
	"^if": { // **
		Argsn: 2,
		Doc:   "Conditional that sets the return flag when true, allowing early return from a function when the condition is a boolean.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			// Check if the first argument is a boolean
			switch cond := arg0.(type) {
			case env.Boolean:
				switch bloc := arg1.(type) {
				case env.Block:
					if cond.Value {
						ser := ps.Ser
						ps.Ser = bloc.Series
						EvalBlockInj(ps, arg0, true)
						if ps.ErrorFlag {
							ps.Ser = ser
							return ps.Res
						}
						ps.Ser = ser
						ps.ReturnFlag = true
						return ps.Res
					}
					return *env.NewInteger(0)
				default:
					return MakeArgError(ps, 2, []env.Type{env.BlockType}, "^if")
				}
			default:
				// if it's not a boolean we return error
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.BooleanType}, "^if")
			}
		},
	},

	// Tests:
	// equal  { either true { 222 } { 333 } } 222
	// equal  { either false { 222 } { 333 } } 333
	// error  { either 1 { 222 } { 333 } }
	// error  { either 0 { 222 } { 333 } }
	// Args:
	// * condition: Boolean value determining which block to execute
	// * true_block: Block or value to return if condition is true
	// * false_block: Block or value to return if condition is false
	// Returns:
	// * result of executing the true_block if condition is true, otherwise result of false_block
	"either": { // **
		Argsn: 3,
		Doc:   "Executes one of two blocks based on a boolean condition, similar to if/else in other languages.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			// Check if the first argument is a boolean
			switch cond := arg0.(type) {
			case env.Boolean:
				switch bloc1 := arg1.(type) {
				case env.Block:
					switch bloc2 := arg2.(type) {
					case env.Block:
						ser := ps.Ser
						if cond.Value {
							ps.Ser = bloc1.Series
							ps.Ser.Reset()
						} else {
							ps.Ser = bloc2.Series
							ps.Ser.Reset()
						}
						EvalBlockInjMultiDialect(ps, arg0, true)
						ps.Ser = ser
						return ps.Res
					default:
						return MakeArgError(ps, 3, []env.Type{env.BlockType}, "either")
					}
				case env.Object:
					switch bloc2 := arg2.(type) {
					case env.Object: // If true value is not block then also false value will be treated as literal
						if cond.Value {
							return bloc1
						} else {
							return bloc2
						}
					default:
						return MakeBuiltinError(ps, "Third argument must be Object Type.", "either")
					}
				default:
					return MakeBuiltinError(ps, "Second argument must be Block or Object Type.", "either")
				}
			default:
				// If it's not a boolean, return an error
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.BooleanType}, "either")
			}
		},
	},
	// Tests:
	// equal  { choose true [ 1 -1 ] } 1
	// equal  { choose false [ 1 -1 ] } -1
	// equal  { choose 10 > 5 [ "yes" "no" ] } "yes"
	// equal  { choose 3 < 2 [ "yes" "no" ] } "no"
	// error  { choose 1 [ 1 -1 ] }
	// error  { choose true [ 1 ] }
	// Args:
	// * condition: Boolean value determining which value to return
	// * values: Block containing exactly two values
	// Returns:
	// * first value if condition is true, second value if condition is false
	"choose": { // **
		Argsn: 2,
		Doc:   "Returns the first or second value from a block of two values based on a boolean condition.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			// Check if the first argument is a boolean
			switch cond := arg0.(type) {
			case env.Boolean:
				// Check if the second argument is a block
				switch bloc := arg1.(type) {
				case env.Block:
					// Check if the block has exactly 2 values
					if bloc.Series.Len() != 2 {
						ps.FailureFlag = true
						return MakeBuiltinError(ps, "Block must contain exactly 2 values.", "choose")
					}
					// Return first value if true, second if false
					if cond.Value {
						return bloc.Series.Get(0)
					}
					return bloc.Series.Get(1)
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 2, []env.Type{env.BlockType}, "choose")
				}
			default:
				// If it's not a boolean, return an error
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.BooleanType}, "choose")
			}
		},
	},

	// Tests:
	// equal  { switch 101 { 101 { 111 } 202 { 222 } } } 111
	// equal  { switch 202 { 101 { 111 } 202 { 222 } } } 222
	// Args:
	// * value: Value to match against case values
	// * cases: Block containing case values and corresponding handler blocks
	// Returns:
	// * result of executing the matching handler block, or the original value if no match
	"switch": { // **
		Argsn:         2,
		Doc:           "Pattern matching construct that executes a block of code corresponding to the first matching case value.",
		AcceptFailure: true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch bloc := arg1.(type) {
			case env.Block:

				var code env.Object

				any_found := false

				for i := 0; i < bloc.Series.Len(); i += 2 {

					if i > bloc.Series.Len()-2 {
						return MakeBuiltinError(ps, "Switch block malformed.", "switch")
					}

					ev := bloc.Series.Get(i)
					if arg0.GetKind() == ev.GetKind() && arg0.Inspect(*ps.Idx) == ev.Inspect(*ps.Idx) {
						any_found = true
						code = bloc.Series.Get(i + 1)
					}
					if ev.Type() == env.VoidType {
						if !any_found {
							code = bloc.Series.Get(i + 1)
							any_found = true
						}
					}
				}
				if any_found {
					switch cc := code.(type) {
					case env.Block:
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
						return ps.Res
					default:
						// if it's not a block we return error for now
						ps.FailureFlag = true
						return MakeBuiltinError(ps, "Malformed switch block.", "switch")
					}
				}
				return arg0
			default:
				// if it's not a block we return error for now
				ps.FailureFlag = true
				return MakeArgError(ps, 2, []env.Type{env.BlockType}, "switch")
			}
		},
	},

	// Tests:
	// equal  { cases 0 { { 1 > 0 } { + 100 } { 2 > 1 } { + 1000 } } } 1100
	// equal  { cases 0 { { 1 > 0 } { + 100 } { 2 < 1 } { + 1000 } } } 100
	// equal  { cases 0 { { 1 < 0 } { + 100 } { 2 > 1 } { + 1000 } } } 1000
	// equal  { cases 0 { { 1 < 0 } { + 100 } { 2 < 1 } { + 1000 } } } 0
	// equal  { cases 1 { { 1 > 0 } { + 100 } { 2 < 1 } { + 1000 } _ { * 3 } } } 101
	// equal  { cases 1 { { 1 < 0 } { + 100 } { 2 > 1 } { + 1000 } _ { * 3 } } } 1001
	// equal  { cases 1 { { 1 < 0 } { + 100 } { 2 < 1 } { + 1000 } _ { * 3 } } } 3
	// Args:
	// * initial: Initial value to be transformed by matching case blocks
	// * cases: Block containing condition blocks and corresponding transformation blocks
	// Returns:
	// * cumulative result after applying all matching transformation blocks to the initial value
	"cases": { // ** , TODO-FIXME: error handling
		Argsn: 2,
		Doc:   "Evaluates multiple condition-action pairs and applies all matching actions cumulatively to the initial value.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			// function accepts 2 args. arg0 is a "boolean" value, arg1 is a block of code
			// we set bloc to block of code
			// (we don't have boolean type yet, because it's not cruicial to important part of design, neither is truthiness ... this will be decided later
			// on more operational level

			// we switch on the type of second argument, so far it should be block (later we could accept native and function)
			switch bloc := arg1.(type) {
			case env.Block:
				// TODO --- istruthy must return error if it's not possible to
				// calculate truthiness and we must here raise failure
				// we switch on type of arg0
				// if it's integer, all except 0 is true
				// if it's string, all except empty string is true
				// we don't care for other types at this stage
				ser := ps.Ser

				cumul := arg0

				foundany := false
				for {

					doblk := false
					cond_ := bloc.Series.Pop()
					blk := bloc.Series.Pop().(env.Block)

					switch cond := cond_.(type) {
					case env.Block:
						ps.Ser = cond.Series
						// we eval the block (current context / scope stays the same as it was in parent block)
						// Inj means we inject the condition value into the block, because it costs us very little. we could do "if name { .print }"
						EvalBlock(ps)
						if ps.ErrorFlag {
							ps.Ser = ser
							return ps.Res
						}
						// we set temporary series back to current program state
						if util.IsTruthy(ps.Res) {
							doblk = true
							foundany = true
						}
					case env.Void:
						if !foundany {
							doblk = true
						}
					default:
						return MakeBuiltinError(ps, "Invalid block series type.", "cases")
					}
					// we set ProgramStates series to series ob the block
					if doblk {
						ps.Ser = blk.Series
						// we eval the block (current context / scope stays the same as it was in parent block)
						// Inj means we inject the condition value into the block, because it costs us very little. we could do "if name { .print }"
						EvalBlockInjMultiDialect(ps, cumul, true)
						if ps.ErrorFlag {
							ps.Ser = ser
							return ps.Res
						}
						cumul = ps.Res
					}
					if bloc.Series.AtLast() {
						break
					}
				}
				ps.Ser = ser
				// we return the last return value (the return value of executing the block) "a: if 1 { 100 }" a becomes 100,
				// in future we will also handle the "else" case, but we have to decide
				return cumul
			default:
				// if it's not a block we return error for now
				ps.FailureFlag = true
				return MakeArgError(ps, 2, []env.Type{env.BlockType}, "cases")
			}
		},
	},
}
