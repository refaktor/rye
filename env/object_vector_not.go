//go:build no_vector
// +build no_vector

package env

// Vector is a minimal stub for when the no_vector build tag is active.
// Without govector, vector operations are unavailable but the type still
// satisfies all interfaces used in builtins_base_*.go.
type Vector struct {
	Value vectorSlice
	Kind  Word
}

// vectorSlice wraps []float64 and provides the govector-compatible methods
// (.Len, .Mean, .Sum) that builtins_base_collections.go calls.
type vectorSlice []float64

func (v vectorSlice) Len() int {
	return len(v)
}

func (v vectorSlice) Mean() float64 {
	if len(v) == 0 {
		return 0
	}
	var sum float64
	for _, x := range v {
		sum += x
	}
	return sum / float64(len(v))
}

func (v vectorSlice) Sum() float64 {
	var sum float64
	for _, x := range v {
		sum += x
	}
	return sum
}

// NewVector creates a stub Vector from a []float64 slice.
func NewVector(vec []float64) *Vector {
	return &Vector{Value: vectorSlice(vec), Kind: Word{0, false}}
}

func ArrayFloat32FromSeries(block TSeries) []float32 {
	data := make([]float32, block.Len())
	for block.Pos() < block.Len() {
		i := block.Pos()
		k1 := block.Pop()
		switch k := k1.(type) {
		case Integer:
			data[i] = float32(k.Value)
		case Decimal:
			data[i] = float32(k.Value)
		}
	}
	return data
}

func NewVectorFromSeries(_ TSeries) *Vector { return nil }

func (i Vector) Type() Type             { return VectorType }
func (i Vector) Inspect(_ Idxs) string  { return "[Vector: unavailable]" }
func (i Vector) Print(_ Idxs) string    { return "V[unavailable]" }
func (i Vector) Trace(msg string)       {}
func (i Vector) GetKind() int           { return int(VectorType) }
func (i Vector) Equal(o Object) bool    { return false }
func (i Vector) Dump(_ Idxs) string     { return "vector { }" }
