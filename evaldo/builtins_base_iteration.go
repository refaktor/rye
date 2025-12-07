package evaldo

import (
	"fmt"
	"strings"

	"github.com/refaktor/rye/env"
	"github.com/refaktor/rye/util"
	// JM 20230825	"github.com/refaktor/rye/term"
)

var builtins_iteration = map[string]*env.Builtin{

	//
	// ##### Iteration ##### "Functions for iterating over collections and executing code repeatedly"
	//
	// Tests:
	// stdout { 3 .loop { prns "x" } } "x x x "
	// equal  { 3 .loop { + 1 } } 4
	// ; equal  { 3 .loop { } } 3  ; TODO should pass the value
	// Args:
	// * count: Integer number of iterations to perform
	// * block: Block of code to execute on each iteration
	// Returns:
	// * result of the last block execution
	"loop": {
		Argsn: 2,
		Doc:   "Executes a block of code, builtin, or function a specified number of times, injecting the current iteration number (starting from 1).",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			// Type checking for arguments
			count, ok := arg0.(env.Integer)
			if !ok {
				return MakeArgError(ps, 1, []env.Type{env.IntegerType}, "loop")
			}

			// Pre-allocate a single Integer object to reuse
			iterObj := env.Integer{Value: 0}

			switch block := arg1.(type) {
			case env.Block:
				// Save original series
				ser := ps.Ser
				ps.Ser = block.Series

				// Main loop
				for i := int64(0); i < count.Value; i++ {
					// Evaluate the block with the current iteration number
					EvalBlockInjMultiDialect(ps, iterObj, true)
					if ps.ErrorFlag {
						MaybeDisplayFailureOrError(ps, ps.Idx, "loop")
						ps.Ser = ser
						return ps.Res
					}

					iterObj.Value++
					// Reset series position for next iteration
					ps.Ser.Reset()
				}

				// Restore original series
				ps.Ser = ser
				return ps.Res
			case env.Builtin:
				// Main loop
				for i := int64(0); i < count.Value; i++ {
					// Call the builtin with the current iteration number
					ps.Res = DirectlyCallBuiltin(ps, block, iterObj, nil)
					if ps.ErrorFlag {
						return ps.Res
					}
					iterObj.Value++
				}
				return ps.Res
			case env.Function:
				// Main loop
				for i := int64(0); i < count.Value; i++ {
					// Call the function with the current iteration number
					CallFunctionArgsN(block, ps, ps.Ctx, iterObj)
					if ps.ErrorFlag {
						return ps.Res
					}
					iterObj.Value++
				}
				return ps.Res
			default:
				return MakeArgError(ps, 2, []env.Type{env.BlockType, env.BuiltinType, env.FunctionType}, "loop")
			}
		},
	},

	// Tests:
	// equal { produce 5 0 { + 3 } } 15
	// equal { produce 3 ">" { ++ "x>" } } ">x>x>x>"
	// equal { produce 3 { } { .concat "x" } } { "x" "x" "x" }
	// equal { produce 3 { } { ::x .concat length? x } } { 0 1 2 }
	// equal { produce 5 { 2 } { ::acc .last ::x * x |concat* acc } } { 2 4 16 256 65536 4294967296 }
	// Args:
	// * count: Integer number of iterations to perform
	// * initial: Initial value to inject into the first block execution
	// * block: Block of code to execute on each iteration
	// Returns:
	// * result of the last block execution
	"produce": {
		Argsn: 3,
		Doc:   "Executes a block of code a specified number of times, passing the result of each execution to the next iteration.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch cond := arg0.(type) {
			case env.Integer:
				switch bloc := arg2.(type) {
				case env.Block:
					acc := arg1
					ser := ps.Ser
					ps.Ser = bloc.Series
					ps.Res = arg1
					for i := 0; int64(i) < cond.Value; i++ {
						EvalBlockInjMultiDialect(ps, acc, true)
						MaybeDisplayFailureOrError(ps, ps.Idx, "produce")
						if ps.ErrorFlag {
							ps.Ser = ser
							return ps.Res
						}
						ps.Ser.Reset()
						acc = ps.Res
					}
					ps.Ser = ser
					return ps.Res
				default:
					return MakeArgError(ps, 3, []env.Type{env.BlockType}, "produce")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType}, "produce")
			}
		},
	},

	// Tests:
	// equal { var 'x 0 , produce\while { x < 100 } 1 { * 2 ::x } } 64
	// stdout { var 'x 0 , produce\while { x < 100 } 1 { * 2 ::x .prns } } "2 4 8 16 32 64 128 "
	// Args:
	// * condition: Block that evaluates to a boolean to determine when to stop iterating
	// * initial: Initial value to inject into the first block execution
	// * block: Block of code to execute on each iteration
	// Returns:
	// * result of the last block execution before the condition became false
	"produce\\while": {
		Argsn: 3,
		Doc:   "Executes a block of code repeatedly while a condition is true, passing the result of each execution to the next iteration.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch cond := arg0.(type) {
			case env.Block:
				switch bloc := arg2.(type) {
				case env.Block:
					acc := arg1
					last := arg1
					ser := ps.Ser
					for {
						ps.Ser = cond.Series
						EvalBlockInjMultiDialect(ps, acc, true)
						MaybeDisplayFailureOrError(ps, ps.Idx, "produce\\while")
						if ps.ErrorFlag {
							ps.Ser = ser
							return ps.Res
						}
						if !util.IsTruthy(ps.Res) {
							ps.Ser.Reset()
							ps.Ser = ser
							return last
						} else {
							last = acc
						}
						ps.Ser.Reset()
						ps.Ser = bloc.Series
						EvalBlockInjMultiDialect(ps, acc, true)
						MaybeDisplayFailureOrError(ps, ps.Idx, "produce\\while")
						if ps.ErrorFlag {
							ps.Ser = ser
							return ps.Res
						}
						ps.Ser = ser
						ps.Ser.Reset()
						acc = ps.Res
					}
				default:
					return MakeArgError(ps, 3, []env.Type{env.BlockType}, "produce\\while")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType}, "produce\\while")
			}
		},
	},

	// Tests:
	//  equal { produce\ 5 1 'acc { * acc , + 1 } } 1  ; Look at what we were trying to do here
	// Args:
	// * count: Integer number of iterations to perform
	// * initial: Initial value for the accumulator
	// * accumulator: Word to store the accumulator value
	// * block: Block of code to execute on each iteration
	// Returns:
	// * final value of the accumulator word after all iterations
	"produce\\": {
		Argsn: 4,
		Doc:   "Executes a block of code a specified number of times, updating an accumulator word with each iteration result.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch cond := arg0.(type) {
			case env.Integer:
				switch bloc := arg3.(type) {
				case env.Block:
					switch accu := arg2.(type) {
					case env.Word:
						acc := arg1
						ps.Ctx.Mod(accu.Index, acc)
						ser := ps.Ser
						ps.Ser = bloc.Series
						for i := 0; int64(i) < cond.Value; i++ {
							EvalBlockInjMultiDialect(ps, acc, true)
							MaybeDisplayFailureOrError(ps, ps.Idx, "produce\\")
							if ps.ErrorFlag {
								ps.Ser = ser
								return ps.Res
							}
							ps.Ser.Reset()
							acc = ps.Res
						}
						ps.Ser = ser
						val, _ := ps.Ctx.Get(accu.Index)
						return val
					default:
						return MakeArgError(ps, 3, []env.Type{env.WordType}, "produce\\")
					}
				default:
					return MakeArgError(ps, 4, []env.Type{env.BlockType}, "produce\\")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType}, "produce\\")
			}
		},
	},

	// Tests:
	//  stdout { forever { "once" .prn .return } } "once"
	//  equal { forever { "once" .return } } "once"
	// Args:
	// * block: Block of code to execute repeatedly
	// Returns:
	// * result of the block when .return is called
	"forever": { // **
		Argsn: 1,
		Doc:   "Executes a block of code repeatedly until .return is called within the block.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch bloc := arg0.(type) {
			case env.Block:
				ser := ps.Ser
				ps.Ser = bloc.Series
				for i := 0; i == i; i++ {
					// Check for interrupt signal (Ctrl+C, Ctrl+Z)
					if ps.InterruptFlag {
						ps.InterruptFlag = false // Reset the flag
						ps.Ser = ser
						return *env.NewError("Operation interrupted by user")
					}
					EvalBlockInjMultiDialect(ps, env.NewInteger(int64(i)), true)
					MaybeDisplayFailureOrError(ps, ps.Idx, "forever")
					if ps.ErrorFlag {
						ps.Ser = ser
						return ps.Res
					}
					if ps.ReturnFlag {
						ps.ReturnFlag = false
						break
					}
					ps.Ser.Reset()
				}
				ps.Ser = ser
				return ps.Res
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.BlockType}, "forever")
			}
		},
	},
	// Tests:
	//  stdout { forever\with 1 { .prn .return } } "1"
	//  equal { var 'x 0 , forever\with 1 { + 5 ::x , if x > 5 { return x } } } 6
	// Args:
	// * value: Value to inject into the block on each iteration
	// * block: Block of code to execute repeatedly
	// Returns:
	// * result of the block when .return is called
	"forever\\with": { // **
		Argsn: 2,
		Doc:   "Accepts a value and a block, and executes the block repeatedly with the value until .return is called.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch bloc := arg1.(type) {
			case env.Block:
				ser := ps.Ser
				ps.Ser = bloc.Series
				for {
					// Check for interrupt signal (Ctrl+C, Ctrl+Z)
					if ps.InterruptFlag {
						ps.InterruptFlag = false // Reset the flag
						ps.Ser = ser
						return *env.NewError("Operation interrupted by user")
					}
					EvalBlockInjMultiDialect(ps, arg0, true)
					MaybeDisplayFailureOrError(ps, ps.Idx, "forever\\with")
					if ps.ErrorFlag {
						ps.Ser = ser
						return ps.Res
					}
					if ps.ReturnFlag {
						ps.ReturnFlag = false
						break
					}
					ps.Ser.Reset()
				}
				ps.Ser = ser
				return ps.Res
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.BlockType}, "forever\\with")
			}
		},
	},
	/*
		"for___": { // **
			Argsn: 2,
			Doc:   "Iterates over each value in a collection, executing a block of code for each value.",
			Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
				switch block := arg0.(type) {
				case env.String:
					switch code := arg1.(type) {
					case env.Block:
						ser := ps.Ser
						ps.Ser = code.Series
						for _, ch := range block.Value {
							EvalBlockInjMultiDialect(ps, *env.NewString(string(ch)), true)
						MaybeDisplayFailureOrError(ps, ps.Idx, "forever\\with")
							if ps.ErrorFlag || ps.ReturnFlag {
								return ps.Res
							}
							ps.Ser.Reset()
						}
						ps.Ser = ser
						return ps.Res
					default:
						return MakeArgError(ps, 2, []env.Type{env.BlockType}, "for")
					}
				case env.Block:
					switch code := arg1.(type) {
					case env.Block:
						ser := ps.Ser
						ps.Ser = code.Series
						for i := 0; i < block.Series.Len(); i++ {
							EvalBlockInjMultiDialect(ps, block.Series.Get(i), true)
							if ps.ErrorFlag || ps.ReturnFlag {
								return ps.Res
							}
							ps.Ser.Reset()
						}
						ps.Ser = ser
						return ps.Res
					default:
						return MakeArgError(ps, 2, []env.Type{env.BlockType}, "for")
					}
				case env.List:
					switch code := arg1.(type) {
					case env.Block:
						ser := ps.Ser
						ps.Ser = code.Series
						for i := 0; i < len(block.Data); i++ {
							EvalBlockInjMultiDialect(ps, env.ToRyeValue(block.Data[i]), true)
							if ps.ErrorFlag || ps.ReturnFlag {
								return ps.Res
							}
							ps.Ser.Reset()
						}
						ps.Ser = ser
						return ps.Res
					default:
						return MakeArgError(ps, 2, []env.Type{env.BlockType}, "for")
					}
				case env.Table:
					switch code := arg1.(type) {
					case env.Block:
						ser := ps.Ser
						ps.Ser = code.Series
						for i := 0; i < len(block.Rows); i++ {
							row := block.Rows[i]
							row.Uplink = &block
							EvalBlockInjMultiDialect(ps, row, true)
							if ps.ErrorFlag || ps.ReturnFlag {
								return ps.Res
							}
							ps.Ser.Reset()
						}
						ps.Ser = ser
						return ps.Res
					default:
						return MakeArgError(ps, 2, []env.Type{env.BlockType}, "for")
					}
				default:
					return MakeArgError(ps, 1, []env.Type{env.StringType, env.BlockType, env.TableType}, "for")
				}
			},
		},
	*/
	// Tests:
	// stdout { for { 1 2 3 } { prns "x" } } "x x x "
	// stdout { { "a" "b" "c" } .for { .prns } } "a b c "
	// Args:
	// * collection: Collection (Block, List, String, or Table) to iterate over
	// * code: Block, Builtin, or Function to execute for each value, injecting the value
	// Returns:
	// * result of the last code execution
	"for": { // **
		Argsn: 2,
		Doc:   "Accepts a block of values and a block of code, builtin, or function, does the code for each of the values, injecting them.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch block := arg0.(type) {
			case env.Collection:
				switch code := arg1.(type) {
				case env.Block:
					ser := ps.Ser
					ps.Ser = code.Series
					for i := 0; i < block.Length(); i++ {
						EvalBlockInjMultiDialect(ps, block.Get(i), true)
						MaybeDisplayFailureOrError(ps, ps.Idx, "for")
						if ps.ErrorFlag || ps.ReturnFlag {
							ps.Ser = ser
							return ps.Res
						}
						ps.Ser.Reset()
					}
					ps.Ser = ser
					return ps.Res
				case env.Builtin:
					for i := 0; i < block.Length(); i++ {
						ps.Res = DirectlyCallBuiltin(ps, code, block.Get(i), nil)
						if ps.ErrorFlag || ps.ReturnFlag {
							return ps.Res
						}
					}
					return ps.Res
				case env.Function:
					for i := 0; i < block.Length(); i++ {
						CallFunctionArgsN(code, ps, ps.Ctx, block.Get(i))
						if ps.ErrorFlag || ps.ReturnFlag {
							return ps.Res
						}
					}
					return ps.Res
				default:
					return MakeArgError(ps, 2, []env.Type{env.BlockType, env.BuiltinType, env.FunctionType}, "for")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType, env.BlockType, env.TableType}, "for")
			}
		},
	},

	// Tests:
	// stdout { { "a" "b" "c" } .for\pos 'i { i .prns , .prns } } "1 a 2 b 3 c "
	// Args:
	// * collection: Collection to iterate over
	// * word: Word to store the current position (1-based index)
	// * code: Block of code to execute for each value
	// Returns:
	// * result of the last block execution
	"for\\pos": { // *TODO -- deduplicate map\pos and map\idx
		Argsn: 3,
		Doc:   "Iterates over a collection with position tracking (1-based), setting a word to the current position and injecting each value.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch list := arg0.(type) {
			case env.Collection:
				switch accu := arg1.(type) {
				case env.Word:
					switch code := arg2.(type) {
					case env.Block:
						l := list.Length()
						ser := ps.Ser
						ps.Ser = code.Series
						for i := 0; i < l; i++ {
							ps.Ctx.Mod(accu.Index, *env.NewInteger(int64(i + 1)))
							EvalBlockInjMultiDialect(ps, list.Get(i), true)
							MaybeDisplayFailureOrError(ps, ps.Idx, "for\\pos")
							if ps.ErrorFlag || ps.ReturnFlag {
								ps.Ser = ser
								return ps.Res
							}
							ps.Ser.Reset()
						}
						ps.Ser = ser
						return ps.Res
					default:
						return MakeArgError(ps, 3, []env.Type{env.BlockType}, "for\\pos")
					}
				default:
					return MakeArgError(ps, 2, []env.Type{env.WordType}, "for\\pos")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType, env.ListType, env.StringType}, "for\\pos")
			}
		},
	},

	// Tests:
	// stdout { { "a" "b" "c" } .for\idx 'i { i .prns , .prns } } "0 a 1 b 2 c "
	// Args:
	// * collection: Collection to iterate over
	// * word: Word to store the current index (0-based)
	// * code: Block of code to execute for each value
	// Returns:
	// * result of the last block execution
	"for\\idx": { // *TODO -- deduplicate map\pos and map\idx
		Argsn: 3,
		Doc:   "Iterates over a collection with index tracking (0-based), setting a word to the current index and injecting each value.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch list := arg0.(type) {
			case env.Collection:
				switch accu := arg1.(type) {
				case env.Word:
					switch code := arg2.(type) {
					case env.Block:
						l := list.Length()
						ser := ps.Ser
						ps.Ser = code.Series
						for i := 0; i < l; i++ {
							ps.Ctx.Mod(accu.Index, *env.NewInteger(int64(i)))
							EvalBlockInjMultiDialect(ps, list.Get(i), true)
							MaybeDisplayFailureOrError(ps, ps.Idx, "for\\idx")
							if ps.ErrorFlag || ps.ReturnFlag {
								ps.Ser = ser
								return ps.Res
							}
							ps.Ser.Reset()
						}
						ps.Ser = ser
						return ps.Res
					default:
						return MakeArgError(ps, 3, []env.Type{env.BlockType}, "for\\idx")
					}
				default:
					return MakeArgError(ps, 2, []env.Type{env.WordType}, "for\\idx")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType, env.ListType, env.StringType}, "for\\idx")
			}
		},
	},

	// Tests:
	//  stdout { walk { 1 2 3 } { .prns .rest } } "1 2 3  2 3  3  "
	//  equal { var 'x 0 , walk { 1 2 3 } { ::b .first + x ::x , b .rest } x } 6
	// Args:
	// * block: Block to walk through
	// * code: Block of code to execute, which should return a modified block
	// Returns:
	// * result of the last block execution
	"walk": { // **
		Argsn: 2,
		Doc:   "Walks through a block, executing code that can modify the block on each iteration (useful for custom iteration patterns).",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch block := arg0.(type) {
			case env.Block:
				switch code := arg1.(type) {
				case env.Block:
					ser := ps.Ser
					ps.Ser = code.Series

					for block.Series.GetPos() < block.Series.Len() {
						EvalBlockInjMultiDialect(ps, block, true)
						MaybeDisplayFailureOrError(ps, ps.Idx, "walk")
						if ps.ErrorFlag || ps.ReturnFlag {
							ps.Ser = ser
							return ps.Res
						}
						block1, ok := ps.Res.(env.Block) // TODO ... switch and throw error if not block
						if ok {
							block = block1
						} else {
							fmt.Println("ERROR 1231241")
						}
						ps.Ser.Reset()
					}
					ps.Ser = ser
					return ps.Res
				default:
					return MakeArgError(ps, 2, []env.Type{env.BlockType}, "walk")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType}, "walk")
			}
		},
	},

	// Higher order functions
	// Tests:
	//  equal { purge { 1 2 3 } { .is-even } } { 1 3 }
	//  equal { purge { } { .is-even } } { }
	//  equal { purge list { 1 2 3 } { .is-even } } list { 1 3 }
	//  equal { purge list { } { .is-even } } list { }
	//  equal { purge "1234" { .probe .to-integer .is-even } } { "1" "3" }
	//  equal { purge "" { .to-integer .is-even } } { }
	// Args:
	// * series: Block, List, String, or Table to purge from
	// * code: Block of code that returns true for values to remove
	// Returns:
	// * modified series with matching values removed
	"purge": { // TODO ... doesn't fully work
		Argsn: 2,
		Doc:   "Purges values from a series based on return of a injected code block.",
		Pure:  false,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch block := arg0.(type) {
			case env.Block:
				switch code := arg1.(type) {
				case env.Block:
					ser := ps.Ser
					ps.Ser = code.Series
					for i := 0; i < block.Series.Len(); i++ {
						EvalBlockInjMultiDialect(ps, block.Series.Get(i), true)
						MaybeDisplayFailureOrError(ps, ps.Idx, "purge")
						if ps.ErrorFlag {
							ps.Ser = ser
							return ps.Res
						}
						if util.IsTruthy(ps.Res) {
							block.Series.S = append(block.Series.S[:i], block.Series.S[i+1:]...)
							i--
						}
						ps.Ser.Reset()
					}
					ps.Ser = ser
					return block
				default:
					return MakeArgError(ps, 2, []env.Type{env.BlockType}, "purge")
				}
			case env.List:
				switch code := arg1.(type) {
				case env.Block:
					ser := ps.Ser
					ps.Ser = code.Series
					for i := 0; i < len(block.Data); i++ {
						EvalBlockInjMultiDialect(ps, env.ToRyeValue(block.Data[i]), true)
						MaybeDisplayFailureOrError(ps, ps.Idx, "purge")
						if ps.ErrorFlag {
							ps.Ser = ser
							return ps.Res
						}
						if util.IsTruthy(ps.Res) {
							block.Data = append(block.Data[:i], block.Data[i+1:]...)
							i--
						}
						ps.Ser.Reset()
					}
					ps.Ser = ser
					return block
				default:
					return MakeArgError(ps, 2, []env.Type{env.BlockType}, "purge")
				}
			case env.String:
				switch code := arg1.(type) {
				case env.Block:
					input := []rune(block.Value)
					var newl []env.Object
					ser := ps.Ser
					ps.Ser = code.Series
					for i := 0; i < len(input); i++ {
						EvalBlockInjMultiDialect(ps, *env.NewString(string(input[i])), true)
						MaybeDisplayFailureOrError(ps, ps.Idx, "purge")
						if ps.ErrorFlag {
							ps.Ser = ser
							return ps.Res
						}
						if !util.IsTruthy(ps.Res) {
							newl = append(newl, *env.NewString(string(input[i])))
						}
						ps.Ser.Reset()
					}
					ps.Ser = ser
					return *env.NewBlock(*env.NewTSeries(newl))
				default:
					return MakeArgError(ps, 2, []env.Type{env.BlockType}, "purge")
				}
			case env.Table:
				switch code := arg1.(type) {
				case env.Block:
					ser := ps.Ser
					ps.Ser = code.Series
					for i := 0; i < len(block.Rows); i++ {
						EvalBlockInjMultiDialect(ps, block.Rows[i], true)
						MaybeDisplayFailureOrError(ps, ps.Idx, "purge")
						if ps.ErrorFlag {
							ps.Ser = ser
							return ps.Res
						}
						if util.IsTruthy(ps.Res) {
							block.Rows = append(block.Rows[:i], block.Rows[i+1:]...)
							i--
						}
						ps.Ser.Reset()
					}
					ps.Ser = ser
					return block
				default:
					return MakeArgError(ps, 2, []env.Type{env.BlockType}, "purge")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType, env.ListType, env.StringType, env.TableType}, "purge")
			}
		},
	},

	// Tests:
	//  equal { { 1 2 3 } :x purge! { .is-even } 'x , x } { 1 3 }
	// Args:
	// * code: Block of code that returns true for values to remove
	// * word: Word referring to a Block to purge from (modified in-place)
	// Returns:
	// * Block containing the purged values
	"purge!": { // TODO ... doesn't fully work
		Argsn: 2,
		Doc:   "Purges values from a block stored in a word (modifying it in-place), returning the purged values.",
		Pure:  false,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch wrd := arg1.(type) {
			case env.Word:
				val, found, ctx := ps.Ctx.Get2(wrd.Index)
				if found {
					switch block := val.(type) {
					case env.Block:
						switch code := arg0.(type) {
						case env.Block:
							ser := ps.Ser
							ps.Ser = code.Series
							purged := make([]env.Object, 0)
							for i := 0; i < block.Series.Len(); i++ {
								EvalBlockInjMultiDialect(ps, block.Series.Get(i), true)
								MaybeDisplayFailureOrError(ps, ps.Idx, "purge!")
								if ps.ErrorFlag {
									ps.Ser = ser
									return ps.Res
								}
								if util.IsTruthy(ps.Res) {
									purged = append(purged, block.Series.S[i])
									block.Series.S = append(block.Series.S[:i], block.Series.S[i+1:]...)
									i--
								}
								ps.Ser.Reset()
							}
							ps.Ser = ser
							ctx.Mod(wrd.Index, block)
							return env.NewBlock(*env.NewTSeries(purged))
						default:
							return MakeArgError(ps, 1, []env.Type{env.BlockType}, "purge!")
						}
					default:
						return MakeBuiltinError(ps, "Context value should be block type.", "purge!")
					}
				} else {
					return MakeBuiltinError(ps, "Word not found in context.", "purge!")
				}
			default:
				return MakeArgError(ps, 2, []env.Type{env.WordType}, "purge!")
			}
		},
	},

	// map should at the end map over block, raw-map, etc ...
	// it should accept a block of code, a function and a builtin
	// it should use injected block so it doesn't need a variable defined like map [ 1 2 3 ] x [ add a 100 ]
	// map [ 1 2 3 ] { .add 3 }
	// Tests:
	//  equal { map { 1 2 3 } { + 1 } } { 2 3 4 }
	//  equal { map { } { + 1 } } { }
	//  equal { map { "aaa" "bb" "c" } { .length? } } { 3 2 1 }
	//  equal { map list { "aaa" "bb" "c" } { .length? } } list { 3 2 1 }
	//  equal { map list { 3 4 5 6 } { .is-multiple-of 3 } } list { 1 0 0 1 }
	//  equal { map list { } { + 1 } } list { }
	//  ; equal { map "abc" { + "-" } .join } "a-b-c-" ; TODO doesn't work, fix join
	//  equal { map "123" { .to-integer } } { 1 2 3 }
	//  equal { map "123" ?to-integer } { 1 2 3 }
	//  equal { map "" { + "-" } } { }
	"map": { // **
		Argsn: 2,
		Doc:   "Maps values of a block to a new block by evaluating a block of code or function.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch list := arg0.(type) {
			case env.Collection:
				switch block := arg1.(type) {
				case env.Block, env.Builtin, env.Function:
					l := list.Length()
					newl := make([]env.Object, l)
					switch block := block.(type) {
					case env.Block:
						ser := ps.Ser
						ps.Ser = block.Series
						for i := 0; i < l; i++ {
							EvalBlockInjMultiDialect(ps, list.Get(i), true)
							MaybeDisplayFailureOrError(ps, ps.Idx, "map")
							if ps.ErrorFlag {
								ps.Ser = ser
								return ps.Res
							}
							newl[i] = ps.Res
							ps.Ser.Reset()
						}
						ps.Ser = ser
					case env.Builtin:
						for i := 0; i < l; i++ {
							newl[i] = DirectlyCallBuiltin(ps, block, list.Get(i), nil)
						}
					case env.Function:
						for i := 0; i < l; i++ {
							CallFunctionArgsN(block, ps, ps.Ctx, list.Get(i))
							if ps.ErrorFlag {
								return ps.Res
							}
							newl[i] = ps.Res
						}
					default:
						return MakeBuiltinError(ps, "Block value should be builtin, block, or function type.", "map")
					}
					return list.MakeNew(newl)
				default:
					return MakeArgError(ps, 2, []env.Type{env.BlockType, env.BuiltinType, env.FunctionType}, "map")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType, env.ListType, env.StringType}, "map")
			}
		},
	},

	// Tests:
	//  equal { map\pos { 1 2 3 } 'i { + i } } { 2 4 6 }
	//  equal { map\pos { } 'i { + i } } { }
	//  equal { map\pos list { 1 2 3 } 'i { + i } } list { 2 4 6 }
	//  equal { map\pos list { } 'i { + i } } list { }
	//  equal { map\pos "abc" 'i { ++ i } } { "a1" "b2" "c3" }
	//  equal { map\pos "" 'i { + i } } { }
	"map\\pos": { // *TODO -- deduplicate map\pos and map\idx
		// TODO --- test { + i } didn't produce / report a valid error, find out why
		Argsn: 3,
		Doc:   "Maps values of a block to a new block by evaluating a block of code.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch list := arg0.(type) {
			case env.Collection:
				switch accu := arg1.(type) {
				case env.Word:
					switch block := arg2.(type) {
					case env.Block:
						l := list.Length()
						newl := make([]env.Object, l)
						ser := ps.Ser
						ps.Ser = block.Series
						for i := 0; i < l; i++ {
							ps.Ctx.Mod(accu.Index, *env.NewInteger(int64(i + 1)))
							EvalBlockInjMultiDialect(ps, list.Get(i), true)
							MaybeDisplayFailureOrError(ps, ps.Idx, "map\\pos")
							if ps.ErrorFlag {
								ps.Ser = ser
								return ps.Res
							}
							newl[i] = ps.Res
							ps.Ser.Reset()
						}
						ps.Ser = ser
						return list.MakeNew(newl)
					default:
						return MakeArgError(ps, 3, []env.Type{env.BlockType}, "map\\pos")
					}
				default:
					return MakeArgError(ps, 2, []env.Type{env.WordType}, "map\\pos")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType, env.ListType, env.StringType}, "map\\pos")
			}
		},
	},

	// Tests:
	// equal { map\idx { 1 2 3 } 'i { + i } } { 1 3 5 }
	// equal { map\idx { } 'i { + i } } { }
	// equal { map\idx list { 1 2 3 } 'i { + i } } list { 1 3 5 }
	// equal { map\idx list { } 'i { + i } } list { }
	// equal { map\idx "abc" 'i { ++ i } } { "a0" "b1" "c2" }
	// equal { map\idx "" 'i { + i } } { }
	"map\\idx": { // TODO -- deduplicate map\idx and map\idx
		Argsn: 3,
		Doc:   "Maps values of a block to a new block by evaluating a block of code.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch list := arg0.(type) {
			case env.Collection:
				switch accu := arg1.(type) {
				case env.Word:
					switch block := arg2.(type) {
					case env.Block:
						l := list.Length()
						newl := make([]env.Object, l)
						ser := ps.Ser
						ps.Ser = block.Series
						for i := 0; i < l; i++ {
							ps.Ctx.Mod(accu.Index, *env.NewInteger(int64(i)))
							EvalBlockInjMultiDialect(ps, list.Get(i), true)
							MaybeDisplayFailureOrError(ps, ps.Idx, "map\\idx")
							if ps.ErrorFlag {
								ps.Ser = ser
								return ps.Res
							}
							newl[i] = ps.Res
							ps.Ser.Reset()
						}
						ps.Ser = ser
						return list.MakeNew(newl)
					default:
						return MakeArgError(ps, 3, []env.Type{env.BlockType}, "map\\idx")
					}
				default:
					return MakeArgError(ps, 2, []env.Type{env.WordType}, "map\\idx")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType, env.ListType, env.StringType}, "map\\idx")
			}
		},
	},
	// Tests:
	//  equal { reduce { 1 2 3 } 'acc { + acc } } 6
	//  equal { reduce list { 1 2 3 } 'acc { + acc } } 6
	//  equal { reduce "abc" 'acc { ++ acc } } "cba"
	//  equal { try { reduce { } 'acc { + acc } } |type? } 'error
	//  equal { try { reduce list { } 'acc { + acc } } |type? } 'error
	//  equal { try { reduce "" 'acc { + acc } } |type? } 'error
	// Args:
	// * collection: Collection to reduce (must not be empty)
	// * accumulator: Word to store the accumulator value
	// * code: Block of code to execute for each value, receiving accumulator as context
	// Returns:
	// * final accumulated value
	"reduce": { // **
		Argsn: 3,
		Doc:   "Reduces values of a block to a single value by evaluating a block of code, using the first element as the initial accumulator.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch list := arg0.(type) {
			case env.Collection:
				l := list.Length()
				if l == 0 {
					return MakeBuiltinError(ps, "Block is empty.", "reduce")
				}
				switch accu := arg1.(type) {
				case env.Word:
					// ps.Ctx.Set(accu.Index)
					switch block := arg2.(type) {
					case env.Block:
						acc := list.Get(0)
						ser := ps.Ser
						ps.Ser = block.Series
						for i := 1; i < l; i++ {
							ps.Ctx.Mod(accu.Index, acc)
							EvalBlockInjMultiDialect(ps, list.Get(i), true)
							MaybeDisplayFailureOrError(ps, ps.Idx, "reduce")
							if ps.ErrorFlag {
								ps.Ser = ser
								return ps.Res
							}
							acc = ps.Res
							ps.Ser.Reset()
						}
						ps.Ser = ser
						return acc
					default:
						return MakeArgError(ps, 3, []env.Type{env.BlockType}, "reduce")
					}
				default:
					return MakeArgError(ps, 2, []env.Type{env.WordType}, "reduce")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType, env.ListType, env.StringType}, "reduce")
			}
		},
	},

	// map should at the end map over block, raw-map, etc ...
	// it should accept a block of code, a function and a builtin
	// it should use injected block so it doesn't need a variable defined like map [ 1 2 3 ] x [ add a 100 ]
	// reduce [ 1 2 3 ] 'acc { + acc }
	// Tests:
	//  equal { fold\do { 1 2 3 } 'acc 1 { + acc } } 7
	//  equal { fold\do { } 'acc 1 { + acc } } 1
	//  equal { fold\do list { 1 2 3 } 'acc 1 { + acc } } 7
	//  equal { fold\do list { } 'acc 1 { + acc } } 1
	//  equal { fold\do "abc" 'acc "123" { ++ acc } } "cba123"
	//  equal { fold\do "" 'acc "123" { + acc } } "123"
	"fold\\do": { // **
		Argsn: 4,
		Doc:   "Reduces values of a block to a new block by evaluating a block of code ...",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch list := arg0.(type) {
			case env.Collection:
				switch accu := arg1.(type) {
				case env.Word:
					// ps.Ctx.Set(accu.Index)
					switch block := arg3.(type) {
					case env.Block:
						l := list.Length()
						acc := arg2
						ser := ps.Ser
						ps.Ser = block.Series
						for i := 0; i < l; i++ {
							ps.Ctx.Mod(accu.Index, acc)
							EvalBlockInjMultiDialect(ps, list.Get(i), true)
							MaybeDisplayFailureOrError(ps, ps.Idx, "fold\\do")
							if ps.ErrorFlag {
								ps.Ser = ser
								return ps.Res
							}
							acc = ps.Res
							ps.Ser.Reset()
						}
						ps.Ser = ser
						return acc
					default:
						return MakeArgError(ps, 4, []env.Type{env.BlockType}, "fold\\do")
					}
				default:
					return MakeArgError(ps, 2, []env.Type{env.WordType}, "fold\\do")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType, env.ListType, env.StringType}, "fold\\do")
			}
		},
	},
	// map should at the end map over block, raw-map, etc ...
	// it should accept a block of code, a function and a builtin
	// it should use injected block so it doesn't need a variable defined like map [ 1 2 3 ] x [ add a 100 ]
	// reduce [ 1 2 3 ] 'acc { + acc }
	// Tests:
	//  equal { fold { 1 2 3 } 1 fn { v acc } { + acc } } 7
	//  equal { fold { } 1 fn { v acc } { + acc } } 1
	//  equal { fold list { 1 2 3 } 1 fn { v acc } { v + acc } } 7
	"fold": { // **
		Argsn: 3,
		Doc:   "Reduces values of a block to a new block by evaluating a block of code ...",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch list := arg0.(type) {
			case env.Collection:
				switch block := arg2.(type) {
				case env.Function:
					l := list.Length()
					acc := arg1
					for i := 0; i < l; i++ {
						var item any
						item = list.Get(i)
						CallFunctionArgsN(block, ps, ps.Ctx, env.ToRyeValue(item), acc) // , env.NewInteger(int64(i)))
						if ps.ErrorFlag {
							return ps.Res
						}
						acc = ps.Res
					}
					return acc
				default:
					return MakeArgError(ps, 4, []env.Type{env.FunctionType}, "fold")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType, env.ListType, env.StringType}, "fold")
			}
		},
	},

	/* This is too specialised and should be removed probably
	"sum-up": { // **
		Argsn: 2,
		Doc:   "Reduces values of a block or list by evaluating a block of code and summing the values.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			var ll []any
			var lo []env.Object
			var llen int
			modeObj := 0
			switch data := arg0.(type) {
			case env.Block:
				lo = data.Series.S
				llen = len(lo)
				modeObj = 2
			case env.List:
				ll = data.Data
				llen = len(ll)
				modeObj = 1
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType, env.ListType}, "sum-up")
			}

			switch block := arg1.(type) {
			case env.Block, env.Builtin:
				acc := *env.NewDecimal(0)
				onlyInts := true
				switch block := block.(type) {
				case env.Block:
					ser := ps.Ser
					ps.Ser = block.Series
					for i := 0; i < llen; i++ {
						var item any
						if modeObj == 1 {
							item = ll[i]
						} else {
							item = lo[i]
						}
						// ps.Ctx.Set(accu.Index, acc)
						ps = EvalBlockInjMultiDialect(ps, env.ToRyeValue(item), true)
						if ps.ErrorFlag {
							return ps.Res
						}
						switch res := ps.Res.(type) {
						case env.Integer:
							acc.Value += float64(res.Value)
						case env.Decimal:
							onlyInts = false
							acc.Value += res.Value
						default:
							return MakeBuiltinError(ps, "Block should return integer or decimal.", "sum-up")
						}
						ps.Ser.Reset()
					}
					ps.Ser = ser
				case env.Builtin:
					for i := 0; i < llen; i++ {
						var item any
						if modeObj == 1 {
							item = ll[i]
						} else {
							item = lo[i]
						}
						res := DirectlyCallBuiltin(ps, block, env.ToRyeValue(item), nil)
						switch res := res.(type) {
						case env.Integer:
							acc.Value += float64(res.Value)
						case env.Decimal:
							onlyInts = false
							acc.Value += res.Value
						default:
							return MakeBuiltinError(ps, "Block should return integer or decimal.", "sum-up")
						}
					}
				default:
					return MakeBuiltinError(ps, "Block type should be Builtin or Block.", "sum-up")
				}
				if onlyInts {
					return *env.NewInteger(int64(acc.Value))
				} else {
					return acc
				}
			default:
				return MakeArgError(ps, 2, []env.Type{env.BlockType, env.BuiltinType}, "sum-up")
			}
		},
	},
	*/

	// Tests:
	//  equal { partition { 1 2 3 4 } { > 2 } } { { 1 2 } { 3 4 } }
	//  equal { partition { "a" "b" 1 "c" "d" } { .is-integer } } { { "a" "b" } { 1 } { "c" "d" } }
	//  equal { partition { "a" "b" 1 "c" "d" } ?is-integer } { { "a" "b" } { 1 } { "c" "d" } }
	//  equal { partition { } { > 2 } } { { } }
	//  equal { partition list { 1 2 3 4 } { > 2 } } list vals { list { 1 2 } list { 3 4 } }
	//  equal { partition list { "a" "b" 1 "c" "d" } ?is-integer } list vals { list { "a" "b" } list { 1 } list { "c" "d" } }
	//  equal { partition list { } { > 2 } } list vals { list { } }
	//  equal { partition "aaabbccc" { , } } list { "aaa" "bb" "ccc" }
	//  equal { partition "" { , } } list { "" }
	//  equal { partition "aaabbccc" ?is-string } list { "aaabbccc" }
	// Args:
	// * series: String, Block, or List to partition
	// * code: Block or Builtin that returns a value for grouping consecutive elements
	// Returns:
	// * collection of partitions (each partition contains consecutive elements with the same result)
	"partition": { // **
		Argsn: 2,
		Doc:   "Partitions a series by evaluating a block of code, grouping consecutive elements that return the same result.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch list := arg0.(type) {
			case env.String:
				switch block := arg1.(type) {
				case env.Block, env.Builtin:
					newl := make([]any, 0)
					var subl strings.Builder
					var prevres env.Object
					switch block := block.(type) {
					case env.Block:
						ser := ps.Ser
						ps.Ser = block.Series
						for _, curval := range list.Value {
							EvalBlockInjMultiDialect(ps, *env.NewString(string(curval)), true)
							MaybeDisplayFailureOrError(ps, ps.Idx, "partition")
							if ps.ErrorFlag {
								ps.Ser = ser
								return ps.Res
							}
							if prevres == nil || ps.Res.Equal(prevres) {
								subl.WriteRune(curval)
							} else {
								newl = append(newl, subl.String())
								subl.Reset()
								subl.WriteRune(curval)
							}
							prevres = ps.Res
							ps.Ser.Reset()
						}
						newl = append(newl, subl.String())
						ps.Ser = ser
					case env.Builtin:
						for _, curval := range list.Value {
							res := DirectlyCallBuiltin(ps, block, env.ToRyeValue(curval), nil)
							if prevres == nil || res.Equal(prevres) {
								subl.WriteRune(curval)
							} else {
								newl = append(newl, subl.String())
							}
						}
						newl = append(newl, subl.String())
					default:
						return MakeBuiltinError(ps, "Block type should be Builtin or Block.", "partition")
					}
					return *env.NewList(newl)
				default:
					return MakeArgError(ps, 2, []env.Type{env.BlockType, env.BuiltinType}, "partition")
				}
			case env.Collection:
				switch block := arg1.(type) {
				case env.Block, env.Builtin:
					l := list.Length()
					newl := make([]env.Object, 0)
					subl := make([]env.Object, 0)
					var prevres env.Object
					switch block := block.(type) {
					case env.Block:
						ser := ps.Ser
						ps.Ser = block.Series
						for i := 0; i < l; i++ {
							curval := list.Get(i)
							EvalBlockInjMultiDialect(ps, curval, true)
							MaybeDisplayFailureOrError(ps, ps.Idx, "partition")
							if ps.ErrorFlag {
								ps.Ser = ser
								return ps.Res
							}
							if prevres == nil || ps.Res.Equal(prevres) {
								subl = append(subl, curval)
							} else {
								newl = append(newl, list.MakeNew(subl))
								//newl = append(newl, *env.NewBlock(*env.NewTSeries(subl)))
								subl = []env.Object{curval}
							}
							prevres = ps.Res
							ps.Ser.Reset()
						}
						newl = append(newl, list.MakeNew(subl))
						// newl = append(newl, *env.NewBlock(*env.NewTSeries(subl)))
						ps.Ser = ser
					case env.Builtin:
						for i := 0; i < l; i++ {
							curval := list.Get(i)
							res := DirectlyCallBuiltin(ps, block, curval, nil)
							if prevres == nil || res.Equal(prevres) {
								subl = append(subl, curval)
							} else {
								newl = append(newl, list.MakeNew(subl))
								//newl = append(newl, *env.NewBlock(*env.NewTSeries(subl)))
								subl = []env.Object{curval}
							}
							prevres = res
						}
						newl = append(newl, list.MakeNew(subl))
						// newl = append(newl, *env.NewBlock(*env.NewTSeries(subl)))
					default:
						return MakeBuiltinError(ps, "Block type should be Builtin or Block.", "partition")
					}
					return list.MakeNew(newl)
				default:
					return MakeArgError(ps, 2, []env.Type{env.BlockType, env.BuiltinType}, "partition")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType, env.ListType, env.StringType}, "partition")
			}
		},
	},

	// Tests:
	//  ; Equality for dicts doesn't yet work consistently
	//  ;equal { { "Anne" "Mitch" "Anya" } .group { .first } } dict vals { "A" list { "Anne" "Anya" } "M" list { "Mitch" } }
	//  ;equal { { "Anne" "Mitch" "Anya" } .group ?first } dict vals { "A" list { "Anne" "Anya" } "M" list { "Mitch" } }
	//  ;equal { { } .group { .first } } dict vals { }
	//  ;equal { { "Anne" "Mitch" "Anya" } .list .group { .first } } dict vals { "A" list { "Anne" "Anya" } "M" list { "Mitch" } }
	//  ;equal { { "Anne" "Mitch" "Anya" } .list .group ?first } dict vals { "A" list { "Anne" "Anya" } "M" list { "Mitch" } }
	//  equal { { } .list .group { .first } } dict vals { }
	//  equal { try { { 1 2 3 4 } .group { .is-even } } |type? } 'error ; TODO keys can only be string currently
	// Args:
	// * series: Block or List to group
	// * code: Block or Builtin that returns a string key for grouping
	// Returns:
	// * Dict with string keys and List values containing grouped elements
	"group": { // **
		Argsn: 2,
		Doc:   "Groups a block or list of values by a key generated from a code block (keys must be strings).",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			var ll []any
			var lo []env.Object
			var llen int
			modeObj := 0
			switch data := arg0.(type) {
			case env.Block:
				lo = data.Series.S
				llen = len(lo)
				modeObj = 2
			case env.List:
				ll = data.Data
				llen = len(ll)
				modeObj = 1
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType, env.ListType}, "group")
			}

			switch block := arg1.(type) {
			case env.Block, env.Builtin:
				newd := make(map[string]any)
				switch block := block.(type) {
				case env.Block:
					ser := ps.Ser
					ps.Ser = block.Series
					for i := 0; i < llen; i++ {
						var curval env.Object
						if modeObj == 1 {
							curval = env.ToRyeValue(ll[i])
						} else {
							curval = lo[i]
						}
						EvalBlockInjMultiDialect(ps, curval, true)
						MaybeDisplayFailureOrError(ps, ps.Idx, "group")
						if ps.ErrorFlag {
							ps.Ser = ser
							return ps.Res
						}
						// TODO !!! -- currently only works if results are strings
						newkeyStr, ok := ps.Res.(env.String)
						if !ok {
							return MakeBuiltinError(ps, "Grouping key should be string.", "group")
						}
						newkey := newkeyStr.Value
						entry, ok := newd[newkey]
						if !ok {
							newd[newkey] = env.NewList(make([]any, 0))
							entry, ok = newd[newkey]
							if !ok {
								return MakeBuiltinError(ps, "Key not found in List.", "group")
							}
						}
						switch ee := entry.(type) { // list in dict is a pointer
						case *env.List:
							ee.Data = append(ee.Data, env.RyeToRaw(curval, ps.Idx))
						default:
							return MakeBuiltinError(ps, "Entry type should be List.", "group")
						}
						ps.Ser.Reset()
					}
					ps.Ser = ser
				case env.Builtin:
					for i := 0; i < llen; i++ {
						var curval env.Object
						if modeObj == 1 {
							curval = env.ToRyeValue(ll[i])
						} else {
							curval = lo[i]
						}
						res := DirectlyCallBuiltin(ps, block, curval, nil)
						// TODO !!! -- currently only works if results are strings
						newkeyStr, ok := res.(env.String)
						if !ok {
							return MakeBuiltinError(ps, "Grouping key should be string.", "group")
						}
						newkey := newkeyStr.Value
						entry, ok := newd[newkey]
						if !ok {
							newd[newkey] = env.NewList(make([]any, 0))
							entry, ok = newd[newkey]
							if !ok {
								return MakeBuiltinError(ps, "Key not found in List.", "group")
							}
						}
						switch ee := entry.(type) { // list in dict is a pointer
						case *env.List:
							ee.Data = append(ee.Data, env.RyeToRaw(curval, ps.Idx))
						default:
							return MakeBuiltinError(ps, "Entry type should be List.", "group")
						}
					}
				default:
					return MakeBuiltinError(ps, "Block must be type of Block or builtin.", "group")
				}
				return *env.NewDict(newd)
			default:
				return MakeArgError(ps, 2, []env.Type{env.BlockType, env.BuiltinType}, "group")
			}
		},
	},

	// filter [ 1 2 3 ] { .add 3 }
	// Tests:
	//  equal { filter { 1 2 3 4 } { .is-even } } { 2 4 }
	//  equal { filter { 1 2 3 4 } ?is-even } { 2 4 }
	//  equal { filter { } { .is-even } } { }
	//  equal { filter list { 1 2 3 4 } { .is-even } } list { 2 4 }
	//  equal { filter list { 1 2 3 4 } ?is-even } list { 2 4 }
	//  ; equal { filter list { } { .is-even } } list { }
	//  ; equal { filter "1234" { .to-integer .is-even } } { "2" "4" }
	//  ; equal { filter "01234" ?to-integer } { "1" "2" "3" "4" }
	//  ; equal { filter "" { .to-integer .is-even } } { }
	"filter": { // **
		Argsn: 2,
		Doc:   "Filters values from a seris based on return of a injected code block.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			var ll []any
			var lo []env.Object
			var ls []rune
			var llen int
			modeObj := 0
			switch data := arg0.(type) {
			case env.String:
				ls = []rune(data.Value)
				llen = len(ls)
				modeObj = 3
			case env.Block:
				lo = data.Series.S
				llen = len(lo)
				modeObj = 2
			case env.List:
				ll = data.Data
				llen = len(ll)
				modeObj = 1
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType, env.ListType, env.StringType}, "filter")
			}

			switch block := arg1.(type) {
			case env.Block, env.Builtin, env.Function:
				var newlo []env.Object
				var newll []any
				switch block := block.(type) {
				case env.Block:
					ser := ps.Ser
					ps.Ser = block.Series
					for i := 0; i < llen; i++ {
						var item any
						if modeObj == 1 {
							item = ll[i]
						} else if modeObj == 2 {
							item = lo[i]
						} else {
							item = env.ToRyeValue(ls[i])
						}
						EvalBlockInjMultiDialect(ps, env.ToRyeValue(item), true)
						MaybeDisplayFailureOrError(ps, ps.Idx, "filter")
						if ps.ErrorFlag {
							ps.Ser = ser
							return ps.Res
						}
						if util.IsTruthy(ps.Res) { // todo -- move these to util or something
							if modeObj == 1 {
								newll = append(newll, ll[i])
							} else if modeObj == 2 {
								newlo = append(newlo, lo[i])
							} else {
								newlo = append(newlo, item.(env.Object))
							}
						}
						ps.Ser.Reset()
					}
					ps.Ser = ser
				case env.Function:
					for i := 0; i < llen; i++ {
						var item any
						if modeObj == 1 {
							item = ll[i]
						} else if modeObj == 2 {
							item = lo[i]
						} else {
							item = env.ToRyeValue(ls[i])
						}
						CallFunctionArgsN(block, ps, ps.Ctx, env.ToRyeValue(item)) // , env.NewInteger(int64(i)))
						if util.IsTruthy(ps.Res) {                                 // todo -- move these to util or something
							if modeObj == 1 {
								newll = append(newll, ll[i])
							} else if modeObj == 2 {
								newlo = append(newlo, lo[i])
							} else {
								newlo = append(newlo, item.(env.Object))
							}
						}
					}
				case env.Builtin:
					for i := 0; i < llen; i++ {
						var item any
						if modeObj == 1 {
							item = ll[i]
						} else if modeObj == 2 {
							item = lo[i]
						} else {
							item = env.ToRyeValue(ls[i])
						}
						res := DirectlyCallBuiltin(ps, block, env.ToRyeValue(item), nil)
						if util.IsTruthy(res) { // todo -- move these to util or something
							if modeObj == 1 {
								newll = append(newll, ll[i])
							} else if modeObj == 2 {
								newlo = append(newlo, lo[i])
							} else {
								newlo = append(newlo, item.(env.Object))
							}
						}
					}
				default:
					return MakeBuiltinError(ps, "Block type should be Builtin or Block.", "filter")
				}
				if modeObj == 1 {
					return *env.NewList(newll)
				} else if modeObj == 2 {
					return *env.NewBlock(*env.NewTSeries(newlo))
				} else {
					return *env.NewBlock(*env.NewTSeries(newlo))
				}

			default:
				return MakeArgError(ps, 2, []env.Type{env.BlockType, env.BuiltinType}, "filter")
			}
		},
	},
	// Tests:
	//  equal { seek { 1 2 3 4 } { .is-even } } 2
	//  equal { seek list { 1 2 3 4 } { .is-even } } 2
	//  equal { seek "1234" { .to-integer .is-even } } "2"
	//  equal { try { seek { 1 2 3 4 } { > 5 } } |type? } 'error
	//  equal { try { seek list { 1 2 3 4 } { > 5 } } |type? } 'error
	//  equal { try { seek "1234" { .to-integer > 5 } } |type? } 'error
	// Args:
	// * series: Block, List, or String to search through
	// * code: Block or Builtin that returns true when the desired value is found
	// Returns:
	// * first value for which the code returns true, error if none found
	"seek": { // **
		Argsn: 2,
		Doc:   "Seek over a series until a Block of code returns True and return the value.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			var ll []any
			var lo []env.Object
			var ls []rune
			var llen int
			modeObj := 0
			switch data := arg0.(type) {
			case env.String:
				ls = []rune(data.Value)
				llen = len(ls)
				modeObj = 3
			case env.Block:
				lo = data.Series.S
				llen = len(lo)
				modeObj = 2
			case env.List:
				ll = data.Data
				llen = len(ll)
				modeObj = 1
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType, env.ListType, env.StringType}, "seek")
			}
			switch block := arg1.(type) {
			case env.Block, env.Builtin:
				switch block := block.(type) {
				case env.Block:
					ser := ps.Ser
					ps.Ser = block.Series
					for i := 0; i < llen; i++ {
						var item any
						if modeObj == 1 {
							item = ll[i]
						} else if modeObj == 2 {
							item = lo[i]
						} else {
							item = *env.NewString(string(ls[i]))
						}
						EvalBlockInjMultiDialect(ps, env.ToRyeValue(item), true)
						MaybeDisplayFailureOrError(ps, ps.Idx, "seek")
						if ps.ErrorFlag {
							ps.Ser = ser
							return ps.Res
						}
						if util.IsTruthy(ps.Res) { // todo -- move these to util or something
							ps.Ser = ser
							return env.ToRyeValue(item)
						}
						ps.Ser.Reset()
					}
					ps.Ser = ser
				case env.Builtin:
					for i := 0; i < llen; i++ {
						var item any
						if modeObj == 1 {
							item = ll[i]
						} else if modeObj == 2 {
							item = lo[i]
						} else {
							item = *env.NewString(string(ls[i]))
						}
						res := DirectlyCallBuiltin(ps, block, env.ToRyeValue(item), nil)
						if util.IsTruthy(res) { // todo -- move these to util or something
							return env.ToRyeValue(item)
						}
					}
				default:
					ps.ErrorFlag = true
					return MakeBuiltinError(ps, "Second argument should be block, builtin (or function).", "seek")
				}
			default:
				return MakeArgError(ps, 2, []env.Type{env.BlockType, env.BuiltinType}, "seek")
			}
			return MakeBuiltinError(ps, "No element found.", "seek")
		},
	},

	// Tests:
	// equal { var 'x 0 while { x < 5 } { x:: x + 1 } x } 5
	// equal { var 'x 0 var 'y 0 while { x < 5 } { x:: x + 1 y:: y + x } y } 15
	// Args:
	// * condition: Block that evaluates to a boolean to determine whether to continue
	// * body: Block of code to execute repeatedly
	// Returns:
	// * result of the last body execution
	"while": {
		Argsn: 2,
		Doc:   "Executes a block of code repeatedly while a condition is true.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch cond := arg0.(type) {
			case env.Block:
				switch bloc := arg1.(type) {
				case env.Block:
					ser := ps.Ser
					var prevRes env.Object
					for {
						// Check for interrupt signal (Ctrl+C, Ctrl+Z)
						if ps.InterruptFlag {
							ps.InterruptFlag = false // Reset the flag
							ps.Ser = ser
							return *env.NewError("Operation interrupted by user")
						}

						// Evaluate the condition block
						ps.Ser = cond.Series
						ps.Ser.Reset()
						EvalBlock(ps)
						MaybeDisplayFailureOrError(ps, ps.Idx, "while")
						if ps.ErrorFlag {
							ps.Ser = ser
							return ps.Res
						}

						// Check if the condition is true
						res, ok := ps.Res.(env.Boolean)
						if !ok {
							ps.Ser = ser
							return MakeBuiltinError(ps, "Result of condition block should be a Boolean value", "while")
						}
						if res.Value == false {
							break
						}

						// Execute the body block
						ps.Ser = bloc.Series
						ps.Ser.Reset()
						EvalBlock(ps)
						MaybeDisplayFailureOrError(ps, ps.Idx, "while")
						if ps.ErrorFlag || ps.ReturnFlag {
							ps.Ser = ser
							return ps.Res
						}
						prevRes = ps.Res
					}
					ps.Ser = ser
					return prevRes
				default:
					return MakeArgError(ps, 2, []env.Type{env.BlockType}, "while")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType}, "while")
			}
		},
	},

	// Tests:
	// equal { var 'x 5 until { x = 0 } { x:: x - 1 } x } 0
	// equal { var 'x 10 var 'y 0 until { x = 5 } { x:: x - 1 y:: y + x } y } 35
	// equal { var 'x 3 until { x < 1 } { x:: x - 1 } x } 0
	// equal { var 'x 0 until { x > 0 } { x:: x + 1 } x } 1
	// Args:
	// * body: Block of code to execute repeatedly (executed at least once)
	// * condition: Block that evaluates to a boolean to determine when to stop
	// Returns:
	// * result of the last body execution
	"until": {
		Argsn: 2,
		Doc:   "Executes a block of code repeatedly until a condition becomes true. Executes the body at least once before checking condition (do-until loop).",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch cond := arg1.(type) {
			case env.Block:
				switch bloc := arg0.(type) {
				case env.Block:
					ser := ps.Ser
					for {
						// Execute the body block first (do-until behavior)
						ps.Ser = bloc.Series
						ps.Ser.Reset()
						EvalBlock(ps)
						MaybeDisplayFailureOrError(ps, ps.Idx, "until")
						if ps.ErrorFlag || ps.ReturnFlag {
							ps.Ser = ser
							return ps.Res
						}

						// Then evaluate the condition block
						ps.Ser = cond.Series
						ps.Ser.Reset()
						EvalBlock(ps)
						MaybeDisplayFailureOrError(ps, ps.Idx, "until")
						if ps.ErrorFlag {
							ps.Ser = ser
							return ps.Res
						}

						// Check if the condition is true - if so, exit
						if util.IsTruthy(ps.Res) {
							break
						}
					}
					ps.Ser = ser
					return ps.Res
				default:
					return MakeArgError(ps, 1, []env.Type{env.BlockType}, "until")
				}
			default:
				return MakeArgError(ps, 2, []env.Type{env.BlockType}, "until")
			}
		},
	},

	// Tests:
	// equal { { 1 2 3 } .for\ 'x { x .prns } } ""  ; Expected output: "1 2 3 "
	// equal { list { 10 20 30 } |for\ 'val { val .prns } } ""  ; Expected output: "10 20 30 "
	// equal { "abc" .for\ 'ch { ch .prns } } ""  ; Expected output: "a b c "
	// Args:
	// * collection: String, Block, List, or other Collection to iterate over
	// * word: Word to store each value during iteration
	// * block: Block of code to execute for each value
	// Returns:
	// * result of the last block execution
	"for\\": {
		Argsn: 3,
		Doc:   "Iterates over a collection, setting a word to each value. Works with lists, blocks, and strings.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch valueWord := arg1.(type) {
			case env.Word:
				switch block := arg2.(type) {
				case env.Block:
					ser := ps.Ser
					ps.Ser = block.Series

					switch collection := arg0.(type) {
					case env.String:
						// Iterate over string characters
						for _, ch := range collection.Value {
							val := *env.NewString(string(ch))
							ps.Ctx.Mod(valueWord.Index, val)
							EvalBlock(ps)
							MaybeDisplayFailureOrError(ps, ps.Idx, "for\\")
							if ps.ErrorFlag || ps.ReturnFlag {
								ps.Ser = ser
								return ps.Res
							}
							ps.Ser.Reset()
						}
					case env.Collection:
						// Iterate over collection items (Block, List, etc.)
						length := collection.Length()
						for i := 0; i < length; i++ {
							val := collection.Get(i)
							ps.Ctx.Mod(valueWord.Index, val)
							EvalBlock(ps)
							MaybeDisplayFailureOrError(ps, ps.Idx, "for\\")
							if ps.ErrorFlag || ps.ReturnFlag {
								ps.Ser = ser
								return ps.Res
							}
							ps.Ser.Reset()
						}
					default:
						ps.Ser = ser
						return MakeArgError(ps, 1, []env.Type{env.BlockType, env.ListType, env.StringType}, "for\\")
					}
					ps.Ser = ser
					return ps.Res
				default:
					return MakeArgError(ps, 3, []env.Type{env.BlockType}, "for\\")
				}
			default:
				return MakeArgError(ps, 2, []env.Type{env.WordType}, "for\\")
			}
		},
	},

	// Tests:
	// equal { dict { "a" 1 "b" 2 } |for\kv 'k 'v { k .prns v .prns } } ""  ; Expected output: "a 1 b 2 "
	// equal { { 10 20 30 } .for\kv 'i 'v { i .prns v .prns } } ""  ; Expected output: "0 10 1 20 2 30 "
	// Args:
	// * collection: Dict, Block, List, or other Collection to iterate over
	// * keyWord: Word to store the key (or index for non-dict collections)
	// * valueWord: Word to store the value
	// * block: Block of code to execute for each key-value pair
	// Returns:
	// * result of the last block execution
	"for\\kv": {
		Argsn: 4,
		Doc:   "Iterates over a collection with separate key and value words. For dicts, uses actual keys and values. For other collections, uses indices as keys.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch keyWord := arg1.(type) {
			case env.Word:
				switch valueWord := arg2.(type) {
				case env.Word:
					switch block := arg3.(type) {
					case env.Block:
						ser := ps.Ser
						ps.Ser = block.Series

						switch collection := arg0.(type) {
						case env.Dict:
							// Iterate over dictionary key-value pairs
							for key, value := range collection.Data {
								ps.Ctx.Mod(keyWord.Index, *env.NewString(key))
								val := env.ToRyeValue(value)
								ps.Ctx.Mod(valueWord.Index, val)
								EvalBlockInjMultiDialect(ps, val, true)
								MaybeDisplayFailureOrError(ps, ps.Idx, "for\\kv")
								if ps.ErrorFlag || ps.ReturnFlag {
									ps.Ser = ser
									return ps.Res
								}
								ps.Ser.Reset()
							}
						case env.Collection:
							// For other collections, use index as key and item as value
							length := collection.Length()
							for i := 0; i < length; i++ {
								ps.Ctx.Mod(keyWord.Index, *env.NewInteger(int64(i)))
								val := collection.Get(i)
								ps.Ctx.Mod(valueWord.Index, val)
								EvalBlockInjMultiDialect(ps, val, true)
								// ERRORING . if we don't display here, nothing of meaning is displayed
								// just runtime error
								MaybeDisplayFailureOrError(ps, ps.Idx, "for\\kv")
								if ps.ErrorFlag || ps.ReturnFlag {
									// fmt.Println("IS ERROR 1")
									// ERRORING . if we return the ser then error location moves to outside of for\kv
									// so we don't want to
									ps.Ser = ser
									return ps.Res
								}
								ps.Ser.Reset()
							}
						default:
							ps.Ser = ser
							return MakeArgError(ps, 1, []env.Type{env.DictType, env.BlockType, env.ListType, env.StringType}, "for\\kw")
						}
						ps.Ser = ser
						return ps.Res
					default:
						return MakeArgError(ps, 4, []env.Type{env.BlockType}, "for\\kw")
					}
				default:
					return MakeArgError(ps, 3, []env.Type{env.WordType}, "for\\kw")
				}
			default:
				return MakeArgError(ps, 2, []env.Type{env.WordType}, "for\\kw")
			}
		},
	},
}
