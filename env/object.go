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
	OpCPathType        Type = 36
	PipeCPathType      Type = 37
	ModwordType        Type = 38
	LModwordType       Type = 39
)

// after adding new type here, also add string to idxs.go

type Object interface {
	Type() Type
	Trace(msg string)
	GetKind() int
	Equal(p Object) bool
	// Print returns a string representation of the Object.
	Print(e Idxs) string
	// Inspect returns a diagnostic string representation of the Object.
	Inspect(e Idxs) string
	// Dump returns a string representation of the Object, intended for serialization.
	Dump(e Idxs) string
}

//
// INTEGER
//

type Integer struct {
	Value int64
}

func NewInteger(val int64) *Integer {
	nat := Integer{val}
	return &nat
}

func (i Integer) Type() Type {
	return IntegerType
}

func (i Integer) Inspect(e Idxs) string {
	return "[Integer: " + i.Print(e) + "]"
}

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

type Decimal struct {
	Value float64 `bson:"value"`
}

func NewDecimal(val float64) *Decimal {
	nat := Decimal{val}
	return &nat
}

func (i Decimal) Type() Type {
	return DecimalType
}

func (i Decimal) Inspect(e Idxs) string {
	return "[Decimal: " + i.Print(e) + "]"
}

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
		fmt.Println(i.Type())
		fmt.Println(o.Type())
		fmt.Println("TYPES")
		return false
	}
	fmt.Println(i.Value)
	fmt.Println(o.(Decimal).Value)
	return i.Value == o.(Decimal).Value
}

func (i Decimal) Dump(e Idxs) string {
	return strconv.FormatFloat(i.Value, 'f', -1, 64)
}

//
// STRING
//

type String struct {
	Value string `bson:"value"`
}

func NewString(val string) *String {
	nat := String{val}
	return &nat
}

func (i String) Type() Type {
	return StringType
}

func (i String) Inspect(e Idxs) string {
	return "[String: " + i.Print(e) + "]"
}

