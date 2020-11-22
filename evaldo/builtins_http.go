package evaldo

import "C"

import (
	"io"
	//"context"
	"fmt"
	"net/http"
	"rye/env"

	//"time"
	//"golang.org/x/time/rate"
	// "nhooyr.io/websocket"
	//"github.com/gorilla/websocket"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/jinzhu/copier"
)

/*

http-handle "/" fn { w req } { write w "Hello world!" }
ws-handle "/ws" fn { c } { forever { msg: receive c write c "GOT:" + msg }
http-serve ":9000"

new-server ":9000" |with {
	.handle "/" fn { w req } { write w "Hello world!" } ,
	.handle-ws "/ws" fn { c } { forever { msg: receive c write c "GOT:" + msg } } ,
	.serve
}

TODO -- integrate gowabs into this and implement their example first just as handle-ws ... no rye code executed
	if this all works with resetc exits multiple at the same time then implement the callFunction ... but we need to make a local programstate probably

*/

var Builtins_http = map[string]*env.Builtin{

	"new-server": {
		Argsn: 1,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch addr := arg0.(type) {
			case env.String:
				return *env.NewNative(env1.Idx, &http.Server{Addr: addr.Value}, "Go-server")
			default:
				env1.FailureFlag = true
				return *env.NewError("arg 0 should be String")
			}

		},
	},

	"Go-server//serve": {
		Argsn: 1,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch server := arg0.(type) {
			case env.Native:
				server.Value.(*http.Server).ListenAndServe()
				return arg0
			default:
				env1.FailureFlag = true
				return env.NewError("arg 2 should be string %s")
			}

		},
	},

	"Go-server//handle": {
		Argsn: 3,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch path := arg1.(type) {
			case env.String:
				switch handler := arg2.(type) {
				case env.String:
					http.HandleFunc(path.Value, func(w http.ResponseWriter, r *http.Request) {
						fmt.Fprintf(w, handler.Value)
					})
					return arg0
				case env.Function:
					http.HandleFunc(path.Value, func(w http.ResponseWriter, r *http.Request) {
						CallFunctionArgs2(handler, env1, *env.NewNative(env1.Idx, w, "Go-server-response-writer"), *env.NewNative(env1.Idx, r, "Go-server-request"), nil)
					})
					return arg0
				default:
					env1.FailureFlag = true
					return env.NewError("arg1 should be string or function")
				}
			default:
				env1.FailureFlag = true
				return env.NewError("arg0 should be string")
			}
		},
	},

	"Go-server-response-writer//write": {
		Argsn: 2,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch path := arg0.(type) {
			case env.Native:
				switch handler := arg1.(type) {
				case env.String:
					fmt.Fprintf(path.Value.(http.ResponseWriter), handler.Value)
					return arg0
				default:
					env1.FailureFlag = true
					return env.NewError("arg1 should be string")
				}
			default:
				env1.FailureFlag = true
				return env.NewError("arg0 should be native")
			}
		},
	},

	"Go-server//handle-ws": {
		Argsn: 3,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch path := arg1.(type) {
			case env.String:
				switch handler := arg2.(type) {
				case env.Function:
					http.HandleFunc(path.Value, func(w http.ResponseWriter, r *http.Request) {
						conn, _, _, err := ws.UpgradeHTTP(r, w)
						if err != nil {
							fmt.Println("< upgrade http error >")
							// handle error
						}
						go func() {
							defer conn.Close()
							env1.FailureFlag = false
							env1.ErrorFlag = false
							env1.ReturnFlag = false
							fmt.Println("<< Call Function Args 2 >>")
							fmt.Println(env1.Ser.Probe(*env1.Idx))
							psTemp := env.ProgramState{}
							copier.Copy(&psTemp, &env1)
							CallFunctionArgs2(handler, &psTemp, *env.NewNative(psTemp.Idx, conn, "Go-server-websocket"), *env.NewNative(psTemp.Idx, "asd", "Go-server-context"), nil)
							/*							for {
														msg, op, err := wsutil.ReadClientData(conn)
														if err != nil {
															// handle error
														}
														err = wsutil.WriteServerMessage(conn, op, msg)
														if err != nil {
															// handle error
														}
													} */
						}()
					})
					return arg0
				default:
					env1.FailureFlag = true
					return env.NewError("arg1 should be string or function")
				}
			default:
				env1.FailureFlag = true
				return env.NewError("arg0 should be string")
			}
		},
	},

	"Go-server-websocket//read": {
		Argsn: 1,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch conn := arg0.(type) {
			case env.Native:
				fmt.Println("BEFORE READ")
				//					_, msg, err := path.Value.(*websocket.Conn).Read(ctx.Value.(context.Context))
				msg, op, err := wsutil.ReadClientData(conn.Value.(io.ReadWriter))
				fmt.Println("AFTER READ")
				fmt.Println(op)
				if err != nil {
					fmt.Println(err.Error())
					fmt.Println("READ ERROR !!!!")
					env1.ReturnFlag = true
					env1.FailureFlag = true
					env1.ErrorFlag = true
					return env.NewError("arg1 should be string 211s")
				}
				// fmt.Fprintf(path.Value.(http.ResponseWriter), handler.Value)
				return env.String{string(msg)}
			default:
				env1.FailureFlag = true
				return env.NewError("arg0 should be native")
			}
		},
	},

	"Go-server-websocket//write": {
		Argsn: 2,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch sock := arg0.(type) {
			case env.Native:
				switch message := arg1.(type) {
				case env.String:
					err := wsutil.WriteServerMessage(sock.Value.(io.Writer), ws.OpText, []byte(message.Value))
					//sock_ := sock.Value.(*websocket.Conn)
					//ctx_ := ctx.Value.(context.Context)
					//err := sock_.Write(ctx_, websocket.MessageText, []byte(message.Value))
					if err != nil {
						fmt.Println("YYOOYOYOYOYOYOYYOYOYOOY")
						env1.FailureFlag = true
						return env.NewError(err.Error())
					}
					return arg1
				default:
					env1.FailureFlag = true
					return env.NewError("arg1 should be string")
				}
			default:
				env1.FailureFlag = true
				return env.NewError("arg0 should be native")
			}
		},
	},

	/*	"Go-server//handle-ws--old": {
			Argsn: 3,
			Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
				switch path := arg1.(type) {
				case env.String:
					switch handler := arg2.(type) {
					case env.Function:
						http.HandleFunc(path.Value, func(w http.ResponseWriter, r *http.Request) {
							fmt.Println("NEW WSOCK")
							c, err := websocket.Accept(w, r, nil)
							fmt.Println("NEW WSOCK")
							if err != nil {
								fmt.Println("NEW WSOCK ERROR")
								env1.ReturnFlag = true
								env1.FailureFlag = true
								return // env.NewError("arg1 should be string or function")
							}
							defer c.Close(websocket.StatusInternalError, "the sky is fallingaa")
							//defer c.Close(websocket.StatusNormalClosure, "bye!")

							// ctx, cancel := context.WithTimeout(r.Context(), time.Second*10)
							//defer cancel()
							env1.FailureFlag = false
							env1.ErrorFlag = false
							env1.ReturnFlag = false
							fmt.Println("<< Call Function Args 2 >>")
							fmt.Println(c)
							// fmt.Println(ctx)
							fmt.Println("<< // Call Function Args 2 >>")
							CallFunctionArgs2(handler, env1, *env.NewNative(env1.Idx, c, "Go-server-websocket"), *env.NewNative(env1.Idx, r.Context(), "Go-server-context"), nil)
						})
						return arg0
					default:
						env1.FailureFlag = true
						return env.NewError("arg1 should be string or function")
					}
				default:
					env1.FailureFlag = true
					return env.NewError("arg0 should be string")
				}
			},
		},

		"Go-server-websocket//read": {
			Argsn: 2,
			Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
				switch path := arg0.(type) {
				case env.Native:
					switch ctx := arg1.(type) {
					case env.Native:
						fmt.Println("BEFORE READ")
						_, msg, err := path.Value.(*websocket.Conn).Read(ctx.Value.(context.Context))
						fmt.Println("AFTER READ")
						if err != nil {
							fmt.Println(err.Error())
							fmt.Println("READ ERROR !!!!")
							env1.ReturnFlag = true
							env1.FailureFlag = true
							env1.ErrorFlag = true
							return env.NewError("arg1 should be string 211s")
						}
						// fmt.Fprintf(path.Value.(http.ResponseWriter), handler.Value)
						return env.String{string(msg)}
					default:
						env1.FailureFlag = true
						return env.NewError("arg1 should be string")
					}
				default:
					env1.FailureFlag = true
					return env.NewError("arg0 should be native")
				}
			},
		},

		"Go-server-websocket//write": {
			Argsn: 3,
			Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
				switch sock := arg0.(type) {
				case env.Native:
					switch ctx := arg1.(type) {
					case env.Native:
						switch message := arg2.(type) {
						case env.String:
							sock_ := sock.Value.(*websocket.Conn)
							ctx_ := ctx.Value.(context.Context)
							err := sock_.Write(ctx_, websocket.MessageText, []byte(message.Value))
							if err != nil {
								env1.FailureFlag = true
								return env.NewError(err.Error())
							}
							return arg1
						default:
							env1.FailureFlag = true
							return env.NewError("arg1 should be string")
						}
					default:
						env1.FailureFlag = true
						return env.NewError("arg0 should be native")
					}
				default:
					env1.FailureFlag = true
					return env.NewError("arg0 should be native")
				}
			},
		},
	*/
}
