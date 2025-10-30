package env

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
)

// TableInterface defines the common interface for all table implementations
type TableInterface interface {
	// Core operations
	AddRow(row TableRow)
	GetRows() []TableRow
	GetRow(ps *ProgramState, index int) TableRow
	RemoveRowByIndex(index int64)

	// Column operations
	GetColumn(name string) Object
	GetColumns() List
	GetColumnIndex(column string) int
	GetColumnNames() []string
	SetCols(vals []string)

	// Query operations
	GetRowValue(column string, rrow TableRow) (any, error)
	Columns(ps *ProgramState, names []string) Object

	// Metadata
	Length() int
	NRows() int
	Type() Type
	GetKind() int
	Equal(o Object) bool

	// Display/serialization
	Inspect(e Idxs) string
	Print(e Idxs) string
	ToHtml() string
	ToTxt() string
	Dump(e Idxs) string
	Trace(msg string)

	// Collections interface
	Get(i int) Object
	MakeNew(data []Object) Object
}

func makeError(env1 *ProgramState, msg string) *Error {
	env1.FailureFlag = true
	return NewError(msg)
}

type TableRow struct {
	Values []any
	Uplink TableInterface
}

func NewTableRow(values []any, uplink TableInterface) *TableRow {
	nat := TableRow{values, uplink}
	return &nat
}

