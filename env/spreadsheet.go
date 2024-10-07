package env

import (
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"strings"
)

func makeError(env1 *ProgramState, msg string) *Error {
	env1.FailureFlag = true
	return NewError(msg)
}

type SpreadsheetRow struct {
	Values []any
	Uplink *Spreadsheet
}

func NewSpreadsheetRow(values []any, uplink *Spreadsheet) *SpreadsheetRow {
	nat := SpreadsheetRow{values, uplink}
	return &nat
}

func SpreadsheetRowFromDict(dict Dict, uplink *Spreadsheet) (bool, string, *SpreadsheetRow) {
	row := SpreadsheetRow{make([]any, len(uplink.Cols)), uplink}
	for i, v := range uplink.Cols {
		if val, ok := dict.Data[v]; ok {
			row.Values[i] = val
		} else {
			return false, v, nil
		}
	}
	return true, "", &row
}

type Spreadsheet struct {
	Cols    []string
	Rows    []SpreadsheetRow
	Kind    Word
	Indexes map[string]map[any][]int
}

func NewSpreadsheet(cols []string) *Spreadsheet {
	var ps Spreadsheet
	ps.Cols = cols
	ps.Rows = make([]SpreadsheetRow, 0)
	/*
		ps := Spreadsheet{
			cols,
			make([]SpreadsheetRow, 1)
		} */
	return &ps
}

// Inspect returns a string representation of the Integer.
func (s *Spreadsheet) AddRow(vals SpreadsheetRow) {
	s.Rows = append(s.Rows, vals)
}

func (s *Spreadsheet) RemoveRowByIndex(index int64) {
	s.Rows = append(s.Rows[:index], s.Rows[index+1:]...)
}

func (s *Spreadsheet) GetRows() []SpreadsheetRow {
	return s.Rows
}

// Inspect returns a string representation of the Integer.
func (s *Spreadsheet) SetCols(vals []string) {
	s.Cols = vals
}

