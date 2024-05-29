//

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

	"spreadsheet": {
		Argsn: 2,
		Doc:   "Create a spreadsheet by accepting block of column names and flat block of values",
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

	"to-spreadsheet": {
		Argsn: 1,
		Doc:   "Create a spreadsheet by accepting block or list of dicts",
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

	"get-rows": {
		Argsn: 1,
		Doc:   "Create a spreadsheet by accepting block of column names and flat block of values",
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

	"add-rows": {
		Argsn: 2,
		Doc:   "Create a spreadsheet by accepting block of column names and flat block of values",
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

	// TODO 2 -- this could move to a go functio so it could be called by general load that uses extension to define the loader
	"load\\csv": {
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

	"sort-by!": {
		Argsn: 3,
		Doc:   "Sorts row by given column.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			dir, ok := arg2.(env.Word)
			if !ok {
				return MakeArgError(ps, 3, []env.Type{env.WordType}, "sort-col!")
			}
			var dirAsc bool
			if dir.Index == ps.Idx.IndexWord("asc") {
				dirAsc = true
			} else if dir.Index == ps.Idx.IndexWord("desc") {
				dirAsc = false
			} else {
				return MakeBuiltinError(ps, "Direction can be just asc or desc.", "sort-col!")
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
					return MakeArgError(ps, 2, []env.Type{env.WordType}, "sort-col!")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.SpreadsheetType}, "sort-col!")
			}
		},
	},
	"columns": {
		Argsn: 2,
		Doc:   "Returs spreasheet with just given columns.",
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
	"columns?": {
		Argsn: 1,
		Doc:   "Gets the column names as block.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch spr := arg0.(type) {
			case env.Spreadsheet:
				return spr.GetColumns()
			default:
				return MakeArgError(ps, 1, []env.Type{env.SpreadsheetType}, "columns?")
			}
		},
	},
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
	"add-col!": {
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
							return MakeArgError(ps, 4, []env.Type{env.BlockType}, "add-col!")
						}
					case env.Word:
						switch replaceBlock := arg3.(type) {
						case env.Block:
							if replaceBlock.Series.Len() != 2 {
								return MakeBuiltinError(ps, "Replacement block must contain a regex object and replacement string.", "add-col!")
							}
							regexNative, ok := replaceBlock.Series.S[0].(env.Native)
							if !ok {
								return MakeBuiltinError(ps, "First element of replacement block must be a regex object.", "add-col!")
							}
							regex, ok := regexNative.Value.(*regexp.Regexp)
							if !ok {
								return MakeBuiltinError(ps, "First element of replacement block must be a regex object.", "add-col!")
							}
							replaceStr, ok := replaceBlock.Series.S[1].(env.String)
							if !ok {
								return MakeBuiltinError(ps, "Second element of replacement block must be a string.", "add-col!")
							}
							err := GenerateColumnRegexReplace(ps, &spr, newCol, fromCols, regex, replaceStr.Value)
							if err != nil {
								return err
							}
							return spr
						default:
							return MakeArgError(ps, 3, []env.Type{env.BlockType}, "add-col!")
						}
					default:
						return MakeArgError(ps, 3, []env.Type{env.BlockType}, "add-col!")
					}
				default:
					return MakeArgError(ps, 2, []env.Type{env.WordType}, "add-col!")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.SpreadsheetType}, "add-col!")
			}
		},
	},
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
					return MakeArgError(ps, 2, []env.Type{env.BlockType}, "add-index!")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.SpreadsheetType}, "add-index!")
			}
		},
	},
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
				if er != nil {
					return nil
				}
				if firstVal == nil {
					firstVal = val.(env.Object)
				}
				ctx.Set(w.Index, val.(env.Object))
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
				intVal, _ := strconv.Atoi(row.Values[colNum].(env.String).Value)
				newS.Rows[i].Values[colNum] = *env.NewInteger(int64(intVal))
			case "dec":
				floatVal, _ := strconv.ParseFloat(row.Values[colNum].(env.String).Value, 64)
				newS.Rows[i].Values[colNum] = *env.NewDecimal(floatVal)
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
		newRow := make([]any, len(combinedCols))

		// the row id of the second spreadsheet that matches the current row
		s2RowId := -1
		// use index if available
		if ix, ok := s2.Indexes[col2]; ok {
			if rowIds, ok := ix[val1]; ok {
				// if there are multiple rows  with the same value (ie. joining on non-unique column), just use the first one
				s2RowId = rowIds[0]
			}
		} else {
			for j, row2 := range s2.GetRows() {
				val2, err := s2.GetRowValue(col2, row2)
				if err != nil {
					return MakeError(ps, fmt.Sprintf("Couldn't retrieve value at row %d (%s)", j, err))
				}
				val1o, ok := val1.(env.Object)
				if ok {
					val2o, ok := val2.(env.Object)
					if ok {
						if val1o.Equal(val2o) {
							s2RowId = j
							break
						}
					} else {
						if env.RyeToRaw(val1o) == val2 {
							s2RowId = j
							break
						}
					}
				}
				val1s, ok := val1.(string)
				if ok {
					if val1s == val2.(string) {
						s2RowId = j
						break
					}
				}
			}
		}
		if innerJoin && s2RowId == -1 {
			continue
		}
		copy(newRow, row1.Values)
		if s2RowId > -1 {
			for i, v := range s2.GetRow(ps, s2RowId).Values {
				newRow[i+len(s1.Cols)] = v
			}
		} else {
			for k := range s2.Cols {
				newRow[k+len(s1.Cols)] = env.Void{}
			}
		}
		nspr.AddRow(*env.NewSpreadsheetRow(newRow, nspr))
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
