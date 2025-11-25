// SECTION: Core/tables

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
	"github.com/xuri/excelize/v2"
)

var Builtins_table = map[string]*env.Builtin{

	//
	// ##### Constructors #####  "Functions that construct a table."
	//
	// Tests:
	//  equal { table { "a" } { 1 2 } |type? } 'table
	//  equal { table { 'a  } { 1 2 } |type? } 'table
	// Args:
	//  * columns
	//  * data
	"table": {
		Argsn: 2,
		Doc:   "Creates a table by accepting block of column names and flat block of values",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch header1 := arg0.(type) {
			case env.Block:
				hlen := header1.Series.Len()
				cols, err := ColNames(ps, header1, "table")
				if err != nil {
					return err
				}
				spr := env.NewTable(cols)
				switch data1 := arg1.(type) {
				case env.Block:
					rdata := data1.Series.S

					if hlen > 0 {
						for i := 0; i < len(rdata)/hlen; i++ {
							rowd := make([]any, hlen)
							for ii := 0; ii < hlen; ii++ {
								rowd[ii] = rdata[i*hlen+ii]
							}
							spr.AddRow(*env.NewTableRow(rowd, spr))
						}
					}
					return *spr
				case env.List:
					rdata := data1.Data
					for i := 0; i < len(rdata)/hlen; i++ {
						rowd := make([]any, hlen)
						for ii := 0; ii < hlen; ii++ {
							rowd[ii] = rdata[i*hlen+ii]
						}
						spr.AddRow(*env.NewTableRow(rowd, spr))
					}
					return *spr
				default:
					return MakeArgError(ps, 2, []env.Type{env.BlockType}, "table")
				}
				/* for data.Pos() < data.Len() {
					rowd := make([]any, header.Len())
					for ii := 0; ii < header.Len(); ii++ {
						k1 := data.Pop()
						rowd[ii] = k1
					}
					spr.AddRow(*env.NewTableRow(rowd, spr))
				} */
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType}, "table")
			}
		},
	},
	// Tests:
	//  equal { table\columns { "a" } { { 1 2 3 } } |type? } 'table
	//  equal { table\columns { "a" "b" } { { 1 2 3 } { 4 5 6 } } |length? } 3
	// Example:
	//  table\columns { 'a 'b } { { 1 2 } { "x" "y" } }
	// Args:
	//  * columns - names of the columns
	//  * data - block or list of columns (each column is a block or list)
	"table\\columns": {
		Argsn: 2,
		Doc:   "Creats a table by accepting a block of columns",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) (res env.Object) {
			return SheetFromColumns(ps, arg0, arg1)
		},
	},

	// Tests:
	//  equal { table\rows { 'a 'b } { { 1 2 } { 3 4 } } } table { 'a 'b } { 1 2 3 4 }
	//  equal { table\rows { 'a 'b } list [ list [ 1 2 ] list [ 3 4 ] ] |type? } 'table
	// Args:
	//  * columns - names of the columns
	//  * data - block or list of rows (each row is a block or list)
	"table\\rows": {
		Argsn: 2,
		Doc:   "Creates a table by accepting a block or list of rows",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) (res env.Object) {
			cols, err := ColNames(ps, arg0, "table\\rows	")
			if err != nil {
				return err
			}
			spr := env.NewTable(cols)
			switch rows := arg1.(type) {
			case env.Block:
				for _, objRow := range rows.Series.S {
					row, err := TableRowsFromBlockOrList(ps, spr, len(cols), objRow)
					if err != nil {
						return err
					}
					spr.AddRow(*row)
				}
				return *spr
			case env.List:
				for _, listRow := range rows.Data {
					row, err := TableRowsFromBlockOrList(ps, spr, len(cols), listRow)
					if err != nil {
						return err
					}
					spr.AddRow(*row)
				}
				return *spr
			default:
				return MakeArgError(ps, 2, []env.Type{env.BlockType, env.ListType}, "table\\rows")
			}
		},
	},
	// WARN: dict seems not to retain order of keys, so columns aren't ordered as in dict

	// Tests:
	//  equal { to-table list [ dict { "a" 1 } dict { "a" 2 } ] |type? } 'table
	//  equal { to-table list [ dict { "a" 1 "b" "Jim" } dict { "a" 2 "b" "Bob" } ] |header? |sort } list { "a" "b" }
	//  equal { to-table list [ dict { "a" 1 "b" "Jim" } dict { "a" 2 "b" "Bob" } ] |column? "b" |first } "Jim"
	// Args:
	//  * data
	"to-table": {
		Argsn: 1,
		Doc:   "Creates a table by accepting block or list of dicts",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) (res env.Object) {
			switch block := arg0.(type) {
			case env.Block:
				data := block.Series
				if data.Len() == 0 {
					return MakeBuiltinError(ps, "Block is empty", "to-table")
				}
				k := make(map[string]struct{})
				for _, obj := range data.S {
					switch dict := obj.(type) {
					case env.Dict:
						for key := range dict.Data {
							k[key] = struct{}{}
						}
					default:
						return MakeBuiltinError(ps, "Block must contain only dicts", "to-table")
					}
				}
				var keys []string
				for key := range k {
					keys = append(keys, key)
				}
				spr := env.NewTable(keys)
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
						spr.AddRow(*env.NewTableRow(row, spr))
					}
				}
				return *spr

			case env.List:
				data := block.Data
				if len(data) == 0 {
					return MakeBuiltinError(ps, "List is empty", "to-table")
				}
				k := make(map[string]struct{})
				for _, obj := range data {
					switch dict := obj.(type) {
					case map[string]any:
						for key := range dict {
							k[key] = struct{}{}
						}
					case env.Dict:
						for key := range dict.Data {
							k[key] = struct{}{}
						}
					default:
						return MakeBuiltinError(ps, "List must contain only dicts", "to-table")
					}
				}
				var keys []string
				for key := range k {
					keys = append(keys, key)
				}
				spr := env.NewTable(keys)
				for _, obj := range data {
					row := make([]any, len(keys))
					switch dict := obj.(type) {
					case map[string]any:
						for i, key := range keys {
							data, ok := dict[key]
							if !ok {
								data = env.Void{}
							}
							row[i] = env.ToRyeValue(data)
						}
						spr.AddRow(*env.NewTableRow(row, spr))
					case env.Dict:
						row := make([]any, len(keys))
						for i, key := range keys {
							data, ok := dict.Data[key]
							if !ok {
								data = env.Void{}
							}
							row[i] = env.ToRyeValue(data)
						}
						spr.AddRow(*env.NewTableRow(row, spr))
					}
				}
				return *spr

			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType, env.ListType}, "to-table")
			}
		},
	},
	//
	// ##### Filtering #####  "Functions that construct a table."
	//
	// Example: filtering for rows with the name "Enno"
	//  sheet: table { "name" } { "Enno" "Enya" "Enid" "Bob" "Bill" }
	//  sheet .where-equal 'name "Enno"
	// Tests:
	//  equal { table { 'a } { 1 2 3 2 } |where-equal "a" 2 |length? } 2
	// Args:
	// * sheet
	// * column
	// * value
	// Tags: #filter #tables
	"where-equal": {
		Argsn: 3,
		Doc:   "Returns table of rows where specific colum is equal to given value.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch spr := arg0.(type) {
			case *env.Table:
				switch col := arg1.(type) {
				case env.Word:
					return WhereEquals(ps, *spr, ps.Idx.GetWord(col.Index), arg2)
				case env.String:
					return WhereEquals(ps, *spr, col.Value, arg2)
				default:
					return MakeArgError(ps, 2, []env.Type{env.WordType, env.StringType}, "where-equal")
				}
			case env.Table:
				switch col := arg1.(type) {
				case env.Word:
					return WhereEquals(ps, spr, ps.Idx.GetWord(col.Index), arg2)
				case env.String:
					return WhereEquals(ps, spr, col.Value, arg2)
				default:
					return MakeArgError(ps, 2, []env.Type{env.WordType, env.StringType}, "where-equal")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.TableType}, "where-equal")
			}
		},
	},
	// Tests:
	//  equal { table { 'a } { 1 2 3 2 } |where-not-equal "a" 2 |length? } 2
	// Args:
	// * sheet
	// * column
	// * value
	// Tags: #filter #tables
	"where-not-equal": {
		Argsn: 3,
		Doc:   "Returns table of rows where specific colum is not equal to given value.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch spr := arg0.(type) {
			case *env.Table:
				switch col := arg1.(type) {
				case env.Word:
					return WhereNotEquals(ps, *spr, ps.Idx.GetWord(col.Index), arg2)
				case env.String:
					return WhereNotEquals(ps, *spr, col.Value, arg2)
				default:
					return MakeArgError(ps, 2, []env.Type{env.WordType, env.StringType}, "where-not-equal")
				}
			case env.Table:
				switch col := arg1.(type) {
				case env.Word:
					return WhereNotEquals(ps, spr, ps.Idx.GetWord(col.Index), arg2)
				case env.String:
					return WhereNotEquals(ps, spr, col.Value, arg2)
				default:
					return MakeArgError(ps, 2, []env.Type{env.WordType, env.StringType}, "where-not-equal")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.TableType}, "where-not-equal")
			}
		},
	},

	// Notes:
	// Void is not a value that any function can accept or return so far in Rye. But there can be missing data
	// in tables ... for example from SQL query with left-join or plain null values in databases

	// Tests:
	//  equal { table { 'a } { 1 _ 3 _ } |where-void "a" |length? } 2
	// Args:
	// * sheet
	// * column
	// Tags: #filter #tables
	"where-void": {
		Argsn: 2,
		Doc:   "Returns table of rows where specific colum is equal to given value.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch spr := arg0.(type) {
			case *env.Table:
				switch col := arg1.(type) {
				case env.Word:
					return WhereEquals(ps, *spr, ps.Idx.GetWord(col.Index), env.NewVoid())
				case env.String:
					return WhereEquals(ps, *spr, col.Value, env.NewVoid())
				default:
					return MakeArgError(ps, 2, []env.Type{env.WordType, env.StringType}, "where-equal")
				}
			case env.Table:
				switch col := arg1.(type) {
				case env.Word:
					return WhereEquals(ps, spr, ps.Idx.GetWord(col.Index), env.NewVoid())
				case env.String:
					return WhereEquals(ps, spr, col.Value, env.NewVoid())
				default:
					return MakeArgError(ps, 2, []env.Type{env.WordType, env.StringType}, "where-equal")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.TableType}, "where-equal")
			}
		},
	},

	// Example: filting for names that start with "En"
	//  sheet: table { "name" } { "Enno" "Enya" "Enid" "Bob" "Bill" }
	//  sheet .where-match 'name "En.+"
	// Tests:
	//  equal { table { 'a } { "1" "2" "a3" "2b" } |where-match 'a regexp "^[0-9]$" |length? } 2
	// Args:
	// * sheet
	// * column
	// * regexp
	// Tags: #filter #tables
	"where-match": {
		Argsn: 3,
		Doc:   "Returns table of rows where a specific colum matches a regex.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) (res env.Object) {
			var spr *env.Table
			switch sheet := arg0.(type) {
			case env.Table:
				spr = &sheet
			case *env.Table:
				spr = sheet
			default:
				return MakeArgError(ps, 1, []env.Type{env.TableType}, "where-match")
			}
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
		},
	},

	// Example:
	//  sheet: table { "name" } { "Enno" "Enya" "Enid" "Bob" "Bill" "Benn" }
	//  sheet .where-contains 'name "nn"
	// Tests:
	//  equal { table { 'a } { "1" "2" "a3" "2b" } |where-contains 'a "2" |length? } 2
	// Args:
	// * sheet
	// * column
	// * substring
	// Tags: #filter #tables
	"where-contains": {
		Argsn: 3,
		Doc:   "Returns table of rows where specific colum contains a given string value.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) (res env.Object) {
			var spr *env.Table
			switch sheet := arg0.(type) {
			case env.Table:
				spr = &sheet
			case *env.Table:
				spr = sheet
			default:
				return MakeArgError(ps, 1, []env.Type{env.TableType}, "where-match")
			}
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
		},
	},

	// Example:
	//  sheet: table { "name" } { "Enno" "Enya" "Enid" "Bob" "Bill" "Benn" }
	//  sheet .where-contains 'name "nn"
	// Tests:
	//  equal { table { 'a } { "1" "2" "a3" "2b" } |where-not-contains 'a "3" |length? } 3
	// Args:
	// * sheet
	// * column
	// * substring
	// Tags: #filter #tables
	"where-not-contains": {
		Argsn: 3,
		Doc:   "Returns table of rows where specific colum contains a given string value.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) (res env.Object) {
			var spr *env.Table
			switch sheet := arg0.(type) {
			case env.Table:
				spr = &sheet
			case *env.Table:
				spr = sheet
			default:
				return MakeArgError(ps, 1, []env.Type{env.TableType}, "where-match")
			}
			switch s := arg2.(type) {
			case env.String:
				switch col := arg1.(type) {
				case env.Word:
					return WhereContains(ps, spr, ps.Idx.GetWord(col.Index), s.Value, true)
				case env.String:
					return WhereContains(ps, spr, col.Value, s.Value, true)
				default:
					return MakeArgError(ps, 2, []env.Type{env.WordType, env.StringType}, "where-not-contains")
				}
			default:
				return MakeArgError(ps, 3, []env.Type{env.StringType}, "where-not-contains")
			}
		},
	},
	// Example:
	//  sheet: table { "name" "age" } { "Enno" 30 "Enya" 25 "Enid" 40 "Bob" 19 "Bill" 45 "Benn" 29 }
	//  sheet .where-greater 'age 29
	// Tests:
	//  equal { table { 'a } { 1 2 3 2 } |where-greater 'a 1 |length? } 3
	// Args:
	// * sheet
	// * column
	// * value
	// Tags: #filter #table
	"where-greater": {
		Argsn: 3,
		Doc:   "Returns table of rows where specific colum is greater than given value.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			var spr *env.Table
			switch sheet := arg0.(type) {
			case env.Table:
				spr = &sheet
			case *env.Table:
				spr = sheet
			default:
				return MakeArgError(ps, 1, []env.Type{env.TableType}, "where-match")
			}
			switch col := arg1.(type) {
			case env.Word:
				return WhereGreater(ps, spr, ps.Idx.GetWord(col.Index), arg2)
			case env.String:
				return WhereGreater(ps, spr, col.Value, arg2)
			default:
				return MakeArgError(ps, 2, []env.Type{env.WordType, env.StringType}, "where-greater")
			}
		},
	},
	// Example: filting for names that contain "nn"
	//  sheet: table { "name" "age" } { "Enno" 30 "Enya" 25 "Enid" 40 "Bob" 19 "Bill" 45 "Benn" 29 }
	//  sheet .where-lesser 'age 29
	// Tests:
	//  equal { table { 'a } { 1 2 3 2 } |where-lesser 'a 3 |length? } 3
	// Args:
	// * sheet
	// * column
	// * value
	// Tags: #filter #table
	"where-lesser": {
		Argsn: 3,
		Doc:   "Returns table of rows where specific colum is lesser than given value.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			var spr *env.Table
			switch sheet := arg0.(type) {
			case env.Table:
				spr = &sheet
			case *env.Table:
				spr = sheet
			default:
				return MakeArgError(ps, 1, []env.Type{env.TableType}, "where-match")
			}
			switch col := arg1.(type) {
			case env.Word:
				return WhereLesser(ps, spr, ps.Idx.GetWord(col.Index), arg2)
			case env.String:
				return WhereLesser(ps, spr, col.Value, arg2)
			default:
				return MakeArgError(ps, 2, []env.Type{env.WordType, env.StringType}, "where-lesser")
			}
		},
	},
	// Returns a spreadhsheet of rows where the given column is between the given
	// values, non-inclusive.
	// Example: filtering for folks in their 20s
	//  sheet: table { "name" "age" } { "Enno" 30 "Enya" 25 "Enid" 40 "Bob" 19 "Bill" 45 "Benn" 29 }
	//  sheet .where-between 'age 19 30
	// Tests:
	//  equal { table { 'a } { 1 2 3 2 } |where-between 'a 1 3 |length? } 2
	// Args:
	// * sheet
	// * column
	// * lower-limit
	// * upper-limit
	// Tags: #filter #table
	"where-between": {
		Argsn: 4,
		Doc:   "Returns table of rows where specific colum is between given values.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) (res env.Object) {
			var spr *env.Table
			switch sheet := arg0.(type) {
			case env.Table:
				spr = &sheet
			case *env.Table:
				spr = sheet
			default:
				return MakeArgError(ps, 1, []env.Type{env.TableType}, "where-match")
			}
			switch col := arg1.(type) {
			case env.Word:
				return WhereBetween(ps, spr, ps.Idx.GetWord(col.Index), arg2, arg3, false)
			case env.String:
				return WhereBetween(ps, spr, col.Value, arg2, arg3, false)
			default:
				return MakeArgError(ps, 2, []env.Type{env.WordType, env.StringType}, "where-between")
			}
		},
	},

	// Returns a spreadhsheet of rows where the given column is between the given
	// values, non-inclusive.
	// Tests:
	//  equal { table { 'a } { 1 2 3 2 5 } |where-between\inclusive 'a 2 3 |length? } 3
	// Args:
	// * sheet
	// * column
	// * lower-limit
	// * upper-limit
	// Tags: #filter #table
	"where-between\\inclusive": {
		Argsn: 4,
		Doc:   "Returns table of rows where specific colum is between given values.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) (res env.Object) {
			var spr *env.Table
			switch sheet := arg0.(type) {
			case env.Table:
				spr = &sheet
			case *env.Table:
				spr = sheet
			default:
				return MakeArgError(ps, 1, []env.Type{env.TableType}, "where-match")
			}
			switch col := arg1.(type) {
			case env.Word:
				return WhereBetween(ps, spr, ps.Idx.GetWord(col.Index), arg2, arg3, true)
			case env.String:
				return WhereBetween(ps, spr, col.Value, arg2, arg3, true)
			default:
				return MakeArgError(ps, 2, []env.Type{env.WordType, env.StringType}, "where-between")
			}
		},
	},

	// Example: filtering for folks named "Enno" or "Enya"
	//  sheet: table { "name" "age" } { "Enno" 30 "Enya" 25 "Enid" 40 "Bob" 19 "Bill" 45 "Benn" 29 }
	//  sheet .where-in 'name { "Enno" "Enya" }
	// Tests:
	//  equal { table { "name" "age" } { "Enno" 30 "Enya" 25 "Bob" 19 }
	//          |where-in 'name { "Enno" "Enya" "Roger" } |column? "age"
	//  } { 30 25 }
	// Args:
	// * sheet
	// * column
	// * values-filtered-for
	// Tags: #filter #table
	"where-in": {
		Argsn: 3,
		Doc:   "Returns table of rows where specific colum value if found in block of values.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) (res env.Object) {
			switch spr := arg0.(type) {
			case env.Table:
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
				return MakeArgError(ps, 1, []env.Type{env.TableType}, "where-in")
			}
		},
	},

	// Example: filtering for folks named "Enno" or "Enya"
	//  sheet: table { "name" "age" } { "Enno" 30 "Enya" 25 "Enid" 40 "Bob" 19 "Bill" 45 "Benn" 29 }
	//  sheet .where-not-in 'name { "Enno" "Enya" }
	// Tests:
	//  equal { table { "name" "age" } { "Enno" 30 "Enya" 25 "Bob" 19 }
	//          |where-not-in 'name { "Enno" "Enya" "Roger" } |column? "age"
	//  } { 19 }
	// Args:
	// * sheet
	// * column
	// * values-filtered-for
	// Tags: #filter #table
	"where-not-in": {
		Argsn: 3,
		Doc:   "Returns table of rows where specific colum value if found in block of values.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) (res env.Object) {
			switch spr := arg0.(type) {
			case env.Table:
				switch s := arg2.(type) {
				case env.Block:
					switch col := arg1.(type) {
					case env.Word:
						return WhereNotIn(ps, spr, ps.Idx.GetWord(col.Index), s.Series.S)
					case env.String:
						return WhereNotIn(ps, spr, col.Value, s.Series.S)
					default:
						return MakeArgError(ps, 2, []env.Type{env.WordType, env.StringType}, "where-in")
					}
				default:
					return MakeArgError(ps, 3, []env.Type{env.StringType}, "where-in")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.TableType}, "where-in")
			}
		},
	},

	//
	// ##### Row level functions #####  "Functions that construct a table."
	//
	// Tests:
	//  equal {
	//	 table { "a" "b" } { 6 60 7 70 } |add-row { 8 80 } -> 2 -> "b"
	//  } 80
	// Args:
	// * sheet
	// * new-row
	"add-row": {
		Argsn: 2,
		Doc:   "Returns a table with new-row added to it",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch table := arg0.(type) {
			case env.Table:
				switch bloc := arg1.(type) {
				case env.Block:
					vals := make([]any, bloc.Series.Len())
					for i := 0; i < bloc.Series.Len(); i++ {
						vals[i] = bloc.Series.Get(i)
					}
					table.AddRow(*env.NewTableRow(vals, &table))
					return table
				}
				return nil
			}
			return nil
		},
	},

	// Tests:
	//  equal {
	//	 table { "a" "b" } { 6 60 7 70 } |get-rows |type?
	//  } 'native
	// Args:
	// * sheet
	"get-rows": {
		Argsn: 1,
		Doc:   "Get rows as a native. This value can be used in `add-rows` and `add-rows!`",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch spr := arg0.(type) {
			case env.Table:
				rows := spr.GetRows()
				return *env.NewNative(ps.Idx, rows, "table-rows")
			default:
				return MakeArgError(ps, 1, []env.Type{env.TableType}, "get-rows")
			}
		},
	},

	// Add one or more rows to a table, returning a new table
	// The `rows` argument can take one of two types:
	// 1) a block that has one or more rows worth of data.
	// This is given as a single flat collection. This means that if your
	//  sheet has `NumColumns` columns, your block should have `NumColumns * NumRows` values.
	// 2) A native that is a slice of TableRows, like the value returned from `get-rows`
	// Tests:
	//  equal {
	//	 ref table { "a" "b" } { 6 60 7 70 } :sheet
	//   sheet .deref |add-rows [ 3 30 ] |length?
	//  } 3
	//  equal {
	//	 ref table { "a" "b" } { 1 80 2 90 } :sheet
	//   sheet .deref |add-rows { 3 30 } |length?
	//  } 3
	// Args:
	// * sheet - the sheet that is getting rows added to it
	// * rows - a block containing one or more rows worth of values, or a TableRow Native value
	"add-rows": {
		Argsn: 2,
		Doc:   "Add one or more rows to a table",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch spr := arg0.(type) {
			case env.Table:
				switch data1 := arg1.(type) {
				case env.Block:
					data := data1.Series
					for data.Pos() < data.Len() {
						rowd := make([]any, len(spr.Cols))
						for ii := 0; ii < len(spr.Cols); ii++ {
							k1 := data.Pop()
							rowd[ii] = k1
						}
						spr.AddRow(*env.NewTableRow(rowd, &spr))
					}
					return spr
				case env.List:
					data := data1.Data
					for item := range data {
						rowd := make([]any, len(spr.Cols))
						for ii := 0; ii < len(spr.Cols); ii++ {
							k1 := item
							rowd[ii] = k1
						}
						spr.AddRow(*env.NewTableRow(rowd, &spr))
					}
					return spr
				case env.Native:
					spr.Rows = append(spr.Rows, data1.Value.([]env.TableRow)...)
					return spr
				default:
					fmt.Println(data1.Inspect(*ps.Idx))
					return MakeArgError(ps, 2, []env.Type{env.BlockType, env.NativeType}, "add-rows")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.TableType}, "add-rows")
			}
		},
	},

	// Add one or more rows to a table ref. Works similarly to `add-rows`, but
	// modified the table ref instead of returning a new copy
	// of the spreasheet
	// Tests:
	//  equal {
	//	 ref table { "a" "b"  } { 1 10 2 20 } :sheet
	//   sheet .add-rows! [ 3 30 ] sheet .deref .length?
	//  } 3
	// Args:
	// * sheet - the reference to the sheet that is getting rows added to it
	// * rows - a block containing one or more rows worth of values, or a TableRow Native value
	// Tags: #spreasheet #mutation
	"add-rows!": {
		Argsn: 2,
		Doc:   "Add one or more rows to a table ref",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch spr := arg0.(type) {
			case *env.Table:
				switch data1 := arg1.(type) {
				case env.Block:
					data := data1.Series
					for data.Pos() < data.Len() {
						rowd := make([]any, len(spr.Cols))
						for ii := 0; ii < len(spr.Cols); ii++ {
							k1 := data.Pop()
							rowd[ii] = k1
						}
						spr.AddRow(*env.NewTableRow(rowd, spr))
					}
					return spr
				case env.Native:
					spr.Rows = append(spr.Rows, data1.Value.([]env.TableRow)...)
					return spr
				default:
					return MakeArgError(ps, 2, []env.Type{env.BlockType, env.NativeType}, "add-rows!")
				}
			case *env.PersistentTable:
				switch data1 := arg1.(type) {
				case env.Block:
					data := data1.Series
					for data.Pos() < data.Len() {
						rowd := make([]any, len(spr.Cols))
						for ii := 0; ii < len(spr.Cols); ii++ {
							k1 := data.Pop()
							rowd[ii] = k1
						}
						spr.AddRow(*env.NewTableRow(rowd, spr))
					}
					return spr
				default:
					return MakeArgError(ps, 2, []env.Type{env.BlockType}, "add-rows!")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.TableType, env.PersistentTableType}, "add-rows!")
			}
		},
	},

	// Update the row at a given index. If given a dict or table row, replace the row with that
	// If given a function, pass the row, its index and replace the row with the return value from the
	// function.
	// Tests:
	//  equal {
	//	 spr1: ref table { "a" "b" } { 1 10 2 20 }
	//	 spr1 .update-row! 1 dict [ "a" 111 ]
	//   spr1 .deref .A1
	//  } 111
	//  equal {
	//	 spr1: ref table { "a" "b" } { 1 10 2 20 }
	//	 incrA: fn { row } { row ++ [ "a" ( "a" <- row ) + 9 ] }
	//	   update-row! spr1 1 ?incrA
	//     spr1 |deref |A1
	//  } 10
	// Args:
	// * sheet-ref - A ref to a table
	// * idx - the index of the row to update, 1-based
	// * updater - One of either a function, a dict, or a Table Row
	// Tags: #table #mutation
	"update-row!": {
		Argsn: 3, // Table, index function/dict
		Doc:   `Update the row at the given index.`,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch spr := arg0.(type) {
			case *env.Table:
				switch idx := arg1.(type) {
				case env.Integer:
					if idx.Value < 1 || (idx.Value-1) > int64(len(spr.Rows)) {
						errMsg := fmt.Sprintf("update-row! called with row index %d, but table only has %d rows", idx.Value, len(spr.Rows))
						return makeError(ps, errMsg)
					}
					switch updater := arg2.(type) {
					case env.Function:
						CallFunction(updater, ps, spr.Rows[idx.Value-1], false, ps.Ctx)
						ret := ps.Res
						if ok, err, row := RyeValueToTableRow(spr, ret); ok {
							spr.Rows[idx.Value-1] = *row
							return spr
						} else if len(err) > 0 {
							return makeError(ps, err)
						} else {
							return makeError(ps, fmt.Sprintf(
								"Function given to update-row! should have returned a Dict or a TableRow, but returned a %s %#v instead",
								NameOfRyeType(ret.Type()), ret,
							))
						}
					case env.Dict:
						// dict should be able to have only the columns you want to update
						row := spr.Rows[idx.Value-1]

						for keyStr, val := range updater.Data {
							index := spr.GetColumnIndex(keyStr)
							if index < 0 {
								return makeError(ps, "Column "+keyStr+" was not found")
							}
							row.Values[index] = val
						}
						return spr
					case env.TableRow:
						spr.Rows[idx.Value-1] = updater
						return spr
					default:
						return MakeArgError(ps, 3, []env.Type{env.FunctionType, env.DictType, env.TableRowType}, "update-row!")
					}
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "update-row!")
				}
			default:
				return MakeNeedsThawedArgError(ps, "update-row!")
			}

		},
	},
	// Tests:
	//  equal {
	//   spr1: ref table { "a" "b" } { 1 10 2 20 }
	//   spr1 .remove-row! 1
	//   spr1 .deref .A1
	//  } 2
	// Args:
	// * sheet-ref
	// * row-idx - Index of row to remove, 1-based
	// Tags: #table #mutation
	"remove-row!": {
		Argsn: 2,
		Doc:   "Remove a row from a table by index",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch spr := arg0.(type) {
			case *env.Table:
				switch data1 := arg1.(type) {
				case env.Integer:
					if data1.Value > 0 && data1.Value <= int64(len(spr.Rows)) {
						spr.RemoveRowByIndex(data1.Value - 1)
						return spr
					} else {
						return makeError(ps, fmt.Sprintf("Table had less then %d rows", data1.Value))
					}
				default:
					return MakeArgError(ps, 2, []env.Type{env.BlockType, env.NativeType}, "remove-row!")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.TableType}, "remove-row!")
			}
		},
	},

	//
	// ##### Column level functions #####  "Functions that construct a table."
	//
	// Example: Select "name" and "age" columns
	//  sheet: table { "name" "age" "job_title" } { "Bob" 25 "Janitor" "Alice" 29 "Librarian" "Charlie" 19 "Line Cook" }
	//  sheet .columns? { 'name 'age }
	// Tests:
	//  equal { table { "name" "age" "job_title" } { "Bob" 25 "Janitor" "Alice" 29 }
	//   |columns? { 'name 'age } |header? } list { "name" "age" }
	"columns?": {
		Argsn: 2,
		Doc:   "Returns table with just given columns. Use lsetwords (:newName) to rename the previous column.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch spr := arg0.(type) {
			case env.Table:
				switch col := arg1.(type) {
				case env.Block:
					cols := make([]string, 0)
					colNames := make([]string, 0)

					for _, obj := range col.Series.S {
						switch ww := obj.(type) {
						case env.String:
							cols = append(cols, ww.Value)
							colNames = append(colNames, ww.Value)
						case env.Tagword:
							colName := ps.Idx.GetWord(ww.Index)
							cols = append(cols, colName)
							colNames = append(colNames, colName)
						case env.LSetword:
							// Rename the previous column
							if len(colNames) == 0 {
								return MakeBuiltinError(ps, "LSetword :"+ps.Idx.GetWord(ww.Index)+" found but no previous column to rename", "columns?")
							}
							// Replace the last column name with the new name from lsetword
							newName := ps.Idx.GetWord(ww.Index)
							colNames[len(colNames)-1] = newName
						default:
							return MakeBuiltinError(ps, "Expected string, tagword, or lsetword in columns specification", "columns?")
						}
					}
					return spr.ColumnsRenamed(ps, cols, colNames)
				default:
					return MakeArgError(ps, 1, []env.Type{env.BlockType}, "columns?")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.TableType}, "columns?")
			}
		},
	},
	// Example: Get sheet column names
	//  sheet: table { "name" "age" "job_title" } { "Bob" 25 "Janitor" "Alice" 29 "Librarian" "Charlie" 19 "Line Cook" }
	//  sheet .header? ; { "name" "age" "job_title" }
	// Tests:
	//  equal { table { "age" "name" } { 123 "Jim" 29 "Anne" }
	//   |header? } list { "age" "name" }
	"header?": {
		Argsn: 1,
		Doc:   "Gets the column names (header) as block.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch spr := arg0.(type) {
			case *env.Table:
				return spr.GetColumns()
			case env.Table:
				return spr.GetColumns()
			default:
				return MakeArgError(ps, 1, []env.Type{env.TableType}, "headers?")
			}
		},
	},

	// Example: Get sheet column names
	//  sheet: table { "name" "age" "job_title" } { "Bob" 25 "Janitor" "Alice" 29 "Librarian" "Charlie" 19 "Line Cook" }
	//  sheet .column? 'name ; => { "Bob" "Alice" "Charlie" }
	// Tests:
	//  equal { table { "name" "age" "job_title" } { "Bob" 25 "Janitor" "Alice" 29 "Teacher" }
	//  |column? 'name }  { "Bob" "Alice" }
	"column?": {
		Argsn: 2,
		Doc:   "Gets all values of a column as a block.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) (res env.Object) {
			switch spr := arg0.(type) {
			case env.Table:
				switch col := arg1.(type) {
				case env.Word:
					res := spr.GetColumn(ps.Idx.GetWord(col.Index))
					if _, isErr := res.(*env.Error); isErr {
						ps.FailureFlag = true
					}
					return res
				case env.String:
					res := spr.GetColumn(col.Value)
					if _, isErr := res.(*env.Error); isErr {
						ps.FailureFlag = true
					}
					return res
				default:
					return MakeArgError(ps, 2, []env.Type{env.WordType, env.StringType}, "column?")
				}
			case env.Block:
				switch col := arg1.(type) {
				case env.Integer:
					col1 := make([]env.Object, len(spr.Series.S))
					if col.Value < 0 {
						return MakeBuiltinError(ps, "Index can't be negative", "column?")
					}
					for i, item_ := range spr.Series.S {
						switch item := item_.(type) {
						case env.Block:
							if len(item.Series.S) < int(col.Value) {
								return MakeBuiltinError(ps, "index out of bounds for item: "+strconv.Itoa(i), "column?")
							}
							col1[i] = item.Series.S[col.Value]
						}
					}
					return *env.NewBlock(*env.NewTSeries(col1))
				default:
					return MakeArgError(ps, 2, []env.Type{env.WordType, env.StringType}, "column?")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.TableType}, "column?")
			}
		},
	},

	// Example: Drop "job_title" column from sheet
	//  sheet: table { "name" "age" "job_title" } { "Bob" 25 "Janitor" "Alice" 29 "Librarian" "Charlie" 19 "Line Cook" }
	//  sheet .drop-column 'job_title ;
	// Tests:
	//  equal { table { "name" "age" "job_title" } { "Bob" 25 "Janitor" "Alice" 29 "Librarian" }
	//  |drop-column "name" |header? } list { "age" "job_title" }
	"drop-column": {
		Argsn: 2,
		Doc:   "Remove a column from a table. Returns new table",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch spr := arg0.(type) {
			case env.Table:
				switch rmCol := arg1.(type) {
				case env.Word:
					return DropColumn(ps, spr, *env.NewString(ps.Idx.GetWord(rmCol.Index)))
				case env.String:
					return DropColumn(ps, spr, rmCol)
				case env.Block:
					return DropColumnBlock(ps, spr, rmCol)
				default:
					return MakeArgError(ps, 2, []env.Type{env.WordType, env.StringType, env.BlockType}, "drop-column")
				}
			}
			return MakeArgError(ps, 1, []env.Type{env.TableType}, "drop-column")
		},
	},
	// Tests:
	//  equal { tab: ref table { "name" "age" } { "Bob" 25 "Alice" 29 "Charlie" 19 }
	//  tab .rename-column! "name" "first_name" , tab .header? } list { "first_name" "age" }
	"rename-column!": {
		Argsn: 3,
		Doc:   "Remove a column from a table. Returns new table",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch spr := arg0.(type) {
			case *env.Table:
				switch oldName := arg1.(type) {
				case env.String:
					switch newName := arg2.(type) {
					case env.String:
						return RenameColumn(ps, spr, oldName, newName)
					default:
						return MakeArgError(ps, 2, []env.Type{env.WordType, env.BlockType}, "rename-column")
					}
				default:
					return MakeArgError(ps, 2, []env.Type{env.WordType, env.BlockType}, "rename-column")
				}
			}
			return MakeArgError(ps, 1, []env.Type{env.TableType}, "rename-column")
		},
	},
	// TODO: also create add-column\i 'name { +100 }   ( or  \idx \pos)
	// Example: Add a column to a sheet
	//  sheet: table { "name" "age" } { "Bob" 25 "Alice" 29 "Charlie" 19 }
	//  sheet .add-column 'job_title { "Jantior" "Librarian" "Line Cook" } ;
	// Tests:
	//  equal { table { "name" "age" } { "Bob" 25 "Alice" 29 "Charlie" 19 }
	//  |add-column 'job { } { "Cook" } |column? "job" } { "Cook" "Cook" "Cook" }
	"add-column": {
		Argsn: 4,
		Doc:   "Adds a new column to table. Changes in-place and returns the new table.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch spr := arg0.(type) {
			case env.Table:
				switch newCol := arg1.(type) {
				case env.Word:
					switch fromCols := arg2.(type) {
					case env.Block:
						switch code := arg3.(type) {
						case env.Block:
							return GenerateColumn(ps, spr, newCol, fromCols, code)
						default:
							return MakeArgError(ps, 4, []env.Type{env.BlockType}, "add-column")
						}
					case env.Word:
						switch replaceBlock := arg3.(type) {
						case env.Block:
							if replaceBlock.Series.Len() != 2 {
								return MakeBuiltinError(ps, "Replacement block must contain a regex object and replacement string.", "add-column")
							}
							regexNative, ok := replaceBlock.Series.S[0].(env.Native)
							if !ok {
								return MakeBuiltinError(ps, "First element of replacement block must be a regex object.", "add-column")
							}
							regex, ok := regexNative.Value.(*regexp.Regexp)
							if !ok {
								return MakeBuiltinError(ps, "First element of replacement block must be a regex object.", "add-column")
							}
							replaceStr, ok := replaceBlock.Series.S[1].(env.String)
							if !ok {
								return MakeBuiltinError(ps, "Second element of replacement block must be a string.", "add-column")
							}
							err := GenerateColumnRegexReplace(ps, &spr, newCol, fromCols, regex, replaceStr.Value)
							if err != nil {
								return err
							}
							return spr
						default:
							return MakeArgError(ps, 3, []env.Type{env.BlockType}, "add-column")
						}
					default:
						return MakeArgError(ps, 3, []env.Type{env.BlockType}, "add-column")
					}
				default:
					return MakeArgError(ps, 2, []env.Type{env.WordType}, "add-column")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.TableType}, "add-column")
			}
		},
	},

	//
	// ##### Miscelaneous #####  ""
	//
	// Example: Order by age ascending
	//  sheet: table { "name" "age" } { "Bob" 25 "Alice" 29 "Charlie" 19  }
	//  sheet .order-by! 'age 'asc
	// Tests:
	//  equal { tab: table { "name" "age" } { "Bob" 25 "Alice" 29 "Charlie" 19  }
	//  tab .order-by! 'age 'asc |column? "age" } { 19 25 29 }
	"order-by!": {
		Argsn: 3,
		Doc:   "Sorts row by given column, changes table in place.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			dir, ok := arg2.(env.Word)
			if !ok {
				return MakeArgError(ps, 3, []env.Type{env.WordType}, "sort-by!")
			}
			var dirAsc bool
			if dir.Index == ps.Idx.IndexWord("asc") {
				dirAsc = true
			} else if dir.Index == ps.Idx.IndexWord("desc") {
				dirAsc = false
			} else {
				return MakeBuiltinError(ps, "Direction can be just asc or desc.", "sort-by!")
			}
			switch spr := arg0.(type) {
			case env.Table:
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
					return MakeArgError(ps, 2, []env.Type{env.WordType}, "sort-by!")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.TableType}, "sort-by!")
			}
		},
	},

	// Example: Order by age ascending
	//  sheet: table { "name" "age" } { "Bob" 25 "Alice" 29 "Charlie" 19  }
	//  sheet .order-by 'age 'asc
	// Tests:
	//  equal { tab: table { "name" "age" } { "Bob" 25 "Alice" 29 "Charlie" 19  }
	//  |order-by 'age 'desc |column? "age" } { 29 25 19 }
	"order-by": {
		Argsn: 3,
		Doc:   "Sorts row by given column, changes table in place.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			dir, ok := arg2.(env.Word)
			if !ok {
				return MakeArgError(ps, 3, []env.Type{env.WordType}, "sort-by!")
			}
			var dirAsc bool
			if dir.Index == ps.Idx.IndexWord("asc") {
				dirAsc = true
			} else if dir.Index == ps.Idx.IndexWord("desc") {
				dirAsc = false
			} else {
				return MakeBuiltinError(ps, "Direction can be just asc or desc.", "sort-by!")
			}
			switch spr := arg0.(type) {
			case env.Table:
				copied := make([]env.TableRow, len(spr.Rows))
				copy(copied, spr.Rows)
				newSpr := env.NewTable(spr.Cols)
				newSpr.Rows = copied
				switch col := arg1.(type) {
				case env.String:
					if dirAsc {
						SortByColumn(ps, newSpr, col.Value)
					} else {
						SortByColumnDesc(ps, newSpr, col.Value)
					}
					return *newSpr
				case env.Word:
					if dirAsc {
						SortByColumn(ps, newSpr, ps.Idx.GetWord(col.Index))
					} else {
						SortByColumnDesc(ps, newSpr, ps.Idx.GetWord(col.Index))
					}
					return *newSpr
				default:
					return MakeArgError(ps, 2, []env.Type{env.WordType}, "sort-by!")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.TableType}, "sort-by!")
			}
		},
	},

	// Tests:
	//  equal { table { "name" "age" } { "Bob" 25 "Alice" 29 "Charlie" 19 }
	//  |add-indexes! { name } |indexes? } { "name" }
	"add-indexes!": {
		Argsn: 2,
		Doc:   "Creates an index for all values in the provided columns. Changes in-place and returns the new table.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch spr := arg0.(type) {
			case env.Table:
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
					return MakeArgError(ps, 2, []env.Type{env.BlockType}, "add-indexes!")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.TableType}, "add-indexes!")
			}
		},
	},
	// Tests:
	//  equal { table { "name" "age" } { "Bob" 25 "Alice" 29 "Charlie" 19 }
	//  |add-indexes! { name age } |indexes? } { "name" "age" }
	"indexes?": {
		Argsn: 1,
		Doc:   "Returns the columns that are indexed in a table.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) (res env.Object) {
			switch spr := arg0.(type) {
			case env.Table:
				res := make([]env.Object, 0)
				for col := range spr.Indexes {
					res = append(res, *env.NewString(col))
				}
				return *env.NewBlock(*env.NewTSeries(res))
			default:
				return MakeArgError(ps, 1, []env.Type{env.TableType}, "indexes?")
			}
		},
	},
	// Tests:
	//  equal { table { "age" } { 123 29 19 }
	//  |autotype 1.0 |types? } { integer }
	"autotype": {
		Argsn: 2,
		Doc:   "Takes a table and tries to determine and change the types of columns.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) (res env.Object) {
			switch spr := arg0.(type) {
			case env.Table:
				switch percent := arg1.(type) {
				case env.Decimal:
					return AutoType(ps, &spr, percent.Value)
				default:
					return MakeArgError(ps, 2, []env.Type{env.DecimalType}, "autotype")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.TableType}, "autotype")
			}
		},
	},
	// Example: join two tables, putting in empty cells if the left sheet doesn't have a value
	//  names: table { "id" "name" } { 1 "Paul" 2 "Chani" 3 "Vladimir" } ,
	//  houses: table { "id" "house" } { 1 "Atreides" 3 "Harkonnen" } ,
	//  names .left-join houses 'id 'id
	// Tests:
	//  equal {
	//    names: table { "id" "name" } { 1 "Paul" 2 "Chani" 3 "Vladimir" } ,
	//    houses: table { "id" "house" } { 1 "Atreides" 3 "Harkonnen" } ,
	//
	//    names .inner-join houses 'id 'id |header?
	//  } list { "id" "name" "id_2" "house" }
	//  equal {
	//    names: table { "id" "name" } { 1 "Paul" 2 "Chani" 3 "Vladimir" } ,
	//    houses: table { "id" "house" } { 1 "Atreides" 3 "Harkonnen" } ,
	//
	//    names .left-join houses 'id 'id |column? "name"
	//  } { "Paul" "Chani" "Vladimir" }
	"left-join": {
		Argsn: 4,
		Doc:   "Left joins two tables on the given columns.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) (res env.Object) {
			switch spr1 := arg0.(type) {
			case env.Table:
				switch spr2 := arg1.(type) {
				case env.Table:
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
					return MakeArgError(ps, 2, []env.Type{env.TableType}, "left-join")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.TableType}, "left-join")
			}
		},
	},
	// Example: join two tables
	//  names: table { "id" "name" } { 1 "Paul" 2 "Chani" 3 "Vladimir" } ,
	//  houses: table { "id" "house" } { 1 "Atreides" 3 "Harkonnen" } ,
	//  names .inner-join houses 'id 'id

	// Tests:
	//  equal {
	//    names: table { "id" "name" } { 1 "Paul" 2 "Chani" 3 "Vladimir" } ,
	//    houses: table { "id" "house" } { 1 "Atreides" 3 "Harkonnen" } ,
	//
	//    names .inner-join houses 'id 'id |header?
	//  } list { "id" "name" "id_2" "house" }
	//  equal {
	//    names: table { "id" "name" } { 1 "Paul" 2 "Chani" 3 "Vladimir" } ,
	//    houses: table { "id" "house" } { 1 "Atreides" 3 "Harkonnen" } ,
	//
	//    names .inner-join houses 'id 'id |column? "name"
	//  } {  "Paul" "Vladimir" }
	"inner-join": {
		Argsn: 4,
		Doc:   "Inner joins two tables on the given columns.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) (res env.Object) {
			switch spr1 := arg0.(type) {
			case env.Table:
				switch spr2 := arg1.(type) {
				case env.Table:
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
					return MakeArgError(ps, 2, []env.Type{env.TableType}, "inner-join")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.TableType}, "inner-join")
			}
		},
	},
	// Example: group table rows by name, running various aggregations on the val column
	//  table { "name" "val" } { "a" 1 "b" 6 "a" 5 "b" 10 "a" 7 }
	// 	|group-by 'name { 'name count 'val sum 'val min 'val max 'val avg }
	// 	|order-by! 'name 'asc
	// Tests:
	//  equal {
	//    table { "name" "val" } { "a" 1 "b" 6 "a" 5 "b" 10 "a" 7 }
	// 	  |group-by 'name { 'name count 'val sum } |column? "val_sum" |sort
	//   } { 13.0 16.0 }
	// Tests:
	//  equal {
	//    table { "name" "val" } { "a" 1 "b" 6 "a" 5 "b" 10 "a" 7 }
	// 	  |group-by 'name { 'name count 'val min } |column? "val_min" |sort
	//   } { 1.0 6.0 }
	// Tests:
	//  equal {
	//    table { "name" "val" } { "a" 1 "b" 6 "a" 5 "b" 10 "a" 12 }
	// 	  |group-by 'name { 'name count 'val avg } |column? "val_avg" |sort
	//   } { 6.0 8.0 }
	"group-by": {
		Argsn: 3,
		Doc:   "Groups a table by the given column(s) and (optional) aggregations.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) (res env.Object) {
			switch spr := arg0.(type) {
			case env.Table:
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
						return GroupBy(ps, spr, []string{ps.Idx.GetWord(col.Index)}, aggregations)
					case env.String:
						return GroupBy(ps, spr, []string{col.Value}, aggregations)
					case env.Block:
						cols := make([]string, col.Series.Len())
						for c := range col.Series.S {
							switch ww := col.Series.S[c].(type) {
							case env.String:
								cols[c] = ww.Value
							case env.Tagword:
								cols[c] = ps.Idx.GetWord(ww.Index)
							case env.Word:
								cols[c] = ps.Idx.GetWord(ww.Index)
							default:
								return MakeBuiltinError(ps, "Block must contain only strings or words for column names", "group-by")
							}
						}
						return GroupBy(ps, spr, cols, aggregations)
					default:
						return MakeArgError(ps, 2, []env.Type{env.WordType, env.StringType, env.BlockType}, "group-by")
					}
				default:
					return MakeArgError(ps, 3, []env.Type{env.BlockType}, "group-by")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.TableType}, "group-by")
			}
		},
	},

	// Tests:
	// equal { table { 'a } { 123 234 345 } |A1 } 123
	"A1": {
		Argsn: 1,
		Doc:   "Accepts a Table and returns the first row first column cell.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s0 := arg0.(type) {
			case env.Table:
				r := s0.Rows[0].Values[0]
				return env.ToRyeValue(r)

			default:
				ps.ErrorFlag = true
				return env.NewError("first arg not table")
			}
		},
	},
	// TODO: Check for size

	// Tests:
	// equal { table { 'a 'b } { 123 234 345 456 } |B1 } 234
	"B1": {
		Argsn: 1,
		Doc:   "Accepts a Table and returns the first row second column cell.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s0 := arg0.(type) {
			case env.Table:
				r := s0.Rows[0].Values[1]
				return env.ToRyeValue(r)

			default:
				ps.ErrorFlag = true
				return env.NewError("first arg not table")
			}
		},
	},
	// TODO: check for size

	// Tests:
	// equal { table { 'a 'b } { 123 234 345 456 } |A2 } 345
	"A2": {
		Argsn: 1,
		Doc:   "Accepts a Table and returns the second row first column cell.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s0 := arg0.(type) {
			case env.Table:
				r := s0.Rows[1].Values[0]
				return env.ToRyeValue(r)

			default:
				ps.ErrorFlag = true
				return env.NewError("first arg not table")
			}
		},
	},
	// Tests:
	// equal { table { 'a 'b } { 123 234 345 456 } |B2 } 456
	"B2": {
		Argsn: 1,
		Doc:   "Accepts a Table and returns the second row second column cell.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s0 := arg0.(type) {
			case env.Table:
				r := s0.Rows[1].Values[1]
				return env.ToRyeValue(r)

			default:
				ps.ErrorFlag = true
				return env.NewError("first arg not table")
			}
		},
	},

	//
	// ##### Loading and saving #####  "Functions that construct a table."
	//
	// Tests:
	//  equal {
	//	 cc os
	//   f:: mktmp ++ "/test.csv"
	//   spr1:: table { "a" "b" "c" } { 1 1.1 "a" 2 2.2 "b" 3 3.3 "c" }
	//   spr1 .save\csv f
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
					return MakeBuiltinError(ps, "Unable to read input file:"+err.Error(), "Load\\csv")
				}
				defer f.Close()

				csvReader := csv.NewReader(f)
				rows, err := csvReader.ReadAll()
				if err != nil {
					// log.Fatal("Unable to parse file as CSV for "+filePath, err)
					return MakeBuiltinError(ps, "Unable to parse file as CSV: "+err.Error(), "Load\\csv")
				}
				if len(rows) == 0 {
					return MakeBuiltinError(ps, "File is empty", "Load\\csv")
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
				return MakeArgError(ps, 1, []env.Type{env.UriType}, "Load\\csv")
			}
		},
	},
	//
	// ##### Loading and saving #####  "Functions that construct a table."
	//
	// Tests:
	//  equal {
	//	 cc os
	//   f:: mktmp ++ "/test.csv"
	//   spr1:: table { "a" "b" "c" } { 1 1.1 "a" 2 2.2 "b" 3 3.3 "c" }
	//   spr1 .save\csv f
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
						return MakeBuiltinError(ps, "Unable to read input file:"+err.Error(), "Load\\csv")
					}
					defer f.Close()

					csvReader := csv.NewReader(f)
					if len(separator.Value) != 1 {
						return MakeBuiltinError(ps, "Separator must be exactly 1 character long", "Load\\csv\\")
					}
					csvReader.Comma = rune(separator.Value[0])
					rows, err := csvReader.ReadAll()
					if err != nil {
						// log.Fatal("Unable to parse file as CSV for "+filePath, err)
						return MakeBuiltinError(ps, "Unable to parse file as CSV: "+err.Error(), "Load\\csv")
					}
					if len(rows) == 0 {
						return MakeBuiltinError(ps, "File is empty", "Load\\csv")
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
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "Load\\csv\\")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.UriType}, "Load\\csv\\")
			}
		},
	},

	// TODO -- deduplicate with above

	// Tests:
	//  equal {
	//	 cc os
	//   f:: mktmp ++ "/test.tsv"
	//   spr1:: table { "a" "b" "c" } { 1 1.1 "a" 2 2.2 "b" 3 3.3 "c" }
	//   spr1 .save\tsv f
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
					return MakeBuiltinError(ps, "Unable to read input file:"+err.Error(), "Load\\csv")
				}
				defer f.Close()

				csvReader := csv.NewReader(f)
				csvReader.Comma = '\t'
				rows, err := csvReader.ReadAll()
				if err != nil {
					// log.Fatal("Unable to parse file as CSV for "+filePath, err)
					return MakeBuiltinError(ps, "Unable to parse file as CSV: "+err.Error(), "Load\\csv")
				}
				if len(rows) == 0 {
					return MakeBuiltinError(ps, "File is empty", "Load\\csv")
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
				return MakeArgError(ps, 1, []env.Type{env.UriType}, "Load\\csv")
			}
		},
	},

	// Tests:
	//  equal {
	//	 cc os
	//   f:: mktmp ++ "/test.csv"
	//   spr1:: table { "a" "b" "c" } { 1 1.1 "a" 2 2.2 "b" 3 3.3 "c" }
	//   spr1 .save\csv f
	//   spr2:: Load\csv f |autotype 1.0
	//   spr1 = spr2
	//  } true
	// Args:
	// * sheet    - the sheet to save
	// * file-url - where to save the sheet as a .csv file
	// Tags: #table #saving #csv
	"save\\csv": {
		Argsn: 2,
		Doc:   "Saves a table to a .csv file.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch spr := arg0.(type) {
			case env.Table:
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
							if i < cLen {
								strVals[i] = sv
							}
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

	// TODO: deduplicate with load\csv

	// Tests:
	//  equal {
	//	 cc os
	//   f:: mktmp ++ "/test.csv"
	//   spr1:: table { "a" "b" "c" } { 1 1.1 "a" 2 2.2 "b" 3 3.3 "c" }
	//   spr1 .save\tsv f
	//   spr2:: Load\tsv f |autotype 1.0
	//   spr1 = spr2
	//  } true
	// Args:
	// * sheet    - the table to save
	// * file-url - where to save the sheet as a .csv file
	// Tags: #table #saving #csv
	"save\\tsv": {
		Argsn: 2,
		Doc:   "Saves a table to a .csv file.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch spr := arg0.(type) {
			case env.Table:
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
					csvWriter.Comma = '\t'
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
							if i < cLen {
								strVals[i] = sv
							}
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

	// Tests:
	//  equal {
	//	 cc os
	//   f:: mktmp ++ "/test.xlsx"
	//   spr1:: table { "a" "b" "c" } { 1 1.1 "a" 2 2.2 "b" 3 3.3 "c" }
	//   spr1 .save\xlsx f
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
					return MakeBuiltinError(ps, fmt.Sprintf("Unable to open file: %s", err), "Load\\xlsx")
				}
				defer f.Close()

				sheetMap := f.GetSheetMap()
				if len(sheetMap) == 0 {
					return MakeBuiltinError(ps, "No sheets found in file", "Load\\xlsx")
				}
				// sheets map index is 1-based
				sheetName := sheetMap[1]
				rows, err := f.Rows(sheetName)
				if err != nil {
					return MakeBuiltinError(ps, fmt.Sprintf("Unable to get rows from sheet: %s", err), "Load\\xlsx")
				}
				rows.Next()
				header, err := rows.Columns()
				if err != nil {
					return MakeBuiltinError(ps, fmt.Sprintf("Unable to get columns from sheet: %s", err), "Load\\xlsx")
				}
				if len(header) == 0 {
					return MakeBuiltinError(ps, "Header row is empty", "Load\\xlsx")
				}
				spr := env.NewTable(header)
				for rows.Next() {
					row, err := rows.Columns()
					if err != nil {
						return MakeBuiltinError(ps, fmt.Sprintf("Unable to get row: %s", err), "Load\\xlsx")
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
				return MakeArgError(ps, 1, []env.Type{env.UriType}, "Load\\xlsx")
			}
		},
	},

	// Tests:
	//  equal {
	//	 cc os
	//   f:: mktmp ++ "/test.xlsx"
	//   spr1:: table { "a" "b" "c" } { 1 1.1 "a" 2 2.2 "b" 3 3.3 "c" }
	//   spr1 .save\xlsx f
	//   spr2:: Load\xlsx f |autotype 1.0
	//   spr1 = spr2
	//  } true
	// Args:
	// * table    - the table to save
	// * file-url 		- where to save the table as a .xlsx file
	// Tags: #table #saving #xlsx
	"save\\xlsx": {
		Argsn: 2,
		Doc:   "Saves a Table to a .xlsx file.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch spr := arg0.(type) {
			case env.Table:
				switch file := arg1.(type) {
				case env.Uri:
					sheetName := "Sheet1"
					f := excelize.NewFile()
					index, err := f.NewSheet(sheetName)
					if err != nil {
						return MakeBuiltinError(ps, fmt.Sprintf("Unable to create new sheet: %s", err), "save\\xlsx")
					}
					err = f.SetSheetRow(sheetName, "A1", &spr.Cols)
					if err != nil {
						return MakeBuiltinError(ps, fmt.Sprintf("Unable to set header row: %s", err), "save\\xlsx")
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
								return MakeBuiltinError(ps, fmt.Sprintf("Unable to save table: unsupported type %T", val), "save\\xlsx")
							}
						}
						err = f.SetSheetRow(sheetName, fmt.Sprintf("A%d", rowIndex), &vals)
						if err != nil {
							return MakeBuiltinError(ps, fmt.Sprintf("Unable to set row %d: %s", rowIndex, err), "save\\xlsx")
						}
					}
					f.SetActiveSheet(index)
					err = f.SaveAs(file.GetPath())
					if err != nil {
						return MakeBuiltinError(ps, fmt.Sprintf("Unable to save table: %s", err), "save\\xlsx")
					}
					return spr
				default:
					return MakeArgError(ps, 1, []env.Type{env.UriType}, "save\\xlsx")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.TableType}, "save\\xlsx")
			}
		},
	},

	// Decided to remove ... .header? .length? seems better than new "vocabulary entry" for frequency of use I expect
	// TODO --- moved from base ... name appropriately and deduplicate
	// Args:
	// * sheet
	/*
		"ncols": {
			Doc:   "Accepts a Table and returns number of columns.",
			Argsn: 1,
			Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
				switch s1 := arg0.(type) {
				case env.Dict:
				case env.Block:
				case env.Table:
					return *env.NewInteger(int64(len(s1.Cols)))
				default:
					fmt.Println("Error")
				}
				return nil
			},
		}, */
	/* 20250126 -- removed ... column? sum makes more sense than another specific word
	// Tests:
	// equal { table { 'a } { 1 2 3 } |col-sum "a" } 6
	"col-sum": {
		Argsn: 2,
		Doc:   "Accepts a table and a column name and returns a sum of a column.", // TODO -- let it accept a block and list also
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			var name string
			switch s1 := arg0.(type) {
			case env.Table:
				switch s2 := arg1.(type) {
				case env.Word:
					name = ps.Idx.GetWord(s2.Index)
				case env.String:
					name = s2.Value
				default:
					ps.ErrorFlag = true
					return env.NewError("second arg not string")
				}
				r := s1.Sum(name)
				if r.Type() == env.ErrorType {
					ps.ErrorFlag = true
				}
				return r

			default:
				ps.ErrorFlag = true
				return env.NewError("first arg not table")
			}
		},
	},

	// Tests:
	// equal { table { 'a } { 1 2 3 } |col-avg 'a } 2.0
	"col-avg": {
		Argsn: 2,
		Doc:   "Accepts a table and a column name and returns a sum of a column.", // TODO -- let it accept a block and list also
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			var name string
			switch s1 := arg0.(type) {
			case env.Table:
				switch s2 := arg1.(type) {
				case env.Word:
					name = ps.Idx.GetWord(s2.Index)
				case env.String:
					name = s2.Value
				default:
					ps.ErrorFlag = true
					return env.NewError("second arg not string")
				}
				r, err := s1.Sum_Just(name)
				if err != nil {
					ps.ErrorFlag = true
					return env.NewError(err.Error())
				}
				n := s1.NRows()
				return *env.NewDecimal(r / float64(n))

			default:
				ps.ErrorFlag = true
				return env.NewError("first arg not table")
			}
		},
	}, */
	// TODO: Check for size

	//
	// ##### Persistent Tables #####  "Functions for persistent tables using BadgerDB."
	//

	// Tests:
	//  ; equal { persistent-table { "a" "b" } "/tmp/test_db" "test_table" |type? } 'persistent-table
	// Args:
	//  * columns - block of column names
	//  * db-path - path to BadgerDB database
	//  * table-name - name of the table
	"persistent-table": {
		Argsn: 3,
		Doc:   "Creates a persistent table using BadgerDB for storage",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			cols, err := ColNames(ps, arg0, "persistent-table")
			if err != nil {
				return err
			}

			switch dbPath := arg1.(type) {
			case env.String:
				switch tableName := arg2.(type) {
				case env.String:
					pt, err := env.NewPersistentTable(cols, dbPath.Value, tableName.Value)
					if err != nil {
						return MakeBuiltinError(ps, fmt.Sprintf("Failed to create persistent table: %v", err), "persistent-table")
					}
					return pt
				default:
					return MakeArgError(ps, 3, []env.Type{env.StringType}, "persistent-table")
				}
			default:
				return MakeArgError(ps, 2, []env.Type{env.StringType}, "persistent-table")
			}
		},
	},

	// Tests:
	//  ; equal { pt: persistent-table { "a" "b" } "/tmp/test_db" "test_table"
	//  ;        pt .close-persistent-table! |type? } 'persistent-table
	// Args:
	//  * persistent-table - the persistent table to close
	"close-persistent-table!": {
		Argsn: 1,
		Doc:   "Closes the BadgerDB connection for a persistent table",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch pt := arg0.(type) {
			case *env.PersistentTable:
				err := pt.Close()
				if err != nil {
					return MakeBuiltinError(ps, fmt.Sprintf("Failed to close persistent table: %v", err), "close-persistent-table!")
				}
				return pt
			default:
				return MakeArgError(ps, 1, []env.Type{env.PersistentTableType}, "close-persistent-table!")
			}
		},
	},
}

