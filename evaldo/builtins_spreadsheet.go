// SECTION: Core/spreadsheets

package evaldo

import (
	"encoding/csv"
	"fmt"
	"os"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/refaktor/rye/env"
	"github.com/refaktor/rye/util"
)

var Builtins_spreadsheet = map[string]*env.Builtin{

	// Tests:
	//  equals { spreadsheet { "a" } { 1 2 } |type? } 'spreadsheet
	// Args:
	//  * columns
	//  * data
	"spreadsheet": {
		Argsn: 2,
		Doc:   "Creates a spreadsheet by accepting block of column names and flat block of values",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch header1 := arg0.(type) {
			case env.Block:
				header := header1.Series
				hlen := header.Len()
				cols := make([]string, hlen)
				for header.Pos() < hlen {
					i := header.Pos()
					k1 := header.Pop()
					switch k := k1.(type) {
					case env.String:
						cols[i] = k.Value
					}
				}
				spr := env.NewSpreadsheet(cols)
				switch data1 := arg1.(type) {
				case env.Block:
					rdata := data1.Series.S

					for i := 0; i < len(rdata)/hlen; i++ {
						rowd := make([]any, hlen)
						for ii := 0; ii < hlen; ii++ {
							rowd[ii] = rdata[i*hlen+ii]
						}
						spr.AddRow(*env.NewSpreadsheetRow(rowd, spr))
					}
					return *spr
				case env.List:
					rdata := data1.Data
					for i := 0; i < len(rdata)/hlen; i++ {
						rowd := make([]any, hlen)
						for ii := 0; ii < hlen; ii++ {
							rowd[ii] = rdata[i*hlen+ii]
						}
						spr.AddRow(*env.NewSpreadsheetRow(rowd, spr))
					}
					return *spr
				default:
					return MakeArgError(ps, 2, []env.Type{env.BlockType}, "spreadsheet")
				}
				/* for data.Pos() < data.Len() {
					rowd := make([]any, header.Len())
					for ii := 0; ii < header.Len(); ii++ {
						k1 := data.Pop()
						rowd[ii] = k1
					}
					spr.AddRow(*env.NewSpreadsheetRow(rowd, spr))
				} */
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType}, "spreadsheet")
			}
		},
	},
	// Example:
	//  spreadsheet\columns { 'a 'b } { { 1 2 } { "x" "y" } }
	"spreadsheet\\columns": {
		Argsn: 2,
		Doc:   "Creats a spreadsheet by accepting a block of columns",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) (res env.Object) {
			switch columns := arg0.(type) {
			case env.Block:
				colNames := columns.Series
				numCols := colNames.Len()
				if numCols == 0 {
					return MakeBuiltinError(ps, "Block of column names is empty", "spreadsheet\\columns")
				}
				cols := make([]string, numCols)
				for colNames.Pos() < numCols {
					i := colNames.Pos()
					k1 := colNames.Pop()
					switch k := k1.(type) {
					case env.String:
						cols[i] = k.Value
					case env.Word:
						cols[i] = ps.Idx.GetWord(k.Index)
					default:
						return MakeBuiltinError(ps, fmt.Sprintf("Expected a string or word instead of %V", k), "spreadsheet\\columns")
					}
					// TODO: Error here?
				}

				spr := env.NewSpreadsheet(cols)

				var colData []env.SpreadsheetRow

				switch colSet := arg1.(type) {
				case env.Block:
					blockColData := colSet.Series.S
					var numRows int
					switch firstCol := blockColData[0].(type) {
					case env.Block:
						numRows = firstCol.Series.Len()
					case env.List:
						numRows = len(firstCol.Data)
					default:
						return MakeBuiltinError(ps, fmt.Sprintf("Expected a block or a list instead of %V", firstCol), "spreadsheet\\columns")
					}

					colData = make([]env.SpreadsheetRow, numRows)
					for rowIdx := 0; rowIdx < numRows; rowIdx++ {
						colData[rowIdx].Values = make([]any, numCols)
					}

					for colIdx, c := range blockColData {
						switch colSeries := c.(type) {
						case env.Block:
							if colSeries.Series.Len() != numRows {
								return MakeBuiltinError(
									ps,
									fmt.Sprintf("Column %s should have %d rows of data, but has %d instead",
										cols[colIdx],
										numRows,
										colSeries.Series.Len(),
									),
									"spreadsheet\\columns",
								)
							}
							for rowIdx, value := range colSeries.Series.S {
								colData[rowIdx].Values[colIdx] = value
							}
						case env.List:
							if len(colSeries.Data) != numRows {
								return MakeBuiltinError(
									ps,
									fmt.Sprintf("Column %s should have %d rows of data, but has %d instead",
										cols[colIdx],
										numRows,
										len(colSeries.Data),
									),
									"spreadsheet\\columns",
								)
							}
							for rowIdx, value := range colSeries.Data {
								colData[rowIdx].Values[colIdx] = value
							}
						}
					}
					spr.Rows = colData
					return spr
				default:
					return MakeBuiltinError(ps, "", "")
				}
			default:
				return MakeBuiltinError(ps, "", "")
			}
		},
	},
	// Tests:
	//  equals { to-spreadsheet dict { "a" 1 "a" b } |type? } 'spreadsheet
	// Args:
	//  * data
	"to-spreadsheet": {
		Argsn: 1,
		Doc:   "Creates a spreadsheet by accepting block or list of dicts",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) (res env.Object) {
			switch block := arg0.(type) {
			case env.Block:
				data := block.Series
				if data.Len() == 0 {
					return MakeBuiltinError(ps, "Block is empty", "to-spreadsheet")
				}
				k := make(map[string]struct{})
				for _, obj := range data.S {
					switch dict := obj.(type) {
					case env.Dict:
						for key := range dict.Data {
							k[key] = struct{}{}
						}
					default:
						return MakeBuiltinError(ps, "Block must contain only dicts", "to-spreadsheet")
					}
				}
				var keys []string
				for key := range k {
					keys = append(keys, key)
				}
				spr := env.NewSpreadsheet(keys)
				for _, obj := range data.S {
					switch dict := obj.(type) {
					case env.Dict:
						row := make([]any, len(keys))
						for i, key := range keys {
							data, ok := dict.Data[key]
							if !ok {
								data = env.Void{}
							}
							row[i] = data
						}
						spr.AddRow(*env.NewSpreadsheetRow(row, spr))
					}
				}
				return *spr

			case env.List:
				data := block.Data
				if len(data) == 0 {
					return MakeBuiltinError(ps, "List is empty", "to-spreadsheet")
				}
				k := make(map[string]struct{})
				for _, obj := range data {
					switch dict := obj.(type) {
					case map[string]any:
						for key := range dict {
							k[key] = struct{}{}
						}
					default:
						return MakeBuiltinError(ps, "List must contain only dicts", "to-spreadsheet")
					}
				}
				var keys []string
				for key := range k {
					keys = append(keys, key)
				}
				spr := env.NewSpreadsheet(keys)
				for _, obj := range data {
					row := make([]any, len(keys))
					switch dict := obj.(type) {
					case map[string]any:
						for i, key := range keys {
							data, ok := dict[key]
							if !ok {
								data = env.Void{}
							}
							row[i] = data
						}
					}
					spr.AddRow(*env.NewSpreadsheetRow(row, spr))
				}
				return *spr

			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType, env.ListType}, "to-spreadsheet")
			}
		},
	},

	// Get rows as a native. This value can be used in `add-rows` and `add-rows!``
	// Args:
	// * sheet
	"get-rows": {
		Argsn: 1,
		Doc:   "Get rows as a native",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch spr := arg0.(type) {
			case env.Spreadsheet:
				rows := spr.GetRows()
				return *env.NewNative(ps.Idx, rows, "spreadsheet-rows")
			default:
				return MakeArgError(ps, 1, []env.Type{env.SpreadsheetType}, "get-rows")
			}
		},
	},

	// Add one or more rows to a spreadsheet, returning a new spreadsheet
	// The `rows` argument can take one of two types:
	// 1) a block that has one or more rows worth of data.
	// This is given as a single flat collection. This means that if your
	//  sheet has `NumColumns` columns, your block should have `NumColumns * NumRows` values.
	// 2) A native that is a slice of SpreadsheetRows, like the value returned from `get-rows`
	// Tests:
	//  equal {
	//	 ref spreadsheet { "a" "b"  } { 1 10 2 20 } :sheet
	//   sheet .add-rows! [ 3 30 ] sheet .deref .length?
	//  } 3
	// Args:
	// * sheet -the sheet that is getting rows added to it
	// * rows - a block containing one or more rows worth of values, or a SpreadsheetRow Native value
	"add-rows": {
		Argsn: 2,
		Doc:   "Add one or more rows to a spreadsheet",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch spr := arg0.(type) {
			case env.Spreadsheet:
				switch data1 := arg1.(type) {
				case env.Block:
					data := data1.Series
					for data.Pos() < data.Len() {
						rowd := make([]any, len(spr.Cols))
						for ii := 0; ii < len(spr.Cols); ii++ {
							k1 := data.Pop()
							rowd[ii] = k1
						}
						spr.AddRow(*env.NewSpreadsheetRow(rowd, &spr))
					}
					return spr
				case env.Native:
					spr.Rows = append(spr.Rows, data1.Value.([]env.SpreadsheetRow)...)
					return spr
				default:
					return MakeArgError(ps, 2, []env.Type{env.BlockType, env.NativeType}, "add-rows")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.SpreadsheetType}, "add-rows")
			}
		},
	},

	// Add one or more rows to a spreadsheet ref. Works similary to `add-rows`, but
	// modified the spreadsheet ref instead of returning a new copy
	// of the spreasheet
	// Tests:
	//  equal {
	//	 ref spreadsheet { "a" "b"  } { 1 10 2 20 } :sheet
	//   sheet .add-rows! [ 3 30 ] sheet .deref .length?
	//  } 3
	// Args:
	// * sheet - the reference to the sheet that is getting rows added to it
	// * rows - a block containing one or more rows worth of values, or a SpreadsheetRow Native value
	// Tags: #spreasheet #mutation
	"add-rows!": {
		Argsn: 2,
		Doc:   "Add one or more rows to a spreadsheet ref",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch spr := arg0.(type) {
			case *env.Spreadsheet:
				switch data1 := arg1.(type) {
				case env.Block:
					data := data1.Series
					for data.Pos() < data.Len() {
						rowd := make([]any, len(spr.Cols))
						for ii := 0; ii < len(spr.Cols); ii++ {
							k1 := data.Pop()
							rowd[ii] = k1
						}
						spr.AddRow(*env.NewSpreadsheetRow(rowd, spr))
					}
					return spr
				case env.Native:
					spr.Rows = append(spr.Rows, data1.Value.([]env.SpreadsheetRow)...)
					return spr
				default:
					return MakeArgError(ps, 2, []env.Type{env.BlockType, env.NativeType}, "add-rows!")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.SpreadsheetType}, "add-rows!")
			}
		},
	},

	// Update the row at a given index. If given a dict or spreadhseet row, replace the row with that
	// If given a function, pass the row, its index and replace the row with the return value from the
	// function.
	// Tests:
	//  equal {
	//	 spr1: ref spreadsheet { "a" "b" } { 1 10 2 20 }
	//	 spr1 .update-row! 1 dict [ "a" 111 ]
	//   spr1 .deref .A1
	//  } 111
	//  equal {
	//	 spr1: ref spreadsheet { "a" "b" } { 1 10 2 20 }
	//	 incrA: fn { row } { row -> "a" + 1 :new-a dict { "a" new-a } }
	//	 spr1 .update-row! 1 dict incrA
	//   spr1 .deref .A1
	//  } 11
	// Args:
	// * sheet-ref - A ref to a spreadsheet
	// * idx - the index of the row to update, 1-based
	// * updater - One of either a function, a dict, or a Spreadsheet Row
	// Tags: #spreadsheet #mutation
	"update-row!": {
		Argsn: 3, // Spreadsheet, index function/dict
		Doc:   `Update the row at the given index.`,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch spr := arg0.(type) {
			case *env.Spreadsheet:
				switch idx := arg1.(type) {
				case env.Integer:
					if idx.Value < 1 || (idx.Value-1) > int64(len(spr.Rows)) {
						errMsg := fmt.Sprintf("update-row! called with row index %i, but spreadsheet only has %i rows", idx.Value, len(spr.Rows))
						return makeError(ps, errMsg)
					}
					switch updater := arg2.(type) {
					case env.Function:
						CallFunctionArgs4(updater, ps, spr.Rows[idx.Value-1], idx, nil, nil, ps.Ctx)
						if !ps.ReturnFlag {
							return makeError(ps, "Function given to update-row! should have returned a value, but didn't")
						}
						if ok, err, row := RyeValueToSpreadsheetRow(spr, ps.Res); ok {
							spr.Rows[idx.Value-1] = *row
							return spr
						} else if len(err) > 0 {
							return makeError(ps, err)
						} else {
							return makeError(ps, fmt.Sprintf(
								"Function given to update-row! should have returned a Dict or a SpreadsheetRow, but returned a %s instead",
								NameOfRyeType(ps.Res.Type()),
							))
						}
					case env.Dict:
						// dict should be able to have only the columns you want to update
						row := spr.Rows[idx.Value-1]

						for keyStr, val := range updater.Data {
							index := spr.GetColumnIndex(keyStr)
							if index < 0 {
								return makeError(ps, "Column "+keyStr+" was not found")
							}
							row.Values[index] = val
						}
						return spr
					case env.SpreadsheetRow:
						spr.Rows[idx.Value-1] = updater
						return spr
					default:
						return MakeArgError(ps, 3, []env.Type{env.FunctionType, env.DictType, env.SpreadsheetRowType}, "update-row")
					}
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "update-row!")
				}
			default:
				return MakeNeedsThawedArgError(ps, "update-row!")
			}

		},
	},
	// Tests:
	//  equal {
	//   spr1: spreadsheet { "a" "b" } { 1 10 2 20 }
	//   spr1 .remove-row! 1
	//   spr1 .deref .A1
	//  } 2
	// Args:
	// * sheet-ref
	// * row-idx - Index of row to remove, 1-based
	// Tags: #spreadsheet #mutation
	"remove-row!": {
		Argsn: 2,
		Doc:   "Remove a row from a spreadsheet by index",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch spr := arg0.(type) {
			case *env.Spreadsheet:
				switch data1 := arg1.(type) {
				case env.Integer:
					if data1.Value > 0 && data1.Value <= int64(len(spr.Rows)) {
						spr.RemoveRowByIndex(data1.Value - 1)
						return spr
					} else {
						return makeError(ps, fmt.Sprintf("Spreadsheet had less then %d rows", data1.Value))
					}
				default:
					return MakeArgError(ps, 2, []env.Type{env.BlockType, env.NativeType}, "remove-row!")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.SpreadsheetType}, "remove-row!")
			}
		},
	},

	// Args:
	// * file-uri - location of csv file to load
	// Tags: #spreadsheet #loading #csv
	"load\\csv": {
		// TODO 2 -- this could move to a go function so it could be called by general load that uses extension to define the loader
		Argsn: 1,
		Doc:   "Loads a .csv file to a spreadsheet datatype.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch file := arg0.(type) {
			case env.Uri:
				// rows, err := db1.Value.(*sql.DB).Query(sqlstr, vals...)
				f, err := os.Open(file.GetPath())
				if err != nil {
					// log.Fatal("Unable to read input file "+filePath, err)
					return MakeBuiltinError(ps, "Unable to read input file.", "load\\csv")
				}
				defer f.Close()

				csvReader := csv.NewReader(f)
				rows, err := csvReader.ReadAll()
				if err != nil {
					// log.Fatal("Unable to parse file as CSV for "+filePath, err)
					return MakeBuiltinError(ps, "Unable to parse file as CSV.", "load\\csv")
				}
				if len(rows) == 0 {
					return MakeBuiltinError(ps, "File is empty", "load\\csv")
				}
				spr := env.NewSpreadsheet(rows[0])
				//				for i, row := range rows {
				//	if i > 0 {
				//		anyRow := make([]any, len(row))
				//		for i, v := range row {
				//			anyRow[i] = v
				if len(rows) > 1 {
					for _, row := range rows[1:] {
						anyRow := make([]any, len(row))
						for i, v := range row {
							anyRow[i] = *env.NewString(v)
						}
						spr.AddRow(*env.NewSpreadsheetRow(anyRow, spr))
					}
				}
				return *spr
			default:
				return MakeArgError(ps, 1, []env.Type{env.UriType}, "load\\csv")
			}
		},
	},
	// Args:
	// * sheet    - the sheet to save
	// * file-url - where to save the sheet as a .csv file
	// Tags: #spreadsheet #saving #csv
	"save\\csv": {
		Argsn: 2,
		Doc:   "Saves a spreadsheet to a .csv file.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch spr := arg0.(type) {
			case env.Spreadsheet:
				switch file := arg1.(type) {
				case env.Uri:
					// rows, err := db1.Value.(*sql.DB).Query(sqlstr, vals...)
					f, err := os.Create(file.GetPath())
					if err != nil {
						// log.Fatal("Unable to read input file "+filePath, err)
						return MakeBuiltinError(ps, "Unable to create input file.", "save\\csv")
					}
					defer f.Close()

					cLen := len(spr.Cols)

					csvWriter := csv.NewWriter(f)

					err1 := csvWriter.Write(spr.Cols)
					if err1 != nil {
						return MakeBuiltinError(ps, "Unable to create write header.", "save\\csv")
					}

					for ir, row := range spr.Rows {
						strVals := make([]string, cLen)
						// TODO -- just adhoc ... move to a general function in utils RyeValsToString or something like it
						for i, v := range row.Values {
							var sv string
							switch tv := v.(type) {
							case string:
								sv = tv
							case int64:
								sv = strconv.Itoa(int(tv))
							case float64:
								sv = strconv.FormatFloat(tv, 'f', -1, 64)
							case env.String:
								sv = tv.Value
							case env.Integer:
								sv = strconv.Itoa(int(tv.Value))
							case env.Decimal:
								sv = fmt.Sprintf("%f", tv.Value)
							}
							strVals[i] = sv
						}
						err := csvWriter.Write(strVals)
						if err != nil {
							return MakeBuiltinError(ps, "Unable to write line: "+strconv.Itoa(ir), "save\\csv")
						}
					}
					csvWriter.Flush()
					f.Close()
					return spr
				default:
					return MakeArgError(ps, 1, []env.Type{env.UriType}, "save\\csv")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.UriType}, "save\\csv")
			}
		},
	},

	// Example: filtering for rows with the name "Enno"
	//  sheet: spreadsheet { "name" } { "Enno" "Enya" "Enid" "Bob" "Bill" }
	//  sheet .where-equal 'name "Enno"
	// Args:
	// * sheet
	// * column
	// * value
	// Tags: #filter #spreadsheets
	"where-equal": {
		Argsn: 3,
		Doc:   "Returns spreadsheet of rows where specific colum is equal to given value.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch spr := arg0.(type) {
			case env.Spreadsheet:
				switch col := arg1.(type) {
				case env.Word:
					return WhereEquals(ps, spr, ps.Idx.GetWord(col.Index), arg2)
				case env.String:
					return WhereEquals(ps, spr, col.Value, arg2)
				default:
					return MakeArgError(ps, 2, []env.Type{env.WordType, env.StringType}, "where-equal")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.SpreadsheetType}, "where-equal")
			}
		},
	},
	// Example: filting for names that start with "En"
	//  sheet: spreadsheet { "name" } { "Enno" "Enya" "Enid" "Bob" "Bill" }
	//  sheet .where-match 'name "En.+"
	// Args:
	// * sheet
	// * column
	// * regexp
	// Tags: #filter #spreadsheets
	"where-match": {
		Argsn: 3,
		Doc:   "Returns spreadsheet of rows where a specific colum matches a regex.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) (res env.Object) {
			switch spr := arg0.(type) {
			case env.Spreadsheet:
				switch reNative := arg2.(type) {
				case env.Native:
					re, ok := reNative.Value.(*regexp.Regexp)
					if !ok {
						return MakeArgError(ps, 2, []env.Type{env.NativeType}, "where-match")
					}
					switch col := arg1.(type) {
					case env.Word:
						return WhereMatch(ps, spr, ps.Idx.GetWord(col.Index), re)
					case env.String:
						return WhereMatch(ps, spr, col.Value, re)
					default:
						return MakeArgError(ps, 2, []env.Type{env.WordType, env.StringType}, "where-match")
					}
				default:
					return MakeArgError(ps, 3, []env.Type{env.NativeType}, "where-match")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.SpreadsheetType}, "where-match")
			}
		},
	},

	// Example: filting for names that contain "nn"
	//  sheet: spreadsheet { "name" } { "Enno" "Enya" "Enid" "Bob" "Bill" "Benn" }
	//  sheet .where-contains 'name "nn"
	// Args:
	// * sheet
	// * column
	// * substring
	// Tags: #filter #spreadsheets
	"where-contains": {
		Argsn: 3,
		Doc:   "Returns spreadsheet of rows where specific colum contains a given string value.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) (res env.Object) {
			switch spr := arg0.(type) {
			case env.Spreadsheet:
				switch s := arg2.(type) {
				case env.String:
					switch col := arg1.(type) {
					case env.Word:
						return WhereContains(ps, spr, ps.Idx.GetWord(col.Index), s.Value, false)
					case env.String:
						return WhereContains(ps, spr, col.Value, s.Value, false)
					default:
						return MakeArgError(ps, 2, []env.Type{env.WordType, env.StringType}, "where-contains")
					}
				default:
					return MakeArgError(ps, 3, []env.Type{env.StringType}, "where-contains")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.SpreadsheetType}, "where-contains")
			}
		},
	},

	// Example: filting for names that contain "nn"
	//  sheet: spreadsheet { "name" } { "Enno" "Enya" "Enid" "Bob" "Bill" "Benn" }
	//  sheet .where-contains 'name "nn"
	// Args:
	// * sheet
	// * column
	// * substring
	// Tags: #filter #spreadsheets
	"where-not-contains": {
		Argsn: 3,
		Doc:   "Returns spreadsheet of rows where specific colum contains a given string value.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) (res env.Object) {
			switch spr := arg0.(type) {
			case env.Spreadsheet:
				switch s := arg2.(type) {
				case env.String:
					switch col := arg1.(type) {
					case env.Word:
						return WhereContains(ps, spr, ps.Idx.GetWord(col.Index), s.Value, true)
					case env.String:
						return WhereContains(ps, spr, col.Value, s.Value, true)
					default:
						return MakeArgError(ps, 2, []env.Type{env.WordType, env.StringType}, "where-not-contains")
					}
				default:
					return MakeArgError(ps, 3, []env.Type{env.StringType}, "where-not-contains")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.SpreadsheetType}, "where-not-contains")
			}
		},
	},
	// Example: filting for ages over 29
	//  sheet: spreadsheet { "name" "age" } { "Enno" 30 "Enya" 25 "Enid" 40 "Bob" 19 "Bill" 45 "Benn" 29 }
	//  sheet .where-greater 'age 29
	// Args:
	// * sheet
	// * column
	// * value
	// Tags: #filter #spreadsheet
	"where-greater": {
		Argsn: 3,
		Doc:   "Returns spreadsheet of rows where specific colum is greater than given value.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch spr := arg0.(type) {
			case env.Spreadsheet:
				switch col := arg1.(type) {
				case env.Word:
					return WhereGreater(ps, spr, ps.Idx.GetWord(col.Index), arg2)
				case env.String:
					return WhereGreater(ps, spr, col.Value, arg2)
				default:
					return MakeArgError(ps, 2, []env.Type{env.WordType, env.StringType}, "where-greater")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.SpreadsheetType}, "where-greater")
			}
		},
	},
	// Example: filting for names that contain "nn"
	//  sheet: spreadsheet { "name" "age" } { "Enno" 30 "Enya" 25 "Enid" 40 "Bob" 19 "Bill" 45 "Benn" 29 }
	//  sheet .where-lesser 'age 29
	// Args:
	// * sheet
	// * column
	// * value
	// Tags: #filter #spreadsheet
	"where-lesser": {
		Argsn: 3,
		Doc:   "Returns spreadsheet of rows where specific colum is lesser than given value.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch spr := arg0.(type) {
			case env.Spreadsheet:
				switch col := arg1.(type) {
				case env.Word:
					return WhereLesser(ps, spr, ps.Idx.GetWord(col.Index), arg2)
				case env.String:
					return WhereLesser(ps, spr, col.Value, arg2)
				default:
					return MakeArgError(ps, 2, []env.Type{env.WordType, env.StringType}, "where-lesser")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.SpreadsheetType}, "where-lesser")
			}
		},
	},
	// Returns a spreadhsheet of rows where the given column is between the given
	// values, non-inclusive.
	// Example: filtering for folks in their 20s
	//  sheet: spreadsheet { "name" "age" } { "Enno" 30 "Enya" 25 "Enid" 40 "Bob" 19 "Bill" 45 "Benn" 29 }
	//  sheet .where-between 'age 19 30
	// Args:
	// * sheet
	// * column
	// * lower-limit
	// * upper-limit
	// Tags: #filter #spreadsheet
	"where-between": {
		Argsn: 4,
		Doc:   "Returns spreadsheet of rows where specific colum is between given values.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) (res env.Object) {
			switch spr := arg0.(type) {
			case env.Spreadsheet:
				switch col := arg1.(type) {
				case env.Word:
					return WhereBetween(ps, spr, ps.Idx.GetWord(col.Index), arg2, arg3)
				case env.String:
					return WhereBetween(ps, spr, col.Value, arg2, arg3)
				default:
					return MakeArgError(ps, 2, []env.Type{env.WordType, env.StringType}, "where-between")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.SpreadsheetType}, "where-between")
			}
		},
	},

	// Example: filtering for folks named "Enno" or "Enya"
	//  sheet: spreadsheet { "name" "age" } { "Enno" 30 "Enya" 25 "Enid" 40 "Bob" 19 "Bill" 45 "Benn" 29 }
	//  sheet .where-in 'name { "Enno" "Enya" }
	// Args:
	// * sheet
	// * column
	// * values-filtered-for
	// Tags: #filter #spreadsheet
	"where-in": {
		Argsn: 3,
		Doc:   "Returns spreadsheet of rows where specific colum value if found in block of values.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) (res env.Object) {
			switch spr := arg0.(type) {
			case env.Spreadsheet:
				switch s := arg2.(type) {
				case env.Block:
					switch col := arg1.(type) {
					case env.Word:
						return WhereIn(ps, spr, ps.Idx.GetWord(col.Index), s.Series.S)
					case env.String:
						return WhereIn(ps, spr, col.Value, s.Series.S)
					default:
						return MakeArgError(ps, 2, []env.Type{env.WordType, env.StringType}, "where-in")
					}
				default:
					return MakeArgError(ps, 3, []env.Type{env.StringType}, "where-in")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.SpreadsheetType}, "where-in")
			}
		},
	},

	// Tags: #spreadsheet
	"limit": {
		Argsn: 2,
		Doc:   "Returns spreadsheet with number of rows limited to second argument.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch spr := arg0.(type) {
			case env.Spreadsheet:
				switch n := arg1.(type) {
				case env.Integer:
					return Limit(ps, spr, int(n.Value))
				default:
					return MakeArgError(ps, 1, []env.Type{env.IntegerType}, "limit")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.SpreadsheetType}, "limit")
			}
		},
	},

	// Example: Order by age ascending
	//  sheet: spreadsheet { "name" "age" } { "Bob" 25 "Alice" 29 "Charlie" 19  }
	//  sheet .order-by! 'age 'asc
	// Tags: #spreadsheet
	"order-by!": {
		Argsn: 3,
		Doc:   "Sorts row by given column, changes spreadsheet in place.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			dir, ok := arg2.(env.Word)
			if !ok {
				return MakeArgError(ps, 3, []env.Type{env.WordType}, "sort-by!")
			}
			var dirAsc bool
			if dir.Index == ps.Idx.IndexWord("asc") {
				dirAsc = true
			} else if dir.Index == ps.Idx.IndexWord("desc") {
				dirAsc = false
			} else {
				return MakeBuiltinError(ps, "Direction can be just asc or desc.", "sort-by!")
			}
			switch spr := arg0.(type) {
			case env.Spreadsheet:
				switch col := arg1.(type) {
				case env.String:
					if dirAsc {
						SortByColumn(ps, &spr, col.Value)
					} else {
						SortByColumnDesc(ps, &spr, col.Value)
					}
					return spr
				case env.Word:
					if dirAsc {
						SortByColumn(ps, &spr, ps.Idx.GetWord(col.Index))
					} else {
						SortByColumnDesc(ps, &spr, ps.Idx.GetWord(col.Index))
					}
					return spr
				default:
					return MakeArgError(ps, 2, []env.Type{env.WordType}, "sort-by!")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.SpreadsheetType}, "sort-by!")
			}
		},
	},

	// Example: Select "name" and "age" columns
	//  sheet: spreadsheet { "name" "age" "job_title" } { "Bob" 25 "Janitor" "Alice" 29 "Librarian" "Charlie" 19 "Line Cook" }
	//  sheet .columns? { 'name 'age }
	// Tags: #spreadsheet
	"columns?": {
		Argsn: 2,
		Doc:   "Returns spreadsheet with just given columns.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch spr := arg0.(type) {
			case env.Spreadsheet:
				switch col := arg1.(type) {
				case env.Block:
					cols := make([]string, col.Series.Len())
					for c := range col.Series.S {
						switch ww := col.Series.S[c].(type) {
						case env.String:
							cols[c] = ww.Value
						case env.Tagword:
							cols[c] = ps.Idx.GetWord(ww.Index)
						}
					}
					return spr.Columns(ps, cols)
				default:
					return MakeArgError(ps, 1, []env.Type{env.BlockType}, "columns")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.SpreadsheetType}, "columns")
			}
		},
	},
	// Example: Get sheet column names
	//  sheet: spreadsheet { "name" "age" "job_title" } { "Bob" 25 "Janitor" "Alice" 29 "Librarian" "Charlie" 19 "Line Cook" }
	//  sheet .header? ; { "name" "age" "job_title" }
	// Tags: #spreadsheet
	"header?": {
		Argsn: 1,
		Doc:   "Gets the column names (header) as block.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch spr := arg0.(type) {
			case *env.Spreadsheet:
				return spr.GetColumns()
			case env.Spreadsheet:
				return spr.GetColumns()
			default:
				return MakeArgError(ps, 1, []env.Type{env.SpreadsheetType}, "headers?")
			}
		},
	},

	// Example: Get sheet column names
	//  sheet: spreadsheet { "name" "age" "job_title" } { "Bob" 25 "Janitor" "Alice" 29 "Librarian" "Charlie" 19 "Line Cook" }
	//  sheet .column? 'name ; => { "Bob" "Alice" "Charlie" }
	// Tags: #spreadsheet
	"column?": {
		Argsn: 2,
		Doc:   "Gets all values of a column as a block.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) (res env.Object) {
			switch spr := arg0.(type) {
			case env.Spreadsheet:
				switch col := arg1.(type) {
				case env.Word:
					return spr.Column(ps.Idx.GetWord(col.Index))
				case env.String:
					return spr.Column(col.Value)
				default:
					return MakeArgError(ps, 2, []env.Type{env.WordType, env.StringType}, "column?")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.SpreadsheetType}, "column?")
			}
		},
	},

	// Example: Drop "job_title" column from sheet
	//  sheet: spreadsheet { "name" "age" "job_title" } { "Bob" 25 "Janitor" "Alice" 29 "Librarian" "Charlie" 19 "Line Cook" }
	//  sheet .drop-column 'job_title ;
	// Example: Drop name and age columns from sheet
	//  sheet: spreadsheet { "name" "age" "job_title" } { "Bob" 25 "Janitor" "Alice" 29 "Librarian" "Charlie" 19 "Line Cook" }
	//  sheet .drop-column { "name" "age" } ;
	// Tags: #spreadsheet
	"drop-column": {
		Argsn: 2,
		Doc:   "Remove a column from a spreadsheet. Returns new spreadsheet",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch spr := arg0.(type) {
			case env.Spreadsheet:
				switch rmCol := arg1.(type) {
				case env.String:
					return DropColumn(ps, spr, rmCol)
				case env.Block:
					return DropColumnBlock(ps, spr, rmCol)
				default:
					return MakeArgError(ps, 2, []env.Type{env.WordType, env.BlockType}, "drop-column")
				}
			}
			return MakeArgError(ps, 1, []env.Type{env.SpreadsheetType}, "drop-column")
		},
	},
	// Example: Add a column to a sheet
	//  sheet: spreadsheet { "name" "age" } { "Bob" 25 "Alice" 29 "Charlie" 19 }
	//  sheet .add-column! 'job_title { "Jantior" "Librarian" "Line Cook" } ;
	// Tags: #spreadsheet
	"add-column!": {
		Argsn: 4,
		Doc:   "Adds a new column to spreadsheet. Changes in-place and returns the new spreadsheet.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch spr := arg0.(type) {
			case env.Spreadsheet:
				switch newCol := arg1.(type) {
				case env.Word:
					switch fromCols := arg2.(type) {
					case env.Block:
						switch code := arg3.(type) {
						case env.Block:
							return GenerateColumn(ps, spr, newCol, fromCols, code)
						default:
							return MakeArgError(ps, 4, []env.Type{env.BlockType}, "add-column!")
						}
					case env.Word:
						switch replaceBlock := arg3.(type) {
						case env.Block:
							if replaceBlock.Series.Len() != 2 {
								return MakeBuiltinError(ps, "Replacement block must contain a regex object and replacement string.", "add-column!")
							}
							regexNative, ok := replaceBlock.Series.S[0].(env.Native)
							if !ok {
								return MakeBuiltinError(ps, "First element of replacement block must be a regex object.", "add-column!")
							}
							regex, ok := regexNative.Value.(*regexp.Regexp)
							if !ok {
								return MakeBuiltinError(ps, "First element of replacement block must be a regex object.", "add-column!")
							}
							replaceStr, ok := replaceBlock.Series.S[1].(env.String)
							if !ok {
								return MakeBuiltinError(ps, "Second element of replacement block must be a string.", "add-column!")
							}
							err := GenerateColumnRegexReplace(ps, &spr, newCol, fromCols, regex, replaceStr.Value)
							if err != nil {
								return err
							}
							return spr
						default:
							return MakeArgError(ps, 3, []env.Type{env.BlockType}, "add-column!")
						}
					default:
						return MakeArgError(ps, 3, []env.Type{env.BlockType}, "add-column!")
					}
				default:
					return MakeArgError(ps, 2, []env.Type{env.WordType}, "add-column!")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.SpreadsheetType}, "add-column!")
			}
		},
	},
	// Tags: #spreadsheet
	"add-indexes!": {
		Argsn: 2,
		Doc:   "Creates an index for all values in the provided columns. Changes in-place and returns the new spreadsheet.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch spr := arg0.(type) {
			case env.Spreadsheet:
				switch col := arg1.(type) {
				case env.Block:
					colWords := make([]env.Word, col.Series.Len())
					for c := range col.Series.S {
						switch ww := col.Series.S[c].(type) {
						case env.Word:
							colWords[c] = ww
						default:
							return MakeError(ps, "Block of tagwords needed")
						}
					}
					err := AddIndexes(ps, &spr, colWords)
					if err != nil {
						return err
					}
					return spr
				default:
					return MakeArgError(ps, 2, []env.Type{env.BlockType}, "add-indexes!")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.SpreadsheetType}, "add-indexes!")
			}
		},
	},
	// Tags: #spreadsheet
	"indexes?": {
		Argsn: 1,
		Doc:   "Returns the columns that are indexed in a spreadsheet.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) (res env.Object) {
			switch spr := arg0.(type) {
			case env.Spreadsheet:
				res := make([]env.Object, 0)
				for col := range spr.Indexes {
					res = append(res, *env.NewString(col))
				}
				return *env.NewBlock(*env.NewTSeries(res))
			default:
				return MakeArgError(ps, 1, []env.Type{env.SpreadsheetType}, "indexes?")
			}
		},
	},
	// Tags: #spreadsheet
	"autotype": {
		Argsn: 2,
		Doc:   "Takes a spreadsheet and tries to determine and change the types of columns.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) (res env.Object) {
			switch spr := arg0.(type) {
			case env.Spreadsheet:
				switch percent := arg1.(type) {
				case env.Decimal:
					return AutoType(ps, &spr, percent.Value)
				default:
					return MakeArgError(ps, 2, []env.Type{env.DecimalType}, "autotype")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.SpreadsheetType}, "autotype")
			}
		},
	},
	// Example: join two spreadsheets, putting in empty cells if the left sheet doesn't have a value
	//  names: spreadsheet { "id" "name" } { 1 "Paul" 2 "Chani" 3 "Vladimir" } ,
	//  houses: spreadsheet { "id" "house" } { 1 "Atreides" 3 "Harkonnen" } ,
	//  names .left-join houses 'id 'id
	// Tags: #spreadsheet
	"left-join": {
		Argsn: 4,
		Doc:   "Left joins two spreadsheets on the given columns.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) (res env.Object) {
			switch spr1 := arg0.(type) {
			case env.Spreadsheet:
				switch spr2 := arg1.(type) {
				case env.Spreadsheet:
					switch col1 := arg2.(type) {
					case env.Word:
						col2, ok := arg3.(env.Word)
						if !ok {
							return MakeArgError(ps, 4, []env.Type{env.WordType}, "left-join")
						}
						return LeftJoin(ps, spr1, spr2, ps.Idx.GetWord(col1.Index), ps.Idx.GetWord(col2.Index), false)
					case env.String:
						col2, ok := arg3.(env.String)
						if !ok {
							MakeArgError(ps, 4, []env.Type{env.StringType}, "left-join")
						}
						return LeftJoin(ps, spr1, spr2, col1.Value, col2.Value, false)
					default:
						return MakeArgError(ps, 3, []env.Type{env.WordType, env.StringType}, "left-join")
					}
				default:
					return MakeArgError(ps, 2, []env.Type{env.SpreadsheetType}, "left-join")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.SpreadsheetType}, "left-join")
			}
		},
	},
	// Example: join two spreadsheets
	//  names: spreadsheet { "id" "name" } { 1 "Paul" 2 "Chani" 3 "Vladimir" } ,
	//  houses: spreadsheet { "id" "house" } { 1 "Atreides" 3 "Harkonnen" } ,
	//  names .inner-join houses 'id 'id
	// Tags: #spreadsheet
	"inner-join": {
		Argsn: 4,
		Doc:   "Inner joins two spreadsheets on the given columns.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) (res env.Object) {
			switch spr1 := arg0.(type) {
			case env.Spreadsheet:
				switch spr2 := arg1.(type) {
				case env.Spreadsheet:
					switch col1 := arg2.(type) {
					case env.Word:
						col2, ok := arg3.(env.Word)
						if !ok {
							return MakeArgError(ps, 4, []env.Type{env.WordType}, "inner-join")
						}
						return LeftJoin(ps, spr1, spr2, ps.Idx.GetWord(col1.Index), ps.Idx.GetWord(col2.Index), true)
					case env.String:
						col2, ok := arg3.(env.String)
						if !ok {
							MakeArgError(ps, 4, []env.Type{env.StringType}, "inner-join")
						}
						return LeftJoin(ps, spr1, spr2, col1.Value, col2.Value, true)
					default:
						return MakeArgError(ps, 3, []env.Type{env.WordType, env.StringType}, "inner-join")
					}
				default:
					return MakeArgError(ps, 2, []env.Type{env.SpreadsheetType}, "inner-join")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.SpreadsheetType}, "inner-join")
			}
		},
	},
	// Example: group spreadsheet rows by name, runing various aggregations on the val column
	//  spreadsheet { "name" "val" } { "a" 1 "b" 6 "a" 5 "b" 10 "a" 7 }
	// 	|group-by 'name { 'name count 'val sum 'val min 'val max 'val avg }
	// 	|order-by! 'name 'asc
	// Tags: #spreadsheet
	"group-by": {
		Argsn: 3,
		Doc:   "Groups a spreadsheet by the given column and (optional) aggregations.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) (res env.Object) {
			switch spr := arg0.(type) {
			case env.Spreadsheet:
				switch aggBlock := arg2.(type) {
				case env.Block:
					if len(aggBlock.Series.S)%2 != 0 {
						return MakeBuiltinError(ps, "Aggregation block must contain pairs of column name and function for each aggregation.", "group-by")
					}
					aggregations := make(map[string][]string)
					for i := 0; i < len(aggBlock.Series.S); i += 2 {
						col := aggBlock.Series.S[i]
						fun, ok := aggBlock.Series.S[i+1].(env.Word)
						if !ok {
							return MakeBuiltinError(ps, "Aggregation function must be a word", "group-by")
						}
						colStr := ""
						switch col := col.(type) {
						case env.Tagword:
							colStr = ps.Idx.GetWord(col.Index)
						case env.String:
							colStr = col.Value
						default:
							return MakeBuiltinError(ps, "Aggregation column must be a word or string", "group-by")
						}
						funStr := ps.Idx.GetWord(fun.Index)
						aggregations[colStr] = append(aggregations[colStr], funStr)
					}
					switch col := arg1.(type) {
					case env.Word:
						return GroupBy(ps, spr, ps.Idx.GetWord(col.Index), aggregations)
					case env.String:
						return GroupBy(ps, spr, col.Value, aggregations)
					default:
						return MakeArgError(ps, 2, []env.Type{env.WordType, env.StringType}, "group-by")
					}
				default:
					return MakeArgError(ps, 3, []env.Type{env.BlockType}, "group-by")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.SpreadsheetType}, "group-by")
			}
		},
	},
}

