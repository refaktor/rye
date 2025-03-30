// series.go
package env

import (
	"strings"
)

// TSeries represents a series of Objects with a position pointer.
// It provides methods for traversing, accessing, and modifying the series.
type TSeries struct {
	S   []Object `bson:"series"` // The underlying slice of Objects
	pos int      // Current position in the series
}

// NewTSeries creates a new TSeries with the given slice of Objects.
// The position pointer is initialized to 0 (beginning of the series).
func NewTSeries(ser []Object) *TSeries {
	ser1 := TSeries{ser, 0}
	return &ser1
}

// Ended returns true if the position is beyond the end of the series.
func (ser TSeries) Ended() bool {
	return ser.pos > len(ser.S)
}

// AtLast returns true if the position is at or beyond the last element of the series.
func (ser TSeries) AtLast() bool {
	return ser.pos > len(ser.S)-1
}

// Pos returns the current position in the series.
func (ser TSeries) Pos() int {
	return ser.pos
}

// Next advances the position pointer by one.
func (ser *TSeries) Next() {
	ser.pos++
}

// Pop returns the Object at the current position and advances the position.
// Returns nil if the position is out of bounds.
func (ser *TSeries) Pop() Object {
	if ser.pos >= len(ser.S) {
		return nil
	}
	obj := ser.S[ser.pos]
	ser.pos++
	return obj
}

// RmLast removes the last element from the series and returns the modified series.
// If the series is empty, returns the series unchanged.
func (ser *TSeries) RmLast() *TSeries {
	if len(ser.S) > 0 {
		ser.S = ser.S[:len(ser.S)-1]
		return ser
	} else {
		return ser
	}
}

// Put replaces the Object at the position before the current position (pos-1) with the given Object.
// Returns true if successful, false if the position is out of bounds.
// Note: This is typically used after Pop() to replace the item that was just popped.
func (ser *TSeries) Put(obj Object) bool {
	if ser.pos > 0 && ser.pos <= len(ser.S) {
		ser.S[ser.pos-1] = obj // -1 ... because we already popped out the word
		return true
	}
	// Return false for out-of-bounds case
	return false
}

// Append adds a single Object to the end of the series and returns the modified series.
func (ser *TSeries) Append(obj Object) *TSeries {
	ser.S = append(ser.S, obj)
	return ser
}

// AppendMul adds multiple Objects to the end of the series and returns the modified series.
func (ser *TSeries) AppendMul(objs []Object) *TSeries {
	ser.S = append(ser.S, objs...)
	return ser
}

// Reset sets the position pointer back to the beginning of the series.
func (ser *TSeries) Reset() {
	ser.pos = 0
}

// SetPos sets the position pointer to the specified position.
// Note: This does not perform bounds checking.
func (ser *TSeries) SetPos(pos int) {
	ser.pos = pos
}

// GetPos returns the current position in the series.
// This is an alias for Pos() but as a pointer receiver method.
func (ser *TSeries) GetPos() int {
	return ser.pos
}

// GetAll returns the underlying slice of Objects.
// Note: This returns a direct reference to the internal slice, not a copy.
func (ser *TSeries) GetAll() []Object {
	return ser.S
}

// Peek returns the Object at the current position without advancing the position.
// Returns nil if the position is out of bounds.
func (ser TSeries) Peek() Object {
	if len(ser.S) > ser.pos {
		return ser.S[ser.pos]
	}
	return nil
}

// Get returns the Object at the specified position n.
// Returns nil if the position is out of bounds.
func (ser TSeries) Get(n int) Object {
	if n >= 0 && n < len(ser.S) {
		return ser.S[n]
	}
	return nil
}

// PGet returns a pointer to the Object at position n in the series.
// WARNING: This returns a direct pointer to the internal object, which allows
// modification of the original object. Use with caution to avoid unintended
// side effects. For read-only access, prefer using Get() instead.
func (ser TSeries) PGet(n int) *Object {
	if n >= 0 && n < len(ser.S) {
		return &ser.S[n]
	}
	return nil
}

// Len returns the length of the series (number of Objects).
func (ser TSeries) Len() int {
	return len(ser.S)
}

// PositionAndSurroundingElements returns a string of the position of the series, marked with (here) and 10 surrounding elements.
func (ser TSeries) PositionAndSurroundingElements(idxs Idxs) string {
	var bu strings.Builder
	bu.WriteString("  ")
	st := 0
	if ser.Pos() > 10 {
		bu.WriteString("... ")
		st = ser.Pos() - 11
	}
	for i := st; i < ser.Pos()+9 && i < ser.Len(); i++ {
		if i == ser.Pos()-1 {
			bu.WriteString("\x1b[1m(here) \x1b[22m")
		}

		v := ser.S[i]
		if v != nil {
			bu.WriteString(v.Print(idxs) + " ")
		} else {
			bu.WriteString("<<< NIL >>>" + " ")
		}
	}
	if ser.Len() == ser.Pos()-1 {
		bu.WriteString("\x1b[1m(here)\x1b[22m")
	}
	if ser.Len() > ser.Pos()+9 {
		bu.WriteString("... ")
	}
	bu.WriteString("")
	return bu.String()
}
