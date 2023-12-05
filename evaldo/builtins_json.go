//go:build !b_no_json
// +build !b_no_json

package evaldo

import (
	"encoding/json"
	"fmt"
	"rye/env"
	"strconv"
	"strings"
)

func _emptyRM() env.Dict {
	return env.Dict{}
}

func resultToJS(res env.Object) interface{} {
	switch v := res.(type) {
	case env.String:
		return v.Value
	case env.Integer:
		return v.Value
	case env.RyeCtx:
		return "{ 'state': 'todo' }"
	default:
		fmt.Println("No matching type available.")
		// TODO-FIXME - ps - ProgramState is not available, can not add error handling.
	}
	return nil
}

func RyeToJSON(res interface{}) string {
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
		return "\"" + v + "\""
	case env.String:
		return "\"" + v.Value + "\""
	case env.Integer:
		return strconv.Itoa(int(v.Value))
	case env.Decimal:
		return strconv.Itoa(int(v.Value))
	case env.Spreadsheet:
		return SpreadsheetToJSON(v)
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

func JsonToRye(res interface{}) env.Object {
	switch v := res.(type) {
	case float64:
		return *env.NewDecimal(v)
	case int:
		return *env.NewInteger(int64(v))
	case int64:
		return *env.NewInteger(v)
	case string:
		return *env.NewString(v)
	case rune:
		return *env.NewString(string(v))
	case map[string]interface{}:
		return *env.NewDict(v)
	case []interface{}:
		return *env.NewList(v)
	case env.Object:
		return v
	case nil:
		return nil
	default:
		fmt.Println(res)
		// TODO-FIXME
		return env.Void{}
	}
}

// Inspect returns a string representation of the Integer.
func SpreadsheetToJSON(s env.Spreadsheet) string {
	//fmt.Println("IN TO Html")
	var bu strings.Builder
	bu.WriteString("[")
	//fmt.Println(len(s.Rows))
	if s.RawMode {
		for i, row := range s.RawRows {
			if i > 0 {
				bu.WriteString(", ")
			}
			bu.WriteString("{")
			for i, val := range row {
				if i > 0 {
					bu.WriteString(", ")
				}
				bu.WriteString("\"")
				bu.WriteString(s.Cols[i])
				bu.WriteString("\": ")
				bu.WriteString(RyeToJSON(val))
			}
			bu.WriteString("} ")
		}
	} else {
		for i, row := range s.Rows {
			if i > 0 {
				bu.WriteString(", ")
			}
			bu.WriteString("{")
			for i, val := range row.Values {
				if i > 0 {
					bu.WriteString(", ")
				}
				bu.WriteString("\"")
				bu.WriteString(s.Cols[i])
				bu.WriteString("\": ")
				bu.WriteString(RyeToJSON(val))
			}
			bu.WriteString("} ")
		}
	}
	bu.WriteString("]")
	//fmt.Println(bu.String())
	return bu.String()
}

// { <person> [ .print ] }
// { <person> { _ [ .print ] <name> <surname> <age> { _ [ .print2 ";" ] } }

func _parse_json(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
	switch input := arg0.(type) {
	case env.String:
		var m interface{}
		err := json.Unmarshal([]byte(input.Value), &m)
		if err != nil {
			return MakeBuiltinError(ps, "Failed to Unmarshal.", "_parse_json")
			//panic(err)
		}
		return JsonToRye(m)
	default:
		return MakeArgError(ps, 1, []env.Type{env.StringType}, "_parse_json")
	}
}

var Builtins_json = map[string]*env.Builtin{

	"parse-json": {
		Argsn: 1,
		Doc:   "Parsing JSON values.",
		Fn: func(es *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return _parse_json(es, arg0, arg1, arg2, arg3, arg4)
		},
	},
	"to-json": {
		Argsn: 1,
		Doc:   "Takes a Rye value and returns it encoded into JSON.",
		Fn: func(es *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewString(RyeToJSON(arg0))
		},
	},
}
