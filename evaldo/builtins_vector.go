package evaldo

import (
	"rye/env"

	"github.com/drewlanenga/govector"
)

func ArrayFloat32FromSeries(block env.TSeries) []float32 {
	data := make([]float32, block.Len())
	for block.Pos() < block.Len() {
		i := block.Pos()
		k1 := block.Pop()
		switch k := k1.(type) {
		case env.Integer:
			data[i] = float32(k.Value)
		case env.Decimal:
			data[i] = float32(k.Value)
		}
	}
	return data
}

var Builtins_vector = map[string]*env.Builtin{

	"vector": {
		Argsn: 1,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s := arg0.(type) {
			case env.Block:

				data := ArrayFloat32FromSeries(s.Series)
				val, err := govector.AsVector(data)
				if err != nil {
					return MakeError(env1, err.Error())
				}
				return *env.NewNative(env1.Idx, val, "vector")
			default:
				return MakeError(env1, "Arg1 not String")
			}
		},
	},

	"vector//len": {
		Argsn: 1,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch val := arg0.(type) {
			case env.Native:
				return env.Integer{int64(val.Value.(govector.Vector).Len())}
			default:
				return MakeError(env1, "Arg1 not String")
			}
		},
	},

	"vector//norm": {
		Argsn: 1,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch val := arg0.(type) {
			case env.Native:
				return env.Decimal{govector.Norm(val.Value.(govector.Vector), 2.0)}
			default:
				return MakeError(env1, "Arg1 not String")
			}
		},
	},

	"vector//std-deviation": {
		Argsn: 1,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch val := arg0.(type) {
			case env.Native:
				return env.Decimal{val.Value.(govector.Vector).Sd()}
			default:
				return MakeError(env1, "Arg1 not String")
			}
		},
	},

	"vector//mean": {
		Argsn: 1,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch val := arg0.(type) {
			case env.Native:
				return env.Decimal{val.Value.(govector.Vector).Mean()}
			default:
				return MakeError(env1, "Arg1 not String")
			}
		},
	},

	"vector//sum": {
		Argsn: 1,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch val := arg0.(type) {
			case env.Native:
				return env.Decimal{val.Value.(govector.Vector).Sum()}
			default:
				return MakeError(env1, "Arg1 not String")
			}
		},
	},

	"vector//cosine": {
		Argsn: 2,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch v1 := arg0.(type) {
			case env.Native:
				switch v2 := arg1.(type) {
				case env.Native:
					res, err := govector.Cosine(v1.Value.(govector.Vector), v2.Value.(govector.Vector))
					if err != nil {
						return MakeError(env1, err.Error())
					}
					return env.Decimal{res}
				default:
					return MakeError(env1, "Arg2 not Native")
				}
			default:
				return MakeError(env1, "Arg1 not Native")
			}
		},
	},

	"vector//correlation": {
		Argsn: 2,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch v1 := arg0.(type) {
			case env.Native:
				switch v2 := arg1.(type) {
				case env.Native:
					res, err := govector.Cor(v1.Value.(govector.Vector), v2.Value.(govector.Vector))
					if err != nil {
						return MakeError(env1, err.Error())
					}
					return env.Decimal{res}
				default:
					return MakeError(env1, "Arg2 not Native")
				}
			default:
				return MakeError(env1, "Arg1 not Native")
			}
		},
	},

	"vector//dot-product": {
		Argsn: 2,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch v1 := arg0.(type) {
			case env.Native:
				switch v2 := arg1.(type) {
				case env.Native:
					res, err := govector.DotProduct(v1.Value.(govector.Vector), v2.Value.(govector.Vector))
					if err != nil {
						return MakeError(env1, err.Error())
					}
					return env.Decimal{res}
				default:
					return MakeError(env1, "Arg2 not Native")
				}
			default:
				return MakeError(env1, "Arg1 not Native")
			}
		},
	},
}
