//go:build !no_psql
// +build !no_psql

package evaldo

import (
	"database/sql"
	"fmt"

	"github.com/refaktor/rye/env"

	_ "github.com/lib/pq"
)

var Builtins_psql = map[string]*env.Builtin{

	//
	// ##### PostgreSQL ##### "PostgreSQL database functions"
	//
	// Tests:
	// equal { postgres-schema//Open %"postgres://user:pass@localhost:5432/dbname" |type? } 'native
	// Args:
	// * uri: PostgreSQL connection string URI
	// Returns:
	// * native PostgreSQL database connection
	"postgres-schema//Open": {
		Argsn: 1,
		Doc:   "Opens a connection to a PostgreSQL database.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch str := arg0.(type) {
			case env.Uri:
				db, err := sql.Open("postgres", str.Path) // TODO -- we need to make path parser in URI then this will be path
				if err != nil {
					// TODO --
					//fmt.Println("Error1")
					ps.FailureFlag = true
					errMsg := fmt.Sprintf("Error opening SQL: %v" + err.Error())
					return MakeBuiltinError(ps, errMsg, "postgres-schema//Open")
				} else {
					//fmt.Println("Error2")
					return *env.NewNative(ps.Idx, db, "Rye-psql")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.UriType}, "postgres-schema//Open")
			}

		},
	},

	// Tests:
	// equal { db: postgres-schema//Open %"postgres://user:pass@localhost:5432/dbname" , db |Rye-psql//Exec "INSERT INTO test VALUES (1, 'test')" |type? } 'integer
	// Args:
	// * db: PostgreSQL database connection
	// * sql: SQL statement as string or block
	// Returns:
	// * integer 1 if rows were affected, error otherwise
	"Rye-psql//Exec": {
		Argsn: 2,
		Doc:   "Executes a SQL statement that modifies data (INSERT, UPDATE, DELETE).",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			var sqlstr string
			var vals []any
			switch db1 := arg0.(type) {
			case env.Native:
				switch str := arg1.(type) {
				case env.Block:
					ser := ps.Ser
					ps.Ser = str.Series
					values := make([]any, 0, 2)
					_, vals = SQL_EvalBlock(ps, MODE_PSQL, values)
					sqlstr = ps.Res.(env.String).Value
					ps.Ser = ser
				case env.String:
					sqlstr = str.Value
				default:
					ps.ErrorFlag = true
					return MakeArgError(ps, 2, []env.Type{env.BlockType, env.StringType}, "Rye-psql//Exec")
				}
				if sqlstr != "" {
					//fmt.Println(sqlstr)
					//fmt.Println(vals)
					db2 := db1.Value.(*sql.DB)
					res, err := db2.Exec(sqlstr, vals...)
					if err != nil {
						ps.FailureFlag = true
						return env.NewError("Error" + err.Error())
					} else {
						num, _ := res.RowsAffected()
						if num > 0 {
							return env.NewInteger(1)
						} else {
							ps.FailureFlag = true
							return MakeBuiltinError(ps, "No rows affected.", "Rye-psql//Exec")
						}

					}
				} else {
					return MakeBuiltinError(ps, "Sql string is empty.", "Rye-psql//Exec")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Rye-psql//Exec")
			}
		},
	},

	// Tests:
	// equal { db: postgres-schema//Open %"postgres://user:pass@localhost:5432/dbname" , db |Rye-psql//Query "SELECT * FROM test" |type? } 'table
	// Args:
	// * db: PostgreSQL database connection
	// * sql: SQL query as string or block
	// Returns:
	// * table containing query results
	"Rye-psql//Query": {
		Argsn: 2,
		Doc:   "Executes a SQL query and returns results as a table.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			var sqlstr string
			var vals []any
			switch db1 := arg0.(type) {
			case env.Native:
				switch str := arg1.(type) {
				case env.Block:
					//fmt.Println("BLOCK ****** *****")
					ser := ps.Ser
					ps.Ser = str.Series
					values := make([]any, 0, 2)
					_, vals = SQL_EvalBlock(ps, MODE_PSQL, values)
					sqlstr = ps.Res.(env.String).Value
					ps.Ser = ser
				case env.String:
					sqlstr = str.Value
				default:
					ps.ErrorFlag = true
					return MakeArgError(ps, 2, []env.Type{env.BlockType, env.StringType}, "Rye-psql//Query")
				}
				if sqlstr != "" {
					//					fmt.Println(sqlstr)
					//					fmt.Println(vals)
					rows, err := db1.Value.(*sql.DB).Query(sqlstr, vals...)
					// result := make([]map[string]any, 0)
					if err != nil {
						ps.FailureFlag = true
						return MakeBuiltinError(ps, err.Error(), "Rye-psql//Query")
					} else {
						cols, _ := rows.Columns()
						spr := env.NewTable(cols)
						i := 0
						for rows.Next() {

							var sr env.TableRow

							columns := make([]any, len(cols))
							columnPointers := make([]any, len(cols))
							for i := range columns {
								columnPointers[i] = &columns[i]
							}

							// Scan the result into the column pointers...
							// if err := rows.Scan(columnPointers...); err != nil {
							//return err
							// }

							// Create our map, and retrieve the value for each column from the pointers slice,
							// storing it in the map with the name of the column as the key.
							m := make(map[string]any)
							for i, colName := range cols {
								val := columnPointers[i].(*any)
								m[colName] = *val
								sr.Values = append(sr.Values, *val)
							}
							spr.AddRow(sr)
							// result = append(result, m)
							// Outputs: map[columnName:value columnName2:value2 columnName3:value3 ...]
							i++
						}
						rows.Close() //good habit to close
						//fmt.Println("+++++")
						//	fmt.Print(result)
						if i == 0 {
							ps.FailureFlag = true
							return MakeBuiltinError(ps, "No data.", "Rye-psql//Query")
						}
						return *spr
						//return *env.NewNative(ps.Idx, *spr, "Rye-table")
					}
				} else {
					return MakeBuiltinError(ps, "Empty SQL.", "Rye-psql//Query")
				}

			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Rye-psql//Query")
			}
		},
	},
}
