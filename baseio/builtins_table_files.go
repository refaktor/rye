//go:build !no_table

package baseio

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"

	"github.com/refaktor/rye/env"
	"github.com/refaktor/rye/evaldo"
	"github.com/xuri/excelize/v2"
)

// Builtins_table_files contains table formats that read or write file URIs.
var Builtins_table_files = map[string]*env.Builtin{
	//
	// ##### Table files #####  "Load and save tables as CSV, TSV or XLSX files."
	//
	// Example:
	//  equal {
	//	 cc os
	//   f:: mktmp ++ "/test.csv"
	//   spr1:: table { "a" "b" "c" } { 1 1.1 "a" 2 2.2 "b" 3 3.3 "c" }
	//   spr1 .Save\csv* f
	//   spr2:: Load\csv f |autotype 1.0
	//   spr1 = spr2
	//  } true
	// Args:
	// * file-uri - location of csv file to load
	// Tags: #table #loading #csv
	"file-uri//Load\\csv": {
		// TODO 2 -- this could move to a go function so it could be called by general load that uses extension to define the loader
		Argsn: 1,
		Doc:   "Loads a .csv file to a table datatype.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch file := arg0.(type) {
			case env.Uri:
				// rows, err := db1.Value.(*sql.DB).Query(sqlstr, vals...)
				f, err := os.Open(file.GetPath())
				if err != nil {
					// log.Fatal("Unable to read input file "+filePath, err)
					return evaldo.MakeBuiltinError(ps, "Unable to read input file:"+err.Error(), "Load\\csv")
				}
				defer f.Close()

				csvReader := csv.NewReader(f)
				rows, err := csvReader.ReadAll()
				if err != nil {
					// log.Fatal("Unable to parse file as CSV for "+filePath, err)
					return evaldo.MakeBuiltinError(ps, "Unable to parse file as CSV: "+err.Error(), "Load\\csv")
				}
				if len(rows) == 0 {
					return evaldo.MakeBuiltinError(ps, "File is empty", "Load\\csv")
				}
				spr := env.NewTable(rows[0])
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
						spr.AddRow(*env.NewTableRow(anyRow, spr))
					}
				}
				return *spr
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.UriType}, "Load\\csv")
			}
		},
	},
	// Example:
	//  equal {
	//	 cc os
	//   f:: mktmp ++ "/test.csv"
	//   spr1:: table { "a" "b" "c" } { 1 1.1 "a" 2 2.2 "b" 3 3.3 "c" }
	//   spr1 .Save\csv* f
	//   spr2:: Load\csv f |autotype 1.0
	//   spr1 = spr2
	//  } true
	// Args:
	// * file-uri - location of csv file to load
	// Tags: #table #loading #csv
	"file-uri//Load\\csv\\": {
		// TODO 2 -- this could move to a go function so it could be called by general load that uses extension to define the loader
		Argsn: 2,
		Doc:   "Loads a .csv file to a table datatype.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch file := arg0.(type) {
			case env.Uri:
				switch separator := arg1.(type) {
				case env.String:
					// rows, err := db1.Value.(*sql.DB).Query(sqlstr, vals...)
					f, err := os.Open(file.GetPath())
					if err != nil {
						// log.Fatal("Unable to read input file "+filePath, err)
						return evaldo.MakeBuiltinError(ps, "Unable to read input file:"+err.Error(), "Load\\csv")
					}
					defer f.Close()

					csvReader := csv.NewReader(f)
					if len(separator.Value) != 1 {
						return evaldo.MakeBuiltinError(ps, "Separator must be exactly 1 character long", "Load\\csv\\")
					}
					csvReader.Comma = rune(separator.Value[0])
					rows, err := csvReader.ReadAll()
					if err != nil {
						// log.Fatal("Unable to parse file as CSV for "+filePath, err)
						return evaldo.MakeBuiltinError(ps, "Unable to parse file as CSV: "+err.Error(), "Load\\csv")
					}
					if len(rows) == 0 {
						return evaldo.MakeBuiltinError(ps, "File is empty", "Load\\csv")
					}
					spr := env.NewTable(rows[0])
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
							spr.AddRow(*env.NewTableRow(anyRow, spr))
						}
					}
					return *spr
				default:
					return evaldo.MakeArgError(ps, 2, []env.Type{env.StringType}, "Load\\csv\\")
				}
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.UriType}, "Load\\csv\\")
			}
		},
	},

	// TODO -- deduplicate with above

	// Example:
	//  equal {
	//	 cc os
	//   f:: mktmp ++ "/test.tsv"
	//   spr1:: table { "a" "b" "c" } { 1 1.1 "a" 2 2.2 "b" 3 3.3 "c" }
	//   spr1 .Save\tsv* f
	//   spr2:: Load\tsv f |autotype 1.0
	//   spr1 = spr2
	//  } true
	// Args:
	// * file-uri - location of csv file to load
	"file-uri//Load\\tsv": {
		// TODO 2 -- this could move to a go function so it could be called by general load that uses extension to define the loader
		Argsn: 1,
		Doc:   "Loads a .csv file to a table datatype.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch file := arg0.(type) {
			case env.Uri:
				// rows, err := db1.Value.(*sql.DB).Query(sqlstr, vals...)
				f, err := os.Open(file.GetPath())
				if err != nil {
					// log.Fatal("Unable to read input file "+filePath, err)
					return evaldo.MakeBuiltinError(ps, "Unable to read input file:"+err.Error(), "Load\\csv")
				}
				defer f.Close()

				csvReader := csv.NewReader(f)
				csvReader.Comma = '\t'
				rows, err := csvReader.ReadAll()
				if err != nil {
					// log.Fatal("Unable to parse file as CSV for "+filePath, err)
					return evaldo.MakeBuiltinError(ps, "Unable to parse file as CSV: "+err.Error(), "Load\\csv")
				}
				if len(rows) == 0 {
					return evaldo.MakeBuiltinError(ps, "File is empty", "Load\\csv")
				}
				spr := env.NewTable(rows[0])
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
						spr.AddRow(*env.NewTableRow(anyRow, spr))
					}
				}
				return *spr
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.UriType}, "Load\\csv")
			}
		},
	},

	// Example:
	//  equal {
	//	 cc os
	//   f:: mktmp ++ "/test.csv"
	//   spr1:: table { "a" "b" "c" } { 1 1.1 "a" 2 2.2 "b" 3 3.3 "c" }
	//   f .Save\csv spr1
	//   spr2:: Load\csv f |autotype 1.0
	//   spr1 = spr2
	//  } true
	// Args:
	// * file-uri - where to save the sheet as a .csv file
	// * table    - the table to save
	// Tags: #table #saving #csv
	"file-uri//Save\\csv": {
		Argsn: 2,
		Doc:   "Saves a table to a .csv file.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch file := arg0.(type) {
			case env.Uri:
				switch spr := arg1.(type) {
				case env.Table:
					// rows, err := db1.Value.(*sql.DB).Query(sqlstr, vals...)
					f, err := os.Create(file.GetPath())
					if err != nil {
						// log.Fatal("Unable to read input file "+filePath, err)
						return evaldo.MakeBuiltinError(ps, "Unable to create input file.", "file-uri//Save\\csv")
					}
					defer f.Close()

					cLen := len(spr.Cols)

					csvWriter := csv.NewWriter(f)

					err1 := csvWriter.Write(spr.Cols)
					if err1 != nil {
						return evaldo.MakeBuiltinError(ps, "Unable to create write header.", "file-uri//Save\\csv")
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
							if i < cLen {
								strVals[i] = sv
							}
						}
						err := csvWriter.Write(strVals)
						if err != nil {
							return evaldo.MakeBuiltinError(ps, "Unable to write line: "+strconv.Itoa(ir), "file-uri//Save\\csv")
						}
					}
					csvWriter.Flush()
					f.Close()
					return spr
				default:
					return evaldo.MakeArgError(ps, 2, []env.Type{env.TableType}, "file-uri//Save\\csv")
				}
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.UriType}, "file-uri//Save\\csv")
			}
		},
	},

	// TODO: deduplicate with load\csv

	// Example:
	//  equal {
	//	 cc os
	//   f:: mktmp ++ "/test.tsv"
	//   spr1:: table { "a" "b" "c" } { 1 1.1 "a" 2 2.2 "b" 3 3.3 "c" }
	//   f .Save\tsv spr1
	//   spr2:: Load\tsv f |autotype 1.0
	//   spr1 = spr2
	//  } true
	// Args:
	// * file-uri - where to save the sheet as a .tsv file
	// * table    - the table to save
	// Tags: #table #saving #tsv
	"file-uri//Save\\tsv": {
		Argsn: 2,
		Doc:   "Saves a table to a .tsv file.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch file := arg0.(type) {
			case env.Uri:
				switch spr := arg1.(type) {
				case env.Table:
					// rows, err := db1.Value.(*sql.DB).Query(sqlstr, vals...)
					f, err := os.Create(file.GetPath())
					if err != nil {
						// log.Fatal("Unable to read input file "+filePath, err)
						return evaldo.MakeBuiltinError(ps, "Unable to create input file.", "file-uri//Save\\tsv")
					}
					defer f.Close()

					cLen := len(spr.Cols)

					csvWriter := csv.NewWriter(f)
					csvWriter.Comma = '\t'
					err1 := csvWriter.Write(spr.Cols)
					if err1 != nil {
						return evaldo.MakeBuiltinError(ps, "Unable to create write header.", "file-uri//Save\\tsv")
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
							if i < cLen {
								strVals[i] = sv
							}
						}
						err := csvWriter.Write(strVals)
						if err != nil {
							return evaldo.MakeBuiltinError(ps, "Unable to write line: "+strconv.Itoa(ir), "file-uri//Save\\tsv")
						}
					}
					csvWriter.Flush()
					f.Close()
					return spr
				default:
					return evaldo.MakeArgError(ps, 2, []env.Type{env.TableType}, "file-uri//Save\\tsv")
				}
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.UriType}, "file-uri//Save\\tsv")
			}
		},
	},

	// Example:
	//  equal {
	//	 cc os
	//   f:: mktmp ++ "/test.xlsx"
	//   spr1:: table { "a" "b" "c" } { 1 1.1 "a" 2 2.2 "b" 3 3.3 "c" }
	//   spr1 .Save\xlsx* f
	//   spr2:: Load\xlsx f |autotype 1.0
	//   spr1 = spr2
	//  } true
	// Args:
	// * file-uri - location of xlsx file to load
	// Tags: #table #loading #xlsx
	"file-uri//Load\\xlsx": {
		Argsn: 1,
		Doc:   "Loads the first sheet in an .xlsx file to a Table.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch file := arg0.(type) {
			case env.Uri:
				f, err := excelize.OpenFile(file.GetPath())
				if err != nil {
					return evaldo.MakeBuiltinError(ps, fmt.Sprintf("Unable to open file: %s", err), "Load\\xlsx")
				}
				defer f.Close()

				sheetMap := f.GetSheetMap()
				if len(sheetMap) == 0 {
					return evaldo.MakeBuiltinError(ps, "No sheets found in file", "Load\\xlsx")
				}
				// sheets map index is 1-based
				sheetName := sheetMap[1]
				rows, err := f.Rows(sheetName)
				if err != nil {
					return evaldo.MakeBuiltinError(ps, fmt.Sprintf("Unable to get rows from sheet: %s", err), "Load\\xlsx")
				}
				rows.Next()
				header, err := rows.Columns()
				if err != nil {
					return evaldo.MakeBuiltinError(ps, fmt.Sprintf("Unable to get columns from sheet: %s", err), "Load\\xlsx")
				}
				if len(header) == 0 {
					return evaldo.MakeBuiltinError(ps, "Header row is empty", "Load\\xlsx")
				}
				spr := env.NewTable(header)
				for rows.Next() {
					row, err := rows.Columns()
					if err != nil {
						return evaldo.MakeBuiltinError(ps, fmt.Sprintf("Unable to get row: %s", err), "Load\\xlsx")
					}
					anyRow := make([]any, len(row))
					for i, v := range row {
						anyRow[i] = *env.NewString(v)
					}
					// fill in any missing columns with empty strings
					for i := len(row); i < len(spr.Cols); i++ {
						anyRow[i] = *env.NewString("")
					}
					spr.AddRow(*env.NewTableRow(anyRow, spr))
				}
				return *spr
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.UriType}, "Load\\xlsx")
			}
		},
	},

	// Example:
	//  equal {
	//	 cc os
	//   f:: mktmp ++ "/test.xlsx"
	//   spr1:: table { "a" "b" "c" } { 1 1.1 "a" 2 2.2 "b" 3 3.3 "c" }
	//   f .Save\xlsx spr1
	//   spr2:: Load\xlsx f |autotype 1.0
	//   spr1 = spr2
	//  } true
	// Args:
	// * file-uri - where to save the table as a .xlsx file
	// * table    - the table to save
	// Tags: #table #saving #xlsx
	"file-uri//Save\\xlsx": {
		Argsn: 2,
		Doc:   "Saves a Table to a .xlsx file.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch file := arg0.(type) {
			case env.Uri:
				switch spr := arg1.(type) {
				case env.Table:
					sheetName := "Sheet1"
					f := excelize.NewFile()
					index, err := f.NewSheet(sheetName)
					if err != nil {
						return evaldo.MakeBuiltinError(ps, fmt.Sprintf("Unable to create new sheet: %s", err), "file-uri//Save\\xlsx")
					}
					err = f.SetSheetRow(sheetName, "A1", &spr.Cols)
					if err != nil {
						return evaldo.MakeBuiltinError(ps, fmt.Sprintf("Unable to set header row: %s", err), "file-uri//Save\\xlsx")
					}
					for i, row := range spr.Rows {
						// 1-based and skip header row
						rowIndex := i + 2
						vals := make([]any, len(row.Values))
						for j, v := range row.Values {
							switch val := v.(type) {
							case env.String:
								vals[j] = val.Value
							case string:
								vals[j] = val
							case env.Integer:
								vals[j] = val.Value
							case int64:
								vals[j] = val
							case env.Decimal:
								vals[j] = val.Value
							case float64:
								vals[j] = val
							default:
								return evaldo.MakeBuiltinError(ps, fmt.Sprintf("Unable to save table: unsupported type %T", val), "file-uri//Save\\xlsx")
							}
						}
						err = f.SetSheetRow(sheetName, fmt.Sprintf("A%d", rowIndex), &vals)
						if err != nil {
							return evaldo.MakeBuiltinError(ps, fmt.Sprintf("Unable to set row %d: %s", rowIndex, err), "file-uri//Save\\xlsx")
						}
					}
					f.SetActiveSheet(index)
					err = f.SaveAs(file.GetPath())
					if err != nil {
						return evaldo.MakeBuiltinError(ps, fmt.Sprintf("Unable to save table: %s", err), "file-uri//Save\\xlsx")
					}
					return spr
				default:
					return evaldo.MakeArgError(ps, 2, []env.Type{env.TableType}, "file-uri//Save\\xlsx")
				}
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.UriType}, "file-uri//Save\\xlsx")
			}
		},
	},
}
