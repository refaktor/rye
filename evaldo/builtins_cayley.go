//go:build b_cayley
// +build b_cayley

package evaldo

import (
	"fmt"
	"rye/env"

	"github.com/cayleygraph/cayley"
	"github.com/cayleygraph/cayley/graph"
	_ "github.com/cayleygraph/cayley/graph/kv/bolt"
	"github.com/cayleygraph/cayley/graph/path"
	"github.com/cayleygraph/quad"
)

func CayleyPath_EvalBlock(ps *env.ProgramState, pth *cayley.Path) *cayley.Path {
	for ps.Ser.Pos() < ps.Ser.Len() {
		object := ps.Ser.Pop()
		switch obj := object.(type) {
		case env.String:
			// cayley.StartPath
			//fmt.Println("asda")
			//fmt.Println(obj.Value)
			pth = pth.Is(quad.String(obj.Value))
			// quad.String("hello").Out(quad.String("mr"))
		case env.Word:
			//fmt.Println("asda222")
			idx, found := ps.Idx.GetIndex("out")
			if found && obj.Index == idx {
				switch arrg0 := ps.Ser.Pop().(type) {
				case env.String:
					pth = pth.Out(quad.String(arrg0.Value))
				default:
					pth = pth.Out()
				}
			}
		default:
			fmt.Println("OTHER CALYEY QUERY NODE")
			return pth
		}
	}
	return pth
}

var Builtins_cayley = map[string]*env.Builtin{

	"init-cayley-store": {
		Argsn: 2,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch storage := arg0.(type) {
			case env.Tagword:
				switch dir := arg1.(type) {
				case env.String:
					err := graph.InitQuadStore(ps.Idx.GetWord(storage.Index), dir.Value, nil)
					if err != nil {
						ps.FailureFlag = true
						return env.NewError("Error initialising the directory: " + err.Error())
					}
					return arg1
				default:
					ps.FailureFlag = true
					return env.NewError("arg 2 should be strin")
				}
			default:
				ps.FailureFlag = true
				return env.NewError("arg 1 should be tagword")
			}
		},
	},

	"new-cayley-graph": {
		Argsn: 2,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch storage := arg0.(type) {
			case env.Tagword:
				switch dir := arg1.(type) {
				case env.String:
					graph, err := cayley.NewGraph(ps.Idx.GetWord(storage.Index), dir.Value, nil)
					if err != nil {
						ps.FailureFlag = true
						return env.NewError("Error initialising the directory: " + err.Error())
					}
					return *env.NewNative(ps.Idx, graph, "cayley-graph")
				default:
					ps.FailureFlag = true
					return env.NewError("arg 2 should be string")
				}
			default:
				ps.FailureFlag = true
				return env.NewError("arg 1 should be native")
			}
		},
	},

	"cayley-graph//add-quad": {
		Argsn: 2,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch graph := arg0.(type) {
			case env.Native:
				switch quad_ := arg1.(type) {
				case env.Native:
					graph.Value.(*cayley.Handle).AddQuad(quad_.Value.(quad.Quad))
					return arg0
				default:
					ps.FailureFlag = true
					return env.NewError("arg 2 should be native")
				}
			default:
				ps.FailureFlag = true
				return env.NewError("arg 1 should be native")
			}
		},
	},

	"new-triple": {
		Argsn: 3,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch v1 := arg0.(type) {
			case env.String:
				switch v2 := arg1.(type) {
				case env.String:
					switch v3 := arg2.(type) {
					case env.String:
						quad_ := quad.Make(v1.Value, v2.Value, v3.Value, nil)
						return *env.NewNative(ps.Idx, quad_, "cayley-quad")
					default:
						ps.FailureFlag = true
						return env.NewError("arg 3 should be string")
					}
				default:
					ps.FailureFlag = true
					return env.NewError("arg 2 should be string")
				}
			default:
				ps.FailureFlag = true
				return env.NewError("arg 1 should be string")
			}

		},
	},

	"cayley-graph//new-path": {
		Argsn: 2,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch store := arg0.(type) {
			case env.Native:
				switch code := arg1.(type) {
				case env.Block:
					// fmt.Println(code.Probe(*ps.Idx))
					ser := ps.Ser
					ps.Ser = code.Series
					// pth := path.NewPath(store.Value.(graph.QuadStore))
					pth := path.StartPath(store.Value.(graph.QuadStore))
					pth = CayleyPath_EvalBlock(ps, pth)
					ps.Ser = ser
					return *env.NewNative(ps.Idx, pth, "cayley-path")
				default:
					ps.FailureFlag = true
					return env.NewError("arg 2 should be string")
				}
			default:
				ps.FailureFlag = true
				return env.NewError("arg 1 should be native")
			}
		},
	},

	"cayley-path//iterate": {
		Argsn: 2,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch path := arg0.(type) {
			case env.Native:
				switch code := arg1.(type) {
				case env.Block:
					// fmt.Println(code.Probe(*ps.Idx))
					// TODO eval block of code second arg
					err := path.Value.(*cayley.Path).Iterate(nil).EachValue(nil, func(value quad.Value) error {
						nativeValue := quad.NativeOf(value) // this converts RDF values to normal Go types
						ser := ps.Ser
						ps.Ser = code.Series
						EvalBlockInj(ps, env.ToRyeValue(nativeValue), true)
						ps.Ser = ser
						//						fmt.Println(nativeValue)
						return nil
					})
					if err != nil {
						ps.FailureFlag = true
						return env.NewError("arg 2 should be string")
					}
					return ps.Res
				default:
					ps.FailureFlag = true
					return env.NewError("arg 2 should be string")
				}
			default:
				ps.FailureFlag = true
				return env.NewError("arg 1 should be native")
			}
		},
	},

	"cayley-token//name-of": {
		Argsn: 3,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return __addAlternative(ps, arg0, arg1, arg2, arg3, arg4)
		},
	},
}
