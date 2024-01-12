package evaldo

import (
	"github.com/refaktor/rye/env"
)

// Integer represents an integer.
type ConversionError struct {
	message string
}

func CopyMap(m map[string]any) map[string]any {
	cp := make(map[string]any)
	for k, v := range m {
		vm, ok := v.(map[string]any)
		if ok {
			cp[k] = CopyMap(vm)
		} else {
			cp[k] = v
		}
	}

	return cp
}

func Conversion_EvalBlockCtx(ps *env.ProgramState, vals env.RyeCtx) env.Object {
	var key int

	out := env.NewEnv(nil)

	for ps.Ser.Pos() < ps.Ser.Len() {
		object := ps.Ser.Pop()
		switch obj := object.(type) {
		case env.Setword:
			key = obj.Index
		case env.LSetword:
			val, _ := vals.Get(obj.Index)
			out.Set(key, val)
		case env.Word:
			val, _ := conversion_evalWord(obj, ps, vals)
			out.Set(key, val)
		}
	}
	return *out
}

func Conversion_EvalBlockDict(ps *env.ProgramState, vals env.Dict) env.Object {
	//var outD map[string]interface{}
	outD := make(map[string]any)
	object := ps.Ser.Peek()
	switch obj := object.(type) {
	case env.Word:
		idxexcl, _ := ps.Idx.GetIndex("exclusive")
		idxinpl, _ := ps.Idx.GetIndex("inplace")
		switch obj.Index {
		case idxexcl:
			ps.Ser.Next()
		case idxinpl:
			outD = vals.Data
		}
	default:
		outD = CopyMap(vals.Data)
	}

	var key string
	var val env.Object
	var toDel any

	for ps.Ser.Pos() < ps.Ser.Len() {
		object := ps.Ser.Pop()
		switch obj := object.(type) {
		case env.Comma:
			if val != nil {
				outD[key] = val
			}
		case env.Setword:
			if val != nil {
				outD[key] = val
			}
			key = ps.Idx.GetWord(obj.Index)
			val = nil
		case env.Tagword:
			srcKey := ps.Idx.GetWord(obj.Index)
			valY := vals.Data[ps.Idx.GetWord(obj.Index)]
			val = env.ToRyeValue(valY)
			if srcKey != key {
				delete(outD, srcKey)
			}
		case env.Word:
			if val != nil {
				val, toDel = conversion_evalWord(obj, ps, val)
			} else {
				val, toDel = conversion_evalWord(obj, ps, vals)
			}
			if toDel != nil {
				delete(outD, ps.Idx.GetWord(toDel.(int)))
			}
		}
	}
	if val != nil {
		outD[key] = val
	}

	return *env.NewDict(outD)
}

func newCE(n string) *ConversionError {
	return &ConversionError{n}
}

func conversion_evalWord(word env.Word, ps *env.ProgramState, vals env.Object) (env.Object, any) {
	// later get all word indexes in adwance and store them only once... then use integer comparison in switch below
	// this is two times BAD ... first it needs to retrieve a string of index (BIG BAD) and then it compares string to string
	// instead of just comparing two integers
	switch ps.Idx.GetWord(word.Index) {
	case "calc":
		switch blk := ps.Ser.Pop().(type) {
		case env.Block:
			ser := ps.Ser
			ps.Ser = blk.Series
			EvalBlockInj(ps, vals, true)
			ps.Ser = ser
			return ps.Res, nil
		default:
			return nil, nil // TODO ... make error
		}
	case "del":
		switch blk := ps.Ser.Pop().(type) {
		case env.Tagword:
			return nil, blk.Index
		default:
			return nil, nil // TODO ... make error
		}
	default:
		return nil, nil
	}
}

func BuiConvert(ps *env.ProgramState, arg0 env.Object, arg1 env.Object) env.Object {
	switch blk := arg1.(type) {
	case env.Block:
		ser1 := ps.Ser
		ps.Ser = blk.Series
		var vals env.Object
		switch rmap := arg0.(type) {
		case env.RyeCtx:
			vals = Conversion_EvalBlockCtx(ps, rmap)
		case env.Dict:
			vals = Conversion_EvalBlockDict(ps, rmap)
		default:
			return MakeArgError(ps, 1, []env.Type{env.DictType, env.CtxType}, "BuiConvert")
		}
		ps.Ser = ser1
		return vals
	default:
		return MakeArgError(ps, 2, []env.Type{env.BlockType}, "BuiConvert")
	}
}

var Builtins_conversion = map[string]*env.Builtin{

	"convert": {
		Argsn: 2,
		Doc:   "Converts value from one kind to another.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return BuiConvert(ps, arg0, arg1)
		},
	},

	"converter": {
		Argsn: 3,
		Doc:   "Sets a converter between two kinds of objects.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			// obj := BuiValidate(ps, arg0, arg1)
			switch obj1 := arg0.(type) {
			case env.Kind:
				switch obj2 := arg1.(type) {
				case env.Kind:
					switch spec := arg2.(type) {
					case env.Block:
						obj2.SetConverter(obj1.Kind.Index, spec)
						return obj2
					default:
						return MakeArgError(ps, 3, []env.Type{env.BlockType}, "converter")
					}
				default:
					return MakeArgError(ps, 2, []env.Type{env.KindType, env.BlockType}, "converter")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.KindType}, "converter")
			}
		},
	},
}
