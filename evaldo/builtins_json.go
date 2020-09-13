package evaldo

import (
	"encoding/json"
	"fmt"
	"rye/env"
)

func _emptyRM() env.RawMap {
	return env.RawMap{}
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

func JSToRye(res interface{}) env.Object {
	switch v := res.(type) {
	case int64:
		return env.Integer{v}
	case string:
		return env.String{v}
	case map[string]interface{}:
		return *env.NewRawMap(v)
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
		fmt.Println(m)
		return JSToRye(m)
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
}