// Inspect returns a string representation of the Integer.
func (s Spreadsheet) ToHtml() string {
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
func (s Spreadsheet) ToTxt() string {
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

func (s Spreadsheet) Column(name string) Object {
	col1 := make([]Object, len(s.Rows))
	idx := slices.Index[[]string](s.Cols, name)
	if idx > -1 {
		for i, row := range s.Rows {
			col1[i] = ToRyeValue(row.Values[idx])
		}
		return *NewBlock(*NewTSeries(col1))
	} else {
		return *NewError("Column not found")
	}
}

func (s Spreadsheet) Sum(name string) Object {
	var sum int64
	var sumf float64
	idx := slices.Index[[]string](s.Cols, name)
	if idx > -1 {
		for _, row := range s.Rows {
			if len(row.Values) > idx {
				switch v := row.Values[idx].(type) {
				case int64:
					sum += v
				case Integer:
					sum += v.Value
				case Decimal:
					sumf += v.Value
				default:
					fmt.Println("row--->")
					fmt.Println(reflect.TypeOf(v))
				}
			}
		}
		if sumf == 0 {
			return Integer{sum}
		} else {
			return Decimal{sumf + float64(sum)}
		}
		//return sumf + float64(sum), nil
		//return Integer{int64(sum)}
	} else {
		return *NewError("Column not found")
	}
}

func (s Spreadsheet) Sum_Just(name string) (float64, error) {
	var sum int64
	var sumf float64
	idx := slices.Index[[]string](s.Cols, name)
	if idx > -1 {
		for _, row := range s.Rows {
			if len(row.Values) > idx {
				switch v := row.Values[idx].(type) {
				case float64:
					sumf += v
				case int64:
					sum += v
				case Integer:
					sum += v.Value
				case Decimal:
					sumf += v.Value
				default:
					fmt.Println("row--->")
					fmt.Println(reflect.TypeOf(v))
				}
			}
		}
		// if sumf == 0 {
		//	return Integer{int64(sum)}
		//} else {
		//	return Decimal{sumf + float64(sum)}
		//}
		return sumf + float64(sum), nil
	} else {
		return 0.0, errors.New("Column not found")
	}
}

func (s Spreadsheet) NRows() int {
	return len(s.Rows)
}

func (s Spreadsheet) Columns(ps *ProgramState, names []string) Object {
	idxs := make([]int, len(names))
	for name := range names {
		idx := slices.Index[[]string](s.Cols, names[name])
		if idx == -1 {
			return makeError(ps, "Col not found")
		}
		idxs[name] = idx
	}
	nspr := NewSpreadsheet(names)

	for _, row := range s.Rows {
		row2 := make([]any, len(names))
		for col := range idxs {
			if len(row.Values) > col {
				//				row2[col] = row.Values[idxs[col]].(Object)
				row2[col] = row.Values[idxs[col]]
			}
		}
		nspr.AddRow(SpreadsheetRow{row2, nspr})
	}
	//nspr.(res)
	return *nspr
}

func (s Spreadsheet) GetRow(ps *ProgramState, index int) SpreadsheetRow {
	row := s.Rows[index]
	row.Uplink = &s
	return row
}

func (s Spreadsheet) GetRowNew(index int) Object {
	row := s.Rows[index]
	row.Uplink = &s
	return row
}

func (s Spreadsheet) GetRowValue(column string, rrow SpreadsheetRow) (any, error) {
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
func (s Spreadsheet) Type() Type {
	return SpreadsheetType
}

// Inspect returns a string
func (s Spreadsheet) Inspect(e Idxs) string {
	rows := strconv.Itoa(len(s.Rows))
	var kindStr string
	//fmt.Println(s.GetKind())
	if s.GetKind() != int(SpreadsheetType) {
		kindStr = " of kind " + s.Kind.Print(e)
	}
	return "[Spreadsheet(" + strconv.Itoa(len(s.Cols)) + " " + rows + ")" + kindStr + "]"
}

// Inspect returns a string representation of the Integer.
func (s Spreadsheet) Print(e Idxs) string {
	return s.ToTxt()
}

func (s Spreadsheet) Trace(msg string) {
	fmt.Print(msg + " (spreadsheet): ")
}

func (s Spreadsheet) GetKind() int {
	return int(SpreadsheetType)
}

func (s Spreadsheet) Equal(o Object) bool {
	if s.Type() != o.Type() {
		return false
	}
	oSpr := o.(Spreadsheet)
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
			if vObj, ok := v.(Object); ok {
				if oObj, ok := o.(Object); ok {
					if !vObj.Equal(oObj) {
						return false
					}
				}
			} else {
				if v != o {
					return false
				}
			}
		}
	}
	return true
}

func (s Spreadsheet) Dump(e Idxs) string {
	// TODO
	return fmt.Sprintf("\"serlization of %s is not yet supported\" ", s.Inspect(e))
}

func (s SpreadsheetRow) GetKind() int {
	return int(0)
}

// Inspect returns a string
func (s SpreadsheetRow) Inspect(e Idxs) string {
	p := ""
	if IsPointer(s) {
		p = "REF:"
	}
	return "[" + p + "SpreadsheetRow(" + strconv.Itoa(len(s.Values)) + ")" + "]"
}

// Inspect returns a string representation of the Integer.
func (s SpreadsheetRow) Print(e Idxs) string {
	return s.ToTxt()
}

// Do not use when comparing a spreadsheet as a whole
// because column ordering is not guaranteed
func (s SpreadsheetRow) Equal(o Object) bool {
	if s.Type() != o.Type() {
		return false
	}
	oSprRow := o.(SpreadsheetRow)
	if len(s.Values) != len(oSprRow.Values) {
		return false
	}
	for i, v := range s.Values {
		if vObj, ok := v.(Object); ok {
			if oObj, ok := oSprRow.Values[i].(Object); ok {
				if !vObj.Equal(oObj) {
					return false
				}
			}
		} else {
			if v != oSprRow.Values[i] {
				return false
			}
		}
	}

	return true
}

func (s SpreadsheetRow) Dump(e Idxs) string {
	// TODO
	return fmt.Sprintf("\"serlization of %s is not yet supported\" ", s.Inspect(e))
}

func (s SpreadsheetRow) ToTxt() string {
	var bu strings.Builder
	bu.WriteString("[ ")
	for i, val := range s.Values {
		bu.WriteString(s.Uplink.Cols[i])
		bu.WriteString(": ")
		bu.WriteString(fmt.Sprint(val))
		bu.WriteString("\t")
	}
	bu.WriteString(" ]")
	return bu.String()
}

func (s SpreadsheetRow) Trace(msg string) {
	fmt.Print(msg + " (spreadsheet): ")
}

// Type returns the type of the Integer.
func (s SpreadsheetRow) Type() Type {
	return SpreadsheetRowType
}

func (s SpreadsheetRow) ToDict() Dict {
	ser := make([]string, len(s.Values))
	for i, v := range s.Values {
		ser[i] = fmt.Sprint(v)
	}
	d := NewDict(map[string]any{})
	for i, v := range s.Values {
		d.Data[s.Uplink.Cols[i]] = v
	}
	return *d
}

func (s Spreadsheet) GetColumns() List {
	lst := make([]any, len(s.Cols))
	for i, v := range s.Cols {
		lst[i] = v
	}
	return *NewList(lst)
}
