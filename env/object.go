// object.go
package env

import (
	"fmt"
	"strconv"
)

type Type int

const (
	BlockType    Type = 1
	IntegerType  Type = 2
	WordType     Type = 3
	SetwordType  Type = 4
	OpwordType   Type = 5
	PipewordType Type = 6
	BuiltinType  Type = 7
	FunctionType Type = 8
	ErrorType    Type = 9
	CommaType    Type = 10
	VoidType     Type = 11
	StringType   Type = 12
)

type Object interface {
	Type() Type
	Inspect(e Idxs) string
	Probe(e Idxs) string
	Trace(msg string)
}

//
// INTEGER
//

// Integer represents an integer.
type Integer struct {
	Value int64
}

// Type returns the type of the Integer.
func (i Integer) Type() Type {
	return IntegerType
}

// Inspect returns a string representation of the Integer.
func (i Integer) Inspect(e Idxs) string {
	return "<Integer: " + strconv.FormatInt(i.Value, 10) + ">"
}

// Inspect returns a string representation of the Integer.
func (i Integer) Probe(e Idxs) string {
	return strconv.FormatInt(i.Value, 10)
}

func (i Integer) Trace(msg string) {
	fmt.Print(msg + "(integer): ")
	fmt.Println(i.Value)
}

//
// STRING
//

// String represents an string.
type String struct {
	Value string
}

// Type returns the type of the Integer.
func (i String) Type() Type {
	return StringType
}

// Inspect returns a string representation of the Integer.
func (i String) Inspect(e Idxs) string {
	return "<String: " + i.Value + ">"
}

// Inspect returns a string representation of the Integer.
func (i String) Probe(e Idxs) string {
	return i.Value
}

func (i String) Trace(msg string) {
	fmt.Print(msg + "(string): ")
	fmt.Println(i.Value)
}

//
// BLOCK
//

// Integer represents an integer.
type Block struct {
	Series TSeries
}

func NewBlock(series TSeries) *Block {
	o := Block{series}
	return &o
}

// Type returns the type of the Integer.
func (i Block) Type() Type {
	return BlockType
}

// Inspect returns a string representation of the Integer.
func (i Block) Inspect(e Idxs) string {
	return "<Block: " + i.Probe(e) + ">"
}

// Inspect returns a string representation of the Integer.
func (b Block) Probe(e Idxs) string {
	res := ""
	for i := 0; i < b.Series.Len(); i += 1 {
		if b.Series.Get(i) != nil {
			res += b.Series.Get(i).Inspect(e)
		}
	}
	return res
}

func (i Block) Trace(msg string) {
	fmt.Print(msg + " (block): ")
	fmt.Println(i.Series)
}

//
// WORD
//

// Integer represents an integer.
type Word struct {
	Index int
}

// Type returns the type of the Integer.
func (i Word) Type() Type {
	return WordType
}

// Inspect returns a string
func (i Word) Inspect(e Idxs) string {
	return "<Word: " + strconv.FormatInt(int64(i.Index), 10) + ", " + e.GetWord(i.Index) + ">"
}

// Inspect returns a string representation of the Integer.
func (b Word) Probe(e Idxs) string {
	return e.GetWord(b.Index)
}

func (i Word) Trace(msg string) {
	fmt.Print(msg + " (word): ")
	fmt.Println(i.Index)
}

//
// SETWORD
//

// Integer represents an integer.
type Setword struct {
	Index int
}

// Type returns the type of the Integer.
func (i Setword) Type() Type {
	return SetwordType
}

// Inspect returns a string representation of the Integer.
func (i Setword) Inspect(e Idxs) string {
	return "<Setword: " + strconv.FormatInt(int64(i.Index), 10) + ", " + e.GetWord(i.Index) + ">"
}

// Inspect returns a string representation of the Integer.
func (b Setword) Probe(e Idxs) string {
	return e.GetWord(b.Index)
}

func (i Setword) Trace(msg string) {
	fmt.Print(msg + "(setword): ")
	fmt.Println(i.Index)
}

//
// OPWORD
//

// Integer represents an integer.
type Opword struct {
	Index int
}

// Type returns the type of the Integer.
func (i Opword) Type() Type {
	return OpwordType
}

// Inspect returns a string
func (i Opword) Inspect(e Idxs) string {
	return "<Opword: " + strconv.FormatInt(int64(i.Index), 10) + ", " + e.GetWord(i.Index) + ">"
}

// Inspect returns a string representation of the Integer.
func (b Opword) Probe(e Idxs) string {
	return e.GetWord(b.Index)
}

func (i Opword) Trace(msg string) {
	fmt.Print(msg + " (opword): ")
	fmt.Println(i.Index)
}

