package evaldo

import (
	"rye/env"
	//	"rye/util"
)

// Integer represents an integer.
type ConversionError struct {
	message string
}

func Conversion_EvalBlock(es *env.ProgramState, vals env.RyeCtx) (env.RyeCtx, map[string]env.Object) {

	notes := make(map[string]env.Object, 0) // TODO ... what is this 2 here ... just for temp

	var key int

	out := env.NewEnv(nil)

	for es.Ser.Pos() < es.Ser.Len() {
		object := es.Ser.Pop()
		/////		var verr *ValidationError
		switch obj := object.(type) {
		case env.Setword:
			key = obj.Index
		case env.LSetword:
			val, _ := vals.Get(obj.Index)
			out.Set(key, val)
		}
	}
	//set the last value too
	/////7///	vals.Data[name] = val
	return *out, notes
}

func newCE(n string) *ConversionError {
	return &ConversionError{n}
}

func conversion_evalWord(word env.Word, es *env.ProgramState, val interface{}) (interface{}, env.Object) {
	// later get all word indexes in adwance and store them only once... then use integer comparisson in switch below
	// this is two times BAD ... first it needs to retrieve a string of index (BIG BAD) and then it compares string to string
	// instead of just comparing two integers
	switch es.Idx.GetWord(word.Index) {
	case "optional":
		def := es.Ser.Pop()
		if val == nil {
			return def, nil
		} else {
			return val, nil
		}
	case "check":
		serr := es.Ser.Pop()
		switch blk := es.Ser.Pop().(type) {
		case env.Block:
			ser := es.Ser
			es.Ser = blk.Series
			EvalBlockInj(es, val.(env.Object), true)
			es.Ser = ser
			if es.Res.(env.Integer).Value > 0 {
				return val, nil
			} else {
				return val, serr
			}
		default:
			return val, nil // TODO ... make error
		}
	case "calc":
		switch blk := es.Ser.Pop().(type) {
		case env.Block:
			ser := es.Ser
			es.Ser = blk.Series
			EvalBlockInj(es, val.(env.Object), true)
			es.Ser = ser
			return es.Res, nil
		default:
			return val, nil // TODO ... make error
		}
	case "required":
		if val == nil {
			return val, env.String{"required"}
		} else {
			return val, nil
		}
	case "integer":
		return evalInteger(val)
	case "string":
		return evalString(val)
	case "email":
		return evalEmail(val)
	case "date":
		return evalDate(val)
	default:
		return val, nil
	}
}

func BuiConvert(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object) env.Object {
	switch rmap := arg0.(type) {
	case env.RyeCtx:
		switch blk := arg1.(type) {
		case env.Block:
			ser1 := env1.Ser
			env1.Ser = blk.Series
			val, _ := Conversion_EvalBlock(env1, rmap)
			env1.Ser = ser1
			return val
		default:
			return env.NewError("arg 2 should be block")
		}
	default:
		return env.NewError("arg 1 should be Dict")
	}
}

var Builtins_conversion = map[string]*env.Builtin{

	"convert": {
		Argsn: 2,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return BuiConvert(env1, arg0, arg1)
		},
	},

	"converter": {
		Argsn: 3,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			// obj := BuiValidate(env1, arg0, arg1)
			switch obj1 := arg0.(type) {
			case env.Kind:
				switch obj2 := arg1.(type) {
				case env.Kind:
					switch spec := arg2.(type) {
					case env.Block:
						obj2.SetConverter(obj1.Kind.Index, spec)
						return obj2
					default:
						return env.NewError("3rd should be block")
					}
				default:
					return env.NewError("2nd should be block")
				}
			default:
				return env.NewError("1st should be block")
			}
			return env.NewError("error at the end")
		},
	},
}
