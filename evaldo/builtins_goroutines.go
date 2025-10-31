//go:build !b_tiny
// +build !b_tiny

package evaldo

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/refaktor/rye/env"

	"github.com/jinzhu/copier"
)

var Builtins_goroutines = map[string]*env.Builtin{

	//
	// ##### Goroutines & Concurrency #####
	//

	// Tests:
	// equal { x: 0 , go-with 5 fn { v } { set 'x v } , sleep 100 , x } 5
	// equal { y: 0 , go-with "test" fn { v } { set 'y length? v } , sleep 100 , y } 4
	// Args:
	// * value: Object to pass to the goroutine function
	// * function: Function to execute in a separate goroutine, receives the value as argument
	// Returns:
	// * the original value that was passed to the goroutine
	"go-with": {
		Argsn: 2,
		Doc:   "Executes a function in a separate goroutine, passing the specified value as an argument.",
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

	// Tests:
	// equal { x: 0 , go fn { } { set 'x 42 } , sleep 100 , x } 42
	// equal { y: "unchanged" , go fn { } { set 'y "changed" } , sleep 100 , y } "changed"
	// Args:
	// * function: Function to execute in a separate goroutine (takes no arguments)
	// Returns:
	// * the function that was executed
	"go": {
		Argsn: 1,
		Doc:   "Executes a function in a separate goroutine without passing any arguments.",
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

	// Tests:
	// equal { ch: channel 0 , ch .send 42 , ch .read } 42
	// equal { ch: channel 2 , ch .send 1 , ch .send 2 , ch .read } 1
	// equal { channel 5 |type? } 'native
	// Args:
	// * buffer-size: Integer specifying the channel buffer size (0 for unbuffered)
	// Returns:
	// * a new channel native object with the specified buffer size
	"channel": {
		Argsn: 1,
		Doc:   "Creates a new channel with the specified buffer size for goroutine communication.",
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
	// Tests:
	// equal { ch: channel 1 , ch .send 123 , ch .read } 123
	// equal { ch: channel 0 , ch .send "test" , ch .read } "test"
	// equal { ch: channel 1 , ch .close , try { ch .read } |type? } 'error
	// Args:
	// * channel: Channel to read from
	// Returns:
	// * the next value from the channel, or an error if the channel is closed
	"Rye-channel//Read": {
		Argsn: 1,
		Doc:   "Reads the next value from a channel, blocking until a value is available or the channel is closed.",
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
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Rye-channel//Read")
			}
		},
	},
	// Tests:
	// equal { ch: channel 1 , ch .send 42 , ch .read } 42
	// equal { ch: channel 3 , ch .send "A" , ch .send "B" , ch .read } "A"
	// equal { ch: channel 0 , ch .send 100 , ch } ch
	// Args:
	// * channel: Channel to send the value to
	// * value: Value to send through the channel
	// Returns:
	// * the channel object
	"Rye-channel//Send": {
		Argsn: 2,
		Doc:   "Sends a value through a channel, blocking if the channel is unbuffered or full.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch chn := arg0.(type) {
			case env.Native:
				chn.Value.(chan *env.Object) <- &arg1
				return arg0
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Rye-channel//Send")
			}
		},
	},

	// Tests:
	// equal { ch: channel 1 , ch .send 42 , ch .close , ch } ch
	// equal { ch: channel 0 , ch .close , try { ch .send 123 } |type? } 'error
	// equal { ch: channel 2 , ch .close , try { ch .read } |type? } 'error
	// Args:
	// * channel: Channel to close
	// Returns:
	// * the closed channel object
	"Rye-channel//Close": {
		Argsn: 1,
		Doc:   "Closes a channel, preventing further sends and causing pending/future reads to return an error.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch chn := arg0.(type) {
			case env.Native:
				close(chn.Value.(chan *env.Object))
				return arg0
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Rye-channel//Close")
			}
		},
	},

	// Tests:
	// equal { mutex |type? } 'native
	// equal { mtx: mutex , mtx .lock , mtx .unlock , mtx } mtx
	// Args:
	// * (none)
	// Returns:
	// * a new mutex native object for synchronization
	"mutex": {
		Argsn: 0,
		Doc:   "Creates a new mutex for synchronizing access to shared resources between goroutines.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			var mtx sync.Mutex
			return *env.NewNative(ps.Idx, &mtx, "Rye-mutex")
		},
	},

	// Tests:
	// equal { waitgroup |type? } 'native
	// equal { wg: waitgroup , wg .add 1 , wg .done , wg .wait , wg } wg
	// Args:
	// * (none)
	// Returns:
	// * a new waitgroup native object for coordinating goroutines
	"waitgroup": {
		Argsn: 0,
		Doc:   "Creates a new waitgroup for coordinating multiple goroutines to wait for completion.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			var wg sync.WaitGroup
			return *env.NewNative(ps.Idx, &wg, "Rye-waitgroup")
		},
	},

	// Tests:
	// equal { mtx: mutex , mtx .lock , mtx } mtx
	// equal { mtx: mutex , mtx .lock , mtx .unlock , mtx .lock , mtx } mtx
	// Args:
	// * mutex: Mutex to acquire the lock on
	// Returns:
	// * the mutex object
	"Rye-mutex//Lock": {
		Argsn: 1,
		Doc:   "Acquires the lock on a mutex, blocking until the lock becomes available.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch mtx := arg0.(type) {
			case env.Native:
				mtx.Value.(*sync.Mutex).Lock()
				return arg0
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Rye-mutex//Lock")
			}
		},
	},

	// Tests:
	// equal { mtx: mutex , mtx .lock , mtx .unlock , mtx } mtx
	// equal { mtx: mutex , mtx .unlock , mtx } mtx
	// Args:
	// * mutex: Mutex to release the lock from
	// Returns:
	// * the mutex object
	"Rye-mutex//Unlock": {
		Argsn: 1,
		Doc:   "Releases the lock on a mutex, allowing other goroutines to acquire it.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch mtx := arg0.(type) {
			case env.Native:
				mtx.Value.(*sync.Mutex).Unlock()
				return arg0
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Rye-mutex//Unlock")
			}
		},
	},

	// Tests:
	// equal { wg: waitgroup , wg .add 3 , wg } wg
	// equal { wg: waitgroup , wg .add 1 , wg .add 2 , wg } wg
	// Args:
	// * waitgroup: Waitgroup to add goroutines to
	// * count: Number of goroutines to add to the wait counter
	// Returns:
	// * the waitgroup object
	"Rye-waitgroup//Add": {
		Argsn: 2,
		Doc:   "Adds the specified count to the waitgroup counter, indicating how many goroutines to wait for.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch wg := arg0.(type) {
			case env.Native:
				switch count := arg1.(type) {
				case env.Integer:
					wg.Value.(*sync.WaitGroup).Add(int(count.Value))
					return arg0
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "Rye-waitgroup//Add")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Rye-waitgroup//Add")
			}
		},
	},

	// Tests:
	// equal { wg: waitgroup , wg .add 1 , wg .done , wg } wg
	// equal { wg: waitgroup , wg .add 2 , wg .done , wg .done , wg } wg
	// Args:
	// * waitgroup: Waitgroup to decrement the counter for
	// Returns:
	// * the waitgroup object
	"Rye-waitgroup//Done": {
		Argsn: 1,
		Doc:   "Decrements the waitgroup counter by one, indicating that a goroutine has completed.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch wg := arg0.(type) {
			case env.Native:
				wg.Value.(*sync.WaitGroup).Done()
				return arg0
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Rye-waitgroup//Done")
			}
		},
	},

	// Tests:
	// equal { wg: waitgroup , wg .add 1 , wg .done , wg .wait , wg } wg
	// equal { wg: waitgroup , wg .wait , wg } wg
	// Args:
	// * waitgroup: Waitgroup to wait on
	// Returns:
	// * the waitgroup object after all goroutines have completed
	"Rye-waitgroup//Wait": {
		Argsn: 1,
		Doc:   "Blocks until the waitgroup counter reaches zero, meaning all goroutines have completed.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch wg := arg0.(type) {
			case env.Native:
				wg.Value.(*sync.WaitGroup).Wait()
				return arg0
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Rye-waitgroup//Wait")
			}
		},
	},

	// Tests:
	// equal { ch1: channel 0 , ch2: channel 0 , ch1 .send 42 , select\fn { ch1 fn { v } { v } ch2 fn { v } { v + 1 } } } 42
	// equal { ch: channel 0 , select\fn { ch fn { v } { v * 2 } fn { } { 999 } } } 999
	// Args:
	// * block: Block containing channel-function pairs and optional default function
	// Returns:
	// * the original block argument
	"select\\fn": {
		Argsn: 1,
		Doc:   "Performs a select operation on multiple channels, executing functions when channels are ready or a default function.",
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
				if psTemp.ErrorFlag {
					return psTemp.Res
				}
				// CallFunction(fn, &psTemp, arg, false, nil)

			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.BlockType}, "select")
			}
			return arg0
		},
	},
}
