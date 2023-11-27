//go:build b_gtk
// +build b_gtk

package evaldo

import "C"

import (
	"rye/env"

	"github.com/gotk3/gotk3/gtk"
)

var Builtins_gtk = map[string]*env.Builtin{

	"gtk-init": {
		Argsn: 0,
		Doc:   "TODODOC.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			gtk.Init(nil)
			return env.Integer{1}
		},
	},
	"gtk-main": {
		Argsn: 0,
		Doc:   "TODODOC.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			gtk.Main()
			return env.Integer{1}
		},
	},
	"gtk-new-window": {
		Argsn: 0,
		Doc:   "Create new gtk window.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			win, _ := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
			return *env.NewNative(ps.Idx, win, "gtk-window")
		},
	},
	"gtk-window//set-title": {
		Argsn: 2,
		Doc:   "Set title of gtk window.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch obj := arg0.(type) {
			case env.Native:
				switch val := arg1.(type) {
				case env.String:
					obj.Value.(*gtk.Window).SetTitle(val.Value)
					return obj
				default:
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "gtk-window//set-title")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "gtk-window//set-title")
			}
		},
	},
	"gtk-window//show": {
		Argsn: 1,
		Doc:   "Show gtk window.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch obj := arg0.(type) {
			case env.Native:
				obj.Value.(*gtk.Window).ShowAll()
				return obj
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "gtk-window//show")
			}
		},
	},

	"gtk-new-label": {
		Argsn: 0,
		Doc:   "TODODOC.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			//switch str := arg0.(type) {
			//case env.String:
			l, _ := gtk.LabelNew("...")
			return *env.NewNative(ps.Idx, l, "gtk-label")
			//default:
			//	return env.NewError("arg 1 should be String")
			//}
		},
	},
	"gtk-label//set-text": {
		Argsn: 2,
		Doc:   "Set text on gtk label.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch win := arg0.(type) {
			case env.Native:
				switch str := arg1.(type) {
				case env.String:
					win.Value.(*gtk.Label).SetText(str.Value)
					return win
				default:
					return MakeArgError(ps, 1, []env.Type{env.StringType}, "gtk-label//set-text")
				}

			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "gtk-label//set-text")
			}
		},
	},
	"gtk-label//set-tooltip": {
		Argsn: 2,
		Doc:   "Set tool tip on gtk label.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch win := arg0.(type) {
			case env.Native:
				switch str := arg1.(type) {
				case env.String:
					win.Value.(*gtk.Label).SetTooltipText(str.Value)
					return win
				default:
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "gtk-label//set-tooltip")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "gtk-label//set-tooltip")
			}
		},
	},
	"gtk-window//add-to": {
		Argsn: 2,
		Doc:   "TODODOC.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch win := arg0.(type) {
			case env.Native:
				switch wig := arg1.(type) {
				case env.Native:
					win.Value.(*gtk.Window).Add(wig.Value.(gtk.IWidget))
					return wig
				default:
					return MakeArgError(ps, 2, []env.Type{env.NativeType}, "gtk-window//add-to")
				}

			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "gtk-window//add-to")
			}
		},
	},
	"gtk-window//set-size": {
		Argsn: 3,
		Doc:   "Set window size in gtk gui.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch win := arg0.(type) {
			case env.Native:
				switch x := arg1.(type) {
				case env.Integer:
					switch y := arg1.(type) {
					case env.Integer:
						win.Value.(*gtk.Window).SetDefaultSize(int(x.Value), int(y.Value))
						return win
					default:
						return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "gtk-window//set-size")
					}
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "gtk-window//set-size")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "gtk-window//set-size")
			}
		},
	},
}
