// +build b_http

package evaldo

import "C"

import (
	"io"
	//"context"
	"fmt"
	"net/http"
	"net/url"

	"rye/env"

	//"time"
	//"golang.org/x/time/rate"
	// "nhooyr.io/websocket"
	//"github.com/gorilla/websocket"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/gorilla/sessions"
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
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
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
						ps.FailureFlag = false
						ps.ErrorFlag = false
						ps.ReturnFlag = false
						psTemp := env.ProgramState{}
						copier.Copy(&psTemp, &ps)
						CallFunctionArgs2(handler, ps, *env.NewNative(ps.Idx, w, "Go-server-response-writer"), *env.NewNative(ps.Idx, r, "Go-server-request"), nil)
					})
					return arg0
				default:
					ps.FailureFlag = true
					return env.NewError("arg1 should be string or function")
				}
			default:
				ps.FailureFlag = true
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

	"Go-server-response-writer//set-content-type": {
		Argsn: 2,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch path := arg0.(type) {
			case env.Native:
				switch handler := arg1.(type) {
				case env.String:
					path.Value.(http.ResponseWriter).Header().Set("Content-Type", handler.Value)
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

	"Go-server-response-writer//write-header": {
		Argsn: 2,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch w := arg0.(type) {
			case env.Native:
				switch code := arg1.(type) {
				case env.Integer:
					w.Value.(http.ResponseWriter).WriteHeader(int(code.Value))
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
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
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
						ps.FailureFlag = true
						return env.NewError(err.Error())
					}
					return arg1
				default:
					ps.FailureFlag = true
					return env.NewError("arg1 should be string")
				}
			default:
				ps.FailureFlag = true
				return env.NewError("arg0 should be native")
			}
		},
	},

	/*	"Go-server-request//form?": {
		Argsn: 2,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch ctx := arg0.(type) {
			case env.Native:
				switch key := arg1.(type) {
				case env.String:
					return env.String{ctx.Value.(echo.Context).FormValue(key.Value)}
				default:
					return env.NewError("second arg should be string, got %s")
				}
			default:
				return env.NewError("first arg should be echo.Context, got %s")
			}
		},
	},*/

	"Go-server-request//query?": {
		Argsn: 2,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			//fmt.Println("YOYOYOYOYOYO ------------- - - -  --")
			//return env.String{"QUERY - VAL"}
			switch req := arg0.(type) {
			case env.Native:
				switch key := arg1.(type) {
				case env.String:

					vals, ok := req.Value.(*http.Request).URL.Query()[key.Value]

					if !ok || len(vals[0]) < 1 {
						ps.FailureFlag = true
						return env.NewError("key is missing")
					}
					//return env.NewError("XOSADOSADOA SDAS DO" + key.Value)
					return env.String{vals[0]}
				default:
					ps.FailureFlag = true
					return env.NewError("second arg should be String")
				}
			default:
				ps.FailureFlag = true
				return env.NewError("first arg should be Native")
			}
		},
	},

	"Go-server-request//url?": {
		Argsn: 1,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch req := arg0.(type) {
			case env.Native:
				vals := req.Value.(*http.Request).URL
				return *env.NewNative(ps.Idx, vals, "Go-server-url")
			default:
				ps.FailureFlag = true
				return env.NewError("first arg should be Native")
			}
		},
	},

	"Go-server-url//path?": {
		Argsn: 1,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch req := arg0.(type) {
			case env.Native:
				val := req.Value.(*url.URL).Path
				return env.String{val}
			default:
				ps.FailureFlag = true
				return env.NewError("first arg should be Native")
			}
		},
	},

	"Go-server-request//cookie-val?": {
		Argsn: 2,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch req := arg0.(type) {
			case env.Native:
				switch key := arg1.(type) {
				case env.String:

					cookie, err := req.Value.(*http.Request).Cookie(key.Value)

					if err != nil {
						ps.FailureFlag = true
						return env.NewError("cookie key is missing")
					}
					return env.String{cookie.Value}
				default:
					ps.FailureFlag = true
					return env.NewError("second arg should be String")
				}
			default:
				ps.FailureFlag = true
				return env.NewError("first arg should be Native")
			}
		},
	},

	"Go-server-request//form?": {
		Argsn: 2,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch req := arg0.(type) {
			case env.Native:
				switch key := arg1.(type) {
				case env.String:
					r := req.Value.(*http.Request)
					r.ParseForm()

					val := r.FormValue(key.Value)

					if len(val) < 1 {
						ps.FailureFlag = true
						return env.NewError("value is missing")
					}
					return env.String{val}
				default:
					ps.FailureFlag = true
					return env.NewError("second arg should be String")
				}
			default:
				ps.FailureFlag = true
				return env.NewError("first arg should be Native")
			}
		},
	},

	"Go-server-request//full-form?": {
		Argsn: 1,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch req := arg0.(type) {
			case env.Native:
				r := req.Value.(*http.Request)
				r.ParseForm()

				dict := make(map[string]interface{})

				for key, val := range r.Form {
					dict[key] = val[0]
				}

				return *env.NewDict(dict)
			default:
				ps.FailureFlag = true
				return env.NewError("first arg should be Native")
			}
		},
	},

	"new-cookie-store": {
		Argsn: 1,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch addr := arg0.(type) {
			case env.String:
				return *env.NewNative(env1.Idx, sessions.NewCookieStore([]byte(addr.Value)), "Http-cookie-store")
			default:
				env1.FailureFlag = true
				return *env.NewError("arg 0 should be String")
			}
		},
	},

	"Http-cookie-store//get": {
		Argsn: 3,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			//fmt.Println("asdsad")
			switch store := arg0.(type) {
			case env.Native:
				switch r := arg1.(type) {
				case env.Native:
					switch name := arg2.(type) {
					case env.String:
						//fmt.Println("asdsad")
						session, err := store.Value.(*sessions.CookieStore).Get(r.Value.(*http.Request), name.Value)
						if err != nil {
							env1.FailureFlag = true
							return env.NewError("can't get session: " + err.Error())
						}
						//fmt.Println("asdsad 1")
						return *env.NewNative(env1.Idx, session, "Http-session")
					default:
						//fmt.Println("asdsad 2")
						env1.FailureFlag = true
						return *env.NewError("arg 0 should be String")
					}
				default:
					//fmt.Println("asdsad 3")
					env1.FailureFlag = true
					return *env.NewError("arg 0 should be String")
				}
			default:
				//fmt.Println("asdsad 4")
				env1.FailureFlag = true
				return *env.NewError("arg 0 should be String")
			}
		},
	},

	"Http-session//set": {
		Argsn: 3,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			//fmt.Println("YOYOYOYOYOYO ------------- - - -  --")
			//return env.String{"QUERY - VAL"}
			switch session := arg0.(type) {
			case env.Native:
				switch key := arg1.(type) {
				case env.String:
					switch val := arg2.(type) {
					case env.String:
						//return env.NewError("XOSADOSADOA SDAS DO" + key.Value)
						session.Value.(*sessions.Session).Values[key.Value] = val.Value
						return arg0
					case env.Integer:
						session.Value.(*sessions.Session).Values[key.Value] = int(val.Value)
						return arg0
					default:
						return env.NewError("second arg should be string, got %s")
					}
					//return env.NewError("XOSADOSADOA SDAS DO" + key.Value)
					return arg2 // env.String{ctx.Value.(echo.Context).QueryParam(key.Value)}
				default:
					return env.NewError("second arg should be string, got %s")
				}
			default:
				return env.NewError("first arg should be echo.Context, got %s")
			}
		},
	},

	"Http-session//get": {
		Argsn: 2,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			//return env.String{"QUERY - VAL"}
			switch session := arg0.(type) {
			case env.Native:
				switch key := arg1.(type) {
				case env.String:
					val := session.Value.(*sessions.Session).Values[key.Value]
					if val != nil {
						switch val2 := val.(type) {
						case int:
							return env.Integer{int64(val2)}
						case string:
							return env.String{val2}
						case env.Object:
							return val2
						default:
							env1.FailureFlag = true
							return env.NewError("unknown type")
						}
					} else {
						env1.FailureFlag = true
						return env.NewError("value is empty")
					}
				default:
					return env.NewError("second arg should be string, got %s")
				}
			default:
				return env.NewError("first arg should be echo.Context, got %s")
			}
		},
	},

	"Http-session//save": {
		Argsn: 3,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch session := arg0.(type) {
			case env.Native:
				switch r := arg1.(type) {
				case env.Native:
					switch w := arg2.(type) {
					case env.Native:
						err := session.Value.(*sessions.Session).Save(r.Value.(*http.Request), w.Value.(http.ResponseWriter))
						if err != nil {
							env1.FailureFlag = true
							return env.NewError("can't save: " + err.Error())
						}
						return env.Integer{1}
					default:
						return env.NewError("second arg should be string, got %s")
					}
				default:
					return env.NewError("second arg should be string, got %s")
				}
			default:
				return env.NewError("first arg should be echo.Context, got %s")
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
