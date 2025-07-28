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
				rctx := ps.Ctx
				ps.Ctx = ctx
				ps.Ser = ser
				if ps.ErrorFlag {
					return ps.Res
				}
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
	// equal { y: 99 c: isolate { x: does { y } } try { c/x } |type? } 'error
	// equal { y: 99 c: isolate { t: ?try x: does { t { y } } } c/x |type? } 'error
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
				rctx := ps.Ctx
				rctx.Parent = nil
				rctx.Kind = *env.NewWord(-1)
				ps.Ctx = ctx
				ps.Ser = ser
				if ps.ErrorFlag {
					return ps.Res
				}
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
				rctx := ps.Ctx
				ps.Ctx = ctx
				ps.Ser = ser
				if ps.ErrorFlag {
					return ps.Res
				}
				return *rctx // return the resulting context
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType}, "context")
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
	// error { ct: context { p: 123 } cn: extends ct { r: p + 234 } cn/r |type? }
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
					rctx := ps.Ctx
					ps.Ctx = ctx
					ps.Ser = ser
					return *rctx // return the resulting context
				default:
					return MakeArgError(ps, 2, []env.Type{env.BlockType}, "extends")
				}
			default:
				ps.ErrorFlag = true
				return MakeArgError(ps, 1, []env.Type{env.CtxType}, "extends")
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
					return MakeArgError(ps, 2, []env.Type{env.CtxType}, "bind!")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.CtxType}, "bind!")
			}
		},
	},

	// Tests:
	// equal { c: context { y: 123 } cc: bind! context { z: does { y + 234 } } c , unbind cc cc/z } 357
	// error { c: context { y: 123 } cc: bind! context { z: does { y + 234 } } c , dd: unbind cc dd/z }
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
				return MakeArgError(ps, 1, []env.Type{env.CtxType}, "unbind!")
			}
		},
	},

	// TODOC
	// Tests:
	// equal { c: context { var 'x 9999 , incr: fn\in { } current { x:: inc x } } c/incr c/x } 10000
	"current": { // **
		Argsn: 0,
		Doc:   "Returns current context.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *ps.Ctx
		},
	},

	// Tests:
	// equal { var 'y 99 c: context { incr: fn\in { } parent { y:: inc y } } c/incr y } 100
	"parent": { // **
		Argsn: 0,
		Doc:   "Returns parent context of the current context.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *ps.Ctx.Parent
		},
	},

	// Tests:
	// equal { ct: context { p: 123 } parent\of ct |= current } 1
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
				return MakeArgError(ps, 1, []env.Type{env.CtxType}, "parent?")
			}
		},
	},

	"lc": {
		Argsn: 0,
		Doc:   "Lists words in current context",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			fmt.Println(ps.Ctx.Preview(*ps.Idx, ""))
			return env.Void{}
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
				return MakeArgError(ps, 1, []env.Type{env.CtxType}, "parent?")
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
			return env.Void{}
		},
	},

	"lc\\": {
		Argsn: 1,
		Doc:   "Lists words in current context with string filter",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.String:
				fmt.Println(ps.Ctx.Preview(*ps.Idx, s1.Value))
				return env.Void{}
			case env.RyeCtx:
				fmt.Println(s1.Preview(*ps.Idx, ""))
				return env.Void{}
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "ls\\")
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
				return env.Void{}
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "lsp\\")
			}
		},
	},

	"cc": {
		Argsn: 1,
		Doc:   "Change to context",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.RyeCtx:
				// s1.Parent = ps.Ctx // TODO ... this is temporary so ccp works, but some other method must be figured out as changing the parent is not OK
				ps.Ctx = &s1
				return s1
			case *env.RyeCtx:
				// s1.Parent = ps.Ctx // TODO ... this is temporary so ccp works, but some other method must be figured out as changing the parent is not OK
				ps.Ctx = s1
				return s1
			default:
				return MakeArgError(ps, 1, []env.Type{env.CtxType}, "cc")
			}
		},
	},

	"ccp": {
		Argsn: 0,
		Doc:   "Change to context",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			cc := ps.Ctx
			ps.Ctx = ps.Ctx.Parent
			return *cc
		},
	},

	"mkcc": {
		Argsn: 1,
		Doc:   "Make context with current as parent and change to it.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch word := arg0.(type) {
			case env.Word:
				newctx := env.NewEnv(ps.Ctx)
				ret := ps.Ctx.Set(word.Index, newctx)
				s, ok := ret.(env.Error)
				if ok {
					return s
				}
				ctx := ps.Ctx
				ps.Ctx = newctx // make new context with current par
				return *ctx
			default:
				return MakeArgError(ps, 1, []env.Type{env.WordType}, "mkcc")
			}
		},
	},

	// Tests:
	// equal { c: context { x: 123 y: 456 } cc: clone c cc/x } 123
	// equal { c: context { x: 123 y: 456 } cc: clone c cc/y } 456
	// equal { c: context { x: 123 } cc: clone c cc .x: 999 c/x } 123 ; original unchanged
	// equal { c: context { x: 123 } cc: clone c cc .x: 999 cc/x } 999 ; clone modified
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
				return *clonedCtx
			case *env.RyeCtx:
				clonedCtx := ctx.Copy()
				return *clonedCtx
			default:
				return MakeArgError(ps, 1, []env.Type{env.CtxType}, "clone")
			}
		},
	},

	// Tests:
	// equal { c: context { x: 123 } cc: clone\ c { y: x + 100 } cc/y } 223
	// equal { c: context { x: 123 } cc: clone\ c { y: x + 100 } c/y } 'error ; y not in original context
	// equal { c: context { x: 123 } cc: clone\ c { x: 999 } c/x } 123 ; original unchanged
	// equal { c: context { x: 123 } cc: clone\ c { x: 999 } cc/x } 999 ; clone modified
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
					ps.Ser = bloc.Series
					ps.Ctx = clonedCtx
					EvalBlock(ps)
					rctx := ps.Ctx
					ps.Ctx = origCtx
					ps.Ser = ser
					if ps.ErrorFlag {
						return ps.Res
					}
					return *rctx // return the resulting cloned context
				default:
					return MakeArgError(ps, 2, []env.Type{env.BlockType}, "clone\\")
				}
			case *env.RyeCtx:
				switch bloc := arg1.(type) {
				case env.Block:
					ser := ps.Ser
					origCtx := ps.Ctx
					clonedCtx := ctx.Copy()
					ps.Ser = bloc.Series
					ps.Ctx = clonedCtx
					EvalBlock(ps)
					rctx := ps.Ctx
					ps.Ctx = origCtx
					ps.Ser = ser
					if ps.ErrorFlag {
						return ps.Res
					}
					return *rctx // return the resulting cloned context
				default:
					return MakeArgError(ps, 2, []env.Type{env.BlockType}, "clone\\")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.CtxType}, "clone\\")
			}
		},
	},
}
