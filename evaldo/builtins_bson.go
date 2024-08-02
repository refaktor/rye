//go:build !no_bson
// +build !no_bson

package evaldo

import (
	"fmt"

	"github.com/refaktor/rye/env"

	"github.com/drewlanenga/govector"
	"go.mongodb.org/mongo-driver/bson"
)

func ValueToBSON(arg0 env.Object, topLevel bool) any {
	//fmt.Println("val->bson")
	//fmt.Println(topLevel)
	var val any
	var typ string
	var met any
	switch obj := arg0.(type) {
	case env.Integer:
		val = obj.Value
	case env.Decimal:
		val = float64(obj.Value)
	case env.String:
		val = obj.Value
	case env.Vector:
		val = obj.Value
		typ = "vec"
	case env.Block:
		vals := make([]any, obj.Series.Len())
		for i, valu := range obj.Series.S {
			vals[i] = ValueToBSON(valu, false)
		}
		val = vals
		typ = "block"
	/*case env.SpreadsheetRow:
	vals := make([]interface{}, len(obj.Values))
	for i, valu := range obj.Values {
		switch val := valu.(type) {
		case env.Object:
			vals[i] = ValueToBSON(val, false)
		}
	}
	val = vals
	typ = "sprr"*/
	case env.SpreadsheetRow:
		vals := make([]any, len(obj.Values))
		for i, valu := range obj.Values {
			switch val := valu.(type) {
			case env.Object:
				vals[i] = ValueToBSON(val, false)
			}
		}
		val = vals
	case env.Spreadsheet:
		// spr["val"] = obj.Rows
		//data := make([]interface{}, len(obj.Rows))
		//cols := make([]string, len(obj.Cols))
		//for i, valu := range obj.Series.S {
		//	vals[i] = ValueToBSON(valu, false)
		//}
		// fmt.Println(spr)
		rows := make([]any, len(obj.Rows))
		for i, valu := range obj.Rows {
			rows[i] = ValueToBSON(valu, false)
		}
		val = rows
		typ = "spr"
		met = obj.Cols
	default:
		fmt.Println("No matching arguments found.")
		// TODO-FIXME
	}
	if topLevel || typ != "" {
		//fmt.Println(bson.M{"val": val, "typ": typ, "met": met})
		return bson.M{"val": val, "typ": typ, "met": met}
	} else {
		return val
	}
}

func BsonToValue_Map(ps *env.ProgramState, val any, typ string, meta any, topLevel bool) env.Object {

	/*fmt.Println("BSONToVALUE_MAP")
	fmt.Println(val)
	fmt.Println(typ)
	fmt.Println(meta)
	fmt.Printf("Type: %T\n", val)*/

	switch rval := val.(type) {
	case int64:
		return env.Integer{int64(rval)}
	case float32:
		return env.Decimal{float64(rval)}
	case float64:
		return env.Decimal{float64(rval)}
	case string:
		return env.String{rval}
	case bson.M:
		return BsonToValue_Map(ps, rval["val"], rval["typ"].(string), rval["met"], false)
	case map[string]any:
		return BsonToValue_Map(ps, rval["val"], rval["typ"].(string), rval["met"], false)
	case bson.A:
		switch typ {
		case "spr":
			//fmt.Println("SPR:")
			//fmt.Printf("Type: %T\n", meta)
			switch cols := meta.(type) {
			case bson.A:
				rcols := make([]string, len(cols))
				for i, rr := range cols {
					rcols[i] = string(rr.(string))
				}
				spr := env.NewSpreadsheet(rcols)
				//rows := make([]interface{}, len(spr.Cols))
				for ii := 0; ii < len(rval); ii++ {

					//fmt.Printf("Type: %T\n", rval[ii])
					//fmt.Println(rval[ii])

					switch rrval := rval[ii].(type) {
					case bson.A:
						cells := make([]any, len(rrval))
						for iii, rrrval := range rrval {
							cells[iii] = BsonToValue_Map(ps, rrrval, "", nil, false)
						}
						spr.AddRow(env.SpreadsheetRow{cells, spr})
					case []any:
						spr.AddRow(env.SpreadsheetRow{rrval, spr})
					}
				}
				return *spr
			}
		case "vec":
			rrval := make([]float32, len(rval))
			for i, element := range rval {
				value, ok := element.(float64)
				if !ok {
					panic("Element is not a float64")
				}
				rrval[i] = float32(value)
			}
			ret, err := govector.AsVector(rrval)
			if err != nil {
				return makeError(ps, err.Error())
			}
			return *env.NewVector(ret)
		case "block":
			rrval := make([]env.Object, len(rval))
			for i, element := range rval {
				value := BsonToValue_Map(ps, element, "", nil, false)
				rrval[i] = value
			}
			return *env.NewBlock(*env.NewTSeries(rrval))
		default:
			return MakeBuiltinError(ps, "No matching arguments found.", "BsonToValue_Map")
		}
	}
	return MakeBuiltinError(ps, "bson type not found.", "BsonToValue_Map")
}

func BsonToValue_Val(ps *env.ProgramState, val any, topLevel bool) env.Object {

	//fmt.Printf("Type: %T\n", val)
	//fmt.Println(val)

	switch rval := val.(type) {
	case bson.M:
		/*fmt.Println("~~")
		fmt.Println(rval["val"])
		fmt.Printf("Type: %T\n", rval["val"])
		fmt.Printf("Meta: %T\n", rval["met"])
		fmt.Println("~~")*/

		val := BsonToValue_Map(ps, rval["col"], rval["typ"].(string), rval["met"], false)
		return val
	default:
		return MakeBuiltinError(ps, "bson type not found.", "BsonToValue_Val")
	}
}

var Builtins_bson = map[string]*env.Builtin{

	"from-bson": {
		Argsn: 1,
		Doc:   "Takes a BSON value and returns it encoded into Rye values.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			var val map[string]any
			// var val interface{}
			// err := bson.Unmarshal(arg0.(env.Native).Value.([]byte), &val)
			err := bson.Unmarshal(arg0.(env.Native).Value.([]byte), &val)
			if err != nil {
				return MakeBuiltinError(ps, err.Error(), "from-bson")
			}

			return BsonToValue_Map(ps, val["val"], val["typ"].(string), val["met"], true)
			//return BsonToValue_Val(ps, val, true)
			// return makeError(ps, "bson type not found")
		},
	},
	"to-bson": {
		Argsn: 1,
		Doc:   "Takes a Rye value and returns it encoded into BSON.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			value := ValueToBSON(arg0, true)
			encoded, err := bson.Marshal(value)
			if err != nil {
				return MakeBuiltinError(ps, err.Error(), "to-bson")
			}
			return *env.NewNative(ps.Idx, encoded, "bytes")
		},
	},
}
