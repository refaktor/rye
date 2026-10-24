//go:build !no_table

package baseio

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/refaktor/rye/env"
	"github.com/refaktor/rye/evaldo"
	"github.com/xuri/excelize/v2"
)

// newBomReader wraps the file in a buffered reader and removes a leading UTF-8 BOM.
// Some CSV/TSV files have a BOM which confuses csv encoding.
func newBomReader(f *os.File) *bufio.Reader {
	br := bufio.NewReader(f)
	if b, err := br.Peek(3); err == nil && len(b) == 3 && b[0] == 0xEF && b[1] == 0xBB && b[2] == 0xBF {
		_, _ = br.Discard(3)
	}
	return br
}

// newCsvReader returns a csv.Reader that skips a leading UTF-8 BOM before parsing.
func newCsvReader(f *os.File) *csv.Reader {
	return csv.NewReader(newBomReader(f))
}

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
	// Returns:
	// * Table parsed from the CSV file.
	"file-uri//Load\\csv__OLD": {
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
					return evaldo.MakeBuiltinError(ps, "Unable to read input file '"+file.GetPath()+"': "+err.Error(), "Load\\csv")
				}
				defer f.Close()

				csvReader := newCsvReader(f)
				rows, err := csvReader.ReadAll()
				if err != nil {
					// log.Fatal("Unable to parse file as CSV for "+filePath, err)
					return evaldo.MakeBuiltinError(ps, "Unable to parse file '"+file.GetPath()+"' as CSV: "+err.Error(), "Load\\csv")
				}
				if len(rows) == 0 {
					return evaldo.MakeBuiltinError(ps, "File '"+file.GetPath()+"' is empty", "Load\\csv")
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
	// Returns:
	// * Table parsed from the CSV file.
	"file-uri//Load\\csv": {
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
						return evaldo.MakeBuiltinError(ps, "Unable to read input file '"+file.GetPath()+"': "+err.Error(), "Load\\csv")
					}
					defer f.Close()

					csvReader := newCsvReader(f)
					if len(separator.Value) != 1 {
						return evaldo.MakeBuiltinError(ps, "Separator must be exactly 1 character long", "Load\\csv\\")
					}
					csvReader.Comma = rune(separator.Value[0])
					rows, err := csvReader.ReadAll()
					if err != nil {
						// log.Fatal("Unable to parse file as CSV for "+filePath, err)
						return evaldo.MakeBuiltinError(ps, "Unable to parse file '"+file.GetPath()+"' as CSV: "+err.Error(), "Load\\csv")
					}
					if len(rows) == 0 {
						return evaldo.MakeBuiltinError(ps, "File '"+file.GetPath()+"' is empty", "Load\\csv")
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

	// Example:
	//  equal {
	//	 cc os
	//   f:: mktmp ++ "/test.csv"
	//   f .Write "a;b;c\n1;1.1;a\nbad\"line\n2;2.2;b\n3;3.3;c\n"
	//   spr:: Load\csv\ignore-bad f ";"
	//   spr .length? 3
	//  } true
	// Args:
	// * file-uri - location of csv file to load
	// * separator - single character used as column separator
	// Tags: #table #loading #csv
	// Returns:
	// * Table parsed from the CSV file, skipping malformed lines.
	"file-uri//Load\\csv\\ignore-bad": {
		Argsn: 2,
		Doc:   "Loads a .csv file to a table datatype, parsing it line by line and skipping lines that can't be parsed as CSV.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch file := arg0.(type) {
			case env.Uri:
				switch separator := arg1.(type) {
				case env.String:
					if len(separator.Value) != 1 {
						return evaldo.MakeBuiltinError(ps, "Separator must be exactly 1 character long", "Load\\csv\\ignore-bad")
					}
					comma := rune(separator.Value[0])

					f, err := os.Open(file.GetPath())
					if err != nil {
						return evaldo.MakeBuiltinError(ps, "Unable to read input file '"+file.GetPath()+"': "+err.Error(), "Load\\csv\\ignore-bad")
					}
					defer f.Close()

					scanner := bufio.NewScanner(newBomReader(f))
					scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
					var spr *env.Table
					for scanner.Scan() {
						r := csv.NewReader(strings.NewReader(scanner.Text()))
						r.Comma = comma
						record, err := r.Read()
						if err != nil {
							// skip lines that can't be parsed as a single CSV record
							continue
						}
						if spr == nil {
							spr = env.NewTable(record)
							continue
						}
						anyRow := make([]any, len(record))
						for i, v := range record {
							anyRow[i] = *env.NewString(v)
						}
						spr.AddRow(*env.NewTableRow(anyRow, spr))
					}
					if err := scanner.Err(); err != nil {
						return evaldo.MakeBuiltinError(ps, "Unable to read input file '"+file.GetPath()+"': "+err.Error(), "Load\\csv\\ignore-bad")
					}
					if spr == nil {
						return evaldo.MakeBuiltinError(ps, "No parseable lines found in '"+file.GetPath()+"'", "Load\\csv\\ignore-bad")
					}
					return *spr
				default:
					return evaldo.MakeArgError(ps, 2, []env.Type{env.StringType}, "Load\\csv\\ignore-bad")
				}
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.UriType}, "Load\\csv\\ignore-bad")
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
	// Returns:
	// * Table parsed from the TSV file.
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
					return evaldo.MakeBuiltinError(ps, "Unable to read input file '"+file.GetPath()+"': "+err.Error(), "Load\\csv")
				}
				defer f.Close()

				csvReader := newCsvReader(f)
				csvReader.Comma = '\t'
				rows, err := csvReader.ReadAll()
				if err != nil {
					// log.Fatal("Unable to parse file as CSV for "+filePath, err)
					return evaldo.MakeBuiltinError(ps, "Unable to parse file '"+file.GetPath()+"' as CSV: "+err.Error(), "Load\\csv")
				}
				if len(rows) == 0 {
					return evaldo.MakeBuiltinError(ps, "File '"+file.GetPath()+"' is empty", "Load\\csv")
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
	// Returns:
	// * the saved file URI.
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
	// Returns:
	// * the saved file URI.
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
	// Returns:
	// * Table containing the first sheet of the XLSX file.
	"file-uri//Load\\xlsx": {
		Argsn: 1,
		Doc:   "Loads the first sheet in an .xlsx file to a Table.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch file := arg0.(type) {
			case env.Uri:
				f, err := excelize.OpenFile(file.GetPath())
				if err != nil {
					return evaldo.MakeBuiltinError(ps, fmt.Sprintf("Unable to open file '%s': %s", file.GetPath(), err), "Load\\xlsx")
				}
				defer f.Close()

				sheetMap := f.GetSheetMap()
				if len(sheetMap) == 0 {
					return evaldo.MakeBuiltinError(ps, fmt.Sprintf("No sheets found in file '%s'", file.GetPath()), "Load\\xlsx")
				}
				// sheets map index is 1-based
				sheetName := sheetMap[1]
				rows, err := f.Rows(sheetName)
				if err != nil {
					return evaldo.MakeBuiltinError(ps, fmt.Sprintf("Unable to get rows from sheet in '%s': %s", file.GetPath(), err), "Load\\xlsx")
				}
				rows.Next()
				header, err := rows.Columns()
				if err != nil {
					return evaldo.MakeBuiltinError(ps, fmt.Sprintf("Unable to get columns from sheet in '%s': %s", file.GetPath(), err), "Load\\xlsx")
				}
				if len(header) == 0 {
					return evaldo.MakeBuiltinError(ps, fmt.Sprintf("Header row is empty in '%s'", file.GetPath()), "Load\\xlsx")
				}
				spr := env.NewTable(header)
				for rows.Next() {
					row, err := rows.Columns()
					if err != nil {
						return evaldo.MakeBuiltinError(ps, fmt.Sprintf("Unable to get row in '%s': %s", file.GetPath(), err), "Load\\xlsx")
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
	// Returns:
	// * the saved file URI.
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
