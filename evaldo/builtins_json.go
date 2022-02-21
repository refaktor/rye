// +build !b_no_json

package evaldo

import (
	"encoding/json"
	"math"
	"rye/env"
	"strconv"
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
	}
	return env.Void{}
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

