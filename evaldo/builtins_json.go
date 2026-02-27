//go:build !no_json
// +build !no_json

package evaldo

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/refaktor/rye/env"
)

func _emptyJSONDict() env.Dict {
	return env.Dict{}
}

func resultToJS(res env.Object) any {
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
	case *env.RyeCtx:
		return "{ 'state': 'todo' }"
	default:
		fmt.Println("No matching type available.")
		// TODO-FIXME - ps - ProgramState is not available, can not add error handling.
	}
	return nil
}

func RyeToJSON(res any) string {
	return RyeToJSONWithIdxs(res, nil)
}

func RyeToJSONWithIdxs(res any, idxs *env.Idxs) string {
	// fmt.Printf("Type: %T", res)
	switch v := res.(type) {
	case nil:
		return "null"
	case int:
		return strconv.Itoa(v)
	case int32:
		return strconv.Itoa(int(v))
	case int64:
		return strconv.Itoa(int(v))
	case float64:
		return strconv.Itoa(int(v))
	case string:
		if v[0] == '[' && v[len(v)-1:] == "]" {
			return v
		}
		return "\"" + EscapeJson(v) + "\""
	case env.String:
		return "\"" + EscapeJson(v.Value) + "\""
	case env.Integer:
		return strconv.Itoa(int(v.Value))
	case env.Decimal:
		return strconv.FormatFloat(v.Value, 'f', -1, 64)
	case env.Block:
		return BlockToJSONWithIdxs(v, idxs)
	case env.List:
		return ListToJSONWithIdxs(v, idxs)
	case []any:
		return ListToJSONWithIdxs(*env.NewList(v), idxs)
	case env.Vector:
		return VectorToJSONWithIdxs(v, idxs)
	case env.Dict:
		return DictToJSONWithIdxs(v, idxs)
	case map[string]any:
		return DictToJSONWithIdxs(*env.NewDict(v), idxs)
	case *env.Table:
		return TableToJSONWithIdxs(*v, idxs)
	case env.Table:
		return TableToJSONWithIdxs(v, idxs)
	case env.TableRow:
		return TableRowToJSONWithIdxs(v, idxs)
	case *env.Error:
		status := ""
		if v.Status != 0 {
			status = "\"status\": " + RyeToJSONWithIdxs(v.Status, idxs)
		}
		var b strings.Builder
		b.WriteString("{ " + status + ", \"message\": " + RyeToJSONWithIdxs(v.Message, idxs))
		if v.Parent != nil {
			b.WriteString(", \"parent\": " + RyeToJSONWithIdxs(v.Parent, idxs))
		}
		b.WriteString(", \"data\": { ")
		for k, v := range v.Values {
			switch ob := v.(type) {
			case env.Object:
				b.WriteString(" " + RyeToJSONWithIdxs(k, idxs) + ": " + RyeToJSONWithIdxs(ob, idxs) + ", ")
			}
		}
		b.WriteString("} }")
		return b.String()
	case *env.RyeCtx:
		if idxs != nil {
			return ContextToJSON(v, idxs)
		}
		return "{ \"error\": \"context requires idxs for JSON conversion\" }"
	default:
		return fmt.Sprintf("\"type %T not handeled\"", v)
		// TODO-FIXME
	}
}

func RyeToJSONLines(res any) string {
	// fmt.Printf("Type: %T", res)
	switch v := res.(type) {
	case env.Table:
		return TableToJSONLines(v)
	case *env.Error:
		status := ""
		if v.Status != 0 {
			status = "\"status\": " + RyeToJSON(v.Status)
		}
		var b strings.Builder
		b.WriteString("{ " + status + ", \"message\": " + RyeToJSON(v.Message))
		if v.Parent != nil {
			b.WriteString(", \"parent\": " + RyeToJSON(v.Parent))
		}
		b.WriteString(", \"data\": { ")
		for k, v := range v.Values {
			switch ob := v.(type) {
			case env.Object:
				b.WriteString(" " + RyeToJSON(k) + ": " + RyeToJSON(ob) + ", ")
			}
		}
		b.WriteString("} }")
		return b.String()
	case *env.RyeCtx:
		return "{ 'state': 'todo' }"
	default:
		return "\"not handeled\""
		// TODO-FIXME
	}
}