func RyeValueToTableRow(spr *env.Table, obj env.Object) (bool, string, *env.TableRow) {
	switch updater := obj.(type) {
	case env.Dict:
		success, missing, row := env.TableRowFromDict(updater, spr)
		if !success {
			return false, "update-row! given a dict that is missing value for the " + missing + " column!", nil
		} else {
			return true, "", row
		}
	case env.TableRow:
		return true, "", &updater
	default:
		return false, "", nil
	}
}

func ColNames(ps *env.ProgramState, from env.Object, fnName string) ([]string, env.Object) {
	switch columns := from.(type) {
	case env.Block:
		colNames := columns.Series
		numCols := colNames.Len()
		if numCols == 0 {
			return nil, MakeBuiltinError(ps, "Block of column names is empty", fnName)
		}
		cols := make([]string, numCols)
		for colNames.Pos() < numCols {
			i := colNames.Pos()
			k1 := colNames.Pop()
			switch k := k1.(type) {
			case env.String:
				cols[i] = k.Value
			case env.Tagword:
				cols[i] = ps.Idx.GetWord(k.Index)
			default:
				return nil, MakeBuiltinError(ps, fmt.Sprintf("Expected a string or word instead of %V", k), fnName)
			}
			// TODO: Error here?
		}
		return cols, nil
	case env.List:
		colNames := columns.Data
		numCols := len(colNames)
		if numCols == 0 {
			return nil, MakeBuiltinError(ps, "Block of column names is empty", fnName)
		}
		cols := make([]string, numCols)
		for i, k1 := range colNames {
			switch k := k1.(type) {
			case env.String:
				cols[i] = k.Value
			case env.Word:
				cols[i] = ps.Idx.GetWord(k.Index)
			default:
				return nil, MakeBuiltinError(ps, fmt.Sprintf("Expected a string or word instead of %V", k), fnName)
			}
			// TODO: Error here?
		}
		return cols, nil
	default:
		return nil, MakeBuiltinError(ps, fmt.Sprintf("Expected a block or a list instead of %V", from), fnName)
	}
}

