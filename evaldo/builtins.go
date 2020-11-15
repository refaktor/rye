// builtins.go
package evaldo

import (
	"reflect"
	//	"bufio"
	"fmt"
	//	"os"
	"rye/env"

	"rye/loader"
	"rye/util"
	"strconv"
	"strings"
	"time"
)

func ss() {
	fmt.Print(1)
}

var builtins = map[string]*env.Builtin{

	"oneone": {
		Argsn: 0,

		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return env.Integer{11}
		},
	},
	"inc": {
		Argsn: 1,
		Pure:  true,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch arg := arg0.(type) {
			case env.Integer:
				return env.Integer{1 + arg.Value}
			default:
				return env.NewError("argument to `len` not supported, got %s")
			}
		},
	},

	// BASIC FUNCTIONS WITH NUMBERS

	"true": {
		Argsn: 0,
		Pure:  true,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return env.Integer{1}
		},
	},

	"false": {
		Argsn: 0,
		Pure:  true,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return env.Integer{0}
		},
	},

	"not": {
		Argsn: 1,
		Pure:  true,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			if util.IsTruthy(arg0) {
				return env.Integer{0}
			} else {
				return env.Integer{1}
			}
		},
	},

	"divides": {
		Argsn: 2,
		Pure:  true,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch arg := arg0.(type) {
			case env.Integer:
				if arg.Value%arg1.(env.Integer).Value == 0 {
					return env.Integer{1}
				} else {
					return env.Integer{0}
				}
			default:
				return env.NewError("argument to `len` not supported, got %s")
			}
		},
	},

	"_+": {
		Argsn: 2,
		Pure:  true,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Integer:
				return env.Integer{s1.Value + arg1.(env.Integer).Value}
			case env.String:
				switch s2 := arg1.(type) {
				case env.String:
					return env.String{s1.Value + s2.Value}
				case env.Integer:
					return env.String{s1.Value + strconv.Itoa(int(s2.Value))}
				}
			case env.Block:
				switch b2 := arg1.(type) {
				case env.Block:
					s := &s1.Series
					s1.Series = *s.AppendMul(b2.Series.GetAll())
					return s1
				}
			default:
				env1.FailureFlag = true
				return env.NewError("Wrong arguments for + function")
			}
			env1.FailureFlag = true
			return env.NewError("Wrong arguments for + function")
		},
	},

	"add": {
		Argsn: 2,
		Pure:  true,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Integer:
				return env.Integer{s1.Value + arg1.(env.Integer).Value}
			case env.String:
				switch s2 := arg1.(type) {
				case env.String:
					return env.String{s1.Value + s2.Value}
				case env.Integer:
					return env.String{s1.Value + strconv.Itoa(int(s2.Value))}
				}
			case env.Block:
				switch b2 := arg1.(type) {
				case env.Block:
					s := &s1.Series
					s1.Series = *s.AppendMul(b2.Series.GetAll())
					return s1
				}
			default:
				env1.FailureFlag = true
				return env.NewError("Wrong arguments for + function")
			}
			env1.FailureFlag = true
			return env.NewError("Wrong arguments for + function")
		},
	},
	"_-": {
		Argsn: 2,
		Pure:  true,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch arg := arg0.(type) {
			case env.Integer:
				return env.Integer{arg.Value - arg1.(env.Integer).Value}
			default:
				return env.NewError("argument to `len` not supported, got %s")
			}
		},
	},
	"_*": {
		Argsn: 2,
		Pure:  true,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch arg := arg0.(type) {
			case env.Integer:
				return env.Integer{arg.Value * arg1.(env.Integer).Value}
			default:
				return env.NewError("argument to `len` not supported, got %s")
			}
		},
	},
	"_=": {
		Argsn: 2,
		Pure:  true,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			var res int64
			if arg0.GetKind() == arg1.GetKind() && arg0.Inspect(*env1.Idx) == arg1.Inspect(*env1.Idx) {
				res = 1
			} else {
				res = 0
			}
			return env.Integer{res}
		},
	},
	"_!": {
		Argsn: 2,
		Pure:  true,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			var res int64
			if arg0.GetKind() == arg1.GetKind() && arg0.Inspect(*env1.Idx) == arg1.Inspect(*env1.Idx) {
				res = 0
			} else {
				res = 1
			}
			return env.Integer{res}
		},
	},
	"_>": {
		Argsn: 2,
		Pure:  true,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch arg := arg0.(type) {
			case env.Integer:
				var res int64
				if arg.Value > arg1.(env.Integer).Value {
					res = 1
				} else {
					res = 0
				}
				return env.Integer{res}
			default:
				return env.NewError("argument to `len` not supported, got %s")
			}
		},
	},
	"_<": {
		Argsn: 2,
		Pure:  true,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch arg := arg0.(type) {
			case env.Integer:
				var res int64
				if arg.Value < arg1.(env.Integer).Value {
					res = 1
				} else {
					res = 0
				}
				return env.Integer{res}
			default:
				return env.NewError("argument to `len` not supported, got %s")
			}
		},
	},

	// BASIC GENERAL FUNCTIONS

	"inspect": {
		Argsn: 1,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			fmt.Println(arg0.Inspect(*env1.Idx))
			return arg0
		},
	},
	"probe": {
		Argsn: 1,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			fmt.Println(arg0.Probe(*env1.Idx))
			return arg0
		},
	},
	"prn": {
		Argsn: 1,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch arg := arg0.(type) {
			case env.String:
				fmt.Print(arg.Value + " ")
			default:
				fmt.Print(arg0.Probe(*env1.Idx) + " ")
			}
			return arg0
		},
	},
	"prin": {
		Argsn: 1,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch arg := arg0.(type) {
			case env.String:
				fmt.Print(arg.Value)
			default:
				fmt.Print(arg0.Probe(*env1.Idx))
			}
			return arg0
		},
	},
	"print": {
		Argsn: 1,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch arg := arg0.(type) {
			case env.String:
				fmt.Println(arg.Value)
			default:
				fmt.Println(arg0.Probe(*env1.Idx))
			}
			return arg0
		},
	},

	// CONTROL WORDS

	"unless": {
		Argsn: 2,
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch cond := arg0.(type) {
			case env.Integer:
				switch bloc := arg1.(type) {
				case env.Block:
					if cond.Value == 0 {
						ser := ps.Ser
						ps.Ser = bloc.Series
						EvalBlock(ps)
						ps.Ser = ser
						return ps.Res
					}
				}
			}
			return nil
		},
	},

	"if": {
		Argsn: 2,
		Pure:  true,
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
				cond1 := util.IsTruthy(arg0)

				// if arg0 is ok and arg1 is block we end up here
				// if cond1 is true (arg0 was truthy), otherwise we don't do anything
				// later we should return void or null, or ... we still have to decide
				if cond1 {
					// we store current series (block of code with position we are at) to temp 'ser'
					ser := ps.Ser
					// we set ProgramStates series to series ob the block
					ps.Ser = bloc.Series
					// we eval the block (current context / scope stays the same as it was in parent block)
					// Inj means we inject the condition value into the block, because it costs us very little. we could do "if name { .print }"
					EvalBlockInj(ps, arg0, true)
					// we set temporary series back to current program state
					ps.Ser = ser
					// we return the last return value (the return value of executing the block) "a: if 1 { 100 }" a becomes 100,
					// in future we will also handle the "else" case, but we have to decide
					return ps.Res
				}
			default:
				// if it's not a block we return error for now
				ps.FailureFlag = true
				return env.NewError("Error if")
			}
			return nil
		},
	},

	"either": {
		Argsn: 3,
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			//arg0.Trace("")
			//arg1.Trace("")
			//arg2.Trace("")
			var cond1 bool
			switch bloc1 := arg1.(type) {
			case env.Block:
				switch bloc2 := arg2.(type) {
				case env.Block:
					switch cond := arg0.(type) {
					case env.Integer:
						cond1 = cond.Value != 0
					case env.String:
						cond1 = cond.Value != ""
					default:
						return env.NewError("Error either")
					}
					ser := ps.Ser
					if cond1 {
						ps.Ser = bloc1.Series
						ps.Ser.Reset()
					} else {
						ps.Ser = bloc2.Series
						ps.Ser.Reset()
					}
					EvalBlockInj(ps, arg0, true)
					ps.Ser = ser
					return ps.Res
				}

			}
			return nil
		},
	},

	"^tidy-switch": {
		Argsn:         2,
		AcceptFailure: true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			fmt.Println("FLAGS")

			ps.FailureFlag = false
			//fmt.Println(arg0.Probe(*ps.Idx))

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
							return env.NewError("switch block malformed")
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
						}
					}
					switch cc := code.(type) {
					case env.Block:
						fmt.Println(code.Probe(*ps.Idx))
						// we store current series (block of code with position we are at) to temp 'ser'
						ser := ps.Ser
						// we set ProgramStates series to series ob the block
						ps.Ser = cc.Series
						// we eval the block (current context / scope stays the same as it was in parent block)
						// Inj means we inject the condition value into the block, because it costs us very little. we could do "if name { .print }"
						EvalBlockInj(ps, arg0, true)
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
						return env.NewError("Malformed switch block")
					}
				default:
					// if it's not a block we return error for now
					ps.FailureFlag = true
					return env.NewError("Second arg not block")
				}
			default:
				return arg0
			}
		},
	},

	"cases": {
		Argsn: 2,
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
						// we set temporary series back to current program state
						if util.IsTruthy(ps.Res) {
							doblk = true
							foundany = true
						}
					case env.Void:
						if foundany == false {
							doblk = true
						}
					}
					// we set ProgramStates series to series ob the block
					if doblk {
						ps.Ser = blk.Series
						// we eval the block (current context / scope stays the same as it was in parent block)
						// Inj means we inject the condition value into the block, because it costs us very little. we could do "if name { .print }"
						EvalBlockInj(ps, cumul, true)
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
				return env.NewError("Error if")
			}
			return nil
		},
	},

	"drop-in": {
		Argsn: 2,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch bloc := arg1.(type) {
			case env.Block:
				ser := ps.Ser
				ps.Ser = bloc.Series
				EvalBlock(ps)
				ps.Ser = ser
				//reader := bufio.NewReader(os.Stdin)

				fmt.Println("You are inside a drop-in: " + arg0.(env.String).Value)
				fmt.Println(" * user (h) for help, (lc) to list context , (r) to return")
				fmt.Println("----------------------------------------------------------")
				/*
					for {
						fmt.Print("{ rye dropin }")
						text, _ := reader.ReadString('\n')
						//fmt.Println(1111)
						// convert CRLF to LF
						text = strings.Replace(text, "\n", "", -1)
						//fmt.Println(1111)
						if strings.Compare("(lc)", text) == 0 {
							fmt.Println(ps.Ctx.Probe(*ps.Idx))
						} else if strings.Compare("(r)", text) == 0 {
							ps.Ser = ser
							return ps.Res
						} else {
							// fmt.Println(1111)
							block, genv := loader.LoadString("{ " + text + " }")
							ps := env.AddToProgramState(ps, block.Series, genv)
							EvalBlock(ps)
							fmt.Println(ps.Res.Inspect(*ps.Idx))
						}
					}*/

				DoRyeRepl(ps)
				fmt.Println("----------------------------------------------------------")
				ps.Ser = ser

				return ps.Res
			}
			return nil
		},
	},

	"do": {
		Argsn: 1,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch bloc := arg0.(type) {
			case env.Block:
				ser := ps.Ser
				ps.Ser = bloc.Series
				EvalBlock(ps)
				ps.Ser = ser
				return ps.Res
			}
			return nil
		},
	},

	"do-with": {
		Argsn: 2,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch bloc := arg1.(type) {
			case env.Block:
				ser := ps.Ser
				ps.Ser = bloc.Series
				EvalBlockInj(ps, arg0, true)
				ps.Ser = ser
				return ps.Res
			}
			return nil
		},
	},

	"do-in": {
		Argsn: 2,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch ctx := arg0.(type) {
			case env.RyeCtx:
				switch bloc := arg1.(type) {
				case env.Block:
					ser := ps.Ser
					ps.Ser = bloc.Series
					EvalBlockInCtx(ps, &ctx)
					ps.Ser = ser
					return ps.Res
				default:
					ps.ErrorFlag = true
					return env.NewError("Secong arg should be block")

				}
			default:
				ps.ErrorFlag = true
				return env.NewError("First arg should be context")
			}

		},
	},

	"range": {
		Argsn: 2,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch i1 := arg0.(type) {
			case env.Integer:
				switch i2 := arg1.(type) {
				case env.Integer:
					objs := make([]env.Object, i2.Value-i1.Value+1)
					idx := 0
					for i := i1.Value; i <= i2.Value; i++ {
						objs[idx] = env.Integer{i}
						idx += 1
					}
					return *env.NewBlock(*env.NewTSeries(objs))
				}

			}
			ps.ErrorFlag = true
			return env.NewError("First arg should be context")
		},
	},

	// SPECIAL FUNCTION FUNCTIONS

	// CONTEXT FUNCTIONS

	"returns": {
		Argsn: 1,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			ps.ForcedResult = arg0
			return arg0
		},
	},

	"collect": {
		Argsn: 1,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			if ps.ForcedResult == nil || ps.ForcedResult.Type() != env.BlockType {
				ps.ForcedResult = *env.NewBlock(*env.NewTSeries(make([]env.Object, 0)))
			}
			block := ps.ForcedResult.(env.Block)
			block.Series.Append(arg0)
			ps.ForcedResult = block
			return arg0
		},
	},

	"collect-key": {
		Argsn: 2,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			if ps.ForcedResult == nil || ps.ForcedResult.Type() != env.BlockType {
				ps.ForcedResult = *env.NewBlock(*env.NewTSeries(make([]env.Object, 0)))
			}
			block := ps.ForcedResult.(env.Block)
			block.Series.Append(arg1)
			block.Series.Append(arg0)
			ps.ForcedResult = block
			return arg0
		},
	},

	"collected": {
		Argsn: 0,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return ps.ForcedResult
		},
	},

	"current-context": {
		Argsn: 0,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *ps.Ctx
		},
	},
	"parent-context": {
		Argsn: 0,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *ps.Ctx.Parent
		},
	},

	"raw-context": {
		Argsn: 1,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch bloc := arg0.(type) {
			case env.Block:
				ser := ps.Ser
				ctx := ps.Ctx
				ps.Ser = bloc.Series
				ps.Ctx = env.NewEnv(nil) // make new context with no parent
				EvalBlock(ps)
				rctx := ps.Ctx
				ps.Ctx = ctx
				ps.Ser = ser
				if ps.ErrorFlag {
					return ps.Res
				}
				return *rctx // return the resulting context
			}
			return nil
		},
	},

	"isolate": {
		Argsn: 1,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch bloc := arg0.(type) {
			case env.Block:
				ser := ps.Ser
				ctx := ps.Ctx
				ps.Ser = bloc.Series
				ps.Ctx = env.NewEnv(ps.Ctx) // make new context with no parent
				EvalBlock(ps)
				rctx := ps.Ctx
				rctx.Parent = nil
				ps.Ctx = ctx
				ps.Ser = ser
				if ps.ErrorFlag {
					return ps.Res
				}
				return *rctx // return the resulting context
			}
			return nil
		},
	},

	"context": {
		Argsn: 1,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch bloc := arg0.(type) {
			case env.Block:
				ser := ps.Ser
				ctx := ps.Ctx
				ps.Ser = bloc.Series
				ps.Ctx = env.NewEnv(ps.Ctx) // make new context with no parent
				EvalBlock(ps)
				rctx := ps.Ctx
				ps.Ctx = ctx
				ps.Ser = ser
				return *rctx // return the resulting context
			}
			return nil
		},
	},

	"extend!": { // exclamation mark, because it as it is now extends/changes the source context too .. in place
		Argsn: 2,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch ctx0 := arg0.(type) {
			case env.RyeCtx:
				switch bloc := arg1.(type) {
				case env.Block:
					ser := ps.Ser
					ctx := ps.Ctx
					ps.Ser = bloc.Series
					ps.Ctx = &ctx0 // make new context with no parent
					EvalBlock(ps)
					rctx := ps.Ctx
					ps.Ctx = ctx
					ps.Ser = ser
					return *rctx // return the resulting context
				}
			}
			ps.ErrorFlag = true
			return env.NewError("Second argument should be block, builtin (or function).")
		},
	},

	"Ctx//bind": { // TODO -- check if this works
		Argsn: 2,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch swCtx1 := arg0.(type) {
			case env.RyeCtx:
				switch swCtx2 := arg1.(type) {
				case env.RyeCtx:
					swCtx1.Parent = &swCtx2
					return swCtx1
				}
			}
			return env.NewError("wrong args")
		},
	},

	"unbind": {
		Argsn: 1,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch swCtx1 := arg0.(type) {
			case env.RyeCtx:
				swCtx1.Parent = nil
				return swCtx1
			}
			return env.NewError("wrong args")
		},
	},

	// COMBINATORS

	"pass": {
		Argsn: 2,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch bloc := arg1.(type) {
			case env.Block:
				res := arg0
				ser := ps.Ser
				ps.Ser = bloc.Series
				EvalBlockInj(ps, arg0, true)
				ps.Ser = ser
				return res
			}
			return nil
		},
	},

	"keep": {
		Argsn: 3,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch b1 := arg1.(type) {
			case env.Block:
				switch b2 := arg2.(type) {
				case env.Block:
					ser := ps.Ser
					ps.Ser = b1.Series
					EvalBlockInj(ps, arg0, true)
					res := ps.Res
					ps.Ser = b2.Series
					EvalBlockInj(ps, arg0, true)
					ps.Ser = ser
					return res
				}
			}
			return nil
		},
	},

	"with": {
		AcceptFailure: true,
		Argsn:         2,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch bloc := arg1.(type) {
			case env.Block:
				ser := ps.Ser
				ps.Ser = bloc.Series
				EvalBlockInj(ps, arg0, true)
				ps.Ser = ser
				return ps.Res
			}
			return nil
		},
	},

	//

	"time-it": {
		Argsn: 1,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch bloc := arg0.(type) {
			case env.Block:
				ser := ps.Ser
				ps.Ser = bloc.Series
				start := time.Now()
				EvalBlock(ps)
				t := time.Now()
				elapsed := t.Sub(start)
				ps.Ser = ser
				return env.Integer{elapsed.Nanoseconds() / 1000000}
			}
			return nil
		},
	},

	"loop": {
		Argsn: 2,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch cond := arg0.(type) {
			case env.Integer:
				switch bloc := arg1.(type) {
				case env.Block:
					ser := ps.Ser
					ps.Ser = bloc.Series
					for i := 0; int64(i) < cond.Value; i++ {
						ps = EvalBlock(ps)
						ps.Ser.Reset()
					}
					ps.Ser = ser
					return ps.Res
				}
			}
			return nil
		},
	},

	"forever": {
		Argsn: 1,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch bloc := arg0.(type) {
			case env.Block:
				ser := ps.Ser
				ps.Ser = bloc.Series
				for {
					ps = EvalBlock(ps)
					ps.Ser.Reset()
				}
				ps.Ser = ser
				return ps.Res
			default:
				ps.FailureFlag = true
				return env.NewError("arg0 should be block	")
			}
		},
	},

	"for": {
		Argsn: 2,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch block := arg0.(type) {
			case env.Block:
				switch code := arg1.(type) {
				case env.Block:
					ser := ps.Ser
					ps.Ser = code.Series
					for i := 0; i < block.Series.Len(); i++ {
						ps = EvalBlockInj(ps, block.Series.Get(i), true)
						ps.Ser.Reset()
					}
					ps.Ser = ser
					return ps.Res
				}
			case env.List:
				switch code := arg1.(type) {
				case env.Block:
					ser := ps.Ser
					ps.Ser = code.Series
					for i := 0; i < len(block.Data); i++ {
						ps = EvalBlockInj(ps, JsonToRye(block.Data[i]), true)
						ps.Ser.Reset()
					}
					ps.Ser = ser
					return ps.Res
				}
			}
			return nil
		},
	},

	// map should at the end map over block, raw-map, etc ...
	// it should accept a block of code, a function and a builtin
	// it should use injected block so it doesn't need a variable defined like map [ 1 2 3 ] x [ add a 100 ]
	// map [ 1 2 3 ] { .add 3 }
	"map": {
		Argsn: 2,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch list := arg0.(type) {
			case env.Block:
				switch block := arg1.(type) {
				case env.Block, env.Builtin:
					l := list.Series.Len()
					newl := make([]env.Object, l)
					switch block := block.(type) {
					case env.Block:
						ser := ps.Ser
						ps.Ser = block.Series
						for i := 0; i < l; i++ {
							ps = EvalBlockInj(ps, list.Series.Get(i), true)
							newl[i] = ps.Res
							ps.Ser.Reset()
						}
						ps.Ser = ser
					case env.Builtin:
						for i := 0; i < l; i++ {
							newl[i] = DirectlyCallBuiltin(ps, block, list.Series.Get(i), nil)
						}
					}
					return *env.NewBlock(*env.NewTSeries(newl))
				}
			case env.List:
				switch block := arg1.(type) {
				case env.Block, env.Builtin:
					l := len(list.Data)
					newl := make([]interface{}, l)
					switch block := block.(type) {
					case env.Block:
						ser := ps.Ser
						ps.Ser = block.Series
						for i := 0; i < l; i++ {
							ps = EvalBlockInj(ps, JsonToRye(list.Data[i]), true)
							newl[i] = ps.Res
							ps.Ser.Reset()
						}
						ps.Ser = ser
					case env.Builtin:
						for i := 0; i < l; i++ {
							newl[i] = DirectlyCallBuiltin(ps, block, JsonToRye(list.Data[i]), nil)
						}
					}
					return *env.NewList(newl)
				}
			}
			return nil
		},
	},

	// filter [ 1 2 3 ] { .add 3 }
	"filter": {
		Argsn: 2,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			var ll []interface{}
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
			}

			if modeObj == 0 {
				ps.FailureFlag = true
				return env.NewError("expects list or block")
			}

			switch block := arg1.(type) {
			case env.Block, env.Builtin:
				var newl []env.Object
				switch block := block.(type) {
				case env.Block:
					ser := ps.Ser
					ps.Ser = block.Series
					for i := 0; i < llen; i++ {
						var item interface{}
						if modeObj == 1 {
							item = ll[i]
						} else {
							item = lo[i]
						}
						ps = EvalBlockInj(ps, JsonToRye(item), true)
						if util.IsTruthy(ps.Res) { // todo -- move these to util or something
							newl = append(newl, JsonToRye(item))
						}
						ps.Ser.Reset()
					}
					ps.Ser = ser
				case env.Builtin:
					for i := 0; i < llen; i++ {
						var item interface{}
						if modeObj == 1 {
							item = ll[i]
						} else {
							item = lo[i]
						}
						res := DirectlyCallBuiltin(ps, block, JsonToRye(item), nil)
						if util.IsTruthy(res) { // todo -- move these to util or something
							newl = append(newl, JsonToRye(item))
						}
					}
				}
				return *env.NewBlock(*env.NewTSeries(newl))
			}
			return nil
		},
	},

	"seek": {
		Argsn: 2,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch list := arg0.(type) {
			case env.Block:
				switch block := arg1.(type) {
				case env.Block, env.Builtin:
					l := list.Series.Len()
					switch block := block.(type) {
					case env.Block:
						ser := ps.Ser
						ps.Ser = block.Series
						for i := 0; i < l; i++ {
							ps = EvalBlockInj(ps, list.Series.Get(i), true)
							if util.IsTruthy(ps.Res) { // todo -- move these to util or something
								return list.Series.Get(i)
							}
							ps.Ser.Reset()
						}
						ps.Ser = ser
					case env.Builtin:
						for i := 0; i < l; i++ {
							res := DirectlyCallBuiltin(ps, block, list.Series.Get(i), nil)
							if util.IsTruthy(res) { // todo -- move these to util or something
								return list.Series.Get(i)
							}
						}
					default:
						ps.ErrorFlag = true
						return env.NewError("Second argument should be block, builtin (or function).")
					}
				}
			}
			return nil
		},
	},

	//test if we can do recur similar to clojure one. Since functions in rejy are of fixed arity we would need recur1 recur2 recur3 and recur [ ] which is less optimal
	//otherwise word recur could somehow be bound to correct version or args depending on number of args of func. Try this at first.
	"recur1if": { //recur1-if
		Argsn: 2,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch cond := arg0.(type) {
			case env.Integer:
				if cond.Value > 0 {
					switch arg := arg1.(type) {
					case env.Integer:
						ps.Ctx.Set(ps.Args[0], arg)
						ps.Ser.Reset()
						return nil
					}
				} else {
					return ps.Res
				}
			}
			return nil
		},
	},

	"recur2if": { //recur1-if
		Argsn: 3,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			//arg0.Trace("a0")
			//arg1.Trace("a1")
			//arg2.Trace("a2")
			switch cond := arg0.(type) {
			case env.Integer:
				if cond.Value > 0 {
					switch argi1 := arg1.(type) {
					case env.Integer:
						switch argi2 := arg2.(type) {
						case env.Integer:
							ps.Ctx.Set(ps.Args[0], argi1)
							ps.Ctx.Set(ps.Args[1], argi2)
							ps.Ser.Reset()
							return ps.Res
						}
					}
				} else {
					return ps.Res
				}
			}
			return nil
		},
	},

	"recur3if": { //recur1-if
		Argsn: 4,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			//arg0.Trace("a0")
			//arg1.Trace("a1")
			//arg2.Trace("a2")
			switch cond := arg0.(type) {
			case env.Integer:
				if cond.Value > 0 {
					switch argi1 := arg1.(type) {
					case env.Integer:
						switch argi2 := arg2.(type) {
						case env.Integer:
							switch argi3 := arg3.(type) {
							case env.Integer:
								ps.Ctx.Set(ps.Args[0], argi1)
								ps.Ctx.Set(ps.Args[1], argi2)
								ps.Ctx.Set(ps.Args[2], argi3)
								ps.Ser.Reset()
								return ps.Res
							}
						}
					}
				} else {
					return ps.Res
				}
			}
			return nil
		},
	},

	"does": {
		Argsn: 1,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch body := arg0.(type) {
			case env.Block:
				//spec := []env.Object{env.Word{aaaidx}}
				//body := []env.Object{env.Word{printidx}, env.Word{aaaidx}, env.Word{recuridx}, env.Word{greateridx}, env.Integer{99}, env.Word{aaaidx}, env.Word{incidx}, env.Word{aaaidx}}
				return *env.NewFunction(*env.NewBlock(*env.NewTSeries(make([]env.Object, 0))), body, false)
			}
			return nil
		},
	},

	"pipe": {
		Argsn: 1,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch body := arg0.(type) {
			case env.Block:
				spec := []env.Object{env.Word{1}}
				//body := []env.Object{env.Word{printidx}, env.Word{aaaidx}, env.Word{recuridx}, env.Word{greateridx}, env.Integer{99}, env.Word{aaaidx}, env.Word{incidx}, env.Word{aaaidx}}
				return *env.NewFunction(*env.NewBlock(*env.NewTSeries(spec)), body, false)
			}
			return nil
		},
	},

	"fn": {
		Argsn: 2,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch args := arg0.(type) {
			case env.Block:
				switch body := arg1.(type) {
				case env.Block:
					//spec := []env.Object{env.Word{aaaidx}}
					//body := []env.Object{env.Word{printidx}, env.Word{aaaidx}, env.Word{recuridx}, env.Word{greateridx}, env.Integer{99}, env.Word{aaaidx}, env.Word{incidx}, env.Word{aaaidx}}
					return *env.NewFunction(args, body, false)
				}
			}
			return nil
		},
	},

	"pfn": {
		Argsn: 2,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch args := arg0.(type) {
			case env.Block:
				switch body := arg1.(type) {
				case env.Block:
					return *env.NewFunction(args, body, true)
				}
			}
			return nil
		},
	},

	"fnc": {
		// a function with context	 bb: 10 add10 [ a ] context [ b: bb ] [ add a b ]
		// 							add10 [ a ] this [ add a b ]
		// later maybe			   add10 [ a ] [ b: b ] [ add a b ]
		//  						   add10 [ a ] [ 'b ] [ add a b ]
		Argsn: 3,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch args := arg0.(type) {
			case env.Block:
				switch ctx := arg1.(type) {
				case env.RyeCtx:
					switch body := arg2.(type) {
					case env.Block:
						return *env.NewFunctionC(args, body, &ctx, false)
					default:
						ps.ErrorFlag = true
						return env.NewError("Third arg should be Block")
					}
				default:
					ps.ErrorFlag = true
					return env.NewError("Second arg should be Context")
				}
			default:
				ps.ErrorFlag = true
				return env.NewError("First argument should be Block")
			}
			return nil
		},
	},

	"kind": {
		Argsn: 2,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Tagword:
				switch spec := arg1.(type) {
				case env.Block:
					return *env.NewKind(s1.ToWord(), spec)
				default:
					return env.NewError("2nd not block")
				}
			default:
				return env.NewError("first not lit-word")
			}
			return nil
		},
	},

	"_>>": {
		Argsn: 2,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch spec := arg1.(type) {
			case env.Kind:
				switch dict := arg0.(type) {
				case env.Dict:
					obj := BuiValidate(ps, dict, spec.Spec)
					switch obj1 := obj.(type) {
					case env.Dict:
						ctx := util.Dict2Context(ps, obj1)
						ctx.Kind = spec.Kind
						return ctx
					default:
						return env.NewError("2nd A isn't ")
					}
				case env.RyeCtx:
					if spec.HasConverter(dict.Kind.Index) {
						obj := BuiConvert(ps, dict, spec.Converters[dict.Kind.Index])
						switch ctx := obj.(type) {
						case env.RyeCtx:
							ctx.Kind = spec.Kind
							return ctx
						default:
							return env.NewError("2344nd xxxx   xxx sn't Dict")
						}
					}
					return env.NewError("2nd xxxx   xxx sn't Dict")
				default:
					return env.NewError("2nd isn't Dict")
				}
			default:
				return env.NewError("1st isn't kind")
			}
			return env.NewError("1st isn't kind xxxx")
		},
	},

	"_<<": {
		Argsn: 2,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch spec := arg0.(type) {
			case env.Kind:
				switch dict := arg1.(type) {
				case env.Dict:
					obj := BuiValidate(ps, dict, spec.Spec)
					switch obj1 := obj.(type) {
					case env.Dict:
						ctx := util.Dict2Context(ps, obj1)
						ctx.Kind = spec.Kind
						return ctx
					default:
						return env.NewError("2nd A isn't ")
					}
				case env.RyeCtx:
					if spec.HasConverter(dict.Kind.Index) {
						obj := BuiConvert(ps, dict, spec.Converters[dict.Kind.Index])
						switch ctx := obj.(type) {
						case env.RyeCtx:
							ctx.Kind = spec.Kind
							return ctx
						default:
							return env.NewError("2344nd xxxx   xxx sn't Dict")
						}
					}
				default:
					return env.NewError("2nd isn't Dict")
				}
			default:
				return env.NewError("1st isn't kind")
			}
			return nil
		},
	},

	"assure-kind": {
		Argsn: 2,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch spec := arg1.(type) {
			case env.Kind:
				switch dict := arg0.(type) {
				case env.Dict:
					obj := BuiValidate(ps, dict, spec.Spec)
					switch obj1 := obj.(type) {
					case env.Dict:
						ctx := util.Dict2Context(ps, obj1)
						ctx.Kind = spec.Kind
						return ctx
					default:
						return env.NewError("2nd A isn't ")
					}
				default:
					return env.NewError("2nd isn't Dict")
				}
			default:
				return env.NewError("1st isn't kind")
			}
			return nil
		},
	},

	// BASIC STRING FUNCTIONS

	"left": {
		Argsn: 2,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.String:
				switch s2 := arg1.(type) {
				case env.Integer:
					return env.String{s1.Value[0:s2.Value]}
				}
			}
			return nil
		},
	},

	"middle": {
		Argsn: 3,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.String:
				switch s2 := arg1.(type) {
				case env.Integer:
					switch s3 := arg2.(type) {
					case env.Integer:
						return env.String{s1.Value[s2.Value:s3.Value]}
					}
				}
			}
			return nil
		},
	},

	"right": {
		Argsn: 2,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.String:
				switch s2 := arg1.(type) {
				case env.Integer:
					return env.String{s1.Value[len(s1.Value)-int(s2.Value):]}
				}
			}
			return nil
		},
	},

	"join": {
		Argsn: 2,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.String:
				switch s2 := arg1.(type) {
				case env.String:
					return env.String{s1.Value + s2.Value}
				case env.Integer:
					return env.String{s1.Value + strconv.Itoa(int(s2.Value))}
				}
			case env.Block:
				switch b2 := arg1.(type) {
				case env.Block:
					s := &s1.Series
					s1.Series = *s.AppendMul(b2.Series.GetAll())
					return s1
				}
			}
			return nil
		},
	},
	"capitalize": {
		Argsn: 1,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.String:
				return env.String{strings.Title(s1.Value)}
			default:
				return env.NewError("first arg must be string")
			}
		},
	},

	"join3": {
		Argsn: 3,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.String:
				switch s2 := arg1.(type) {
				case env.String:
					switch s3 := arg2.(type) {
					case env.String:
						return env.String{s1.Value + s2.Value + s3.Value}
					}
				}
			}
			return nil
		},
	},

	// BASIC SERIES FUNCTIONS

	"first": {
		Argsn: 1,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Block:
				return s1.Series.Get(int(0))
			case env.List:
				return JsonToRye(s1.Data[int(0)])
			default:
				return env.NewError("not block or list")
			}
		},
	},
	"second": {
		Argsn: 1,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Block:
				return s1.Series.Get(int(1))
			}
			return nil
		},
	},
	"third": {
		Argsn: 1,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Block:
				return s1.Series.Get(int(2))
			}
			return nil
		},
	},
	"last": {
		Argsn: 1,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Block:
				return s1.Series.Get(s1.Series.Len() - 1)
			}
			return nil
		},
	},

	"head": {
		Argsn: 2,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s2 := arg1.(type) {
			case env.Integer:
				switch s1 := arg0.(type) {
				case env.Block:
					return *env.NewBlock(*env.NewTSeries(s1.Series.S[0:s2.Value]))
				case env.List:
					return *env.NewList(s1.Data[0:s2.Value])
				default:
					return *env.NewError("not block or list")
				}
			default:
				return *env.NewError("not integer")
			}
		},
	},

	"nth": {
		Argsn: 2,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Block:
				switch s2 := arg1.(type) {
				case env.Integer:
					return s1.Series.Get(int(s2.Value - 1))
				}
			}
			return nil
		},
	},
	"length": {
		Argsn: 1,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Block:
				return env.Integer{int64(s1.Series.Len())}
			}
			return nil
		},
	},
	"peek": {
		Argsn: 1,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Block:
				return s1.Series.Peek()
			}
			return nil
		},
	},
	"pop": {
		Argsn: 1,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Block:
				return s1.Series.Pop()
			}
			return nil
		},
	},
	"pos": {
		Argsn: 1,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Block:
				return env.Integer{int64(s1.Series.Pos())}
			}
			return nil
		},
	},
	"next": {
		Argsn: 1,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Block:
				s1.Series.Next()
				return s1
			}
			return nil
		},
	},
	"append": {
		Argsn: 2,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Block:
				s := &s1.Series
				s1.Series = *s.Append(arg1)
				return s1
			}
			return nil
		},
	},
	// FUNCTIONALITY AROUND GENERIC METHODS
	// generic <integer> <add> fn [ a b ] [ a + b ] // tagwords are temporary here
	"generic": {
		Argsn: 3,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Tagword:
				switch s2 := arg1.(type) {
				case env.Tagword:
					switch s3 := arg2.(type) {
					case env.Object:
						fmt.Println(s1.Index)
						fmt.Println(s2.Index)
						fmt.Println("Generic")

						registerGeneric(ps, s1.Index, s2.Index, s3)
						return s3
					}
				}
			}
			ps.ErrorFlag = true
			return env.NewError("Wrong args when creating generic function")
		},
	},

	"dict": {
		Argsn: 1,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch bloc := arg0.(type) {
			case env.Block:
				return env.NewDictFromSeries(bloc.Series)
			}
			return nil
		},
	},

	// BASIC ENV / Dict FUNCTIONS
	"_->": {
		Argsn: 2,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Dict:
				switch s2 := arg1.(type) {
				case env.String:
					v := s1.Data[s2.Value]
					switch v1 := v.(type) {
					case int, int64, float64, string, []interface{}, map[string]interface{}:
						return JsonToRye(v1)
					case env.Integer:
						return v1
					case env.String:
						return v1
					case env.Date:
						return v1
					default:
						return env.NewError("Value of type: " + reflect.TypeOf(v1).String())
					}
				}
			case env.RyeCtx:
				switch s2 := arg1.(type) {
				case env.Tagword:
					v, ok := s1.Get(s2.Index)
					if ok {
						return v
					} else {
						ps.FailureFlag = true
						return env.NewError1(5) // NOT_FOUND
					}
				}
			case env.List:
				switch s2 := arg1.(type) {
				case env.Integer:
					v := s1.Data[s2.Value]
					ok := true
					if ok {
						return JsonToRye(v)
					} else {
						ps.FailureFlag = true
						return env.NewError1(5) // NOT_FOUND
					}
				}
			case env.Block:
				switch s2 := arg1.(type) {
				case env.Integer:
					v := s1.Series.Get(int(s2.Value))
					ok := true
					if ok {
						return v
					} else {
						ps.FailureFlag = true
						return env.NewError1(5) // NOT_FOUND
					}
				}
			}
			return nil
		},
	},

	// return , error , failure functions
	"return": {
		Argsn: 1,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			//fmt.Println("RETURN")
			ps.ReturnFlag = true
			return arg0
		},
	},

	"fail": {
		Argsn: 1,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			//fmt.Println("FAIL")
			ps.FailureFlag = true
			ps.ReturnFlag = true
			switch val := arg0.(type) {
			case env.String: // todo .. make Error type .. make error construction micro dialect, return the error wrapping error that caused it
				return *env.NewError(val.Value)
			case env.Integer: // todo .. make Error type .. make error construction micro dialect, return the error wrapping error that caused it
				return *env.NewError1(int(val.Value))
			}
			return arg0
		},
	},

	"failure": {
		Argsn: 1,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			//fmt.Println("FAIL")
			ps.FailureFlag = true
			switch val := arg0.(type) {
			case env.String: // todo .. make Error type .. make error construction micro dialect, return the error wrapping error that caused it
				return *env.NewError(val.Value)
			case env.Integer: // todo .. make Error type .. make error construction micro dialect, return the error wrapping error that caused it
				return *env.NewError1(int(val.Value))
			}
			return arg0
		},
	},

	"error": {
		Argsn: 1,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			fmt.Println("ERROR")
			ps.ErrorFlag = true
			switch val := arg0.(type) {
			case env.String: // todo .. make Error type .. make error construction micro dialect, return the error wrapping error that caused it
				return *env.NewError(val.Value)
			case env.Integer: // todo .. make Error type .. make error construction micro dialect, return the error wrapping error that caused it
				return *env.NewError1(int(val.Value))
			}
			return arg0
		},
	},

	"disarm": {
		AcceptFailure: true,
		Argsn:         1,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			ps.FailureFlag = false
			return arg0
		},
	},

	"status": {
		AcceptFailure: true,
		Argsn:         1,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			arg0.Trace("STATUS")
			switch er := arg0.(type) {
			case env.Error:
				return env.Integer{int64(er.Status)}
			}
			return env.NewError("wrong arg")
		},
	},

	"^check": {
		AcceptFailure: true,
		Argsn:         2,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			if ps.FailureFlag {
				ps.ReturnFlag = true
				switch er := arg0.(type) {
				case env.Error: // todo .. make Error type .. make error construction micro dialect, return the error wrapping error that caused it
					switch val := arg1.(type) {
					case env.String: // todo .. make Error type .. make error construction micro dialect, return the error wrapping error that caused it
						return *env.NewError4(0, val.Value, &er, nil)
					case env.Integer: // todo .. make Error type .. make error construction micro dialect, return the error wrapping error that caused it
						return *env.NewError4(int(val.Value), "", &er, nil)
					}
				}
				return env.NewError("error 1")
			}
			return arg0
		},
	},

	"^assert": {
		AcceptFailure: true,
		Argsn:         2,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch cond := arg0.(type) {
			case env.Integer:
				if cond.Value == 0 {
					ps.FailureFlag = true
					ps.ReturnFlag = true
					switch er := arg1.(type) {
					case env.String:
						return *env.NewError(er.Value)
					case env.Integer:
						return *env.NewError1(int(er.Value))
					}
				} else {
					return env.Void{}
				}
				return env.Void{}
			}
			return env.Void{}
		},
	},

	"fix": {
		AcceptFailure: true,
		Argsn:         2,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			if ps.FailureFlag || arg0.Type() == env.ErrorType {
				ps.FailureFlag = false
				// TODO -- create function do_block and call in all cases
				switch bloc := arg1.(type) {
				case env.Block:
					ser := ps.Ser
					ps.Ser = bloc.Series
					EvalBlock(ps)
					ps.Ser = ser
					return ps.Res
				default:
					return env.NewError("Error ..TODO")
				}
			} else {
				return arg0
			}
		},
	},

	"fix-either": {
		AcceptFailure: true,
		Argsn:         3,
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
					return env.NewError("Error ..TODO")
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
					return env.NewError("Error ..TODO")
				}
			}
		},
	},

	"fix-else": {
		AcceptFailure: true,
		Argsn:         2,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			if !(ps.FailureFlag || arg0.Type() == env.ErrorType) {
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
					return env.NewError("Error ..TODO")
				}
			} else {
				return arg0
			}
		},
	},

	"load": {
		Argsn: 1,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.String:
				block, _ := loader.LoadString(s1.Value)
				//ps = env.AddToProgramState(ps, block.Series, genv)
				return block
			default:
				ps.FailureFlag = true
				return env.NewError("Must be string or file TODO")
			}
		},
	},

	// BASIC ENV / Dict FUNCTIONS
	"format": {
		Argsn: 1,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			var r strings.Builder
			switch s1 := arg0.(type) {
			case env.Dict:
				for k, v := range s1.Data {
					r.WriteString(k)
					r.WriteString(":\n\t")
					r.WriteString(fmt.Sprintln(v))
				}
			default:
				return env.String{arg0.Probe(*ps.Idx)}
			}
			return env.String{r.String()}
		},
	},

	// date time functions
	"now": {
		Argsn: 0,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return env.Time{time.Now()}
		},
	},

	// end of date time functions

	"to-context": {
		Argsn: 1,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Dict:

				return util.Dict2Context(ps, s1)
				// make new context with no parent

			default:
				fmt.Println("Error")
			}
			return nil
		},
	},

	"len": {
		Argsn: 1,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Dict:
				return env.Integer{int64(len(s1.Data))}
			case env.Block:
				return env.Integer{int64(s1.Series.Len())}
			case env.Spreadsheet:
				return env.Integer{int64(len(s1.Rows))}
			default:
				fmt.Println("Error")
			}
			return nil
		},
	},
	"ncols": {
		Argsn: 1,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Dict:
			case env.Block:
			case env.Spreadsheet:
				return env.Integer{int64(len(s1.Cols))}
			default:
				fmt.Println("Error")
			}
			return nil
		},
	},
	"keys": {
		Argsn: 1,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Dict:
				keys := make([]env.Object, len(s1.Data))
				i := 0
				for k, _ := range s1.Data {
					keys[i] = env.String{k}
					i++
				}
			case env.Spreadsheet:
				keys := make([]env.Object, len(s1.Cols))
				for i, k := range s1.Cols {
					keys[i] = env.String{k}
				}
				return *env.NewBlock(*env.NewTSeries(keys))
			default:
				fmt.Println("Error")
			}
			return nil
		},
	},

	"sum": {
		Argsn: 2,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			var name string
			switch s1 := arg0.(type) {
			case env.Spreadsheet:
				switch s2 := arg1.(type) {
				case env.Tagword:
					name = ps.Idx.GetWord(s2.Index)
				case env.String:
					name = s2.Value
				default:
					ps.ErrorFlag = true
					return env.NewError("second arg not string")
				}
				r := s1.Sum(name)
				if r.Type() == env.ErrorType {
					ps.ErrorFlag = true
				}
				return r

			default:
				ps.ErrorFlag = true
				return env.NewError("first arg not spreadsheet")
			}
			return nil
		},
	},
}

