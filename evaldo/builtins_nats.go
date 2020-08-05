package evaldo

import "C"

import (
	"Ryelang/env"
	"fmt"

	nats "github.com/nats-io/nats.go"
)

func strimp() { fmt.Println("") }

var Builtins_nats = map[string]*env.Builtin{

	"nats-schema//open": {
		Argsn: 1,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch arg0.(type) {
			case env.Uri:
				//fmt.Println(str.Path)
				nc, _ := nats.Connect("demo.nats.io")
				return *env.NewNative(env1.Idx, nc, "Rye-nats")
			default:
				return env.NewError("arg 1 should be Uri")
			}

		},
	},

	"Rye-nats//pub": {
		Argsn: 3,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch con := arg0.(type) {
			case env.Native:
				switch subj := arg1.(type) {
				case env.String:
					switch msg := arg2.(type) {
					case env.String:
						con.Value.(*nats.Conn).Publish(subj.Value, []byte(msg.Value))
						return arg0
					default:
						return env.NewError("arg 3 should be string")
					}
				default:
					return env.NewError("arg 2 should be string")
				}
			default:
				return env.NewError("arg 1 should be Native")
			}
		},
	},

	"Rye-nats//sub-do": {
		Argsn: 3,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch con := arg0.(type) {
			case env.Native:
				switch subj := arg1.(type) {
				case env.String:
					switch bloc := arg2.(type) {
					case env.Block:
						con.Value.(*nats.Conn).Subscribe(subj.Value, func(m *nats.Msg) {
							ser := env1.Ser
							env1.Ser = bloc.Series
							EvalBlockInj(env1, env.String{string(m.Data)}, true)
							env1.Ser = ser
						})
						return arg0
					default:
						return env.NewError("arg 3 should be string")
					}
				default:
					return env.NewError("arg 2 should be string")
				}
			default:
				return env.NewError("arg 1 should be Native")
			}
		},
	},

	"nats-schema//chan": {
		Argsn: 1,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch arg0.(type) {
			case env.Uri:
				//fmt.Println(str.Value)
				ch := make(chan *nats.Msg, 64)
				return *env.NewNative(env1.Idx, ch, "Rye-nats-chan")
			default:
				return env.NewError("arg 1 should be Uri")
			}

		},
	},
	"Rye-nats-chan//sub": {
		Argsn: 3,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch chn := arg0.(type) {
			case env.Native:
				switch con := arg1.(type) {
				case env.Native:
					switch subj := arg2.(type) {
					case env.String:
						sub, _ := con.Value.(*nats.Conn).ChanSubscribe(subj.Value, chn.Value.(chan *nats.Msg))
						return env.NewNative(env1.Idx, sub, "Rye-nats-chan-sub")
					default:
						return env.NewError("arg 3 should be string")
					}
				default:
					return env.NewError("arg 2 should be string")
				}
			default:
				return env.NewError("arg 1 should be Native")
			}
		},
	},
	"Rye-nats-chan//read": {
		Argsn: 1,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch chn := arg0.(type) {
			case env.Native:
				msg := <-chn.Value.(chan *nats.Msg)
				return env.String{string(msg.Data)}
			default:
				return env.NewError("arg 1 should be Uri")
			}
		},
	},
}
