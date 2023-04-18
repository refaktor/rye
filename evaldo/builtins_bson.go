//go:build !b_bsonaa
// +build !b_bsonaa

package evaldo

import (
	"fmt"
	"rye/env"

	"github.com/drewlanenga/govector"
	"go.mongodb.org/mongo-driver/bson"
)

var Builtins_bson = map[string]*env.Builtin{

	"from-bson": {
		Argsn: 1,
		Fn: func(es *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			var val map[string]interface{}
			err := bson.Unmarshal(arg0.(env.Native).Value.([]byte), &val)
			if err != nil {
				return makeError(es, err.Error())
			}
			fmt.Println(val["val"])
			fmt.Printf("Type: %T", val["val"])
			switch rval := val["val"].(type) {
			case int64:
				return env.Integer{int64(rval)}
			case string:
				return env.String{rval}
			case bson.A:
				switch val["typ"] {
				case "vec":
					rrval := make([]float32, len(rval))
					for i, element := range rval {
						value, ok := element.(float64)
						if !ok {
							panic("Element is not a float64")
						}
						rrval[i] = float32(value)
					}
					// 	rrval := rval.([]float32)
					ret, err := govector.AsVector(rrval)
					if err != nil {
						return makeError(es, err.Error())
					}
					return *env.NewVector(ret)
				}
			case env.Integer:
				return rval
			case env.String:
				return rval
			}
			return makeError(es, "bson type not found")
		},
	},
	"to-bson": {
		Argsn: 1,
		Doc:   "Takes a Rye value and returns it encoded into BSON.",
		Fn: func(es *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			var val interface{}
			var typ string
			switch obj := arg0.(type) {
			case env.Integer:
				val = obj.Value
			case env.String:
				val = obj.Value
			case env.Vector:
				val = obj.Value
				typ = "vec"
			case env.Block:
				val = obj.Series.S
				// for val := range obj.Series.S
				typ = "ser"
			}
			encoded, err := bson.Marshal(bson.M{"val": val, "typ": typ})
			if err != nil {
				return makeError(es, err.Error())
			}
			return *env.NewNative(es.Idx, encoded, "bytes")
		},
	},
}