func RyeValueToSpreadsheetRow(spr *env.Spreadsheet, obj env.Object) (bool, string, *env.SpreadsheetRow) {
	switch updater := obj.(type) {
	case env.Dict:
		success, missing, row := env.SpreadsheetRowFromDict(updater, spr)
		if !success {
			return false, "update-row! given a dict that is missing value for the " + missing + " column!", nil
		} else {
			return true, "", row

		}
	case env.SpreadsheetRow:
		return true, "", &updater
	default:
		return false, "", nil
	}

}

func DropColumnBlock(ps *env.ProgramState, s env.Spreadsheet, names env.Block) env.Object {
	toDrop := make([]env.String, 0)
	for _, obj := range names.Series.S {
		switch word := obj.(type) {
		case env.Word:
			toDrop = append(toDrop, *env.NewString(ps.Idx.GetWord(word.Index)))
		default:
			return MakeError(ps, "Cannot use a non-word to specify a column to drop")
		}
	}
	return DropColumns(ps, s, toDrop)
}

func DropColumn(ps *env.ProgramState, s env.Spreadsheet, name env.String) env.Object {
	return DropColumns(ps, s, []env.String{name})
}

// Drop one or more columns from a spreadsheet, returning a new spreadsheet
func DropColumns(ps *env.ProgramState, s env.Spreadsheet, names []env.String) env.Object {
	var columnsToCopy []int = make([]int, len(s.Cols)-len(names))
	var keepColIdx int = 0

	for colIdx, col := range s.Cols {
		keep := true
		for _, name := range names {
			nameStr := name.Value
			if col == nameStr {
				keep = false
				break
			}
		}
		if keep {
			columnsToCopy[keepColIdx] = colIdx
			keepColIdx++
		}
	}

	newCols := make([]string, len(columnsToCopy))

	for toIdx, fromIdx := range columnsToCopy {
		newCols[toIdx] = s.Cols[fromIdx]
	}

	newSheet := env.NewSpreadsheet(newCols)
	newSheet.Rows = make([]env.SpreadsheetRow, len(s.Rows))

	for rowIdx, row := range s.Rows {
		newValues := make([]any, len(columnsToCopy))
		for toIdx, fromIdx := range columnsToCopy {
			newValues[toIdx] = row.Values[fromIdx]
		}
		newSheet.Rows[rowIdx] = *env.NewSpreadsheetRow(newValues, newSheet)
	}
	newSheet.Indexes = make(map[string]map[any][]int)

	for _, colName := range newCols {
		newSheet.Indexes[colName] = s.Indexes[colName]
	}
	newSheet.Kind = s.Kind

	return newSheet
}

