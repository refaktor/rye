package evaldo

import (
	"fmt"

	"github.com/refaktor/rye/env"
	// JM 20230825	"github.com/refaktor/rye/term"
)

var builtins_contexts = map[string]*env.Builtin{

	//
	// ##### Contexts ##### "Context related functions"
	//
	// Tests:
	// equal { c: raw-context { x: 123 } c/x } 123
	// equal { y: 123 try { c: raw-context { x: y } } |type? } 'error ; word not found y
	// equal { try { c: raw-context { x: inc 10 } } |type? } 'error ; word not found inc
	// Args:
	// * block: Block of expressions to evaluate in a new isolated context
	// Returns:
	// * context object with the values defined in the block
	"raw-context": { // **
		Argsn: 1,
		Doc:   "Creates a completely isolated context with no parent, where only built-in functions are available.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch bloc := arg0.(type) {
			case env.Block:
				ser := ps.Ser
				ctx := ps.Ctx
				ps.Ser = bloc.Series
				ps.Ctx = env.NewEnv(nil) // make new context with no parent
				EvalBlock(ps)
				MaybeDisplayFailureOrError(ps, ps.Idx, "raw-context")
				if ps.ReturnFlag || ps.ErrorFlag {
					ps.Ctx = ctx
					ps.Ser = ser
					return ps.Res
				}
				rctx := ps.Ctx
				ps.Ctx = ctx
				ps.Ser = ser
				return *rctx // return the resulting context
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType}, "raw-context")
			}
		},
	},

	// Tests:
	// equal { c: isolate { x: 123 } c/x } 123
	// equal { y: 123 c: isolate { x: y } c/x } 123
	// equal { c: isolate { x: inc 10 } c/x } 11
	// ; equal { y: 99 c: isolate { x: does { y } } try { c/x } |type? } 'error
	// ; equal { y: 99 c: isolate { t: ?try x: does { t { y } } } c/x |type? } 'error
	// Args:
	// * block: Block of expressions to evaluate in a temporary context
	// Returns:
	// * context object with the values defined in the block, but isolated from parent contexts
	"isolate": {
		Argsn: 1,
		Doc:   "Creates a context that can access the parent context during creation, but becomes isolated afterward.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch bloc := arg0.(type) {
			case env.Block:
				ser := ps.Ser
				ctx := ps.Ctx
				ps.Ser = bloc.Series
				ps.Ctx = env.NewEnv(ps.Ctx) // make new context with no parent
				EvalBlock(ps)
				MaybeDisplayFailureOrError(ps, ps.Idx, "isolate")
				if ps.ReturnFlag || ps.ErrorFlag {
					ps.Ctx = ctx
					ps.Ser = ser
					return ps.Res
				}
				rctx := ps.Ctx
				rctx.Parent = nil
				rctx.Kind = *env.NewWord(-1)
				ps.Ctx = ctx
				ps.Ser = ser
				return *rctx // return the resulting context
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType}, "isolate")
			}
		},
	},

	// Tests:
	// equal { c: context { x: 123 } c/x } 123
	// equal { y: 123 c: context { x: y } c/x } 123
	// equal { c: context { x: inc 10 } c/x } 11
	// equal { y: 123 c: context { x: does { y } } c/x } 123
	// Args:
	// * block: Block of expressions to evaluate in a new context
	// Returns:
	// * context object with the values defined in the block and access to parent context
	"context": {
		Argsn: 1,
		Doc:   "Creates a new context that maintains access to its parent context.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch bloc := arg0.(type) {
			case env.Block:
				ser := ps.Ser
				ctx := ps.Ctx
				ps.Ser = bloc.Series
				ps.Ctx = env.NewEnv(ps.Ctx) // make new context with no parent
				EvalBlock(ps)
				MaybeDisplayFailureOrError(ps, ps.Idx, "context")
				if ps.ReturnFlag || ps.ErrorFlag {
					ps.Ctx = ctx
					ps.Ser = ser
					return ps.Res
				}
				rctx := ps.Ctx
				ps.Ctx = ctx
				ps.Ser = ser
				return *rctx // return the resulting context
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType}, "context")
			}
		},
	},

	// Tests:
	// equal { c: context\pure { x: 123 } c/x } 123
	// ; error { y: 123 c: context\pure { x: y } } ; y not accessible in pure context
	// equal { c: context\pure { x: add 10 5 } c/x } 15
	// Args:
	// * block: Block of expressions to evaluate in a new pure context
	// Returns:
	// * context object with the values defined in the block, using Pure Context as parent
	"context\\pure": {
		Argsn: 1,
		Doc:   "Creates a new context using Pure Context (PCtx) as parent, preventing access to regular context changes.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch bloc := arg0.(type) {
			case env.Block:
				ser := ps.Ser
				ctx := ps.Ctx
				ps.Ser = bloc.Series
				ps.Ctx = env.NewEnv(ps.PCtx) // make new context with PCtx as parent instead of regular Ctx
				EvalBlock(ps)
				MaybeDisplayFailureOrError(ps, ps.Idx, "context\\pure")
				if ps.ReturnFlag || ps.ErrorFlag {
					ps.Ctx = ctx
					ps.Ser = ser
					return ps.Res
				}
				rctx := ps.Ctx
				ps.Ctx = ctx
				ps.Ser = ser
				return *rctx // return the resulting context
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType}, "context\\pure")
			}
		},
	},

	// Tests:
	// equal { private { x: 123 } } 123
	// equal { y: 123 private { x: y } } 123
	// equal { private { x: inc 10 } } 11
	// equal { y: 123 private { does { y } } :f f } 123
	// Args:
	// * block: Block of expressions to evaluate in a private context
	// Returns:
	// * the last value from evaluating the block (not the context itself)
	"private": { // **
		Argsn: 1,
		Doc:   "Creates a temporary private context for evaluating expressions, returning the last value instead of the context.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch bloc := arg0.(type) {
			case env.Block:
				ser := ps.Ser
				ctx := ps.Ctx
				ps.Ser = bloc.Series
				ps.Ctx = env.NewEnv(ps.Ctx) // make new context with no parent
				EvalBlock(ps)
				MaybeDisplayFailureOrError(ps, ps.Idx, "private")
				if ps.ReturnFlag || ps.ErrorFlag {
					ps.Ctx = ctx
					ps.Ser = ser
					return ps.Res
				}
				// rctx := ps.Ctx
				ps.Ctx = ctx
				ps.Ser = ser
				return ps.Res // return the resulting context
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType}, "private")
			}
		},
	},

	// Tests:
	// equal { private\ "what are we doing here" { x: 234 1000 + x } } 1234
	// Args:
	// * doc: String containing documentation for the context
	// * block: Block of expressions to evaluate in a private context
	// Returns:
	// * the last value from evaluating the block (not the context itself)
	"private\\": {
		Argsn: 2,
		Doc:   "Creates a documented private context for evaluating expressions, returning the last value instead of the context.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch doc := arg0.(type) {
			case env.String:
				switch bloc := arg1.(type) {
				case env.Block:
					ser := ps.Ser
					ctx := ps.Ctx
					ps.Ser = bloc.Series
					ps.Ctx = env.NewEnv2(ps.Ctx, doc.Value) // make new context with no parent
					EvalBlock(ps)
					MaybeDisplayFailureOrError(ps, ps.Idx, "private\\")
					if ps.ReturnFlag || ps.ErrorFlag {
						ps.Ctx = ctx
						ps.Ser = ser
						return ps.Res
					}
					// rctx := ps.Ctx
					ps.Ctx = ctx
					ps.Ser = ser
					return ps.Res // return the resulting context
				default:
					return MakeArgError(ps, 2, []env.Type{env.BlockType}, "private\\")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "private\\")
			}
		},
	},

	// Tests:
	// equal { ct: context { p: 123 } cn: extends ct { r: p + 234 } cn/r } 357
	// ; error { ct: context { p: 123 } cn: extends ct { r: p + 234 } cn/r }
	// Args:
	// * parent: Context object to extend
	// * block: Block of expressions to evaluate in the new context
	// Returns:
	// * new context object that inherits from the parent context
	"extends": { // ** add one with exclamation mark, which it as it is now extends/changes the source context too .. in place
		Argsn: 2,
		Doc:   "Creates a new context that inherits from a specified parent context.",
		Pure:  false,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch ctx0 := arg0.(type) {
			case env.RyeCtx:
				switch bloc := arg1.(type) {
				case env.Block:
					ser := ps.Ser
					ctx := ps.Ctx
					ps.Ser = bloc.Series
					// ps.Ctx = ctx0.Copy() // make new context with no parent
					ps.Ctx = env.NewEnv(&ctx0) // make new context with no parent
					EvalBlock(ps)
					MaybeDisplayFailureOrError(ps, ps.Idx, "extends")
					if ps.ReturnFlag || ps.ErrorFlag {
						ps.Ctx = ctx
						ps.Ser = ser
						return ps.Res
					}
					rctx := ps.Ctx
					ps.Ctx = ctx
					ps.Ser = ser
					return *rctx // return the resulting context
				default:
					return MakeArgError(ps, 2, []env.Type{env.BlockType}, "extends")
				}
			default:
				ps.ErrorFlag = true
				return MakeArgError(ps, 1, []env.Type{env.ContextType}, "extends")
			}
		},
	},

	// Tests:
	// equal { c: context { y: 123 } cc: bind! context { z: does { y + 234 } } c , cc/z } 357
	// Args:
	// * child: Context object to be bound
	// * parent: Context object to bind to as parent
	// Returns:
	// * the modified child context with its parent set to the specified parent context
	"bind!": { // **
		Argsn: 2,
		Doc:   "Binds a context to a parent context, allowing it to access the parent's values.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch swCtx1 := arg0.(type) {
			case env.RyeCtx:
				switch swCtx2 := arg1.(type) {
				case env.RyeCtx:
					swCtx1.Parent = &swCtx2
					return swCtx1
				default:
					return MakeArgError(ps, 2, []env.Type{env.ContextType}, "bind!")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.ContextType}, "bind!")
			}
		},
	},

	// Tests:
	// equal { c: context { y: 123 } cc: bind! context { z: does { y + 234 } } c , unbind cc cc/z } 357
	// ; error { c: context { y: 123 } cc: bind! context { z: does { y + 234 } } c , dd: unbind cc dd/z }
	// Args:
	// * ctx: Context object to unbind from its parent
	// Returns:
	// * the modified context with no parent
	"unbind": { // **
		Argsn: 1,
		Doc:   "Removes the parent relationship from a context, making it a standalone context.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch swCtx1 := arg0.(type) {
			case env.RyeCtx:
				swCtx1.Parent = nil
				return swCtx1
			default:
				return MakeArgError(ps, 1, []env.Type{env.ContextType}, "unbind!")
			}
		},
	},

	// TODOC
	// Tests:
	// equal { c: context { var 'x 9999 , incr: fn\inside { } current { x:: inc x } } c/incr c/x } 10000
	"current": { // **
		Argsn: 0,
		Doc:   "Returns current context.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *ps.Ctx
		},
	},

	// Tests:
	// equal { var 'y 99 c: context { incr: fn\inside { } parent? { y:: inc y } } c/incr y } 100
	"parent?": { // **
		Argsn: 0,
		Doc:   "Returns parent context of the current context.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *ps.Ctx.Parent
		},
	},

	// Tests:
	// equal { ct: context { p: 123 } parent\of ct |= current } true
	"parent\\of": {
		Argsn: 1,
		Doc:   "Returns parent context of the current context.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch c := arg0.(type) {
			case env.RyeCtx:
				return *c.Parent
			case *env.RyeCtx:
				return *c.Parent
			default:
				return MakeArgError(ps, 1, []env.Type{env.ContextType}, "parent?")
			}
		},
	},

	"lc": {
		Argsn: 0,
		Pure:  true,
		Doc:   "Lists words in current context",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			fmt.Println(ps.Ctx.Preview(*ps.Idx, ""))
			return ps.Ctx
		},
	},

	"lc\\data": {
		Argsn: 0,
		Doc:   "Lists words in current context",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return ps.Ctx.GetWords(*ps.Idx)
		},
	},

	"lc\\data\\": {
		Argsn: 1,
		Doc:   "Lists words in current context",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch c := arg0.(type) {
			case env.RyeCtx:
				return c.GetWords(*ps.Idx)
			case *env.RyeCtx:
				return c.GetWords(*ps.Idx)
			default:
				return MakeArgError(ps, 1, []env.Type{env.ContextType}, "parent?")
			}
		},
	},

	"lcp": {
		Argsn: 0,
		Doc:   "Lists words in current context",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			if ps.Ctx.Parent != nil {
				fmt.Println(ps.Ctx.Parent.Preview(*ps.Idx, ""))
			} else {
				fmt.Println("No parent")
			}
			return ps.Ctx
		},
	},

	"lc\\": {
		Argsn: 1,
		Doc:   "Lists words in current context with string filter, by type (word: 'function, 'builtin, 'context), or regex filter (native of kind 'regexp)",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.String:
				fmt.Println(ps.Ctx.Preview(*ps.Idx, s1.Value))
				return ps.Ctx
			case env.RyeCtx:
				fmt.Println(s1.Preview(*ps.Idx, ""))
				return ps.Ctx
			case env.Word:
				// Handle type-based filtering
				typeName := ps.Idx.GetWord(s1.Index)
				if typeName == "function" || typeName == "builtin" || typeName == "context" {
					fmt.Println(ps.Ctx.PreviewByType(*ps.Idx, typeName))
					return ps.Ctx
				} else {
					return MakeBuiltinError(ps, fmt.Sprintf("Invalid type filter '%s'. Use 'function, 'builtin, or 'context", typeName), "lc\\")
				}
			case env.Native:
				// Handle regex filtering for Native of kind regexp
				kindName := ps.Idx.GetWord(s1.Kind.Index)
				if kindName == "regexp" {
					// Cast the Native value to regex and use it for filtering
					if regexVal, ok := s1.Value.(interface{ MatchString(string) bool }); ok {
						fmt.Println(ps.Ctx.PreviewByRegex(*ps.Idx, regexVal))
						return ps.Ctx
					} else {
						return MakeBuiltinError(ps, "Native regexp does not support MatchString method", "lc\\")
					}
				} else {
					return MakeBuiltinError(ps, fmt.Sprintf("Native filter only supports 'regexp kind, got '%s'", kindName), "lc\\")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType, env.WordType, env.NativeType}, "lc\\")
			}
		},
	},

	"lcp\\": {
		Argsn: 1,
		Doc:   "Lists words in current context with string filter",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.String:
				if ps.Ctx.Parent != nil {
					fmt.Println(ps.Ctx.Parent.Preview(*ps.Idx, s1.Value))
				} else {
					fmt.Println("No parent")
				}
				return ps.Ctx
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "lsp\\")
			}
		},
	},

	"cc": {
		Argsn: 1,
		Doc:   "Change to context (pushes current context to stack for ccb)",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.RyeCtx:
				// Push current context to stack before changing
				ps.PushContext(ps.Ctx)
				ps.Ctx = &s1
				return ps.Ctx
			case *env.RyeCtx:
				// Push current context to stack before changing
				ps.PushContext(ps.Ctx)
				ps.Ctx = s1
				return ps.Ctx
			default:
				return MakeArgError(ps, 1, []env.Type{env.ContextType}, "cc")
			}
		},
	},

	"ccp": {
		Argsn: 0,
		Doc:   "Change to parent context (pushes current context to stack for ccb)",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			if ps.Ctx.Parent == nil {
				return MakeBuiltinError(ps, "No parent context available.", "ccp")
			}
			// Push current context to stack before changing to parent
			ps.PushContext(ps.Ctx)
			ps.Ctx = ps.Ctx.Parent
			return ps.Ctx
		},
	},

	"ccb": {
		Argsn: 0,
		Doc:   "Change context back (pops from context stack)",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			prevCtx, ok := ps.PopContext()
			if !ok {
				return MakeBuiltinError(ps, "No previous context in stack to return to.", "ccb")
			}
			ps.Ctx = prevCtx
			return *ps.Ctx
		},
	},

	"mkcc": {
		Argsn: 1,
		Doc:   "Make context with current as parent and change to it (pushes current context to stack for ccb).",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch word := arg0.(type) {
			case env.Word:
				newctx := env.NewEnv(ps.Ctx)
				ret := ps.Ctx.Set(word.Index, newctx)
				s, ok := ret.(env.Error)
				if ok {
					return s
				}
				// Push current context to stack before changing to new context
				ps.PushContext(ps.Ctx)
				ps.Ctx = newctx // make new context with current par
				return ps.Ctx
			default:
				return MakeArgError(ps, 1, []env.Type{env.WordType}, "mkcc")
			}
		},
	},

	"cc-stack-size": {
		Argsn: 0,
		Doc:   "Returns the current size of the context navigation stack.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewInteger(int64(ps.ContextStackSize()))
		},
	},

	"cc-clear-stack": {
		Argsn: 0,
		Doc:   "Clears the context navigation stack (removes all stored previous contexts).",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			oldSize := ps.ContextStackSize()
			ps.ContextStack = make([]*env.RyeCtx, 0)
			return *env.NewInteger(int64(oldSize))
		},
	},

	// Tests:
	// equal { c: context { x: 123 y: 456 } cc: clone c cc/x } 123
	// equal { c: context { x: 123 y: 456 } cc: clone c cc/y } 456
	// equal { c: context { x:: 123 } cc: clone c do\inside cc { x:: 999 } c/x  } 123 ; original unchanged
	// equal { c: context { x:: 123 } cc: clone c do\inside cc { x:: 999 } cc/x } 999 ; clone modified
	// Args:
	// * ctx: Context object to clone
	// Returns:
	// * a new context object that is a copy of the original context
	"clone": {
		Argsn: 1,
		Doc:   "Creates a copy of a context with the same state and parent relationship.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch ctx := arg0.(type) {
			case env.RyeCtx:
				clonedCtx := ctx.Copy()
				if ryeCtx, ok := clonedCtx.(*env.RyeCtx); ok {
					return *ryeCtx
				}
				return MakeArgError(ps, 1, []env.Type{env.ContextType}, "clone")
			case *env.RyeCtx:
				clonedCtx := ctx.Copy()
				if ryeCtx, ok := clonedCtx.(*env.RyeCtx); ok {
					return *ryeCtx
				}
				return MakeArgError(ps, 1, []env.Type{env.ContextType}, "clone")
			default:
				return MakeArgError(ps, 1, []env.Type{env.ContextType}, "clone")
			}
		},
	},

	// Tests:
	// equal { c: context { x: 123 } cc: clone\ c { y: x + 100 } cc/y } 223
	// ; error { c: context { x:: 123 } cc: clone\ c { y:: x + 100 } c/y } 'error ; y not in original context
	// equal { c: context { x:: 123 } cc: clone\ c { x:: 999 } c/x } 123 ; original unchanged
	// equal { c: context { x:: 123 } cc: clone\ c { x:: 999 } cc/x } 999 ; clone modified
	// Args:
	// * ctx: Context object to clone
	// * block: Block of expressions to evaluate in the cloned context
	// Returns:
	// * the cloned context with the block evaluated inside it
	"clone\\": {
		Argsn: 2,
		Doc:   "Creates a copy of a context and evaluates a block of code inside the clone.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch ctx := arg0.(type) {
			case env.RyeCtx:
				switch bloc := arg1.(type) {
				case env.Block:
					ser := ps.Ser
					origCtx := ps.Ctx
					clonedCtx := ctx.Copy()
					if ryeCtx, ok := clonedCtx.(*env.RyeCtx); ok {
						ps.Ser = bloc.Series
						ps.Ctx = ryeCtx
						EvalBlock(ps)
						MaybeDisplayFailureOrError(ps, ps.Idx, "clone\\")
						if ps.ReturnFlag || ps.ErrorFlag {
							ps.Ctx = origCtx
							ps.Ser = ser
							return ps.Res
						}
						rctx := ps.Ctx
						ps.Ctx = origCtx
						ps.Ser = ser
						return *rctx // return the resulting cloned context
					}
					return MakeArgError(ps, 1, []env.Type{env.ContextType}, "clone\\")
				default:
					return MakeArgError(ps, 2, []env.Type{env.BlockType}, "clone\\")
				}
			case *env.RyeCtx:
				switch bloc := arg1.(type) {
				case env.Block:
					ser := ps.Ser
					origCtx := ps.Ctx
					clonedCtx := ctx.Copy()
					if ryeCtx, ok := clonedCtx.(*env.RyeCtx); ok {
						ps.Ser = bloc.Series
						ps.Ctx = ryeCtx
						EvalBlock(ps)
						MaybeDisplayFailureOrError(ps, ps.Idx, "clone\\")
						if ps.ReturnFlag || ps.ErrorFlag {
							ps.Ctx = origCtx
							ps.Ser = ser
							return ps.Res
						}
						rctx := ps.Ctx
						ps.Ctx = origCtx
						ps.Ser = ser
						return *rctx // return the resulting cloned context
					}
					return MakeArgError(ps, 1, []env.Type{env.ContextType}, "clone\\")
				default:
					return MakeArgError(ps, 2, []env.Type{env.BlockType}, "clone\\")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.ContextType}, "clone\\")
			}
		},
	},

	// Tests:
	// equal { c: context { x: [ 1 2 3 ] } cc: clone\deep c cc/x } [ 1 2 3 ]
	// equal { c: context { x: [ 1 2 3 ] } cc: clone\deep c do\in cc { change\nth! ref x 1 999 } c/x -> 0 } 1 ; original unchanged
	// equal { c: context { x: [ 1 2 3 ] } cc: clone\deep c do\in cc { change\nth! ref x 1 999 } cc/x -> 0 } 999 ; deep clone modified
	// Args:
	// * ctx: Context object to deep clone
	// Returns:
	// * a new context object that is a deep copy of the original context (including nested objects)
	"clone\\deep": {
		Argsn: 1,
		Doc:   "Creates a deep copy of a context with the same state and parent relationship, recursively copying all nested objects.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch ctx := arg0.(type) {
			case env.RyeCtx:
				clonedCtx := ctx.DeepCopy()
				if ryeCtx, ok := clonedCtx.(*env.RyeCtx); ok {
					return *ryeCtx
				}
				return MakeArgError(ps, 1, []env.Type{env.ContextType}, "clone\\deep")
			case *env.RyeCtx:
				clonedCtx := ctx.DeepCopy()
				if ryeCtx, ok := clonedCtx.(*env.RyeCtx); ok {
					return *ryeCtx
				}
				return MakeArgError(ps, 1, []env.Type{env.ContextType}, "clone\\deep")
			default:
				return MakeArgError(ps, 1, []env.Type{env.ContextType}, "clone\\deep")
			}
		},
	},
}
