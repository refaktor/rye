//

package evaldo

import (
	"encoding/csv"
	"os"
	"rye/env"
	"rye/util"
	"sort"
	"strconv"
)

var Builtins_spreadsheet = map[string]*env.Builtin{

	"new-spreadsheet": {
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
						rowd := make([]interface{}, header.Len())
						for ii := 0; ii < header.Len(); ii++ {
							k1 := data.Pop()
							rowd[ii] = k1
						}
						spr.AddRow(env.SpreadsheetRow{rowd, spr})
					}
					return *spr
				}
			}
			return MakeError(ps, "Some error")
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
			}
			return MakeError(ps, "Some error")
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
						rowd := make([]interface{}, len(spr.Cols))
						for ii := 0; ii < len(spr.Cols); ii++ {
							k1 := data.Pop()
							rowd[ii] = k1
						}
						spr.AddRow(env.SpreadsheetRow{rowd, &spr})
					}
					return spr
				case env.Native:
					spr.Rows = append(spr.Rows, data1.Value.([]env.SpreadsheetRow)...)
					return spr
				}
			}
			return MakeError(ps, "Some errora")
		},
	},

	// TODO 2 -- this could move to a go functio so it could be called by general load that uses extension to define the loader
	"load\\csv": {
		Argsn: 1,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch file := arg0.(type) {
			case env.Uri:
				// rows, err := db1.Value.(*sql.DB).Query(sqlstr, vals...)
				f, err := os.Open(file.GetPath())
				if err != nil {
					// log.Fatal("Unable to read input file "+filePath, err)
				}
				defer f.Close()

				csvReader := csv.NewReader(f)
				rows, err := csvReader.ReadAll()
				if err != nil {
					// log.Fatal("Unable to parse file as CSV for "+filePath, err)
				}
				spr := env.NewSpreadsheet(rows[0])
				spr.SetRaw(rows[1:])
				return *spr
			}
			return MakeError(ps, "Some error")
		},
	},
	"where-equal": {
		Argsn: 3,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch spr := arg0.(type) {
			case env.Spreadsheet:
				switch col := arg1.(type) {
				case env.Word:
					return WhereEquals(ps, spr, ps.Idx.GetWord(col.Index), arg2)
				}
			}
			return MakeError(ps, "Some error")
		},
	},
	"where-greater": {
		Argsn: 3,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch spr := arg0.(type) {
			case env.Spreadsheet:
				switch col := arg1.(type) {
				case env.Word:
					return WhereGreater(ps, spr, ps.Idx.GetWord(col.Index), arg2)
				}
			}
			return MakeError(ps, "Some error")
		},
	},
	"limit": {
		Argsn: 2,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch spr := arg0.(type) {
			case env.Spreadsheet:
				switch n := arg1.(type) {
				case env.Integer:
					return Limit(ps, spr, int(n.Value))
				}
			}
			return MakeError(ps, "Some error")
		},
	},

	"sort-col!": {
		Argsn: 2,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch spr := arg0.(type) {
			case env.Spreadsheet:
				switch col := arg1.(type) {
				case env.Word:
					SortByColumn(ps, &spr, ps.Idx.GetWord(col.Index))
					return spr
				}
			}
			return MakeError(ps, "Some error")
		},
	},
	"sort-col\\desc!": {
		Argsn: 2,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch spr := arg0.(type) {
			case env.Spreadsheet:
				switch col := arg1.(type) {
				case env.Word:
					SortByColumnDesc(ps, &spr, ps.Idx.GetWord(col.Index))
					return spr
				}
			}
			return MakeError(ps, "Some error")
		},
	},
	"columns": {
		Argsn: 2,
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
				}
			}
			return MakeError(ps, "Some error")
		},
	},
	"columns?": {
		Argsn: 1,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch spr := arg0.(type) {
			case env.Spreadsheet:
				return spr.GetColumns()
			}
			return MakeError(ps, "Some error")
		},
	},
	"gen-col": {
		Argsn: 4,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch spr := arg0.(type) {
			case env.Spreadsheet:
				switch name := arg1.(type) {
				case env.Word:
					switch extract := arg2.(type) {
					case env.Block:
						switch code := arg3.(type) {
						case env.Block:
							return GenerateColumn(ps, spr, name, extract, code)
						}
					}
				}
			}
			return MakeError(ps, "Some error")
		},
	},
	"add-index": {
		Argsn: 2,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch spr := arg0.(type) {
			case env.Spreadsheet:
				switch col := arg1.(type) {
				case env.Block:
					for c := range col.Series.S {
						switch ww := col.Series.S[c].(type) {
						case env.Word:
							res := AddIndex(ps, spr, ww)
							switch res2 := res.(type) {
							case env.Spreadsheet:
								spr = res2
							default:
								return res
							}
						default:
							return MakeError(ps, "Block of tagwords needed")
						}
					}
					return spr
				}
			}
			return MakeError(ps, "Some error TODO")
		},
	},
}

