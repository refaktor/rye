//go:build b_raylib
// +build b_raylib

package evaldo

import (
	"rye/env"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var Builtins_raylib = map[string]*env.Builtin{

	"raylib-init": {
		Argsn: 3,
		Doc:   "TODODOC",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch width := arg0.(type) {
			case env.Integer:
				switch height := arg1.(type) {
				case env.Integer:
					switch title := arg2.(type) {
					case env.String:
						rl.InitWindow(int32(width.Value), int32(height.Value), title.Value)
						return *env.NewNative(ps.Idx, 1, "raylib")
					default:
						return MakeArgError(ps, 3, []env.Type{env.StringType}, "raylib-init")
					}
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "raylib-init")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType}, "raylib-init")
			}
		},
	},
	"raylib-set-target-fps": {
		Argsn: 1,
		Doc:   "TODODOC",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch fps := arg0.(type) {
			case env.Integer:
				rl.SetTargetFPS(int32(fps.Value))
				return arg0
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType}, "raylib-set-target-fps")
			}
		},
	},
	"raylib-window-should-close": {
		Argsn: 0,
		Doc:   "TODODOC",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			if rl.WindowShouldClose() {
				return env.Integer{1}
			} else {
				return env.Integer{0}
			}
		},
	},
	"raylib-main-loop": {
		Argsn: 1,
		Doc:   "TODODOC",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch bloc := arg0.(type) {
			case env.Block:
				ser := ps.Ser
				ps.Ser = bloc.Series
				for !rl.WindowShouldClose() {
					ps = EvalBlock(ps)
					ps.Ser.Reset()
				}
				ps.Ser = ser
				return ps.Res
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType}, "raylib-main-loop")
			}
		},
	},

	"raylib-begin-drawing": {
		Argsn: 0,
		Doc:   "TODODOC",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			rl.BeginDrawing()
			return arg0
		},
	},
	"raylib-end-drawing": {
		Argsn: 0,
		Doc:   "TODODOC",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			rl.EndDrawing()
			return arg0
		},
	},
	"raylib-close-window": {
		Argsn: 0,
		Doc:   "TODODOC",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			rl.CloseWindow()
			return arg0
		},
	},
	"raylib-draw-circle": {
		Argsn: 4,
		Doc:   "TODODOC",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch cx := arg0.(type) {
			case env.Integer:
				switch cy := arg1.(type) {
				case env.Integer:
					switch r := arg2.(type) {
					case env.Integer:
						switch col := arg3.(type) {
						case env.Native:
							//rl.DrawCircle(int32(cx.Value), int32(cy.Value), float32(r.Value), col.Value.(rl.Color))
							rl.DrawCircle(int32(cx.Value), int32(cy.Value), float32(r.Value), col.Value.(rl.Color))
							return arg0
						default:
							return MakeArgError(ps, 4, []env.Type{env.NativeType}, "raylib-draw-circle")
						}
					default:
						return MakeArgError(ps, 3, []env.Type{env.IntegerType}, "raylib-draw-circle")
					}
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "raylib-draw-circle")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType}, "raylib-draw-circle")
			}
		},
	},

	"raylib-clear-background": {
		Argsn: 1,
		Doc:   "TODODOC",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch col := arg0.(type) {
			case env.Native:
				rl.ClearBackground(col.Value.(rl.Color))
				return arg0
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "raylib-clear-background")
			}
		},
	},

	"raylib-gold": {
		Argsn: 0,
		Doc:   "TODODOC",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewNative(ps.Idx, rl.Gold, "rl-color")
		},
	},
	"raylib-black": {
		Argsn: 0,
		Doc:   "TODODOC",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewNative(ps.Idx, rl.Black, "rl-color")
		},
	},
}