func (i String) Print(e Idxs) string {
	return i.Value
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
	return fmt.Sprintf("\"%s\"", i.Value)
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
	return "[Date: " + i.Print(e) + "]"
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

func (i Uri) GetFullUri(e Idxs) string {
	return e.GetWord(i.Scheme.Index) + "://" + i.Path
}

func (i Uri) GetProtocol() Word {
	return i.Scheme
}

func (i Uri) Type() Type {
	return UriType
}

func (i Uri) Inspect(e Idxs) string {
	return "[Uri: " + i.Print(e) + "]"
}

func (i Uri) Print(e Idxs) string {
	return i.Scheme.Print(e) + "://" + i.GetPath()
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
// EMAIL
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

func (i Email) Inspect(e Idxs) string {
	return "[Email: " + i.Print(e) + "]"
}

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

func (i Block) Type() Type {
	return BlockType
}

func (b Block) Inspect(e Idxs) string {
	var r strings.Builder
	r.WriteString("[Block: ")
	for i := 0; i < b.Series.Len(); i += 1 {
		if b.Series.Get(i) != nil {
			if b.Series.GetPos() == i {
				r.WriteString("^")
			}
			r.WriteString(b.Series.Get(i).Inspect(e))
			r.WriteString(" ")
		}
	}
	r.WriteString("]")
	return r.String()
}

func (b Block) Print(e Idxs) string {
	var r strings.Builder
	// r.WriteString("{ ")
	for i := 0; i < b.Series.Len(); i += 1 {
		if b.Series.Get(i) != nil {
			r.WriteString(b.Series.Get(i).Print(e))
			r.WriteString(" ")
		} else {
			r.WriteString("[NIL]")
		}
	}
	// r.WriteString("}")
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
			bu.WriteString(obj.Dump(e))
			bu.WriteString(" ")
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

type Word struct {
	Index int
}

func NewWord(val int) *Word {
	nat := Word{val}
	return &nat
}

func (i Word) Type() Type {
	return WordType
}

func (i Word) Inspect(e Idxs) string {
	return "[Word: " + i.Print(e) + "]"
}

func (i Word) Print(e Idxs) string {
	return e.GetWord(i.Index)
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

type Setword struct {
	Index int
}

func NewSetword(index int) *Setword {
	nat := Setword{index}
	return &nat
}

func (i Setword) Type() Type {
	return SetwordType
}

func (i Setword) Inspect(e Idxs) string {
	return "[Setword: " + e.GetWord(i.Index) + "]"
}

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
	return e.GetWord(i.Index) + ":"
}

//
// LSETWORD
//

type LSetword struct {
	Index int
}

func NewLSetword(index int) *LSetword {
	nat := LSetword{index}
	return &nat
}

func (i LSetword) Type() Type {
	return LSetwordType
}

func (i LSetword) Inspect(e Idxs) string {
	return "[LSetword: " + e.GetWord(i.Index) + "]"
}

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
	return ":" + e.GetWord(i.Index)
}

//
// MODWORD
//

type Modword struct {
	Index int
}

func NewModword(index int) *Modword {
	nat := Modword{index}
	return &nat
}

func (i Modword) Type() Type {
	return ModwordType
}

func (i Modword) Inspect(e Idxs) string {
	return "[Modword: " + e.GetWord(i.Index) + "]"
}

func (b Modword) Print(e Idxs) string {
	return e.GetWord(b.Index) + "::"
}

func (i Modword) Trace(msg string) {
	fmt.Print(msg + "(Modword): ")
	fmt.Println(i.Index)
}

func (i Modword) GetKind() int {
	return int(ModwordType)
}

func (i Modword) Equal(o Object) bool {
	if i.Type() != o.Type() {
		return false
	}
	return i.Index == o.(Modword).Index
}

func (i Modword) Dump(e Idxs) string {
	return e.GetWord(i.Index) + "::"
}

//
// LMODWORD
//

type LModword struct {
	Index int
}

func NewLModword(index int) *LModword {
	nat := LModword{index}
	return &nat
}

func (i LModword) Type() Type {
	return LModwordType
}

func (i LModword) Inspect(e Idxs) string {
	return "[LModword: " + e.GetWord(i.Index) + "]"
}

func (b LModword) Print(e Idxs) string {
	return "::" + e.GetWord(b.Index)
}

func (i LModword) Trace(msg string) {
	fmt.Print(msg + "(lModword): ")
	fmt.Println(i.Index)
}

func (i LModword) GetKind() int {
	return int(LModwordType)
}

func (i LModword) Equal(o Object) bool {
	if i.Type() != o.Type() {
		return false
	}
	return i.Index == o.(LModword).Index
}

func (i LModword) Dump(e Idxs) string {
	return ":" + e.GetWord(i.Index)
}

//
// OPWORD
//

type Opword struct {
	Index int
	Force int
}

func NewOpword(index, force int) *Opword {
	nat := Opword{index, force}
	return &nat
}

func (i Opword) Type() Type {
	return OpwordType
}

func (i Opword) Inspect(e Idxs) string {
	return "[Opword: " + e.GetWord(i.Index) + "]"
}

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
	return "." + e.GetWord(i.Index)
}

//
// PIPEWORD
//

type Pipeword struct {
	Index int
	Force int
}

func NewPipeword(index, force int) *Pipeword {
	nat := Pipeword{index, force}
	return &nat
}

func (i Pipeword) Type() Type {
	return PipewordType
}

func (i Pipeword) Inspect(e Idxs) string {
	return "[Pipeword: " + e.GetWord(i.Index) + "]"
}

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
	return "|" + e.GetWord(i.Index)
}

//
// TAGWORD
//

type Tagword struct {
	Index int
}

func NewTagword(index int) *Tagword {
	nat := Tagword{index}
	return &nat
}

func (i Tagword) Type() Type {
	return TagwordType
}

func (i Tagword) Inspect(e Idxs) string {
	return "[Tagword: " + e.GetWord(i.Index) + "]"
}

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
	return "'" + e.GetWord(i.Index)
}

//
// XWORD
//

type Xword struct {
	Index int
	Args  string
}

func NewXword(index int, args string) *Xword {
	nat := Xword{index, args}
	return &nat
}

func (i Xword) Type() Type {
	return XwordType
}

// + strconv.FormatInt(int64(i.Index), 10) +
func (i Xword) Inspect(e Idxs) string {
	return "[Xword: " + e.GetWord(i.Index) + "]"
}