/*
func isTruthy(arg env.Object) env.Object {
	switch cond := arg.(type) {
	case env.Integer:
		return cond.Value != 0
	case env.String:
		return cond.Value != ""
	default:
		// if it's neither we just return error for now
		ps.FailureFlag = true
		return env.NewError("Error determining if truty")
	}
}
*/

func RegisterBuiltins(ps *env.ProgramState) {
	RegisterBuiltins2(builtins, ps)
	RegisterBuiltins2(Builtins_io, ps)
	RegisterBuiltins2(Builtins_web, ps)
	RegisterBuiltins2(Builtins_sxml, ps)
	RegisterBuiltins2(Builtins_sqlite, ps)
	RegisterBuiltins2(Builtins_gtk, ps)
	RegisterBuiltins2(Builtins_validation, ps)
	RegisterBuiltins2(Builtins_ps, ps)
	RegisterBuiltins2(Builtins_nats, ps)
	RegisterBuiltins2(Builtins_qframe, ps)
	RegisterBuiltins2(Builtins_webview, ps)
	RegisterBuiltins2(Builtins_json, ps)
	RegisterBuiltins2(Builtins_stackless, ps)
	RegisterBuiltins2(Builtins_eyr, ps)
	RegisterBuiltins2(Builtins_conversion, ps)
	RegisterBuiltins2(Builtins_nng, ps)
	RegisterBuiltins2(Builtins_http, ps)
	// RegisterBuiltins2(Builtins_psql, ps)
}

func RegisterBuiltins2(builtins map[string]*env.Builtin, ps *env.ProgramState) {
	for k, v := range builtins {
		bu := env.NewBuiltin(v.Fn, v.Argsn, v.AcceptFailure, v.Pure)
		registerBuiltin(ps, k, *bu)
	}
}

func registerBuiltin(ps *env.ProgramState, word string, builtin env.Builtin) {
	// indexWord
	// TODO -- this with string separator is a temporary way of how we define generic builtins
	// in future a map will probably not be a map but an array and builtin will also support the Kind value

	idxk := 0
	if strings.Index(word, "//") > 0 {
		temp := strings.Split(word, "//")
		word = temp[1]
		idxk = ps.Idx.IndexWord(temp[0])
	}
	idxw := ps.Idx.IndexWord(word)
	// set global word with builtin
	if idxk == 0 {
		ps.Ctx.Set(idxw, builtin)
		if builtin.Pure {
			ps.PCtx.Set(idxw, builtin)
		}

	} else {
		ps.Gen.Set(idxk, idxw, builtin)
	}
}
