//go:build no_validation
// +build no_validation

package evaldo

import (
	"fmt"
	//	"net/mail"
	"github.com/refaktor/rye/env"
	// "github.com/refaktor/rye/util"
	"strconv"
	"strings"
	"time"
)

// Integer represents an integer.
type ValidationError struct {
	message string
}

func Validation_EvalBlock(es *env.ProgramState, vals env.Dict) (env.Dict, map[string]ValidationError) {

	notes := make(map[string]ValidationError, 0) // TODO ... what is this 2 here ... just for temp

	var name string
	var val any

	for es.Ser.Pos() < es.Ser.Len() {
		object := es.Ser.Pop()
		var verr *ValidationError
		switch obj := object.(type) {
		case env.Setword:
			if name != "" {
				// sets the previous value
				vals.Data[name] = val
			}
			name = es.Idx.GetWord(obj.Index)
		case env.Word:
			if name != "" {
				val, verr = evalWord(obj, es, vals.Data[name])
				if verr != nil {
					notes[name] = *verr
				} else {
					vals.Data[name] = val
				}
			}
		}
	}
	//set the last value too
	vals.Data[name] = val
	return vals, notes
}

func newVE(n string) *ValidationError {
	return &ValidationError{n}
}

func evalWord(word env.Word, es *env.ProgramState, val any) (any, *ValidationError) {
	// later get all word indexes in adwance and store them only once... then use integer comparison in switch below
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
				return val, newVE(serr.(env.String).Value)
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
			return val, newVE("required")
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

func evalInteger(val any) (any, *ValidationError) {
	switch val1 := val.(type) {
	case int64:
		return env.Integer{val1}, nil
	case env.Integer:
		return val1, nil
	case string:
		v, e := strconv.Atoi(val1)
		if e != nil {
			return val, newVE("not integer")
		} else {
			return env.Integer{int64(v)}, nil
		}
	case env.String:
		v, e := strconv.Atoi(val1.Value)
		if e != nil {
			return val, newVE("not integer")
		} else {
			return env.Integer{int64(v)}, nil
		}
	default:
		return val, newVE("not integer")
	}
}

func evalString(val any) (any, *ValidationError) {
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
		return val1, newVE("not integer")
	}
}

func parseEmail(v string) (any, *ValidationError) {
	/* e, err := mail.ParseAddress(v)
	if err != nil {
		return v, newVE("not email")
	} */
	return env.String{v}, nil
}

func evalEmail(val any) (any, *ValidationError) {
	switch val1 := val.(type) {
	case env.String:
		return parseEmail(val1.Value)
	case string:
		return parseEmail(val1)
	default:
		return val, newVE("not email")
	}
}

func parseDate(v string) (any, *ValidationError) {
	if strings.Index(v[0:3], ".") > 0 {
		d, e := time.Parse("02.01.2006", v)
		if e != nil {
			return v, newVE("not date")
		}
		fmt.Println(d)
		return env.Date{d}, nil
	} else if strings.Index(v[3:5], ":") > 0 {
		d, e := time.Parse("2006-01-02", v)
		if e != nil {
			return v, newVE("not date")
		}
		return env.Date{d}, nil
	}
	return v, newVE("not date")
}

func evalDate(val any) (any, *ValidationError) {
	switch val1 := val.(type) {
	case env.String:
		return parseDate(val1.Value)
	case string:
		return parseDate(val1)
	default:
		return val, newVE("not date")
	}
}

func something() {

	fmt.Print("1")
}

func BuiValidate(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object) env.Object {
	switch rmap := arg0.(type) {
	case env.Dict:
		switch blk := arg1.(type) {
		case env.Block:
			ser1 := env1.Ser
			env1.Ser = blk.Series
			val, _ := Validation_EvalBlock(env1, rmap)
			env1.Ser = ser1
			return val
		default:
			return env.NewError("arg 2 should be block")
		}
	default:
		return env.NewError("arg 1 should be Dict")
	}
}

var Builtins_validation = map[string]*env.Builtin{}
