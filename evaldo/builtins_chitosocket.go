//go:build !no_chitosocket
// +build !no_chitosocket

package evaldo

import (
	"encoding/json"
	"net/http"

	"github.com/refaktor/rye/env"
	"github.com/sairash/chitosocket"
)

// Builtins_chitosocket provides WebSocket server functionality with room support
var Builtins_chitosocket = map[string]*env.Builtin{

	//
	// ##### ChitoSocket Server Functions ##### "WebSocket server with room support"
	//

	// Tests:
	// equal { chitosocket 1 |type? } 'native
	// Args:
	// * cores: Integer number of CPU cores for the socket server (0 for all available)
	// Returns:
	// * native Chitosocket-server object
	"chitosocket": {
		Argsn: 1,
		Doc:   "Creates a new ChitoSocket WebSocket server with the specified number of CPU cores.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch cores := arg0.(type) {
			case env.Integer:
				socket, err := chitosocket.StartUp(int(cores.Value))
				if err != nil {
					ps.FailureFlag = true
					return MakeBuiltinError(ps, err.Error(), "chitosocket")
				}
				return *env.NewNative(ps.Idx, socket, "Chitosocket-server")
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.IntegerType}, "chitosocket")
			}
		},
	},

	// Tests:
	// Args:
	// * server: Native Chitosocket-server object
	// Returns:
	// * Integer 1 on success
	"Chitosocket-server//close": {
		Argsn: 1,
		Doc:   "Closes the ChitoSocket server and releases resources.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch server := arg0.(type) {
			case env.Native:
				sock := server.Value.(*chitosocket.Socket)
				err := sock.Close()
				if err != nil {
					ps.FailureFlag = true
					return MakeBuiltinError(ps, err.Error(), "Chitosocket-server//close")
				}
				return *env.NewInteger(1)
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Chitosocket-server//close")
			}
		},
	},

	// Tests:
	// Args:
	// * server: Native Chitosocket-server object
	// * event: String event name (e.g., "connected", "disconnect", "chat")
	// * handler: Function that receives subscriber and data block arguments
	// Returns:
	// * the server object for method chaining
	"Chitosocket-server//on": {
		Argsn: 3,
		Doc:   "Registers an event handler for the specified event name. Handler receives subscriber and data as arguments.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch server := arg0.(type) {
			case env.Native:
				switch eventName := arg1.(type) {
				case env.String:
					switch handler := arg2.(type) {
					case env.Function:
						sock := server.Value.(*chitosocket.Socket)
						sock.On[eventName.Value] = func(sub *chitosocket.Subscriber, data []byte) {
							// Reset program state flags for clean handler execution
							ps.FailureFlag = false
							ps.ErrorFlag = false
							ps.ReturnFlag = false

							// Create subscriber native and data string
							subNative := *env.NewNative(ps.Idx, sub, "Chitosocket-subscriber")
							dataStr := *env.NewString(string(data))

							// Call the Rye handler function
							CallFunctionArgs2(handler, ps, subNative, dataStr, nil)
						}
						return arg0
					default:
						ps.FailureFlag = true
						return MakeArgError(ps, 3, []env.Type{env.FunctionType}, "Chitosocket-server//on")
					}
				case env.Word:
					switch handler := arg2.(type) {
					case env.Function:
						sock := server.Value.(*chitosocket.Socket)
						eventNameStr := ps.Idx.GetWord(eventName.Index)
						sock.On[eventNameStr] = func(sub *chitosocket.Subscriber, data []byte) {
							ps.FailureFlag = false
							ps.ErrorFlag = false
							ps.ReturnFlag = false

							subNative := *env.NewNative(ps.Idx, sub, "Chitosocket-subscriber")
							dataStr := *env.NewString(string(data))

							CallFunctionArgs2(handler, ps, subNative, dataStr, nil)
						}
						return arg0
					default:
						ps.FailureFlag = true
						return MakeArgError(ps, 3, []env.Type{env.FunctionType}, "Chitosocket-server//on")
					}
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 2, []env.Type{env.StringType, env.WordType}, "Chitosocket-server//on")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Chitosocket-server//on")
			}
		},
	},

	// Tests:
	// Args:
	// * server: Native Chitosocket-server object
	// * request: Native Go-server-request object
	// * writer: Native Go-server-response-writer object
	// Returns:
	// * Block containing [connection, response-writer, subscriber] or error
	"Chitosocket-server//upgrade-connection": {
		Argsn: 3,
		Doc:   "Upgrades an HTTP connection to WebSocket. Returns a block with connection, response writer, and subscriber.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch server := arg0.(type) {
			case env.Native:
				switch req := arg1.(type) {
				case env.Native:
					switch w := arg2.(type) {
					case env.Native:
						sock := server.Value.(*chitosocket.Socket)

						// Type assert to *http.Request and http.ResponseWriter
						httpReq, ok1 := req.Value.(*http.Request)
						httpWriter, ok2 := w.Value.(http.ResponseWriter)

						if !ok1 || !ok2 {
							ps.FailureFlag = true
							return MakeBuiltinError(ps, "Invalid HTTP request or response writer types", "Chitosocket-server//upgrade-connection")
						}

						conn, rw, sub, err := sock.UpgradeConnection(httpReq, httpWriter)
						if err != nil {
							ps.FailureFlag = true
							return MakeBuiltinError(ps, err.Error(), "Chitosocket-server//upgrade-connection")
						}

						// Return block with connection, response writer, and subscriber
						result := make([]env.Object, 3)
						result[0] = *env.NewNative(ps.Idx, conn, "Chitosocket-connection")
						result[1] = *env.NewNative(ps.Idx, rw, "Chitosocket-response-writer")
						result[2] = *env.NewNative(ps.Idx, sub, "Chitosocket-subscriber")

						return *env.NewBlock(*env.NewTSeries(result))
					default:
						ps.FailureFlag = true
						return MakeArgError(ps, 3, []env.Type{env.NativeType}, "Chitosocket-server//upgrade-connection")
					}
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 2, []env.Type{env.NativeType}, "Chitosocket-server//upgrade-connection")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Chitosocket-server//upgrade-connection")
			}
		},
	},

	// Tests:
	// Args:
	// * server: Native Chitosocket-server object
	// * subscriber: Native Chitosocket-subscriber object
	// * event: String event name
	// * data: String or Dict data to send
	// Returns:
	// * Integer 1 on success
	"Chitosocket-server//emit-direct": {
		Argsn: 4,
		Doc:   "Emits an event directly to a specific subscriber with the given data.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch server := arg0.(type) {
			case env.Native:
				switch sub := arg1.(type) {
				case env.Native:
					switch eventName := arg2.(type) {
					case env.String:
						sock := server.Value.(*chitosocket.Socket)
						subscriber := sub.Value.(*chitosocket.Subscriber)

						var dataBytes []byte
						switch data := arg3.(type) {
						case env.String:
							dataBytes = []byte(data.Value)
						case env.Dict:
							jsonData, err := json.Marshal(data.Data)
							if err != nil {
								ps.FailureFlag = true
								return MakeBuiltinError(ps, err.Error(), "Chitosocket-server//emit-direct")
							}
							dataBytes = jsonData
						default:
							dataBytes = []byte(arg3.Inspect(*ps.Idx))
						}

						sock.EmitDirect(subscriber, eventName.Value, dataBytes)
						return *env.NewInteger(1)
					default:
						ps.FailureFlag = true
						return MakeArgError(ps, 3, []env.Type{env.StringType}, "Chitosocket-server//emit-direct")
					}
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 2, []env.Type{env.NativeType}, "Chitosocket-server//emit-direct")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Chitosocket-server//emit-direct")
			}
		},
	},

	// Tests:
	// Args:
	// * server: Native Chitosocket-server object
	// * event: String event name
	// * data: String or Dict data to broadcast
	// * room: String room name to broadcast to
	// Returns:
	// * Integer 1 on success
	"Chitosocket-server//emit": {
		Argsn: 4,
		Doc:   "Emits an event to all subscribers in a specific room (excluding none).",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch server := arg0.(type) {
			case env.Native:
				switch eventName := arg1.(type) {
				case env.String:
					switch room := arg3.(type) {
					case env.String:
						sock := server.Value.(*chitosocket.Socket)

						var dataBytes []byte
						switch data := arg2.(type) {
						case env.String:
							dataBytes = []byte(data.Value)
						case env.Dict:
							jsonData, err := json.Marshal(data.Data)
							if err != nil {
								ps.FailureFlag = true
								return MakeBuiltinError(ps, err.Error(), "Chitosocket-server//emit")
							}
							dataBytes = jsonData
						default:
							dataBytes = []byte(arg2.Inspect(*ps.Idx))
						}

						// Emit to room with nil excludeSub (no exclusion)
						sock.Emit(eventName.Value, nil, dataBytes, room.Value)
						return *env.NewInteger(1)
					default:
						ps.FailureFlag = true
						return MakeArgError(ps, 4, []env.Type{env.StringType}, "Chitosocket-server//emit")
					}
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "Chitosocket-server//emit")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Chitosocket-server//emit")
			}
		},
	},

	// Tests:
	// Args:
	// * server: Native Chitosocket-server object
	// * event: String event name
	// * data: String or Dict data to broadcast
	// Returns:
	// * Integer 1 on success
	"Chitosocket-server//broadcast": {
		Argsn: 3,
		Doc:   "Broadcasts an event to all connected subscribers.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch server := arg0.(type) {
			case env.Native:
				switch eventName := arg1.(type) {
				case env.String:
					sock := server.Value.(*chitosocket.Socket)

					var dataBytes []byte
					switch data := arg2.(type) {
					case env.String:
						dataBytes = []byte(data.Value)
					case env.Dict:
						jsonData, err := json.Marshal(data.Data)
						if err != nil {
							ps.FailureFlag = true
							return MakeBuiltinError(ps, err.Error(), "Chitosocket-server//broadcast")
						}
						dataBytes = jsonData
					default:
						dataBytes = []byte(arg2.Inspect(*ps.Idx))
					}

					sock.Broadcast(eventName.Value, dataBytes)
					return *env.NewInteger(1)
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "Chitosocket-server//broadcast")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Chitosocket-server//broadcast")
			}
		},
	},

	// Tests:
	// Args:
	// * server: Native Chitosocket-server object
	// * subscriber: Native Chitosocket-subscriber object
	// * room: String room name
	// Returns:
	// * Integer 1 on success
	"Chitosocket-server//add-to-room": {
		Argsn: 3,
		Doc:   "Adds a subscriber to a room.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch server := arg0.(type) {
			case env.Native:
				switch sub := arg1.(type) {
				case env.Native:
					switch room := arg2.(type) {
					case env.String:
						sock := server.Value.(*chitosocket.Socket)
						subscriber := sub.Value.(*chitosocket.Subscriber)
						sock.AddToRoom(subscriber, room.Value)
						return *env.NewInteger(1)
					default:
						ps.FailureFlag = true
						return MakeArgError(ps, 3, []env.Type{env.StringType}, "Chitosocket-server//add-to-room")
					}
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 2, []env.Type{env.NativeType}, "Chitosocket-server//add-to-room")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Chitosocket-server//add-to-room")
			}
		},
	},

	// Tests:
	// Args:
	// * server: Native Chitosocket-server object
	// * subscriber: Native Chitosocket-subscriber object
	// * room: String room name
	// Returns:
	// * Integer 1 on success
	"Chitosocket-server//remove-from-room": {
		Argsn: 3,
		Doc:   "Removes a subscriber from a room.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch server := arg0.(type) {
			case env.Native:
				switch sub := arg1.(type) {
				case env.Native:
					switch room := arg2.(type) {
					case env.String:
						sock := server.Value.(*chitosocket.Socket)
						subscriber := sub.Value.(*chitosocket.Subscriber)
						sock.RemoveFromRoom(subscriber, room.Value)
						return *env.NewInteger(1)
					default:
						ps.FailureFlag = true
						return MakeArgError(ps, 3, []env.Type{env.StringType}, "Chitosocket-server//remove-from-room")
					}
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 2, []env.Type{env.NativeType}, "Chitosocket-server//remove-from-room")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Chitosocket-server//remove-from-room")
			}
		},
	},

	// Tests:
	// Args:
	// * server: Native Chitosocket-server object
	// * room: String room name
	// Returns:
	// * Integer count of subscribers in the room
	"Chitosocket-server//room-count?": {
		Argsn: 2,
		Doc:   "Returns the number of subscribers in a room.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch server := arg0.(type) {
			case env.Native:
				switch room := arg1.(type) {
				case env.String:
					sock := server.Value.(*chitosocket.Socket)
					count := sock.GetRoomCount(room.Value)
					return *env.NewInteger(int64(count))
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "Chitosocket-server//room-count?")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Chitosocket-server//room-count?")
			}
		},
	},

	//
	// ##### ChitoSocket Subscriber Functions ##### "Working with WebSocket subscribers"
	//

	// Tests:
	// Args:
	// * subscriber: Native Chitosocket-subscriber object
	// Returns:
	// * String subscriber ID
	"Chitosocket-subscriber//id?": {
		Argsn: 1,
		Doc:   "Returns the unique ID of the subscriber.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch sub := arg0.(type) {
			case env.Native:
				subscriber := sub.Value.(*chitosocket.Subscriber)
				return *env.NewString(subscriber.ID)
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Chitosocket-subscriber//id?")
			}
		},
	},

	// Tests:
	// Args:
	// * server: Native Chitosocket-server object
	// * event: String event name
	// Returns:
	// * Function handler or void if not found
	"Chitosocket-server//handler?": {
		Argsn: 2,
		Doc:   "Returns the handler function for a specific event, or void if not set.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch server := arg0.(type) {
			case env.Native:
				switch eventName := arg1.(type) {
				case env.String:
					sock := server.Value.(*chitosocket.Socket)
					if handler, ok := sock.On[eventName.Value]; ok {
						return *env.NewNative(ps.Idx, handler, "Chitosocket-handler")
					}
					return env.Void{}
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "Chitosocket-server//handler?")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Chitosocket-server//handler?")
			}
		},
	},

	// Tests:
	// Args:
	// * handler: Native Chitosocket-handler object
	// * subscriber: Native Chitosocket-subscriber object
	// * data: String data to pass to handler (can be empty)
	// Returns:
	// * Integer 1 on success
	"Chitosocket-handler//call": {
		Argsn: 3,
		Doc:   "Calls a handler function with the given subscriber and data.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch handler := arg0.(type) {
			case env.Native:
				switch sub := arg1.(type) {
				case env.Native:
					handlerFn := handler.Value.(chitosocket.HandlerFunc)
					subscriber := sub.Value.(*chitosocket.Subscriber)

					var dataBytes []byte
					switch data := arg2.(type) {
					case env.String:
						dataBytes = []byte(data.Value)
					case env.Void:
						dataBytes = nil
					default:
						dataBytes = []byte(arg2.Inspect(*ps.Idx))
					}

					handlerFn(subscriber, dataBytes)
					return *env.NewInteger(1)
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 2, []env.Type{env.NativeType}, "Chitosocket-handler//call")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Chitosocket-handler//call")
			}
		},
	},

	// Tests:
	// Args:
	// * server: Native Chitosocket-server object
	// * subscriber: Native Chitosocket-subscriber object
	// Returns:
	// * Integer 1 on success
	"Chitosocket-server//remove": {
		Argsn: 2,
		Doc:   "Removes a subscriber from the server.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch server := arg0.(type) {
			case env.Native:
				switch sub := arg1.(type) {
				case env.Native:
					sock := server.Value.(*chitosocket.Socket)
					subscriber := sub.Value.(*chitosocket.Subscriber)
					err := sock.Remove(subscriber)
					if err != nil {
						ps.FailureFlag = true
						return MakeBuiltinError(ps, err.Error(), "Chitosocket-server//remove")
					}
					return *env.NewInteger(1)
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 2, []env.Type{env.NativeType}, "Chitosocket-server//remove")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Chitosocket-server//remove")
			}
		},
	},
}
