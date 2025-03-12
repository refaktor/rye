//go:build !no_validation
// +build !no_validation

package evaldo

import (
	"fmt"
	"net/mail"
	"strconv"
	"strings"
	"time"

	"github.com/refaktor/rye/env"
	"github.com/refaktor/rye/util"
)

// Integer represents an integer.
type ValidationError struct {
	message string
}

func Validation_EvalBlock(es *env.ProgramState, vals env.Dict) (env.Dict, map[string]env.Object) {
	notes := make(map[string]env.Object, 0) // TODO ... what is this 2 here ... just for temp

	var name string
	var val any
	res := make(map[string]any)

	for es.Ser.Pos() < es.Ser.Len() {
		object := es.Ser.Pop()
		var verr env.Object
		switch obj := object.(type) {
		case env.Setword:
			if name != "" {
				// sets the previous value
				res[name] = env.ToRyeValue(val)
			}
			name = es.Idx.GetWord(obj.Index)
			res[name] = env.ToRyeValue(vals.Data[name])
		case env.Word:
			if name != "" {
				val, verr = evalWord(obj, es, res[name])
				if verr != nil {
					notes[name] = verr
				} else {
					res[name] = val
				}
			}
		default:
			fmt.Println("Type is not matching - Validation_EvalBlock.")
			//TODO-FIXME
		}
	}
	//set the last value too
	res[name] = env.ToRyeValue(val)
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
		default:
			fmt.Println("Type is not matching - Validation_EvalBlock_List.")
			//TODO-FIXME
		}
	}

	//set the last value too
	return res, notes
}

func newVE(n string) *ValidationError {
	return &ValidationError{n}
}

func evalWord(word env.Word, es *env.ProgramState, val any) (any, env.Object) {
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
			return val, *env.NewString("required")
		} else {
			return val, nil
		}
	case "integer":
		return evalInteger(val)
	case "decimal":
		return evalDecimal(val)
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
	// later get all word indexes in advance and store them only once... then use integer comparison in switch below
	// this is two times BAD ... first it needs to retrieve a string of index (BIG BAD) and then it compares string to string
	// instead of just comparing two integers

	res := make([]any, 0)
	switch es.Idx.GetWord(word.Index) {
	case "some":
		switch blk := es.Ser.Pop().(type) {
		case env.Block:
			for _, v := range vals.Data {
				rit := BuiValidate(es, env.ToRyeValue(v), blk)
				res = append(res, rit)
			}
			return *env.NewList(res), nil
		default:
			return *env.NewList(res), nil // TODO ... make error
		}
	default:
		return vals, *env.NewString("unknown word in list validation") // TODO --- this is not a validation error exactly, but more like error in validation code .. think about
	}
}

func evalInteger(val any) (any, env.Object) {
	switch val1 := val.(type) {
	case int64:
		return *env.NewInteger(val1), nil
	case env.Integer:
		return val1, nil
	case string:
		v, e := strconv.Atoi(val1)
		if e != nil {
			return val, *env.NewString("not integer")
		} else {
			return *env.NewInteger(int64(v)), nil
		}
	case env.String:
		v, e := strconv.Atoi(val1.Value)
		if e != nil {
			return val, *env.NewString("not integer")
		} else {
			return *env.NewInteger(int64(v)), nil
		}
	default:
		return val, *env.NewString("not integer")
	}
}

func evalDecimal(val any) (any, env.Object) {
	switch val1 := val.(type) {
	case float64:
		return *env.NewDecimal(val1), nil
	case int64:
		return *env.NewDecimal(float64(val1)), nil
	case env.Decimal:
		return val1, nil
	case env.Integer:
		return *env.NewDecimal(float64(val1.Value)), nil
	case string:
		v, e := strconv.ParseFloat(val1, 64)
		if e != nil {
			return val, *env.NewString("not decimal")
		} else {
			return *env.NewDecimal(v), nil
		}
	case env.String:
		v, e := strconv.ParseFloat(val1.Value, 64)
		if e != nil {
			return val, *env.NewString("not decimal")
		} else {
			return *env.NewDecimal(v), nil
		}
	default:
		return val, *env.NewString("not decimal")
	}
}

func evalString(val any) (any, env.Object) {
	switch val1 := val.(type) {
	case int64:
		return *env.NewString(strconv.FormatInt(val1, 10)), nil
	case env.Integer:
		return *env.NewString(strconv.FormatInt(val1.Value, 10)), nil
	case string:
		return *env.NewString(val1), nil
	case env.String:
		return val1, nil
	default:
		return val1, *env.NewString("not string")
	}
}

func parseEmail(v string) (any, env.Object) {
	e, err := mail.ParseAddress(v)
	if err != nil {
		return v, *env.NewString("not email")
	}
	return *env.NewString(e.Address), nil
}

func evalEmail(val any) (any, env.Object) {
	switch val1 := val.(type) {
	case env.String:
		return parseEmail(val1.Value)
	case string:
		return parseEmail(val1)
	default:
		return val, *env.NewString("not email")
	}
}

