//go:build !no_vector
// +build !no_vector

package env

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/drewlanenga/govector"
)

//
// VECTOR
//

// Vector -- feature vector (uses govector)
type Vector struct {
	Value govector.Vector
	Kind  Word
}

func NewVector(vec govector.Vector) *Vector {
	return &Vector{vec, Word{0, false}}
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

func NewVectorFromSeries(block TSeries) *Vector {
	data := ArrayFloat32FromSeries(block)
	vec, err := govector.AsVector(data)
	if err != nil {
		return nil
	}
	return &Vector{vec, Word{0, false}}
}

func (i Vector) Type() Type {
	return VectorType
}

func (i Vector) Inspect(idxs Idxs) string {
	var bu strings.Builder
	bu.WriteString("[Vector:") //(" + i.Kind.Print(idxs) + "):")
	bu.WriteString(" Len " + strconv.Itoa(i.Value.Len()))
	bu.WriteString(" Norm " + fmt.Sprintf("%.2f", govector.Norm(i.Value, 2.0)))
	bu.WriteString(" Mean " + fmt.Sprintf("%.2f", i.Value.Mean()))
	bu.WriteString("]")
	return bu.String()
}

func (i Vector) Print(idxs Idxs) string {
	var bu strings.Builder
	bu.WriteString("V[")
	bu.WriteString("Len " + strconv.Itoa(i.Value.Len()))
	bu.WriteString(" Norm " + fmt.Sprintf("%.2f", govector.Norm(i.Value, 2.0)))
	bu.WriteString(" Mean " + fmt.Sprintf("%.2f", i.Value.Mean()))
	bu.WriteString("]")
	return bu.String()
}

func (i Vector) Trace(msg string) {
	fmt.Print(msg + "(Vector): ")
}

func (i Vector) GetKind() int {
	return int(VectorType)
}

func (i Vector) Equal(o Object) bool {
	if i.Type() != o.Type() {
		return false
	}
	oVector := o.(Vector)
	if !i.Kind.Equal(oVector.Kind) {
		return false
	}
	if i.Value.Len() != oVector.Value.Len() {
		return false
	}
	for j := 0; j < i.Value.Len(); j++ {
		if i.Value[j] != oVector.Value[j] {
			return false
		}
	}
	return true
}

func (i Vector) Dump(e Idxs) string {
	var b strings.Builder
	b.WriteString("vector { ")
	for _, v := range i.Value {
		b.WriteString(fmt.Sprintf("%f ", v))
	}
	b.WriteString("}")
	return b.String()
}