func (b Xword) Print(e Idxs) string {
	spc := ""
	if len(b.Args) > 0 {
		spc = " "
	}
	return "<" + e.GetWord(b.Index) + spc + b.Args + ">"
}

func (i Xword) Trace(msg string) {
	fmt.Print(msg + " (Xword): ")
	fmt.Println(i.Index)
}

func (i Xword) ToWord() Word {
	return *NewWord(i.Index)
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
	return "<" + e.GetWord(i.Index) + ">"
}

//
// EXWORD
//

type EXword struct {
	Index int
}

func NewEXword(index int) *EXword {
	nat := EXword{index}
	return &nat
}

func (i EXword) Type() Type {
	return EXwordType
}

func (i EXword) Inspect(e Idxs) string {
	return "[EXword: " + e.GetWord(i.Index) + "]"
}

func (b EXword) Print(e Idxs) string {
	return "</" + e.GetWord(b.Index) + ">"
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
	return "</" + e.GetWord(i.Index) + ">"
}

//
// KINDWORD
//

type Kindword struct {
	Index int
}

func NewKindword(index int) *Kindword {
	nat := Kindword{index}
	return &nat
}

func (i Kindword) Type() Type {
	return KindwordType
}

func (i Kindword) Inspect(e Idxs) string {
	return "[Kindword: " + e.GetWord(i.Index) + "]"
}

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
	return "~(" + e.GetWord(i.Index) + ")"
}

//
// GETWORD
//

type Getword struct {
	Index int
}

func NewGetword(index int) *Getword {
	nat := Getword{index}
	return &nat
}

func (i Getword) Type() Type {
	return GetwordType
}

func (i Getword) Inspect(e Idxs) string {
	return "[Getword: " + e.GetWord(i.Index) + "]"
}

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
	return "?" + e.GetWord(i.Index)
}

//
// GENWORD
//

type Genword struct {
	Index int
}

func NewGenword(index int) *Genword {
	nat := Genword{index}
	return &nat
}

func (i Genword) Type() Type {
	return GenwordType
}

func (i Genword) Inspect(e Idxs) string {
	return "[Genword: " + e.GetWord(i.Index) + "]"
}

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
	return "~" + e.GetWord(i.Index)
}

//
// COMMA
//

type Comma struct{}

func (i Comma) Type() Type {
	return CommaType
}

func (i Comma) Inspect(e Idxs) string {
	return "[Comma]"
}

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

type Void struct{}

func (i Void) Type() Type {
	return VoidType
}

func (i Void) Inspect(e Idxs) string {
	return "[Void]"
}

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
// FUNCTION
//

type Function struct {
	Argsn int
	Spec  Block
	Body  Block
	Ctx   *RyeCtx
	Pure  bool
	Doc   string
	InCtx bool
}

func NewFunction(spec Block, body Block, pure bool) *Function {
	o := Function{spec.Series.Len(), spec, body, nil, pure, "", false}
	return &o
}

func NewFunctionC(spec Block, body Block, ctx *RyeCtx, pure bool, inCtx bool, doc string) *Function {
	var argn int
	if doc > "" {
		argn = spec.Series.Len() - 1
	} else {
		argn = spec.Series.Len()
	}
	o := Function{argn, spec, body, ctx, pure, doc, inCtx}
	return &o
}

func NewFunctionDoc(spec Block, body Block, pure bool, doc string) *Function {
	var argn int
	if doc > "" {
		argn = spec.Series.Len() - 1
	} else {
		argn = spec.Series.Len()
	}
	o := Function{argn, spec, body, nil, pure, doc, false}
	return &o
}

func (i Function) Type() Type {
	return FunctionType
}

func (i Function) Inspect(e Idxs) string {
	// LONG DISPLAY OF FUNCTION NODES return "[Function: " + i.Spec.Inspect(e) + ", " + i.Body.Inspect(e) + "]"
	var pure string
	if i.Pure {
		pure = "Pure "
	}
	docs := ""
	if len(i.Doc) > 0 {
		docs = ": " + i.Doc
	}
	return "[" + pure + "Function(" + strconv.FormatInt(int64(i.Argsn), 10) + ")" + docs + "]"
}

