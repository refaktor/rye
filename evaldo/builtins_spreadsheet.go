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
				switch data1 := arg1.(type) {
				case env.Block:
					header := header1.Series
					cols := make([]string, header.Len())
					for header.Pos() < header.Len() {
						i := header.Pos()
						k1 := header.Pop()
						switch k := k1.(type) {
						case env.String:
							cols[i] = k.Value
						}
					}
					spr := env.NewSpreadsheet(cols)
					data := data1.Series
					for data.Pos() < data.Len() {
						rowd := make([]any, header.Len())
						for ii := 0; ii < header.Len(); ii++ {
							k1 := data.Pop()
							rowd[ii] = k1
						}
						spr.AddRow(*env.NewSpreadsheetRow(rowd, spr))
					}
					return *spr
				default:
					return MakeArgError(ps, 2, []env.Type{env.BlockType}, "spreadsheet")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType}, "spreadsheet")
			}
		},
	},

	"to-spreadsheet": {
		Argsn: 1,
		Doc:   "Create a spreadsheet by accepting block of dicts",
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

			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType}, "to-spreadsheet")
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
							return MakeBuiltinError(ps, "Unable to write line: "+strconv.Itoa(int(ir)), "save\\csv")
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

	"sort-col!": {
		Argsn: 2,
		Doc:   "Sorts row by given column.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch spr := arg0.(type) {
			case env.Spreadsheet:
				switch col := arg1.(type) {
				case env.Word:
					SortByColumn(ps, &spr, ps.Idx.GetWord(col.Index))
					return spr
				default:
					return MakeArgError(ps, 2, []env.Type{env.WordType}, "sort-col!")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.SpreadsheetType}, "sort-col!")
			}
		},
	},
	"sort-col\\desc!": {
		Argsn: 2,
		Doc:   "Sorts rows by given column, descending.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch spr := arg0.(type) {
			case env.Spreadsheet:
				switch col := arg1.(type) {
				case env.Word:
					SortByColumnDesc(ps, &spr, ps.Idx.GetWord(col.Index))
					return spr
				default:
					return MakeArgError(ps, 2, []env.Type{env.WordType}, "sort-col\\desc!")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.SpreadsheetType}, "sort-col\\desc!")
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
							return GenerateColumnRegexReplace(ps, spr, newCol, fromCols, regex, replaceStr.Value)
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
	"add-index!": {
		Argsn: 2,
		Doc:   "Indexes all values in a colun and istre it,",
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
					res := AddIndexes(ps, &spr, colWords)
					if res != nil {
						return res
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

func GenerateColumnRegexReplace(ps *env.ProgramState, s env.Spreadsheet, name env.Word, fromColName env.Word, re *regexp.Regexp, pattern string) env.Object {
	// add name to columns
	s.Cols = append(s.Cols, ps.Idx.GetWord(name.Index))
	for ix, row := range s.Rows {
		// get value from current row
		val, err := s.GetRowValue(ps.Idx.GetWord(fromColName.Index), row)
		if err != nil {
			return MakeError(ps, "Couldn't retrieve value at row "+strconv.Itoa(ix))
		}

		var newVal any
		valStr, ok := val.(env.String)
		if !ok {
			newVal = ""
		} else {
			// replace the value with the regex
			newVal = env.NewString(re.ReplaceAllString(valStr.Value, pattern))
		}
		// set the result of code block as the new column value in this row
		row.Values = append(row.Values, newVal)
		s.Rows[ix] = row
	}
	return s
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
