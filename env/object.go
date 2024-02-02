// object.go
package env

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/drewlanenga/govector"
)

type Type int

const (
	BlockType          Type = 1
	IntegerType        Type = 2
	WordType           Type = 3
	SetwordType        Type = 4
	OpwordType         Type = 5
	PipewordType       Type = 6
	BuiltinType        Type = 7
	FunctionType       Type = 8
	ErrorType          Type = 9
	CommaType          Type = 10
	VoidType           Type = 11
	StringType         Type = 12
	TagwordType        Type = 13
	GenwordType        Type = 14
	GetwordType        Type = 15
	ArgwordType        Type = 16
	NativeType         Type = 17
	UriType            Type = 18
	LSetwordType       Type = 19
	CtxType            Type = 20
	DictType           Type = 21
	ListType           Type = 22
	DateType           Type = 23
	CPathType          Type = 24
	XwordType          Type = 25
	EXwordType         Type = 26
	SpreadsheetType    Type = 27
	EmailType          Type = 28
	KindType           Type = 29
	KindwordType       Type = 30
	ConverterType      Type = 31
	TimeType           Type = 32
	SpreadsheetRowType Type = 33
	DecimalType        Type = 34
	VectorType         Type = 35
)

// after adding new type here, also add string to idxs.go

type Object interface {
	Type() Type
	Trace(msg string)
	GetKind() int
	Equal(p Object) bool
	Inspect(e Idxs) string
	Print(e Idxs) string
	Dump(e Idxs) string
}

//
// INTEGER
//

// Integer represents an integer.
type Integer struct {
	Value int64
}

func NewInteger(val int64) *Integer {
	nat := Integer{val}
	return &nat
}

// Type returns the type of the Integer.
func (i Integer) Type() Type {
	return IntegerType
}

// Inspect returns a string representation of the Integer.
func (i Integer) Inspect(e Idxs) string {
	return "[Integer: " + strconv.FormatInt(i.Value, 10) + "]"
}

// Inspect returns a string representation of the Integer.
func (i Integer) Print(e Idxs) string {
	return strconv.FormatInt(i.Value, 10)
}

func (i Integer) Trace(msg string) {
	fmt.Print(msg + "(integer): ")
	fmt.Println(i.Value)
}

func (i Integer) GetKind() int {
	return int(IntegerType)
}

func (i Integer) Equal(o Object) bool {
	if i.Type() != o.Type() {
		return false
	}
	return i.Value == o.(Integer).Value
}

func (i Integer) Dump(e Idxs) string {
	return strconv.FormatInt(i.Value, 10)
}

//
// DECIMAL
//

// Decimal
type Decimal struct {
	Value float64 `bson:"value"`
}

func NewDecimal(val float64) *Decimal {
	nat := Decimal{val}
	return &nat
}

// Type returns the type of the Decimal.
func (i Decimal) Type() Type {
	return DecimalType
}

// Inspect returns a string representation of the Decimal.
func (i Decimal) Inspect(e Idxs) string {
	return "[Decimal: " + strconv.FormatFloat(i.Value, 'f', 6, 64) + "]"
}

// Inspect returns a string representation of the Decimal.
func (i Decimal) Print(e Idxs) string {
	return strconv.FormatFloat(i.Value, 'f', 6, 64)
}

func (i Decimal) Trace(msg string) {
	fmt.Print(msg + "(Decimal): ")
	fmt.Println(i.Value)
}

func (i Decimal) GetKind() int {
	return int(DecimalType)
}

func (i Decimal) Equal(o Object) bool {
	if i.Type() != o.Type() {
		return false
	}
	return i.Value == o.(Decimal).Value
}

func (i Decimal) Dump(e Idxs) string {
	return strconv.FormatFloat(i.Value, 'f', -1, 64)
}

//
// STRING
//

// String represents an string.
type String struct {
	Value string `bson:"value"`
}

func NewString(val string) *String {
	nat := String{val}
	return &nat
}

// Type returns the type of the Integer.
func (i String) Type() Type {
	return StringType
}

// Inspect returns a string representation of the Integer.
func (i String) Inspect(e Idxs) string {
	return "[String: " + i.Value + "]"
}

// Inspect returns a string representation of the Integer.
func (i String) Print(e Idxs) string {
	s := i.Value
	if len(s) > 80 {
		return s[:80] + "..."
	}
	return s
	// return "\"" + s + "\""
}

func (i String) Trace(msg string) {
	fmt.Print(msg + "(string): ")
	fmt.Println(i.Value)
}

func (i String) GetKind() int {
	return int(StringType)
}

func (i String) Equal(o Object) bool {
	if i.Type() != o.Type() {
		return false
	}
	return i.Value == o.(String).Value
}

