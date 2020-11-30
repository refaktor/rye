// object.go
package env

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Type int

const (
	BlockType       Type = 1
	IntegerType     Type = 2
	WordType        Type = 3
	SetwordType     Type = 4
	OpwordType      Type = 5
	PipewordType    Type = 6
	BuiltinType     Type = 7
	FunctionType    Type = 8
	ErrorType       Type = 9
	CommaType       Type = 10
	VoidType        Type = 11
	StringType      Type = 12
	TagwordType     Type = 13
	GenwordType     Type = 14
	GetwordType     Type = 15
	ArgwordType     Type = 16
	NativeType      Type = 17
	UriType         Type = 18
	LSetwordType    Type = 19
	CtxType         Type = 20
	DictType        Type = 21
	ListType        Type = 22
	DateType        Type = 23
	CPathType       Type = 24
	XwordType       Type = 25
	EXwordType      Type = 26
	SpreadsheetType Type = 27
	EmailType       Type = 28
	KindType        Type = 29
	KindwordType    Type = 30
	ConverterType   Type = 31
	TimeType        Type = 32
)

// after adding new type here, also add string to idxs.go

type Object interface {
	Type() Type
	Inspect(e Idxs) string
	Probe(e Idxs) string
	Trace(msg string)
	GetKind() int
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

func (i Integer) GetKind() int {
	return int(IntegerType)
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
	return "\"" + i.Value + "\""
}

func (i String) Trace(msg string) {
	fmt.Print(msg + "(string): ")
	fmt.Println(i.Value)
}

func (i String) GetKind() int {
	return int(StringType)
}

//
// DATE
//

type Date struct {
	Value time.Time
}

func (i Date) Type() Type {
	return DateType
}

func (i Date) Inspect(e Idxs) string {
	return "<Date: " + i.Value.Format(time.RFC822Z) + ">"
}

func (i Date) Probe(e Idxs) string {
	return i.Value.Format(time.RFC822Z)
}

func (i Date) Trace(msg string) {
	fmt.Print(msg + "(date): ")
	fmt.Println(i.Value.Format(time.RFC822Z))
}

func (i Date) GetKind() int {
	return int(DateType)
}

//
// URI
//

type Uri struct {
	Scheme Word
	Path   string
	Kind   Word
}

func NewUri(index *Idxs, scheme Word, path string) *Uri {
	kindstr := strings.Split(path, "://")[0] + "-schema" // TODO -- this is just temporary .. so we test it further, make propper once at that level
	idx := index.IndexWord(kindstr)
	nat := Uri{scheme, path, Word{idx}}
	return &nat
}

func (i Uri) GetPath() string {
	return strings.SplitAfter(i.Path, "://")[1]
}

func (i Uri) GetProtocol() string {
	return strings.Split(i.Path, "://")[0]
}

func (i Uri) Type() Type {
	return UriType
}

// Inspect returns a string representation of the Integer.
func (i Uri) Inspect(e Idxs) string {
	return "<Uri: " + i.Scheme.Probe(e) + "://" + i.Path + ">"
}

// Inspect returns a string representation of the Integer.
func (i Uri) Probe(e Idxs) string {
	return i.Path
}

func (i Uri) Trace(msg string) {
	fmt.Print(msg + "(uri): ")
	fmt.Println(i.Path)
}

func (i Uri) GetKind() int {
	return i.Kind.Index
}

//
// Email
//

type Email struct {
	Address string
}

func (i Email) Type() Type {
	return EmailType
}

// Inspect returns a string representation of the Integer.
func (i Email) Inspect(e Idxs) string {
	return "<Email: " + i.Probe(e) + ">"
}

// Inspect returns a string representation of the Integer.
func (i Email) Probe(e Idxs) string {
	return i.Address
}

func (i Email) Trace(msg string) {
	fmt.Print(msg + "(email): ")
	fmt.Println(i.Address)
}

func (i Email) GetKind() int {
	return int(EmailType)
}

//
// BLOCK
//

// Integer represents an integer.
type Block struct {
	Series TSeries
	Mode   int
}

func NewBlock(series TSeries) *Block {
	o := Block{series, 0}
	return &o
}

func NewBlock2(series TSeries, m int) *Block {
	o := Block{series, m}
	return &o
}

// Type returns the type of the Integer.
func (i Block) Type() Type {
	return BlockType
}

// Inspect returns a string representation of the Integer.
func (b Block) Inspect(e Idxs) string {
	var r strings.Builder
	r.WriteString("<Block: ")
	for i := 0; i < b.Series.Len(); i += 1 {
		if b.Series.Get(i) != nil {
			r.WriteString(b.Series.Get(i).Inspect(e))
			r.WriteString(" ")
		}
	}
	r.WriteString(">")
	return r.String()
}

// Inspect returns a string representation of the Integer.
func (b Block) Probe(e Idxs) string {
	var r strings.Builder
	r.WriteString("{ ")
	for i := 0; i < b.Series.Len(); i += 1 {
		if b.Series.Get(i) != nil {
			r.WriteString(b.Series.Get(i).Probe(e))
			r.WriteString(" ")
		}
	}
	r.WriteString(" }")
	return r.String()
}

func (i Block) Trace(msg string) {
	fmt.Print(msg + " (block): ")
	fmt.Println(i.Series)
}

func (i Block) GetKind() int {
	return int(BlockType)
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

func (i Word) GetKind() int {
	return int(WordType)
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
	return e.GetWord(b.Index) + ":"
}

func (i Setword) Trace(msg string) {
	fmt.Print(msg + "(setword): ")
	fmt.Println(i.Index)
}

func (i Setword) GetKind() int {
	return int(SetwordType)
}

//
// LSETWORD
//

// Integer represents an integer.
type LSetword struct {
	Index int
}

// Type returns the type of the Integer.
func (i LSetword) Type() Type {
	return LSetwordType
}

// Inspect returns a string representation of the Integer.
func (i LSetword) Inspect(e Idxs) string {
	return "<LSetword: " + strconv.FormatInt(int64(i.Index), 10) + ", " + e.GetWord(i.Index) + ">"
}

// Inspect returns a string representation of the Integer.
func (b LSetword) Probe(e Idxs) string {
	return ":" + e.GetWord(b.Index)
}

func (i LSetword) Trace(msg string) {
	fmt.Print(msg + "(lsetword): ")
	fmt.Println(i.Index)
}

func (i LSetword) GetKind() int {
	return int(LSetwordType)
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
	return "." + e.GetWord(b.Index)
}

func (i Opword) Trace(msg string) {
	fmt.Print(msg + " (opword): ")
	fmt.Println(i.Index)
}

func (i Opword) ToWord() Word {
	return Word{i.Index}
}

func (i Opword) GetKind() int {
	return int(OpwordType)
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
	return "|" + e.GetWord(b.Index)
}

func (i Pipeword) Trace(msg string) {
	fmt.Print(msg + " (pipeword): ")
	fmt.Println(i.Index)
}

func (i Pipeword) ToWord() Word {
	return Word{i.Index}
}

func (i Pipeword) GetKind() int {
	return int(PipewordType)
}

//
// TAGWORD
//

// Integer represents an integer.
type Tagword struct {
	Index int
}

// Type returns the type of the Integer.
func (i Tagword) Type() Type {
	return TagwordType
}

// Inspect returns a string
func (i Tagword) Inspect(e Idxs) string {
	return "<Tagword: " + strconv.FormatInt(int64(i.Index), 10) + ", " + e.GetWord(i.Index) + ">"
}

// Inspect returns a string representation of the Integer.
func (b Tagword) Probe(e Idxs) string {
	return e.GetWord(b.Index)
}

func (i Tagword) Trace(msg string) {
	fmt.Print(msg + " (tagword): ")
	fmt.Println(i.Index)
}

func (i Tagword) ToWord() Word {
	return Word{i.Index}
}

func (i Tagword) GetKind() int {
	return int(TagwordType)
}

//
// XWORD
//

// Integer represents an integer.
type Xword struct {
	Index int
}

// Type returns the type of the Integer.
func (i Xword) Type() Type {
	return XwordType
}

// Inspect returns a string
func (i Xword) Inspect(e Idxs) string {
	return "<Xword: " + strconv.FormatInt(int64(i.Index), 10) + ", " + e.GetWord(i.Index) + ">"
}

// Inspect returns a string representation of the Integer.
func (b Xword) Probe(e Idxs) string {
	return e.GetWord(b.Index)
}

func (i Xword) Trace(msg string) {
	fmt.Print(msg + " (Xword): ")
	fmt.Println(i.Index)
}

func (i Xword) ToWord() Word {
	return Word{i.Index}
}

func (i Xword) GetKind() int {
	return int(XwordType)
}

//
// EXWORD
//

// Integer represents an integer.
type EXword struct {
	Index int
}

// Type returns the type of the Integer.
func (i EXword) Type() Type {
	return EXwordType
}

// Inspect returns a string
func (i EXword) Inspect(e Idxs) string {
	return "<EXword: " + strconv.FormatInt(int64(i.Index), 10) + ", " + e.GetWord(i.Index) + ">"
}

// Inspect returns a string representation of the Integer.
func (b EXword) Probe(e Idxs) string {
	return e.GetWord(b.Index)
}

func (i EXword) Trace(msg string) {
	fmt.Print(msg + " (EXword): ")
	fmt.Println(i.Index)
}

func (i EXword) ToWord() Word {
	return Word{i.Index}
}

func (i EXword) GetKind() int {
	return int(EXwordType)
}

//
// KINDWORD
//

// Integer represents an integer.
type Kindword struct {
	Index int
}

// Type returns the type of the Integer.
func (i Kindword) Type() Type {
	return KindwordType
}

// Inspect returns a string
func (i Kindword) Inspect(e Idxs) string {
	return "<Kindword: " + strconv.FormatInt(int64(i.Index), 10) + ", " + e.GetWord(i.Index) + ">"
}

// Inspect returns a string representation of the Integer.
func (b Kindword) Probe(e Idxs) string {
	return e.GetWord(b.Index)
}

func (i Kindword) Trace(msg string) {
	fmt.Print(msg + " (Kidnword): ")
	fmt.Println(i.Index)
}

func (i Kindword) ToWord() Word {
	return Word{i.Index}
}

func (i Kindword) GetKind() int {
	return int(KindwordType)
}

//
// GETWORD
//

// Integer represents an integer.
type Getword struct {
	Index int
}

// Type returns the type of the Integer.
func (i Getword) Type() Type {
	return GetwordType
}

// Inspect returns a string
func (i Getword) Inspect(e Idxs) string {
	return "<Getword: " + strconv.FormatInt(int64(i.Index), 10) + ", " + e.GetWord(i.Index) + ">"
}

// Inspect returns a string representation of the Integer.
func (b Getword) Probe(e Idxs) string {
	return e.GetWord(b.Index)
}

func (i Getword) Trace(msg string) {
	fmt.Print(msg + " (getword): ")
	fmt.Println(i.Index)
}

func (i Getword) ToWord() Word {
	return Word{i.Index}
}

func (i Getword) GetKind() int {
	return int(GetwordType)
}

//
// GENWORD
//

// Integer represents an integer.
type Genword struct {
	Index int
}

// Type returns the type of the Integer.
func (i Genword) Type() Type {
	return GenwordType
}

// Inspect returns a string
func (i Genword) Inspect(e Idxs) string {
	return "<Genword: " + strconv.FormatInt(int64(i.Index), 10) + ", " + e.GetWord(i.Index) + ">"
}

// Inspect returns a string representation of the Integer.
func (b Genword) Probe(e Idxs) string {
	return e.GetWord(b.Index)
}

func (i Genword) Trace(msg string) {
	fmt.Print(msg + " (genword): ")
	fmt.Println(i.Index)
}

func (i Genword) ToWord() Word {
	return Word{i.Index}
}

func (i Genword) GetKind() int {
	return int(GenwordType)
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

func (i Comma) GetKind() int {
	return int(CommaType)
}

//
// VOID
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

func (i Void) GetKind() int {
	return int(VoidType)
}

//
// Function
//

// Integer represents an integer.
type Function struct {
	Argsn int
	Spec  Block
	Body  Block
	Ctx   *RyeCtx
	Pure  bool
}

func NewFunction(spec Block, body Block, pure bool) *Function {
	o := Function{spec.Series.Len(), spec, body, nil, pure}
	return &o
}

func NewFunctionC(spec Block, body Block, ctx *RyeCtx, pure bool) *Function {
	o := Function{spec.Series.Len(), spec, body, ctx, pure}
	return &o
}

// Type returns the type of the Integer.
func (i Function) Type() Type {
	return FunctionType
}

// Inspect returns a string representation of the Integer.
func (i Function) Inspect(e Idxs) string {
	// LONG DISPLAY OF FUNCTION NODES return "<Function: " + i.Spec.Inspect(e) + ", " + i.Body.Inspect(e) + ">"
	var pure_s string
	if i.Pure {
		pure_s = "Pure "
	}
	return "<" + pure_s + "Function: " + strconv.FormatInt(int64(i.Argsn), 10) + ">"
}

// Inspect returns a string representation of the Integer.
func (b Function) Probe(e Idxs) string {
	return "<Function: " + strconv.FormatInt(int64(b.Argsn), 10) + ">"
}

func (i Function) Trace(msg string) {
	fmt.Print(msg + " (function): ")
	fmt.Println(i.Spec)
}

func (i Function) GetKind() int {
	return int(FunctionType)
}

//
// BuiltinFunction
//

// BuiltinFunction represents a function signature of builtin functions.
/////type BuiltinFunction func(ps *ProgramState, args ...Object) Object
type BuiltinFunction func(ps *ProgramState, arg0 Object, arg1 Object, arg2 Object, arg3 Object, arg4 Object) Object

// Builtin represents a builtin function.
type Builtin struct {
	Fn            BuiltinFunction
	Argsn         int
	Cur0          Object
	Cur1          Object
	Cur2          Object
	Cur3          Object
	Cur4          Object
	AcceptFailure bool
	Pure          bool
}

func NewBuiltin(fn BuiltinFunction, argsn int, acceptFailure bool, pure bool) *Builtin {
	bl := Builtin{fn, argsn, nil, nil, nil, nil, nil, acceptFailure, pure}
	return &bl
}

// Type returns the type of the Builtin.
func (b Builtin) Type() Type {
	return BuiltinType
}

// Inspect returns a string representation of the Builtin.
func (b Builtin) Inspect(e Idxs) string {
	var pure_s string
	if b.Pure {
		pure_s = "Pure "
	}

	return "<" + pure_s + "Builtin>"
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
	Status      int
	Message     string
	Parent      *Error
	Values      map[string]Object
	CodeContext *RyeCtx
	CodeBlock   TSeries
}

// Type returns the type of the Integer.
func (i Error) Type() Type {
	return ErrorType
}

// Inspect returns a string representation of the Integer.
func (i Error) Inspect(e Idxs) string {
	var b strings.Builder
	b.WriteString("<Error " + strconv.Itoa(i.Status) + ": " + i.Message + " ")
	if i.Parent != nil {
		b.WriteString(i.Parent.Probe(e))
	}
	b.WriteString(">")
	return b.String()
}

// Inspect returns a string representation of the Integer.
func (i Error) Probe(e Idxs) string {
	var b strings.Builder
	b.WriteString("Error: " + i.Message + " ")
	if i.Parent != nil {
		b.WriteString("\n\t" + i.Parent.Probe(e))
	}
	b.WriteString("")
	return b.String()
}

func NewError1(status int) *Error {
	var e Error
	e.Status = status
	return &e
}

func NewError2(status int, message string) *Error {
	var e Error
	e.Message = message
	e.Status = status
	return &e
}

func NewError(message string) *Error {
	var e Error
	e.Message = message
	return &e
}

func NewError4(status int, message string, error *Error, values map[string]Object) *Error {
	var e Error
	e.Status = status
	e.Message = message
	e.Parent = error
	e.Values = values
	return &e
}

func (i Error) Trace(msg string) {
	fmt.Print(msg + "(error): ")
	fmt.Println(i.Message)
}

func (i Error) GetKind() int {
	return int(IntegerType)
}

//
// ARGWORD
//

type Argword struct {
	Name Word
	Kind Word
}

// Type returns the type of the Integer.
func (i Argword) Type() Type {
	return ArgwordType
}

// Inspect returns a string
func (i Argword) Inspect(e Idxs) string {
	return "<Argword: " + i.Name.Inspect(e) + ":" + i.Kind.Inspect(e) + ">"
}

// Inspect returns a string representation of the Integer.
func (b Argword) Probe(e Idxs) string {
	return b.Name.Probe(e)
}

func (i Argword) Trace(msg string) {
	fmt.Print(msg + " (argword): ")
	//fmt.Println(i.Name.Probe())
}

func (i Argword) GetKind() int {
	return int(WordType)
}

//
// CPATH
//

type CPath struct {
	Cnt   int
	Word1 Word
	Word2 Word
	Word3 Word
}

// Type returns the type of the Integer.
func (i CPath) Type() Type {
	return CPathType
}

// Inspect returns a string
func (i CPath) Inspect(e Idxs) string {
	switch i.Cnt {
	case 2:
		return "<CPath: " + i.Word1.Inspect(e) + "/" + i.Word2.Inspect(e) + ">"
	case 3:
		return "<CPath: " + i.Word1.Inspect(e) + "/" + i.Word2.Inspect(e) + "/" + i.Word3.Inspect(e) + ">"
	}
	return "<CPath: " + i.Word1.Inspect(e) + "/ ... >"
}

// Inspect returns a string
func (o CPath) GetWordNumber(i int) Word {
	switch i {
	case 1:
		return o.Word1
	case 2:
		return o.Word2
	default:
		return o.Word3 // TODO -- just temporary this wasy ... make ultil depth 5 or 6 and return error otherwises
	}
}

// Inspect returns a string representation of the Integer.
func (b CPath) Probe(e Idxs) string {
	return b.Word1.Probe(e)
}

func (i CPath) Trace(msg string) {
	fmt.Print(msg + " (cpath): ")
}

func (i CPath) GetKind() int {
	return int(CPathType)
}

func NewCPath2(w1 Word, w2 Word) *CPath {
	var cp CPath
	cp.Cnt = 2
	cp.Word1 = w1
	cp.Word2 = w2
	return &cp

}
func NewCPath3(w1 Word, w2 Word, w3 Word) *CPath {
	var cp CPath
	cp.Cnt = 3
	cp.Word1 = w1
	cp.Word2 = w2
	cp.Word3 = w3
	return &cp
}

//
// NATIVE
//

// String represents an string.
type Native struct {
	Value interface{}
	Kind  Word
}

func NewNative(index *Idxs, val interface{}, kind string) *Native {
	idx := index.IndexWord(kind)
	nat := Native{val, Word{idx}}
	return &nat
}

// Type returns the type of the Integer.
func (i Native) Type() Type {
	return NativeType
}

// Inspect returns a string representation of the Integer.
func (i Native) Inspect(e Idxs) string {
	return "<Native of kind " + i.Kind.Probe(e) + ">"
}

// Inspect returns a string representation of the Integer.
func (i Native) Probe(e Idxs) string {
	return "<Native of kind " + i.Kind.Probe(e) + ">"
}

func (i Native) Trace(msg string) {
	fmt.Print(msg + "(native): ")
	//fmt.Println(i.Value)
}

func (i Native) GetKind() int {
	return i.Kind.Index
}

//
// Dict -- nonindexed and unboxed map ... for example for params from request etc, so we don't neet to idex keys and it doesn't need boxed values
// I think it should have option of having Kind too ...
//

// String represents an string.
type Dict struct {
	Data map[string]interface{}
	Kind Word
}

func NewDict(data map[string]interface{}) *Dict {
	return &Dict{data, Word{0}}
}

func NewDictFromSeries(block TSeries) Dict {
	data := make(map[string]interface{})
	for block.Pos() < block.Len() {
		key := block.Pop()
		val := block.Pop()
		// v001 -- only process the typical case of string val
		switch k := key.(type) {
		case String:
			data[k.Value] = val
		}
	}
	return Dict{data, Word{0}}
}

// Type returns the type of the Integer.
func (i Dict) Type() Type {
	return DictType
}

// Inspect returns a string representation of the Integer.
func (i Dict) Inspect(idxs Idxs) string {
	var bu strings.Builder
	bu.WriteString("<Dict (" + i.Kind.Probe(idxs) + "): ")
	for k, v := range i.Data {
		switch ob := v.(type) {
		case Object:
			bu.WriteString(k + ": " + ob.Inspect(idxs) + " ")
		default:
			bu.WriteString(k + ": " + fmt.Sprint(ob) + " ")
		}
	}
	bu.WriteString(">")
	return bu.String()
}

// Inspect returns a string representation of the Integer.
func (i Dict) Probe(e Idxs) string {
	return i.Inspect(e)
}

func (i Dict) Trace(msg string) {
	fmt.Print(msg + "(Dict): ")
}

func (i Dict) GetKind() int {
	return i.Kind.Index
}

//
// List -- nonindexed and unboxed list (block)
//

type List struct {
	Data []interface{}
	Kind Word
}

func NewList(data []interface{}) *List {
	return &List{data, Word{0}}
}

func NewListFromSeries(block TSeries) List {
	data := make([]interface{}, block.Len())
	for block.Pos() < block.Len() {
		k1 := block.Pop()
		i := block.Pos()
		switch k := k1.(type) {
		case String:
			data[i] = k.Value
		case Integer:
			data[i] = k.Value
		}
	}
	return List{data, Word{0}}
}

// Type returns the type of the Integer.
func (i List) Type() Type {
	return ListType
}

// Inspect returns a string representation of the Integer.
func (i List) Inspect(idxs Idxs) string {
	var bu strings.Builder
	bu.WriteString("<List (" + i.Kind.Probe(idxs) + "): ")
	for _, v := range i.Data {
		switch ob := v.(type) {
		case map[string]interface{}:
			vv := NewDict(ob)
			bu.WriteString(" " + vv.Inspect(idxs) + " ")
		case Object:
			bu.WriteString(" " + ob.Inspect(idxs) + " ")
		default:
			bu.WriteString(" " + fmt.Sprint(ob) + " ")
		}
	}
	bu.WriteString(">")
	return bu.String()
}

// Inspect returns a string representation of the Integer.
func (i List) Probe(e Idxs) string {
	return i.Inspect(e)
}

func (i List) Trace(msg string) {
	fmt.Print(msg + "(List): ")
}

func (i List) GetKind() int {
	return i.Kind.Index
}

// KIND Type

//
// Kind
//

type Kind struct {
	Kind       Word
	Spec       Block
	Converters map[int]Block
}

func NewKind(kind Word, spec Block) *Kind {
	var o Kind // o := Kind{kind, spec}
	o.Kind = kind
	o.Spec = spec
	o.Converters = make(map[int]Block)
	return &o
}

func (i Kind) Type() Type {
	return KindType
}

// Inspect returns a string representation of the Integer.
func (i Kind) Inspect(e Idxs) string {
	// LONG DISPLAY OF FUNCTION NODES return "<Function: " + i.Spec.Inspect(e) + ", " + i.Body.Inspect(e) + ">"
	return "<Kind(" + i.Kind.Probe(e) + "): " + i.Spec.Inspect(e) + ">"
}

// Inspect returns a string representation of the Integer.
func (i Kind) Probe(e Idxs) string {
	return "<Kind: " + i.Spec.Inspect(e) + ">"
}

func (i Kind) Trace(msg string) {
	fmt.Print(msg + " (Kind): ")
	fmt.Println(i.Spec)
}

func (i Kind) GetKind() int {
	return int(KindType)
}

func (i Kind) SetConverter(from int, spec Block) {
	i.Converters[from] = spec
}

func (i Kind) HasConverter(from int) bool {
	if _, ok := i.Converters[from]; ok {
		return true
	} else {
		return false
	}
}

//
// Converter
//

type Converter struct {
	From Word
	To   Word
	Spec Block
}

func NewConverter(from Word, to Word, spec Block) *Converter {
	o := Converter{from, to, spec}
	return &o
}

func (i Converter) Type() Type {
	return ConverterType
}

// Inspect returns a string representation of the Integer.
func (i Converter) Inspect(e Idxs) string {
	// LONG DISPLAY OF FUNCTION NODES return "<Function: " + i.Spec.Inspect(e) + ", " + i.Body.Inspect(e) + ">"
	return "<Converter(" + i.From.Probe(e) + "->" + i.To.Probe(e) + "): " + i.Spec.Inspect(e) + ">"
}

// Inspect returns a string representation of the Integer.
func (i Converter) Probe(e Idxs) string {
	return i.Spec.Inspect(e)
}

func (i Converter) Trace(msg string) {
	fmt.Print(msg + " (Converter): ")
	fmt.Println(i.Spec)
}

func (i Converter) GetKind() int {
	return int(ConverterType)
}

//
// TIME
//

type Time struct {
	Value time.Time
}

// Type returns the type of the Integer.
func (i Time) Type() Type {
	return TimeType
}

// Inspect returns a string representation of the Integer.
func (i Time) Inspect(e Idxs) string {
	return "<Time: " + i.Probe(e) + ">"
}

// Inspect returns a string representation of the Integer.
func (i Time) Probe(e Idxs) string {
	return i.Value.Format("2006-01-02 15:04:05")
}

func (i Time) Trace(msg string) {
	fmt.Print(msg + "(time): ")
	fmt.Println(i.Value)
}

func (i Time) GetKind() int {
	return int(TimeType)
}