func (i Function) Print(e Idxs) string {
	var pure string
	if i.Pure {
		pure = "Pure "
	}
	return "[" + pure + "Function(" + strconv.FormatInt(int64(i.Argsn), 10) + ")]"
}

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
	var b strings.Builder
	b.WriteString("fn { ")
	for _, obj := range i.Spec.Series.GetAll() {
		if obj != nil {
			b.WriteString(obj.Dump(e))
			b.WriteString(" ")
		} else {
			b.WriteString("'nil ")
		}
	}
	b.WriteString("} { ")
	for _, obj := range i.Body.Series.GetAll() {
		if obj != nil {
			b.WriteString(obj.Dump(e))
			b.WriteString(" ")
		} else {
			b.WriteString("'nil ")
		}
	}
	b.WriteString("}")

	return b.String()
}

//
// BUILTIN FUNCTION
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

func (b Builtin) Type() Type {
	return BuiltinType
}

func (b Builtin) Inspect(e Idxs) string {
	return "[" + b.Print(e) + "]"
}

func (b Builtin) Print(e Idxs) string {
	var pure string
	if b.Pure {
		pure = "Pure "
	}
	return pure + "BFunction(" + strconv.Itoa(b.Argsn) + "): " + b.Doc
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
	// Serializing builtins is not supported
	return ""
}

//
// ERROR
//

type Error struct {
	Status      int
	Message     string
	Parent      *Error
	Values      map[string]Object
	CodeContext *RyeCtx
	CodeBlock   TSeries
}

func (i Error) Type() Type {
	return ErrorType
}

func (i Error) Inspect(e Idxs) string {
	return "[" + i.Print(e) + "]"
}

func (i Error) Print(e Idxs) string {
	return i.Print2(e, 1)
}