func (i String) Dump(e Idxs) string {
	return i.Value
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

func NewDate(val time.Time) *Date {
	nat := Date{val}
	return &nat
}

func (i Date) Inspect(e Idxs) string {
	return "[Date: " + i.Value.Format(time.RFC822Z) + "]"
}

func (i Date) Print(e Idxs) string {
	return i.Value.Format(time.RFC822Z)
}

func (i Date) Trace(msg string) {
	fmt.Print(msg + "(date): ")
	fmt.Println(i.Value.Format(time.RFC822Z))
}

func (i Date) GetKind() int {
	return int(DateType)
}

func (i Date) Equal(o Object) bool {
	if i.Type() != o.Type() {
		return false
	}
	return i.Value == o.(Date).Value
}

func (i Date) Dump(e Idxs) string {
	return i.Value.Format(time.DateOnly)
}

//
// URI
//

type Uri struct {
	Scheme Word
	Path   string
	Kind   Word
}

func NewUri1(index *Idxs, path string) *Uri {
	scheme2 := strings.Split(path, "://")
	scheme := scheme2[0] // + "-schema" // TODO -- this is just temporary .. so we test it further, make proper once at that level
	idxSch := index.IndexWord(scheme)
	kind := scheme + "-schema"
	idxKind := index.IndexWord(kind)
	nat := Uri{Word{idxSch}, scheme2[1], Word{idxKind}}
	return &nat
}

func NewUri(index *Idxs, scheme Word, path string) *Uri {
	scheme2 := strings.Split(path, "://")
	kindstr := strings.Split(path, "://")[0] + "-schema" // TODO -- this is just temporary .. so we test it further, make proper once at that level
	idx := index.IndexWord(kindstr)
	nat := Uri{scheme, scheme2[1], Word{idx}}
	//	nat := Uri{Word{idxSch}, scheme2[1], Word{idxKind}}
	return &nat
}

func (i Uri) GetPath() string {
	return i.Path
}

func (i Uri) GetProtocol() Word {
	return i.Scheme
}

func (i Uri) Type() Type {
	return UriType
}

// Inspect returns a string representation of the Integer.
func (i Uri) Inspect(e Idxs) string {
	return "[Uri: " + i.Scheme.Print(e) + "://" + i.GetPath() + "]"
}

// Inspect returns a string representation of the Integer.
func (i Uri) Print(e Idxs) string {
	return i.Path
}

func (i Uri) Trace(msg string) {
	fmt.Print(msg + "(uri): ")
	fmt.Println(i.Path)
}

func (i Uri) GetKind() int {
	return i.Kind.Index
}

func (i Uri) Equal(o Object) bool {
	if i.Type() != o.Type() {
		return false
	}
	oUri := o.(Uri)
	return i.Path == oUri.Path && i.Scheme.Equal(oUri.Scheme) && i.Kind.Equal(oUri.Kind)
}

func (i Uri) Dump(e Idxs) string {
	return e.GetWord(i.Scheme.Index) + "://" + i.Path
}

//
// Email
//

type Email struct {
	Address string
}

func NewEmail(address string) *Email {
	nat := Email{address}
	return &nat
}

func (i Email) Type() Type {
	return EmailType
}

// Inspect returns a string representation of the Integer.
func (i Email) Inspect(e Idxs) string {
	return "[Email: " + i.Print(e) + "]"
}

// Inspect returns a string representation of the Integer.
func (i Email) Print(e Idxs) string {
	return i.Address
}

func (i Email) Trace(msg string) {
	fmt.Print(msg + "(email): ")
	fmt.Println(i.Address)
}

func (i Email) GetKind() int {
	return int(EmailType)
}

func (i Email) Equal(o Object) bool {
	if i.Type() != o.Type() {
		return false
	}
	return i.Address == o.(Email).Address
}

func (i Email) Dump(e Idxs) string {
	return i.Address
}

//
// BLOCK
//

// Integer represents an integer.
type Block struct {
	Series TSeries `bson:"series"`
	Mode   int     `bson:"mode"`
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
	r.WriteString("[Block: ")
	for i := 0; i < b.Series.Len(); i += 1 {
		if b.Series.Get(i) != nil {
			r.WriteString(b.Series.Get(i).Inspect(e))
			r.WriteString(" ")
		}
	}
	r.WriteString("]")
	return r.String()
}

// Inspect returns a string representation of the Integer.
func (b Block) Print(e Idxs) string {
	var r strings.Builder
	r.WriteString("{ ")
	for i := 0; i < b.Series.Len(); i += 1 {
		if b.Series.Get(i) != nil {
			r.WriteString(b.Series.Get(i).Print(e))
			r.WriteString(" ")
		} else {
			r.WriteString("[NIL]")
		}
	}
	r.WriteString("}")
	return r.String()
}

func (i Block) Trace(msg string) {
	fmt.Print(msg + " (block): ")
	fmt.Println(i.Series)
}

func (i Block) GetKind() int {
	return int(BlockType)
}

func (i Block) Equal(o Object) bool {
	if i.Type() != o.Type() {
		return false
	}
	oBlock := o.(Block)
	if i.Series.Len() != oBlock.Series.Len() {
		return false
	}
	if i.Mode != oBlock.Mode {
		return false
	}
	for j := 0; j < i.Series.Len(); j += 1 {
		if !i.Series.Get(j).Equal(oBlock.Series.Get(j)) {
			return false
		}
	}
	return true
}

func (i Block) Dump(e Idxs) string {
	var bu strings.Builder
	bu.WriteString("{ ")
	for _, obj := range i.Series.GetAll() {
		if obj != nil {
			bu.WriteString(fmt.Sprintf("%s ", obj.Dump(e)))
		} else {
			bu.WriteString("'nil ")
		}
	}
	bu.WriteString("}")
	return bu.String()
}

//
// WORD
//

// Integer represents an integer.
type Word struct {
	Index int
}

func NewWord(val int) *Word {
	nat := Word{val}
	return &nat
}

// Type returns the type of the Integer.
func (i Word) Type() Type {
	return WordType
}

// Inspect returns a string
func (i Word) Inspect(e Idxs) string {
	return "[Word: " + strconv.FormatInt(int64(i.Index), 10) + ", " + e.GetWord(i.Index) + "]"
}

// Inspect returns a string representation of the Integer.
func (b Word) Print(e Idxs) string {
	return e.GetWord(b.Index)
}

func (i Word) Trace(msg string) {
	fmt.Print(msg + " (word): ")
	fmt.Println(i.Index)
}

func (i Word) GetKind() int {
	return int(WordType)
}

func (i Word) Equal(o Object) bool {
	if i.Type() != o.Type() {
		return false
	}
	return i.Index == o.(Word).Index
}

func (i Word) Dump(e Idxs) string {
	return e.GetWord(i.Index)
}

//
// SETWORD
//

// Integer represents an integer.
type Setword struct {
	Index int
}

func NewSetword(index int) *Setword {
	nat := Setword{index}
	return &nat
}

// Type returns the type of the Integer.
func (i Setword) Type() Type {
	return SetwordType
}

// Inspect returns a string representation of the Integer.
func (i Setword) Inspect(e Idxs) string {
	return "[Setword: " + strconv.FormatInt(int64(i.Index), 10) + ", " + e.GetWord(i.Index) + "]"
}

// Inspect returns a string representation of the Integer.
func (b Setword) Print(e Idxs) string {
	return e.GetWord(b.Index) + ":"
}

func (i Setword) Trace(msg string) {
	fmt.Print(msg + "(setword): ")
	fmt.Println(i.Index)
}

func (i Setword) GetKind() int {
	return int(SetwordType)
}

func (i Setword) Equal(o Object) bool {
	if i.Type() != o.Type() {
		return false
	}
	return i.Index == o.(Setword).Index
}

func (i Setword) Dump(e Idxs) string {
	return fmt.Sprintf("%s:", e.GetWord(i.Index))
}

//
// LSETWORD
//

// Integer represents an integer.
type LSetword struct {
	Index int
}

func NewLSetword(index int) *LSetword {
	nat := LSetword{index}
	return &nat
}

// Type returns the type of the Integer.
func (i LSetword) Type() Type {
	return LSetwordType
}

// Inspect returns a string representation of the Integer.
func (i LSetword) Inspect(e Idxs) string {
	return "[LSetword: " + strconv.FormatInt(int64(i.Index), 10) + ", " + e.GetWord(i.Index) + "]"
}

// Inspect returns a string representation of the Integer.
func (b LSetword) Print(e Idxs) string {
	return ":" + e.GetWord(b.Index)
}

func (i LSetword) Trace(msg string) {
	fmt.Print(msg + "(lsetword): ")
	fmt.Println(i.Index)
}

func (i LSetword) GetKind() int {
	return int(LSetwordType)
}

func (i LSetword) Equal(o Object) bool {
	if i.Type() != o.Type() {
		return false
	}
	return i.Index == o.(LSetword).Index
}

func (i LSetword) Dump(e Idxs) string {
	return fmt.Sprintf(":%s", e.GetWord(i.Index))
}

//
// OPWORD
//

// Integer represents an integer.
type Opword struct {
	Index int
	Force int
}

func NewOpword(index, force int) *Opword {
	nat := Opword{index, force}
	return &nat
}

// Type returns the type of the Integer.
func (i Opword) Type() Type {
	return OpwordType
}

// Inspect returns a string
func (i Opword) Inspect(e Idxs) string {
	return "[Opword: " + strconv.FormatInt(int64(i.Index), 10) + ", " + e.GetWord(i.Index) + "]"
}

// Inspect returns a string representation of the Integer.
func (b Opword) Print(e Idxs) string {
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

func (i Opword) Equal(o Object) bool {
	if i.Type() != o.Type() {
		return false
	}
	oOpword := o.(Opword)
	return i.Index == oOpword.Index && i.Force == oOpword.Force
}

func (i Opword) Dump(e Idxs) string {
	return fmt.Sprintf(".%s", e.GetWord(i.Index))
}

//
// PIPEWORD
//

// Integer represents an integer.
type Pipeword struct {
	Index int
	Force int
}

func NewPipeword(index, force int) *Pipeword {
	nat := Pipeword{index, force}
	return &nat
}

// Type returns the type of the Integer.
func (i Pipeword) Type() Type {
	return PipewordType
}

// Inspect returns a string
func (i Pipeword) Inspect(e Idxs) string {
	return "[Pipeword: " + strconv.FormatInt(int64(i.Index), 10) + ", " + e.GetWord(i.Index) + "]"
}

// Inspect returns a string representation of the Integer.
func (b Pipeword) Print(e Idxs) string {
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

func (i Pipeword) Equal(o Object) bool {
	if i.Type() != o.Type() {
		return false
	}
	oPipeword := o.(Pipeword)
	return i.Index == oPipeword.Index && i.Force == oPipeword.Force
}

func (i Pipeword) Dump(e Idxs) string {
	return fmt.Sprintf("|%s", e.GetWord(i.Index))
}

//
// TAGWORD
//

// Integer represents an integer.
type Tagword struct {
	Index int
}

func NewTagword(index int) *Tagword {
	nat := Tagword{index}
	return &nat
}

// Type returns the type of the Integer.
func (i Tagword) Type() Type {
	return TagwordType
}

// Inspect returns a string
func (i Tagword) Inspect(e Idxs) string {
	return "[Tagword: " + strconv.FormatInt(int64(i.Index), 10) + ", " + e.GetWord(i.Index) + "]"
}

// Inspect returns a string representation of the Integer.
func (b Tagword) Print(e Idxs) string {
	return "'" + e.GetWord(b.Index)
}

func (i Tagword) Trace(msg string) {
	fmt.Print(msg + " (tagword): ")
	fmt.Println(i.Index)
}

func (i Tagword) ToWord() Word {
	return Word(i)
}

func (i Tagword) GetKind() int {
	return int(TagwordType)
}

func (i Tagword) Equal(o Object) bool {
	if i.Type() != o.Type() {
		return false
	}
	return i.Index == o.(Tagword).Index
}

func (i Tagword) Dump(e Idxs) string {
	return fmt.Sprintf("'%s", e.GetWord(i.Index))
}

//
// XWORD
//

// Integer represents an integer.
type Xword struct {
	Index int
}

func NewXword(index int) *Xword {
	nat := Xword{index}
	return &nat
}

// Type returns the type of the Integer.
func (i Xword) Type() Type {
	return XwordType
}

// Inspect returns a string
func (i Xword) Inspect(e Idxs) string {
	return "[Xword: " + strconv.FormatInt(int64(i.Index), 10) + ", " + e.GetWord(i.Index) + "]"
}

// Inspect returns a string representation of the Integer.
func (b Xword) Print(e Idxs) string {
	return e.GetWord(b.Index)
}

func (i Xword) Trace(msg string) {
	fmt.Print(msg + " (Xword): ")
	fmt.Println(i.Index)
}

func (i Xword) ToWord() Word {
	return Word(i)
}

func (i Xword) GetKind() int {
	return int(XwordType)
}

func (i Xword) Equal(o Object) bool {
	if i.Type() != o.Type() {
		return false
	}
	return i.Index == o.(Xword).Index
}

func (i Xword) Dump(e Idxs) string {
	return fmt.Sprintf("<%s>", e.GetWord(i.Index))
}

//
// EXWORD
//

// Integer represents an integer.
type EXword struct {
	Index int
}

func NewEXword(index int) *EXword {
	nat := EXword{index}
	return &nat
}

// Type returns the type of the Integer.
func (i EXword) Type() Type {
	return EXwordType
}

// Inspect returns a string
func (i EXword) Inspect(e Idxs) string {
	return "[EXword: " + strconv.FormatInt(int64(i.Index), 10) + ", " + e.GetWord(i.Index) + "]"
}

// Inspect returns a string representation of the Integer.
func (b EXword) Print(e Idxs) string {
	return e.GetWord(b.Index)
}

func (i EXword) Trace(msg string) {
	fmt.Print(msg + " (EXword): ")
	fmt.Println(i.Index)
}

func (i EXword) ToWord() Word {
	return Word(i)
}

func (i EXword) GetKind() int {
	return int(EXwordType)
}

func (i EXword) Equal(o Object) bool {
	if i.Type() != o.Type() {
		return false
	}
	return i.Index == o.(EXword).Index
}

func (i EXword) Dump(e Idxs) string {
	return fmt.Sprintf("</%s>", e.GetWord(i.Index))
}

//
// KINDWORD
//

// Integer represents an integer.
type Kindword struct {
	Index int
}

func NewKindword(index int) *Kindword {
	nat := Kindword{index}
	return &nat
}

// Type returns the type of the Integer.
func (i Kindword) Type() Type {
	return KindwordType
}

// Inspect returns a string
func (i Kindword) Inspect(e Idxs) string {
	return "[Kindword: " + strconv.FormatInt(int64(i.Index), 10) + ", " + e.GetWord(i.Index) + "]"
}

// Inspect returns a string representation of the Integer.
func (b Kindword) Print(e Idxs) string {
	return e.GetWord(b.Index)
}

func (i Kindword) Trace(msg string) {
	fmt.Print(msg + " (Kidnword): ")
	fmt.Println(i.Index)
}

func (i Kindword) ToWord() Word {
	return Word(i)
}

func (i Kindword) GetKind() int {
	return int(KindwordType)
}

func (i Kindword) Equal(o Object) bool {
	if i.Type() != o.Type() {
		return false
	}
	return i.Index == o.(Kindword).Index
}

func (i Kindword) Dump(e Idxs) string {
	return fmt.Sprintf("~(%s)", e.GetWord(i.Index))
}

//
// GETWORD
//

// Integer represents an integer.
type Getword struct {
	Index int
}

func NewGetword(index int) *Getword {
	nat := Getword{index}
	return &nat
}

// Type returns the type of the Integer.
func (i Getword) Type() Type {
	return GetwordType
}

// Inspect returns a string
func (i Getword) Inspect(e Idxs) string {
	return "[Getword: " + strconv.FormatInt(int64(i.Index), 10) + ", " + e.GetWord(i.Index) + "]"
}

// Inspect returns a string representation of the Integer.
func (b Getword) Print(e Idxs) string {
	return "?" + e.GetWord(b.Index)
}

func (i Getword) Trace(msg string) {
	fmt.Print(msg + " (getword): ")
	fmt.Println(i.Index)
}

func (i Getword) ToWord() Word {
	return Word(i)
}

func (i Getword) GetKind() int {
	return int(GetwordType)
}

func (i Getword) Equal(o Object) bool {
	if i.Type() != o.Type() {
		return false
	}
	return i.Index == o.(Getword).Index
}

func (i Getword) Dump(e Idxs) string {
	return fmt.Sprintf("?%s", e.GetWord(i.Index))
}

//
// GENWORD
//

// Integer represents an integer.
type Genword struct {
	Index int
}

func NewGenword(index int) *Genword {
	nat := Genword{index}
	return &nat
}

// Type returns the type of the Integer.
func (i Genword) Type() Type {
	return GenwordType
}

// Inspect returns a string
func (i Genword) Inspect(e Idxs) string {
	return "[Genword: " + strconv.FormatInt(int64(i.Index), 10) + ", " + e.GetWord(i.Index) + "]"
}

// Inspect returns a string representation of the Integer.
func (b Genword) Print(e Idxs) string {
	return e.GetWord(b.Index)
}

func (i Genword) Trace(msg string) {
	fmt.Print(msg + " (genword): ")
	fmt.Println(i.Index)
}

func (i Genword) ToWord() Word {
	return Word(i)
}

func (i Genword) GetKind() int {
	return int(GenwordType)
}

func (i Genword) Equal(o Object) bool {
	if i.Type() != o.Type() {
		return false
	}
	return i.Index == o.(Genword).Index
}

func (i Genword) Dump(e Idxs) string {
	// TODO not sure if this is correct
	return fmt.Sprintf("~%s", e.GetWord(i.Index))
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
	return "[Comma]"
}

// Inspect returns a string representation of the Integer.
func (b Comma) Print(e Idxs) string {
	return ","
}

func (i Comma) Trace(msg string) {
	fmt.Print(msg + " (comma)")
}

func (i Comma) GetKind() int {
	return int(CommaType)
}

func (i Comma) Equal(o Object) bool {
	return i.Type() == o.Type()
}

func (i Comma) Dump(e Idxs) string {
	return ","
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
	return "[Void]"
}

// Inspect returns a string representation of the Integer.
func (b Void) Print(e Idxs) string {
	return "_"
}

func (i Void) Trace(msg string) {
	fmt.Print(msg + " (void)")
}

func (i Void) GetKind() int {
	return int(VoidType)
}

func (i Void) Equal(o Object) bool {
	return i.Type() == o.Type()
}

func (i Void) Dump(e Idxs) string {
	return "_"
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
	Doc   string
}

func NewFunction(spec Block, body Block, pure bool) *Function {
	o := Function{spec.Series.Len(), spec, body, nil, pure, ""}
	return &o
}

func NewFunctionC(spec Block, body Block, ctx *RyeCtx, pure bool) *Function {
	o := Function{spec.Series.Len(), spec, body, ctx, pure, ""}
	return &o
}

func NewFunctionDoc(spec Block, body Block, pure bool, doc string) *Function {
	var argn int
	if doc > "" {
		argn = spec.Series.Len() - 1
	} else {
		argn = spec.Series.Len()
	}
	o := Function{argn, spec, body, nil, pure, doc}
	return &o
}

// Type returns the type of the Integer.
func (i Function) Type() Type {
	return FunctionType
}

// Inspect returns a string representation of the Integer.
func (i Function) Inspect(e Idxs) string {
	// LONG DISPLAY OF FUNCTION NODES return "[Function: " + i.Spec.Inspect(e) + ", " + i.Body.Inspect(e) + "]"
	var pure_s string
	if i.Pure {
		pure_s = "Pure "
	}
	docs := ""
	if len(i.Doc) > 0 {
		docs = ": " + i.Doc
	}
	return "[" + pure_s + "Function(" + strconv.FormatInt(int64(i.Argsn), 10) + ")" + docs + "]"
}

// Inspect returns a string representation of the Integer.
func (b Function) Print(e Idxs) string {
	return "[Function(" + strconv.FormatInt(int64(b.Argsn), 10) + ")]"
}

// Inspect returns a string representation of the Integer.
// func (i Function) Dump(e Idxs) Block {
// 	// LONG DISPLAY OF FUNCTION NODES return "[Function: " + i.Spec.Inspect(e) + ", " + i.Body.Inspect(e) + "]"
// 	ser := make([]Object, 0)
// 	idx, found := e.GetIndex("fn")
// 	if !found {
// 		goto ENE // TODO
// 	}
// 	ser = append(ser, Word{idx})
// 	ser = append(ser, i.Spec)
// 	ser = append(ser, i.Body)
// ENE:
// 	return *NewBlock(*NewTSeries(ser))
// }

func (i Function) Trace(msg string) {
	fmt.Print(msg + " (function): ")
	fmt.Println(i.Spec)
}

func (i Function) GetKind() int {
	return int(FunctionType)
}

func (i Function) Equal(o Object) bool {
	if i.Type() != o.Type() {
		return false
	}
	oFunction := o.(Function)
	if i.Argsn != oFunction.Argsn {
		return false
	}
	if !i.Spec.Equal(oFunction.Spec) {
		return false
	}
	if !i.Body.Equal(oFunction.Body) {
		return false
	}
	if i.Pure != oFunction.Pure {
		return false
	}
	return true
}

func (i Function) Dump(e Idxs) string {
	return fmt.Sprintf("\"serlization of %s is not yet supported\" ", i.Inspect(e))
}

//
// BuiltinFunction
//

// BuiltinFunction represents a function signature of builtin functions.
// ///type BuiltinFunction func(ps *ProgramState, args ...Object) Object
type BuiltinFunction func(ps *ProgramState, arg0 Object, arg1 Object, arg2 Object, arg3 Object, arg4 Object) Object

// Builtin represents a builtin function.
// TODO: Builtin is just temporary ... we need to make something else, that holds natives and user functions. Interface should be the same ...
// would it be better (faster) to have concrete type probably.
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
	Doc           string
}

func NewBuiltin(fn BuiltinFunction, argsn int, acceptFailure bool, pure bool, doc string) *Builtin {
	bl := Builtin{fn, argsn, nil, nil, nil, nil, nil, acceptFailure, pure, doc}
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

	return "[" + pure_s + "BFunction(" + strconv.Itoa(b.Argsn) + "): " + b.Doc + "]"
}

func (b Builtin) Print(e Idxs) string {
	var pure_s string
	if b.Pure {
		pure_s = "Pure "
	}

	// return "[" + pure_s + "Builtin]"
	return "[" + pure_s + "BFunction(" + strconv.Itoa(b.Argsn) + "): " + b.Doc + "]"
}

func (i Builtin) Trace(msg string) {
	fmt.Print(msg + " (bfunction): ")
	fmt.Println(i.Argsn)
}

func (i Builtin) GetKind() int {
	return int(BuiltinType)
}

func (i Builtin) Equal(o Object) bool {
	if i.Type() != o.Type() {
		return false
	}
	oBuiltin := o.(Builtin)
	if i.Argsn != oBuiltin.Argsn {
		return false
	}
	if i.Cur0 != oBuiltin.Cur0 {
		return false
	}
	if i.Cur1 != oBuiltin.Cur1 {
		return false
	}
	if i.Cur2 != oBuiltin.Cur2 {
		return false
	}
	if i.Cur3 != oBuiltin.Cur3 {
		return false
	}
	if i.Cur4 != oBuiltin.Cur4 {
		return false
	}
	if i.AcceptFailure != oBuiltin.AcceptFailure {
		return false
	}
	if i.Pure != oBuiltin.Pure {
		return false
	}
	return true
}

func (i Builtin) Dump(e Idxs) string {
	// TODO
	return fmt.Sprintf("\"serlization of %s is not yet supported\" ", i.Inspect(e))
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
	return i.Print(e)
}

// Inspect returns a string representation of the Integer.
func (i Error) Print(e Idxs) string {
	status := ""
	if i.Status != 0 {
		status = "(" + strconv.Itoa(i.Status) + ")"
	}
	var b strings.Builder
	b.WriteString("Error" + status + ": " + i.Message + " ")
	if i.Parent != nil {
		b.WriteString("\n\t" + i.Parent.Print(e))
	}
	for k, v := range i.Values {
		switch ob := v.(type) {
		case Object:
			b.WriteString("\n\t" + k + ": " + ob.Inspect(e) + " ")
		default:
			b.WriteString("\n\t" + k + ": " + fmt.Sprint(ob) + " ")
		}
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

func (i Error) Equal(o Object) bool {
	if i.Type() != o.Type() {
		return false
	}
	oError := o.(*Error)
	if i.Status != oError.Status {
		return false
	}
	if i.Message != oError.Message {
		return false
	}
	if i.Parent != oError.Parent {
		return false
	}
	if len(i.Values) != len(oError.Values) {
		return false
	}
	for k, v := range i.Values {
		if !v.Equal(oError.Values[k]) {
			return false
		}
	}
	return true
}

func (i Error) Dump(e Idxs) string {
	// TODO
	return fmt.Sprintf("\"serlization of %s is not yet supported\" ", i.Inspect(e))
}

//
// ARGWORD
//

type Argword struct {
	Name Word
	Kind Word
}

func NewArgword(name Word, kind Word) *Argword {
	nat := Argword{name, kind}
	return &nat
}

// Type returns the type of the Integer.
func (i Argword) Type() Type {
	return ArgwordType
}

// Inspect returns a string
func (i Argword) Inspect(e Idxs) string {
	return "[Argword: " + i.Name.Inspect(e) + ":" + i.Kind.Inspect(e) + "]"
}

// Inspect returns a string representation of the Integer.
func (b Argword) Print(e Idxs) string {
	return b.Name.Print(e)
}

func (i Argword) Trace(msg string) {
	fmt.Print(msg + " (argword): ")
	//fmt.Println(i.Name.Probe())
}

func (i Argword) GetKind() int {
	return int(WordType)
}

func (i Argword) Equal(o Object) bool {
	if i.Type() != o.Type() {
		return false
	}
	oArgword := o.(Argword)
	return i.Name.Equal(oArgword.Name) && i.Kind.Equal(oArgword.Kind)
}

func (i Argword) Dump(e Idxs) string {
	// TODO not sure if this is correct
	return fmt.Sprintf("{ %s : %s }", i.Name.Dump(e), i.Kind.Dump(e))
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
		return "[CPath: " + i.Word1.Inspect(e) + "/" + i.Word2.Inspect(e) + "]"
	case 3:
		return "[CPath: " + i.Word1.Inspect(e) + "/" + i.Word2.Inspect(e) + "/" + i.Word3.Inspect(e) + "]"
	}
	return "[CPath: " + i.Word1.Inspect(e) + "/ ... ]"
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
func (b CPath) Print(e Idxs) string {
	return b.Word1.Print(e)
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

func (i CPath) Equal(o Object) bool {
	if i.Type() != o.Type() {
		return false
	}
	oCPath := o.(CPath)
	if i.Cnt != oCPath.Cnt {
		return false
	}
	if !i.Word1.Equal(oCPath.Word1) {
		return false
	}
	if !i.Word2.Equal(oCPath.Word2) {
		return false
	}
	if !i.Word3.Equal(oCPath.Word3) {
		return false
	}
	return true
}

func (i CPath) Dump(e Idxs) string {
	// TODO
	return fmt.Sprintf("\"serlization of %s is not yet supported\" ", i.Inspect(e))
}

//
// NATIVE
//

// String represents an string.
type Native struct {
	Value any
	Kind  Word
}

func NewNative(index *Idxs, val any, kind string) *Native {
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
	return "[Native of kind " + i.Kind.Print(e) + "]"
}

// Inspect returns a string representation of the Integer.
func (i Native) Print(e Idxs) string {
	return "[Native of kind " + i.Kind.Print(e) + "]"
}

func (i Native) Trace(msg string) {
	fmt.Print(msg + "(native): ")
	//fmt.Println(i.Value)
}

func (i Native) GetKind() int {
	return i.Kind.Index
}

func (i Native) Equal(o Object) bool {
	if i.Type() != o.Type() {
		return false
	}
	oNative := o.(Native)
	if !i.Kind.Equal(oNative.Kind) {
		return false
	}
	if iValObj, ok := i.Value.(Object); ok {
		if oValObj, ok := oNative.Value.(Object); ok {
			return iValObj.Equal(oValObj)
		}
	}
	return i.Value == oNative.Value
}

func (i Native) Dump(e Idxs) string {
	// TODO
	return fmt.Sprintf("\"serlization of %s is not yet supported\" ", i.Inspect(e))
}

//
// Dict -- nonindexed and unboxed map ... for example for params from request etc, so we don't neet to idex keys and it doesn't need boxed values
// I think it should have option of having Kind too ...
//

// String represents an string.
type Dict struct {
	Data map[string]any
	Kind Word
}

func NewDict(data map[string]any) *Dict {
	return &Dict{data, Word{0}}
}

func NewDictFromSeries(block TSeries, idx *Idxs) Dict {
	data := make(map[string]any)
	for block.Pos() < block.Len() {
		key := block.Pop()
		val := block.Pop()
		// v001 -- only process the typical case of string val
		switch k := key.(type) {
		case String:
			data[k.Value] = val
		case Tagword:
			data[idx.GetWord(k.Index)] = val
		case Word:
			data[idx.GetWord(k.Index)] = val
		case Setword:
			data[idx.GetWord(k.Index)] = val
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
	bu.WriteString("[Dict (" + i.Kind.Print(idxs) + "): ")
	for k, v := range i.Data {
		switch ob := v.(type) {
		case Object:
			bu.WriteString(k + ": " + ob.Inspect(idxs) + " ")
		default:
			bu.WriteString(k + ": " + fmt.Sprint(ob) + " ")
		}
	}
	bu.WriteString("]")
	return bu.String()
}

// Inspect returns a string representation of the Integer.
func (i Dict) Print(idxs Idxs) string {
	var bu strings.Builder
	bu.WriteString("[\n")
	for k, v := range i.Data {
		switch ob := v.(type) {
		case Object:
			bu.WriteString(" " + k + ": " + ob.Print(idxs) + "\n")
		default:
			bu.WriteString(" " + k + ": " + fmt.Sprint(ob) + "\n")
		}
	}
	bu.WriteString("]")
	return bu.String()
}

func (i Dict) Trace(msg string) {
	fmt.Print(msg + "(Dict): ")
}

func (i Dict) GetKind() int {
	return i.Kind.Index
}

func (i Dict) Equal(o Object) bool {
	if i.Type() != o.Type() {
		return false
	}
	oDict := o.(Dict)
	if !i.Kind.Equal(oDict.Kind) {
		return false
	}
	if len(i.Data) != len(oDict.Data) {
		return false
	}
	for k, v := range i.Data {
		if vObj, ok := v.(Object); ok {
			if oObj, ok := oDict.Data[k].(Object); ok {
				if !vObj.Equal(oObj) {
					return false
				}
			}
		} else {
			if v != oDict.Data[k] {
				return false
			}
		}
	}
	return true
}

func (i Dict) Dump(e Idxs) string {
	var bu strings.Builder
	bu.WriteString("dict { ")
	for k, v := range i.Data {
		switch obj := v.(type) {
		case Object:
			bu.WriteString(fmt.Sprintf("%s %s ", k, obj.Dump(e)))
		default:
			bu.WriteString(fmt.Sprintf("%s \"WARN: serlization of %s is not yet supported\"", k, obj))
		}
	}
	bu.WriteString("}")
	return bu.String()
}

//
// List -- nonindexed and unboxed list (block)
//

type List struct {
	Data []any
	Kind Word
}

func NewList(data []any) *List {
	return &List{data, Word{0}}
}

func RyeToRaw(res Object) any {
	// fmt.Printf("Type: %T", res)
	switch v := res.(type) {
	case nil:
		return "null"
	case String:
		return v.Value
	case Integer:
		return v.Value
		// return strconv.Itoa(int(v.Value))
	case Decimal:
		return v.Value
		// return strconv.Itoa(int(v.Value))
	case Word:
		return "word"
	case Block:
		return v
	case List:
		return v
	case *List:
		return *v
	default:
		return "not handeled 2"
		// TODO-FIXME
	}
}

func NewListFromSeries(block TSeries) List {
	data := make([]any, block.Len())
	for block.Pos() < block.Len() {
		i := block.Pos()
		k1 := block.Pop()
		switch k := k1.(type) {
		case String:
			data[i] = k.Value
		case Integer:
			data[i] = k.Value
		case Decimal:
			data[i] = k.Value
		case List:
			data[i] = k
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
	bu.WriteString("[List (" + i.Kind.Print(idxs) + "): ")
	for _, v := range i.Data {
		switch ob := v.(type) {
		case map[string]any:
			vv := NewDict(ob)
			bu.WriteString(" " + vv.Inspect(idxs) + " ")
		case Object:
			bu.WriteString(" " + ob.Inspect(idxs) + " ")
		default:
			bu.WriteString(" " + fmt.Sprint(ob) + " ")
		}
	}
	bu.WriteString("]")
	return bu.String()
}

// Inspect returns a string representation of the Integer.
func (i List) Print(idxs Idxs) string {
	var bu strings.Builder
	bu.WriteString("L[")
	for _, v := range i.Data {
		switch ob := v.(type) {
		case map[string]any:
			vv := NewDict(ob)
			bu.WriteString(" " + vv.Print(idxs) + " ")
		case Object:
			bu.WriteString(" " + ob.Print(idxs) + " ")
		default:
			bu.WriteString(" " + fmt.Sprint(ob) + " ")
		}
	}
	bu.WriteString("]")
	return bu.String()
}

func (i List) Trace(msg string) {
	fmt.Print(msg + "(List): ")
}

func (i List) GetKind() int {
	return i.Kind.Index
}

func (i List) Equal(o Object) bool {
	if i.Type() != o.Type() {
		return false
	}
	oList := o.(List)
	if !i.Kind.Equal(oList.Kind) {
		return false
	}
	if len(i.Data) != len(oList.Data) {
		return false
	}
	for i, v := range i.Data {
		if vObj, ok := v.(Object); ok {
			if oObj, ok := oList.Data[i].(Object); ok {
				if !vObj.Equal(oObj) {
					return false
				}
			}
		} else {
			if v != oList.Data[i] {
				return false
			}
		}
	}
	return true
}

func (i List) Dump(e Idxs) string {
	// TODO
	return fmt.Sprintf("\"serlization of %s is not yet supported\" ", i.Inspect(e))
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
	// LONG DISPLAY OF FUNCTION NODES return "[Function: " + i.Spec.Inspect(e) + ", " + i.Body.Inspect(e) + "]"
	return "[Kind(" + i.Kind.Print(e) + "): " + i.Spec.Inspect(e) + "]"
}

// Inspect returns a string representation of the Integer.
func (i Kind) Print(e Idxs) string {
	return "[Kind: " + i.Spec.Inspect(e) + "]"
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

func (i Kind) Equal(o Object) bool {
	if i.Type() != o.Type() {
		return false
	}
	oKind := o.(Kind)
	if !i.Kind.Equal(oKind.Kind) {
		return false
	}
	if !i.Spec.Equal(oKind.Spec) {
		return false
	}
	if len(i.Converters) != len(oKind.Converters) {
		return false
	}
	for k, v := range i.Converters {
		if !v.Equal(oKind.Converters[k]) {
			return false
		}
	}
	return true
}

func (i Kind) Dump(e Idxs) string {
	return fmt.Sprintf("kind %s %s", i.Kind.Dump(e), i.Spec.Dump(e))
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
	// LONG DISPLAY OF FUNCTION NODES return "[Function: " + i.Spec.Inspect(e) + ", " + i.Body.Inspect(e) + "]"
	return "[Converter(" + i.From.Print(e) + "->" + i.To.Print(e) + "): " + i.Spec.Inspect(e) + "]"
}

// Inspect returns a string representation of the Integer.
func (i Converter) Print(e Idxs) string {
	return i.Spec.Inspect(e)
}

func (i Converter) Trace(msg string) {
	fmt.Print(msg + " (Converter): ")
	fmt.Println(i.Spec)
}

func (i Converter) GetKind() int {
	return int(ConverterType)
}

func (i Converter) Equal(o Object) bool {
	if i.Type() != o.Type() {
		return false
	}
	oConverter := o.(Converter)
	if !i.From.Equal(oConverter.From) {
		return false
	}
	if !i.To.Equal(oConverter.To) {
		return false
	}
	if !i.Spec.Equal(oConverter.Spec) {
		return false
	}
	return true
}

func (i Converter) Dump(e Idxs) string {
	// TODO
	return fmt.Sprintf("\"serlization of %s is not yet supported\" ", i.Inspect(e))
}

//
// TIME
//

type Time struct {
	Value time.Time
}

func NewTime(val time.Time) *Time {
	nat := Time{val}
	return &nat
}

// Type returns the type of the Integer.
func (i Time) Type() Type {
	return TimeType
}

// Inspect returns a string representation of the Integer.
func (i Time) Inspect(e Idxs) string {
	return "[Time: " + i.Print(e) + "]"
}

// Inspect returns a string representation of the Integer.
func (i Time) Print(e Idxs) string {
	return i.Value.Format("2006-01-02 15:04:05")
}

func (i Time) Trace(msg string) {
	fmt.Print(msg + "(time): ")
	fmt.Println(i.Value)
}

func (i Time) GetKind() int {
	return int(TimeType)
}

func (i Time) Equal(o Object) bool {
	if i.Type() != o.Type() {
		return false
	}
	return i.Value.Equal(o.(Time).Value)
}

func (i Time) Dump(e Idxs) string {
	return fmt.Sprintf("datetime \"%s\"", i.Value.Format("2006-01-02T15:04:05"))
}

//
// VECTOR TYPE -- feture vector (uses govector)
//

type Vector struct {
	Value govector.Vector
	Kind  Word
}

func NewVector(vec govector.Vector) *Vector {
	//vec, err := govector.AsVector(data)
	//if err != nil {
	//	return MakeError(env1, err.Error())
	//}
	return &Vector{vec, Word{0}}
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

// data []float32

func NewVectorFromSeries(block TSeries) *Vector {
	data := ArrayFloat32FromSeries(block)
	vec, err := govector.AsVector(data)
	if err != nil {
		return nil
	}
	return &Vector{vec, Word{0}}
}

// Type returns the type of the Integer.
func (i Vector) Type() Type {
	return VectorType
}

// Inspect returns a string representation of the Integer.
func (i Vector) Inspect(idxs Idxs) string {
	var bu strings.Builder
	bu.WriteString("[Vector:") //(" + i.Kind.Probe(idxs) + "):")
	bu.WriteString(" Len " + strconv.Itoa(i.Value.Len()))
	bu.WriteString(" Norm " + fmt.Sprintf("%.2f", govector.Norm(i.Value, 2.0)))
	bu.WriteString(" Mean " + fmt.Sprintf("%.2f", i.Value.Mean()))
	bu.WriteString("]")
	return bu.String()
}

// Inspect returns a string representation of the Integer.
func (i Vector) Print(idxs Idxs) string {
	var bu strings.Builder
	bu.WriteString("[")
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
	// TODO
	return fmt.Sprintf("\"serlization of %s is not yet supported\" ", i.Inspect(e))
}
