package env

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

func makeError(env1 *ProgramState, msg string) *Error {
	env1.FailureFlag = true
	return NewError(msg)
}

type SpreadsheetRow struct {
	Values []interface{}
	Uplink *Spreadsheet
}

type Spreadsheet struct {
	Cols      []string
	Rows      []SpreadsheetRow
	RawRows   [][]string
	Kind      Word
	RawMode   bool
	Index     map[string][]int
	IndexName string
}

func NewSpreadsheet(cols []string) *Spreadsheet {
	var ps Spreadsheet
	ps.Cols = cols
	ps.Rows = make([]SpreadsheetRow, 0)
	ps.RawMode = false
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

func (s *Spreadsheet) SetRaw(vals [][]string) {
	s.RawRows = vals
	s.RawMode = true
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
	col1 := make([]Object, len(s.Cols))
	idx := IndexOfString(name, s.Cols)
	if idx > -1 {
		for i, row := range s.Rows {
			switch v := row.Values[idx].(type) {
			case int:
				col1[i] = Integer{int64(v)}
			case Integer:
				col1[i] = v
			}
		}
		return *NewBlock(*NewTSeries(col1))
	} else {
		return *NewError("Column not found")
	}
}

func (s Spreadsheet) Sum(name string) Object {
	var sum int64
	idx := IndexOfString(name, s.Cols)
	if idx > -1 {
		for _, row := range s.Rows {
			if len(row.Values) > idx {
				switch v := row.Values[idx].(type) {
				case int64:
					sum += v
				case Integer:
					sum += v.Value
				default:
					fmt.Println("row--->")
					fmt.Println(reflect.TypeOf(v))
				}
			} else {
				// TODO fmt.Println("no VAL")
			}
		}
		return Integer{int64(sum)}
	} else {
		return *NewError("Column not found")
	}
}

func (s Spreadsheet) Sum_Just(name string) (float64, error) {
	var sum int64
	var sumf float64
	idx := IndexOfString(name, s.Cols)
	if idx > -1 {
		for _, row := range s.Rows {
			if len(row.Values) > idx {
				switch v := row.Values[idx].(type) {
				case float64:
					sumf += v
				case int64:
					sum += v
					break
				case Integer:
					sum += v.Value
					break
				case Decimal:
					sumf += v.Value
					break
				default:
					fmt.Println("row--->")
					fmt.Println(reflect.TypeOf(v))
				}
			} else {
				// TODO fmt.Println("no VAL")
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
	if s.RawMode {
		return len(s.RawRows)
	} else {
		return len(s.Rows)
	}
}

func (s Spreadsheet) Columns(ps *ProgramState, names []string) Object {

	idxs := make([]int, len(names))
	for name := range names {
		idx := IndexOfString(names[name], s.Cols)
		if idx == -1 {
			return makeError(ps, "Col not found")
		}
		idxs[name] = idx

	}
	nspr := NewSpreadsheet(names)
	if s.RawMode {
		res := make([][]string, 0)
		for _, row := range s.RawRows {
			row2 := make([]string, len(names))
			for col := range idxs {
				if len(row) > col {
					row2[col] = row[idxs[col]]
				}
			}
			res = append(res, row2)
		}
		nspr.SetRaw(res)
		return *nspr
	}
	return makeError(ps, "Only raw spreadsheet for now TODO")

}

func (s Spreadsheet) GetRow(ps *ProgramState, index int) Object {
	if s.RawMode {
		row := s.RawRows[index]
		row2 := make([]interface{}, len(row))
		for i := range row {
			row2[i] = row[i]
		}
		return SpreadsheetRow{row2, &s}
	} else {
		row := s.Rows[index]
		row.Uplink = &s
		return row
	}
}

func (s Spreadsheet) GetRowNew(index int) Object {
	if s.RawMode {
		row := s.RawRows[index]
		row2 := make([]interface{}, len(row))
		for i := range row {
			row2[i] = row[i]
		}
		return SpreadsheetRow{row2, &s}
	} else {
		row := s.Rows[index]
		row.Uplink = &s
		return row
	}
}

func (s Spreadsheet) GetRawRowValue(column string, rrow []string) (string, error) {
	index := -1
	for i, v := range s.Cols {
		if v == column {
			index = i
			break
		}
	}
	if index < 0 {
		return "", nil
	}
	return rrow[index], nil
}

func (s Spreadsheet) GetRowValue(column string, rrow SpreadsheetRow) (interface{}, error) {
	index := -1
	for i, v := range s.Cols {
		if v == column {
			index = i
			break
		}
	}
	if index < 0 {
		return "", nil
	}
	return rrow.Values[index], nil
}

// Type returns the type of the Integer.
func (s Spreadsheet) Type() Type {
	return SpreadsheetType
}

// Inspect returns a string
func (s Spreadsheet) Inspect(e Idxs) string {
	rows := ""
	if s.RawMode {
		rows = strconv.Itoa(len(s.RawRows))
	} else {
		rows = strconv.Itoa(len(s.Rows))
	}
	var kindStr string
	//fmt.Println(s.GetKind())
	if s.GetKind() != int(SpreadsheetType) {
		kindStr = " of kind " + s.Kind.Probe(e)
	}
	return "[Spreadsheet(" + strconv.Itoa(len(s.Cols)) + " " + rows + ")" + kindStr + "]"
}

// Inspect returns a string representation of the Integer.
func (s Spreadsheet) Probe(e Idxs) string {
	return s.ToTxt()
}

func (s Spreadsheet) Trace(msg string) {
	fmt.Print(msg + " (spreadsheet): ")
}

func (s Spreadsheet) GetKind() int {
	return int(SpreadsheetType)
}

func (s SpreadsheetRow) GetKind() int {
	return int(0)
}

// Inspect returns a string
func (s SpreadsheetRow) Inspect(e Idxs) string {
	return "[SpreadsheetRow(" + strconv.Itoa(len(s.Values)) + ")" + "]"
}

// Inspect returns a string representation of the Integer.
func (s SpreadsheetRow) Probe(e Idxs) string {
	return s.ToTxt()
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

func (s Spreadsheet) GetColumns() List {
	lst := make([]interface{}, len(s.Cols))
	for i, v := range s.Cols {
		lst[i] = v
	}
	return *NewList(lst)
}