func GenerateColumn(ps *env.ProgramState, s env.Spreadsheet, name env.Word, extractCols env.Block, code env.Block) env.Object {
	// add name to columns
	s.Cols = append(s.Cols, ps.Idx.GetWord(name.Index))
	for ix, row := range s.Rows {
		// create a empty context connected to current context
		ctx := env.NewEnv(ps.Ctx)
		// for each word in extractCols get a value from current row and set word in context to it
		var firstVal env.Object
		for _, word := range extractCols.Series.S {
			switch w := word.(type) {
			case env.Word:
				val, er := s.GetRowValue(ps.Idx.GetWord(w.Index), row)
				if val == nil {
					val = ""
				}
				// fmt.Println(val)
				if er != nil {
					return MakeBuiltinError(ps, er.Error(), "add-column!")
				}
				if firstVal == nil {
					var ok bool
					firstVal, ok = val.(env.Object)
					if !ok {
						firstVal = *env.NewString(val.(string))
					}
				}
				val1, ok := val.(env.Object)
				if !ok {
					val1 = *env.NewString(val.(string))
				}
				ctx.Set(w.Index, val1)
			}
		}
		// execute the block of code injected with first value
		ser := ps.Ser
		ps.Ser = code.Series
		EvalBlockInCtxInj(ps, ctx, firstVal, firstVal != nil)
		if ps.ErrorFlag {
			return ps.Res
		}
		ps.Ser = ser
		// set the result of code block as the new column value in this row
		// TODO -- make
		row.Values = append(row.Values, ps.Res)
		row.Uplink = &s
		s.Rows[ix] = row
	}
	return s
}

