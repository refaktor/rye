//go:build b_nats
// +build b_nats

package evaldo

import "C"

import (
	"fmt"
	"rye/env"

	nats "github.com/nats-io/nats.go"
)

func strimp() { fmt.Println("") }

var Builtins_nats = map[string]*env.Builtin{

	"nats-schema//open": {
		Argsn: 1,
		Doc:   "TODODOC",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch arg0.(type) {
			case env.Uri:
				//fmt.Println(str.Path)
				nc, _ := nats.Connect("demo.nats.io")
				return *env.NewNative(ps.Idx, nc, "Rye-nats")
			default:
				return MakeArgError(ps, 1, []env.Type{env.UriType}, "nats-schema//open")
			}

		},
	},

	"Rye-nats//pub": {
		Argsn: 3,
		Doc:   "TODODOC",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch con := arg0.(type) {
			case env.Native:
				switch subj := arg1.(type) {
				case env.String:
					switch msg := arg2.(type) {
					case env.String:
						con.Value.(*nats.Conn).Publish(subj.Value, []byte(msg.Value))
						return arg0
					default:
						return MakeArgError(ps, 3, []env.Type{env.StringType}, "Rye-nats//pub")
					}
				default:
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "Rye-nats//pub")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Rye-nats//pub")
			}
		},
	},

	"Rye-nats//sub-do": {
		Argsn: 3,
		Doc:   "TODODOC",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch con := arg0.(type) {
			case env.Native:
				switch subj := arg1.(type) {
				case env.String:
					switch bloc := arg2.(type) {
					case env.Block:
						con.Value.(*nats.Conn).Subscribe(subj.Value, func(m *nats.Msg) {
							ser := ps.Ser
							ps.Ser = bloc.Series
							EvalBlockInj(ps, env.String{string(m.Data)}, true)
							ps.Ser = ser
						})
						return arg0
					default:
						return MakeArgError(ps, 3, []env.Type{env.BlockType}, "Rye-nats//sub-do")
					}
				default:
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "Rye-nats//sub-do")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Rye-nats//sub-do")
			}
		},
	},

	"nats-schema//chan": {
		Argsn: 1,
		Doc:   "TODODOC",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch arg0.(type) {
			case env.Uri:
				//fmt.Println(str.Value)
				ch := make(chan *nats.Msg, 64)
				return *env.NewNative(ps.Idx, ch, "Rye-nats-chan")
			default:
				return MakeArgError(ps, 1, []env.Type{env.UriType}, "nats-schema//chan")
			}

		},
	},
	"Rye-nats-chan//sub": {
		Argsn: 3,
		Doc:   "TODODOC",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch chn := arg0.(type) {
			case env.Native:
				switch con := arg1.(type) {
				case env.Native:
					switch subj := arg2.(type) {
					case env.String:
						sub, _ := con.Value.(*nats.Conn).ChanSubscribe(subj.Value, chn.Value.(chan *nats.Msg))
						return env.NewNative(ps.Idx, sub, "Rye-nats-chan-sub")
					default:
						return MakeArgError(ps, 3, []env.Type{env.StringType}, "Rye-nats-chan//sub")
					}
				default:
					return MakeArgError(ps, 2, []env.Type{env.NativeType}, "Rye-nats-chan//sub")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Rye-nats-chan//sub")
			}
		},
	},
	"Rye-nats-chan//read": {
		Argsn: 1,
		Doc:   "TODODOC",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch chn := arg0.(type) {
			case env.Native:
				msg := <-chn.Value.(chan *nats.Msg)
				return env.String{string(msg.Data)}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Rye-nats-chan//read")
			}
		},
	},
}
