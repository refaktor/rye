//go:build !no_mysql
// +build !no_mysql

package evaldo

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/refaktor/rye/env"

	_ "github.com/go-sql-driver/mysql"
)

var Builtins_mysql = map[string]*env.Builtin{

	//
	// ##### MySQL ##### "MySQL database functions"
	//
	// Example:
	// Open mysql://user:pass@tcp(localhost)/dbname
	// Args:
	// * uri: MySQL connection string URI
	// Returns:
	// * native MySQL database connection
	"mysql-uri//Open": {
		Argsn: 1,
		Doc:   "Opens a connection to a MySQL database.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch str := arg0.(type) {
			case env.Uri:
				fmt.Println(str.Path)
				fmt.Println(str.GetPath())
				fmt.Println(str.GetFullUri(*ps.Idx))
				db, err := sql.Open("mysql", str.Path) // TODO -- we need to make path parser in URI then this will be path
				if err != nil {
					// TODO --
					//fmt.Println("Error1")
					ps.FailureFlag = true
					errMsg := fmt.Sprintf("Error opening SQL: %v", err.Error())
					return MakeBuiltinError(ps, errMsg, "mysql-uri//Open")
				} else {
					//fmt.Println("Error2")
					return *env.NewNative(ps.Idx, db, "Rye-mysql")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.UriType}, "mysql-uri//Open")
			}

		},
	},

	// Example:
	// equal { Open\pwd mysql://user@tcp(localhost:3306)/dbname" "password"
	// Args:
	// * uri: MySQL connection string URI without password
	// * password: Password for the database connection
	// Returns:
	// * native MySQL database connection
	"mysql-uri//Open\\pwd": {
		Argsn: 2,
		Doc:   "Opens a connection to a MySQL database with separate password parameter.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch str := arg0.(type) {
			case env.Uri:
				switch pwd := arg1.(type) {
				case env.String:
					path := strings.Replace(str.Path, "@", ":"+pwd.Value+"@", 1)
					fmt.Println(path)
					db, err := sql.Open("mysql", path)
					if err != nil {
						// TODO --
						//fmt.Println("Error1")
						ps.FailureFlag = true
						errMsg := fmt.Sprintf("Error opening SQL: %v", err.Error())
						return MakeBuiltinError(ps, errMsg, "mysql-uri//Open\\pwd")
					} else {
						//fmt.Println("Error2")
						return *env.NewNative(ps.Idx, db, "Rye-mysql")
					}
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "mysql-uri//Open\\pwd")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.UriType}, "mysql-uri//Open\\pwd")
			}
		},
	},

	// Example:
	// Open mysql://user:pass@tcp(localhost:3306)/dbname" ,
	// |Exec { INSERT INTO test VALUES ( 1 , "test" ) }
	// Args:
	// * db: MySQL database connection
	// * sql: SQL statement as string or block
	// Returns:
	// * integer 1 if rows were affected, error otherwise
	"Rye-mysql//Exec": {
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
					_, vals = SQL_EvalBlock(ps, MODE_SQLITE, values)
					sqlstr = ps.Res.(env.String).Value
					ps.Ser = ser
				case env.String:
					sqlstr = str.Value
				default:
					ps.ErrorFlag = true
					return MakeArgError(ps, 2, []env.Type{env.BlockType, env.StringType}, "Rye-mysql//Exec")
				}
				if sqlstr != "" {
					//fmt.Println(sqlstr)
					//fmt.Println(vals)
					db2 := db1.Value.(*sql.DB)
					res, err := db2.Exec(sqlstr, vals...)
					if err != nil {
						ps.FailureFlag = true
						return MakeBuiltinError(ps, err.Error(), "Rye-mysql//Exec")
					} else {
						num, _ := res.RowsAffected()
						if num > 0 {
							return env.NewInteger(1)
						} else {
							ps.FailureFlag = true
							return MakeBuiltinError(ps, "No rows affected.", "Rye-mysql//Exec")
						}
					}
				} else {
					return MakeBuiltinError(ps, "SQL string is blank.", "Rye-mysql//Exec")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Rye-mysql//Exec")
			}
		},
	},

	// Example:
	// id: 101
	// db: Open mysql://user:pass@tcp(localhost:3306)/dbname
	// db .Query { SELECT * FROM test where id = ?id }
	// Args:
	// * db: MySQL database connection
	// * sql: SQL query as string or block
	// Returns:
	// * table containing query results
	"Rye-mysql//Query": {
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
					_, vals = SQL_EvalBlock(ps, MODE_SQLITE, values)
					sqlstr = ps.Res.(env.String).Value
					ps.Ser = ser
				case env.String:
					sqlstr = str.Value
				default:
					ps.ErrorFlag = true
					return MakeArgError(ps, 2, []env.Type{env.BlockType, env.StringType}, "Rye-mysql//Query")
				}
				if sqlstr != "" {
					//					fmt.Println(sqlstr)
					//					fmt.Println(vals)
					rows, err := db1.Value.(*sql.DB).Query(sqlstr, vals...)
					// result := make([]map[string]any, 0)
					if err != nil {
						ps.FailureFlag = true
						return MakeBuiltinError(ps, err.Error(), "Rye-mysql//Query")
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
							if err := rows.Scan(columnPointers...); err != nil {
								ps.FailureFlag = true
								return MakeBuiltinError(ps, err.Error(), "Rye-mysql//Query")
							}

							// Create our map, and retrieve the value for each column from the pointers slice,
							// storing it in the map with the name of the column as the key.
							m := make(map[string]any)
							for i, colName := range cols {
								val := *columnPointers[i].(*any)
								switch vval := val.(type) {
								case []uint8:
									//								fmt.Println(val)
									//								fmt.Printf("%T", vval)
									m[colName] = env.ToRyeValue(string(vval))
									sr.Values = append(sr.Values, env.ToRyeValue(string(vval)))
								default:
									m[colName] = env.ToRyeValue(vval)
									sr.Values = append(sr.Values, env.ToRyeValue(vval))
								}
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
							return MakeBuiltinError(ps, "No data.", "Rye-mysql//Query")
						}
						return *spr
						//return *env.NewNative(ps.Idx, *spr, "Rye-table")
					}
				} else {
					return MakeBuiltinError(ps, "Empty SQL.", "Rye-mysql//Query")
				}

			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "Rye-mysql//Query")
			}
		},
	},
}
