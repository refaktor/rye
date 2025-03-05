//go:build !no_http
// +build !no_http

package evaldo

// import "C"

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"time"

	"github.com/refaktor/rye/env"

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

	"http-server": {
		Argsn: 1,
		Doc:   "Create new http server.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch addr := arg0.(type) {
			case env.String:
				server := &http.Server{Addr: addr.Value, ReadHeaderTimeout: 10 * time.Second}
				return *env.NewNative(ps.Idx, server, "Go-server")
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "http-server")
			}

		},
	},

	"Go-server//serve": {
		Argsn: 1,
		Doc:   "Listen and serve new server.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch server := arg0.(type) {
			case env.Native:
				err := server.Value.(*http.Server).ListenAndServe()
				if err != nil {
					return makeError(ps, err.Error())
				}
				return arg0
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Go-server//serve")
			}

		},
	},

	/* "Go-server//serve\\port": {
		Argsn: 1,
		Doc:   "Listen and serve with port.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch host := arg0.(type) {
			case env.String:
				err := server.Value.(*http.Server).ListenAndServe(host.Value, nil)
				if err != nil {
					return makeError(ps, err.Error())
				}
				return arg0
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "Go-server//serve\\port")
			}

		},
	}, */

	"Go-server//handle": {
		Argsn: 3,
		Doc:   "HTTP handle function for server.",
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
						err := copier.Copy(&psTemp, &ps)
						if err != nil {
							fmt.Println(err.Error())
							// TODO return makeError(ps, err.Error())
						}
						CallFunctionArgs2(handler, ps, *env.NewNative(ps.Idx, w, "Go-server-response-writer"), *env.NewNative(ps.Idx, r, "Go-server-request"), nil)
					})
					return arg0
				case env.Native:
					http.Handle(path.Value, handler.Value.(http.Handler))
					return arg0
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 3, []env.Type{env.StringType, env.FunctionType, env.NativeType}, "Go-server//handle")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 2, []env.Type{env.StringType}, "Go-server//handle")
			}
		},
	},

	"Go-server-response-writer//write": {
		Argsn: 2,
		Doc:   "Http response writer write function.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch path := arg0.(type) {
			case env.Native:
				switch handler := arg1.(type) {
				case env.String:
					fmt.Fprintf(path.Value.(http.ResponseWriter), handler.Value)
					return arg0
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "Go-server-response-writer//write")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Go-server-response-writer//write")
			}
		},
	},

	"Go-server-response-writer//set-content-type": {
		Argsn: 2,
		Doc:   "Set http content type.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch path := arg0.(type) {
			case env.Native:
				switch handler := arg1.(type) {
				case env.String:
					path.Value.(http.ResponseWriter).Header().Set("Content-Type", handler.Value)
					return arg0
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "Go-server-response-writer//set-content-type")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Go-server-response-writer//set-content-type")
			}
		},
	},

	"Go-server-response-writer//set-header": {
		Argsn: 3,
		Doc:   "Set header for http server.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch writer := arg0.(type) {
			case env.Native:
				switch name := arg1.(type) {
				case env.Word:
					name_ := ps.Idx.GetWord(name.Index)
					switch value := arg2.(type) {
					case env.String:
						writer.Value.(http.ResponseWriter).Header().Set(name_, value.Value)
						return arg0
					default:
						ps.FailureFlag = true
						return MakeArgError(ps, 3, []env.Type{env.StringType}, "Go-server-response-writer//set-header")
					}
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 2, []env.Type{env.WordType}, "Go-server-response-writer//set-header")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Go-server-response-writer//set-header")
			}
		},
	},

	"Go-server-response-writer//write-header": {
		Argsn: 2,
		Doc:   "Write header for http server.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch w := arg0.(type) {
			case env.Native:
				switch code := arg1.(type) {
				case env.Integer:
					w.Value.(http.ResponseWriter).WriteHeader(int(code.Value))
					return arg0
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "Go-server-response-writer//write-header")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Go-server-response-writer//write-header")
			}
		},
	},

	"Go-server//handle-ws": {
		Argsn: 3,
		Doc:   "Define handler for websockets",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch path := arg1.(type) {
			case env.String:
				switch handler := arg2.(type) {
				case env.Function:
					http.HandleFunc(path.Value, func(w http.ResponseWriter, r *http.Request) {
						conn, _, _, err := ws.UpgradeHTTP(r, w)
						if err != nil {
							fmt.Println("< upgrade http error >")
							// handle error
							//TODO-FIXME
							//return MakeBuiltinError(ps, "Unable to upgrade HTTP.", "Go-server//handle-ws"), nil
						}
						go func() {
							defer conn.Close()
							ps.FailureFlag = false
							ps.ErrorFlag = false
							ps.ReturnFlag = false
							fmt.Println("<< Call Function Args 2 >>")
							fmt.Println(ps.Ser.PositionAndSurroundingElements(*ps.Idx))
							psTemp := env.ProgramState{}
							err := copier.Copy(&psTemp, &ps)
							if err != nil {
								fmt.Println(err.Error())
								// return makeError(ps, "Can't Listen and Serve")
							}
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
					ps.FailureFlag = true
					return MakeArgError(ps, 1, []env.Type{env.FunctionType}, "Go-server//handle-ws")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "Go-server//handle-ws")
			}
		},
	},

	"Go-server-websocket//read": {
		Argsn: 1,
		Doc:   "Reading websocket.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
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
					ps.ReturnFlag = true
					ps.FailureFlag = true
					ps.ErrorFlag = true
					return MakeBuiltinError(ps, "Error in reading client data.", "Go-server-websocket//read")
				}
				// fmt.Fprintf(path.Value.(http.ResponseWriter), handler.Value)
				return env.NewString(string(msg))
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Go-server-websocket//read")
			}
		},
	},

	"Go-server-websocket//write": {
		Argsn: 2,
		Doc:   "Writing websocket.",
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
						return MakeBuiltinError(ps, "Failed to write server message.", "Go-server-websocket//write")
					}
					return arg1
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 2, []env.Type{env.NativeType}, "Go-server-websocket//write")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Go-server-websocket//write")
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
		Doc:   "Get query parameter from HTTP request.",
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
						return MakeBuiltinError(ps, "Key is missing.", "Go-server-request//query?")
					}
					//return env.NewError("XOSADOSADOA SDAS DO" + key.Value)
					return *env.NewString(vals[0])
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "Go-server-request//query?")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Go-server-request//query?")
			}
		},
	},

	"Go-server-request//url?": {
		Argsn: 1,
		Doc:   "Get URL from HTTP request.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch req := arg0.(type) {
			case env.Native:
				vals := req.Value.(*http.Request).URL
				return *env.NewNative(ps.Idx, vals, "Go-server-url")
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Go-server-request//url?")
			}
		},
	},

	"Go-server-url//path?": {
		Argsn: 1,
		Doc:   "Get path from server url.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch req := arg0.(type) {
			case env.Native:
				val := req.Value.(*url.URL).Path
				return *env.NewString(val)
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Go-server-url//path?")
			}
		},
	},

	"Go-server-request//cookie-val?": {
		Argsn: 2,
		Doc:   "Get cookie value from server request.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch req := arg0.(type) {
			case env.Native:
				switch key := arg1.(type) {
				case env.String:
					cookie, err := req.Value.(*http.Request).Cookie(key.Value)
					if err != nil {
						ps.FailureFlag = true
						return MakeBuiltinError(ps, "Cookie key is missing.", "Go-server-request//cookie-val?")
					}
					return *env.NewString(cookie.Value)
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "Go-server-request//cookie-val?")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Go-server-request//cookie-val?")
			}
		},
	},

	"Go-server-request//form?": {
		Argsn: 2,
		Doc:   "Get form field from server request.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch req := arg0.(type) {
			case env.Native:
				switch key := arg1.(type) {
				case env.String:
					r := req.Value.(*http.Request)
					err := r.ParseForm()
					if err != nil {
						return makeError(ps, err.Error())
					}
					val := r.FormValue(key.Value)
					if len(val) < 1 {
						ps.FailureFlag = true
						return MakeBuiltinError(ps, "Value is missing.", "Go-server-request//form?")
					}
					return *env.NewString(val)
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "Go-server-request//form?")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Go-server-request//form?")
			}
		},
	},

	"Go-server-request//full-form?": {
		Argsn: 1,
		Doc:   "Get full form data as Dict from server request.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch req := arg0.(type) {
			case env.Native:
				r := req.Value.(*http.Request)
				err := r.ParseForm()
				if err != nil {
					return makeError(ps, err.Error())
				}
				dict := make(map[string]any)
				for key, val := range r.Form {
					dict[key] = val[0]
				}
				return *env.NewDict(dict)
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Go-server-request//full-form?")
			}
		},
	},

	"Go-server-request//parse-multipart-form!": {
		Argsn: 1,
		Doc:   "Parse multipart form from server request.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch req := arg0.(type) {
			case env.Native:
				r := req.Value.(*http.Request)
				// 10 MB files max
				err := r.ParseMultipartForm(10 << 20)
				if err != nil {
					return makeError(ps, err.Error())
				}
				return arg0
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Go-server-request//parse-multipart-form!")
			}
		},
	},

	// file-handler: r.form-file "image"
	// dst: create file-handler

	"Go-server-request//form-file?": {
		Argsn: 2,
		Doc:   "Get form file from server request as block with reader and multipart header.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch req := arg0.(type) {
			case env.Native:
				switch key := arg1.(type) {
				case env.String:
					r := req.Value.(*http.Request)
					file, handler, err := r.FormFile(key.Value)
					if err != nil {
						ps.FailureFlag = true
						return MakeBuiltinError(ps, fmt.Sprintf("Failed to read from file: '%v'", err.Error()), "Go-server-request//form-file?")
					}
					pair := make([]env.Object, 2)
					pair[0] = *env.NewNative(ps.Idx, file, "reader")
					pair[1] = *env.NewNative(ps.Idx, handler, "rye-multipart-header")
					return *env.NewBlock(*env.NewTSeries(pair))
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "Go-server-request//form-file?")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Go-server-request//form-file?")
			}
		},
	},

	"rye-multipart-header//filename?": {
		Argsn: 1,
		Doc:   "Get filename from multipart header.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch req := arg0.(type) {
			case env.Native:
				r := req.Value.(*multipart.FileHeader)
				return *env.NewString(r.Filename)
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "rye-multipart-header//filename?")
			}
		},
	},

	"new-cookie-store": {
		Argsn: 1,
		Doc:   "Create new cookie store.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch addr := arg0.(type) {
			case env.String:
				return *env.NewNative(ps.Idx, sessions.NewCookieStore([]byte(addr.Value)), "Http-cookie-store")
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "new-cookie-store")
			}
		},
	},

	"Http-cookie-store//get": {
		Argsn: 3,
		Doc:   "Get http cookie store.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
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
							ps.FailureFlag = true
							errMsg := fmt.Sprintf("Can't get session: %v", err.Error())
							return MakeBuiltinError(ps, errMsg, "Http-cookie-store//get")
						}
						//fmt.Println("asdsad 1")s
						return *env.NewNative(ps.Idx, session, "Http-session")
					default:
						//fmt.Println("asdsad 2")
						ps.FailureFlag = true
						return *env.NewError("arg 0 should be String")
						// return MakeArgError(ps, 3, []env.Type{env.StringType}, "Http-cookie-store//get")
					}
				default:
					//fmt.Println("asdsad 3")
					ps.FailureFlag = true
					return *env.NewError("arg 0 should be String")
					// return MakeArgError(ps, 2, []env.Type{env.NativeType}, "Http-cookie-store//get")
				}
			default:
				//fmt.Println("asdsad 4")
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Http-cookie-store//get")
			}
		},
	},

	"Http-session//set": {
		Argsn: 3,
		Doc:   "Set http session.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
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
						return MakeArgError(ps, 3, []env.Type{env.StringType, env.IntegerType}, "Http-session//set")
					}
					//return env.NewError("XOSADOSADOA SDAS DO" + key.Value)
					// return arg2 // env.String{ctx.Value.(echo.Context).QueryParam(key.Value)}
				default:
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "Http-session//set")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Http-session//set")
			}
		},
	},

	"Http-session//get": {
		Argsn: 2,
		Doc:   "Get http session.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			//return env.String{"QUERY - VAL"}
			switch session := arg0.(type) {
			case env.Native:
				switch key := arg1.(type) {
				case env.String:
					val := session.Value.(*sessions.Session).Values[key.Value]
					if val != nil {
						switch val2 := val.(type) {
						case int:
							return env.NewInteger(int64(val2))
						case string:
							return env.NewString(val2)
						case env.Object:
							return val2
						default:
							ps.FailureFlag = true
							return MakeBuiltinError(ps, "Unknown type.", "Http-session//get")
						}
					} else {
						ps.FailureFlag = true
						return MakeBuiltinError(ps, "Value is empty.", "Http-session//get")
					}
				default:
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "Http-session//get")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Http-session//get")
			}
		},
	},

	"Http-session//save": {
		Argsn: 3,
		Doc:   "Save http session.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch session := arg0.(type) {
			case env.Native:
				switch r := arg1.(type) {
				case env.Native:
					switch w := arg2.(type) {
					case env.Native:
						err := session.Value.(*sessions.Session).Save(r.Value.(*http.Request), w.Value.(http.ResponseWriter))
						if err != nil {
							ps.FailureFlag = true
							errMsg := fmt.Sprintf("Can't save: %v", err.Error())
							return MakeBuiltinError(ps, errMsg, "Http-session//save")
						}
						return *env.NewInteger(1)
					default:
						return MakeArgError(ps, 3, []env.Type{env.NativeType}, "Http-session//save")
					}
				default:
					return MakeArgError(ps, 2, []env.Type{env.NativeType}, "Http-session//save")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Http-session//save")
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

	// Serving static files

	"new-http-dir": {
		Argsn: 1,
		Doc:   "Create new http directory.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch addr := arg0.(type) {
			case env.Uri:
				return *env.NewNative(ps.Idx, http.Dir(addr.Path), "Go-http-dir")
			default:
				return MakeArgError(ps, 1, []env.Type{env.UriType}, "new-http-dir")
			}
		},
	},
	"new-static-handler": {
		Argsn: 1,
		Doc:   "Create new static handler.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch addr := arg0.(type) {
			case env.Uri:
				return *env.NewNative(ps.Idx, http.FileServer(http.Dir(addr.Path)), "Http-handler")
			default:
				return MakeArgError(ps, 1, []env.Type{env.UriType}, "new-static-handler")
			}

		},
	},
	"Http-handler//strip-prefix": {
		Argsn: 2,
		Doc:   "TODODOC.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch prefix := arg1.(type) {
			case env.String:
				switch servr := arg0.(type) {
				case env.Native:
					return *env.NewNative(ps.Idx, http.StripPrefix(prefix.Value, servr.Value.(http.Handler)), "Http-handler")
				default:
					return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Http-handler//strip-prefix")
				}
			default:
				return MakeArgError(ps, 2, []env.Type{env.StringType}, "Http-handler//strip-prefix")
			}

		},
	},
}
