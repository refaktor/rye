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
						CallFunction_CollectArgs(middleware, &psTemp, *env.NewNative(ps.Idx, c, "Gin-context"), false, nil)
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
						CallFunction_CollectArgs(middleware, &psTemp, *env.NewNative(ps.Idx, c, "Gin-context"), false, nil)
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

	// Built-in middleware constructors (simplified)
	"middleware\\CORS": {
		Argsn: 0,
		Doc:   "Create basic CORS middleware function.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			// Return a simple Rye function that implements CORS
			return *env.NewString("{ ctx | ctx//Header ctx \"Access-Control-Allow-Origin\" \"*\" ctx//Header ctx \"Access-Control-Allow-Methods\" \"GET, POST, PUT, DELETE, OPTIONS\" ctx//Header ctx \"Access-Control-Allow-Headers\" \"Content-Type, Authorization\" ctx//Next ctx }")
		},
	},

	"middleware\\Logger": {
		Argsn: 0,
		Doc:   "Create basic logging middleware function.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			// Return a simple Rye function that implements logging
			return *env.NewString("{ ctx | print [ \"Request from:\" ctx//ClientIP ctx ] ctx//Next ctx }")
		},
	},

	"middleware\\Auth": {
		Argsn: 1,
		Doc:   "Create authentication middleware with secret key.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch secret := arg0.(type) {
			case env.String:
				// Return a simple Rye function that implements auth checking
				authCode := "{ ctx | auth-header: ctx//GetHeader ctx \"Authorization\" either [ equals? auth-header \"Bearer " + secret.Value + "\" ] [ ctx//Next ctx ] [ ctx//AbortWithStatus ctx status\\unauthorized ] }"
				return *env.NewString(authCode)
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
							CallFunction_CollectArgs(handler, &psTemp, *env.NewNative(ps.Idx, c, "Gin-context"), false, nil)
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
							CallFunction_CollectArgs(handler, &psTemp, *env.NewNative(ps.Idx, c, "Gin-context"), false, nil)
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
							CallFunction_CollectArgs(handler, &psTemp, *env.NewNative(ps.Idx, c, "Gin-context"), false, nil)
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
							CallFunction_CollectArgs(handler, &psTemp, *env.NewNative(ps.Idx, c, "Gin-context"), false, nil)
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
							CallFunction_CollectArgs(handler, &psTemp, *env.NewNative(ps.Idx, c, "Gin-context"), false, nil)
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
							CallFunction_CollectArgs(handler, &psTemp, *env.NewNative(ps.Idx, c, "Gin-context"), false, nil)
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
					case env.Dict:
						ctx.Value.(*gin.Context).JSON(int(code.Value), data.Data)
						return arg0
					case env.List:
						ctx.Value.(*gin.Context).JSON(int(code.Value), data.Data)
						return arg0
					case env.Block:
						// Convert block to slice
						items := make([]interface{}, data.Series.Len())
						for i := 0; i < data.Series.Len(); i++ {
							items[i] = data.Series.Get(i)
						}
						ctx.Value.(*gin.Context).JSON(int(code.Value), items)
						return arg0
					default:
						ps.FailureFlag = true
						return MakeArgError(ps, 3, []env.Type{env.DictType, env.ListType, env.BlockType}, "Gin-context//JSON")
					}
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "Gin-context//JSON")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Gin-context//JSON")
			}
		},
	},

	"Gin-context//HTML": {
		Argsn: 4,
		Doc:   "Render HTML template with data.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch ctx := arg0.(type) {
			case env.Native:
				switch code := arg1.(type) {
				case env.Integer:
					switch tmplName := arg2.(type) {
					case env.String:
						switch data := arg3.(type) {
						case env.Dict:
							// Convert Rye dict to gin.H
							ginH := make(gin.H)
							for k, v := range data.Data {
								ginH[k] = v
							}
							ctx.Value.(*gin.Context).HTML(int(code.Value), tmplName.Value, ginH)
							return arg0
						default:
							ps.FailureFlag = true
							return MakeArgError(ps, 4, []env.Type{env.DictType}, "Gin-context//HTML")
						}
					default:
						ps.FailureFlag = true
						return MakeArgError(ps, 3, []env.Type{env.StringType}, "Gin-context//HTML")
					}
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "Gin-context//HTML")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Gin-context//HTML")
			}
		},
	},

	"Gin-context//ShouldBindJSON": {
		Argsn: 1,
		Doc:   "Bind request JSON to Rye dict.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch ctx := arg0.(type) {
			case env.Native:
				var jsonData map[string]interface{}
				err := ctx.Value.(*gin.Context).ShouldBindJSON(&jsonData)
				if err != nil {
					ps.FailureFlag = true
					return MakeBuiltinError(ps, err.Error(), "Gin-context//ShouldBindJSON")
				}
				// Convert to Rye dict
				ryeDict := env.NewDict(jsonData)
				return *ryeDict
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Gin-context//ShouldBindJSON")
			}
		},
	},

	"Gin-context//Header": {
		Argsn: 3,
		Doc:   "Set response header in Gin context.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch ctx := arg0.(type) {
			case env.Native:
				switch name := arg1.(type) {
				case env.String:
					switch value := arg2.(type) {
					case env.String:
						ctx.Value.(*gin.Context).Header(name.Value, value.Value)
						return arg0
					default:
						ps.FailureFlag = true
						return MakeArgError(ps, 3, []env.Type{env.StringType}, "Gin-context//Header")
					}
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "Gin-context//Header")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Gin-context//Header")
			}
		},
	},

	"Gin-context//Status": {
		Argsn: 2,
		Doc:   "Set response status code.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch ctx := arg0.(type) {
			case env.Native:
				switch code := arg1.(type) {
				case env.Integer:
					ctx.Value.(*gin.Context).Status(int(code.Value))
					return arg0
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "Gin-context//Status")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Gin-context//Status")
			}
		},
	},

	"Gin-context//FullPath": {
		Argsn: 1,
		Doc:   "Get the full path of the matched route.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch ctx := arg0.(type) {
			case env.Native:
				fullPath := ctx.Value.(*gin.Context).FullPath()
				return *env.NewString(fullPath)
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Gin-context//FullPath")
			}
		},
	},

	"Gin-context//ClientIP": {
		Argsn: 1,
		Doc:   "Get client IP address from Gin context.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch ctx := arg0.(type) {
			case env.Native:
				clientIP := ctx.Value.(*gin.Context).ClientIP()
				return *env.NewString(clientIP)
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Gin-context//ClientIP")
			}
		},
	},

	"Gin-context//GetHeader": {
		Argsn: 2,
		Doc:   "Get request header value from Gin context.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch ctx := arg0.(type) {
			case env.Native:
				switch name := arg1.(type) {
				case env.String:
					value := ctx.Value.(*gin.Context).GetHeader(name.Value)
					return *env.NewString(value)
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "Gin-context//GetHeader")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Gin-context//GetHeader")
			}
		},
	},

	// HTTP Status constants
	"status\\OK": {
		Argsn: 0,
		Doc:   "HTTP 200 OK status code constant.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewInteger(http.StatusOK)
		},
	},

	"status\\created": {
		Argsn: 0,
		Doc:   "HTTP 201 Created status code constant.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewInteger(http.StatusCreated)
		},
	},

	"status\\bad-request": {
		Argsn: 0,
		Doc:   "HTTP 400 Bad Request status code constant.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewInteger(http.StatusBadRequest)
		},
	},

	"status\\not-found": {
		Argsn: 0,
		Doc:   "HTTP 404 Not Found status code constant.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewInteger(http.StatusNotFound)
		},
	},

	"status\\unauthorized": {
		Argsn: 0,
		Doc:   "HTTP 401 Unauthorized status code constant.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewInteger(http.StatusUnauthorized)
		},
	},

	"status\\internal-server-error": {
		Argsn: 0,
		Doc:   "HTTP 500 Internal Server Error status code constant.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewInteger(http.StatusInternalServerError)
		},
	},

	"status\\conflict": {
		Argsn: 0,
		Doc:   "HTTP 409 Conflict status code constant.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewInteger(http.StatusConflict)
		},
	},

	// Cookie support
	"Gin-context//SetCookie": {
		Argsn: 8,
		Doc:   "Set HTTP cookie with name, value, maxAge, path, domain, secure, httpOnly.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch ctx := arg0.(type) {
			case env.Native:
				switch name := arg1.(type) {
				case env.String:
					switch value := arg2.(type) {
					case env.String:
						switch maxAge := arg3.(type) {
						case env.Integer:
							switch path := arg4.(type) {
							case env.String:
								// For simplicity, use basic SetCookie with name, value, maxAge, path
								// Additional parameters (domain, secure, httpOnly) would be in a more complex version
								ctx.Value.(*gin.Context).SetCookie(name.Value, value.Value, int(maxAge.Value), path.Value, "", false, false)
								return arg0
							default:
								ps.FailureFlag = true
								return MakeArgError(ps, 5, []env.Type{env.StringType}, "Gin-context//SetCookie")
							}
						default:
							ps.FailureFlag = true
							return MakeArgError(ps, 4, []env.Type{env.IntegerType}, "Gin-context//SetCookie")
						}
					default:
						ps.FailureFlag = true
						return MakeArgError(ps, 3, []env.Type{env.StringType}, "Gin-context//SetCookie")
					}
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "Gin-context//SetCookie")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Gin-context//SetCookie")
			}
		},
	},

	"Gin-context//Cookie": {
		Argsn: 2,
		Doc:   "Get cookie value by name from request.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch ctx := arg0.(type) {
			case env.Native:
				switch name := arg1.(type) {
				case env.String:
					value, err := ctx.Value.(*gin.Context).Cookie(name.Value)
					if err != nil {
						// Return empty string if cookie not found
						return *env.NewString("")
					}
					return *env.NewString(value)
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "Gin-context//Cookie")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Gin-context//Cookie")
			}
		},
	},

	// Session support (simple in-memory sessions)
	"Gin-context//SetSession": {
		Argsn: 3,
		Doc:   "Set session value by key (stored in context).",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch ctx := arg0.(type) {
			case env.Native:
				switch key := arg1.(type) {
				case env.String:
					// Store in gin context
					ctx.Value.(*gin.Context).Set("session_"+key.Value, arg2)
					return arg0
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "Gin-context//SetSession")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Gin-context//SetSession")
			}
		},
	},

	"Gin-context//GetSession": {
		Argsn: 2,
		Doc:   "Get session value by key from context.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch ctx := arg0.(type) {
			case env.Native:
				switch key := arg1.(type) {
				case env.String:
					value, exists := ctx.Value.(*gin.Context).Get("session_" + key.Value)
					if !exists {
						return *env.NewString("")
					}
					// Convert back to Rye object if possible
					switch v := value.(type) {
					case string:
						return *env.NewString(v)
					case int64:
						return *env.NewInteger(v)
					case env.Object:
						return v
					default:
						return *env.NewString("")
					}
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "Gin-context//GetSession")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Gin-context//GetSession")
			}
		},
	},

	"Gin-context//ClearSession": {
		Argsn: 2,
		Doc:   "Clear/remove session value by key.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch ctx := arg0.(type) {
			case env.Native:
				switch key := arg1.(type) {
				case env.String:
					// Remove from gin context
					ctx.Value.(*gin.Context).Set("session_"+key.Value, nil)
					return arg0
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "Gin-context//ClearSession")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Gin-context//ClearSession")
			}
		},
	},

	// Session helpers
	"session\\login": {
		Argsn: 3,
		Doc:   "Helper function to create user login session.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch ctx := arg0.(type) {
			case env.Native:
				switch userID := arg1.(type) {
				case env.String:
					switch username := arg2.(type) {
					case env.String:
						// Set session data
						ctx.Value.(*gin.Context).Set("session_user_id", userID.Value)
						ctx.Value.(*gin.Context).Set("session_username", username.Value)
						ctx.Value.(*gin.Context).Set("session_logged_in", true)

						// Set session cookie (1 hour expiry)
						ctx.Value.(*gin.Context).SetCookie("session_id", userID.Value, 3600, "/", "", false, true)

						return arg0
					default:
						ps.FailureFlag = true
						return MakeArgError(ps, 3, []env.Type{env.StringType}, "session\\login")
					}
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "session\\login")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "session\\login")
			}
		},
	},

	"session\\logout": {
		Argsn: 1,
		Doc:   "Helper function to destroy user session.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch ctx := arg0.(type) {
			case env.Native:
				// Clear session data
				ctx.Value.(*gin.Context).Set("session_user_id", nil)
				ctx.Value.(*gin.Context).Set("session_username", nil)
				ctx.Value.(*gin.Context).Set("session_logged_in", false)

				// Clear session cookie
				ctx.Value.(*gin.Context).SetCookie("session_id", "", -1, "/", "", false, true)

				return arg0
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "session\\logout")
			}
		},
	},

	"session\\is-logged-in": {
		Argsn: 1,
		Doc:   "Check if user is logged in.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch ctx := arg0.(type) {
			case env.Native:
				loggedIn, exists := ctx.Value.(*gin.Context).Get("session_logged_in")
				if exists && loggedIn == true {
					return *env.NewInteger(1)
				}
				return *env.NewInteger(0)
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "session\\is-logged-in")
			}
		},
	},

	"session\\get-user": {
		Argsn: 1,
		Doc:   "Get current logged in user info as dict.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch ctx := arg0.(type) {
			case env.Native:
				userID, _ := ctx.Value.(*gin.Context).Get("session_user_id")
				username, _ := ctx.Value.(*gin.Context).Get("session_username")
				loggedIn, _ := ctx.Value.(*gin.Context).Get("session_logged_in")

				userData := make(map[string]interface{})
				if userID != nil {
					userData["user_id"] = userID
				}
				if username != nil {
					userData["username"] = username
				}
				userData["logged_in"] = loggedIn == true

				return *env.NewDict(userData)
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "session\\get-user")
			}
		},
	},
}
