//

package evaldo

import (
	"encoding/csv"
	"os"
	"rye/env"
	"strconv"
)

var Builtins_spreadsheet = map[string]*env.Builtin{

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
				case env.Tagword:
					return spr.WhereEquals(ps, ps.Idx.GetWord(col.Index), arg2)
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
				case env.Tagword:
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
						case env.Tagword:
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

func GenerateColumn(ps *env.ProgramState, s env.Spreadsheet, name env.Tagword, extractCols env.Block, code env.Block) env.Object {
	// add name to columns
	s.Cols = append(s.Cols, ps.Idx.GetWord(name.Index))
	// for each row in spreadsheet
	for ir, row := range s.RawRows {
		// create a empty context connected to current context
		ctx := env.NewEnv(ps.Ctx)
		// for each word in extractCols get a value from current row and set word in context to it
		for _, word := range extractCols.Series.S {
			switch w := word.(type) {
			case env.Word:
				val, er := s.GetRawRowValue(ps.Idx.GetWord(w.Index), row)
				if er != nil {
					return nil
				}
				ctx.Set(w.Index, env.String{val})
			}
		}
		// execute the block of code injected with first value
		ser := ps.Ser
		ps.Ser = code.Series
		EvalBlockInCtx(ps, ctx)
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
}

func AddIndex(ps *env.ProgramState, s env.Spreadsheet, column env.Tagword) env.Object {
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