func parseDate(v string) (any, env.Object) {
	if strings.Index(v[0:3], ".") > 0 {
		d, e := time.Parse("02.01.2006", v)
		if e != nil {
			return v, *env.NewString("not date")
		}
		return *env.NewDate(d), nil
	} else if strings.Index(v[3:5], "-") > 0 {
		d, e := time.Parse("2006-01-02", v)
		if e != nil {
			return v, *env.NewString("not date")
		}
		return *env.NewDate(d), nil
	}
	return v, *env.NewString("not date")
}

func evalDate(val any) (any, env.Object) {
	switch val1 := val.(type) {
	case env.String:
		return parseDate(val1.Value)
	case string:
		return parseDate(val1)
	default:
		return val, *env.NewString("not date")
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

var Builtins_validation = map[string]*env.Builtin{

	//
	// ##### Validation ##### "validation dialect for Rye values"
	//
	// Tests:
	// equal { validate dict { a: 1 } { a: required } } dict { a: 1 }
	// equal { validate dict { a: 1 } { b: optional 2 } } dict { b: 2 }
	// equal { validate dict { a: 1 } { a: optional 0 b: optional 2 } } dict { a: 1 b: 2 }
	// equal { validate dict { a: 1 } { a: required integer } } dict { a: 1 }
	// equal { validate dict { a: "1" } { a: required integer } } dict { a: 1 }
	// equal { validate dict { a: "1" } { a: required integer } -> "a" |type? } 'integer
	// equal { validate dict { a: 3.14 } { a: required decimal } } dict { a: 3.14 }
	// equal { validate dict { a: 3 } { a: required decimal } } dict { a: 3.0 }
	// equal { validate dict { a: "3.14" } { a: required decimal } } dict { a: 3.14 }
	// equal { validate dict { a: "3.14" } { a: required decimal } -> "a" |type? } 'decimal
	// equal { validate dict { a: "jim" } { a: required string } } dict { a: "jim" }
	// equal { validate dict { a: "e@ma.il" } { a: required email } } dict { a: "e@ma.il" }
	// equal { validate dict { a: "e@ma.il" } { a: required email } -> "a" |type? } 'string
	// equal { validate dict { a: "30.12.2024" } { a: required date } } dict [ "a" date "2024-12-30" ]
	// equal { validate dict { a: "2024-12-30" } { a: required date } } dict [ "a" date "2024-12-30" ]
	// equal { validate dict { a: "2024-12-30" } { a: required date } -> "a" |type? } 'date
	// equal { validate dict { a: 5 } { a: required integer check { < 10 } } } dict [ "a" 5 ]
	// equal { validate dict { a: 5 } { a: required integer calc { + 10 } } } dict [ "a" 15 ]
	// equal { validate dict { a: 5 } { b: required } |disarm |type? } 'error
	// equal { validate dict { b: "5c" } { b: optional 0 integer } |disarm |type? } 'error
	// equal { validate dict { b: "2x0" } { b: required decimal } |disarm |status? } 403   ;  ("The server understood the request, but is refusing to fulfill it"). Contrary to popular opinion, RFC2616 doesn't say "403 is only intended for failed authentication", but "403: I know what you want, but I won't do that". That condition may or may not be due to authentication.
	// equal { validate dict { b: "not-mail" } { b: required email } |disarm |message? } "validation error"
	// equal { validate dict { b: "2023-1-1" } { b: required date } |disarm |details? } dict { b: "not date" }
	// Args:
	// * data: Dictionary or List to validate
	// * rules: Block containing validation rules
	// Returns:
	// * validated Dictionary with converted values or error if validation fails
	"validate": {
		Argsn: 2,
		Doc:   "Validates and transforms data according to specified rules, returning a dictionary with converted values or an error.",
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return BuiValidate(env1, arg0, arg1)
		},
	},

	// Tests:
	// equal { validate>ctx dict { a: 1 } { a: required } |type? } 'ctx    ; TODO rename to context
	// equal { validate>ctx dict { a: 1 } { a: optional 0 } -> 'a } 1
	// Args:
	// * data: Dictionary to validate
	// * rules: Block containing validation rules
	// Returns:
	// * validated Context with converted values or error if validation fails
	"validate>ctx": {
		Argsn: 2,
		Doc:   "Validates and transforms data according to specified rules, returning a context object for easy field access.",
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			obj := BuiValidate(env1, arg0, arg1)
			switch obj1 := obj.(type) {
			case env.Dict:
				return util.Dict2Context(env1, obj1)
			default:
				return obj1
			}
		},
	},

	/*	"collect": {
			Argsn: 1,
			Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
				arg0.Trace("OPEN :::::::::")
				switch str := arg0.(type) {
				case env.Uri:
					fmt.Println(str.Path)
					db, _ := sql.Open("sqlite3", "temp-database") // TODO -- we need to make path parser in URI then this will be path
					return *env.NewNative(env1.Idx, db, "Rye-sqlite")
				default:
					return env.NewError("arg 2 should be Uri")
				}
			},
		},
		"pulldown": {
			Argsn: 2,
			Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
				arg0.Trace("OPEN :::::::::")
				switch str := arg0.(type) {
				case env.Uri:
					fmt.Println(str.Path)
					db, _ := sql.Open("sqlite3", "temp-database") // TODO -- we need to make path parser in URI then this will be path
					return *env.NewNative(env1.Idx, db, "Rye-sqlite")
				default:
					return env.NewError("arg 2 should be Uri")
				}
			},
		},*/
}
