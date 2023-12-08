//go:build !b_tiny
// +build !b_tiny

package evaldo

// import "C"

import (
	"fmt"
	"rye/env"
	"time"

	"github.com/jinzhu/copier"
	"golang.org/x/sync/errgroup"
)

func strimpg() { fmt.Println("") }

var Builtins_goroutines = map[string]*env.Builtin{

	"sleep": {
		Argsn: 1,
		Doc:   "Accepts an integer and Sleeps for given number of miliseconds.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch arg := arg0.(type) {
			case env.Integer:
				time.Sleep(time.Duration(int(arg.Value)) * time.Millisecond)
				return arg
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType}, "sleep")
			}
		},
	},

	"go-with": {
		Argsn: 2,
		Doc:   "TODODOC.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch arg := arg0.(type) {
			case env.Object:
				switch handler := arg1.(type) {
				case env.Function:
					g := errgroup.Group{}
					g.Go(func() error {
						ps.FailureFlag = false
						ps.ErrorFlag = false
						ps.ReturnFlag = false
						psTemp := env.ProgramState{}
						err := copier.Copy(&psTemp, &ps)
						if err != nil {
							ps.FailureFlag = true
							ps.ErrorFlag = true
							ps.ReturnFlag = true
							return fmt.Errorf("failed to copy ps: %w", err)
						}
						CallFunction(handler, &psTemp, arg, false, nil)
						// CallFunctionArgs2(handler, &psTemp, arg, *env.NewNative(psTemp.Idx, "asd", "Go-server-context"), nil)
						return nil
					})
					if err := g.Wait(); err != nil {
						return MakeBuiltinError(ps, err.Error(), "go-with")
					}
					return arg0
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 2, []env.Type{env.FunctionType}, "go-with")
				}
			default:
				ps.FailureFlag = true
				return MakeBuiltinError(ps, "First argument should be object type.", "go-with")
			}
		},
	},

	"go": {
		Argsn: 1,
		Doc:   "TODODOC.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch handler := arg0.(type) {
			case env.Function:
				g := errgroup.Group{}
				g.Go(func() error {
					ps.FailureFlag = false
					ps.ErrorFlag = false
					ps.ReturnFlag = false
					psTemp := env.ProgramState{}
					err := copier.Copy(&psTemp, &ps)
					if err != nil {
						ps.FailureFlag = true
						ps.ErrorFlag = true
						ps.ReturnFlag = true
						return fmt.Errorf("failed to copy ps: %w", err)
					}
					CallFunction(handler, &psTemp, nil, false, nil)
					// CallFunctionArgs2(handler, &psTemp, arg, *env.NewNative(psTemp.Idx, "asd", "Go-server-context"), nil)
					return nil
				})
				if err := g.Wait(); err != nil {
					return MakeBuiltinError(ps, err.Error(), "go")
				}
				return arg0
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.FunctionType}, "go")
			}
		},
	},

	"new-channel": {
		Argsn: 1,
		Doc:   "TODODOC.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch buflen := arg0.(type) {
			case env.Integer:
				//fmt.Println(str.Value)
				ch := make(chan *env.Object, int(buflen.Value))
				return *env.NewNative(ps.Idx, ch, "Rye-channel")
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.IntegerType}, "new-channel")
			}
		},
	},
	"Rye-channel//read": {
		Argsn: 1,
		Doc:   "TODODOC.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch chn := arg0.(type) {
			case env.Native:
				msg := <-chn.Value.(chan *env.Object)
				return *msg
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Rye-channel//read")
			}
		},
	},
	"Rye-channel//send": {
		Argsn: 2,
		Doc:   "TODODOC.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch chn := arg0.(type) {
			case env.Native:
				chn.Value.(chan *env.Object) <- &arg1
				return arg0
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Rye-channel//send")
			}
		},
	},
}