func GenerateColumnRegexReplace(ps *env.ProgramState, s *env.Spreadsheet, name env.Word, fromColName env.Word, re *regexp.Regexp, pattern string) env.Object {
	// add name to columns
	s.Cols = append(s.Cols, ps.Idx.GetWord(name.Index))
	for ix, row := range s.Rows {
		// get value from current row
		val, err := s.GetRowValue(ps.Idx.GetWord(fromColName.Index), row)
		if err != nil {
			return MakeError(ps, fmt.Sprintf("Couldn't retrieve value at row %d (%s)", ix, err))
		}

		var newVal any
		valStr, ok := val.(env.String)
		if !ok {
			newVal = ""
		} else {
			// replace the value with the regex
			newVal = *env.NewString(re.ReplaceAllString(valStr.Value, pattern))
		}
		// set the result of code block as the new column value in this row
		row.Values = append(row.Values, newVal)
		s.Rows[ix] = row
	}
	return nil
}

func AddIndexes(ps *env.ProgramState, s *env.Spreadsheet, columns []env.Word) env.Object {
	s.Indexes = make(map[string]map[any][]int, 0)
	for _, column := range columns {
		colstr := ps.Idx.GetWord(column.Index)
		s.Indexes[colstr] = make(map[any][]int, 0)
		for ir, row := range s.Rows {
			val, err := s.GetRowValue(colstr, row)
			if err != nil {
				return MakeError(ps, "Couldn't retrieve index at row "+strconv.Itoa(ir))
			}
			if subidx, ok := s.Indexes[colstr][val]; ok {
				s.Indexes[colstr][val] = append(subidx, ir)
			} else {
				subidx := []int{ir}
				s.Indexes[colstr][val] = subidx
			}
		}
	}
	return nil
}

