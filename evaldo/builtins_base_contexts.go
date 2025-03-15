package evaldo

import (
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
}
