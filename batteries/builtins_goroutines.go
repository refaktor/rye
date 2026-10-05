//go:build !b_tiny
// +build !b_tiny

package batteries

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/refaktor/rye/env"
	"github.com/refaktor/rye/evaldo"

	"github.com/jinzhu/copier"
)

// RyeMutex wraps sync.Mutex with state tracking to prevent fatal errors on unlocking unlocked mutex
type RyeMutex struct {
	mu     sync.Mutex
	locked atomic.Bool
}

func (m *RyeMutex) Lock() {
	m.mu.Lock()
	m.locked.Store(true)
}

func (m *RyeMutex) Unlock() error {
	if !m.locked.CompareAndSwap(true, false) {
		return fmt.Errorf("unlock of unlocked mutex")
	}
	m.mu.Unlock()
	return nil
}

var Builtins_goroutines = map[string]*env.Builtin{

	//
	// ##### Goroutines & Concurrency ##### ""
	//
	// Example:
	//  ; Simple goroutine with channel communication
	//  ch: channel 1
	//  go fn { } { ch .Send "Hello from goroutine" }
	//  print ch .Read
	//
	//  ; Using waitgroup to coordinate multiple goroutines
	//  wg: waitgroup
	//  results: channel 10
	//  loop 5 { i |
	//    wg .Add 1
	//    go-with i fn { n } { results .Send n * 2 , wg .Done }
	//  }
	//  wg .Wait
	//  results .Close
	//
	//  ; Mutex for safe shared state
	//  counter:: 0
	//  mtx: mutex
	//  go fn { } { mtx .Lock , change! counter + 1 'counter , mtx .Unlock }
	//
	// Tests:
	// equal { x:: 0 , go-with 5 fn { v } { change! v 'x } , sleep 100 , x } 5
	// equal { y:: 0 , go-with "test" fn { v } { change! length? v 'y } , sleep 100 , y } 4
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
					if handler.Argsn != 1 {
						ps.FailureFlag = true
						return evaldo.MakeBuiltinError(ps, "function with exactly 1 argument required", "go-with")
					}
					// Create a copy of the program state for the goroutine
					psTemp := env.ProgramState{}
					err := copier.Copy(&psTemp, &ps)
					if err != nil {
						return evaldo.MakeBuiltinError(ps, fmt.Sprintf("failed to copy program state: %s", err), "go-with")
					}

					// Reset flags for the goroutine state
					psTemp.FailureFlag = false
					psTemp.ErrorFlag = false
					psTemp.ReturnFlag = false

					// Launch goroutine with panic recovery
					go func() {
						defer func() {
							if r := recover(); r != nil {
								// Log panic but don't crash the main program
								fmt.Printf("Goroutine panic in go-with: %v\n", r)
							}
						}()
						evaldo.CallFunction_CollectArgs(handler, &psTemp, arg, false, nil)
					}()

					return arg0
				default:
					ps.FailureFlag = true
					return evaldo.MakeArgError(ps, 2, []env.Type{env.FunctionType}, "go-with")
				}
			default:
				ps.FailureFlag = true
				return evaldo.MakeBuiltinError(ps, "First argument should be object type.", "go-with")
			}
		},
	},

	// Tests:
	// equal { x:: 0 , go fn { } { change! 42 'x } , sleep 100 , x } 42
	// equal { y:: "unchanged" , go fn { } { change! "changed" 'y } , sleep 100 , y } "changed"
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
				if handler.Argsn != 0 {
					ps.FailureFlag = true
					return evaldo.MakeBuiltinError(ps, "function with 0 arguments required", "go")
				}
				// Create a copy of the program state for the goroutine
				psTemp := env.ProgramState{}
				err := copier.Copy(&psTemp, &ps)
				if err != nil {
					return evaldo.MakeBuiltinError(ps, fmt.Sprintf("failed to copy program state: %s", err), "go")
				}

				// Reset flags for the goroutine state
				psTemp.FailureFlag = false
				psTemp.ErrorFlag = false
				psTemp.ReturnFlag = false

				// Launch goroutine with panic recovery
				go func() {
					defer func() {
						if r := recover(); r != nil {
							// Log panic but don't crash the main program
							fmt.Printf("Goroutine panic in go: %v\n", r)
						}
					}()
					evaldo.CallFunction_CollectArgs(handler, &psTemp, nil, false, nil)
				}()

				return arg0
			default:
				ps.FailureFlag = true
				return evaldo.MakeArgError(ps, 1, []env.Type{env.FunctionType}, "go")
			}
		},
	},

	// Tests:
	// equal { ch: channel 1 ,  ch .probe .lg .Send 42 , ch .Read } 42
	// equal { ch: channel 2 , ch .Send 1 , ch .Send 2 , ch .Read } 1
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
				if buflen.Value < 0 || int64(int(buflen.Value)) != buflen.Value {
					ps.FailureFlag = true
					return evaldo.MakeBuiltinError(ps, "buffer size must be a non-negative integer representable on this platform", "channel")
				}
				ch := make(chan *env.Object, int(buflen.Value))
				return *env.NewNative(ps.Idx, ch, "Rye-channel")
			default:
				ps.FailureFlag = true
				return evaldo.MakeArgError(ps, 1, []env.Type{env.IntegerType}, "channel")
			}
		},
	},
	// Tests:
	// equal { ch: channel 1 , ch .Send 123 , ch .Read } 123
	// equal { ch: channel 1 , ch .Send "test" , ch .Read } "test"
	// equal { ch: channel 1 , ch .Close , try { ch .Read } |type? } 'error
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
				// Type assertion with proper error handling
				ch, ok := chn.Value.(chan *env.Object)
				if !ok {
					ps.FailureFlag = true
					return evaldo.MakeBuiltinError(ps, "Invalid channel type", "Rye-channel//Read")
				}

				msg, ok := <-ch
				if ok {
					return *msg
				} else {
					ps.FailureFlag = true
					return evaldo.MakeBuiltinError(ps, "channel closed", "Rye-channel//Read")
				}
			default:
				ps.FailureFlag = true
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "Rye-channel//Read")
			}
		},
	},
	// Tests:
	// equal { ch: channel 1 , ch .Send 42 , ch .Read } 42
	// equal { ch: channel 3 , ch .Send "A" , ch .Send "B" , ch .Read } "A"
	// ; equal { ch: channel 1 , ch .Send 100 , ch } ch
	// Args:
	// * channel: Channel to send the value to
	// * value: Value to send through the channel
	// Returns:
	// * the channel object
	"Rye-channel//Send": {
		Argsn: 2,
		Doc:   "Sends a value through a channel, blocking if the channel is unbuffered or full.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) (result env.Object) {
			// Recover from panic if channel is closed
			defer func() {
				if r := recover(); r != nil {
					ps.FailureFlag = true
					result = *env.NewError("send on closed channel")
				}
			}()
			switch chn := arg0.(type) {
			case env.Native:
				ch, ok := chn.Value.(chan *env.Object)
				if !ok {
					ps.FailureFlag = true
					return evaldo.MakeBuiltinError(ps, "Invalid channel type", "Rye-channel//Send")
				}
				ch <- &arg1
				return arg0
			default:
				ps.FailureFlag = true
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "Rye-channel//Send")
			}
		},
	},

	// Tests:
	// equal { ch: channel 1 , ch .Send 42 , ch .Close , ch |type? } 'native
	// equal { ch: channel 1 , ch .Close , ch .Send 123 |disarm |type? } 'error
	// equal { ch: channel 2 , ch .Close , ch .Read |disarm |type? } 'error
	// Args:
	// * channel: Channel to close
	// Returns:
	// * the closed channel object
	"Rye-channel//Close": {
		Argsn: 1,
		Doc:   "Closes a channel, preventing further sends. Buffered values remain readable until the channel is drained.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) (result env.Object) {
			defer func() {
				if r := recover(); r != nil {
					ps.FailureFlag = true
					result = evaldo.MakeBuiltinError(ps, "channel already closed", "Rye-channel//Close")
				}
			}()
			switch chn := arg0.(type) {
			case env.Native:
				ch, ok := chn.Value.(chan *env.Object)
				if !ok || ch == nil {
					ps.FailureFlag = true
					return evaldo.MakeBuiltinError(ps, "Invalid channel type or nil channel", "Rye-channel//Close")
				}
				close(ch)
				return arg0
			default:
				ps.FailureFlag = true
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "Rye-channel//Close")
			}
		},
	},

	// Tests:
	// equal { mutex |type? } 'native
	// equal { mtx: mutex , mtx .Lock , mtx .Unlock , mtx |type? } 'native
	// Args:
	// * (none)
	// Returns:
	// * a new mutex native object for synchronization
	"mutex": {
		Argsn: 0,
		Doc:   "Creates a new mutex for synchronizing access to shared resources between goroutines.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			mtx := &RyeMutex{}
			return *env.NewNative(ps.Idx, mtx, "Rye-mutex")
		},
	},

	// Tests:
	// equal { waitgroup |type? } 'native
	// equal { wg: waitgroup , wg .Add 1 , wg .Done , wg .Wait , wg |type? } 'native
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
	// equal { mtx: mutex , mtx .Lock , mtx |type? } 'native
	// equal { mtx: mutex , mtx .Lock , mtx .Unlock , mtx .Lock , mtx |type? } 'native
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
				mtx.Value.(*RyeMutex).Lock()
				return arg0
			default:
				ps.FailureFlag = true
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "Rye-mutex//Lock")
			}
		},
	},

	// Tests:
	// equal { mtx: mutex , mtx .Lock , mtx .Unlock , mtx |type? } 'native
	// equal { mtx: mutex , mtx .Unlock |disarm |type? } 'error
	// Args:
	// * mutex: Mutex to release the lock from
	// Returns:
	// * the mutex object, or error if unlocking an unlocked mutex
	"Rye-mutex//Unlock": {
		Argsn: 1,
		Doc:   "Releases the lock on a mutex, allowing other goroutines to acquire it.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch mtx := arg0.(type) {
			case env.Native:
				err := mtx.Value.(*RyeMutex).Unlock()
				if err != nil {
					ps.FailureFlag = true
					return *env.NewError(err.Error())
				}
				return arg0
			default:
				ps.FailureFlag = true
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "Rye-mutex//Unlock")
			}
		},
	},

	// Tests:
	// equal { wg: waitgroup , wg .Add 3 , wg |type? } 'native
	// equal { wg: waitgroup , wg .Add 1 , wg .Add 2 , wg |type? } 'native
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
					return evaldo.MakeArgError(ps, 2, []env.Type{env.IntegerType}, "Rye-waitgroup//Add")
				}
			default:
				ps.FailureFlag = true
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "Rye-waitgroup//Add")
			}
		},
	},

	// Tests:
	// equal { wg: waitgroup , wg .Add 1 , wg .Done , wg |type? } 'native
	// equal { wg: waitgroup , wg .Add 2 , wg .Done , wg .Done , wg |type? } 'native
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
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "Rye-waitgroup//Done")
			}
		},
	},

	// Tests:
	// equal { wg: waitgroup , wg .Add 1 , wg .Done , wg .Wait , wg |type? } 'native
	// equal { wg: waitgroup , wg .Wait , wg |type? } 'native
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
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "Rye-waitgroup//Wait")
			}
		},
	},

	// Tests:
	// equal { ch1: channel 1 , ch2: channel 1 , ch1 .Send 42 , select\fn { ch1 fn { v } { v } ch2 fn { v } { v + 1 } } } 42
	// equal { ch: channel 1 , select\fn { ch fn { v } { v * 2 } fn { } { 999 } } } 999
	// Args:
	// * block: Block containing channel-function pairs and optional default function
	// Returns:
	// * the selected handler's result, or failure if a closed channel is selected
	"select\\fn": {
		Argsn: 1,
		Doc:   "Selects a channel or default function and returns its result. Selecting a drained closed channel fails without calling a handler.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return runChannelSelect(ps, arg0, true)
		},
	},

	// Block handlers run in the caller's context with the received value injected.
	// A default is written as _ { ... }; select { } intentionally blocks forever.
	// Compatibility: returns the handler result, not the specification block.

	"select": {
		Argsn: 1,
		Doc:   "Selects a channel or default block and returns its result. Selecting a drained closed channel fails without running a handler.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return runChannelSelect(ps, arg0, false)
		},
	},
}
