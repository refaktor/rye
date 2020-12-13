package env

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type SpreadsheetRow struct {
	Values []interface{}
}

type Spreadsheet struct {
	Cols []string
	Rows []SpreadsheetRow
	Kind Word
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

// Inspect returns a string representation of the Integer.
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

// Type returns the type of the Integer.
func (s Spreadsheet) Type() Type {
	return SpreadsheetType
}

// Inspect returns a string
func (s Spreadsheet) Inspect(e Idxs) string {
	return "<Spreadsheet [" + strconv.Itoa(len(s.Cols)) + " " + strconv.Itoa(len(s.Rows)) + "] of kind " + s.Kind.Probe(e) + ">"
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