func AddTableRowAndBlock(row TableRow, updatesBlock TSeries, idx *Idxs) TableRow {
	data := make(map[string]any)

	for updatesBlock.Pos() < updatesBlock.Len() {
		key := updatesBlock.Pop()
		val := updatesBlock.Pop()
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

	return AddTableRowAndMap(row, data)
}

func AddTableRowAndDict(row TableRow, dict Dict) TableRow {
	return AddTableRowAndMap(row, dict.Data)
}

func AddTableRowAndMap(row TableRow, dict map[string]any) TableRow {
	newRow := TableRow{row.Values, row.Uplink}

	for i, v := range row.Uplink.GetColumnNames() {
		if val, ok := dict[v]; ok {
			newRow.Values[i] = val
		}
	}
	return newRow
}

func TableRowFromDict(dict Dict, uplink *Table) (bool, string, *TableRow) {
	row := TableRow{make([]any, len(uplink.Cols)), uplink}
	for i, v := range uplink.Cols {
		if val, ok := dict.Data[v]; ok {
			row.Values[i] = val
		} else {
			return false, v, nil
		}
	}
	return true, "", &row
}

type Table struct {
	Cols    []string
	Rows    []TableRow
	Kind    Word
	Indexes map[string]map[any][]int
}

func NewTable(cols []string) *Table {
	var ps Table
	ps.Cols = cols
	ps.Rows = make([]TableRow, 0)
	/*
		ps := Table{
			cols,
			make([]TableRow, 1)
		} */
	return &ps
}

// Inspect returns a string representation of the Integer.
func (s *Table) AddRow(vals TableRow) {
	s.Rows = append(s.Rows, vals)
}

func (s *Table) RemoveRowByIndex(index int64) {
	s.Rows = append(s.Rows[:index], s.Rows[index+1:]...)
}

func (s *Table) GetRows() []TableRow {
	return s.Rows
}

// Inspect returns a string representation of the Integer.
func (s *Table) SetCols(vals []string) {
	s.Cols = vals
}

// Inspect returns a string representation of the Integer.
func (s Table) ToHtml() string {
	//fmt.Println("IN TO Html")
	var bu strings.Builder
	bu.WriteString("<table>")
	for _, row := range s.Rows {
		bu.WriteString("<tr>")
		for _, val := range row.Values {
			bu.WriteString("<td>")
			bu.WriteString(fmt.Sprint(val))
			bu.WriteString("</td>")
		}
		bu.WriteString("</tr>")
	}
	bu.WriteString("</table>")
	//fmt.Println(bu.String())
	return bu.String()
}

// Inspect returns a string representation of the Integer.
func (s Table) ToTxt() string {
	var bu strings.Builder
	for _, name := range s.Cols {
		bu.WriteString(fmt.Sprint(name))
		bu.WriteString("\t|")
	}
	bu.WriteString("\n")
	for _, row := range s.Rows {
		for _, val := range row.Values {
			bu.WriteString(fmt.Sprint(val))
			bu.WriteString("\t|")
		}
		bu.WriteString("\n")
	}
	//fmt.Println(bu.String())
	return bu.String()
}

func (s Table) GetColumn(name string) Object {
	col1 := make([]Object, len(s.Rows))
	idx := slices.Index[[]string](s.Cols, name)
	if idx > -1 {
		for i, row := range s.Rows {
			col1[i] = ToRyeValue(row.Values[idx])
		}
		return *NewBlock(*NewTSeries(col1))
	} else {
		return NewError("Column not found")
	}
}

func (s Table) NRows() int {
	return len(s.Rows)
}

func (s Table) Columns(ps *ProgramState, names []string) Object {
	idxs := make([]int, len(names))
	for name := range names {
		idx := slices.Index[[]string](s.Cols, names[name])
		if idx == -1 {
			return makeError(ps, "Col not found")
		}
		idxs[name] = idx
	}
	nspr := NewTable(names)

	for _, row := range s.Rows {
		row2 := make([]any, len(names))
		for col := range idxs {
			if len(row.Values) > col {
				//				row2[col] = row.Values[idxs[col]].(Object)
				row2[col] = row.Values[idxs[col]]
			}
		}
		nspr.AddRow(TableRow{row2, nspr})
	}
	//nspr.(res)
	return *nspr
}

func (s Table) ColumnsRenamed(ps *ProgramState, originalNames []string, newNames []string) Object {
	if len(originalNames) != len(newNames) {
		return makeError(ps, "Original and new column names must have same length")
	}

	idxs := make([]int, len(originalNames))
	for i, name := range originalNames {
		idx := slices.Index[[]string](s.Cols, name)
		if idx == -1 {
			return makeError(ps, "Col not found: "+name)
		}
		idxs[i] = idx
	}
	nspr := NewTable(newNames)

	for _, row := range s.Rows {
		row2 := make([]any, len(originalNames))
		for col := range idxs {
			if len(row.Values) > idxs[col] {
				row2[col] = row.Values[idxs[col]]
			}
		}
		nspr.AddRow(TableRow{row2, nspr})
	}
	return *nspr
}

func (s Table) GetRow(ps *ProgramState, index int) TableRow {
	row := s.Rows[index]
	row.Uplink = &s
	return row
}

func (s Table) GetRowNew(index int) Object {
	row := s.Rows[index]
	row.Uplink = &s
	return row
}

func (s Table) GetRowValue(column string, rrow TableRow) (any, error) {
	index := -1
	for i, v := range s.Cols {
		if v == column {
			index = i
			break
		}
	}
	if index < 0 {
		return "", fmt.Errorf("column %s not found", column)
	}
	return rrow.Values[index], nil
}

// Type returns the type of the Integer.
func (s Table) Type() Type {
	return TableType
}

// Inspect returns a string
func (s Table) Inspect(e Idxs) string {
	rows := strconv.Itoa(len(s.Rows))
	var kindStr string
	//fmt.Println(s.GetKind())
	if s.GetKind() != int(TableType) {
		kindStr = " of kind " + s.Kind.Print(e)
	}
	return "[Table(" + rows + " " + strconv.Itoa(len(s.Cols)) + ")" + kindStr + "]"
}

// Inspect returns a string representation of the Integer.
func (s Table) Print(e Idxs) string {
	return s.ToTxt()
}

func (s Table) Trace(msg string) {
	fmt.Print(msg + " (table): ")
}

func (s Table) GetKind() int {
	return int(TableType)
}

func (s Table) Equal(o Object) bool {
	if s.Type() != o.Type() {
		return false
	}
	oSpr := o.(Table)
	if len(s.Cols) != len(oSpr.Cols) {
		return false
	}
	columnMapping := make(map[int]int, len(s.Cols))
	for i, v := range s.Cols {
		idx := slices.Index[[]string](oSpr.Cols, v)
		if idx == -1 {
			return false
		}
		columnMapping[i] = idx
	}
	if len(s.Rows) != len(oSpr.Rows) {
		return false
	}
	for i, row := range s.Rows {
		for j, v := range row.Values {
			o := oSpr.Rows[i].Values[columnMapping[j]]
			vObj := ToRyeValue(v)
			oObj := ToRyeValue(o)
			if !vObj.Equal(oObj) {
				return false
			}
		}
	}
	return true
}

func (s Table) Dump(e Idxs) string {
	var sb strings.Builder
	sb.WriteString("table {")

	for _, col := range s.Cols {
		sb.WriteString(" ")
		sb.WriteString("\"")
		sb.WriteString(col)
		sb.WriteString("\"")
	}
	sb.WriteString(" } [")

	for _, row := range s.Rows {
		for _, val := range row.Values {
			sb.WriteString(" ")
			ryeVal := ToRyeValue(val)
			if ryeVal != nil {
				sb.WriteString(ryeVal.Dump(e))
			} else {
				sb.WriteString("_")
			}
		}
		// Fill in missing columns (if they exist) with void (_)
		for i := len(row.Values); i < len(s.Cols); i++ {
			sb.WriteString(" _")
		}
	}
	sb.WriteString(" ]")
	return sb.String()
}

func (s TableRow) GetKind() int {
	return int(0)
}

// Inspect returns a string
func (s TableRow) Inspect(e Idxs) string {
	return "[TableRow(" + strconv.Itoa(len(s.Values)) + ")" + "]"
}

// Inspect returns a string representation of the Integer.
func (s TableRow) Print(e Idxs) string {
	return s.ToTxt()
}

// Do not use when comparing a table as a whole
// because column ordering is not guaranteed
func (s TableRow) Equal(o Object) bool {
	if s.Type() != o.Type() {
		return false
	}
	oSprRow := o.(TableRow)
	if len(s.Values) != len(oSprRow.Values) {
		return false
	}
	for i, v := range s.Values {
		vObj := ToRyeValue(v)
		oObj := ToRyeValue(oSprRow.Values[i])
		if !vObj.Equal(oObj) {
			return false
		}
	}

	return true
}

func (s TableRow) Dump(e Idxs) string {
	// TODO
	return fmt.Sprintf("\"serlization of %s is not yet supported\" ", s.Inspect(e))
}

func (s TableRow) ToTxt() string {
	var bu strings.Builder
	bu.WriteString("[ ")
	for i, val := range s.Values {
		bu.WriteString(s.Uplink.GetColumnNames()[i])
		bu.WriteString(": ")
		bu.WriteString(fmt.Sprint(val))
		bu.WriteString("\t")
	}
	bu.WriteString(" ]")
	return bu.String()
}

func (s TableRow) Trace(msg string) {
	fmt.Print(msg + " (table): ")
}

// Type returns the type of the Integer.
func (s TableRow) Type() Type {
	return TableRowType
}

func (s TableRow) ToDict() Dict {
	ser := make([]string, len(s.Values))
	for i, v := range s.Values {
		ser[i] = fmt.Sprint(v)
	}
	d := NewDict(map[string]any{})
	for i, v := range s.Values {
		d.Data[s.Uplink.GetColumnNames()[i]] = v
	}
	return *d
}

func (s Table) GetColumns() List {
	lst := make([]any, len(s.Cols))
	for i, v := range s.Cols {
		lst[i] = v
	}
	return *NewList(lst)
}

func (s Table) GetColumnIndex(column string) int {
	index := -1
	for i, v := range s.Cols {
		if v == column {
			index = i
			break
		}
	}
	return index
}

func (s Table) GetColumnNames() []string {
	return s.Cols
}

// Collections

func (o Table) Length() int {
	return len(o.Rows)
}

func (o Table) Get(i int) Object {
	return o.Rows[i]
}

func (o Table) MakeNew(data []Object) Object {
	return *NewBlock(*NewTSeries(data))
}