func MakeColError(ps *env.ProgramState, builtinName string, colName string, expectedRowCount int, actualRowCount int) *env.Error {
	return MakeBuiltinError(
		ps,
		fmt.Sprintf("Column %s should have %d rows of data, but has %d instead",
			colName,
			expectedRowCount,
			actualRowCount,
		),
		builtinName,
	)
}

func LoadColumnData(ps *env.ProgramState, data any, colIdx int, numRows int, colData []env.TableRow, cols []string) *env.Error {
	switch colSeries := data.(type) {
	case env.Block:
		if colSeries.Series.Len() != numRows {
			return MakeColError(ps, "table\\columns", cols[colIdx], numRows, colSeries.Series.Len())
		}
		for rowIdx, value := range colSeries.Series.S {
			colData[rowIdx].Values[colIdx] = value
		}
	case env.List:
		if len(colSeries.Data) != numRows {
			return MakeColError(ps, "table\\columns", cols[colIdx], numRows, len(colSeries.Data))
		}
		for rowIdx, value := range colSeries.Data {
			colData[rowIdx].Values[colIdx] = value
		}
	}
	return nil
}

func GetNumRowsFrom(ps *env.ProgramState, data any) (int, *env.Error) {
	switch firstCol := data.(type) {
	case env.Block:
		return firstCol.Series.Len(), nil
	case env.List:
		return len(firstCol.Data), nil
	default:
		return -1, MakeBuiltinError(ps, fmt.Sprintf("Expected a block or a list instead of %V", firstCol), "table\\columns")
	}
}

