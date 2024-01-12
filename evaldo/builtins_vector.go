package evaldo

import (
	"github.com/refaktor/rye/env"

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
		Doc:   "Creates vector object.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s := arg0.(type) {
			case env.Block:
				data := ArrayFloat32FromSeries(s.Series)
				val, err := govector.AsVector(data)
				if err != nil {
					return MakeError(ps, err.Error())
				}
				return *env.NewVector(val)
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType}, "vector")
			}
		},
	},

	"normalize": {
		Argsn: 1,
		Doc:   "Normalize vector.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch val := arg0.(type) {
			case env.Vector:
				return *env.NewDecimal(govector.Norm(val.Value, 2.0))
			default:
				return MakeArgError(ps, 1, []env.Type{env.VectorType}, "norm")
			}
		},
	},

	"std-deviation?": {
		Argsn: 1,
		Doc:   "Calculate standard deviation of a vector",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch val := arg0.(type) {
			case env.Vector:
				return *env.NewDecimal(val.Value.Sd())
			default:
				return MakeArgError(ps, 1, []env.Type{env.VectorType}, "std-deviation")
			}
		},
	},

	"cosine-similarity?": {
		Argsn: 2,
		Doc:   "Calculate cosine similarity.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch v1 := arg0.(type) {
			case env.Vector:
				switch v2 := arg1.(type) {
				case env.Vector:
					res, err := govector.Cosine(v1.Value, v2.Value)
					if err != nil {
						return MakeBuiltinError(ps, err.Error(), "cosine-similarity")
					}
					return *env.NewDecimal(res)
				default:
					return MakeArgError(ps, 2, []env.Type{env.VectorType}, "cosine-similarity")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.VectorType}, "cosine-similarity")
			}
		},
	},

	"correlation": {
		Argsn: 2,
		Doc:   "Get correlation between two vectors",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch v1 := arg0.(type) {
			case env.Vector:
				switch v2 := arg1.(type) {
				case env.Vector:
					res, err := govector.Cor(v1.Value, v2.Value)
					if err != nil {
						return MakeError(ps, err.Error())
					}
					return *env.NewDecimal(res)
				default:
					return MakeArgError(ps, 2, []env.Type{env.VectorType}, "correlation")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.VectorType}, "correlation")
			}
		},
	},

	"dot-product": {
		Argsn: 2,
		Doc:   "Calculate dot product between two vectors.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch v1 := arg0.(type) {
			case env.Vector:
				switch v2 := arg1.(type) {
				case env.Vector:
					res, err := govector.DotProduct(v1.Value, v2.Value)
					if err != nil {
						return MakeError(ps, err.Error())
					}
					return *env.NewDecimal(res)
				default:
					return MakeArgError(ps, 2, []env.Type{env.VectorType}, "dot-product")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.VectorType}, "dot-product")
			}
		},
	},
}