func SortByColumn(ps *env.ProgramState, s *env.Spreadsheet, name string) {
	idx := slices.Index[[]string](s.Cols, name)

	compareCol := func(i, j int) bool {
		return greaterThanNew(s.Rows[j].Values[idx].(env.Object), s.Rows[i].Values[idx].(env.Object))
	}

	sort.Slice(s.Rows, compareCol)
}

func SortByColumnDesc(ps *env.ProgramState, s *env.Spreadsheet, name string) {
	idx := slices.Index[[]string](s.Cols, name)

	compareCol := func(i, j int) bool {
		return greaterThanNew(s.Rows[i].Values[idx].(env.Object), s.Rows[j].Values[idx].(env.Object))
	}

	sort.Slice(s.Rows, compareCol)
}

func WhereEquals(ps *env.ProgramState, s env.Spreadsheet, name string, val env.Object) env.Object {
	idx := slices.Index(s.Cols, name)
	nspr := env.NewSpreadsheet(s.Cols)
	if idx > -1 {
		if index, ok := s.Indexes[name]; ok {
			idxs := index[val]
			for _, idx := range idxs {
				nspr.AddRow(s.Rows[idx])
			}
		} else {
			for _, row := range s.Rows {
				if len(row.Values) > idx {
					if val.Equal(env.ToRyeValue(row.Values[idx])) {
						nspr.AddRow(row)
					}
				}
			}
		}
		return *nspr
	} else {
		return MakeBuiltinError(ps, "Column not found.", "WhereEquals")
	}
}