func SheetFromColumnsMapData(ps *env.ProgramState, cols []string, arg1 env.Object) env.Object {
	spr := *env.NewTable(cols)
	numCols := len(cols)

	var colData []env.TableRow

	switch colSet := arg1.(type) {
	case env.Block:
		blockColData := colSet.Series.S
		numRows, err := GetNumRowsFrom(ps, blockColData[0])
		if err != nil {
			return err
		}

		colData = make([]env.TableRow, numRows)
		for rowIdx := 0; rowIdx < numRows; rowIdx++ {
			colData[rowIdx].Values = make([]any, numCols)
		}

		for colIdx, c := range blockColData {
			err := LoadColumnData(ps, c, colIdx, numRows, colData, cols)
			if err != nil {
				return err
			}
		}
		spr.Rows = colData
		return spr
	case env.List:
		blockColData := colSet.Data
		numRows, err := GetNumRowsFrom(ps, blockColData[0])
		if err != nil {
			return err
		}

		colData = make([]env.TableRow, numRows)
		for rowIdx := 0; rowIdx < numRows; rowIdx++ {
			colData[rowIdx].Values = make([]any, numCols)
		}
		for colIdx, c := range blockColData {
			err := LoadColumnData(ps, c, colIdx, numRows, colData, cols)
			if err != nil {
				return err
			}
		}
		return spr
	default:
		return MakeBuiltinError(ps, fmt.Sprintf("Expected either a Block of a list of data columns, got %v instead", colSet), "table\\columns")
	}
}

