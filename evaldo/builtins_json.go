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

func _emptyRM() env.Dict {
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
	case env.RyeCtx:
		return "{ 'state': 'todo' }"
	default:
		fmt.Println("No matching type available.")
		// TODO-FIXME - ps - ProgramState is not available, can not add error handling.
	}
	return nil
}

func RyeToJSON(res any) string {
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
	case env.List:
		return ListToJSON(v)
	case env.Vector:
		return VectorToJSON(v)
	case env.Dict:
		return DictToJSON(v)
	case *env.Spreadsheet:
		return SpreadsheetToJSON(*v)
	case env.Spreadsheet:
		return SpreadsheetToJSON(v)
	case env.SpreadsheetRow:
		return SpreadsheetRowToJSON(v)
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
	case env.RyeCtx:
		return "{ 'state': 'todo' }"
	default:
		return fmt.Sprintf("\"type %T not handeled\"", v)
		// TODO-FIXME
	}
}

func RyeToJSONLines(res any) string {
	// fmt.Printf("Type: %T", res)
	switch v := res.(type) {
	case env.Spreadsheet:
		return SpreadsheetToJSONLines(v)
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
	case env.RyeCtx:
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
	var bu strings.Builder
	bu.WriteString("[")
	for i, val := range vector.Value {
		if i > 0 {
			bu.WriteString(", ")
		}
		bu.WriteString(RyeToJSON(val))
	}
	bu.WriteString("]")
	return bu.String()
}

// Inspect returns a string representation of the Integer.
func ListToJSON(list env.List) string {
	var bu strings.Builder
	bu.WriteString("[")
	for i, val := range list.Data {
		if i > 0 {
			bu.WriteString(", ")
		}
		bu.WriteString(RyeToJSON(val))
	}
	bu.WriteString("] ")
	return bu.String()
}

// Inspect returns a string representation of the Integer.
func DictToJSON(dict env.Dict) string {
	var bu strings.Builder
	bu.WriteString("{")
	i := 0
	for key, val := range dict.Data {
		if i > 0 {
			bu.WriteString(", ")
		}
		bu.WriteString(RyeToJSON(key))
		bu.WriteString(": ")
		bu.WriteString(RyeToJSON(val))
		i = i + 1
	}
	bu.WriteString("} ")
	return bu.String()
}

// Inspect returns a string representation of the Integer.
func SpreadsheetRowToJSON(row env.SpreadsheetRow) string {
	var bu strings.Builder
	bu.WriteString("{")
	for i, val := range row.Values {
		if i > 0 {
			bu.WriteString(", ")
		}
		bu.WriteString("\"")
		bu.WriteString(row.Uplink.Cols[i])
		bu.WriteString("\": ")
		bu.WriteString(RyeToJSON(val))
	}
	bu.WriteString("} ")
	return bu.String()
}

// Inspect returns a string representation of the Integer.
func SpreadsheetToJSON(s env.Spreadsheet) string {
	//fmt.Println("IN TO Html")
	var bu strings.Builder
	bu.WriteString("[")
	//fmt.Println(len(s.Rows))
	for i, row := range s.Rows {
		if i > 0 {
			bu.WriteString(", ")
		}
		bu.WriteString(SpreadsheetRowToJSON(row))
	}
	bu.WriteString("]")
	//fmt.Println(bu.String())
	return bu.String()
}

func SpreadsheetToJSONLines(s env.Spreadsheet) string {
	//fmt.Println("IN TO Html")
	var bu strings.Builder
	//fmt.Println(len(s.Rows))
	for _, row := range s.Rows {
		bu.WriteString(SpreadsheetRowToJSON(row))
		bu.WriteString("\n")
	}
	//fmt.Println(bu.String())
	return bu.String()
}

// { <person> [ .print ] }
// { <person> { _ [ .print ] <name> <surname> <age> { _ [ .print2 ";" ] } }

var Builtins_json = map[string]*env.Builtin{

	// Tests:
	// equals { "[ 1, 2, 3 ]" |parse-json |length? } 3
	"parse-json": {
		Argsn: 1,
		Doc:   "Parses JSON to and turns them to Rye values.",
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
	"to-json": {
		Argsn: 1,
		Doc:   "Takes a Rye value and returns it encoded into JSON.",
		Fn: func(es *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewString(RyeToJSON(arg0))
		},
	},
	"to-json\\lines": {
		Argsn: 1,
		Doc:   "Takes a Rye value and returns it encoded into JSON.",
		Fn: func(es *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewString(RyeToJSONLines(arg0))
		},
	},
}