func WhereMatch(ps *env.ProgramState, s env.Spreadsheet, name string, r *regexp.Regexp) env.Object {
	idx := slices.Index(s.Cols, name)
	nspr := env.NewSpreadsheet(s.Cols)
	if idx > -1 {
		for _, row := range s.Rows {
			if len(row.Values) > idx {
				rv := row.Values[idx]
				if rvStr, ok := rv.(env.String); ok {
					if r.MatchString(rvStr.Value) {
						nspr.AddRow(row)
					}
				}
			}
		}
		return *nspr
	} else {
		return MakeBuiltinError(ps, "Column not found.", "WhereMatch")
	}
}

func WhereContains(ps *env.ProgramState, s env.Spreadsheet, name string, val string, not bool) env.Object {
	idx := slices.Index(s.Cols, name)
	nspr := env.NewSpreadsheet(s.Cols)
	if idx > -1 {
		for _, row := range s.Rows {
			if len(row.Values) > idx {
				rv := row.Values[idx]
				if rvStr, ok := rv.(env.String); ok {
					if !not {
						if strings.Contains(rvStr.Value, val) {
							nspr.AddRow(row)
						}
					} else {
						if !strings.Contains(rvStr.Value, val) {
							nspr.AddRow(row)
						}
					}
				}
			}
		}
		return *nspr
	} else {
		return MakeBuiltinError(ps, "Column not found.", "WhereMatch")
	}
}

