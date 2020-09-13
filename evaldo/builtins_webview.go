// +build b_webview

package evaldo

import "C"

import (
	//	"fmt"
	"rye/env"

	"github.com/webview/webview"
)

var Builtins_webview = map[string]*env.Builtin{

	"new-webview": {
		Argsn: 0,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			wv := webview.New(true)
			return *env.NewNative(env1.Idx, wv, "webview")
		},
	},
	"webview//set-title": {
		Argsn: 2,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch obj := arg0.(type) {
			case env.Native:
				switch val := arg1.(type) {
				case env.String:
					obj.Value.(webview.WebView).SetTitle(val.Value)
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
	"webview//run": {
		Argsn: 1,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch obj := arg0.(type) {
			case env.Native:
				obj.Value.(webview.WebView).Run()
				return obj
			default:
				return env.NewError("arg 2 should be Native")
			}
			return env.NewError("arg 2 should be Native")
		},
	},
	"webview//destroy": {
		Argsn: 1,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch obj := arg0.(type) {
			case env.Native:
				obj.Value.(webview.WebView).Destroy()
				return obj
			default:
				return env.NewError("arg 2 should be Native")
			}
			return env.NewError("arg 2 should be Native")
		},
	},
	"webview//navigate": {
		Argsn: 2,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch win := arg0.(type) {
			case env.Native:
				switch str := arg1.(type) {
				case env.String:
					win.Value.(webview.WebView).Navigate(str.Value)
					return win
				default:
					return env.NewError("arg 2 should be String")
				}

			default:
				return env.NewError("arg 1 should be Native")
			}
		},
	},
	"webview//set-size": {
		Argsn: 3,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch win := arg0.(type) {
			case env.Native:
				switch x := arg1.(type) {
				case env.Integer:
					switch y := arg2.(type) {
					case env.Integer:
						win.Value.(webview.WebView).SetSize(int(x.Value), int(y.Value), webview.HintNone)
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

	"webview//bind": {
		Argsn: 3,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch win := arg0.(type) {
			case env.Native:
				switch word := arg1.(type) {
				case env.Tagword:
					switch fn := arg2.(type) {
					case env.Function:
						if fn.Argsn == 0 {
							win.Value.(webview.WebView).Bind(env1.Idx.GetWord(word.Index), func() interface{} {
								CallFunction(fn, env1, nil, false, env1.Ctx)
								return resultToJS(env1.Res)
							})
						}
						if fn.Argsn == 1 {
							win.Value.(webview.WebView).Bind(env1.Idx.GetWord(word.Index), func(a0 interface{}) interface{} {
								a0_ := JSToRye(a0)
								CallFunction(fn, env1, a0_, false, env1.Ctx)
								return resultToJS(env1.Res)
							})
						}
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
