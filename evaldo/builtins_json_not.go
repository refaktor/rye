//go:build no_json
// +build no_json

package evaldo

import (
	"fmt"
	"strconv"

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
	case env.RyeCtx:
		return "{ 'state': 'todo' }"
	}
	return nil
}

func RyeToJSON(res any) string {
	switch v := res.(type) {
	case nil:
		return "null"
	case int:
		return strconv.Itoa(v)
	case int32:
		return strconv.Itoa(int(v))
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
	fmt.Println(res)
	return "not handeled1"
}

var Builtins_json = map[string]*env.Builtin{}