func (i Error) Print2(e Idxs, depth int) string {
	status := ""
	if i.Status != 0 {
		status = "(" + strconv.Itoa(i.Status) + ")"
	}
	var b strings.Builder
	b.WriteString("Error" + status + ": " + i.Message + " ")
	if i.Parent != nil {
		b.WriteString("\n")
		for i := 0; i < depth; i++ {
			b.WriteString("  ")
		}
		b.WriteString(i.Parent.Print2(e, depth+1))
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
	if (i.Parent == nil) != (oError.Parent == nil) {
		return false
	}
	if i.Parent != nil && oError.Parent != nil {
		if !i.Parent.Equal(oError.Parent) {
			return false
		}
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
	if i.Parent == nil {
		return fmt.Sprintf("failure { %d \"%s\" }", i.Status, i.Message)
	} else {
		return fmt.Sprintf("wrap\\failure { %d \"%s\" } %s", i.Status, i.Message, i.Parent.Dump(e))
	}
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

func (i Argword) Type() Type {
	return ArgwordType
}

func (i Argword) Inspect(e Idxs) string {
	return "[Argword: " + i.Name.Inspect(e) + ":" + i.Kind.Inspect(e) + "]"
}

func (b Argword) Print(e Idxs) string {
	return b.Name.Print(e)
}

func (i Argword) Trace(msg string) {
	fmt.Print(msg + " (argword): ")
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
	Mode  int // 0 Cpath, 1 OpCpath , 2 PipeCPath
	Cnt   int
	Word1 Word
	Word2 Word
	Word3 Word
}

func (i CPath) Type() Type {
	return CPathType
}

func (i CPath) Inspect(e Idxs) string {
	switch i.Cnt {
	case 2:
		return "[CPath: " + i.Word1.Inspect(e) + "/" + i.Word2.Inspect(e) + "]"
	case 3:
		return "[CPath: " + i.Word1.Inspect(e) + "/" + i.Word2.Inspect(e) + "/" + i.Word3.Inspect(e) + "]"
	}
	return "[CPath: " + i.Word1.Inspect(e) + "/ ... ]"
}

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

func (b CPath) Print(e Idxs) string {
	return b.Word1.Print(e)
}

func (i CPath) Trace(msg string) {
	fmt.Print(msg + " (cpath): ")
}

func (i CPath) GetKind() int {
	return int(CPathType)
}

func NewCPath2(mode int, w1 Word, w2 Word) *CPath {
	var cp CPath
	cp.Mode = mode
	cp.Cnt = 2
	cp.Word1 = w1
	cp.Word2 = w2
	return &cp
}
func NewCPath3(mode int, w1 Word, w2 Word, w3 Word) *CPath {
	var cp CPath
	cp.Mode = mode
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
	var b strings.Builder
	b.WriteString(i.Word1.Dump(e))
	if i.Cnt > 1 {
		b.WriteString("/" + i.Word2.Dump(e))
	}
	if i.Cnt > 2 {
		b.WriteString("/" + i.Word3.Dump(e))
	}
	return b.String()
}

//
// NATIVE
//

type Native struct {
	Value any
	Kind  Word
}

func NewNative(index *Idxs, val any, kind string) *Native {
	idx := index.IndexWord(kind)
	nat := Native{val, Word{idx}}
	return &nat
}

func (i Native) Type() Type {
	return NativeType
}

func (i Native) Inspect(e Idxs) string {
	return "[" + i.Print(e) + "]"
}

func (i Native) Print(e Idxs) string {
	return "Native of kind " + i.Kind.Print(e) + ""
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
	// Serializing natives is not supported
	return ""
}

//
// DICT
//

// Dict -- nonindexed and unboxed map ... for example for params from request etc, so we don't neet to idex keys and it doesn't need boxed values
// I think it should have option of having Kind too ...
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

func (i Dict) Type() Type {
	return DictType
}

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
// LIST
//

// List -- nonindexed and unboxed list (block)
type List struct {
	Data []any
	Kind Word
}

func NewList(data []any) *List {
	return &List{data, Word{0}}
}

func RyeToRaw(res Object) any { // TODO -- MOVE TO UTIL ... provide reverse named symmetrically
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
		k1 := block.Pop() // TODO -- USE RyeToRaw
		switch k := k1.(type) {
		case String:
			data[i] = k.Value
		case Integer:
			data[i] = k.Value
		case Decimal:
			data[i] = k.Value
		case List:
			data[i] = k
		case Dict:
			data[i] = k
		}
	}
	return List{data, Word{0}}
}

func NewBlockFromList(list List) TSeries {
	data := make([]Object, len(list.Data))
	for i, v := range list.Data {
		switch k := v.(type) {
		case string:
			data[i] = *NewString(k)
		case int64:
			data[i] = *NewInteger(k)
		case float64:
			data[i] = *NewDecimal(k)
		case List:
			data[i] = *NewString("not handeled 3") // TODO -- just temp result
		}
	}
	return *NewTSeries(data)
}

func (i List) Type() Type {
	return ListType
}

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
	var b strings.Builder
	b.WriteString("list { ")
	for _, v := range i.Data {
		b.WriteString(ToRyeValue(v).Dump(e) + " ")
	}
	b.WriteString("}")
	return b.String()
}

//
// KIND
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

func (i Kind) Inspect(e Idxs) string {
	// LONG DISPLAY OF FUNCTION NODES return "[Function: " + i.Spec.Inspect(e) + ", " + i.Body.Inspect(e) + "]"
	return "[Kind(" + i.Kind.Print(e) + "): " + i.Spec.Inspect(e) + "]"
}

func (i Kind) Print(e Idxs) string {
	return "Kind: " + i.Spec.Inspect(e)
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
// CONVERTER
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

func (i Converter) Inspect(e Idxs) string {
	// LONG DISPLAY OF FUNCTION NODES return "[Function: " + i.Spec.Inspect(e) + ", " + i.Body.Inspect(e) + "]"
	return "[" + i.Print(e) + "]"
}

func (i Converter) Print(e Idxs) string {
	return "Converter(" + i.From.Print(e) + "->" + i.To.Print(e) + "): " + i.Spec.Inspect(e)
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
	// Serializing converters is not supported
	return ""
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

func (i Time) Type() Type {
	return TimeType
}

func (i Time) Inspect(e Idxs) string {
	return "[Time: " + i.Print(e) + "]"
}

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
// VECTOR
//

// Vector -- feture vector (uses govector)
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

func NewVectorFromSeries(block TSeries) *Vector {
	data := ArrayFloat32FromSeries(block)
	vec, err := govector.AsVector(data)
	if err != nil {
		return nil
	}
	return &Vector{vec, Word{0}}
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
