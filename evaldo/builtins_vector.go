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
				return *env.NewVector(val)
			default:
				return MakeError(env1, "Arg1 not String")
			}
		},
	},

	"norm": {
		Argsn: 1,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch val := arg0.(type) {
			case env.Vector:
				return env.Decimal{govector.Norm(val.Value, 2.0)}
			default:
				return MakeError(env1, "Arg1 not Vector")
			}
		},
	},

	"std-deviation": {
		Argsn: 1,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch val := arg0.(type) {
			case env.Vector:
				return env.Decimal{val.Value.Sd()}
			default:
				return MakeError(env1, "Arg1 not String")
			}
		},
	},

	"cosine-similarity": {
		Argsn: 2,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch v1 := arg0.(type) {
			case env.Vector:
				switch v2 := arg1.(type) {
				case env.Vector:
					res, err := govector.Cosine(v1.Value, v2.Value)
					if err != nil {
						return MakeError(env1, err.Error())
					}
					return env.Decimal{res}
				default:
					return MakeError(env1, "Arg2 not Vector")
				}
			default:
				return MakeError(env1, "Arg1 not Vector")
			}
		},
	},

	"correlation": {
		Argsn: 2,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch v1 := arg0.(type) {
			case env.Vector:
				switch v2 := arg1.(type) {
				case env.Vector:
					res, err := govector.Cor(v1.Value, v2.Value)
					if err != nil {
						return MakeError(env1, err.Error())
					}
					return env.Decimal{res}
				default:
					return MakeError(env1, "Arg2 not Vector")
				}
			default:
				return MakeError(env1, "Arg1 not Vector")
			}
		},
	},

	"dot-product": {
		Argsn: 2,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch v1 := arg0.(type) {
			case env.Vector:
				switch v2 := arg1.(type) {
				case env.Vector:
					res, err := govector.DotProduct(v1.Value, v2.Value)
					if err != nil {
						return MakeError(env1, err.Error())
					}
					return env.Decimal{res}
				default:
					return MakeError(env1, "Arg2 not Vector")
				}
			default:
				return MakeError(env1, "Arg1 not Vector")
			}
		},
	},
}
