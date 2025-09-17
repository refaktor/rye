//go:build !no_gin
// +build !no_gin

package evaldo

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	"github.com/refaktor/rye/env"
)

var Builtins_gin = map[string]*env.Builtin{

	"router": {
		Argsn: 0,
		Doc:   "Create new Gin router with default middleware.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			router := gin.Default()
			return *env.NewNative(ps.Idx, router, "Gin-router")
		},
	},

	"router\\new": {
		Argsn: 0,
		Doc:   "Create new Gin router without default middleware.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			router := gin.New()
			return *env.NewNative(ps.Idx, router, "Gin-router")
		},
	},

	// Middleware support
	"Gin-router//Use": {
		Argsn: 2,
		Doc:   "Add middleware function to Gin router.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch router := arg0.(type) {
			case env.Native:
				switch middleware := arg1.(type) {
				case env.Function:
					router.Value.(*gin.Engine).Use(func(c *gin.Context) {
						ps.FailureFlag = false
						ps.ErrorFlag = false
						ps.ReturnFlag = false
						psTemp := env.ProgramState{}
						err := copier.Copy(&psTemp, &ps)
						if err != nil {
							c.String(http.StatusInternalServerError, "Middleware error")
							return
						}
						CallFunction(middleware, &psTemp, *env.NewNative(ps.Idx, c, "Gin-context"), false, nil)
						if psTemp.FailureFlag || psTemp.ErrorFlag {
							c.Abort()
							return
						}
					})
					return arg0
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 2, []env.Type{env.FunctionType}, "Gin-router//Use")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Gin-router//Use")
			}
		},
	},

	"Gin-group//Use": {
		Argsn: 2,
		Doc:   "Add middleware function to Gin route group.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch group := arg0.(type) {
			case env.Native:
				switch middleware := arg1.(type) {
				case env.Function:
					group.Value.(*gin.RouterGroup).Use(func(c *gin.Context) {
						ps.FailureFlag = false
						ps.ErrorFlag = false
						ps.ReturnFlag = false
						psTemp := env.ProgramState{}
						err := copier.Copy(&psTemp, &ps)
						if err != nil {
							c.String(http.StatusInternalServerError, "Middleware error")
							return
						}
						CallFunction(middleware, &psTemp, *env.NewNative(ps.Idx, c, "Gin-context"), false, nil)
						if psTemp.FailureFlag || psTemp.ErrorFlag {
							c.Abort()
							return
						}
					})
					return arg0
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 2, []env.Type{env.FunctionType}, "Gin-group//Use")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Gin-group//Use")
			}
		},
	},

	"Gin-context//Next": {
		Argsn: 1,
		Doc:   "Continue to next middleware/handler in chain.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch ctx := arg0.(type) {
			case env.Native:
				ctx.Value.(*gin.Context).Next()
				return arg0
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Gin-context//Next")
			}
		},
	},

	"Gin-context//Abort": {
		Argsn: 1,
		Doc:   "Abort the middleware chain and prevent further handlers from executing.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch ctx := arg0.(type) {
			case env.Native:
				ctx.Value.(*gin.Context).Abort()
				return arg0
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Gin-context//Abort")
			}
		},
	},

	"Gin-context//AbortWithStatus": {
		Argsn: 2,
		Doc:   "Abort with HTTP status code.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch ctx := arg0.(type) {
			case env.Native:
				switch code := arg1.(type) {
				case env.Integer:
					ctx.Value.(*gin.Context).AbortWithStatus(int(code.Value))
					return arg0
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "Gin-context//AbortWithStatus")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Gin-context//AbortWithStatus")
			}
		},
	},

	"Gin-context//IsAborted": {
		Argsn: 1,
		Doc:   "Check if context has been aborted.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch ctx := arg0.(type) {
			case env.Native:
				aborted := ctx.Value.(*gin.Context).IsAborted()
				if aborted {
					return *env.NewInteger(1)
				}
				return *env.NewInteger(0)
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Gin-context//IsAborted")
			}
		},
	},

	// Built-in middleware constructors
	"middleware\\CORS": {
		Argsn: 0,
		Doc:   "Create basic CORS middleware function.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			// Return a Rye function that implements CORS
			corsFunc := env.NewFunction(
				*env.NewBlock(*env.NewTSeries([]env.Object{
					*env.NewWord("ctx"),
					*env.NewString("|"),
					*env.NewWord("ctx//Header"),
					*env.NewWord("ctx"),
					*env.NewString("Access-Control-Allow-Origin"),
					*env.NewString("*"),
					*env.NewWord("ctx//Header"),
					*env.NewWord("ctx"),
					*env.NewString("Access-Control-Allow-Methods"),
					*env.NewString("GET, POST, PUT, DELETE, OPTIONS"),
					*env.NewWord("ctx//Header"),
					*env.NewWord("ctx"),
					*env.NewString("Access-Control-Allow-Headers"),
					*env.NewString("Content-Type, Authorization"),
					*env.NewWord("ctx//Next"),
					*env.NewWord("ctx"),
				})),
				*env.NewBlock(*env.NewTSeries([]env.Object{})),
				false)
			return *corsFunc
		},
	},

	"middleware\\Logger": {
		Argsn: 0,
		Doc:   "Create basic logging middleware function.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			// Return a Rye function that implements logging
			logFunc := env.NewFunction(
				*env.NewBlock(*env.NewTSeries([]env.Object{
					*env.NewWord("ctx"),
					*env.NewString("|"),
					*env.NewWord("print"),
					*env.NewBlock(*env.NewTSeries([]env.Object{
						*env.NewString("Request:"),
						*env.NewWord("ctx//GetHeader"),
						*env.NewWord("ctx"),
						*env.NewString("User-Agent"),
					})),
					*env.NewWord("ctx//Next"),
					*env.NewWord("ctx"),
				})),
				*env.NewBlock(*env.NewTSeries([]env.Object{})),
				false)
			return *logFunc
		},
	},

	"middleware\\Auth": {
		Argsn: 1,
		Doc:   "Create authentication middleware with secret key.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch secret := arg0.(type) {
			case env.String:
				// Return a Rye function that implements auth checking
				authFunc := env.NewFunction(
					*env.NewBlock(*env.NewTSeries([]env.Object{
						*env.NewWord("ctx"),
						*env.NewString("|"),
						*env.NewWord("auth-header"),
						*env.NewWord(":"),
						*env.NewWord("ctx//GetHeader"),
						*env.NewWord("ctx"),
						*env.NewString("Authorization"),
						*env.NewWord("either"),
						*env.NewBlock(*env.NewTSeries([]env.Object{
							*env.NewWord("equals?"),
							*env.NewWord("auth-header"),
							*env.NewString("Bearer " + secret.Value),
						})),
						*env.NewBlock(*env.NewTSeries([]env.Object{
							*env.NewWord("ctx//Next"),
							*env.NewWord("ctx"),
						})),
						*env.NewBlock(*env.NewTSeries([]env.Object{
							*env.NewWord("ctx//AbortWithStatus"),
							*env.NewWord("ctx"),
							*env.NewWord("status\\unauthorized"),
						})),
					})),
					*env.NewBlock(*env.NewTSeries([]env.Object{})),
					false)
				return *authFunc
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "middleware\\Auth")
			}
		},
	},

	"Gin-router//GET": {
		Argsn: 3,
		Doc:   "Add GET route handler to Gin router.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch router := arg0.(type) {
			case env.Native:
				switch path := arg1.(type) {
				case env.String:
					switch handler := arg2.(type) {
					case env.Function:
						router.Value.(*gin.Engine).GET(path.Value, func(c *gin.Context) {
							ps.FailureFlag = false
							ps.ErrorFlag = false
							ps.ReturnFlag = false
							psTemp := env.ProgramState{}
							err := copier.Copy(&psTemp, &ps)
							if err != nil {
								c.String(http.StatusInternalServerError, "Internal error")
								return
							}
							CallFunction(handler, &psTemp, *env.NewNative(ps.Idx, c, "Gin-context"), false, nil)
							// TODO: don't display if not in debug mode
							if psTemp.FailureFlag {
								c.String(http.StatusInternalServerError, psTemp.Res.Inspect(*ps.Idx))
								return
							}
							if psTemp.ErrorFlag {
								c.String(http.StatusInternalServerError, psTemp.Res.Inspect(*ps.Idx))
								return
							}
						})
						return arg0
					default:
						ps.FailureFlag = true
						return MakeArgError(ps, 3, []env.Type{env.FunctionType}, "Gin-router//GET")
					}
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "Gin-router//GET")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Gin-router//GET")
			}
		},
	},

	"Gin-router//POST": {
		Argsn: 3,
		Doc:   "Add POST route handler to Gin router.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch router := arg0.(type) {
			case env.Native:
				switch path := arg1.(type) {
				case env.String:
					switch handler := arg2.(type) {
					case env.Function:
						router.Value.(*gin.Engine).POST(path.Value, func(c *gin.Context) {
							ps.FailureFlag = false
							ps.ErrorFlag = false
							ps.ReturnFlag = false
							psTemp := env.ProgramState{}
							err := copier.Copy(&psTemp, &ps)
							if err != nil {
								c.String(http.StatusInternalServerError, "Internal error")
								return
							}
							CallFunction(handler, &psTemp, *env.NewNative(ps.Idx, c, "Gin-context"), false, nil)
						})
						return arg0
					default:
						ps.FailureFlag = true
						return MakeArgError(ps, 3, []env.Type{env.FunctionType}, "Gin-router//POST")
					}
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "Gin-router//POST")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Gin-router//POST")
			}
		},
	},

	"Gin-router//PUT": {
		Argsn: 3,
		Doc:   "Add PUT route handler to Gin router.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch router := arg0.(type) {
			case env.Native:
				switch path := arg1.(type) {
				case env.String:
					switch handler := arg2.(type) {
					case env.Function:
						router.Value.(*gin.Engine).PUT(path.Value, func(c *gin.Context) {
							ps.FailureFlag = false
							ps.ErrorFlag = false
							ps.ReturnFlag = false
							psTemp := env.ProgramState{}
							err := copier.Copy(&psTemp, &ps)
							if err != nil {
								c.String(http.StatusInternalServerError, "Internal error")
								return
							}
							CallFunction(handler, &psTemp, *env.NewNative(ps.Idx, c, "Gin-context"), false, nil)
						})
						return arg0
					default:
						ps.FailureFlag = true
						return MakeArgError(ps, 3, []env.Type{env.FunctionType}, "Gin-router//PUT")
					}
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "Gin-router//PUT")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Gin-router//PUT")
			}
		},
	},

	"Gin-router//DELETE": {
		Argsn: 3,
		Doc:   "Add DELETE route handler to Gin router.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch router := arg0.(type) {
			case env.Native:
				switch path := arg1.(type) {
				case env.String:
					switch handler := arg2.(type) {
					case env.Function:
						router.Value.(*gin.Engine).DELETE(path.Value, func(c *gin.Context) {
							ps.FailureFlag = false
							ps.ErrorFlag = false
							ps.ReturnFlag = false
							psTemp := env.ProgramState{}
							err := copier.Copy(&psTemp, &ps)
							if err != nil {
								c.String(http.StatusInternalServerError, "Internal error")
								return
							}
							CallFunction(handler, &psTemp, *env.NewNative(ps.Idx, c, "Gin-context"), false, nil)
						})
						return arg0
					default:
						ps.FailureFlag = true
						return MakeArgError(ps, 3, []env.Type{env.FunctionType}, "Gin-router//DELETE")
					}
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "Gin-router//DELETE")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Gin-router//DELETE")
			}
		},
	},

	"Gin-router//Run": {
		Argsn: 2,
		Doc:   "Start the Gin server listening on specified address.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch router := arg0.(type) {
			case env.Native:
				switch addr := arg1.(type) {
				case env.String:
					err := router.Value.(*gin.Engine).Run(addr.Value)
					if err != nil {
						ps.FailureFlag = true
						return MakeBuiltinError(ps, err.Error(), "Gin-router//Run")
					}
					return arg0
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "Gin-router//Run")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Gin-router//Run")
			}
		},
	},

	// Template handling
	"Gin-router//LoadHTMLGlob": {
		Argsn: 2,
		Doc:   "Load HTML templates from a glob pattern.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch router := arg0.(type) {
			case env.Native:
				switch pattern := arg1.(type) {
				case env.String:
					router.Value.(*gin.Engine).LoadHTMLGlob(pattern.Value)
					return arg0
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "Gin-router//LoadHTMLGlob")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Gin-router//LoadHTMLGlob")
			}
		},
	},

	// Route groups
	"Gin-router//Group": {
		Argsn: 2,
		Doc:   "Create route group with prefix.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch router := arg0.(type) {
			case env.Native:
				switch prefix := arg1.(type) {
				case env.String:
					group := router.Value.(*gin.Engine).Group(prefix.Value)
					return *env.NewNative(ps.Idx, group, "Gin-group")
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "Gin-router//Group")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Gin-router//Group")
			}
		},
	},

	"Gin-group//GET": {
		Argsn: 3,
		Doc:   "Add GET route handler to Gin route group.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch group := arg0.(type) {
			case env.Native:
				switch path := arg1.(type) {
				case env.String:
					switch handler := arg2.(type) {
					case env.Function:
						group.Value.(*gin.RouterGroup).GET(path.Value, func(c *gin.Context) {
							ps.FailureFlag = false
							ps.ErrorFlag = false
							ps.ReturnFlag = false
							psTemp := env.ProgramState{}
							err := copier.Copy(&psTemp, &ps)
							if err != nil {
								c.String(http.StatusInternalServerError, "Internal error")
								return
							}
							CallFunction(handler, &psTemp, *env.NewNative(ps.Idx, c, "Gin-context"), false, nil)
						})
						return arg0
					default:
						ps.FailureFlag = true
						return MakeArgError(ps, 3, []env.Type{env.FunctionType}, "Gin-group//GET")
					}
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "Gin-group//GET")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Gin-group//GET")
			}
		},
	},

	"Gin-group//POST": {
		Argsn: 3,
		Doc:   "Add POST route handler to Gin route group.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch group := arg0.(type) {
			case env.Native:
				switch path := arg1.(type) {
				case env.String:
					switch handler := arg2.(type) {
					case env.Function:
						group.Value.(*gin.RouterGroup).POST(path.Value, func(c *gin.Context) {
							ps.FailureFlag = false
							ps.ErrorFlag = false
							ps.ReturnFlag = false
							psTemp := env.ProgramState{}
							err := copier.Copy(&psTemp, &ps)
							if err != nil {
								c.String(http.StatusInternalServerError, "Internal error")
								return
							}
							CallFunction(handler, &psTemp, *env.NewNative(ps.Idx, c, "Gin-context"), false, nil)
						})
						return arg0
					default:
						ps.FailureFlag = true
						return MakeArgError(ps, 3, []env.Type{env.FunctionType}, "Gin-group//POST")
					}
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "Gin-group//POST")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Gin-group//POST")
			}
		},
	},

	"Gin-context//Param": {
		Argsn: 2,
		Doc:   "Get URL parameter from Gin context.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch ctx := arg0.(type) {
			case env.Native:
				switch key := arg1.(type) {
				case env.String:
					value := ctx.Value.(*gin.Context).Param(key.Value)
					return *env.NewString(value)
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "Gin-context//Param")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Gin-context//Param")
			}
		},
	},

	"Gin-context//Query": {
		Argsn: 2,
		Doc:   "Get query parameter from Gin context.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch ctx := arg0.(type) {
			case env.Native:
				switch key := arg1.(type) {
				case env.String:
					value := ctx.Value.(*gin.Context).Query(key.Value)
					return *env.NewString(value)
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "Gin-context//Query")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Gin-context//Query")
			}
		},
	},

	"Gin-context//PostForm": {
		Argsn: 2,
		Doc:   "Get form parameter from Gin context.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch ctx := arg0.(type) {
			case env.Native:
				switch key := arg1.(type) {
				case env.String:
					value := ctx.Value.(*gin.Context).PostForm(key.Value)
					return *env.NewString(value)
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "Gin-context//PostForm")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Gin-context//PostForm")
			}
		},
	},

	"Gin-context//String": {
		Argsn: 3,
		Doc:   "Send string response with status code.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch ctx := arg0.(type) {
			case env.Native:
				switch code := arg1.(type) {
				case env.Integer:
					switch message := arg2.(type) {
					case env.String:
						ctx.Value.(*gin.Context).String(int(code.Value), message.Value)
						return arg0
					default:
						ps.FailureFlag = true
						return MakeArgError(ps, 3, []env.Type{env.StringType}, "Gin-context//String")
					}
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "Gin-context//String")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Gin-context//String")
			}
		},
	},

	"Gin-context//JSON": {
		Argsn: 3,
		Doc:   "Send JSON response with status code.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch ctx := arg0.(type) {
			case env.Native:
				switch code := arg1.(type) {
				case env.Integer:
					switch data := arg2.(type) {
					case env.Dict
