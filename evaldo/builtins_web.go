//go:build b_echo
// +build b_echo

package evaldo

import "C"

import (
	"fmt"
	"strconv"

	"github.com/refaktor/rye/env"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo"
)

//	"github.com/labstack/echo-contrib/session"

var OutBuffer = "" // how does this work with multiple threads / ... in server use ... probably we would need some per environment variable, not global / global?

func PopOutBuffer() string {
	r := OutBuffer
	OutBuffer = ""
	return r
}

var Builtins_web = map[string]*env.Builtin{

	"out-buffer": {
		Argsn: 0,
		Doc:   "TODODOC",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return env.String{OutBuffer}
		},
	},

	"echo": {
		Argsn: 1,
		Doc:   "TODODOC",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch str := arg0.(type) {
			case env.String:
				OutBuffer += str.Value
				return str
			case env.Integer:
				OutBuffer += strconv.FormatInt(ps.Res.(env.Integer).Value, 10)
				return str
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType, env.StringType}, "echo")
			}

		},
	},

	"tag": {
		Argsn: 2,
		Doc:   "TODODOC",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch wrp := arg1.(type) {
			case env.String:
				switch str := arg0.(type) {
				case env.String:
					return env.String{"<" + wrp.Value + ">" + str.Value + "</" + wrp.Value + ">"}
				default:
					return MakeArgError(ps, 1, []env.Type{env.StringType}, "wrap")
				}
			default:
				return MakeArgError(ps, 2, []env.Type{env.StringType}, "wrap")
			}

		},
	},

	// BASIC FUNCTIONS WITH NUMBERS

	"form?": {
		Argsn: 2,
		Doc:   "TODODOC",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch ctx := arg0.(type) {
			case env.Native:
				switch key := arg1.(type) {
				case env.String:
					return env.String{ctx.Value.(echo.Context).FormValue(key.Value)}
				default:
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "form?")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "form?")
			}
		},
	},

	"query?": {
		Argsn: 2,
		Doc:   "TODODOC",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			fmt.Println("YOYOYOYOYOYO ------------- - - -  --")
			//return env.String{"QUERY - VAL"}
			switch ctx := arg0.(type) {
			case env.Native:
				switch key := arg1.(type) {
				case env.String:
					//return env.NewError("XOSADOSADOA SDAS DO" + key.Value)
					return env.String{ctx.Value.(echo.Context).QueryParam(key.Value)}
				default:
					return MakeArgError(ps, 1, []env.Type{env.NativeType}, "query?")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "query?")
			}
		},
	},

	"Rye-echo-session//set": { // after we make kinds ... session native will be tagged with session, and set will be multimetod on session
		Argsn: 3,
		Doc:   "TODODOC",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
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
						return MakeArgError(ps, 3, []env.Type{env.StringType}, "Rye-echo-session//set")
					}
					//return env.NewError("XOSADOSADOA SDAS DO" + key.Value)
					return env.String{ctx.Value.(echo.Context).QueryParam(key.Value)}
				default:
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "Rye-echo-session//set")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Rye-echo-session//set")
			}
		},
	},

	"Rye-echo-session//get": { // after we make kinds ... session native will be tagged with session, and set will be multimetod on sessio
		Argsn: 2,
		Doc:   "TODODOC",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
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
						default:
							return MakeBuiltinError(ps, "No matching type found.", "Rye-echo-session//get")
						}
						//return env.NewError("XOSADOSADOA SDAS DO" + key.Value)
					} else {
						return env.String{"NO VALUE"}
					}
				default:
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "Rye-echo-session//get")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Rye-echo-session//get")
			}
		},
	},

	/*
		"queryvals": { // returns the block with query vals
			Argsn: 1,
			Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
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
