// builtins.go
package evaldo

import (
	"Rejy_go_v1/env"
	"Rejy_go_v1/loader"
	"fmt"
	"time"
)

var builtins = map[string]*env.Builtin{

	"oneone": {
		Argsn: 0,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return env.Integer{11}
		},
	},
	"inc": {
		Argsn: 1,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch arg := arg0.(type) {
			case env.Integer:
				return env.Integer{1 + arg.Value}
			default:
				return env.NewError("argument to `len` not supported, got %s")
			}
		},
	},

	// BASIC FUNCTIONS WITH NUMBERS

	"add": {
		Argsn: 2,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch arg := arg0.(type) {
			case env.Integer:
				return env.Integer{arg.Value + arg1.(env.Integer).Value}
			default:
				return env.NewError("argument to `len` not supported, got %s")
			}
		},
	},
	"subtract": {
		Argsn: 2,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch arg := arg0.(type) {
			case env.Integer:
				return env.Integer{arg.Value - arg1.(env.Integer).Value}
			default:
				return env.NewError("argument to `len` not supported, got %s")
			}
		},
	},
	"multiply": {
		Argsn: 2,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch arg := arg0.(type) {
			case env.Integer:
				return env.Integer{arg.Value * arg1.(env.Integer).Value}
			default:
				return env.NewError("argument to `len` not supported, got %s")
			}
		},
	},

	"equals": {
		Argsn: 2,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			var res int64
			if arg0.GetKind() == arg1.GetKind() && arg0.Inspect(*env1.Idx) == arg1.Inspect(*env1.Idx) {
				res = 1
			} else {
				res = 0
			}
			return env.Integer{res}
		},
	},
	"greater": {
		Argsn: 2,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch arg := arg0.(type) {
			case env.Integer:
				var res int64
				if arg.Value > arg1.(env.Integer).Value {
					res = 1
				} else {
					res = 0
				}
				return env.Integer{res}
			default:
				return env.NewError("argument to `len` not supported, got %s")
			}
		},
	},
	"lesser": {
		Argsn: 2,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch arg := arg0.(type) {
			case env.Integer:
				var res int64
				if arg.Value < arg1.(env.Integer).Value {
					res = 1
				} else {
					res = 0
				}
				return env.Integer{res}
			default:
				return env.NewError("argument to `len` not supported, got %s")
			}
		},
	},

	// BASIC GENERAL FUNCTIONS

	"inspect": {
		Argsn: 1,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			fmt.Println(arg0.Inspect(*env1.Idx))
			return arg0
		},
	},
	"prin": {
		Argsn: 1,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			fmt.Print(arg0.Probe(*env1.Idx))
			return arg0
		},
	},
	"print": {
		Argsn: 1,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			fmt.Println(arg0.Probe(*env1.Idx))
			return arg0
		},
	},

	// CONTROL WORDS

	"unless": {
		Argsn: 2,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch cond := arg0.(type) {
			case env.Integer:
				switch bloc := arg1.(type) {
				case env.Block:
					if cond.Value == 0 {
						ser := ps.Ser
						ps.Ser = bloc.Series
						EvalBlock(ps)
						ps.Ser = ser
						return ps.Res
					}
				}
			}
			return nil
		},
	},

	"if": {
		Argsn: 2,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch cond := arg0.(type) {
			case env.Integer:
				switch bloc := arg1.(type) {
				case env.Block:
					if cond.Value > 0 {
						ser := ps.Ser
						ps.Ser = bloc.Series
						EvalBlock(ps)
						ps.Ser = ser
						return ps.Res
					}
				}
			}
			return nil
		},
	},

	"either": {
		Argsn: 3,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			//arg0.Trace("")
			//arg1.Trace("")
			//arg2.Trace("")
			switch cond := arg0.(type) {
			case env.Integer:
				switch bloc1 := arg1.(type) {
				case env.Block:
					switch bloc2 := arg2.(type) {
					case env.Block:
						ser := ps.Ser
						if cond.Value > 0 {
							ps.Ser = bloc1.Series
							ps.Ser.Reset()
						} else {
							ps.Ser = bloc2.Series
							ps.Ser.Reset()
						}
						EvalBlock(ps)
						ps.Ser = ser
						return ps.Res
					}
				}
			}
			return nil
		},
	},

	"do": {
		Argsn: 1,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch bloc := arg0.(type) {
			case env.Block:
				ser := ps.Ser
				ps.Ser = bloc.Series
				EvalBlock(ps)
				ps.Ser = ser
				return ps.Res
			}
			return nil
		},
	},

	"skip": {
		Argsn: 2,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch bloc := arg1.(type) {
			case env.Block:
				res := arg0
				ser := ps.Ser
				ps.Ser = bloc.Series
				EvalBlock(ps)
				ps.Ser = ser
				return res
			}
			return nil
		},
	},

	"dotime": {
		Argsn: 1,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch bloc := arg0.(type) {
			case env.Block:
				ser := ps.Ser
				ps.Ser = bloc.Series
				start := time.Now()
				EvalBlock(ps)
				t := time.Now()
				elapsed := t.Sub(start)
				ps.Ser = ser
				return env.Integer{elapsed.Nanoseconds() / 1000000}
			}
			return nil
		},
	},

	"loop": {
		Argsn: 2,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch cond := arg0.(type) {
			case env.Integer:
				switch bloc := arg1.(type) {
				case env.Block:
					ser := ps.Ser
					ps.Ser = bloc.Series
					for i := 0; int64(i) < cond.Value; i++ {
						ps = EvalBlock(ps)
						ps.Ser.Reset()
					}
					ps.Ser = ser
					return ps.Res
				}
			}
			return nil
		},
	},

	//test if we can do recur similar to clojure one. Since functions in rejy are of fixed arity we would need recur1 recur2 recur3 and recur [ ] which is less optimal
	//otherwise word recur could somehow be bound to correct version or args depending on number of args of func. Try this at first.
	"recur1if": { //recur1-if
		Argsn: 2,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch cond := arg0.(type) {
			case env.Integer:
				if cond.Value > 0 {
					switch arg := arg1.(type) {
					case env.Integer:
						ps.Env.Set(ps.Args[0], arg)
						ps.Ser.Reset()
						return nil
					}
				} else {
					return ps.Res
				}
			}
			return nil
		},
	},

	"recur2if": { //recur1-if
		Argsn: 3,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			//arg0.Trace("a0")
			//arg1.Trace("a1")
			//arg2.Trace("a2")
			switch cond := arg0.(type) {
			case env.Integer:
				if cond.Value > 0 {
					switch argi1 := arg1.(type) {
					case env.Integer:
						switch argi2 := arg2.(type) {
						case env.Integer:
							ps.Env.Set(ps.Args[0], argi1)
							ps.Env.Set(ps.Args[1], argi2)
							ps.Ser.Reset()
							return ps.Res
						}
					}
				} else {
					return ps.Res
				}
			}
			return nil
		},
	},

	"recur3if": { //recur1-if
		Argsn: 4,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			//arg0.Trace("a0")
			//arg1.Trace("a1")
			//arg2.Trace("a2")
			switch cond := arg0.(type) {
			case env.Integer:
				if cond.Value > 0 {
					switch argi1 := arg1.(type) {
					case env.Integer:
						switch argi2 := arg2.(type) {
						case env.Integer:
							switch argi3 := arg3.(type) {
							case env.Integer:
								ps.Env.Set(ps.Args[0], argi1)
								ps.Env.Set(ps.Args[1], argi2)
								ps.Env.Set(ps.Args[2], argi3)
								ps.Ser.Reset()
								return ps.Res
							}
						}
					}
				} else {
					return ps.Res
				}
			}
			return nil
		},
	},

	"fn": {
		Argsn: 2,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch args := arg0.(type) {
			case env.Block:
				switch body := arg1.(type) {
				case env.Block:
					//spec := []env.Object{env.Word{aaaidx}}
					//body := []env.Object{env.Word{printidx}, env.Word{aaaidx}, env.Word{recuridx}, env.Word{greateridx}, env.Integer{99}, env.Word{aaaidx}, env.Word{incidx}, env.Word{aaaidx}}
					return *env.NewFunction(args, body)
				}
			}
			return nil
		},
	},

	// BASIC STRING FUNCTIONS

	"left": {
		Argsn: 2,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.String:
				switch s2 := arg1.(type) {
				case env.Integer:
					return env.String{s1.Value[0:s2.Value]}
				}
			}
			return nil
		},
	},

	"middle": {
		Argsn: 3,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.String:
				switch s2 := arg1.(type) {
				case env.Integer:
					switch s3 := arg2.(type) {
					case env.Integer:
						return env.String{s1.Value[s2.Value:s3.Value]}
					}
				}
			}
			return nil
		},
	},

	"right": {
		Argsn: 2,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.String:
				switch s2 := arg1.(type) {
				case env.Integer:
					return env.String{s1.Value[len(s1.Value)-int(s2.Value):]}
				}
			}
			return nil
		},
	},

	"join": {
		Argsn: 2,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.String:
				switch s2 := arg1.(type) {
				case env.String:
					return env.String{s1.Value + s2.Value}
				}
			}
			return nil
		},
	},

	"join3": {
		Argsn: 3,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.String:
				switch s2 := arg1.(type) {
				case env.String:
					switch s3 := arg2.(type) {
					case env.String:
						return env.String{s1.Value + s2.Value + s3.Value}
					}
				}
			}
			return nil
		},
	},

	// BASIC SERIES FUNCTIONS

	"nth": {
		Argsn: 2,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Block:
				switch s2 := arg1.(type) {
				case env.Integer:
					return s1.Series.Get(int(s2.Value - 1))
				}
			}
			return nil
		},
	},
	"length": {
		Argsn: 1,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Block:
				return env.Integer{int64(s1.Series.Len())}
			}
			return nil
		},
	},
	"peek": {
		Argsn: 1,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Block:
				return s1.Series.Peek()
			}
			return nil
		},
	},
	"pop": {
		Argsn: 1,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Block:
				return s1.Series.Pop()
			}
			return nil
		},
	},
	"pos": {
		Argsn: 1,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Block:
				return env.Integer{int64(s1.Series.Pos())}
			}
			return nil
		},
	},
	"next": {
		Argsn: 1,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Block:
				s1.Series.Next()
				return s1
			}
			return nil
		},
	},
	// FUNCTIONALITY AROUND GENERIC METHODS
	// generic <integer> <add> fn [ a b ] [ a + b ] // tagwords are temporary here
	"generic": {
		Argsn: 3,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Tagword:
				switch s2 := arg1.(type) {
				case env.Tagword:
					switch s3 := arg2.(type) {
					case env.Object:
						fmt.Println(s1.Index)
						fmt.Println(s2.Index)
						fmt.Println("Generic")

						registerGeneric(ps, s1.Index, s2.Index, s3)
					}
				}
			}
			return nil
		},
	},
}

func RegisterBuiltins(ps *env.ProgramState) {
	RegisterBuiltins2(builtins, ps)
	RegisterBuiltins2(Builtins_web, ps)
	RegisterBuiltins2(Builtins_sqlite, ps)
}

func RegisterBuiltins2(builtins map[string]*env.Builtin, ps *env.ProgramState) {
	for k, v := range builtins {
		bu := env.NewBuiltin(v.Fn, v.Argsn)
		registerBuiltin(ps, k, *bu)
	}
}

func registerBuiltin(ps *env.ProgramState, word string, builtin env.Builtin) {
	// indexWord
	idxs := loader.GetIdxs()
	idx := idxs.IndexWord(word)
	// set global word with builtin
	ps.Env.Set(idx, builtin)
}