func EscapeJson(val string) string {
	res := strings.ReplaceAll(val, "\"", "\\\"")
	return res
}

func VectorToJSON(vector env.Vector) string {
	return VectorToJSONWithIdxs(vector, nil)
}

func VectorToJSONWithIdxs(vector env.Vector, idxs *env.Idxs) string {
	var bu strings.Builder
	bu.WriteString("[")
	for i, val := range vector.Value {
		if i > 0 {
			bu.WriteString(", ")
		}
		bu.WriteString(RyeToJSONWithIdxs(val, idxs))
	}
	bu.WriteString("]")
	return bu.String()
}

// BlockToJSON converts a Block to JSON array format.
func BlockToJSON(block env.Block) string {
	return BlockToJSONWithIdxs(block, nil)
}

// BlockToJSONWithIdxs converts a Block to JSON array format with Idxs for context support.
func BlockToJSONWithIdxs(block env.Block, idxs *env.Idxs) string {
	var bu strings.Builder
	bu.WriteString("[")
	for i, val := range block.Series.S {
		if i > 0 {
			bu.WriteString(", ")
		}
		bu.WriteString(RyeToJSONWithIdxs(val, idxs))
	}
	bu.WriteString("] ")
	return bu.String()
}

// ListToJSON converts a List to JSON array format.
func ListToJSON(list env.List) string {
	return ListToJSONWithIdxs(list, nil)
}

// ListToJSONWithIdxs converts a List to JSON array format with Idxs for context support.
func ListToJSONWithIdxs(list env.List, idxs *env.Idxs) string {
	var bu strings.Builder
	bu.WriteString("[")
	for i, val := range list.Data {
		if i > 0 {
			bu.WriteString(", ")
		}
		bu.WriteString(RyeToJSONWithIdxs(val, idxs))
	}
	bu.WriteString("] ")
	return bu.String()
}

// DictToJSON converts a Dict to JSON object format.
func DictToJSON(dict env.Dict) string {
	return DictToJSONWithIdxs(dict, nil)
}

// DictToJSONWithIdxs converts a Dict to JSON object format with Idxs for context support.
func DictToJSONWithIdxs(dict env.Dict, idxs *env.Idxs) string {
	var bu strings.Builder
	bu.WriteString("{")
	i := 0
	for key, val := range dict.Data {
		if i > 0 {
			bu.WriteString(", ")
		}
		bu.WriteString(RyeToJSONWithIdxs(key, idxs))
		bu.WriteString(": ")
		bu.WriteString(RyeToJSONWithIdxs(val, idxs))
		i = i + 1
	}
	bu.WriteString("} ")
	return bu.String()
}

// ContextToJSON converts a RyeCtx to JSON object format.
func ContextToJSON(ctx *env.RyeCtx, idxs *env.Idxs) string {
	var bu strings.Builder
	bu.WriteString("{")
	i := 0
	for key, val := range ctx.GetState() {
		if i > 0 {
			bu.WriteString(", ")
		}
		// Convert word index to string key
		bu.WriteString("\"")
		bu.WriteString(EscapeJson(idxs.GetWord(key)))
		bu.WriteString("\": ")
		bu.WriteString(RyeToJSONWithIdxs(val, idxs))
		i = i + 1
	}
	bu.WriteString("} ")
	return bu.String()
}

// TableRowToJSON converts a TableRow to JSON object format.
func TableRowToJSON(row env.TableRow) string {
	return TableRowToJSONWithIdxs(row, nil)
}

// TableRowToJSONWithIdxs converts a TableRow to JSON object format with Idxs for context support.
func TableRowToJSONWithIdxs(row env.TableRow, idxs *env.Idxs) string {
	var bu strings.Builder
	bu.WriteString("{")
	for i, val := range row.Values {
		if i > 0 {
			bu.WriteString(", ")
		}
		bu.WriteString("\"")
		bu.WriteString(row.Uplink.GetColumnNames()[i])
		bu.WriteString("\": ")
		bu.WriteString(RyeToJSONWithIdxs(val, idxs))
	}
	bu.WriteString("} ")
	return bu.String()
}

