//go:build b_webview
// +build b_webview

package evaldo

import "C"

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"mime"
	"net"
	"net/http"
	"path/filepath"

	webview "github.com/webview/webview_go"

	"github.com/refaktor/rye/env"
	"github.com/refaktor/rye/util"
)

var Builtins_webview = map[string]*env.Builtin{

	"new-webview": {
		Argsn: 0,
		Doc:   "Create new webview.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			wv := webview.New(true)
			return *env.NewNative(ps.Idx, wv, "webview")
		},
	},
	"webview//set-title": {
		Argsn: 2,
		Doc:   "Set title for webview.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			fmt.Println("TITLE ---")
			switch obj := arg0.(type) {
			case env.Native:
				switch val := arg1.(type) {
				case env.String:
					obj.Value.(webview.WebView).SetTitle(val.Value)
					return obj
				default:
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "webview//set-title")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "webview//set-title")
			}
		},
	},
	"webview//run": {
		Argsn: 1,
		Doc:   "TODODOC",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch obj := arg0.(type) {
			case env.Native:
				obj.Value.(webview.WebView).Run()
				return obj
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "webview//run")
			}
		},
	},
	"webview//destroy": {
		Argsn: 1,
		Doc:   "TODODOC",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch obj := arg0.(type) {
			case env.Native:
				obj.Value.(webview.WebView).Destroy()
				return obj
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "webview//destroy")
			}
		},
	},
	"webview//navigate": {
		Argsn: 2,
		Doc:   "TODODOC",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch win := arg0.(type) {
			case env.Native:
				switch str := arg1.(type) {
				case env.String:
					win.Value.(webview.WebView).Navigate(str.Value)
					return win
				default:
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "webview//navigate")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "webview//navigate")
			}
		},
	},
	"webview//set-size": {
		Argsn: 3,
		Doc:   "TODODOC",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch win := arg0.(type) {
			case env.Native:
				switch x := arg1.(type) {
				case env.Integer:
					switch y := arg2.(type) {
					case env.Integer:
						win.Value.(webview.WebView).SetSize(int(x.Value), int(y.Value), webview.HintNone)
						return win
					default:
						return MakeArgError(ps, 3, []env.Type{env.IntegerType}, "webview//set-size")
					}
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "webview//set-size")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "webview//set-size")
			}
		},
	},

	"webview//fn-bind": {
		Argsn: 3,
		Doc:   "TODODOC",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			fmt.Println("YOYO1")
			switch win := arg0.(type) {
			case env.Native:
				win.Value.(webview.WebView).Bind("hellorye", func() any {
					fmt.Println("YOYO2")
					return "RETURNED"
				})
				switch word := arg1.(type) {
				case env.Tagword:
					switch fn := arg2.(type) {
					case env.Function:
						if fn.Argsn == 0 {
							win.Value.(webview.WebView).Bind(ps.Idx.GetWord(word.Index), func() any {
								fmt.Println("YOYO3")
								CallFunction(fn, ps, nil, false, ps.Ctx)
								return resultToJS(ps.Res)
							})
						}
						if fn.Argsn == 1 {
							win.Value.(webview.WebView).Bind(ps.Idx.GetWord(word.Index), func(a0 any) any {
								a0_ := env.ToRyeValue(a0)
								CallFunction(fn, ps, a0_, false, ps.Ctx)
								return resultToJS(ps.Res)
							})
						}
						return win
					default:
						return MakeArgError(ps, 3, []env.Type{env.FunctionType}, "webview//fn-bind")
					}
				default:
					return MakeArgError(ps, 2, []env.Type{env.TagwordType}, "webview//fn-bind")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "webview//fn-bind")
			}
		},
	},

	"start-server": {
		Argsn: 0,
		Doc:   "TODODOC",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			go func() {
				fs := http.FileServer(http.Dir("./assets"))
				http.Handle("/", fs)

				log.Println("Listening on :3301...")
				err := http.ListenAndServe(":3301", nil)
				if err != nil {
					log.Fatal(err)
				}
			}()
			return env.String{"http://localhost:3301"}
		},
	},
	"start-serverX": {
		Argsn: 0,
		Doc:   "TODODOC",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {

			ln, err := net.Listen("tcp", "127.0.0.1:0")
			if err != nil {
				log.Fatal(err)
			}
			go func() {
				defer ln.Close()
				http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
					path := r.URL.Path
					if len(path) > 0 && path[0] == '/' {
						path = path[1:]
					}
					if path == "" {
						path = "index.html"
					}
					if bs, err := util.Asset(path); err != nil {
						w.WriteHeader(http.StatusNotFound)
					} else {
						w.Header().Add("Content-Type", mime.TypeByExtension(filepath.Ext(path)))
						io.Copy(w, bytes.NewBuffer(bs))
					}
				})
				log.Fatal(http.Serve(ln, nil))
			}()
			return env.String{"http://" + ln.Addr().String()}
		},
	},
}
