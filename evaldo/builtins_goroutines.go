// ### +build !b_tiny

package evaldo

// import "C"

import (
	"fmt"
	"rye/env"
	"time"

	"github.com/jinzhu/copier"
)

func strimpg() { fmt.Println("") }

var Builtins_goroutines = map[string]*env.Builtin{

	"sleep": {
		Argsn: 1,
		Doc:   "Accepts an integer and Sleeps for given number of miliseconds.",
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch arg := arg0.(type) {
			case env.Integer:
				time.Sleep(time.Duration(int(arg.Value)) * time.Millisecond)
				return arg
			default:
				return makeError(env1, "Arg 1 should be Integer.")
			}
		},
	},

	"go-with": {
		Argsn: 2,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch arg := arg0.(type) {
			case env.Object:
				switch handler := arg1.(type) {
				case env.Function:
					go func() {
						env1.FailureFlag = false
						env1.ErrorFlag = false
						env1.ReturnFlag = false
						psTemp := env.ProgramState{}
						copier.Copy(&psTemp, &env1)
						CallFunction(handler, &psTemp, arg, false, nil)
						// CallFunctionArgs2(handler, &psTemp, arg, *env.NewNative(psTemp.Idx, "asd", "Go-server-context"), nil)
					}()
					return arg0
				default:
					env1.FailureFlag = true
					return env.NewError("arg0 should be string")
				}
			default:
				env1.FailureFlag = true
				return env.NewError("arg0 should be string")
			}
		},
	},

	"new-channel": {
		Argsn: 1,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch buflen := arg0.(type) {
			case env.Integer:
				//fmt.Println(str.Value)
				ch := make(chan *env.Object, int(buflen.Value))
				return *env.NewNative(ps.Idx, ch, "Rye-channel")
			default:
				ps.FailureFlag = true
				return env.NewError("first arg should be integer")
			}

		},
	},
	"Rye-channel//read": {
		Argsn: 1,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch chn := arg0.(type) {
			case env.Native:
				msg := <-chn.Value.(chan *env.Object)
				return *msg
			default:
				ps.FailureFlag = true
				return env.NewError("arg 1 should be Uri")
			}
		},
	},
	"Rye-channel//send": {
		Argsn: 2,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch chn := arg0.(type) {
			case env.Native:
				chn.Value.(chan *env.Object) <- &arg1
				return arg0
			default:
				ps.FailureFlag = true
				return env.NewError("first ar0 should be native")
			}
		},
	},
}
