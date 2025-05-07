package env

import (
	"fmt"
	"strconv"
)

// CurriedCaller represents a curried function or builtin.
// This type is separate from Builtin and Function to avoid slowing down regular execution.
type CurriedCaller struct {
	CallerType int       // 0 for Builtin, 1 for Function
	Builtin    *Builtin  // Non-nil if CallerType == 0
	Function   *Function // Non-nil if CallerType == 1
	Cur0       Object    // Curried arguments
	Cur1       Object
	Cur2       Object
	Cur3       Object
	Cur4       Object
	Argsn      int    // Total number of arguments expected
	Pure       bool   // Whether the caller is pure
	Doc        string // Documentation
}

// NewCurriedCallerFromBuiltin creates a new CurriedCaller from a Builtin
func NewCurriedCallerFromBuiltin(bi Builtin, cur0, cur1, cur2, cur3, cur4 Object) *CurriedCaller {
	return &CurriedCaller{
		CallerType: 0,
		Builtin:    &bi,
		Function:   nil,
		Cur0:       cur0,
		Cur1:       cur1,
		Cur2:       cur2,
		Cur3:       cur3,
		Cur4:       cur4,
		Argsn:      bi.Argsn,
		Pure:       bi.Pure,
		Doc:        bi.Doc,
	}
}

// NewCurriedCallerFromFunction creates a new CurriedCaller from a Function
func NewCurriedCallerFromFunction(fn Function, cur0, cur1, cur2, cur3, cur4 Object) *CurriedCaller {
	return &CurriedCaller{
		CallerType: 1,
		Builtin:    nil,
		Function:   &fn,
		Cur0:       cur0,
		Cur1:       cur1,
		Cur2:       cur2,
		Cur3:       cur3,
		Cur4:       cur4,
		Argsn:      fn.Argsn,
		Pure:       fn.Pure,
		Doc:        fn.Doc,
	}
}

// Type returns the type of the object
func (c CurriedCaller) Type() Type {
	return CurriedCallerType
}

// Inspect returns a string representation of the object for debugging
func (c CurriedCaller) Inspect(e Idxs) string {
	return "[" + c.Print(e) + "]"
}

// Print returns a string representation of the object
func (c CurriedCaller) Print(e Idxs) string {
	var pure string
	if c.Pure {
		pure = "Pure "
	}

	var callerType string
	if c.CallerType == 0 {
		callerType = "Builtin"
	} else {
		callerType = "Function"
	}

	return pure + "CurriedCaller(" + callerType + ", " + strconv.Itoa(c.Argsn) + "): " + c.Doc
}

// Trace prints a trace message
func (c CurriedCaller) Trace(msg string) {
	fmt.Print(msg + " (curriedcaller): ")
	fmt.Println(c.Argsn)
}

// GetKind returns the kind of the object
func (c CurriedCaller) GetKind() int {
	return int(CurriedCallerType)
}

// Equal checks if two objects are equal
func (c CurriedCaller) Equal(o Object) bool {
	if c.Type() != o.Type() {
		return false
	}

	oCurriedCaller := o.(CurriedCaller)
	if c.CallerType != oCurriedCaller.CallerType {
		return false
	}

	if c.Argsn != oCurriedCaller.Argsn {
		return false
	}

	if c.Pure != oCurriedCaller.Pure {
		return false
	}

	// Compare curried arguments
	if !objectsEqual(c.Cur0, oCurriedCaller.Cur0) {
		return false
	}
	if !objectsEqual(c.Cur1, oCurriedCaller.Cur1) {
		return false
	}
	if !objectsEqual(c.Cur2, oCurriedCaller.Cur2) {
		return false
	}
	if !objectsEqual(c.Cur3, oCurriedCaller.Cur3) {
		return false
	}
	if !objectsEqual(c.Cur4, oCurriedCaller.Cur4) {
		return false
	}

	return true
}

// Helper function to compare two objects that might be nil
func objectsEqual(a, b Object) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.Equal(b)
}

// Dump returns a string representation of the object for serialization
func (c CurriedCaller) Dump(e Idxs) string {
	// Serializing curried callers is not supported
	return ""
}