// TableToJSON converts a Table to JSON array format.
func TableToJSON(s env.Table) string {
	return TableToJSONWithIdxs(s, nil)
}

// TableToJSONWithIdxs converts a Table to JSON array format with Idxs for context support.
func TableToJSONWithIdxs(s env.Table, idxs *env.Idxs) string {
	var bu strings.Builder
	bu.WriteString("[")
	for i, row := range s.Rows {
		if i > 0 {
			bu.WriteString(", ")
		}
		bu.WriteString(TableRowToJSONWithIdxs(row, idxs))
	}
	bu.WriteString("]")
	return bu.String()
}

func TableToJSONLines(s env.Table) string {
	var bu strings.Builder
	for _, row := range s.Rows {
		bu.WriteString(TableRowToJSON(row))
		bu.WriteString("\n")
	}
	return bu.String()
}

// { <person> [ .print ] }
// { <person> { _ [ .print ] <name> <surname> <age> { _ [ .print2 ";" ] } }

var Builtins_json = map[string]*env.Builtin{

	//
	// ##### JSON #####  "Parsing and generating JSON"
	//
	// Tests:
	// equal { "[ 1, 2, 3 ]" |parse-json |length? } 3
	// equal { "[ 1, 2, 3 ]" |parse-json |type? } 'list
	// Args:
	// * json: string containing JSON data
	// Returns:
	// * parsed Rye value (list, dict, string, integer, etc.)
	"parse-json": {
		Argsn: 1,
		Doc:   "Parses JSON string into Rye values.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch input := arg0.(type) {
			case env.String:
				var m any
				err := json.Unmarshal([]byte(input.Value), &m)
				if err != nil {
					return MakeBuiltinError(ps, "Failed to Unmarshal.", "_parse_json")
					//panic(err)
				}
				return env.ToRyeValue(m)
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "_parse_json")
			}
		},
	},

	// Tests:
	// equal { `{"a": 2, "b": "x"} \n{"a": 3, "b": "y"} \n` |parse-json\lines |to-table } table { "a" "b" } { 2 "x" 3 "y" }
	// Args:
	// * json: string containing consecutive JSON values
	// Returns:
	// * list of parsed Rye values
	"parse-json\\lines": {
		Argsn: 1,
		Doc:   "Parses JSON string into Rye values.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch input := arg0.(type) {
			case env.String:
				var vals []any
				decoder := json.NewDecoder(strings.NewReader(input.Value))
				for decoder.More() {
					var m any
					err := decoder.Decode(&m)
					if err != nil {
						return MakeBuiltinError(ps, "Failed to Unmarshal.", "_parse_json\\lines")
					}
					vals = append(vals, env.ToRyeValue(m))
				}
				return *env.NewList(vals)
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "_parse_json\\lines")
			}
		},
	},

	// Tests:
	// equal { list { 1 2 3 } |to-json } "[1, 2, 3] "
	// equal { dict { a: 1 b: 2 c: 3 } |to-json } `{"a": 1, "b": 2, "c": 3} `
	// equal { { 1 2 3 } |to-json } "[1, 2, 3] "
	// equal { context { a: 1 b: 2 } |to-json |parse-json -> "a" } 1
	// Args:
	// * value: any Rye value to encode (block, list, dict, context, string, integer, etc.)
	// Returns:
	// * string containing the JSON representation
	"to-json": {
		Argsn: 1,
		Doc:   "Converts a Rye value to a JSON string. Supports block (like list) and context (like dict).",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewString(RyeToJSONWithIdxs(arg0, ps.Idx))
		},
	},
	// Tests:
	// equal { table { "a" "b" } { 2 "x" 3 "y" } |to-json\lines } `{"a": 2, "b": "x"} \n{"a": 3, "b": "y"} \n`
	// Args:
	// * table: table value to encode
	// Returns:
	// * string containing the JSON representation with each row on a new line
	"to-json\\lines": {
		Argsn: 1,
		Doc:   "Converts a table to JSON with each row on a separate line.",
		Fn: func(es *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewString(RyeToJSONLines(arg0))
		},
	},
}