func WhereIn(ps *env.ProgramState, s env.Spreadsheet, name string, b []env.Object) env.Object {
	idx := slices.Index(s.Cols, name)
	nspr := env.NewSpreadsheet(s.Cols)
	if idx > -1 {
		for _, row := range s.Rows {
			if len(row.Values) > idx {
				rv := row.Values[idx]
				if rvObj, ok := rv.(env.Object); ok {
					if util.ContainsVal(ps, b, rvObj) {
						nspr.AddRow(row)
					}
				}
			}
		}
		return *nspr
	} else {
		return MakeBuiltinError(ps, "Column not found.", "WhereIn")
	}
}

func WhereGreater(ps *env.ProgramState, s env.Spreadsheet, name string, val env.Object) env.Object {
	idx := slices.Index(s.Cols, name)
	nspr := env.NewSpreadsheet(s.Cols)
	if idx > -1 {
		for _, row := range s.Rows {
			if len(row.Values) > idx {
				if greaterThanNew(row.Values[idx].(env.Object), val) {
					nspr.AddRow(row)
				}
			}
		}
		return *nspr
	} else {
		return MakeBuiltinError(ps, "Column not found.", "WhereGreater")
	}
}

func WhereLesser(ps *env.ProgramState, s env.Spreadsheet, name string, val env.Object) env.Object {
	idx := slices.Index(s.Cols, name)
	nspr := env.NewSpreadsheet(s.Cols)
	if idx > -1 {
		for _, row := range s.Rows {
			if len(row.Values) > idx {
				if lesserThanNew(row.Values[idx].(env.Object), val) {
					nspr.AddRow(row)
				}
			}
		}
		return *nspr
	} else {
		return MakeBuiltinError(ps, "Column not found.", "WhereGreater")
	}
}

func WhereBetween(ps *env.ProgramState, s env.Spreadsheet, name string, val1 env.Object, val2 env.Object) env.Object {
	idx := slices.Index(s.Cols, name)
	nspr := env.NewSpreadsheet(s.Cols)
	if idx > -1 {
		for _, row := range s.Rows {
			if len(row.Values) > idx {
				rv := row.Values[idx].(env.Object)
				if greaterThanNew(rv, val1) && lesserThanNew(rv, val2) {
					nspr.AddRow(row)
				}
			}
		}
		return *nspr
	} else {
		return MakeBuiltinError(ps, "Column not found.", "WhereBetween")
	}
}

func Limit(ps *env.ProgramState, s env.Spreadsheet, n int) env.Object {
	nspr := env.NewSpreadsheet(s.Cols)
	nspr.Rows = s.Rows[0:n]
	return *nspr
}

func AutoType(ps *env.ProgramState, s *env.Spreadsheet, percent float64) env.Object {
	colTypeCount := make(map[int]map[string]int)
	for i := range s.Cols {
		colTypeCount[i] = make(map[string]int)
	}
	for _, row := range s.Rows {
		for i, val := range row.Values {
			switch stringVal := val.(type) {
			case env.String:
				if _, err := strconv.Atoi(stringVal.Value); err == nil {
					colTypeCount[i]["int"]++
				} else if _, err = strconv.ParseFloat(stringVal.Value, 64); err == nil {
					colTypeCount[i]["dec"]++
				} else {
					colTypeCount[i]["str"]++
				}
			case env.Integer:
				colTypeCount[i]["int"]++
			case env.Decimal:
				colTypeCount[i]["dec"]++
			default:
				continue
			}
		}
	}

	lenRows := len(s.Rows)
	newS := env.NewSpreadsheet(s.Cols)
	for range s.Rows {
		newRow := make([]any, len(s.Cols))
		newS.AddRow(*env.NewSpreadsheetRow(newRow, newS))
	}

	for colNum, typeCount := range colTypeCount {
		minRows := int(float64(lenRows) * percent)
		var newType string
		// if there's a mix of floats and ints, make it a float
		if typeCount["dec"] > 0 && typeCount["dec"]+typeCount["int"] >= minRows {
			newType = "dec"
		} else if typeCount["int"] >= minRows {
			newType = "int"
		} else {
			newType = "str"
		}
		for i, row := range s.Rows {
			switch newType {
			case "int":
				switch val := row.Values[colNum].(type) {
				case env.String:
					intVal, _ := strconv.Atoi(val.Value)
					newS.Rows[i].Values[colNum] = *env.NewInteger(int64(intVal))
				case env.Integer:
					//intVal, _ := strconv.Atoi(row.Values[colNum].(env.String).Value)
					newS.Rows[i].Values[colNum] = val
				case env.Decimal:
					//intVal, _ := strconv.Atoi(row.Values[colNum].(env.String).Value)
					newS.Rows[i].Values[colNum] = val
				}
			case "dec":
				switch val1 := row.Values[colNum].(type) {
				case env.String:
					floatVal, _ := strconv.ParseFloat(val1.Value, 64)
					newS.Rows[i].Values[colNum] = *env.NewDecimal(floatVal)
				case env.Integer:
					//intVal, _ := strconv.Atoi(row.Values[colNum].(env.String).Value)
					//newS.Rows[i].Values[colNum] = *env.NewInteger(int64(intVal))
					newS.Rows[i].Values[colNum] = *env.NewDecimal(float64(val1.Value))
				case env.Decimal:
					//intVal, _ := strconv.Atoi(row.Values[colNum].(env.String).Value)
					//newS.Rows[i].Values[colNum] = *env.NewInteger(int64(intVal))
					newS.Rows[i].Values[colNum] = val1
				}
			case "str":
				newS.Rows[i].Values[colNum] = row.Values[colNum]
			}
		}
	}

	return *newS
}

