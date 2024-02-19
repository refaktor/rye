//go:build b_fyne

package fyne

// import "C"

import (
	"fmt"

	"github.com/refaktor/rye/env"
	"github.com/refaktor/rye/evaldo"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

var Builtins_fyne = map[string]*env.Builtin{

	"fyne-app": {
		Argsn: 0,
		Doc:   "TODODOC.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			app1 := app.New()
			return *env.NewNative(ps.Idx, app1, "fyne-app")
		},
	},
	"fyne-app//new-window": {
		Argsn: 2,
		Doc:   "TODODOC.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch obj := arg0.(type) {
			case env.Native:
				switch val := arg1.(type) {
				case env.String:
					wind := obj.Value.(fyne.App).NewWindow(val.Value)
					return *env.NewNative(ps.Idx, wind, "fyne-window")
				default:
					return evaldo.MakeArgError(ps, 2, []env.Type{env.StringType}, "gtk-window//set-title")
				}
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "gtk-window//set-title")
			}
		},
	},
	"fyne-label": {
		Argsn: 1,
		Doc:   "Create new gtk window.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch val := arg0.(type) {
			case env.String:
				win := widget.NewLabel(val.Value)
				return *env.NewNative(ps.Idx, win, "fyne-widget")
			default:
				return evaldo.MakeArgError(ps, 2, []env.Type{env.StringType}, "gtk-window//set-title")
			}
		},
	},
	"fyne-widget//set-text": {
		Argsn: 2,
		Doc:   "Create new gtk window.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch val := arg0.(type) {
			case env.Native:
				switch txt := arg1.(type) {
				case env.String:
					val.Value.(*widget.Label).SetText(txt.Value)
					return arg0
				default:
					return evaldo.MakeArgError(ps, 2, []env.Type{env.StringType}, "gtk-window//set-title")
				}
			default:
				return evaldo.MakeArgError(ps, 2, []env.Type{env.StringType}, "gtk-window//set-title")
			}
		},
	},
	"fyne-container(2)": {
		Argsn: 2,
		Doc:   "Create new gtk window.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch val := arg0.(type) {
			case env.Native:
				switch val1 := arg1.(type) {
				case env.Native:
					win := container.New(layout.NewVBoxLayout(), val.Value.(fyne.CanvasObject), val1.Value.(fyne.CanvasObject))
					return *env.NewNative(ps.Idx, win, "fyne-container")
				default:
					return evaldo.MakeArgError(ps, 2, []env.Type{env.StringType}, "gtk-window//set-title")
				}
			default:
				return evaldo.MakeArgError(ps, 2, []env.Type{env.StringType}, "gtk-window//set-title")
			}
		},
	},

	"fyne-button": {
		Argsn: 2,
		Doc:   "Create new gtk window.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch val := arg0.(type) {
			case env.String:
				switch fn := arg1.(type) {
				case env.Function:
					win := widget.NewButton(val.Value, func() {
						evaldo.CallFunction(fn, ps, nil, false, ps.Ctx)
						// return ps.Res
					})
					return *env.NewNative(ps.Idx, win, "fyne-widget")
				case env.Block:
					win := widget.NewButton(val.Value, func() {
						ser := ps.Ser
						ps.Ser = fn.Series
						fmt.Println("BEFORE")
						r := evaldo.EvalBlockInj(ps, nil, false)
						ps.Ser = ser
						fmt.Println("AFTER")
						if r.Res != nil && r.Res.Type() == env.ErrorType {
							fmt.Println(r.Res.(*env.Error).Message)
						}
					})
					return *env.NewNative(ps.Idx, win, "fyne-widget")
				default:
					return evaldo.MakeArgError(ps, 2, []env.Type{env.StringType}, "gtk-window//set-title")
				}
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.StringType}, "gtk-window//set-title")
			}
		},
	},

	"fyne-window//set-content": {
		Argsn: 2,
		Doc:   "Set title of gtk window.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch obj := arg0.(type) {
			case env.Native:
				switch val := arg1.(type) {
				case env.Native:
					obj.Value.(fyne.Window).SetContent(val.Value.(fyne.CanvasObject))
					return obj
				default:
					return evaldo.MakeArgError(ps, 2, []env.Type{env.StringType}, "gtk-window//set-title")
				}
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "gtk-window//set-title")
			}
		},
	},
	"fyne-window//show-and-run": {
		Argsn: 1,
		Doc:   "Show gtk window.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch obj := arg0.(type) {
			case env.Native:
				obj.Value.(fyne.Window).ShowAndRun()
				return obj
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "gtk-window//show")
			}
		},
	},
}
