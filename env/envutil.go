package env

import "fmt"

func ToRyeValue(val any) Object {
	switch v := val.(type) {
	case float64:
		return *NewDecimal(v)
	case int:
		return *NewInteger(int64(v))
	case int64:
		return *NewInteger(v)
	case string:
		return *NewString(v)
	case rune:
		return *NewString(string(v))
	case map[string]any:
		return *NewDict(v)
	case []any:
		return *NewList(v)
	case *List:
		return *v
	case *Block:
		return *v
	case *Object:
		return *v
	case Object:
		return v
	case nil:
		return nil
	default:
		fmt.Println(val)
		// TODO-FIXME
		return Void{}
	}
}