func LeftJoin(ps *env.ProgramState, s1 env.Spreadsheet, s2 env.Spreadsheet, col1 string, col2 string, innerJoin bool) env.Object {
	if !slices.Contains(s1.Cols, col1) {
		return MakeBuiltinError(ps, "Column not found in first spreadsheet.", "left-join")
	}
	if !slices.Contains(s2.Cols, col2) {
		return MakeBuiltinError(ps, "Column not found in second spreadsheet.", "left-join")
	}

	combinedCols := make([]string, len(s1.Cols)+len(s2.Cols))
	copy(combinedCols, s1.Cols)
	for i, v := range s2.Cols {
		if slices.Contains(combinedCols, v) {
			combinedCols[i+len(s1.Cols)] = v + "_2"
		} else {
			combinedCols[i+len(s1.Cols)] = v
		}
	}
	nspr := env.NewSpreadsheet(combinedCols)
	for i, row1 := range s1.GetRows() {
		val1, err := s1.GetRowValue(col1, row1)
		if err != nil {
			return MakeError(ps, fmt.Sprintf("Couldn't retrieve value at row %d (%s)", i, err))
		}

		// the row ids of the second spreadsheet which match the values in the current first spreadsheet row
		var s2RowIds []int
		// use index if available
		if ix, ok := s2.Indexes[col2]; ok {
			if rowIds, ok := ix[val1]; ok {
				s2RowIds = rowIds
			}
		} else {
			for j, row2 := range s2.GetRows() {
				val2, err := s2.GetRowValue(col2, row2)
				if err != nil {
					return MakeError(ps, fmt.Sprintf("Couldn't retrieve value at row %d (%s)", j, err))
				}
				val1o, ok := val1.(env.Object)
				if ok {
					val2o, ok1 := val2.(env.Object)
					if ok1 {
						if val1o.Equal(val2o) {
							s2RowIds = append(s2RowIds, j)
							continue
						}
					} else {
						if env.RyeToRaw(val1o, ps.Idx) == val2 {
							s2RowIds = append(s2RowIds, j)
							continue
						}
					}
				}
				val1s, ok := val1.(string)
				if ok {
					if val1s == val2.(string) {
						s2RowIds = append(s2RowIds, j)
						continue
					}
				}
			}
		}
		if innerJoin && len(s2RowIds) == 0 {
			continue
		}
		buildCombinedRow := func(row1Values []any, row2Values []any) []any {
			newRow := make([]any, len(combinedCols))
			copy(newRow, row1Values)

			if row2Values != nil {
				// copy values from second spreadsheet
				for i, v := range row2Values {
					newRow[i+len(s1.Cols)] = v
				}
			} else {
				// fill with Void{} values when no match
				for i := len(s1.Cols); i < len(combinedCols); i++ {
					newRow[i] = env.Void{}
				}
			}
			return newRow
		}

		if len(s2RowIds) > 0 {
			// add a row for each matching record from s2
			for _, s2RowId := range s2RowIds {
				row2Values := s2.GetRow(ps, s2RowId).Values
				combinedValues := buildCombinedRow(row1.Values, row2Values)
				nspr.AddRow(*env.NewSpreadsheetRow(combinedValues, nspr))
			}
		} else {
			// add a single row with Void values for s2 columns
			combinedValues := buildCombinedRow(row1.Values, nil)
			nspr.AddRow(*env.NewSpreadsheetRow(combinedValues, nspr))
		}
	}
	return *nspr
}

func GroupBy(ps *env.ProgramState, s env.Spreadsheet, col string, aggregations map[string][]string) env.Object {
	if !slices.Contains(s.Cols, col) {
		return MakeBuiltinError(ps, "Column not found.", "group-by")
	}

	aggregatesByGroup := make(map[string]map[string]float64)
	countByGroup := make(map[string]int)
	for i, row := range s.Rows {
		groupingVal, err := s.GetRowValue(col, row)
		if err != nil {
			return MakeError(ps, fmt.Sprintf("Couldn't retrieve value at row %d (%s)", i, err))
		}
		var groupValStr string
		switch val := groupingVal.(type) {
		case env.String:
			groupValStr = val.Value
		case string:
			groupValStr = val
		case env.Integer:
			groupValStr = strconv.Itoa(int(val.Value))
		case int:
			groupValStr = strconv.Itoa(val)
		default:
			return MakeBuiltinError(ps, "Grouping column value must be a string", "group-by")
		}

		if _, ok := aggregatesByGroup[groupValStr]; !ok {
			aggregatesByGroup[groupValStr] = make(map[string]float64)
		}
		groupAggregates := aggregatesByGroup[groupValStr]

		for aggCol, funs := range aggregations {
			for _, fun := range funs {
				colAgg := aggCol + "_" + fun
				if fun == "count" {
					if aggCol != col {
						return MakeBuiltinError(ps, "Count aggregation can only be applied on the grouping column", "group-by")
					}
					groupAggregates[colAgg]++
					continue
				}
				valObj, err := s.GetRowValue(aggCol, row)
				if err != nil {
					return MakeError(ps, fmt.Sprintf("Couldn't retrieve value at row %d (%s)", i, err))
				}
				var val float64
				switch valObj := env.ToRyeValue(valObj).(type) {
				case env.Integer:
					val = float64(valObj.Value)
				case env.Decimal:
					val = valObj.Value
				default:
					return MakeBuiltinError(ps, "Aggregation column value must be a number", "group-by")
				}
				switch fun {
				case "sum":
					groupAggregates[colAgg] += val
				case "avg":
					groupAggregates[colAgg] += val
					countByGroup[groupValStr]++
				case "min":
					if min, ok := groupAggregates[colAgg]; !ok || val < min {
						groupAggregates[colAgg] = val
					}
				case "max":
					if max, ok := groupAggregates[colAgg]; !ok || val > max {
						groupAggregates[colAgg] = val
					}
				default:
					return MakeBuiltinError(ps, fmt.Sprintf("Unknown aggregation function: %s", fun), "group-by")
				}
			}
		}
	}
	newCols := []string{col}
	for aggCol, funs := range aggregations {
		for _, fun := range funs {
			newCols = append(newCols, aggCol+"_"+fun)
		}
	}
	newS := env.NewSpreadsheet(newCols)
	for groupVal, groupAggregates := range aggregatesByGroup {
		newRow := make([]any, len(newCols))
		newRow[0] = *env.NewString(groupVal)
		for i, col := range newCols[1:] {
			if strings.HasSuffix(col, "_count") {
				newRow[i+1] = *env.NewInteger(int64(groupAggregates[col]))
			} else if strings.HasSuffix(col, "_avg") {
				newRow[i+1] = *env.NewDecimal(groupAggregates[col] / float64(countByGroup[groupVal]))
			} else {
				newRow[i+1] = *env.NewDecimal(groupAggregates[col])
			}
		}
		newS.AddRow(*env.NewSpreadsheetRow(newRow, newS))
	}
	return *newS
}
