//go:build !no_sql
// +build !no_sql

package batteries

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/refaktor/rye/env"
	"github.com/refaktor/rye/evaldo"
)

// queryOneValue reuses the corresponding Query implementation so SQL-DSL,
// parameter binding and value conversions remain identical. Query returns a
// table; Query\one returns its first row's first column instead.
func queryOneValue(ps *env.ProgramState, db, statement env.Object, query *env.Builtin, name string) env.Object {
	result := query.Fn(ps, db, statement, nil, nil, nil)
	if ps.ErrorFlag || ps.FailureFlag {
		return result
	}
	// Some existing Query implementations return an Error without setting a
	// flag (for example, a SQLite scan error). Preserve that diagnostic.
	if _, isError := result.(*env.Error); isError {
		ps.FailureFlag = true
		return result
	}
	table, ok := result.(env.Table)
	if !ok {
		return evaldo.MakeBuiltinError(ps, "Query did not return a table", name)
	}
	if len(table.Rows) == 0 || len(table.Rows[0].Values) == 0 {
		return evaldo.MakeBuiltinError(ps, "Query returned no value", name)
	}
	if table.Rows[0].Values[0] == nil {
		return env.Void{} // SQL NULL is a value, not a missing row.
	}
	value, ok := table.Rows[0].Values[0].(env.Object)
	if !ok {
		return evaldo.MakeBuiltinError(ps, "Query returned an unsupported value", name)
	}
	return value
}

// SQL mode constants for different database parameter styles
const MODE_SQLITE = 1 // Uses ? placeholders
const MODE_PSQL = 2   // Uses $1, $2, etc. placeholders

// SQL_EvalBlock evaluates a block of SQL expressions and builds a SQL string
func SQL_EvalBlock(es *env.ProgramState, mode int, values []any) (*env.ProgramState, []any) {
	var bu strings.Builder
	var str string
	for es.Ser.Pos() < es.Ser.Len() {
		es, str, values = SQL_EvalExpression(es, values, mode)
		bu.WriteString(str + " ")
	}
	es.Res = *env.NewString(bu.String())
	return es, values
}

// SQL_EvalExpression evaluates a single SQL expression and converts Rye objects to SQL syntax
// mode 1: SQLite (uses ? placeholders), mode 2: PostgreSQL (uses $1, $2, etc. placeholders)
func SQL_EvalExpression(es *env.ProgramState, vals []any, mode int) (*env.ProgramState, string, []any) {
	object := es.Ser.Pop()

	switch obj := object.(type) {
	case env.Integer:
		return es, strconv.FormatInt(obj.Value, 10), vals
	case env.Decimal:
		return es, strconv.FormatFloat(obj.Value, 'f', 6, 64), vals
	case env.String:
		return es, "'" + obj.Value + "'", vals
	case env.Word:
		return es, es.Idx.GetWord(obj.Index), vals
	case env.Opword:
		return es, es.Idx.GetWord(obj.Index)[1:], vals
	case env.Pipeword:
		return es, es.Idx.GetWord(obj.Index)[1:], vals
	case env.Block:
		ser := es.Ser
		es.Ser = obj.Series
		es1, vals1 := SQL_EvalBlock(es, mode, vals)
		es.Ser = ser
		if obj.Mode == 2 {
			return es1, "( " + es.Res.(env.String).Value + " )", vals1
		} else if obj.Mode == 1 {
			return es1, "[ " + es.Res.(env.String).Value + " ]", vals1
		} else {
			return es1, "{ " + es.Res.(env.String).Value + " }", vals1
		}
	case env.Getword:
		val, _ := es.Ctx.Get(obj.Index)
		vals = append(vals, sqlResultToJS(val))
		var ph string
		switch mode {
		case MODE_SQLITE:
			ph = "?"
		case MODE_PSQL:
			ph = "$" + strconv.Itoa(len(vals))
		}
		return es, ph, vals
	case env.Comma:
		return es, ", ", vals
	default:
		fmt.Println("OTHER SQL NODE")
		return es, "Error 123112431", vals
	}
}

// sqlResultToJS converts Rye objects to appropriate Go types for SQL parameters
// This is a simplified version of resultToJS that doesn't depend on the JSON module
func sqlResultToJS(res env.Object) any {
	switch v := res.(type) {
	case env.String:
		return v.Value
	case env.Integer:
		return v.Value
	case *env.Integer:
		return v.Value
	case env.Decimal:
		return v.Value
	case *env.Decimal:
		return v.Value
	case env.Boolean:
		return v.Value
	case *env.Boolean:
		return v.Value
	case env.Void:
		return nil
	default:
		// For unsupported types, convert to string representation
		return fmt.Sprintf("%v", res)
	}
}
