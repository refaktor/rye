//go:build b_ebitengine
// +build b_ebitengine

package ebitengine

import (
	"fmt"

	"github.com/refaktor/rye/env"
	"github.com/refaktor/rye/evaldo"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

var onDraw *env.Block
var onUpdate *env.Block

var Ps *env.ProgramState

var LayoutScale int = 1

type Game struct {
	w int
	h int
}

var game Game

// Update proceeds the game state.
// Update is called every tick (1/60 [s] by default).
func (g *Game) Update() error {
	// Write your game's logical update.
	if onUpdate == nil {
		return nil
	}

	ser := Ps.Ser
	Ps.Ser = onUpdate.Series
	r := evaldo.EvalBlockInj(Ps, nil, false)
	Ps.Ser = ser
	if r.Res != nil && r.Res.Type() == env.ErrorType {
		fmt.Println(r.Res.(*env.Error).Message)
	}
	// TODO error handling
	return nil
}

// Draw draws the game screen.
// Draw is called every frame (typically 1/60[s] for 60Hz display).
func (g *Game) Draw(screen *ebiten.Image) {
	// Write your game's rendering.
	if onDraw == nil {
		return
	}

	scr := *env.NewNative(Ps.Idx, screen, "screen")

	// fmt.Println("on--draw")
	// fmt.Println(onDraw)
	ser := Ps.Ser
	Ps.Ser = onDraw.Series
	r := evaldo.EvalBlockInj(Ps, scr, true)
	if r.Res != nil && r.Res.Type() == env.ErrorType {
		fmt.Println(r.Res.(*env.Error).Message)
	}
	Ps.Ser = ser
}

// Layout takes the outside size (e.g., the window size) and returns the (logical) screen size.
// If you don't have to adjust the screen size with the outside size, just return a fixed size.
func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return g.w / LayoutScale, g.h / LayoutScale
}

// ** Just a proof of concept module so far **

// ## TODO
// Global Ps and onDraw are undesired but they seemed necesarry for how (statically?) ebitengine semms to work.
// If I overlooked something and there is any solution without the need for this it should be removed!

// ## IDEA
// Should move to external repo rye-alterego, which contrary to main Rye, would focus on desktop / UI / game / windows?
// This would make main Rye and Contrib cleaner and focused on the linux backend tasks, information (pre)processing, ...
// It would also serve as a test if we can move contrib to external module instead of it being a git submodule which
// complicates many things.

var Builtins_ebitengine = map[string]*env.Builtin{

	"ebitengine-run": {
		Argsn: 2,
		Doc:   "TODODOC",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			Ps = ps
			w, ok := arg0.(env.Integer)
			if !ok {
				return evaldo.MakeArgError(ps, 0, []env.Type{env.IntegerType}, "ebitengine-run")
			}
			h, ok := arg1.(env.Integer)
			if !ok {
				return evaldo.MakeArgError(ps, 1, []env.Type{env.IntegerType}, "ebitengine-run")
			}
			ebiten.SetWindowSize(int(w.Value), int(h.Value))
			ebiten.SetWindowTitle("Your game's title")
			game := &Game{
				w: int(w.Value),
				h: int(h.Value),
			}

			// Call ebiten.RunGame to start your game loop.
			if err := ebiten.RunGame(game); err != nil {
				// log.Fatal(err)
				return nil
			}
			return nil
		},
	},
	"set-layout-scale": {
		Argsn: 1,
		Doc:   "TODODOC",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch scale := arg0.(type) {
			case env.Integer:
				LayoutScale = int(scale.Value)
			}
			return nil
		},
	},
	"on-draw": {
		Argsn: 1,
		Doc:   "TODODOC",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			onDraw_ := arg0.(env.Block)
			onDraw = &onDraw_
			Ps = ps
			return nil
		},
	},
	"on-update": {
		Argsn: 1,
		Doc:   "TODODOC",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			onUpdate_ := arg0.(env.Block)
			onUpdate = &onUpdate_
			Ps = ps
			return nil
		},
	},
	"new-image": {
		Argsn: 1,
		Doc:   "TODODOC",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch fileName := arg0.(type) {
			case env.String:
				img, _, err := ebitenutil.NewImageFromFile(fileName.Value)
				if err != nil {
					return evaldo.MakeError(ps, err.Error())
				}
				return *env.NewNative(ps.Idx, img, "image")
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.StringType}, "new-image")
			}
		},
	},
	"draw-image": {
		Argsn: 2,
		Doc:   "TODODOC",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			arg0.(env.Native).Value.(*ebiten.Image).DrawImage(arg1.(env.Native).Value.(*ebiten.Image), nil)
			return nil
		},
	},
	"write-pixels": {
		Argsn: 2,
		Doc:   "TODODOC",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch img := arg0.(type) {
			case env.Native:
				switch block := arg1.(type) {
				case env.Block:
					pixels := make([]byte, block.Series.Len())
					for i := 0; i < block.Series.Len(); i++ {
						integer, ok := block.Series.Get(i).(env.Integer)
						if !ok {
							return evaldo.MakeBuiltinError(ps, "pixel block must contain only integers", "write-pixels")
						}
						pixels[i] = byte(integer.Value)
					}
					img.Value.(*ebiten.Image).WritePixels(pixels)
					return nil
				case env.List:
					pixels := make([]byte, len(block.Data))
					for i := 0; i < len(block.Data); i++ {
						integer, ok := block.Data[i].(int64)
						if !ok {
							fmt.Printf("pixel: %T\n", block.Data[i])
							return evaldo.MakeBuiltinError(ps, "pixel list must contain only integers", "write-pixels")
						}
						pixels[i] = byte(integer)
					}
					img.Value.(*ebiten.Image).WritePixels(pixels)
					return nil
				default:
					return evaldo.MakeArgError(ps, 1, []env.Type{env.BlockType}, "write-pixels")
				}
			default:
				return evaldo.MakeArgError(ps, 0, []env.Type{env.NativeType}, "write-pixels")
			}
		},
	},
}
