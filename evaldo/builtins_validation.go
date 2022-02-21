// +build !b_no_validate

package evaldo

import (
	"fmt"
	"net/mail"
	"rye/env"
	"strconv"
	"strings"
	"time"
)

// Integer represents an integer.
type ValidationError struct {
	message string
}

func Validation_EvalBlock(es *env.ProgramState, vals env.Dict) (env.Dict, map[string]env.Object) {

	notes := make(map[string]env.Object, 0) // TODO ... what is this 2 here ... just for temp

	var name string
	var val interface{}
	res := make(map[string]interface{})

	for es.Ser.Pos() < es.Ser.Len() {
		object := es.Ser.Pop()
		var verr env.Object
		switch obj := object.(type) {
		case env.Setword:
			if name != "" {
				// sets the previous value
				res[name] = JsonToRye(val)
			}
			name = es.Idx.GetWord(obj.Index)
			res[name] = JsonToRye(vals.Data[name])
		case env.Word:
			if name != "" {
				val, verr = evalWord(obj, es, res[name])
				if verr != nil {
					notes[name] = verr
				} else {
					res[name] = val
				}
			}
		}
	}
	//set the last value too
	res[name] = JsonToRye(val)
	return *env.NewDict(res), notes
}

func Validation_EvalBlock_List(es *env.ProgramState, vals env.List) (env.Object, []env.Object) {

	notes := make([]env.Object, 0) // TODO ... what is this 2 here ... just for temp

	var res env.Object

	for es.Ser.Pos() < es.Ser.Len() {
		object := es.Ser.Pop()
		var verr env.Object
		switch obj := object.(type) {
		case env.Word:
			res, verr = evalWord_List(obj, es, vals)
			if verr != nil {
				notes = append(notes, verr)
			}
		}
	}

	//set the last value too
	return res, notes
}

func newVE(n string) *ValidationError {
	return &ValidationError{n}
}

func evalWord(word env.Word, es *env.ProgramState, val interface{}) (interface{}, env.Object) {
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

func evalWord_List(word env.Word, es *env.ProgramState, vals env.List) (env.List, env.Object) {
	// later get all word indexes in adwance and store them only once... then use integer comparisson in switch below
	// this is two times BAD ... first it needs to retrieve a string of index (BIG BAD) and then it compares string to string
	// instead of just comparing two integers

	res := make([]interface{}, 0)
	switch es.Idx.GetWord(word.Index) {
	case "some":
		switch blk := es.Ser.Pop().(type) {
		case env.Block:
			for _, v := range vals.Data {
				rit := BuiValidate(es, JsonToRye(v), blk)
				res = append(res, rit)
			}
			return *env.NewList(res), nil
		default:
			return *env.NewList(res), nil // TODO ... make error
		}
	default:
		return vals, env.String{"unknown word in list validation"} // TODO --- this is not a validation error exactly, but more like error in validation code .. think about
	}
}

func evalInteger(val interface{}) (interface{}, env.Object) {
	switch val1 := val.(type) {
	case int64:
		return env.Integer{val1}, nil
	case env.Integer:
		return val1, nil
	case string:
		v, e := strconv.Atoi(val1)
		if e != nil {
			return val, env.String{"not integer"}
		} else {
			return env.Integer{int64(v)}, nil
		}
	case env.String:
		v, e := strconv.Atoi(val1.Value)
		if e != nil {
			return val, env.String{"not integer"}
		} else {
			return env.Integer{int64(v)}, nil
		}
	default:
		return val, env.String{"not integer"}
	}
}

func evalString(val interface{}) (interface{}, env.Object) {
	switch val1 := val.(type) {
	case int64:
		return env.String{strconv.FormatInt(val1, 10)}, nil
	case env.Integer:
		return env.String{strconv.FormatInt(val1.Value, 10)}, nil
	case string:
		return env.String{val1}, nil
	case env.String:
		return val1, nil
	default:
		return val1, env.String{"not string"}
	}
}

func parseEmail(v string) (interface{}, env.Object) {
	e, err := mail.ParseAddress(v)
	if err != nil {
		return v, env.String{"not email"}
	}
	return env.String{e.Address}, nil
}

func evalEmail(val interface{}) (interface{}, env.Object) {
	switch val1 := val.(type) {
	case env.String:
		return parseEmail(val1.Value)
	case string:
		return parseEmail(val1)
	default:
		return val, env.String{"not email"}
	}
}

func parseDate(v string) (interface{}, env.Object) {
	if strings.Index(v[0:3], ".") > 0 {
		d, e := time.Parse("02.01.2006", v)
		if e != nil {
			return v, env.String{"not date"}
		}
		fmt.Println(d)
		return env.Date{d}, nil
	} else if strings.Index(v[3:5], ":") > 0 {
		d, e := time.Parse("2006-01-02", v)
		if e != nil {
			return v, env.String{"not date"}
		}
		return env.Date{d}, nil
	}
	return v, env.String{"not date"}
}

func evalDate(val interface{}) (interface{}, env.Object) {
	switch val1 := val.(type) {
	case env.String:
		return parseDate(val1.Value)
	case string:
		return parseDate(val1)
	default:
		return val, env.String{"not date"}
	}
}

func BuiValidate(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object) env.Object {
	switch blk := arg1.(type) {
	case env.Block:
		switch rmap := arg0.(type) {
		case env.Dict:
			ser1 := env1.Ser
			env1.Ser = blk.Series
			val, verrs := Validation_EvalBlock(env1, rmap)
			env1.Ser = ser1
			if len(verrs) > 0 {
				env1.FailureFlag = true
				return env.NewError4(403, "validation error", nil, verrs)
			}
			return val
		case env.List:
			ser1 := env1.Ser
			env1.Ser = blk.Series
			val, _ := Validation_EvalBlock_List(env1, rmap)
			env1.Ser = ser1
			return val
		default:
			return *env.NewError("arg 1 should be Dict or List")
		}
	default:
		return *env.NewError("arg 2 should be block")
	}
}

func something() {

	fmt.Print("1")
}