func GenerateColumn(ps *env.ProgramState, s env.Spreadsheet, name env.Word, extractCols env.Block, code env.Block) env.Object {
	// add name to columns
	s.Cols = append(s.Cols, ps.Idx.GetWord(name.Index))
	if s.RawMode {
		// for each row in spreadsheet
		for ir, row := range s.RawRows {
			// create a empty context connected to current context
			ctx := env.NewEnv(ps.Ctx)
			// for each word in extractCols get a value from current row and set word in context to it
			var firstVal env.Object
			for _, word := range extractCols.Series.S {
				switch w := word.(type) {
				case env.Word:
					val, er := s.GetRawRowValue(ps.Idx.GetWord(w.Index), row)
					if er != nil {
						return nil
					}
					if firstVal == nil {
						firstVal = env.String{val}
					}
					ctx.Set(w.Index, env.String{val})
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
			row = append(row, ps.Res.(env.String).Value)
			s.RawRows[ir] = row
		}
		return s
	} else {
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
}

func AddIndex(ps *env.ProgramState, s env.Spreadsheet, column env.Word) env.Object {
	// make a new index map
	s.Index = make(map[string][]int, 0)
	// for each row in spreadsheet
	colstr := ps.Idx.GetWord(column.Index)
	s.IndexName = colstr
	for ir, row := range s.RawRows {
		val, er := s.GetRawRowValue(colstr, row)
		// fmt.Println(val)
		if er != nil {
			return MakeError(ps, "Couldn't retrieve index at row "+strconv.Itoa(ir))
		}
		if subidx, ok := s.Index[val]; ok {
			s.Index[val] = append(subidx, ir)
		} else {
			subidx := make([]int, 1)
			subidx[0] = ir
			s.Index[val] = subidx
		}
	}
	//fmt.Println(s.IndexName)
	//fmt.Println(len(s.Index))
	return s
}

func SortByColumn(ps *env.ProgramState, s *env.Spreadsheet, name string) {

	idx := env.IndexOfString(name, s.Cols)

	compareCol := func(i, j int) bool {
		return greaterThanNew(s.Rows[j].Values[idx].(env.Object), s.Rows[i].Values[idx].(env.Object))
	}

	sort.Slice(s.Rows, compareCol)

}

func SortByColumnDesc(ps *env.ProgramState, s *env.Spreadsheet, name string) {

	idx := env.IndexOfString(name, s.Cols)

	compareCol := func(i, j int) bool {
		return greaterThanNew(s.Rows[i].Values[idx].(env.Object), s.Rows[j].Values[idx].(env.Object))
	}

	sort.Slice(s.Rows, compareCol)

}

func WhereEquals(ps *env.ProgramState, s env.Spreadsheet, name string, val interface{}) env.Object {
	idx := env.IndexOfString(name, s.Cols)
	nspr := env.NewSpreadsheet(s.Cols)
	if idx > -1 {
		if s.RawMode {
			var res [][]string
			if name == s.IndexName {
				//				fmt.Println("Using index")
				switch ov := val.(type) {
				case env.String:
					idxs := s.Index[ov.Value]
					res = make([][]string, len(idxs))
					for i, idx := range idxs {
						res[i] = s.RawRows[idx]
					}
				}
			} else {
				//			fmt.Println("Not using index")
				res = make([][]string, 0)
				for _, row := range s.RawRows {
					if len(row) > idx {
						switch ov := val.(type) {
						case env.String:
							// fmt.Println(ov.Value)
							// fmt.Println(row[idx])
							// fmt.Println(idx)
							if ov.Value == row[idx] {
								// fmt.Println("appending")
								res = append(res, row)
								// fmt.Println(res)
							}
						}
					}
				}
			}
			// fmt.Println(res)
			nspr.SetRaw(res)
			return *nspr
		} else {
			for _, row := range s.Rows {
				if len(s.Cols) > idx {
					switch val2 := val.(type) {
					case env.Object:
						if util.EqualValues(ps, val2, row.Values[idx].(env.Object)) {
							nspr.AddRow(row)
						}
					}
				}
			}
			return *nspr
		}
	} else {
		return makeError(ps, "Column not found")
	}
}

func WhereGreater(ps *env.ProgramState, s env.Spreadsheet, name string, val interface{}) env.Object {
	idx := env.IndexOfString(name, s.Cols)
	nspr := env.NewSpreadsheet(s.Cols)
	if idx > -1 {
		if s.RawMode {
			var res [][]string
			if name == s.IndexName {
				switch ov := val.(type) {
				case env.String:
					idxs := s.Index[ov.Value]
					res = make([][]string, len(idxs))
					for i, idx := range idxs {
						res[i] = s.RawRows[idx]
					}
				}
			} else {
				res = make([][]string, 0)
				for _, row := range s.RawRows {
					if len(row) > idx {
						switch ov := val.(type) {
						case env.String:
							if ov.Value == row[idx] {
								res = append(res, row)
							}
						}
					}
				}
			}
			nspr.SetRaw(res)
			return *nspr
		} else {
			for _, row := range s.Rows {
				if len(s.Cols) > idx {
					switch val2 := val.(type) {
					case env.Object:
						if greaterThanNew(val2, row.Values[idx].(env.Object)) {
							nspr.AddRow(row)
						}
					}
				}
			}
			return *nspr
		}
	} else {
		return makeError(ps, "Column not found")
	}
}

func Limit(ps *env.ProgramState, s env.Spreadsheet, n int) env.Object {
	nspr := env.NewSpreadsheet(s.Cols)
	if s.RawMode {
		nspr.SetRaw(s.RawRows[0:n])
		return *nspr
	} else {
		nspr.Rows = s.Rows[0:n]
		return *nspr
	}
}
