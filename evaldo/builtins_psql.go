//go:build b_psql
// +build b_psql

package evaldo

// import "C"

import (
	"database/sql"
	"rye/env"

	_ "github.com/lib/pq"
)

var Builtins_psql = map[string]*env.Builtin{

	"postgres-schema//open": {
		Argsn: 1,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch str := arg0.(type) {
			case env.Uri:
				db, err := sql.Open("postgres", str.Path) // TODO -- we need to make path parser in URI then this will be path
				if err != nil {
					// TODO --
					//fmt.Println("Error1")
					env1.FailureFlag = true
					return env.NewError("Error opening SQL: " + err.Error())
				} else {
					//fmt.Println("Error2")
					return *env.NewNative(env1.Idx, db, "Rye-psql")
				}
			default:
				env1.FailureFlag = true
				return env.NewError("arg 1 should be Uri")
			}

		},
	},

	"Rye-psql//exec": {
		Argsn: 2,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			var sqlstr string
			var vals []interface{}
			switch db1 := arg0.(type) {
			case env.Native:
				switch str := arg1.(type) {
				case env.Block:
					ser := env1.Ser
					env1.Ser = str.Series
					values := make([]interface{}, 0, 2)
					_, vals = SQL_EvalBlock(env1, MODE_PSQL, values)
					sqlstr = env1.Res.(env.String).Value
					env1.Ser = ser
				case env.String:
					sqlstr = str.Value
				default:
					env1.ErrorFlag = true
					return env.NewError("First argument should be block or string.")
				}
				if sqlstr != "" {
					//fmt.Println(sqlstr)
					//fmt.Println(vals)
					db2 := db1.Value.(*sql.DB)
					res, err := db2.Exec(sqlstr, vals...)
					if err != nil {
						env1.FailureFlag = true
						return env.NewError("Error" + err.Error())
					} else {
						num, _ := res.RowsAffected()
						if num > 0 {
							return env.Integer{1}
						} else {
							env1.FailureFlag = true
							return env.NewError("no rows affected")
						}

					}
				}
			default:
				return env.NewError("arg 1111 should be string %s")
			}
			return env.NewError("arg 0000 should be string %s")
		},
	},

	"Rye-psql//query": {
		Argsn: 2,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			var sqlstr string
			var vals []interface{}
			switch db1 := arg0.(type) {
			case env.Native:
				switch str := arg1.(type) {
				case env.Block:
					//fmt.Println("BLOCK ****** *****")
					ser := env1.Ser
					env1.Ser = str.Series
					values := make([]interface{}, 0, 2)
					_, vals = SQL_EvalBlock(env1, MODE_PSQL, values)
					sqlstr = env1.Res.(env.String).Value
					env1.Ser = ser
				case env.String:
					sqlstr = str.Value
				default:
					env1.ErrorFlag = true
					return env.NewError("First argument should be block or string.")
				}
				if sqlstr != "" {
					//					fmt.Println(sqlstr)
					//					fmt.Println(vals)
					rows, err := db1.Value.(*sql.DB).Query(sqlstr, vals...)
					result := make([]map[string]interface{}, 0)
					if err != nil {
						env1.FailureFlag = true
						return env.NewError("Error" + err.Error())
					} else {
						cols, _ := rows.Columns()
						spr := env.NewSpreadsheet(cols)
						i := 0
						for rows.Next() {

							var sr env.SpreadsheetRow

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
								sr.Values = append(sr.Values, *val)
							}
							spr.AddRow(sr)
							result = append(result, m)
							// Outputs: map[columnName:value columnName2:value2 columnName3:value3 ...]
							i++
						}
						rows.Close() //good habit to close
						//fmt.Println("+++++")
						//	fmt.Print(result)
						if i == 0 {
							env1.FailureFlag = true
							return env.NewError("no data")
						}
						return *spr
						//return *env.NewNative(env1.Idx, *spr, "Rye-spreadsheet")
					}
				} else {
					return env.NewError("Empty SQL")
				}

			default:
				return env.NewError("First argument should be native.")
			}
		},
	},
}
