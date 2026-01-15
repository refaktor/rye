package evaldo

import (
	"math"

	"github.com/refaktor/rye/env"

	"github.com/drewlanenga/govector"
)

func sqrtFloat64(x float64) float64 {
	return math.Sqrt(x)
}

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

	//
	// ##### Distance Functions #####
	//

	// Tests:
	// equal { euclidean-distance vector [ 0 0 ] vector [ 3 4 ] } 5.0
	// equal { euclidean-distance vector [ 1 2 3 ] vector [ 1 2 3 ] } 0.0
	// Args:
	// * vector1: first vector object
	// * vector2: second vector object
	// Returns:
	// * decimal representing the Euclidean distance between the two vectors
	"euclidean-distance": {
		Argsn: 2,
		Doc:   "Calculates the Euclidean distance between two vectors. Useful for K-Means clustering.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch v1 := arg0.(type) {
			case env.Vector:
				switch v2 := arg1.(type) {
				case env.Vector:
					if len(v1.Value) != len(v2.Value) {
						return MakeBuiltinError(ps, "Vectors must have the same length", "euclidean-distance")
					}
					var sum float64
					for i := 0; i < len(v1.Value); i++ {
						diff := v1.Value[i] - v2.Value[i]
						sum += diff * diff
					}
					return *env.NewDecimal(sqrtFloat64(sum))
				default:
					return MakeArgError(ps, 2, []env.Type{env.VectorType}, "euclidean-distance")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.VectorType}, "euclidean-distance")
			}
		},
	},

	//
	// ##### Element-wise Arithmetic #####
	//

	// Tests:
	// equal { vector [ 1 2 3 ] .add vector [ 4 5 6 ] |to-block } { 5.0 7.0 9.0 }
	// Args:
	// * vector1: first vector object
	// * vector2: second vector object to add
	// Returns:
	// * new vector with element-wise sum
	"Vector//add": {
		Argsn: 2,
		Doc:   "Adds two vectors element-wise. Returns a new vector.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch v1 := arg0.(type) {
			case env.Vector:
				switch v2 := arg1.(type) {
				case env.Vector:
					if len(v1.Value) != len(v2.Value) {
						return MakeBuiltinError(ps, "Vectors must have the same length", "Vector//add")
					}
					result := make(govector.Vector, len(v1.Value))
					for i := 0; i < len(v1.Value); i++ {
						result[i] = v1.Value[i] + v2.Value[i]
					}
					return *env.NewVector(result)
				default:
					return MakeArgError(ps, 2, []env.Type{env.VectorType}, "Vector//add")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.VectorType}, "Vector//add")
			}
		},
	},

	// Tests:
	// equal { vector [ 5 7 9 ] .sub vector [ 4 5 6 ] |to-block } { 1.0 2.0 3.0 }
	// Args:
	// * vector1: first vector object
	// * vector2: second vector object to subtract
	// Returns:
	// * new vector with element-wise difference
	"Vector//sub": {
		Argsn: 2,
		Doc:   "Subtracts two vectors element-wise. Returns a new vector.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch v1 := arg0.(type) {
			case env.Vector:
				switch v2 := arg1.(type) {
				case env.Vector:
					if len(v1.Value) != len(v2.Value) {
						return MakeBuiltinError(ps, "Vectors must have the same length", "Vector//sub")
					}
					result := make(govector.Vector, len(v1.Value))
					for i := 0; i < len(v1.Value); i++ {
						result[i] = v1.Value[i] - v2.Value[i]
					}
					return *env.NewVector(result)
				default:
					return MakeArgError(ps, 2, []env.Type{env.VectorType}, "Vector//sub")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.VectorType}, "Vector//sub")
			}
		},
	},

	// Tests:
	// equal { vector [ 1 2 3 ] .scale 2.0 |to-block } { 2.0 4.0 6.0 }
	// equal { vector [ 10 20 ] .scale 0.5 |to-block } { 5.0 10.0 }
	// Args:
	// * vector: vector object
	// * scalar: decimal or integer to multiply by
	// Returns:
	// * new vector with each element multiplied by the scalar
	"Vector//scale": {
		Argsn: 2,
		Doc:   "Multiplies a vector by a scalar. Returns a new vector.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch v := arg0.(type) {
			case env.Vector:
				var scalar float64
				switch s := arg1.(type) {
				case env.Decimal:
					scalar = s.Value
				case env.Integer:
					scalar = float64(s.Value)
				default:
					return MakeArgError(ps, 2, []env.Type{env.DecimalType, env.IntegerType}, "Vector//scale")
				}
				result := make(govector.Vector, len(v.Value))
				for i := 0; i < len(v.Value); i++ {
					result[i] = v.Value[i] * scalar
				}
				return *env.NewVector(result)
			default:
				return MakeArgError(ps, 1, []env.Type{env.VectorType}, "Vector//scale")
			}
		},
	},

	// Tests:
	// equal { mean-vectors { vector [ 1 2 ] vector [ 3 4 ] } |to-block } { 2.0 3.0 }
	// Args:
	// * block: block of vectors to average
	// Returns:
	// * new vector representing the element-wise mean of all input vectors
	"mean-vectors": {
		Argsn: 1,
		Doc:   "Calculates the element-wise mean of multiple vectors. Useful for creating 'Master Anchors'.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch block := arg0.(type) {
			case env.Block:
				if block.Series.Len() == 0 {
					return MakeBuiltinError(ps, "Block must contain at least one vector", "mean-vectors")
				}
				// Get first vector to determine length
				first := block.Series.Get(0)
				firstVec, ok := first.(env.Vector)
				if !ok {
					return MakeBuiltinError(ps, "Block must contain only vectors", "mean-vectors")
				}
				vecLen := len(firstVec.Value)
				result := make(govector.Vector, vecLen)
				// Sum all vectors
				for i := 0; i < block.Series.Len(); i++ {
					item := block.Series.Get(i)
					vec, ok := item.(env.Vector)
					if !ok {
						return MakeBuiltinError(ps, "Block must contain only vectors", "mean-vectors")
					}
					if len(vec.Value) != vecLen {
						return MakeBuiltinError(ps, "All vectors must have the same length", "mean-vectors")
					}
					for j := 0; j < vecLen; j++ {
						result[j] += vec.Value[j]
					}
				}
				// Divide by count
				count := float64(block.Series.Len())
				for i := 0; i < vecLen; i++ {
					result[i] /= count
				}
				return *env.NewVector(result)
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType}, "mean-vectors")
			}
		},
	},

	//
	// ##### Geometric & Transformation #####
	//

	// Tests:
	// equal { vector [ 3 4 ] .unit |normalize } 1.0
	// Args:
	// * vector: vector object to normalize
	// Returns:
	// * new vector with length 1 (unit vector)
	"Vector//unit": {
		Argsn: 1,
		Doc:   "Returns a new normalized vector (length = 1). Essential for consistent dot products and overlays.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch v := arg0.(type) {
			case env.Vector:
				norm := govector.Norm(v.Value, 2.0)
				if norm == 0 {
					return MakeBuiltinError(ps, "Cannot normalize zero vector", "Vector//unit")
				}
				result := make(govector.Vector, len(v.Value))
				for i := 0; i < len(v.Value); i++ {
					result[i] = v.Value[i] / norm
				}
				return *env.NewVector(result)
			default:
				return MakeArgError(ps, 1, []env.Type{env.VectorType}, "Vector//unit")
			}
		},
	},

	// Tests:
	// equal { vector [ 3 4 ] .project vector [ 1 0 ] |to-block } { 3.0 0.0 }
	// Args:
	// * vector1: vector to project
	// * vector2: vector to project onto
	// Returns:
	// * new vector representing the projection of vector1 onto vector2
	"Vector//project": {
		Argsn: 2,
		Doc:   "Projects vector onto another vector. Answers: 'How much of this vector is in the direction of another?'",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch v1 := arg0.(type) {
			case env.Vector:
				switch v2 := arg1.(type) {
				case env.Vector:
					if len(v1.Value) != len(v2.Value) {
						return MakeBuiltinError(ps, "Vectors must have the same length", "Vector//project")
					}
					// project(a, b) = (a·b / b·b) * b
					dotAB, err := govector.DotProduct(v1.Value, v2.Value)
					if err != nil {
						return MakeBuiltinError(ps, err.Error(), "Vector//project")
					}
					dotBB, err := govector.DotProduct(v2.Value, v2.Value)
					if err != nil {
						return MakeBuiltinError(ps, err.Error(), "Vector//project")
					}
					if dotBB == 0 {
						return MakeBuiltinError(ps, "Cannot project onto zero vector", "Vector//project")
					}
					scalar := dotAB / dotBB
					result := make(govector.Vector, len(v2.Value))
					for i := 0; i < len(v2.Value); i++ {
						result[i] = v2.Value[i] * scalar
					}
					return *env.NewVector(result)
				default:
					return MakeArgError(ps, 2, []env.Type{env.VectorType}, "Vector//project")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.VectorType}, "Vector//project")
			}
		},
	},

	// Tests:
	// equal { vector [ 3 4 ] .reject vector [ 1 0 ] |to-block } { 0.0 4.0 }
	// Args:
	// * vector1: vector to reject from
	// * vector2: vector representing the direction to remove
	// Returns:
	// * new vector with the projection onto vector2 removed
	"Vector//reject": {
		Argsn: 2,
		Doc:   "Removes projection from vector. Use this to 'subtract' a concept direction from a vector.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch v1 := arg0.(type) {
			case env.Vector:
				switch v2 := arg1.(type) {
				case env.Vector:
					if len(v1.Value) != len(v2.Value) {
						return MakeBuiltinError(ps, "Vectors must have the same length", "Vector//reject")
					}
					// reject(a, b) = a - project(a, b)
					// project(a, b) = (a·b / b·b) * b
					dotAB, err := govector.DotProduct(v1.Value, v2.Value)
					if err != nil {
						return MakeBuiltinError(ps, err.Error(), "Vector//reject")
					}
					dotBB, err := govector.DotProduct(v2.Value, v2.Value)
					if err != nil {
						return MakeBuiltinError(ps, err.Error(), "Vector//reject")
					}
					if dotBB == 0 {
						// If b is zero vector, rejection is just the original vector
						return v1
					}
					scalar := dotAB / dotBB
					result := make(govector.Vector, len(v1.Value))
					for i := 0; i < len(v1.Value); i++ {
						result[i] = v1.Value[i] - (v2.Value[i] * scalar)
					}
					return *env.NewVector(result)
				default:
					return MakeArgError(ps, 2, []env.Type{env.VectorType}, "Vector//reject")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.VectorType}, "Vector//reject")
			}
		},
	},

	// Tests:
	// equal { vector [ 1 2 3 ] .to-block } { 1.0 2.0 3.0 }
	// Args:
	// * vector: vector object
	// Returns:
	// * block containing the vector elements as decimals
	"Vector//to-block": {
		Argsn: 1,
		Doc:   "Converts a vector to a block of decimals.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch v := arg0.(type) {
			case env.Vector:
				items := make([]env.Object, len(v.Value))
				for i, val := range v.Value {
					items[i] = *env.NewDecimal(val)
				}
				return *env.NewBlock(*env.NewTSeries(items))
			default:
				return MakeArgError(ps, 1, []env.Type{env.VectorType}, "Vector//to-block")
			}
		},
	},
}
