package evaldo

import "C"

import (
	"Rejy_go_v1/env"
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

type SpreadsheetRow struct {
	values []interface{}
}

type Spreadsheet struct {
	cols []string
	rows []SpreadsheetRow
}

func NewSpreadsheet(cols []string) *Spreadsheet {
	var ps Spreadsheet
	ps.cols = cols
	ps.rows = make([]SpreadsheetRow, 1)
	/*
		ps := Spreadsheet{
			cols,
			make([]SpreadsheetRow, 1)
		} */
	return &ps
}

// Inspect returns a string representation of the Integer.
func (s *Spreadsheet) addRow(vals SpreadsheetRow) {
	s.rows = append(s.rows, vals)
}

// Inspect returns a string representation of the Integer.
func (s *Spreadsheet) setCols(vals []string) {
	s.cols = vals
}

// Inspect returns a string representation of the Integer.
func (s Spreadsheet) toHtml() string {
	fmt.Println("IN TO Html")
	var bu strings.Builder
	bu.WriteString("<table>")
	for _, row := range s.rows {
		bu.WriteString("<tr>")
		for _, val := range row.values {
			bu.WriteString("<td>")
			bu.WriteString(fmt.Sprint(val))
			bu.WriteString("</td>")
		}
		bu.WriteString("</tr>")
	}
	bu.WriteString("</table>")
	fmt.Println(bu.String())
	return bu.String()
}

func SQL_EvalBlock(es *env.ProgramState) (*env.ProgramState, []interface{}) {
	var bu strings.Builder
	var str string
	values := make([]interface{}, 0, 2) // TODO ... what is this 2 here ... just for temp
	for es.Ser.Pos() < es.Ser.Len() {
		es, str, values = SQL_EvalExpression(es, values)
		bu.WriteString(str + " ")
		fmt.Println(bu.String())
	}
	es.Res = env.String{bu.String()}
	return es, values
}

func SQL_EvalExpression(es *env.ProgramState, vals []interface{}) (*env.ProgramState, string, []interface{}) {
	object := es.Ser.Pop()

	switch obj := object.(type) {
	case env.Integer:
		return es, strconv.FormatInt(obj.Value, 10), vals
	case env.String:
		return es, "\"" + obj.Value + "\"", vals
	/*case env.VoidType:
		es.Res = object
	case env.TagwordType:
		es.Res = object
	case env.UriType:
		es.Res = object */
	case env.Word:
		return es, es.Idx.GetWord(obj.Index), vals
	/*case env.GenwordType:
		return EvalGenword(es, object.(env.Genword), nil, false)
	case env.SetwordType:
		return EvalSetword(es, object.(env.Setword)) */
	case env.Getword:
		val, _ := es.Env.Get(obj.Index)
		vals = append(vals, val.(env.Integer).Value)
		return es, "?", vals
	case env.Comma:
		return es, ", ", vals
	default:
		return es, "Error 123112431", vals
	}
	return es, "ERROR", vals
}

var Builtins_sqlite = map[string]*env.Builtin{

	"sqlite-schema//open": {
		Argsn: 1,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			arg0.Trace("OPEN :::::::::")
			switch str := arg0.(type) {
			case env.Uri:
				fmt.Println(str.Path)
				db, _ := sql.Open("sqlite3", "temp-database") // TODO -- we need to make path parser in URI then this will be path
				return *env.NewNative(env1.Idx, db, "Rye-sqlite")
			default:
				return env.NewError("arg 2 should be Uri")
			}

		},
	},

	"Rye-spreadsheet//htmlize": {
		Argsn: 1,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch str := arg0.(type) {
			case env.Native:
				return env.String{str.Value.(Spreadsheet).toHtml()}
			default:
				return env.NewError("arg 2 should be string %s")
			}

		},
	},

	"Rye-sqlite//exec": {
		Argsn: 2,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch db1 := arg0.(type) {
			case env.Native:
				switch str := arg1.(type) {
				case env.String:
					db2 := db1.Value.(*sql.DB)
					db2.Exec(str.Value)
					return env.Void{}
				default:
					return env.NewError("arg 2222 should be string %s")
				}
			default:
				return env.NewError("arg 1111 should be string %s")
			}
			return env.NewError("arg 0000 should be string %s")
		},
	},

	"Rye-sqlite//query": {
		Argsn: 2,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			var sqlstr string
			var vals []interface{}
			switch db1 := arg0.(type) {
			case env.Native:
				switch str := arg1.(type) {
				case env.Block:
					fmt.Println("BLOCK ****** *****")
					ser := env1.Ser
					env1.Ser = str.Series
					_, vals = SQL_EvalBlock(env1)
					sqlstr = env1.Res.(env.String).Value
					env1.Ser = ser
				case env.String:
					sqlstr = str.Value
				default:
					return env.NewError("arg 2222 should be string %s")
				}
				if sqlstr != "" {
					spr := NewSpreadsheet([]string{"name", "id"})
					rows, err := db1.Value.(*sql.DB).Query(sqlstr, vals...)
					result := make([]map[string]interface{}, 0)
					if err != nil {
						fmt.Println(err.Error())
					} else {
						cols, _ := rows.Columns()
						for rows.Next() {

							var sr SpreadsheetRow

							columns := make([]interface{}, len(cols))
							columnPointers := make([]interface{}, len(cols))
							for i, _ := range columns {
								columnPointers[i] = &columns[i]
							}

							// Scan the result into the column pointers...
							if err := rows.Scan(columnPointers...); err != nil {
								//return err
							}

							// Create our map, and retrieve the value for each column from the pointers slice,
							// storing it in the map with the name of the column as the key.
							m := make(map[string]interface{})
							for i, colName := range cols {
								val := columnPointers[i].(*interface{})
								m[colName] = *val
								sr.values = append(sr.values, *val)
							}
							spr.addRow(sr)
							result = append(result, m)
							// Outputs: map[columnName:value columnName2:value2 columnName3:value3 ...]
						}
						rows.Close() //good habit to close
						fmt.Println("+++++")
						fmt.Print(result)
						return *env.NewNative(env1.Idx, *spr, "Rye-spreadsheet")
					}
					return env.NewError("Empty SQL")
				}
			default:
				return env.NewError("arg 1111 should be string %s")
			}
			return env.NewError("arg 1111 should be string %s")

		},
	},
}
