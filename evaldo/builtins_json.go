// +build !b_no_json

package evaldo

import (
	"encoding/json"
	"fmt"
	"math"
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
	}
	return nil
}

func RyeToJSON(res interface{}) string {
	switch v := res.(type) {
	case nil:
		return "null"
	case int:
		return strconv.Itoa(v)
	case int64:
		return strconv.Itoa(int(v))
	case string:
		return "\"" + v + "\""
	case env.String:
		return "\"" + v.Value + "\""
	case env.Integer:
		return strconv.Itoa(int(v.Value))
	case env.Spreadsheet:
		return SpreadsheetToJSON(v)
	case *env.Error:
		if v != nil {
			return "{ \"code\": " + RyeToJSON(v.Status) + ", \"message\": " + RyeToJSON(v.Message) + ", \"parent\": " + RyeToJSON(v.Parent) + " }"
		} else {
			return "null"
		}
	case env.RyeCtx:
		return "{ 'state': 'todo' }"
	}
	return "\"not handeled\""
}

func JsonToRye(res interface{}) env.Object {
	switch v := res.(type) {
	case float64:
		return env.Integer{int64(math.Round(v))}
	case int:
		return env.Integer{int64(v)}
	case int64:
		return env.Integer{v}
	case string:
		return env.String{v}
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
		return env.Void{}
	}
}

// Inspect returns a string representation of the Integer.
func SpreadsheetToJSON(s env.Spreadsheet) string {
	//fmt.Println("IN TO Html")
	var bu strings.Builder
	bu.WriteString("[")
	fmt.Println(len(s.Rows))
	for _, row := range s.Rows {
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
		bu.WriteString("}")
	}
	bu.WriteString("]")
	//fmt.Println(bu.String())
	return bu.String()
}

// { <person> [ .print ] }
// { <person> { _ [ .print ] <name> <surname> <age> { _ [ .print2 ";" ] } }

func _parse_json(es *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
	switch input := arg0.(type) {
	case env.String:
		var m interface{}
		err := json.Unmarshal([]byte(input.Value), &m)
		if err != nil {
			panic(err)
		}
		return JsonToRye(m)
	}
	return env.Void{}
}

var Builtins_json = map[string]*env.Builtin{

	"parse-json": {
		Argsn: 1,
		Fn: func(es *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return _parse_json(es, arg0, arg1, arg2, arg3, arg4)
		},
	},
	"to-json": {
		Argsn: 1,
		Doc:   "Takes a Rye value and returns it encoded into JSON.",
		Fn: func(es *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return env.String{RyeToJSON(arg0)}
		},
	},
}