func (i Opword) ToWord() Word {
	return Word{i.Index}
}

//
// PIPEWORD
//

// Integer represents an integer.
type Pipeword struct {
	Index int
}

// Type returns the type of the Integer.
func (i Pipeword) Type() Type {
	return PipewordType
}

// Inspect returns a string
func (i Pipeword) Inspect(e Idxs) string {
	return "<Pipeword: " + strconv.FormatInt(int64(i.Index), 10) + ", " + e.GetWord(i.Index) + ">"
}

// Inspect returns a string representation of the Integer.
func (b Pipeword) Probe(e Idxs) string {
	return e.GetWord(b.Index)
}

func (i Pipeword) Trace(msg string) {
	fmt.Print(msg + " (pipeword): ")
	fmt.Println(i.Index)
}

func (i Pipeword) ToWord() Word {
	return Word{i.Index}
}

//
// COMMA
//

// Integer represents an integer.
type Comma struct{}

// Type returns the type of the Integer.
func (i Comma) Type() Type {
	return CommaType
}

// Inspect returns a string
func (i Comma) Inspect(e Idxs) string {
	return "<Comma>"
}

// Inspect returns a string representation of the Integer.
func (b Comma) Probe(e Idxs) string {
	return ","
}

func (i Comma) Trace(msg string) {
	fmt.Print(msg + " (comma)")
}

//
// COMMA
//

// Integer represents an integer.
type Void struct{}

// Type returns the type of the Integer.
func (i Void) Type() Type {
	return VoidType
}

// Inspect returns a string
func (i Void) Inspect(e Idxs) string {
	return "<Void>"
}

// Inspect returns a string representation of the Integer.
func (b Void) Probe(e Idxs) string {
	return "_"
}

func (i Void) Trace(msg string) {
	fmt.Print(msg + " (void)")
}

//
// Function
//

// Integer represents an integer.
type Function struct {
	Argsn int
	Spec  Block
	Body  Block
}

func NewFunction(spec Block, body Block) *Function {
	o := Function{spec.Series.Len(), spec, body}
	return &o
}

// Type returns the type of the Integer.
func (i Function) Type() Type {
	return FunctionType
}

// Inspect returns a string representation of the Integer.
func (i Function) Inspect(e Idxs) string {
	return "<Function: " + i.Spec.Inspect(e) + ", " + i.Body.Inspect(e) + ">"
}

// Inspect returns a string representation of the Integer.
func (b Function) Probe(e Idxs) string {
	return "<Function: " + strconv.FormatInt(int64(b.Argsn), 10) + ">"
}

func (i Function) Trace(msg string) {
	fmt.Print(msg + " (function): ")
	fmt.Println(i.Spec)
}

//
// BuiltinFunction
//

// BuiltinFunction represents a function signature of builtin functions.
/////type BuiltinFunction func(ps *ProgramState, args ...Object) Object
type BuiltinFunction func(ps *ProgramState, arg0 Object, arg1 Object, arg2 Object, arg3 Object, arg4 Object) Object

// Builtin represents a builtin function.
type Builtin struct {
	Fn    BuiltinFunction
	Argsn int
	Cur0  Object
	Cur1  Object
	Cur2  Object
	Cur3  Object
	Cur4  Object
}

func NewBuiltin(fn BuiltinFunction, argsn int) *Builtin {
	bl := Builtin{fn, argsn, nil, nil, nil, nil, nil}
	return &bl
}

// Type returns the type of the Builtin.
func (b Builtin) Type() Type {
	return BuiltinType
}

// Inspect returns a string representation of the Builtin.
func (b Builtin) Inspect(e Idxs) string {
	return "<Builtin>"
}

func (b Builtin) Probe(e Idxs) string {
	return "<Builtin>"
}

func (i Builtin) Trace(msg string) {
	fmt.Print(msg + " (builtin): ")
	fmt.Println(i.Argsn)
}

//
// Error
//

// Integer represents an integer.
type Error struct {
	message string
}

// Type returns the type of the Integer.
func (i Error) Type() Type {
	return ErrorType
}

// Inspect returns a string representation of the Integer.
func (i Error) Inspect(e Idxs) string {
	return "<Error: " + i.message + ">"
}

// Inspect returns a string representation of the Integer.
func (b Error) Probe(e Idxs) string {
	return "<Error: " + b.message + ">"
}

func NewError(message string) *Error {
	var e Error
	fmt.Println("ERROR: " + message)
	e.message = message
	return &e
}

func (i Error) Trace(msg string) {
	fmt.Print(msg + "(error): ")
	fmt.Println(i.message)
}
