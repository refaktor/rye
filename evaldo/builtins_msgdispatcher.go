package evaldo

import (
	"fmt"
	"sync"
	"time"

	"github.com/refaktor/rye/env"
)

// MessageDispatcher provides a high-performance message routing system
type MessageDispatcher struct {
	msgChan  chan *env.Object
	handlers map[string]env.Function
	state    env.Dict
	running  bool
	mu       sync.RWMutex
	wg       sync.WaitGroup
}

// NewMessageDispatcher creates a new message dispatcher
func NewMessageDispatcher(bufferSize int, initialState env.Dict) *MessageDispatcher {
	return &MessageDispatcher{
		msgChan:  make(chan *env.Object, bufferSize),
		handlers: make(map[string]env.Function),
		state:    initialState,
		running:  false,
	}
}

// RegisterHandler registers a message handler function
func (md *MessageDispatcher) RegisterHandler(msgType string, handler env.Function) {
	md.mu.Lock()
	defer md.mu.Unlock()
	md.handlers[msgType] = handler
}

// UnregisterHandler removes a message handler
func (md *MessageDispatcher) UnregisterHandler(msgType string) {
	md.mu.Lock()
	defer md.mu.Unlock()
	delete(md.handlers, msgType)
}

// Send sends a message to the dispatcher
func (md *MessageDispatcher) Send(msg env.Object) error {
	if !md.running {
		return fmt.Errorf("dispatcher not running")
	}
	select {
	case md.msgChan <- &msg:
		return nil
	default:
		return fmt.Errorf("message channel full")
	}
}

// GetState returns a copy of the current state
func (md *MessageDispatcher) GetState() env.Dict {
	md.mu.RLock()
	defer md.mu.RUnlock()
	return md.state
}

// SetState updates the state
func (md *MessageDispatcher) SetState(state env.Dict) {
	md.mu.Lock()
	defer md.mu.Unlock()
	md.state = state
}

// Start begins processing messages in a goroutine
func (md *MessageDispatcher) Start(ps *env.ProgramState, updateCallback env.Function) {
	md.mu.Lock()
	if md.running {
		md.mu.Unlock()
		return
	}
	md.running = true
	md.mu.Unlock()

	md.wg.Add(1)
	go md.runLoop(ps, updateCallback)
}

// Stop stops the message processing loop
func (md *MessageDispatcher) Stop() {
	md.mu.Lock()
	if !md.running {
		md.mu.Unlock()
		return
	}
	md.running = false
	md.mu.Unlock()

	close(md.msgChan)
	md.wg.Wait()
}

// runLoop is the main message processing loop
func (md *MessageDispatcher) runLoop(ps *env.ProgramState, updateCallback env.Function) {
	defer md.wg.Done()

	ticker := time.NewTicker(16 * time.Millisecond) // ~60 FPS
	defer ticker.Stop()

	for {
		select {
		case msg, ok := <-md.msgChan:
			if !ok {
				// Channel closed, exit
				return
			}
			md.processMessage(ps, *msg)

		case <-ticker.C:
			// Regular update tick
			md.mu.RLock()
			running := md.running
			md.mu.RUnlock()

			if !running {
				return
			}

			// Call update callback if provided
			if updateCallback.Argsn >= 0 {
				md.mu.RLock()
				state := md.state
				md.mu.RUnlock()

				CallFunctionArgs2(updateCallback, ps, nil, state, nil)
				if newState, ok := ps.Res.(env.Dict); ok {
					md.SetState(newState)
				}
			}
		}
	}
}

// processMessage handles a single message
func (md *MessageDispatcher) processMessage(ps *env.ProgramState, msg env.Object) {
	// Extract message type
	msgDict, ok := msg.(env.Dict)
	if !ok {
		fmt.Printf("Warning: Invalid message format (not a Dict)\n")
		return
	}

	msgTypeObj, ok := msgDict.Data["type"]
	if !ok {
		fmt.Printf("Warning: Message missing 'type' field\n")
		return
	}

	msgTypeStr, ok := msgTypeObj.(env.String)
	if !ok {
		fmt.Printf("Warning: Message 'type' is not a string\n")
		return
	}

	msgType := msgTypeStr.Value

	// Get handler
	md.mu.RLock()
	handler, exists := md.handlers[msgType]
	currentState := md.state
	md.mu.RUnlock()

	if !exists {
		fmt.Printf("Warning: No handler for message type '%s'\n", msgType)
		return
	}

	// Call handler with message and state
	CallFunctionArgs2(handler, ps, msg, currentState, nil)

	// Update state if handler returned a Dict
	if newState, ok := ps.Res.(env.Dict); ok {
		md.SetState(newState)
	}
}

