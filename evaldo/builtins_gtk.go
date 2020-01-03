package evaldo

import "C"

import (
	"Ryelang/env"

	"github.com/gotk3/gotk3/gtk"
)

var Builtins_gtk = map[string]*env.Builtin{

	"gtk-init": {
		Argsn: 0,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			gtk.Init(nil)
			return env.Integer{1}
		},
	},
	"gtk-main": {
		Argsn: 0,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			gtk.Main()
			return env.Integer{1}
		},
	},
	"gtk-new-window": {
		Argsn: 0,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			win, _ := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
			return *env.NewNative(env1.Idx, win, "gtk-window")
		},
	},
	"gtk-window//set-title": {
		Argsn: 2,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch obj := arg0.(type) {
			case env.Native:
				switch val := arg1.(type) {
				case env.String:
					obj.Value.(*gtk.Window).SetTitle(val.Value)
					return obj
				default:
					return env.NewError("arg 2 should be String")
				}
			default:
				return env.NewError("arg 2 should be Native")
			}
			return env.NewError("arg 2 should be Native")

		},
	},
	"gtk-window//show": {
		Argsn: 1,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch obj := arg0.(type) {
			case env.Native:
				obj.Value.(*gtk.Window).ShowAll()
				return obj
			default:
				return env.NewError("arg 2 should be Native")
			}
			return env.NewError("arg 2 should be Native")
		},
	},

	"gtk-new-label": {
		Argsn: 0,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			//switch str := arg0.(type) {
			//case env.String:
			l, _ := gtk.LabelNew("...")
			return *env.NewNative(env1.Idx, l, "gtk-label")
			//default:
			//	return env.NewError("arg 1 should be String")
			//}
		},
	},
	"gtk-label//set-text": {
		Argsn: 2,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch win := arg0.(type) {
			case env.Native:
				switch str := arg1.(type) {
				case env.String:
					win.Value.(*gtk.Label).SetText(str.Value)
					return win
				default:
					return env.NewError("arg 2 should be String")
				}

			default:
				return env.NewError("arg 1 should be Native")
			}
		},
	},
	"gtk-label//set-tooltip": {
		Argsn: 2,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch win := arg0.(type) {
			case env.Native:
				switch str := arg1.(type) {
				case env.String:
					win.Value.(*gtk.Label).SetTooltipText(str.Value)
					return win
				default:
					return env.NewError("arg 2 should be String")
				}

			default:
				return env.NewError("arg 1 should be Native")
			}
		},
	},
	"gtk-window//add-to": {
		Argsn: 2,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch win := arg0.(type) {
			case env.Native:
				switch wig := arg1.(type) {
				case env.Native:
					win.Value.(*gtk.Window).Add(wig.Value.(gtk.IWidget))
					return wig
				default:
					return env.NewError("arg 2 should be Native")
				}

			default:
				return env.NewError("arg 1 should be Native")
			}
		},
	},
	"gtk-window//set-size": {
		Argsn: 3,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch win := arg0.(type) {
			case env.Native:
				switch x := arg1.(type) {
				case env.Integer:
					switch y := arg1.(type) {
					case env.Integer:
						win.Value.(*gtk.Window).SetDefaultSize(int(x.Value), int(y.Value))
						return win
					default:
						return env.NewError("arg 3 should be Int")
					}
				default:
					return env.NewError("arg 2 should be Int")
				}
			default:
				return env.NewError("arg 1 should be Native")
			}
		},
	},
}
