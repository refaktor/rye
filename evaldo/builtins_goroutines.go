//go:build !b_tiny
// +build !b_tiny

package evaldo

// import "C"

import (
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/refaktor/rye/env"

	"github.com/jinzhu/copier"
)

func strimpg() { fmt.Println("") }

var Builtins_goroutines = map[string]*env.Builtin{

	"sleep": {
		Argsn: 1,
		Doc:   "Accepts an integer and Sleeps for given number of miliseconds.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch arg := arg0.(type) {
			case env.Integer:
				time.Sleep(time.Duration(int(arg.Value)) * time.Millisecond)
				return arg
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType}, "sleep")
			}
		},
	},

	"go-with": {
		Argsn: 2,
		Doc:   "TODODOC.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch arg := arg0.(type) {
			case env.Object:
				switch handler := arg1.(type) {
				case env.Function:
					errC := make(chan error)
					go func() {
						ps.FailureFlag = false
						ps.ErrorFlag = false
						ps.ReturnFlag = false
						psTemp := env.ProgramState{}
						err := copier.Copy(&psTemp, &ps)
						if err != nil {
							ps.FailureFlag = true
							ps.ErrorFlag = true
							ps.ReturnFlag = true
							errC <- fmt.Errorf("failed to copy ps: %w", err)
						}
						close(errC)
						CallFunction(handler, &psTemp, arg, false, nil)
					}()
					if err := <-errC; err != nil {
						return MakeBuiltinError(ps, err.Error(), "go-with")
					}
					return arg0
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 2, []env.Type{env.FunctionType}, "go-with")
				}
			default:
				ps.FailureFlag = true
				return MakeBuiltinError(ps, "First argument should be object type.", "go-with")
			}
		},
	},

	"go": {
		Argsn: 1,
		Doc:   "TODODOC.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch handler := arg0.(type) {
			case env.Function:
				errC := make(chan error)
				go func() {
					ps.FailureFlag = false
					ps.ErrorFlag = false
					ps.ReturnFlag = false
					psTemp := env.ProgramState{}
					err := copier.Copy(&psTemp, &ps)
					if err != nil {
						ps.FailureFlag = true
						ps.ErrorFlag = true
						ps.ReturnFlag = true
						errC <- fmt.Errorf("failed to copy ps: %w", err)
					}
					close(errC)
					CallFunction(handler, &psTemp, nil, false, nil)
					// CallFunctionArgs2(handler, &psTemp, arg, *env.NewNative(psTemp.Idx, "asd", "Go-server-context"), nil)
				}()
				if err := <-errC; err != nil {
					return MakeBuiltinError(ps, err.Error(), "go")
				}
				return arg0
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.FunctionType}, "go")
			}
		},
	},

	"channel": {
		Argsn: 1,
		Doc:   "TODODOC.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch buflen := arg0.(type) {
			case env.Integer:
				//fmt.Println(str.Value)
				ch := make(chan *env.Object, int(buflen.Value))
				return *env.NewNative(ps.Idx, ch, "Rye-channel")
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.IntegerType}, "channel")
			}
		},
	},
	"Rye-channel//read": {
		Argsn: 1,
		Doc:   "TODODOC.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch chn := arg0.(type) {
			case env.Native:
				msg, ok := <-chn.Value.(chan *env.Object)
				if ok {
					return *msg
				} else {
					return *env.NewError("channel closed")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Rye-channel//read")
			}
		},
	},
	"Rye-channel//send": {
		Argsn: 2,
		Doc:   "TODODOC.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch chn := arg0.(type) {
			case env.Native:
				chn.Value.(chan *env.Object) <- &arg1
				return arg0
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Rye-channel//send")
			}
		},
	},

	"Rye-channel//close": {
		Argsn: 1,
		Doc:   "TODODOC.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch chn := arg0.(type) {
			case env.Native:
				close(chn.Value.(chan *env.Object))
				return arg0
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Rye-channel//close")
			}
		},
	},

	"waitgroup": {
		Argsn: 0,
		Doc:   "Create a waitgroup.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			var wg sync.WaitGroup
			return *env.NewNative(ps.Idx, &wg, "Rye-waitgroup")
		},
	},

	"Rye-waitgroup//add": {
		Argsn: 2,
		Doc:   "TODODOC.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch wg := arg0.(type) {
			case env.Native:
				switch count := arg1.(type) {
				case env.Integer:
					wg.Value.(*sync.WaitGroup).Add(int(count.Value))
					return arg0
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "Rye-waitgroup//add")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Rye-waitgroup//add")
			}
		},
	},

	"Rye-waitgroup//done": {
		Argsn: 1,
		Doc:   "TODODOC.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch wg := arg0.(type) {
			case env.Native:
				wg.Value.(*sync.WaitGroup).Done()
				return arg0
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Rye-waitgroup//done")
			}
		},
	},

	"Rye-waitgroup//wait": {
		Argsn: 1,
		Doc:   "Wait on a waitgroup.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch wg := arg0.(type) {
			case env.Native:
				wg.Value.(*sync.WaitGroup).Wait()
				return arg0
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Rye-waitgroup//wait")
			}
		},
	},

	"select\\fn": {
		Argsn: 1,
		Doc:   "Select on a message on multiple channels or default.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch block := arg0.(type) {
			case env.Block:
				ser := ps.Ser
				ps.Ser = block.Series

				var hasDeafult bool
				var cases []reflect.SelectCase
				var funcs []env.Function
				for ps.Ser.Pos() < ps.Ser.Len() {
					EvalExpression2(ps, false)
					defaultFn, ok := ps.Res.(env.Function)
					// handle default case
					if ok {
						if hasDeafult {
							ps.FailureFlag = true
							return MakeBuiltinError(ps, "select can only have one default case", "select\\fn")
						}
						if defaultFn.Argsn != 0 {
							ps.FailureFlag = true
							return MakeBuiltinError(ps, "function with 0 args required", "select\\fn")
						}
						defaultCase := make(chan struct{})
						close(defaultCase) // close it immediately so it's always ready to receive
						cases = append(cases, reflect.SelectCase{
							Dir:  reflect.SelectRecv,
							Chan: reflect.ValueOf(defaultCase),
						})
						funcs = append(funcs, defaultFn)
						hasDeafult = true
						continue
					}
					// handle regular channel case
					native, ok := ps.Res.(env.Native)
					if !ok {
						ps.FailureFlag = true
						return MakeBuiltinError(ps, "first argument of a case must be a channel", "select\\fn")
					}
					ch, ok := native.Value.(chan *env.Object)
					if !ok {
						ps.FailureFlag = true
						return MakeBuiltinError(ps, "first argument of a case must be a channel", "select\\fn")
					}
					cases = append(cases, reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ch)})

					EvalExpression2(ps, false)
					fn, ok := ps.Res.(env.Function)
					if !ok {
						ps.FailureFlag = true
						return MakeBuiltinError(ps, "second argument of a case must be a function", "select\\fn")
					}
					if fn.Argsn > 1 {
						ps.FailureFlag = true
						return MakeBuiltinError(ps, "function with 0 or 1 arg required", "select\\fn")
					}
					funcs = append(funcs, fn)
				}
				ps.Ser = ser

				chosen, value, recvOK := reflect.Select(cases)
				fn := funcs[chosen]

				psTemp := env.ProgramState{}
				err := copier.Copy(&psTemp, &ps)
				if err != nil {
					ps.FailureFlag = true
					return MakeBuiltinError(ps, fmt.Sprintf("failed to copy ps: %s", err), "select\\fn")
				}
				var arg env.Object = nil
				if recvOK {
					val, ok := value.Interface().(*env.Object)
					if !ok {
						ps.FailureFlag = true
						return MakeBuiltinError(ps, "value from channel is not an object", "select\\fn")
					}
					arg = *val
				}
				if fn.Argsn == 0 {
					arg = nil
				}
				CallFunction(fn, &psTemp, arg, false, nil)

			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.BlockType}, "select\\fn")
			}
			return arg0
		},
	},

	// Modified select\fn code to accept blocks, at the end there will only be one select probably, accepting blocks, functions and get-words
	// Further modified so that default case is prepedned by void _ , like we do in switch function for example
	// Rok did one thing differently than we did so faw. He evaluated the second value in pair, block or fn ...
	// So faw we haven't done this. We didn't evaluate/retrieve except a value was a get-word ... I have to think what is
	// better in the long run. In normal use you don't see the difference, but in more edge cases the difference is big and it
	// must be done consistent across similar functions

	"select": {
		Argsn: 1,
		Doc:   "Select on a message on multiple channels or default.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch block := arg0.(type) {
			case env.Block:
				ser := ps.Ser
				ps.Ser = block.Series

				var hasDeafult bool
				var cases []reflect.SelectCase
				var funcs []env.Block
				for ps.Ser.Pos() < ps.Ser.Len() {
					EvalExpression2(ps, false)
					// handle default case
					switch maybeChan := ps.Res.(type) {
					case env.Native:
						ch, ok := maybeChan.Value.(chan *env.Object)
						if !ok {
							ps.FailureFlag = true
							return MakeBuiltinError(ps, "first argument of a case must be a channel", "select")
						}
						cases = append(cases, reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ch)})

						EvalExpression2(ps, false)
						fn, ok := ps.Res.(env.Block)
						if !ok {
							ps.FailureFlag = true
							return MakeBuiltinError(ps, "second argument of a case must be a block", "select")
						}
						/* if fn.Argsn > 1 {
							ps.FailureFlag = true
							return MakeBuiltinError(ps, "function with 0 or 1 arg required", "select")
						}*/
						funcs = append(funcs, fn)

					case env.Void:
						if hasDeafult {
							ps.FailureFlag = true
							return MakeBuiltinError(ps, "select can only have one default case", "select")
						}
						/* if defaultFn.Argsn != 0 {
							ps.FailureFlag = true
							return MakeBuiltinError(ps, "function with 0 args required", "select")
						} */
						defaultCase := make(chan struct{})
						close(defaultCase) // close it immediately so it's always ready to receive
						cases = append(cases, reflect.SelectCase{
							Dir:  reflect.SelectRecv,
							Chan: reflect.ValueOf(defaultCase),
						})
						EvalExpression2(ps, false)
						fn, ok := ps.Res.(env.Block)
						if !ok {
							ps.FailureFlag = true
							return MakeBuiltinError(ps, "second argument of a case must be a block", "select")
						}
						funcs = append(funcs, fn)
						hasDeafult = true
					}
				}
				ps.Ser = ser

				chosen, value, recvOK := reflect.Select(cases)
				fn := funcs[chosen]

				psTemp := env.ProgramState{}
				err := copier.Copy(&psTemp, &ps)
				if err != nil {
					ps.FailureFlag = true
					return MakeBuiltinError(ps, fmt.Sprintf("failed to copy ps: %s", err), "select")
				}
				var arg env.Object = nil
				if recvOK {
					val, ok := value.Interface().(*env.Object)
					if !ok {
						ps.FailureFlag = true
						return MakeBuiltinError(ps, "value from channel is not an object", "select")
					}
					arg = *val
				}
				/* if fn.Argsn == 0 {
					arg = nil
				}*/
				psTemp.Ser = fn.Series
				EvalBlockInj(&psTemp, arg, true)
				// CallFunction(fn, &psTemp, arg, false, nil)

			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.BlockType}, "select")
			}
			return arg0
		},
	},
}