func SheetFromColumns(ps *env.ProgramState, arg0 env.Object, arg1 env.Object) (res env.Object) {
	cols, err := ColNames(ps, arg0, "table\\columns")
	if err != nil {
		return err
	}

	return SheetFromColumnsMapData(ps, cols, arg1)
}

func TableRowsFromBlockOrList(ps *env.ProgramState, spr *env.Table, numCols int, arg1 any) (*env.TableRow, *env.Error) {
	switch row := arg1.(type) {
	case env.Block:
		if len(row.Series.S) != numCols {
			return nil, MakeBuiltinError(
				ps,
				fmt.Sprintf("All rows must have the same number elements as the number of columns (%d)", numCols),
				"table\\rows",
			)
		}
		rowAny := make([]any, len(row.Series.S))
		for i, d := range row.Series.S {
			rowAny[i] = d
		}

		return env.NewTableRow(rowAny, spr), nil
	case env.List:
		if len(row.Data) != numCols {
			return nil, MakeBuiltinError(
				ps,
				fmt.Sprintf("All rows must have the same number elements as the number of columns (%d)", numCols),
				"table\\rows",
			)
		}
		return env.NewTableRow(row.Data, spr), nil
	default:
		return nil, MakeBuiltinError(ps, "Rows must be blocks or lists", "table\\rows")
	}
}

