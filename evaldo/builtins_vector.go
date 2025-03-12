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

	//
	// ##### Vector ##### "Vector operations"
	//
	// Tests:
	// equal { vector [ 1 2 3 ] |type? } 'vector
	// Args:
	// * block: block of numbers to convert to a vector
	// Returns:
	// * vector object
	"vector": {
		Argsn: 1,
		Doc:   "Creates a vector from a block of numbers.",
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

	// Tests:
	// equal { vector [ 3 4 ] |normalize } 5.0
	// Args:
	// * vector: vector object to normalize
	// Returns:
	// * decimal representing the L2 norm (Euclidean length) of the vector
	"normalize": {
		Argsn: 1,
		Doc:   "Calculates the L2 norm (Euclidean length) of a vector.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch val := arg0.(type) {
			case env.Vector:
				return *env.NewDecimal(govector.Norm(val.Value, 2.0))
			default:
				return MakeArgError(ps, 1, []env.Type{env.VectorType}, "normalize")
			}
		},
	},

	// Tests:
	// equal { vector [ 1 2 3 4 5 ] |std-deviation? |round 2 } 1.58
	// Args:
	// * vector: vector object
	// Returns:
	// * decimal representing the standard deviation of the vector elements
	"std-deviation?": {
		Argsn: 1,
		Doc:   "Calculates the standard deviation of a vector's elements.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch val := arg0.(type) {
			case env.Vector:
				return *env.NewDecimal(val.Value.Sd())
			default:
				return MakeArgError(ps, 1, []env.Type{env.VectorType}, "std-deviation?")
			}
		},
	},

	// Tests:
	// equal { cosine-similarity? vector [ 1 0 ] vector [ 0 1 ] } 0.0
	// equal { cosine-similarity? vector [ 1 1 ] vector [ 1 1 ] } 1.0
	// Args:
	// * vector1: first vector object
	// * vector2: second vector object
	// Returns:
	// * decimal representing the cosine similarity between the two vectors
	"cosine-similarity?": {
		Argsn: 2,
		Doc:   "Calculates the cosine similarity between two vectors.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch v1 := arg0.(type) {
			case env.Vector:
				switch v2 := arg1.(type) {
				case env.Vector:
					res, err := govector.Cosine(v1.Value, v2.Value)
					if err != nil {
						return MakeBuiltinError(ps, err.Error(), "cosine-similarity?")
					}
					return *env.NewDecimal(res)
				default:
					return MakeArgError(ps, 2, []env.Type{env.VectorType}, "cosine-similarity?")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.VectorType}, "cosine-similarity?")
			}
		},
	},

	// Tests:
	// equal { correlation vector [ 1 2 3 4 5 ] vector [ 1 2 3 4 5 ] } 1.0
	// equal { correlation vector [ 1 2 3 4 5 ] vector [ 5 4 3 2 1 ] } -1.0
	// Args:
	// * vector1: first vector object
	// * vector2: second vector object
	// Returns:
	// * decimal representing the correlation coefficient between the two vectors
	"correlation": {
		Argsn: 2,
		Doc:   "Calculates the correlation coefficient between two vectors.",
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

	// Tests:
	// equal { dot-product vector [ 1 2 3 ] vector [ 4 5 6 ] } 32.0
	// Args:
	// * vector1: first vector object
	// * vector2: second vector object
	// Returns:
	// * decimal representing the dot product of the two vectors
	"dot-product": {
		Argsn: 2,
		Doc:   "Calculates the dot product between two vectors.",
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
