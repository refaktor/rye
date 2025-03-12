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

	//
	// ##### Web ##### "Web and HTML generation functions"
	//
	// Tests:
	// equal { echo "Hello" , out-buffer } "Hello"
	// Args:
	// * none
	// Returns:
	// * string containing the current output buffer
	"out-buffer": {
		Argsn: 0,
		Doc:   "Returns the current content of the output buffer.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return env.String{OutBuffer}
		},
	},

	// Tests:
	// equal { echo "Hello" } "Hello"
	// equal { echo 123 } 123
	// Args:
	// * value: String or integer to append to the output buffer
	// Returns:
	// * the input value
	"echo": {
		Argsn: 1,
		Doc:   "Appends a string or integer to the output buffer.",
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

	// Tests:
	// equal { "Hello" |tag "div" } "<div>Hello</div>"
	// Args:
	// * content: String content to wrap in HTML tag
	// * tag: String name of the HTML tag
	// Returns:
	// * string with content wrapped in the specified HTML tag
	"tag": {
		Argsn: 2,
		Doc:   "Wraps content in an HTML tag.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch wrp := arg1.(type) {
			case env.String:
				switch str := arg0.(type) {
				case env.String:
					return env.String{"<" + wrp.Value + ">" + str.Value + "</" + wrp.Value + ">"}
				default:
					return MakeArgError(ps, 1, []env.Type{env.StringType}, "tag")
				}
			default:
				return MakeArgError(ps, 2, []env.Type{env.StringType}, "tag")
			}

		},
	},

	// Tests:
	// equal { ctx: echo-context-mock , ctx |form? "username" } "user1"
	// Args:
	// * context: Echo context object
	// * name: String name of the form field
	// Returns:
	// * string value of the form field
	"form?": {
		Argsn: 2,
		Doc:   "Gets a form field value from an Echo context.",
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

	// Tests:
	// equal { ctx: echo-context-mock , ctx |query? "search" } "keyword"
	// Args:
	// * context: Echo context object
	// * name: String name of the query parameter
	// Returns:
	// * string value of the query parameter
	"query?": {
		Argsn: 2,
		Doc:   "Gets a query parameter value from an Echo context.",
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

	// Tests:
	// equal { session: echo-session-mock , session |Rye-echo-session//set "user" "john" |type? } 'string
	// Args:
	// * session: Echo session object
	// * key: String key for the session value
	// * value: String value to store in the session
	// Returns:
	// * the stored value
	"Rye-echo-session//set": { // after we make kinds ... session native will be tagged with session, and set will be multimetod on session
		Argsn: 3,
		Doc:   "Sets a value in an Echo session.",
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

	// Tests:
	// equal { session: echo-session-mock , session |Rye-echo-session//get "user" } "john"
	// Args:
	// * session: Echo session object
	// * key: String key for the session value
	// Returns:
	// * the stored value or "NO VALUE" if not found
	"Rye-echo-session//get": { // after we make kinds ... session native will be tagged with session, and set will be multimetod on sessio
		Argsn: 2,
		Doc:   "Gets a value from an Echo session.",
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
