package evaldo

import (
	"github.com/refaktor/rye/env"
)

var builtins_apply = map[string]*env.Builtin{

	// Tests:
	// equal { apply + { 1 2 3 } } 6
	// equal { apply fn { x y } { x + y } { 5 10 } } 15
	// equal { f: fn { x y } { x * y } , apply ?f { 7 6 } } 42
	// Args:
	// * function: Function or builtin to apply
	// * args: Block of arguments to pass to the function
	// Returns:
	// * The result of applying the function to the arguments
	"apply": {
		Argsn: 2,
		Doc:   "Applies a function or builtin to a block of arguments.",
		Pure:  false,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch fn := arg0.(type) {
			case env.Function:
				switch args := arg1.(type) {
				case env.Block:
					// Check if we have enough arguments
					if args.Series.Len() < fn.Argsn {
						ps.FailureFlag = true
						return env.NewError("Not enough arguments for function. Expected " +
							string(rune(fn.Argsn+'0')) + ", got " +
							string(rune(args.Series.Len()+'0')))
					}

					// Save current series position
					ser := ps.Ser

					// Create a new context for function execution
					var ctx *env.RyeCtx
					if fn.Ctx != nil {
						// Use function's captured context if available
						if fn.InCtx {
							ctx = fn.Ctx
						} else {
							ctx = env.NewEnv(fn.Ctx)
						}
					} else {
						// Otherwise create a new context with current context as parent
						ctx = env.NewEnv(ps.Ctx)
					}

					// Bind arguments to parameters
					for i := 0; i < fn.Argsn; i++ {
						if i < args.Series.Len() {
							// Get parameter name from function spec
							paramWord, ok := fn.Spec.Series.Get(i).(env.Word)
							if !ok {
								ps.FailureFlag = true
								return env.NewError("Invalid parameter specification in function")
							}

							// Bind argument to parameter
							ctx.Set(paramWord.Index, args.Series.Get(i))
						}
					}

					// Save current context
					oldCtx := ps.Ctx

					// Set new context for function execution
					ps.Ctx = ctx

					// Execute function body
					ps.Ser = fn.Body.Series
					EvalBlock(ps)

					// Restore original context and series
					ps.Ctx = oldCtx
					ps.Ser = ser

					return ps.Res

				default:
					return MakeArgError(ps, 2, []env.Type{env.BlockType}, "apply")
				}

			case env.Builtin:
				switch args := arg1.(type) {
				case env.Block:
					// Check if we have enough arguments
					if args.Series.Len() < fn.Argsn {
						ps.FailureFlag = true
						return env.NewError("Not enough arguments for builtin. Expected " +
							string(rune(fn.Argsn+'0')) + ", got " +
							string(rune(args.Series.Len()+'0')))
					}

					// Prepare arguments for the builtin
					var arg0, arg1, arg2, arg3, arg4 env.Object

					// Set default values
					arg0 = env.Void{}
					arg1 = env.Void{}
					arg2 = env.Void{}
					arg3 = env.Void{}
					arg4 = env.Void{}

					// Fill in arguments from the block
					if args.Series.Len() > 0 {
						arg0 = args.Series.Get(0)
					}
					if args.Series.Len() > 1 {
						arg1 = args.Series.Get(1)
					}
					if args.Series.Len() > 2 {
						arg2 = args.Series.Get(2)
					}
					if args.Series.Len() > 3 {
						arg3 = args.Series.Get(3)
					}
					if args.Series.Len() > 4 {
						arg4 = args.Series.Get(4)
					}

					// Call the builtin with the arguments
					return fn.Fn(ps, arg0, arg1, arg2, arg3, arg4)

				default:
					return MakeArgError(ps, 2, []env.Type{env.BlockType}, "apply")
				}

			case env.VarBuiltin:
				switch args := arg1.(type) {
				case env.Block:
					// Check if we have enough arguments
					if args.Series.Len() < fn.Argsn {
						ps.FailureFlag = true
						return env.NewError("Not enough arguments for variadic builtin. Expected at least " +
							string(rune(fn.Argsn+'0')) + ", got " +
							string(rune(args.Series.Len()+'0')))
					}

					// Convert block to slice of arguments
					argsSlice := make([]env.Object, args.Series.Len())
					for i := 0; i < args.Series.Len(); i++ {
						argsSlice[i] = args.Series.Get(i)
					}

					// Call the variadic builtin with the arguments
					return fn.Fn(ps, argsSlice...)

				default:
					return MakeArgError(ps, 2, []env.Type{env.BlockType}, "apply")
				}

			default:
				return MakeArgError(ps, 1, []env.Type{env.FunctionType, env.BuiltinType, env.VarBuiltinType}, "apply")
			}
		},
	},
}
