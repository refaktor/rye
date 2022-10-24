//

package evaldo

import (
	"encoding/csv"
	"os"
	"rye/env"
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
}