func DropColumnBlock(ps *env.ProgramState, s env.Table, names env.Block) env.Object {
	toDrop := make([]env.String, 0)
	for _, obj := range names.Series.S {
		switch word := obj.(type) {
		case env.Word:
			toDrop = append(toDrop, *env.NewString(ps.Idx.GetWord(word.Index)))
		default:
			return MakeError(ps, "Cannot use a non-word to specify a column to drop")
		}
	}
	return DropColumns(ps, s, toDrop)
}

func DropColumn(ps *env.ProgramState, s env.Table, name env.String) env.Object {
	return DropColumns(ps, s, []env.String{name})
}

// Drop one or more columns from a table, returning a new table
func DropColumns(ps *env.ProgramState, s env.Table, names []env.String) env.Object {
	var columnsToCopy []int = make([]int, len(s.Cols)-len(names))
	var keepColIdx int = 0

	for colIdx, col := range s.Cols {
		keep := true
		for _, name := range names {
			nameStr := name.Value
			if col == nameStr {
				keep = false
				break
			}
		}
		if keep {
			columnsToCopy[keepColIdx] = colIdx
			keepColIdx++
		}
	}

	newCols := make([]string, len(columnsToCopy))

	for toIdx, fromIdx := range columnsToCopy {
		newCols[toIdx] = s.Cols[fromIdx]
	}

	newSheet := env.NewTable(newCols)
	newSheet.Rows = make([]env.TableRow, len(s.Rows))

	for rowIdx, row := range s.Rows {
		newValues := make([]any, len(columnsToCopy))
		for toIdx, fromIdx := range columnsToCopy {
			newValues[toIdx] = row.Values[fromIdx]
		}
		newSheet.Rows[rowIdx] = *env.NewTableRow(newValues, newSheet)
	}
	newSheet.Indexes = make(map[string]map[any][]int)

	for _, colName := range newCols {
		newSheet.Indexes[colName] = s.Indexes[colName]
	}
	newSheet.Kind = s.Kind

	return newSheet
}

