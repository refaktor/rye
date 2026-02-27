package evaldo

import (
	"github.com/refaktor/rye/env"
	"github.com/refaktor/rye/util"
	// JM 20230825	"github.com/refaktor/rye/term"
)

var builtins_functions = map[string]*env.Builtin{

	//
	// ##### Functions ##### "functions that create functions"
	//
	// Tests:
	// equal { var 'x 10 x:: 20 x } 20
	// Args:
	// * word: Tagword representing the variable name
	// * value: Initial value for the variable
	// Returns:
	// * The initial value
	"var": {
		Argsn: 2,
		Doc:   "Declares a word as a variable with the given value, allowing it to be modified. Returns the word for use with on-change.",
		Pure:  false,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch word := arg0.(type) {
			case env.Tagword:
				// Convert tagword to index
				idx := word.Index
				// Check if word already exists in context
				if _, exists := ps.Ctx.GetCurrent(idx); exists {
					ps.FailureFlag = true
					return env.NewError("Cannot redefine existing word '" + ps.Idx.GetWord(idx) + "' with var")
				}
				// Set the value
				ps.Ctx.SetNew(idx, arg1, ps.Idx)
				// Mark as variable
				ps.Ctx.MarkAsVariable(idx)
				// Return the word instead of the value for use with on-change
				return *env.NewWord(idx)
			case env.Word:
				// Use word index directly
				idx := word.Index
				// Check if word already exists in context
				if _, exists := ps.Ctx.GetCurrent(idx); exists {
					ps.FailureFlag = true
					return env.NewError("Cannot redefine existing word '" + ps.Idx.GetWord(idx) + "' with var")
				}
				ps.Ctx.SetNew(idx, arg1, ps.Idx)
				ps.Ctx.MarkAsVariable(idx)
				// Return the word instead of the value for use with on-change
				return word
			default:
				return MakeArgError(ps, 1, []env.Type{env.TagwordType, env.WordType}, "var")
			}
		},
	},
	// Tests:
	// equal { does { 123 } |type? } 'function
	// equal { x: does { 123 } x } 123
	// equal { x: does { 1 + 2 } x } 3
	// Args:
	// * body: Block containing the function body code
	// Returns:
	// * function object with no parameters
	"does": { // **
		Argsn: 1,
		Doc:   "Creates a function with no arguments that executes the given block when called.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch body := arg0.(type) {
			case env.Block:
				//spec := []env.Object{*env.NewWord(aaaidx)}
				//body := []env.Object{*env.NewWord(printidx), *env.NewWord(aaaidx), *env.NewWord(recuridx), *env.NewWord(greateridx), *env.NewInteger(99), *env.NewWord(aaaidx), *env.NewWord(incidx), *env.NewWord(aaaidx)}
				return *env.NewFunction(*env.NewBlock(*env.NewTSeries(make([]env.Object, 0))), body, false)
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType}, "does")
			}
		},
	},

	// Tests:
	// equal { fn1 { .pass { } } |type? } 'function
	// equal { x: fn1 { } , x 123 } 123
	// equal { x: fn1 { .pass { } } , x 123 } 123
	// equal { x: fn1 { + 1 } , x 123 } 124
	// Args:
	// * body: Block containing the function body code
	// Returns:
	// * function object that accepts one anonymous argument
	"fn1": { // **
		Argsn: 1,
		Doc:   "Creates a function that accepts one anonymous argument and executes the given block with that argument.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch body := arg0.(type) {
			case env.Block:
				spec := []env.Object{*env.NewWord(1)}
				//body := []env.Object{*env.NewWord(printidx), *env.NewWord(aaaidx), *env.NewWord(recuridx), *env.NewWord(greateridx), *env.NewInteger(99), *env.NewWord(aaaidx), *env.NewWord(incidx), *env.NewWord(aaaidx)}
				return *env.NewFunction(*env.NewBlock(*env.NewTSeries(spec)), body, false)
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType}, "fn1")
			}
		},
	},

	// Tests:
	// equal { fn { } { } |type? } 'function
	// equal { x: fn { } { 234 } , x } 234
	// equal { x: fn { x } { x } , x 123 } 123
	// equal { x: fn { x } { + 123 } , x 123 } 246
	// Args:
	// * spec: Block containing parameter specifications
	// * body: Block containing the function body code
	// Returns:
	// * function object with the specified parameters
	"fn": {
		Argsn: 2,
		Doc:   "Creates a function with named parameters specified in the first block and code in the second block.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch args := arg0.(type) {
			case env.Block:
				ok, doc := util.ProcessFunctionSpec(args)
				if !ok {
					return MakeBuiltinError(ps, doc, "fn")
				}
				switch body := arg1.(type) {
				case env.Block:
					//spec := []env.Object{*env.NewWord(aaaidx)}
					//body := []env.Object{*env.NewWord(printidx), *env.NewWord(aaaidx), *env.NewWord(recuridx), *env.NewWord(greateridx), *env.NewInteger(99), *env.NewWord(aaaidx), *env.NewWord(incidx), *env.NewWord(aaaidx)}
					// fmt.Println(doc)
					return *env.NewFunctionDoc(args, body, false, doc)
				default:
					return MakeArgError(ps, 2, []env.Type{env.BlockType}, "fn")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType}, "fn")
			}
		},
	},

	// Tests:
	// equal { f: fn { x y } { x + y } , ?f .fn\spec? } { x y }
	// equal { does { 1 } |fn\spec? } { }
	// Args:
	// * function: Function to get the argument spec from
	// Returns:
	// * Block containing the function's parameter specification
	"fn\\spec?": {
		Argsn: 1,
		Doc:   "Returns the argument specification block of a function.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch fn := arg0.(type) {
			case env.Function:
				return fn.Spec
			default:
				return MakeArgError(ps, 1, []env.Type{env.FunctionType}, "fn\\spec?")
			}
		},
	},

	// Tests:
	// equal { f: fn { x y } { x + y } , ?f .fn\body? } { x + y }
	// equal { does { 123 } |fn\body? } { 123 }
	// Args:
	// * function: Function to get the body from
	// Returns:
	// * Block containing the function's body code
	"fn\\body?": {
		Argsn: 1,
		Doc:   "Returns the body block of a function.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch fn := arg0.(type) {
			case env.Function:
				return fn.Body
			default:
				return MakeArgError(ps, 1, []env.Type{env.FunctionType}, "fn\\body?")
			}
		},
	},

	// Tests:
	// equal { pfn { } { } |type? } 'function
	// equal { x: pfn { x } { + 123 } , x 123 } 246
	// ; TODO -- it seems pure namespace not also has print and append! error { x: pfn { } { ?append! } , x 123 }
	// ; TODO -- it seems pure namespace not also has print and append! error { x: pfn { x } { .print } , x 123 }
	// Args:
	// * spec: Block containing parameter specifications
	// * body: Block containing the function body code
	// Returns:
	// * pure function object with the specified parameters
	"pfn": {
		Argsn: 2,
		Doc:   "Creates a pure function (no side effects allowed) with named parameters and code body.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch args := arg0.(type) {
			case env.Block:
				ok, doc := util.ProcessFunctionSpec(args)
				if !ok {
					return MakeBuiltinError(ps, doc, "fn")
				}
				switch body := arg1.(type) {
				case env.Block:
					return *env.NewFunctionDoc(args, body, true, doc)
				default:
					return MakeArgError(ps, 2, []env.Type{env.BlockType}, "pfn")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType}, "pfn")
			}
		},
	},

	/*
		"fnc": { // TODO -- fnc will maybe become fn\par context is set as parrent, fn\in will be executed directly in context
			// a function with context	 bb: 10 add10 [ a ] context [ b: bb ] [ add a b ]
			// 							add10 [ a ] this [ add a b ]
			// later maybe			   add10 [ a ] [ b: b ] [ add a b ]
			//  						   add10 [ a ] [ 'b ] [ add a b ]
			Argsn: 3,
			Doc:   "Creates a function with specific context.",
			Pure:  true,
			Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
				switch args := arg0.(type) {
				case env.Block:
					ok, doc := util.ProcessFunctionSpec(args)
					if !ok {
						return MakeBuiltinError(ps, doc, "fn")
					}
					switch ctx := arg1.(type) {
					case *env.RyeCtx:
						switch body := arg2.(type) {
						case env.Block:
							return *env.NewFunctionC(args, body, &ctx, false, false, doc)
						default:
							ps.ErrorFlag = true
							return MakeArgError(ps, 3, []env.Type{env.BlockType}, "fnc")
						}
					default:
						ps.ErrorFlag = true
						return MakeArgError(ps, 2, []env.Type{env.ContextType}, "fnc")
					}
				default:
					ps.ErrorFlag = true
					return MakeArgError(ps, 1, []env.Type{env.BlockType}, "fnc")
				}
			},
		}, */

	// Tests:
	// equal { fn\cc { x } { x + y } |type? } 'function
	// equal { y: 5 , f: fn\cc { x } { x + y } , f 3 } 8
	// Args:
	// * spec: Block containing parameter specifications
	// * body: Block containing the function body code
	// Returns:
	// * function object with the current context captured
	"fn\\cc": {
		// a function with context	 bb: 10 add10 [ a ] context [ b: bb ] [ add a b ]
		// 							add10 [ a ] this [ add a b ]
		// later maybe			   add10 [ a ] [ b: b ] [ add a b ]
		//  						   add10 [ a ] [ 'b ] [ add a b ]
		Argsn: 2,
		Doc:   "Creates a function that captures the current context, allowing access to variables from the enclosing scope.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch args := arg0.(type) {
			case env.Block:
				ok, doc := util.ProcessFunctionSpec(args)
				if !ok {
					return MakeBuiltinError(ps, doc, "fn")
				}
				switch body := arg1.(type) {
				case env.Block:
					return *env.NewFunctionC(args, body, ps.Ctx, false, false, doc)
				default:
					ps.ErrorFlag = true
					return MakeArgError(ps, 2, []env.Type{env.BlockType}, "fn\\cc")
				}
			default:
				ps.ErrorFlag = true
				return MakeArgError(ps, 1, []env.Type{env.BlockType}, "fn\\cc")
			}
		},
	},

	// Tests:
	// equal { ctx: context { y: 5 } , f: fn\in { x } ctx { x + y } , f 3 } 8
	// Args:
	// * spec: Block containing parameter specifications
	// * context: Context object to use as parent context
	// * body: Block containing the function body code
	// Returns:
	// * function object with the specified parent context
	"fn\\in": {
		// a function with context	 bb: 10 add10 [ a ] context [ b: bb ] [ add a b ]
		// 							add10 [ a ] this [ add a b ]
		// later maybe			   add10 [ a ] [ b: b ] [ add a b ]
		//  						   add10 [ a ] [ 'b ] [ add a b ]
		Argsn: 3,
		Doc:   "Creates a function with a specified parent context, allowing access to variables from that context.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch args := arg0.(type) {
			case env.Block:
				ok, doc := util.ProcessFunctionSpec(args)
				if !ok {
					return MakeBuiltinError(ps, doc, "fn")
				}
				switch ctx := arg1.(type) {
				case *env.RyeCtx:
					switch body := arg2.(type) {
					case env.Block:
						return *env.NewFunctionC(args, body, ctx, false, false, doc)
					default:
						ps.ErrorFlag = true
						return MakeArgError(ps, 3, []env.Type{env.BlockType}, "fn\\in")
					}
				default:
					ps.ErrorFlag = true
					return MakeArgError(ps, 2, []env.Type{env.ContextType}, "fn\\in")
				}
			default:
				ps.ErrorFlag = true
				return MakeArgError(ps, 1, []env.Type{env.BlockType}, "fn\\in")
			}
		},
	},

	// Tests:
	// equal { ctx: context { y: 5 } , f: fn\inside { x } ctx { x + y } , f 3 } 8
	// Args:
	// * spec: Block containing parameter specifications
	// * context: Context object to execute the function in
	// * body: Block containing the function body code
	// Returns:
	// * function object that executes directly in the specified context
	"fn\\inside": {
		// a function with context	 bb: 10 add10 [ a ] context [ b: bb ] [ add a b ]
		// 							add10 [ a ] this [ add a b ]
		// later maybe			   add10 [ a ] [ b: b ] [ add a b ]
		//  						   add10 [ a ] [ 'b ] [ add a b ]
		Argsn: 3,
		Doc:   "Creates a function that executes directly in the specified context rather than creating a new execution context.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch args := arg0.(type) {
			case env.Block:
				ok, doc := util.ProcessFunctionSpec(args)
				if !ok {
					return MakeBuiltinError(ps, doc, "fn\\inside")
				}
				switch ctx := arg1.(type) {
				case *env.RyeCtx:
					switch body := arg2.(type) {
					case env.Block:
						return *env.NewFunctionC(args, body, ctx, false, true, doc)
					default:
						ps.ErrorFlag = true
						return MakeArgError(ps, 3, []env.Type{env.BlockType}, "fn\\inside")
					}
				default:
					ps.ErrorFlag = true
					return MakeArgError(ps, 2, []env.Type{env.ContextType}, "fn\\inside")
				}
			default:
				ps.ErrorFlag = true
				return MakeArgError(ps, 1, []env.Type{env.BlockType}, "fn\\inside")
			}
		},
	},

	// Tests:
	// equal { y: 5 , f: closure { x } { x + y } , f 3 } 8
	// equal { mk-cntr: does { var 'c 0 , closure { } { inc! 'c } } cnt: mk-cntr , cnt + cnt + cnt } 6
	// Args:
	// * spec: Block containing parameter specifications
	// * body: Block containing the function body code
	// Returns:
	// * function object that captures the current context at creation time
	"closure": {
		// a function with context	 bb: 10 add10 [ a ] context [ b: bb ] [ add a b ]
		// 							add10 [ a ] this [ add a b ]
		// later maybe			   add10 [ a ] [ b: b ] [ add a b ]
		//  						   add10 [ a ] [ 'b ] [ add a b ]
		Argsn: 2,
		Doc:   "Creates a closure that captures the current context at creation time, preserving access to variables in that scope.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			// Reference the current context directly - no copying
			// This allows closures created in the same execution context to share state
			// while different function calls naturally get different contexts
			ctx := ps.Ctx
			ctx.IsClosure = true // Mark this context as a closure context to prevent pooling

			switch args := arg0.(type) {
			case env.Block:
				ok, doc := util.ProcessFunctionSpec(args)
				if !ok {
					return MakeBuiltinError(ps, doc, "closure")
				}
				switch body := arg1.(type) {
				case env.Block:
					return *env.NewFunctionC(args, body, ctx, false, false, doc)
				default:
					ps.ErrorFlag = true
					return MakeArgError(ps, 2, []env.Type{env.BlockType}, "closure")
				}
			default:
				ps.ErrorFlag = true
				return MakeArgError(ps, 1, []env.Type{env.BlockType}, "closure")
			}
		},
	},

	// Tests:
	// equal { prepend-star: partial ?concat [ "* " _ ] , prepend-star "hello" } "* hello"
	// equal { add-5: partial ?_+ [ _ 5 ] , add-5 10 } 15
	// equal { fn-add: fn { x y } { x + y } , add-5: partial ?fn-add [ _ 5 ] , add-5 10 } 15
	// Args:
	// * func: Function or builtin to partially apply
	// * args: Block of arguments, with _ (void) for arguments to be filled later
	// Returns:
	// * CurriedCaller object that can be called with the remaining arguments
	"partial": {
		Argsn: 2,
		Doc:   "Creates a partially applied function with specified arguments, using _ (void) for arguments to be filled later.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			// Check if first argument is a function or builtin
			var callerType int
			var originalArgsn int

			switch fn := arg0.(type) {
			case env.Builtin:
				callerType = 0
				originalArgsn = fn.Argsn
			case env.Function:
				callerType = 1
				originalArgsn = fn.Argsn
			default:
				return MakeArgError(ps, 1, []env.Type{env.BuiltinType, env.FunctionType}, "partial")
			}

			// Check if second argument is a block
			var args []env.Object
			switch block := arg1.(type) {
			case env.Block:
				args = block.Series.GetAll()
				if len(args) > 5 {
					ps.FailureFlag = true
					return env.NewError("partial currently supports up to 5 arguments")
				}
			default:
				return MakeArgError(ps, 2, []env.Type{env.BlockType}, "partial")
			}

			// Extract arguments from the block and count nils (placeholders)
			var cur0, cur1, cur2, cur3, cur4 env.Object
			argsn := 0

			// Set arguments based on the block contents, only count nils up to originalArgsn
			if len(args) > 0 && originalArgsn > 0 {
				if args[0].Type() == env.VoidType {
					cur0 = nil
					argsn++
				} else {
					cur0 = args[0]
				}
			}

			if len(args) > 1 && originalArgsn > 1 {
				if args[1].Type() == env.VoidType {
					cur1 = nil
					argsn++
				} else {
					cur1 = args[1]
				}
			}

			if len(args) > 2 && originalArgsn > 2 {
				if args[2].Type() == env.VoidType {
					cur2 = nil
					argsn++
				} else {
					cur2 = args[2]
				}
			}

			if len(args) > 3 && originalArgsn > 3 {
				if args[3].Type() == env.VoidType {
					cur3 = nil
					argsn++
				} else {
					cur3 = args[3]
				}
			}

			if len(args) > 4 && originalArgsn > 4 {
				if args[4].Type() == env.VoidType {
					cur4 = nil
					argsn++
				} else {
					cur4 = args[4]
				}
			}

			// Create the CurriedCaller based on the function type
			if callerType == 0 {
				// Builtin
				return *env.NewCurriedCallerFromBuiltin(arg0.(env.Builtin), cur0, cur1, cur2, cur3, cur4, argsn)
			} else {
				// Function
				return *env.NewCurriedCallerFromFunction(arg0.(env.Function), cur0, cur1, cur2, cur3, cur4, argsn)
			}
		},
	},

	// Tests:
	// equal { apply ?_+ { 12 23 } } 35
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

					// Use DetermineContext to properly set up the context
					// This handles fn.InCtx (fn\in) and fn.Ctx (fn\par) correctly
					ctx := DetermineContext(fn, ps, nil)
					if ctx == nil {
						ps.Ser = ser
						return ps.Res
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

							// TODO --- check if it exists in context .. should return error and be set? in/into dillema
							// Bind argument to parameter
							ctx.Mod(paramWord.Index, args.Series.Get(i))
						}
					}

					// Save current context
					oldCtx := ps.Ctx

					// Set new context for function execution
					ps.Ctx = ctx

					// Execute function body
					ps.Ser = fn.Body.Series
					EvalBlock(ps)
					MaybeDisplayFailureOrError(ps, ps.Idx, "apply")
					if ps.ErrorFlag {
						ps.Ctx = oldCtx
						ps.Ser = ser
						return ps.Res
					}

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

	// Tests:
	// equal { x: var 'counter 0 , on-change x { print "changed" } , counter:: 5 } 5
	// Args:
	// * word: Word representing the variable to observe
	// * observer: Block containing code to execute when variable changes
	// Returns:
	// * The observer block
	"on-change": {
		Argsn: 2,
		Doc:   "Registers an observer block to be executed when a variable changes. Use with 'var' result.",
		Pure:  false,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch word := arg0.(type) {
			case env.Word:
				switch observerBlock := arg1.(type) {
				case env.Block:
					// Check if the word is actually a variable in the current context
					if _, exists := ps.Ctx.GetCurrent(word.Index); !exists {
						ps.FailureFlag = true
						return env.NewError("Word '" + ps.Idx.GetWord(word.Index) + "' not found in current context")
					}

					if !ps.Ctx.IsVariable(word.Index) {
						ps.FailureFlag = true
						return env.NewError("Word '" + ps.Idx.GetWord(word.Index) + "' is not a variable. Use 'var' to declare it as variable first.")
					}

					// Register the observer in the current context
					ps.Ctx.AddObserver(word.Index, observerBlock)
					return observerBlock
				default:
					return MakeArgError(ps, 2, []env.Type{env.BlockType}, "on-change")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.WordType}, "on-change")
			}
		},
	},
}