// Builtins for message dispatcher
var Builtins_msgdispatcher = map[string]*env.Builtin{

	// Tests:
	// equal { md: msg-dispatcher 10 dict { "x" 0 } , md |type? } 'native
	// Args:
	// * buffer-size: Integer for channel buffer size
	// * initial-state: Dict containing initial game state
	// Returns:
	// * a new MessageDispatcher native object
	"msg-dispatcher": {
		Argsn: 2,
		Doc:   "Creates a new message dispatcher with specified buffer size and initial state.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch bufSize := arg0.(type) {
			case env.Integer:
				switch initialState := arg1.(type) {
				case env.Dict:
					dispatcher := NewMessageDispatcher(int(bufSize.Value), initialState)
					return *env.NewNative(ps.Idx, dispatcher, "msg-dispatcher")
				default:
					return MakeArgError(ps, 2, []env.Type{env.DictType}, "msg-dispatcher")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType}, "msg-dispatcher")
			}
		},
	},

	// Tests:
	// equal { md: msg-dispatcher 10 dict { "x" 0 } , md .register "test" fn { msg state } { state } , md } md
	// Args:
	// * dispatcher: MessageDispatcher native object
	// * message-type: String identifying the message type
	// * handler: Function that takes (message, state) and returns new state
	// Returns:
	// * the dispatcher object
	"msg-dispatcher//Register": {
		Argsn: 3,
		Doc:   "Registers a message handler function for a specific message type.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch native := arg0.(type) {
			case env.Native:
				dispatcher, ok := native.Value.(*MessageDispatcher)
				if !ok {
					return MakeBuiltinError(ps, "Expected MessageDispatcher", "Rye-msg-dispatcher//Register")
				}

				switch msgType := arg1.(type) {
				case env.String:
					switch handler := arg2.(type) {
					case env.Function:
						if handler.Argsn != 2 {
							return MakeBuiltinError(ps, "Handler must accept 2 arguments (message, state)", "Rye-msg-dispatcher//Register")
						}
						dispatcher.RegisterHandler(msgType.Value, handler)
						return arg0
					default:
						return MakeArgError(ps, 3, []env.Type{env.FunctionType}, "Rye-msg-dispatcher//Register")
					}
				default:
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "Rye-msg-dispatcher//Register")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Rye-msg-dispatcher//Register")
			}
		},
	},

	// Tests:
	// equal { md: msg-dispatcher 10 dict { "x" 0 } , md .register "test" fn { msg state } { state } , md .unregister "test" , md } md
	// Args:
	// * dispatcher: MessageDispatcher native object
	// * message-type: String identifying the message type to unregister
	// Returns:
	// * the dispatcher object
	"msg-dispatcher//Unregister": {
		Argsn: 2,
		Doc:   "Unregisters a message handler for a specific message type.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch native := arg0.(type) {
			case env.Native:
				dispatcher, ok := native.Value.(*MessageDispatcher)
				if !ok {
					return MakeBuiltinError(ps, "Expected MessageDispatcher", "Rye-msg-dispatcher//Unregister")
				}

				switch msgType := arg1.(type) {
				case env.String:
					dispatcher.UnregisterHandler(msgType.Value)
					return arg0
				default:
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "Rye-msg-dispatcher//Unregister")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Rye-msg-dispatcher//Unregister")
			}
		},
	},

	// Tests:
	// equal { md: msg-dispatcher 10 dict { "x" 0 } , msg: dict { "type" "test" "data" dict { } } , md .send msg , md } md
	// Args:
	// * dispatcher: MessageDispatcher native object
	// * message: Dict containing message data (must have "type" field)
	// Returns:
	// * the dispatcher object
	"msg-dispatcher//Send": {
		Argsn: 2,
		Doc:   "Sends a message to the dispatcher for processing.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch native := arg0.(type) {
			case env.Native:
				dispatcher, ok := native.Value.(*MessageDispatcher)
				if !ok {
					return MakeBuiltinError(ps, "Expected MessageDispatcher", "Rye-msg-dispatcher//Send")
				}

				err := dispatcher.Send(arg1)
				if err != nil {
					return MakeBuiltinError(ps, fmt.Sprintf("Send failed: %v", err), "Rye-msg-dispatcher//Send")
				}
				return arg0
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Rye-msg-dispatcher//Send")
			}
		},
	},

	// Tests:
	// equal { md: msg-dispatcher 10 dict { "x" 0 } , md .state |type? } 'dict
	// Args:
	// * dispatcher: MessageDispatcher native object
	// Returns:
	// * the current state Dict
	"msg-dispatcher//State": {
		Argsn: 1,
		Doc:   "Gets the current state from the dispatcher.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch native := arg0.(type) {
			case env.Native:
				dispatcher, ok := native.Value.(*MessageDispatcher)
				if !ok {
					return MakeBuiltinError(ps, "Expected MessageDispatcher", "Rye-msg-dispatcher//State")
				}
				return dispatcher.GetState()
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Rye-msg-dispatcher//State")
			}
		},
	},

	// Tests:
	// equal { md: msg-dispatcher 10 dict { "x" 0 } , md .set-state dict { "x" 100 } , md .state .get "x" } 100
	// Args:
	// * dispatcher: MessageDispatcher native object
	// * new-state: Dict containing the new state
	// Returns:
	// * the dispatcher object
	"msg-dispatcher//Set-state": {
		Argsn: 2,
		Doc:   "Sets the current state in the dispatcher.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch native := arg0.(type) {
			case env.Native:
				dispatcher, ok := native.Value.(*MessageDispatcher)
				if !ok {
					return MakeBuiltinError(ps, "Expected MessageDispatcher", "Rye-msg-dispatcher//Set-state")
				}

				switch newState := arg1.(type) {
				case env.Dict:
					dispatcher.SetState(newState)
					return arg0
				default:
					return MakeArgError(ps, 2, []env.Type{env.DictType}, "Rye-msg-dispatcher//Set-state")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Rye-msg-dispatcher//Set-state")
			}
		},
	},

	// Tests:
	// equal { md: msg-dispatcher 10 dict { "x" 0 } , md .start fn { state } { state } , md .stop , md } md
	// Args:
	// * dispatcher: MessageDispatcher native object
	// * update-callback: Function called every frame with current state, should return new state (optional, can be void)
	// Returns:
	// * the dispatcher object
	"msg-dispatcher//Start": {
		Argsn: 2,
		Doc:   "Starts the message dispatcher loop in a goroutine. Optionally takes an update callback function.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch native := arg0.(type) {
			case env.Native:
				dispatcher, ok := native.Value.(*MessageDispatcher)
				if !ok {
					return MakeBuiltinError(ps, "Expected MessageDispatcher", "Rye-msg-dispatcher//Start")
				}

				var updateCallback env.Function
				switch cb := arg1.(type) {
				case env.Function:
					if cb.Argsn != 1 {
						return MakeBuiltinError(ps, "Update callback must accept 1 argument (state)", "Rye-msg-dispatcher//Start")
					}
					updateCallback = cb
				case env.Void:
					// No callback, use empty function
					updateCallback = env.Function{Argsn: -1}
				default:
					return MakeArgError(ps, 2, []env.Type{env.FunctionType, env.VoidType}, "Rye-msg-dispatcher//Start")
				}

				dispatcher.Start(ps, updateCallback)
				return arg0
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Rye-msg-dispatcher//Start")
			}
		},
	},

	// Tests:
	// equal { md: msg-dispatcher 10 dict { "x" 0 } , md .start fn { state } { state } , md .stop , md } md
	// Args:
	// * dispatcher: MessageDispatcher native object
	// Returns:
	// * the dispatcher object
	"msg-dispatcher//Stop": {
		Argsn: 1,
		Doc:   "Stops the message dispatcher loop and waits for it to finish.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch native := arg0.(type) {
			case env.Native:
				dispatcher, ok := native.Value.(*MessageDispatcher)
				if !ok {
					return MakeBuiltinError(ps, "Expected MessageDispatcher", "Rye-msg-dispatcher//Stop")
				}
				dispatcher.Stop()
				return arg0
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Rye-msg-dispatcher//Stop")
			}
		},
	},
}