// Drop one or more columns from a table, returning a new table
func RenameColumn(ps *env.ProgramState, s *env.Table, oldName env.String, newName env.String) env.Object {
	var colI int

	for i, name := range s.Cols {
		if name == oldName.Value {
			colI = i
			break
		}
	}

	s.Cols[colI] = newName.Value
	return s
}

func GenerateColumn(ps *env.ProgramState, s env.Table, name env.Word, extractCols env.Block, code env.Block) env.Object {
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
				if val == nil {
					val = ""
				}
				// fmt.Println(val)
				if er != nil {
					return MakeBuiltinError(ps, er.Error(), "add-column!")
				}
				if firstVal == nil {
					var ok bool
					firstVal, ok = val.(env.Object)
					if !ok {
						firstVal = *env.NewString(val.(string))
					}
				}
				val1, ok := val.(env.Object)
				if !ok {
					val1 = *env.NewString(val.(string))
				}
				ctx.Set(w.Index, val1)
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
		row.Uplink = &s
		s.Rows[ix] = row
	}
	return s
}

func GenerateColumnRegexReplace(ps *env.ProgramState, s *env.Table, name env.Word, fromColName env.Word, re *regexp.Regexp, pattern string) env.Object {
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

func AddIndexes(ps *env.ProgramState, s *env.Table, columns []env.Word) env.Object {
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

func SortByColumn(ps *env.ProgramState, s *env.Table, name string) {
	idx := slices.Index[[]string](s.Cols, name)

	compareCol := func(i, j int) bool {
		if s.Rows[i].Values[idx] == nil {
			return true
		}
		if s.Rows[j].Values[idx] == nil {
			return false
		}
		return greaterThanNew(s.Rows[j].Values[idx].(env.Object), s.Rows[i].Values[idx].(env.Object))
	}

	sort.Slice(s.Rows, compareCol)
}

func SortByColumnDesc(ps *env.ProgramState, s *env.Table, name string) {
	idx := slices.Index[[]string](s.Cols, name)

	compareCol := func(i, j int) bool {
		if s.Rows[j].Values[idx] == nil {
			return true
		}
		if s.Rows[i].Values[idx] == nil {
			return false
		}
		return greaterThanNew(s.Rows[i].Values[idx].(env.Object), s.Rows[j].Values[idx].(env.Object))
	}

	sort.Slice(s.Rows, compareCol)
}

func WhereEquals(ps *env.ProgramState, s env.Table, name string, val env.Object) env.Object {
	idx := slices.Index(s.Cols, name)
	nspr := env.NewTable(s.Cols)
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

func WhereNotEquals(ps *env.ProgramState, s env.Table, name string, val env.Object) env.Object {
	idx := slices.Index(s.Cols, name)
	nspr := env.NewTable(s.Cols)
	if idx > -1 {
		for _, row := range s.Rows {
			if len(row.Values) > idx {
				if !val.Equal(env.ToRyeValue(row.Values[idx])) {
					nspr.AddRow(row)
				}
			}
		}
		return *nspr
	} else {
		return MakeBuiltinError(ps, "Column not found.", "WhereNotEquals")
	}
}

func WhereMatch(ps *env.ProgramState, s *env.Table, name string, r *regexp.Regexp) env.Object {
	idx := slices.Index(s.Cols, name)
	nspr := env.NewTable(s.Cols)
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

func WhereContains(ps *env.ProgramState, s *env.Table, name string, val string, not bool) env.Object {
	idx := slices.Index(s.Cols, name)
	nspr := env.NewTable(s.Cols)
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

func WhereIn(ps *env.ProgramState, s env.Table, name string, b []env.Object) env.Object {
	idx := slices.Index(s.Cols, name)
	nspr := env.NewTable(s.Cols)
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

func WhereNotIn(ps *env.ProgramState, s env.Table, name string, b []env.Object) env.Object {
	idx := slices.Index(s.Cols, name)
	nspr := env.NewTable(s.Cols)
	if idx > -1 {
		for _, row := range s.Rows {
			if len(row.Values) > idx {
				rv := row.Values[idx]
				if rvObj, ok := rv.(env.Object); ok {
					if !util.ContainsVal(ps, b, rvObj) {
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

func WhereGreater(ps *env.ProgramState, s *env.Table, name string, val env.Object) env.Object {
	idx := slices.Index(s.Cols, name)
	nspr := env.NewTable(s.Cols)
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

func WhereLesser(ps *env.ProgramState, s *env.Table, name string, val env.Object) env.Object {
	idx := slices.Index(s.Cols, name)
	nspr := env.NewTable(s.Cols)
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

func WhereBetween(ps *env.ProgramState, s *env.Table, name string, val1 env.Object, val2 env.Object, inclusiveMode bool) env.Object {
	idx := slices.Index(s.Cols, name)
	nspr := env.NewTable(s.Cols)
	if idx > -1 {
		for _, row := range s.Rows {
			if len(row.Values) > idx {
				rv := row.Values[idx].(env.Object)
				if inclusiveMode {
					if !greaterThanNew(rv, val2) && !lesserThanNew(rv, val1) {
						nspr.AddRow(row)
					}
				} else {
					if greaterThanNew(rv, val1) && lesserThanNew(rv, val2) {
						nspr.AddRow(row)
					}
				}
			}
		}
		return *nspr
	} else {
		return MakeBuiltinError(ps, "Column not found.", "WhereBetween")
	}
}

func AutoType(ps *env.ProgramState, s *env.Table, percent float64) env.Object {
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
			case env.Integer:
				colTypeCount[i]["int"]++
			case env.Decimal:
				colTypeCount[i]["dec"]++
			default:
				continue
			}
		}
	}

	lenRows := len(s.Rows)
	newS := env.NewTable(s.Cols)
	for range s.Rows {
		newRow := make([]any, len(s.Cols))
		newS.AddRow(*env.NewTableRow(newRow, newS))
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
				switch val := row.Values[colNum].(type) {
				case env.String:
					intVal, _ := strconv.Atoi(val.Value)
					newS.Rows[i].Values[colNum] = *env.NewInteger(int64(intVal))
				case env.Integer:
					//intVal, _ := strconv.Atoi(row.Values[colNum].(env.String).Value)
					newS.Rows[i].Values[colNum] = val
				case env.Decimal:
					//intVal, _ := strconv.Atoi(row.Values[colNum].(env.String).Value)
					newS.Rows[i].Values[colNum] = val
				}
			case "dec":
				switch val1 := row.Values[colNum].(type) {
				case env.String:
					floatVal, _ := strconv.ParseFloat(val1.Value, 64)
					newS.Rows[i].Values[colNum] = *env.NewDecimal(floatVal)
				case env.Integer:
					//intVal, _ := strconv.Atoi(row.Values[colNum].(env.String).Value)
					//newS.Rows[i].Values[colNum] = *env.NewInteger(int64(intVal))
					newS.Rows[i].Values[colNum] = *env.NewDecimal(float64(val1.Value))
				case env.Decimal:
					//intVal, _ := strconv.Atoi(row.Values[colNum].(env.String).Value)
					//newS.Rows[i].Values[colNum] = *env.NewInteger(int64(intVal))
					newS.Rows[i].Values[colNum] = val1
				}
			case "str":
				newS.Rows[i].Values[colNum] = row.Values[colNum]
			}
		}
	}

	return *newS
}

func LeftJoin(ps *env.ProgramState, s1 env.Table, s2 env.Table, col1 string, col2 string, innerJoin bool) env.Object {
	if !slices.Contains(s1.Cols, col1) {
		return MakeBuiltinError(ps, "Column not found in first table.", "left-join")
	}
	if !slices.Contains(s2.Cols, col2) {
		return MakeBuiltinError(ps, "Column not found in second table.", "left-join")
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
	nspr := env.NewTable(combinedCols)
	for i, row1 := range s1.GetRows() {
		val1, err := s1.GetRowValue(col1, row1)
		if err != nil {
			return MakeError(ps, fmt.Sprintf("Couldn't retrieve value at row %d (%s)", i, err))
		}

		// the row ids of the second table which match the values in the current first table row
		var s2RowIds []int
		// use index if available
		if ix, ok := s2.Indexes[col2]; ok {
			if rowIds, ok := ix[val1]; ok {
				s2RowIds = rowIds
			}
		} else {
			for j, row2 := range s2.GetRows() {
				val2, err := s2.GetRowValue(col2, row2)
				if err != nil {
					return MakeError(ps, fmt.Sprintf("Couldn't retrieve value at row %d (%s)", j, err))
				}
				val1o, ok := val1.(env.Object)
				if ok {
					val2o, ok1 := val2.(env.Object)
					if ok1 {
						if val1o.Equal(val2o) {
							s2RowIds = append(s2RowIds, j)
							continue
						}
					} else {
						if env.RyeToRaw(val1o, ps.Idx) == val2 {
							s2RowIds = append(s2RowIds, j)
							continue
						}
					}
				}
				val1s, ok := val1.(string)
				if ok {
					if val1s == val2.(string) {
						s2RowIds = append(s2RowIds, j)
						continue
					}
				}
			}
		}
		if innerJoin && len(s2RowIds) == 0 {
			continue
		}
		buildCombinedRow := func(row1Values []any, row2Values []any) []any {
			newRow := make([]any, len(combinedCols))
			copy(newRow, row1Values)

			if row2Values != nil {
				// copy values from second table
				for i, v := range row2Values {
					newRow[i+len(s1.Cols)] = v
				}
			} else {
				// fill with Void{} values when no match
				for i := len(s1.Cols); i < len(combinedCols); i++ {
					newRow[i] = env.Void{}
				}
			}
			return newRow
		}

		if len(s2RowIds) > 0 {
			// add a row for each matching record from s2
			for _, s2RowId := range s2RowIds {
				row2Values := s2.GetRow(ps, s2RowId).Values
				combinedValues := buildCombinedRow(row1.Values, row2Values)
				nspr.AddRow(*env.NewTableRow(combinedValues, nspr))
			}
		} else {
			// add a single row with Void values for s2 columns
			combinedValues := buildCombinedRow(row1.Values, nil)
			nspr.AddRow(*env.NewTableRow(combinedValues, nspr))
		}
	}
	return *nspr
}

func GroupBy(ps *env.ProgramState, s env.Table, cols []string, aggregations map[string][]string) env.Object {
	// Validate that all grouping columns exist
	for _, col := range cols {
		if !slices.Contains(s.Cols, col) {
			return MakeBuiltinError(ps, fmt.Sprintf("Column '%s' not found.", col), "group-by")
		}
	}

	aggregatesByGroup := make(map[string]map[string]float64)
	countByGroup := make(map[string]int)
	for i, row := range s.Rows {
		// Create composite key for multi-column grouping
		groupKeyParts := make([]string, len(cols))
		for j, col := range cols {
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
			case env.Integer:
				groupValStr = strconv.Itoa(int(val.Value))
			case int:
				groupValStr = strconv.Itoa(val)
			default:
				return MakeBuiltinError(ps, "Grouping column value must be a string or number", "group-by")
			}
			groupKeyParts[j] = groupValStr
		}
		// Join with "|" separator to create unique composite key
		groupKey := strings.Join(groupKeyParts, "|")

		if _, ok := aggregatesByGroup[groupKey]; !ok {
			aggregatesByGroup[groupKey] = make(map[string]float64)
		}
		groupAggregates := aggregatesByGroup[groupKey]

		for aggCol, funs := range aggregations {
			for _, fun := range funs {
				colAgg := aggCol + "_" + fun
				if fun == "count" {
					// Count aggregation can be applied on any of the grouping columns
					isGroupingCol := false
					for _, gcol := range cols {
						if aggCol == gcol {
							isGroupingCol = true
							break
						}
					}
					if !isGroupingCol {
						return MakeBuiltinError(ps, "Count aggregation can only be applied on the grouping columns", "group-by")
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
					countByGroup[groupKey]++
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

	// Create result columns: grouping columns + aggregation columns
	newCols := make([]string, len(cols))
	copy(newCols, cols)
	for aggCol, funs := range aggregations {
		for _, fun := range funs {
			newCols = append(newCols, aggCol+"_"+fun)
		}
	}
	newS := env.NewTable(newCols)

	for groupKey, groupAggregates := range aggregatesByGroup {
		newRow := make([]any, len(newCols))
		// Split the composite key back into individual column values
		groupKeyParts := strings.Split(groupKey, "|")
		for i, part := range groupKeyParts {
			newRow[i] = *env.NewString(part)
		}

		// Add aggregation results
		for i, colName := range newCols[len(cols):] {
			if strings.HasSuffix(colName, "_count") {
				newRow[len(cols)+i] = *env.NewInteger(int64(groupAggregates[colName]))
			} else if strings.HasSuffix(colName, "_avg") {
				newRow[len(cols)+i] = *env.NewDecimal(groupAggregates[colName] / float64(countByGroup[groupKey]))
			} else {
				newRow[len(cols)+i] = *env.NewDecimal(groupAggregates[colName])
			}
		}
		newS.AddRow(*env.NewTableRow(newRow, newS))
	}
	return *newS
}
