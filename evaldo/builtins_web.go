package evaldo

import "C"

import (
	"rye/env"
	"fmt"
	"strconv"

	"github.com/labstack/echo"
)
import (
	"github.com/gorilla/sessions"
	//	"github.com/labstack/echo-contrib/session"
)

var OutBuffer = "" // how does this work with multiple threads / ... in server use ... probably we would need some per environment variable, not global / global?

func PopOutBuffer() string {
	r := OutBuffer
	OutBuffer = ""
	return r
}

var Builtins_web = map[string]*env.Builtin{

	"out-buffer": {
		Argsn: 0,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return env.String{OutBuffer}
		},
	},

	"echo": {
		Argsn: 1,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch str := arg0.(type) {
			case env.String:
				OutBuffer += str.Value
				return str
			case env.Integer:
				OutBuffer += strconv.FormatInt(env1.Res.(env.Integer).Value, 10)
				return str
			default:
				return env.NewError("arg 2 should be string %s")
			}

		},
	},

	"wrap": {
		Argsn: 2,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch wrp := arg1.(type) {
			case env.String:
				switch str := arg0.(type) {
				case env.String:
					return env.String{"<" + wrp.Value + ">" + str.Value + "</" + wrp.Value + ">"}
				default:
					return env.NewError("arg should be string %s")
				}
			default:
				return env.NewError("arg should be string %s")
			}

		},
	},

	// BASIC FUNCTIONS WITH NUMBERS

	"form?": {
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
	},

	"query?": {
		Argsn: 2,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			fmt.Println("YOYOYOYOYOYO ------------- - - -  --")
			//return env.String{"QUERY - VAL"}
			switch ctx := arg0.(type) {
			case env.Native:
				switch key := arg1.(type) {
				case env.String:
					//return env.NewError("XOSADOSADOA SDAS DO" + key.Value)
					return env.String{ctx.Value.(echo.Context).QueryParam(key.Value)}
				default:
					return env.NewError("second arg should be string, got %s")
				}
			default:
				return env.NewError("first arg should be echo.Context, got %s")
			}
		},
	},

	"Rye-echo-session//set": { // after we make kinds ... session native will be tagged with session, and set will be multimetod on session
		Argsn: 3,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			fmt.Println("YOYOYOYOYOYO ------------- - - -  --")
			//return env.String{"QUERY - VAL"}
			switch ctx := arg0.(type) {
			case env.Native:
				switch key := arg1.(type) {
				case env.String:
					switch val := arg2.(type) {
					case env.String:
						//return env.NewError("XOSADOSADOA SDAS DO" + key.Value)
						ctx.Value.(*sessions.Session).Values[key.Value] = val.Value
						return val
					default:
						return env.NewError("second arg should be string, got %s")
					}
					//return env.NewError("XOSADOSADOA SDAS DO" + key.Value)
					return env.String{ctx.Value.(echo.Context).QueryParam(key.Value)}
				default:
					return env.NewError("second arg should be string, got %s")
				}
			default:
				return env.NewError("first arg should be echo.Context, got %s")
			}
		},
	},

	"Rye-echo-session//get": { // after we make kinds ... session native will be tagged with session, and set will be multimetod on sessio
		Argsn: 2,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			fmt.Println("YOYOYOYOYOYO ------------- - - -  --")
			//return env.String{"QUERY - VAL"}
			switch ctx := arg0.(type) {
			case env.Native:
				switch key := arg1.(type) {
				case env.String:
					val := ctx.Value.(*sessions.Session).Values[key.Value]
					if val != nil {
						fmt.Println("***************************************************************")
						fmt.Println(val)
						switch val2 := val.(type) {
						case string:
							return env.String{val2}
						case env.Object:
							return val2
						}
						return env.NewError("bla 123141")
						//return env.NewError("XOSADOSADOA SDAS DO" + key.Value)
					} else {
						return env.String{"NO VALUE"}
					}
				default:
					return env.NewError("second arg should be string, got %s")
				}
			default:
				return env.NewError("first arg should be echo.Context, got %s")
			}
		},
	},

	/*
		"queryvals": { // returns the block with query vals
			Argsn: 1,
			Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
				fmt.Println("YOYOYOYOYOYO ------------- - - -  --")
				//return env.String{"QUERY - VAL"}
				switch ctx := arg0.(type) {
				case env.Native:
					switch key := arg1.(type) {
					case env.String:
						//return env.NewError("XOSADOSADOA SDAS DO" + key.Value)
						return env.String{ctx.Value.(echo.Context).QueryParams()}
					default:
						return env.NewError("second arg should be string, got %s")
					}
				default:
					return env.NewError("first arg should be echo.Context, got %s")
				}
			},
		}, */
}
