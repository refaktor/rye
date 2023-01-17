// builtins.go
package evaldo

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"reflect"
	"rye/env"
	"sort"

	"rye/loader"
	"rye/term"
	"rye/util"
	"strconv"
	"strings"
	"time"
)

func ss() {
	fmt.Print(1)
}

func MakeError(env1 *env.ProgramState, msg string) *env.Error {
	env1.FailureFlag = true
	return env.NewError(msg)
}

func MakeRyeError(env1 *env.ProgramState, val env.Object, er *env.Error) *env.Error {
	switch val := val.(type) {
	case env.String: // todo .. make Error type .. make error construction micro dialect, return the error wrapping error that caused it
		return env.NewError4(0, val.Value, er, nil)
	case env.Integer: // todo .. make Error type .. make error construction micro dialect, return the error wrapping error that caused it
		if val.Value == 0 {
			return er
		}
		return env.NewError4(int(val.Value), "", er, nil)
	case env.Block: // todo .. make Error type .. make error construction micro dialect, return the error wrapping error that caused it
		// TODO -- this is only temporary it takes numeric value as first and string as second arg
		code := val.Series.Get(0)
		message := val.Series.Get(1)
		if code.Type() == env.IntegerType && message.Type() == env.StringType {
			return env.NewError4(int(code.(env.Integer).Value), message.(env.String).Value, er, nil)
		}

		return makeError(env1, "Wrong error constructor")
	default:
		return makeError(env1, "Wrong error constructor")
	}
}

func makeError(env1 *env.ProgramState, msg string) *env.Error {
	env1.FailureFlag = true
	return env.NewError(msg)
}

// todo -- move to util
func equalValues(ps *env.ProgramState, arg0 env.Object, arg1 env.Object) bool {
	return arg0.GetKind() == arg1.GetKind() && arg0.Inspect(*ps.Idx) == arg1.Inspect(*ps.Idx)
}

func greaterThan(ps *env.ProgramState, arg0 env.Object, arg1 env.Object) bool {
	var valA float64
	var valB float64
	switch vA := arg0.(type) {
	case env.Integer:
		valA = float64(vA.Value)
	case env.Decimal:
		valA = vA.Value
	}
	switch vB := arg1.(type) {
	case env.Integer:
		valB = float64(vB.Value)
	case env.Decimal:
		valB = vB.Value
	}
	return valA > valB
}

func greaterThanNew(arg0 env.Object, arg1 env.Object) bool {
	var valA float64
	var valB float64
	switch vA := arg0.(type) {
	case env.Integer:
		valA = float64(vA.Value)
	case env.Decimal:
		valA = vA.Value
	}
	switch vB := arg1.(type) {
	case env.Integer:
		valB = float64(vB.Value)
	case env.Decimal:
		valB = vB.Value
	}
	return valA > valB
}

func lesserThan(ps *env.ProgramState, arg0 env.Object, arg1 env.Object) bool {
	var valA float64
	var valB float64
	switch vA := arg0.(type) {
	case env.Integer:
		valA = float64(vA.Value)
	case env.Decimal:
		valA = vA.Value
	}
	switch vB := arg1.(type) {
	case env.Integer:
		valB = float64(vB.Value)
	case env.Decimal:
		valB = vB.Value
	}
	return valA < valB
}

func getFrom(ps *env.ProgramState, data interface{}, key interface{}, posMode bool) env.Object {
	switch s1 := data.(type) {
	case env.Dict:
		switch s2 := key.(type) {
		case env.String:
			v := s1.Data[s2.Value]
			switch v1 := v.(type) {
			case int, int64, float64, string, []interface{}, map[string]interface{}:
				return JsonToRye(v1)
			case env.Integer:
				return v1
			case env.String:
				return v1
			case env.Date:
				return v1
			case env.Block:
				return v1
			case env.Dict:
				return v1
			case env.List:
				return v1
			case nil:
				ps.FailureFlag = true
				return env.NewError("missing key")
			default:
				ps.FailureFlag = true
				return env.NewError("Value of type: " + reflect.TypeOf(v1).String())
			}
		}
	case env.RyeCtx:
		switch s2 := key.(type) {
		case env.Tagword:
			v, ok := s1.Get(s2.Index)
			if ok {
				return v
			} else {
				return makeError(ps, "Not found in context")
			}
		case env.Word:
			v, ok := s1.Get(s2.Index)
			if ok {
				return v
			} else {
				return makeError(ps, "Not found in context")
			}
		default:
			return makeError(ps, "Wrong type or missing key for get-arrow")
		}
	case env.List:
		switch s2 := key.(type) {
		case env.Integer:
			idx := s2.Value
			if posMode {
				idx--
			}
			v := s1.Data[idx]
			ok := true
			if ok {
				return JsonToRye(v)
			} else {
				ps.FailureFlag = true
				return env.NewError1(5) // NOT_FOUND
			}
		}
	case env.Block:
		switch s2 := key.(type) {
		case env.Integer:
			idx := s2.Value
			if posMode {
				idx--
			}
			v := s1.Series.Get(int(idx))
			ok := true
			if ok {
				return v
			} else {
				ps.FailureFlag = true
				return env.NewError1(5) // NOT_FOUND
			}
		}
	case env.SpreadsheetRow:
		switch s2 := key.(type) {
		case env.String:
			index := 0
			// find the column index
			for i := 0; i < len(s1.Uplink.Cols); i++ {
				if s1.Uplink.Cols[i] == s2.Value {
					index = i
				}
			}
			v := s1.Values[index]
			if true {
				return JsonToRye(v)
			} else {
				ps.FailureFlag = true
				return env.NewError1(5) // NOT_FOUND
			}
		case env.Integer:
			idx := s2.Value
			if posMode {
				idx--
			}
			v := s1.Values[idx]
			ok := true
			if ok {
				return JsonToRye(v)
			} else {
				ps.FailureFlag = true
				return env.NewError1(5) // NOT_FOUND
			}
		}
	}
	return makeError(ps, "Wrong type or missing key for get-arrow")
}

// Sort interface
type RyeBlockSort []env.Object

func (s RyeBlockSort) Len() int {
	return len(s)
}
func (s RyeBlockSort) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s RyeBlockSort) Less(i, j int) bool {
	return greaterThanNew(s[j], s[i])
}

var ShowResults bool

var builtins = map[string]*env.Builtin{

	"to-word": {
		Argsn: 1,
		Doc:   "Takes a String and returns a Word with that name.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch str := arg0.(type) {
			case env.String:
				idx := ps.Idx.IndexWord(str.Value)
				return env.Word{idx}
			default:
				return makeError(ps, "Arg 1 not String.")
			}
		},
	},

	"to-string": {
		Argsn: 1,
		Doc:   "Takes a Rye value and returns a string representation.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return env.String{arg0.Probe(*ps.Idx)}
		},
	},

	"is-string": {
		Argsn: 1,
		Doc:   "Returns true if value is string.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			if arg0.Type() == env.StringType {
				return env.Integer{1}
			} else {
				return env.Integer{0}
			}
		},
	},

	"to-uri": {
		Argsn: 1,
		Doc:   "Takes a Rye value and returns a string representation.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewUri1(ps.Idx, arg0.(env.String).Value) // TODO turn to switch
		},
	},

	"inc": {
		Argsn: 1,
		Doc:   "Returns integer value incremented by 1.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch arg := arg0.(type) {
			case env.Integer:
				return env.Integer{1 + arg.Value}
			default:
				return makeError(ps, "Arg 1 not Integer.")
			}
		},
	},

	"positive?": {
		Argsn: 1,
		Doc:   "Returns true if integer is positive.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch arg := arg0.(type) {
			case env.Integer:
				if arg.Value > 0 {
					return env.Integer{1}
				} else {
					return env.Integer{0}
				}
			default:
				return makeError(ps, "Arg 1 not Integer.")
			}
		},
	},

	"inc!": {
		Argsn: 1,
		Doc:   "Increments integer value by 1 in place.",
		Pure:  false,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch arg := arg0.(type) {
			case env.Tagword:
				intval, found, ctx := ps.Ctx.Get2(arg.Index)
				if found {
					switch iintval := intval.(type) {
					case env.Integer:
						ctx.Set(arg.Index, env.Integer{1 + iintval.Value})
						return env.Integer{1 + iintval.Value}
					}
				}
				return makeError(ps, "Arg 1 not Integer.")

			default:
				return makeError(ps, "Arg 1 not Integer.")
			}
		},
	},

	"change!": {
		Argsn: 2,
		Doc:   "Changes value in a word, if value changes returns true otherwise false",
		Pure:  false,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch arg := arg1.(type) {
			case env.Tagword:
				val, found, ctx := ps.Ctx.Get2(arg.Index)
				if found {
					ctx.Set(arg.Index, arg0)
					var res int64
					if arg0.GetKind() == val.GetKind() && arg0.Inspect(*ps.Idx) == val.Inspect(*ps.Idx) {
						res = 0
					} else {
						res = 1
					}
					return env.Integer{res}
				}
				return makeError(ps, "Arg 1 not Integer.")
			default:
				return makeError(ps, "Arg 1 not Integer.")
			}
		},
	},

	"set": {
		Argsn: 2,
		Doc:   "Set words by deconstructing block",
		Pure:  false,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch vals := arg0.(type) {
			case env.Block:
				switch words := arg1.(type) {
				case env.Block:
					for i, word_ := range words.Series.S {
						switch word := word_.(type) {
						case env.Word:
							// get nth value from values
							if len(vals.Series.S) < i {
								return makeError(ps, "More words than values.")
							}
							val := vals.Series.S[i]
							// if it exists then we set it to word from words
							ps.Ctx.Set(word.Index, val)
						default:
							fmt.Println(word)
							return makeError(ps, "Only words in words block")
						}
					}
					return arg0
				default:
					return makeError(ps, "Arg 1 not Integer.")
				}
			default:
				return makeError(ps, "Arg 1 not Integer.")
			}
		},
	},

	// BASIC FUNCTIONS WITH NUMBERS

	"type?": {
		Argsn: 1,
		Doc:   "Return type of a value as a word.",
		Pure:  true,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return env.Word{arg0.GetKind()}
		},
	},

	"true": {
		Argsn: 0,
		Doc:   "Retutns a truthy value.",
		Pure:  true,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return env.Integer{1}
		},
	},

	"false": {
		Argsn: 0,
		Doc:   "Retutns a falsy value.",
		Pure:  true,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return env.Integer{0}
		},
	},

	"not": {
		Argsn: 1,
		Doc:   "Turns a truthy value to non-truthy and reverse.",
		Pure:  true,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			if util.IsTruthy(arg0) {
				return env.Integer{0}
			} else {
				return env.Integer{1}
			}
		},
	},

	"require_": {
		Argsn: 1,
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			if util.IsTruthy(arg0) {
				return env.Integer{1}
			} else {
				return makeError(ps, "Requirement failed.")
			}
		},
	},

	"factor-of": {
		Argsn: 2,
		Doc:   "Checks if a Arg 1 is factor of Arg 2.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch a := arg0.(type) {
			case env.Integer:
				switch b := arg1.(type) {
				case env.Integer:
					if a.Value%b.Value == 0 {
						return env.Integer{1}
					} else {
						return env.Integer{0}
					}
				default:
					return makeError(ps, "Arg 1 not Int")
				}
			default:
				return makeError(ps, "Arg 2 not Int")
			}
		},
	},
	"odd": {
		Argsn: 1,
		Doc:   "Checks if a Arg 1 is even.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch a := arg0.(type) {
			case env.Integer:
				if a.Value%2 != 0 {
					return env.Integer{1}
				} else {
					return env.Integer{0}
				}
			default:
				return makeError(ps, "Arg 2 not Int")
			}
		},
	},
	"even": {
		Argsn: 1,
		Doc:   "Checks if a Arg 1 is even.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch a := arg0.(type) {
			case env.Integer:
				if a.Value%2 == 0 {
					return env.Integer{1}
				} else {
					return env.Integer{0}
				}
			default:
				return makeError(ps, "Arg 2 not Int")
			}
		},
	},

	"mod": {
		Argsn: 2,
		Doc:   "Calculates module of two integers.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch a := arg0.(type) {
			case env.Integer:
				switch b := arg1.(type) {
				case env.Integer:
					return env.Integer{a.Value % b.Value}
				default:
					return makeError(ps, "Arg 2 not Int")
				}
			default:
				return makeError(ps, "Arg 1 not Int")
			}
		},
	},

	"_+": {
		Argsn: 2,
		Doc:   "Adds or joins two values together (Integers, Strings, Uri-s and Blocks)",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Integer:
				switch s2 := arg1.(type) {
				case env.Integer:
					return env.Integer{s1.Value + s2.Value}
				case env.Decimal:
					return env.Decimal{float64(s1.Value) + s2.Value}
				default:
					return makeError(ps, "Integer and a wrong type")
				}
			case env.Decimal:
				switch s2 := arg1.(type) {
				case env.Integer:
					return env.Decimal{s1.Value + float64(s2.Value)}
				case env.Decimal:
					return env.Decimal{s1.Value + s2.Value}
				default:
					return makeError(ps, "Decimal and a wrong type")
				}
			case env.String:
				switch s2 := arg1.(type) {
				case env.String:
					return env.String{s1.Value + s2.Value}
				case env.Integer:
					return env.String{s1.Value + strconv.Itoa(int(s2.Value))}
				case env.Decimal:
					return env.String{s1.Value + strconv.FormatFloat(s2.Value, 'f', -1, 64)}
				default:
					return makeError(ps, "If Arg 1 is String, Arg 2 should also be String or Integer")
				}
			case env.Uri:
				switch s2 := arg1.(type) {
				case env.String:
					return *env.NewUri(ps.Idx, s1.Scheme, s1.Path+s2.Value)
				case env.Integer:
					return *env.NewUri(ps.Idx, s1.Scheme, s1.Path+strconv.Itoa(int(s2.Value)))
				case env.Block: // -- TODO turn tagwords and valvar sb strings.Builderues to uri encoded values , turn files into paths ... think more about it
					var str strings.Builder
					sepa := ""
					for i := 0; i < s2.Series.Len(); i++ {
						switch node := s2.Series.Get(i).(type) {
						case env.Tagword:
							str.WriteString(sepa + ps.Idx.GetWord(node.Index) + "=")
							sepa = "&"
						case env.String:
							str.WriteString(node.Value)
						case env.Integer:
							str.WriteString(strconv.Itoa(int(node.Value)))
						case env.Uri:
							str.WriteString(node.GetPath())
						}
					}
					return *env.NewUri(ps.Idx, s1.Scheme, s1.Path+str.String())
				default:
					return makeError(ps, "If Arg 1 is Uri, Arg 2 should be Integer, String or Block.")
				}
			case env.Block:
				switch b2 := arg1.(type) {
				case env.Block:
					s := &s1.Series
					s1.Series = *s.AppendMul(b2.Series.GetAll())
					return s1
				default:
					return makeError(ps, "If Arg 1 is Block, Arg 2 should also be Block.")
				}
			default:
				return makeError(ps, "If Arg 1 is Uri, Arg 2 should be Integer, String or Block")
			}
		},
	},

	"_-": {
		Argsn: 2,
		Doc:   "Substract two integers.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch a := arg0.(type) {
			case env.Integer:
				switch b := arg1.(type) {
				case env.Integer:
					return env.Integer{a.Value - b.Value}
				case env.Decimal:
					return env.Decimal{float64(a.Value) - b.Value}
				default:
					return makeError(ps, "Arg 1 is not Integer.")
				}
			case env.Decimal:
				switch b := arg1.(type) {
				case env.Integer:
					return env.Decimal{a.Value - float64(b.Value)}
				case env.Decimal:
					return env.Decimal{a.Value - b.Value}
				default:
					return makeError(ps, "Arg 1 is not Integer.")
				}
			default:
				return makeError(ps, "Arg 2 is not Integer.")
			}
		},
	},
	"_*": {
		Argsn: 2,
		Doc:   "Multiply two integers.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch a := arg0.(type) {
			case env.Integer:
				switch b := arg1.(type) {
				case env.Integer:
					return env.Integer{a.Value * b.Value}
				case env.Decimal:
					return env.Decimal{float64(a.Value) * b.Value}
				default:
					return makeError(ps, "Arg 1 is not Integer.")
				}
			case env.Decimal:
				switch b := arg1.(type) {
				case env.Integer:
					return env.Decimal{a.Value * float64(b.Value)}
				case env.Decimal:
					return env.Decimal{a.Value * b.Value}
				default:
					return makeError(ps, "Arg 1 is not Integer.")
				}
			default:
				return makeError(ps, "Arg 2 is not Integer.")
			}
		},
	},
	"_/": {
		Argsn: 2,
		Doc:   "Divide two integers.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch a := arg0.(type) {
			case env.Integer:
				switch b := arg1.(type) {
				case env.Integer:
					if b.Value == 0 {
						ps.FailureFlag = true
						return makeError(ps, "Can't divide by Zero.")
					}
					return env.Integer{a.Value / b.Value}
				case env.Decimal:
					if b.Value == 0.0 {
						ps.FailureFlag = true
						return makeError(ps, "Can't divide by Zero.")
					}
					return env.Decimal{float64(a.Value) / b.Value}
				default:
					return makeError(ps, "Arg 1 is not Integer.")
				}
			case env.Decimal:
				switch b := arg1.(type) {
				case env.Integer:
					if b.Value == 0 {
						ps.FailureFlag = true
						return makeError(ps, "Can't divide by Zero.")
					}
					return env.Decimal{a.Value / float64(b.Value)}
				case env.Decimal:
					if b.Value == 0.0 {
						ps.FailureFlag = true
						return makeError(ps, "Can't divide by Zero.")
					}
					return env.Decimal{a.Value / b.Value}
				default:
					return makeError(ps, "Arg 1 is not Integer.")
				}
			default:
				return makeError(ps, "Arg 1 is not Integer.")
			}
		},
	},
	"_=": {
		Argsn: 2,
		Doc:   "Test if two values are equal.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			var res int64
			if equalValues(ps, arg0, arg1) {
				res = 1
			} else {
				res = 0
			}
			return env.Integer{res}
		},
	},
	"_!": {
		Argsn: 2,
		Doc:   "Reverses the truthines.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			var res int64
			if arg0.GetKind() == arg1.GetKind() && arg0.Inspect(*ps.Idx) == arg1.Inspect(*ps.Idx) {
				res = 0
			} else {
				res = 1
			}
			return env.Integer{res}
		},
	},
	"_>": {
		Argsn: 2,
		Doc:   "Tests if Arg1 is greater than Arg 2.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			if greaterThan(ps, arg0, arg1) {
				return env.Integer{1}
			} else {
				return env.Integer{0}
			}
		},
	},
	"_>=": {
		Argsn: 2,
		Doc:   "Tests if Arg1 is greater than Arg 2.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			if equalValues(ps, arg0, arg1) || greaterThan(ps, arg0, arg1) {
				return env.Integer{1}
			} else {
				return env.Integer{0}
			}
		},
	},
	"_<": {
		Argsn: 2,
		Doc:   "Tests if Arg1 is lesser than Arg 2.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			if lesserThan(ps, arg0, arg1) {
				return env.Integer{1}
			} else {
				return env.Integer{0}
			}
		},
	},
	"_<=": {
		Argsn: 2,
		Doc:   "Tests if Arg1 is lesser than Arg 2.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			if equalValues(ps, arg0, arg1) || lesserThan(ps, arg0, arg1) {
				return env.Integer{1}
			} else {
				return env.Integer{0}
			}
		},
	},

	// BASIC GENERAL FUNCTIONS

	"prnl": {
		Argsn: 0,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			fmt.Print("\n")
			return nil
		},
	},

	"prn": {
		Argsn: 1,
		Doc:   "Prints a value and adds a space.",
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch arg := arg0.(type) {
			case env.String:
				fmt.Print(arg.Value + " ")
			default:
				fmt.Print(arg0.Probe(*env1.Idx) + " ")
			}
			return arg0
		},
	},
	"prin": {
		Argsn: 1,
		Doc:   "Prints a value without newline.",
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch arg := arg0.(type) {
			case env.String:
				fmt.Print(arg.Value)
			default:
				fmt.Print(arg0.Probe(*env1.Idx))
			}
			return arg0
		},
	},
	"print": {
		Argsn: 1,
		Doc:   "Prints a value and adds a newline.",
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch arg := arg0.(type) {
			case env.String:
				fmt.Println(arg.Value)
			default:
				fmt.Println(arg0.Probe(*env1.Idx))
			}
			return arg0
		},
	},
	"print-val": {
		Argsn: 2,
		Doc:   "Prints a value and adds a newline.",
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch arg := arg1.(type) {
			case env.String:
				vals := arg0.Probe(*env1.Idx)
				news := strings.ReplaceAll(arg.Value, "{{}}", vals)
				fmt.Println(news)
			default:
				fmt.Println(arg0.Probe(*env1.Idx))
			}
			return arg0
		},
	},
	"print-ssv": {
		Argsn: 1,
		Doc:   "Prints a value and adds a newline.",
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch arg := arg0.(type) {
			case env.Object:
				fmt.Println(util.FormatSsv(arg, *env1.Idx))
			default:
				return makeError(env1, "Not Rye object")
			}
			return arg0
		},
	},
	"print-csv": {
		Argsn: 1,
		Doc:   "Prints a value and adds a newline.",
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch arg := arg0.(type) {
			case env.Object:
				fmt.Println(util.FormatCsv(arg, *env1.Idx))
			default:
				return makeError(env1, "Not Rye object")
			}
			return arg0
		},
	},
	"print-json": {
		Argsn: 1,
		Doc:   "Prints a value and adds a newline.",
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch arg := arg0.(type) {
			case env.Object:
				fmt.Println(util.FormatJson(arg, *env1.Idx))
			default:
				return makeError(env1, "Not Rye object")
			}
			return arg0
		},
	},
	"probe": {
		Argsn: 1,
		Doc:   "Prints a probe of a value.",
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			fmt.Println(arg0.Inspect(*env1.Idx))
			return arg0
		},
	},
	"display": {
		Argsn: 1,
		Doc:   "Work in progress Interactively displays a value.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			// This is temporary implementation for experimenting what it would work like at all
			// later it should belong to the object (and the medium of display, terminal, html ..., it's part of the frontend)
			term.SaveCurPos()
			switch bloc := arg0.(type) {
			case env.Block:
				obj, esc := term.DisplayBlock(bloc, ps.Idx)
				if !esc {
					return obj
				}
			case env.Dict:
				obj, esc := term.DisplayDict(bloc, ps.Idx)
				if !esc {
					return obj
				}
			case env.Spreadsheet:
				obj, esc := term.DisplayTable(bloc, ps.Idx)
				if !esc {
					return obj
				}
			}

			return nil
		},
	},
	"mold": {
		Argsn: 1,
		Doc:   "Turn value to it's string representation.",
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			// fmt.Println()
			return env.String{arg0.Inspect(*env1.Idx)}
		},
	},

	// CONTROL WORDS

	"otherwise": {
		Argsn: 2,
		Doc:   "Conditional if not. Takes condition and a block of code.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch bloc := arg1.(type) {
			case env.Block:
				cond1 := util.IsTruthy(arg0)
				if !cond1 {
					ser := ps.Ser
					ps.Ser = bloc.Series
					EvalBlock(ps)
					ps.Ser = ser
					return ps.Res
				}
			}
			return nil
		},
	},

	"if": {
		Argsn: 2,
		Doc:   "Basic conditional. Takes a condition and a block of code.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			// function accepts 2 args. arg0 is a "boolean" value, arg1 is a block of code
			// we set bloc to block of code
			// (we don't have boolean type yet, because it's not cruicial to important part of design, neither is truthiness ... this will be decided later
			// on more operational level

			// we switch on the type of second argument, so far it should be block (later we could accept native and function)
			switch bloc := arg1.(type) {
			case env.Block:
				// TODO --- istruthy must return error if it's not possible to
				// calculate truthiness and we must here raise failure
				// we switch on type of arg0
				// if it's integer, all except 0 is true
				// if it's string, all except empty string is true
				// we don't care for other types at this stage
				cond1 := util.IsTruthy(arg0)

				// if arg0 is ok and arg1 is block we end up here
				// if cond1 is true (arg0 was truthy), otherwise we don't do anything
				// later we should return void or null, or ... we still have to decide
				if cond1 {
					// we store current series (block of code with position we are at) to temp 'ser'
					ser := ps.Ser
					// we set ProgramStates series to series ob the block
					ps.Ser = bloc.Series
					// we eval the block (current context / scope stays the same as it was in parent block)
					// Inj means we inject the condition value into the block, because it costs us very little. we could do "if name { .print }"
					EvalBlockInj(ps, arg0, true)
					// we set temporary series back to current program state
					ps.Ser = ser
					// we return the last return value (the return value of executing the block) "a: if 1 { 100 }" a becomes 100,
					// in future we will also handle the "else" case, but we have to decide
					return ps.Res
				}
			default:
				// if it's not a block we return error for now
				ps.FailureFlag = true
				return env.NewError("Error if")
			}
			return nil
		},
	},

	"^if": {
		Argsn: 2,
		Doc:   "Basic conditional with a Returning mechanism when true. Takes a condition and a block of code.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch bloc := arg1.(type) {
			case env.Block:
				cond1 := util.IsTruthy(arg0)
				if cond1 {
					ser := ps.Ser
					ps.Ser = bloc.Series
					EvalBlockInj(ps, arg0, true)
					ps.Ser = ser
					ps.ReturnFlag = true
					return ps.Res
				}
			default:
				return makeError(ps, "Arg 2 not Block.")
			}
			return nil
		},
	},

	"^otherwise": {
		Argsn: 2,
		Doc:   "Basic conditional with a Returning mechanism when true. Takes a condition and a block of code.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch bloc := arg1.(type) {
			case env.Block:
				cond1 := util.IsTruthy(arg0)
				if !cond1 {
					ser := ps.Ser
					ps.Ser = bloc.Series
					EvalBlockInj(ps, arg0, true)
					ps.Ser = ser
					ps.ReturnFlag = true
					return ps.Res
				}
			default:
				return makeError(ps, "Arg 2 not Block.")
			}
			return nil
		},
	},

	"either": {
		Argsn: 3,
		Doc:   "The if/else conditional. Takes a value and true and false block of code.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			//arg0.Trace("")
			//arg1.Trace("")
			//arg2.Trace("")
			var cond1 bool
			switch bloc1 := arg1.(type) {
			case env.Block:
				switch bloc2 := arg2.(type) {
				case env.Block:
					switch cond := arg0.(type) {
					case env.Integer:
						cond1 = cond.Value != 0
					case env.String:
						cond1 = cond.Value != ""
					default:
						return env.NewError("Error either")
					}
					ser := ps.Ser
					if cond1 {
						ps.Ser = bloc1.Series
						ps.Ser.Reset()
					} else {
						ps.Ser = bloc2.Series
						ps.Ser.Reset()
					}
					EvalBlockInj(ps, arg0, true)
					ps.Ser = ser
					return ps.Res
				}

			}
			return nil
		},
	},

	"^tidy-switch": {
		Argsn:         2,
		AcceptFailure: true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			fmt.Println("FLAGS")

			ps.FailureFlag = false
			//fmt.Println(arg0.Probe(*ps.Idx))

			switch er := arg0.(type) {
			case env.Error:
				fmt.Println("ERR")

				switch bloc := arg1.(type) {
				case env.Block:

					var code env.Object

					any_found := false
					fmt.Println("BLOCK")

					for i := 0; i < bloc.Series.Len(); i += 2 {
						fmt.Println("LOOP")

						if i > bloc.Series.Len()-2 {
							return env.NewError("switch block malformed")
						}

						switch ev := bloc.Series.Get(i).(type) {
						case env.Integer:
							if er.Status == int(ev.Value) {
								any_found = true
								code = bloc.Series.Get(i + 1)
							}
						case env.Void:
							fmt.Println("VOID")
							if !any_found {
								code = bloc.Series.Get(i + 1)
								any_found = false
							}
						}
					}
					switch cc := code.(type) {
					case env.Block:
						fmt.Println(code.Probe(*ps.Idx))
						// we store current series (block of code with position we are at) to temp 'ser'
						ser := ps.Ser
						// we set ProgramStates series to series ob the block
						ps.Ser = cc.Series
						// we eval the block (current context / scope stays the same as it was in parent block)
						// Inj means we inject the condition value into the block, because it costs us very little. we could do "if name { .print }"
						EvalBlockInj(ps, arg0, true)
						// we set temporary series back to current program state
						ps.Ser = ser
						// we return the last return value (the return value of executing the block) "a: if 1 { 100 }" a becomes 100,
						// in future we will also handle the "else" case, but we have to decide
						//						ps.ReturnFlag = true

						ps.ReturnFlag = true
						ps.FailureFlag = true
						return arg0
					default:
						// if it's not a block we return error for now
						ps.FailureFlag = true
						return env.NewError("Malformed switch block")
					}
				default:
					// if it's not a block we return error for now
					ps.FailureFlag = true
					return env.NewError("Second arg not block")
				}
			default:
				return arg0
			}
		},
	},

	"switch": {
		Argsn:         2,
		Doc:           "Classic switch function. Takes a word and multiple possible values and block of code to do.",
		AcceptFailure: true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch bloc := arg1.(type) {
			case env.Block:

				var code env.Object

				any_found := false
				//fmt.Println("BLOCK")

				for i := 0; i < bloc.Series.Len(); i += 2 {
					//fmt.Println("LOOP")

					if i > bloc.Series.Len()-2 {
						return env.NewError("switch block malformed")
					}

					ev := bloc.Series.Get(i)
					if arg0.GetKind() == ev.GetKind() && arg0.Inspect(*ps.Idx) == ev.Inspect(*ps.Idx) {
						any_found = true
						code = bloc.Series.Get(i + 1)
					}
				}
				if any_found {
					switch cc := code.(type) {
					case env.Block:
						// fmt.Println(code.Probe(*ps.Idx))
						// we store current series (block of code with position we are at) to temp 'ser'
						ser := ps.Ser
						// we set ProgramStates series to series ob the block
						ps.Ser = cc.Series
						// we eval the block (current context / scope stays the same as it was in parent block)
						// Inj means we inject the condition value into the block, because it costs us very little. we could do "if name { .print }"
						EvalBlockInj(ps, arg0, true)
						// we set temporary series back to current program state
						ps.Ser = ser
						// we return the last return value (the return value of executing the block) "a: if 1 { 100 }" a becomes 100,
						// in future we will also handle the "else" case, but we have to decide
						//						ps.ReturnFlag = true
						return ps.Res
					default:
						// if it's not a block we return error for now
						ps.FailureFlag = true
						return env.NewError("Malformed switch block")
					}
				}
				return arg0
			default:
				// if it's not a block we return error for now
				ps.FailureFlag = true
				return env.NewError("Second arg not block")
			}
		},
	},

	"cases": {
		Argsn: 2,
		Doc:   "Similar to Case function, but checks all the cases, even after a match. It combines the outputs.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			// function accepts 2 args. arg0 is a "boolean" value, arg1 is a block of code
			// we set bloc to block of code
			// (we don't have boolean type yet, because it's not cruicial to important part of design, neither is truthiness ... this will be decided later
			// on more operational level

			// we switch on the type of second argument, so far it should be block (later we could accept native and function)
			switch bloc := arg1.(type) {
			case env.Block:
				// TODO --- istruthy must return error if it's not possible to
				// calculate truthiness and we must here raise failure
				// we switch on type of arg0
				// if it's integer, all except 0 is true
				// if it's string, all except empty string is true
				// we don't care for other types at this stage
				ser := ps.Ser

				cumul := arg0

				foundany := false
				for {

					doblk := false
					cond_ := bloc.Series.Pop()
					blk := bloc.Series.Pop().(env.Block)

					switch cond := cond_.(type) {
					case env.Block:
						ps.Ser = cond.Series
						// we eval the block (current context / scope stays the same as it was in parent block)
						// Inj means we inject the condition value into the block, because it costs us very little. we could do "if name { .print }"
						EvalBlock(ps)
						// we set temporary series back to current program state
						if util.IsTruthy(ps.Res) {
							doblk = true
							foundany = true
						}
					case env.Void:
						if foundany == false {
							doblk = true
						}
					}
					// we set ProgramStates series to series ob the block
					if doblk {
						ps.Ser = blk.Series
						// we eval the block (current context / scope stays the same as it was in parent block)
						// Inj means we inject the condition value into the block, because it costs us very little. we could do "if name { .print }"
						EvalBlockInj(ps, cumul, true)
						cumul = ps.Res
					}
					if bloc.Series.AtLast() {
						break
					}
				}
				ps.Ser = ser
				// we return the last return value (the return value of executing the block) "a: if 1 { 100 }" a becomes 100,
				// in future we will also handle the "else" case, but we have to decide
				return cumul
			default:
				// if it's not a block we return error for now
				ps.FailureFlag = true
				return env.NewError("Error if")
			}
			return nil
		},
	},

	"enter-console": {
		Argsn: 1,
		Doc:   "Stops execution and gives you a Rye console, to test the code inside environment.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch name := arg0.(type) {
			case env.String:
				/* ser := ps.Ser
				ps.Ser = bloc.Series
				EvalBlock(ps)
				ps.Ser = ser */
				//reader := bufio.NewReader(os.Stdin)

				fmt.Println("Welcome to console: \033[1m" + name.Value + "\033[0m")
				fmt.Println("* use \033[1mls\033[0m to list current context")
				fmt.Println("-------------------------------------------------------------")
				/*
					for {
						fmt.Print("{ rye dropin }")
						text, _ := reader.ReadString('\n')
						//fmt.Println(1111)
						// convert CRLF to LF
						text = strings.Replace(text, "\n", "", -1)
						//fmt.Println(1111)
						if strings.Compare("(lc)", text) == 0 {
							fmt.Println(ps.Ctx.Probe(*ps.Idx))
						} else if strings.Compare("(r)", text) == 0 {
							ps.Ser = ser
							return ps.Res
						} else {
							// fmt.Println(1111)
							block, genv := loader.LoadString("{ " + text + " }")
							ps := env.AddToProgramState(ps, block.Series, genv)
							EvalBlock(ps)
							fmt.Println(ps.Res.Inspect(*ps.Idx))
						}
					}*/

				DoRyeRepl(ps, ShowResults)
				fmt.Println("-------------------------------------------------------------")
				// ps.Ser = ser

				return ps.Res
			}
			return nil
		},
	},

	"do": {
		Argsn: 1,
		Doc:   "Takes a block of code and does (runs) it.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch bloc := arg0.(type) {
			case env.Block:
				ser := ps.Ser
				ps.Ser = bloc.Series
				EvalBlock(ps)
				ps.Ser = ser
				return ps.Res
			}
			return nil
		},
	},

	"do-with": {
		Argsn: 2,
		Doc:   "Takes a value and a block of code. It does the code with the value injected.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch bloc := arg1.(type) {
			case env.Block:
				ser := ps.Ser
				ps.Ser = bloc.Series
				EvalBlockInj(ps, arg0, true)
				ps.Ser = ser
				return ps.Res
			}
			return nil
		},
	},

	"do-in": {
		Argsn: 2,
		Doc:   "Takes a Context and a Block. It Does a block inside a given Context.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch ctx := arg0.(type) {
			case env.RyeCtx:
				switch bloc := arg1.(type) {
				case env.Block:
					ser := ps.Ser
					ps.Ser = bloc.Series
					EvalBlockInCtx(ps, &ctx)
					ps.Ser = ser
					return ps.Res
				default:
					ps.ErrorFlag = true
					return env.NewError("Arg 2 should be block")

				}
			default:
				ps.ErrorFlag = true
				return env.NewError("Arg 1 should be context")
			}

		},
	},

	"eval": {
		Argsn: 1,
		Doc:   "Takes a block of Rye values and evaluates each value or expression.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch bloc := arg0.(type) {
			case env.Block:
				ser := ps.Ser
				ps.Ser = bloc.Series
				res := make([]env.Object, 0)
				for ps.Ser.Pos() < ps.Ser.Len() {
					// ps, injnow = EvalExpressionInj(ps, inj, injnow)
					EvalExpression2(ps, false)
					res = append(res, ps.Res)
					// check and raise the flags if needed if true (error) return
					//if checkFlagsAfterBlock(ps, 101) {
					//	return ps
					//}
					// if return flag was raised return ( errorflag I think would return in previous if anyway)
					//if checkErrorReturnFlag(ps) {
					//	return ps
					//}
					// ps, injnow = MaybeAcceptComma(ps, inj, injnow)
				}
				ps.Ser = ser
				return *env.NewBlock(*env.NewTSeries(res))
			case env.Word:
				val, found := ps.Ctx.Get(bloc.Index)
				if found {
					return val
				}
				return makeError(ps, "Value not bound")
			}
			return nil
		},
	},

	"eval\\with": {
		Argsn: 2,
		Doc:   "",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch bloc := arg1.(type) {
			case env.Block:
				ser := ps.Ser
				ps.Ser = bloc.Series
				res := make([]env.Object, 0)
				for ps.Ser.Pos() < ps.Ser.Len() {
					// ps, injnow = EvalExpressionInj(ps, inj, injnow)
					EvalExpressionInjLimited(ps, arg0, true)
					res = append(res, ps.Res)
					// check and raise the flags if needed if true (error) return
					//if checkFlagsAfterBlock(ps, 101) {
					//	return ps
					//}
					// if return flag was raised return ( errorflag I think would return in previous if anyway)
					//if checkErrorReturnFlag(ps) {
					//	return ps
					//}
					// ps, injnow = MaybeAcceptComma(ps, inj, injnow)
				}
				ps.Ser = ser
				return *env.NewBlock(*env.NewTSeries(res))
			}
			return nil
		},
	},

	"all": {
		Argsn: 1,
		Doc:   "Takes a block, if all values or expressions are truthy it returns the last one, otherwise false.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch bloc := arg0.(type) {
			case env.Block:
				ser := ps.Ser
				ps.Ser = bloc.Series
				for ps.Ser.Pos() < ps.Ser.Len() {
					EvalExpression2(ps, false)
					if !util.IsTruthy(ps.Res) {
						break
					}
				}
				ps.Ser = ser
				return ps.Res
			}
			return makeError(ps, "Arg 1 not Block")
		},
	},

	"any": {
		Argsn: 1,
		Doc:   "Takes a block, if any of the values or expressions are truthy, the it returns that one, in none false.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch bloc := arg0.(type) {
			case env.Block:
				ser := ps.Ser
				ps.Ser = bloc.Series
				for ps.Ser.Pos() < ps.Ser.Len() {
					EvalExpression2(ps, false)
					if util.IsTruthy(ps.Res) {
						break
					}
				}
				ps.Ser = ser
				return ps.Res
			}
			return makeError(ps, "Arg 1 not Block")
		},
	},

	"any\\with": {
		Argsn: 2,
		Doc:   "Takes a block, if any of the values or expressions are truthy, the it returns that one, in none false.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch bloc := arg1.(type) {
			case env.Block:
				ser := ps.Ser
				ps.Ser = bloc.Series
				for ps.Ser.Pos() < ps.Ser.Len() {
					EvalExpressionInjLimited(ps, arg0, true)
					//					EvalExpression2(ps, false)
					if util.IsTruthy(ps.Res) {
						break
					}
				}
				ps.Ser = ser
				return ps.Res
			}
			return makeError(ps, "Arg 1 not Block")
		},
	},

	"range": {
		Argsn: 2,
		Doc:   "Takes two integers and returns a block of integers between them. (Will change to lazy list/generator later)",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch i1 := arg0.(type) {
			case env.Integer:
				switch i2 := arg1.(type) {
				case env.Integer:
					objs := make([]env.Object, i2.Value-i1.Value+1)
					idx := 0
					for i := i1.Value; i <= i2.Value; i++ {
						objs[idx] = env.Integer{i}
						idx += 1
					}
					return *env.NewBlock(*env.NewTSeries(objs))
				}
				return makeError(ps, "Arg 1 not Int")

			}
			return makeError(ps, "Arg 1 not Int")
		},
	},

	// SPECIAL FUNCTION FUNCTIONS

	// CONTEXT FUNCTIONS

	"returns": {
		Argsn: 1,
		Doc:   "Sets up a value to return at the end of function.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			ps.ForcedResult = arg0
			return arg0
		},
	},

	"collect": {
		Argsn: 1,
		Doc:   "Collects values into an implicit block.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			if ps.ForcedResult == nil || ps.ForcedResult.Type() != env.BlockType {
				ps.ForcedResult = *env.NewBlock(*env.NewTSeries(make([]env.Object, 0)))
			}
			block := ps.ForcedResult.(env.Block)
			block.Series.Append(arg0)
			ps.ForcedResult = block
			return arg0
		},
	},

	"collect-key": {
		Argsn: 2,
		Doc:   "Collecs key value pars to implicit block.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			if ps.ForcedResult == nil || ps.ForcedResult.Type() != env.BlockType {
				ps.ForcedResult = *env.NewBlock(*env.NewTSeries(make([]env.Object, 0)))
			}
			block := ps.ForcedResult.(env.Block)
			block.Series.Append(arg1)
			block.Series.Append(arg0)
			ps.ForcedResult = block
			return arg0
		},
	},

	"collect-update-key": {
		Argsn: 2,
		Doc:   "Collecs key value pars to implicit block.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			if ps.ForcedResult == nil || ps.ForcedResult.Type() != env.BlockType {
				ps.ForcedResult = *env.NewBlock(*env.NewTSeries(make([]env.Object, 0)))
			}
			block := ps.ForcedResult.(env.Block)
			block.Series.Append(arg1)
			block.Series.Append(arg0)
			ps.ForcedResult = block
			return arg0
		},
	},

	"collected": {
		Argsn: 0,
		Doc:   "Returns the implicit data structure that we collected t",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return ps.ForcedResult
		},
	},

	"pop-collected": {
		Argsn: 0,
		Doc:   "Retursn the implicit collected data structure and resets it.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			result := ps.ForcedResult
			ps.ForcedResult = nil
			return result
		},
	},

	"current-context": {
		Argsn: 0,
		Doc:   "Returns current context.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *ps.Ctx
		},
	},

	"ls": {
		Argsn: 0,
		Doc:   "Lists words in current context",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			fmt.Println(ps.Ctx.Preview(*ps.Idx, ""))
			return env.Integer{1}
		},
	},

	"ls\\": {
		Argsn: 1,
		Doc:   "Lists words in current context with string filter",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.String:
				fmt.Println(ps.Ctx.Preview(*ps.Idx, s1.Value))
				return env.Integer{1}
			}
			return nil

		},
	},

	"parent-context": {
		Argsn: 0,
		Doc:   "Returns parent context of the current context.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *ps.Ctx.Parent
		},
	},

	"raw-context": {
		Argsn: 1,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch bloc := arg0.(type) {
			case env.Block:
				ser := ps.Ser
				ctx := ps.Ctx
				ps.Ser = bloc.Series
				ps.Ctx = env.NewEnv(nil) // make new context with no parent
				EvalBlock(ps)
				rctx := ps.Ctx
				ps.Ctx = ctx
				ps.Ser = ser
				if ps.ErrorFlag {
					return ps.Res
				}
				return *rctx // return the resulting context
			}
			return nil
		},
	},

	"isolate": {
		Argsn: 1,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch bloc := arg0.(type) {
			case env.Block:
				ser := ps.Ser
				ctx := ps.Ctx
				ps.Ser = bloc.Series
				ps.Ctx = env.NewEnv(ps.Ctx) // make new context with no parent
				EvalBlock(ps)
				rctx := ps.Ctx
				rctx.Parent = nil
				ps.Ctx = ctx
				ps.Ser = ser
				if ps.ErrorFlag {
					return ps.Res
				}
				return *rctx // return the resulting context
			}
			return nil
		},
	},

	"context": {
		Argsn: 1,
		Doc:   "Creates a new context with no parent",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch bloc := arg0.(type) {
			case env.Block:
				ser := ps.Ser
				ctx := ps.Ctx
				ps.Ser = bloc.Series
				ps.Ctx = env.NewEnv(ps.Ctx) // make new context with no parent
				EvalBlock(ps)
				rctx := ps.Ctx
				ps.Ctx = ctx
				ps.Ser = ser
				return *rctx // return the resulting context
			}
			return nil
		},
	},

	"private": {
		Argsn: 1,
		Doc:   "Creates a new context with current parent, returns last value",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch bloc := arg0.(type) {
			case env.Block:
				ser := ps.Ser
				ctx := ps.Ctx
				ps.Ser = bloc.Series
				ps.Ctx = env.NewEnv(ps.Ctx) // make new context with no parent
				EvalBlock(ps)
				// rctx := ps.Ctx
				ps.Ctx = ctx
				ps.Ser = ser
				return ps.Res // return the resulting context
			}
			return nil
		},
	},

	"private\\": {
		Argsn: 2,
		Doc:   "Creates a new context with current parent, returns last value",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch doc := arg0.(type) {
			case env.String:
				switch bloc := arg1.(type) {
				case env.Block:
					ser := ps.Ser
					ctx := ps.Ctx
					ps.Ser = bloc.Series
					ps.Ctx = env.NewEnv2(ps.Ctx, doc.Value) // make new context with no parent
					EvalBlock(ps)
					// rctx := ps.Ctx
					ps.Ctx = ctx
					ps.Ser = ser
					return ps.Res // return the resulting context

				default:
					return makeError(ps, "Arg 2 not Block")
				}
			default:
				return makeError(ps, "Arg 1 not String")
			}
		},
	},

	"extend": { // exclamation mark, because it as it is now extends/changes the source context too .. in place
		Argsn: 2,
		Doc:   "Extends a context with a new context in place and returns it.",
		Pure:  false,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch ctx0 := arg0.(type) {
			case env.RyeCtx:
				switch bloc := arg1.(type) {
				case env.Block:
					ser := ps.Ser
					ctx := ps.Ctx
					ps.Ser = bloc.Series
					ps.Ctx = ctx0.Copy() // make new context with no parent
					EvalBlock(ps)
					rctx := ps.Ctx
					ps.Ctx = ctx
					ps.Ser = ser
					return *rctx // return the resulting context
				}
			}
			ps.ErrorFlag = true
			return env.NewError("Second argument should be block, builtin (or function).")
		},
	},

	"bind": { // TODO -- check if this works
		Argsn: 2,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch swCtx1 := arg0.(type) {
			case env.RyeCtx:
				switch swCtx2 := arg1.(type) {
				case env.RyeCtx:
					swCtx1.Parent = &swCtx2
					return swCtx1
				}
			}
			return env.NewError("wrong args")
		},
	},

	"unbind": {
		Argsn: 1,
		Doc:   "Accepts a Context and unbinds it from it's parent Context.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch swCtx1 := arg0.(type) {
			case env.RyeCtx:
				swCtx1.Parent = nil
				return swCtx1
			}
			return env.NewError("wrong args")
		},
	},

	// COMBINATORS

	"pass": {
		Argsn: 2,
		Doc:   "Accepts a value and a block. It does the block, with value injected, and returns (passes on) the initial value.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch bloc := arg1.(type) {
			case env.Block:
				res := arg0
				ser := ps.Ser
				ps.Ser = bloc.Series
				EvalBlockInj(ps, arg0, true)
				ps.Ser = ser
				if ps.ReturnFlag {
					return ps.Res
				}
				return res
			default:
				return makeError(ps, "Arg 2 should be Block.")
			}
		},
	},

	"keep": {
		Argsn: 3,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch b1 := arg1.(type) {
			case env.Block:
				switch b2 := arg2.(type) {
				case env.Block:
					ser := ps.Ser
					ps.Ser = b1.Series
					EvalBlockInj(ps, arg0, true)
					res := ps.Res
					ps.Ser = b2.Series
					EvalBlockInj(ps, arg0, true)
					ps.Ser = ser
					return res
				}
			}
			return nil
		},
	},

	"with": {
		AcceptFailure: true,
		Doc:           "Do a block with Arg 1 injected.",
		Argsn:         2,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch bloc := arg1.(type) {
			case env.Block:
				ser := ps.Ser
				ps.Ser = bloc.Series
				EvalBlockInj(ps, arg0, true)
				ps.Ser = ser
				return ps.Res
			}
			return nil
		},
	},

	//

	"time-it": {
		Argsn: 1,
		Doc:   "Accepts a block, does it and times it's execution time.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch bloc := arg0.(type) {
			case env.Block:
				ser := ps.Ser
				ps.Ser = bloc.Series
				start := time.Now()
				EvalBlock(ps)
				t := time.Now()
				elapsed := t.Sub(start)
				ps.Ser = ser
				return env.Integer{elapsed.Nanoseconds() / 1000000}
			}
			return nil
		},
	},

	"loop": {
		Argsn: 2,
		Doc:   "Accepts a number and a block of code. Does the block of code number times, injecting the number.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch cond := arg0.(type) {
			case env.Integer:
				switch bloc := arg1.(type) {
				case env.Block:
					ser := ps.Ser
					ps.Ser = bloc.Series
					for i := 0; int64(i) < cond.Value; i++ {
						ps = EvalBlockInj(ps, env.Integer{int64(i + 1)}, true)
						ps.Ser.Reset()
					}
					ps.Ser = ser
					return ps.Res
				}
			}
			return nil
		},
	},

	"produce": {
		Argsn: 3,
		Doc:   "Accepts a number and a block of code. Does the block of code number times, injecting the number.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch cond := arg0.(type) {
			case env.Integer:
				switch bloc := arg2.(type) {
				case env.Block:
					acc := arg1
					ser := ps.Ser
					ps.Ser = bloc.Series
					for i := 0; int64(i) < cond.Value; i++ {
						ps = EvalBlockInj(ps, acc, true)
						ps.Ser.Reset()
						acc = ps.Res
					}
					ps.Ser = ser
					return ps.Res
				}
			}
			return nil
		},
	},

	"produce\\": {
		Argsn: 4,
		Doc:   "produce\\ 5 1 'acc { * acc , + 1 }",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch cond := arg0.(type) {
			case env.Integer:
				switch bloc := arg3.(type) {
				case env.Block:
					switch accu := arg2.(type) {
					case env.Tagword:
						acc := arg1
						ps.Ctx.Set(accu.Index, acc)
						ser := ps.Ser
						ps.Ser = bloc.Series
						for i := 0; int64(i) < cond.Value; i++ {
							ps = EvalBlockInj(ps, acc, true)
							ps.Ser.Reset()
							acc = ps.Res
						}
						ps.Ser = ser
						val, _ := ps.Ctx.Get(accu.Index)
						return val
					}
				}
			}
			return nil
		},
	},

	"forever": {
		Argsn: 1,
		Doc:   "Accepts a block and does it forever.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch bloc := arg0.(type) {
			case env.Block:
				ser := ps.Ser
				ps.Ser = bloc.Series
				for {
					ps = EvalBlock(ps)
					ps.Ser.Reset()
				}
				ps.Ser = ser
				return ps.Res
			default:
				ps.FailureFlag = true
				return env.NewError("arg0 should be block	")
			}
		},
	},
	"forever-with": {
		Argsn: 2,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch bloc := arg1.(type) {
			case env.Block:
				ser := ps.Ser
				ps.Ser = bloc.Series
				for {
					EvalBlockInj(ps, arg0, true)
					ps.Ser.Reset()
				}
				ps.Ser = ser
				return ps.Res
			default:
				ps.FailureFlag = true
				return env.NewError("arg0 should be block	")
			}
		},
	},

	"for": {
		Argsn: 2,
		Doc:   "Accepts a block of values and a block of code, does the code for each of the values, injecting them.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch block := arg0.(type) {
			case env.String:
				switch code := arg1.(type) {
				case env.Block:
					ser := ps.Ser
					ps.Ser = code.Series
					for i := 0; i < len(block.Value); i++ {
						ps = EvalBlockInj(ps, env.String{string(block.Value[i])}, true)
						if ps.ErrorFlag {
							return ps.Res
						}
						ps.Ser.Reset()
					}
					ps.Ser = ser
					return ps.Res
				}
			case env.Block:
				switch code := arg1.(type) {
				case env.Block:
					ser := ps.Ser
					ps.Ser = code.Series
					for i := 0; i < block.Series.Len(); i++ {
						ps = EvalBlockInj(ps, block.Series.Get(i), true)
						ps.Ser.Reset()
					}
					ps.Ser = ser
					return ps.Res
				}
			case env.List:
				switch code := arg1.(type) {
				case env.Block:
					ser := ps.Ser
					ps.Ser = code.Series
					for i := 0; i < len(block.Data); i++ {
						ps = EvalBlockInj(ps, JsonToRye(block.Data[i]), true)
						if ps.ErrorFlag {
							return ps.Res
						}
						ps.Ser.Reset()
					}
					ps.Ser = ser
					return ps.Res
				}
			case env.Spreadsheet:
				switch code := arg1.(type) {
				case env.Block:
					ser := ps.Ser
					ps.Ser = code.Series
					if block.RawMode {
						for i := 0; i < len(block.RawRows); i++ {
							row := block.RawRows[i]
							row2 := make([]interface{}, len(row))
							for i := range row {
								row2[i] = row[i]
							}

							row3 := env.NewList(row2)
							ps = EvalBlockInj(ps, row3, true)
							if ps.ErrorFlag {
								return ps.Res
							}
							ps.Ser.Reset()
						}
					} else {
						for i := 0; i < len(block.Rows); i++ {
							row := block.Rows[i]
							row.Uplink = &block
							ps = EvalBlockInj(ps, row, true)
							if ps.ErrorFlag {
								return ps.Res
							}
							ps.Ser.Reset()
						}
					}
					ps.Ser = ser
					return ps.Res
				}
			}
			return nil
		},
	},

	"purge": { // TODO ... doesn't fully work
		Argsn: 2,
		Doc:   "Purges values from a seris based on return of a injected code block.",
		Pure:  false,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch block := arg0.(type) {
			case env.Block:
				switch code := arg1.(type) {
				case env.Block:
					ser := ps.Ser
					ps.Ser = code.Series
					for i := 0; i < block.Series.Len(); i++ {
						ps = EvalBlockInj(ps, block.Series.Get(i), true)
						if util.IsTruthy(ps.Res) {
							block.Series.S = append(block.Series.S[:i], block.Series.S[i+1:]...)
							i--
						}
						ps.Ser.Reset()
					}
					ps.Ser = ser
					return block
				}
			case env.List:
				switch code := arg1.(type) {
				case env.Block:
					ser := ps.Ser
					ps.Ser = code.Series
					for i := 0; i < len(block.Data); i++ {
						ps = EvalBlockInj(ps, JsonToRye(block.Data[i]), true)
						if util.IsTruthy(ps.Res) {
							block.Data = append(block.Data[:i], block.Data[i+1:]...)
							i--
						}
						ps.Ser.Reset()
					}
					ps.Ser = ser
					return block
				}
			case env.Spreadsheet:
				switch code := arg1.(type) {
				case env.Block:
					ser := ps.Ser
					ps.Ser = code.Series
					for i := 0; i < len(block.Rows); i++ {
						ps = EvalBlockInj(ps, block.Rows[i], true)
						if util.IsTruthy(ps.Res) {
							block.Rows = append(block.Rows[:i], block.Rows[i+1:]...)
							i--
						}
						ps.Ser.Reset()
					}
					ps.Ser = ser
					return block
				}
			}
			return nil
		},
	},

	"purge!": { // TODO ... doesn't fully work
		Argsn: 2,
		Doc:   "Purges values from a seris based on return of a injected code block.",
		Pure:  false,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch wrd := arg1.(type) {
			case env.Tagword:
				val, found, ctx := ps.Ctx.Get2(wrd.Index)
				if found {
					switch block := val.(type) {
					case env.Block:
						switch code := arg0.(type) {
						case env.Block:
							ser := ps.Ser
							ps.Ser = code.Series
							for i := 0; i < block.Series.Len(); i++ {
								ps = EvalBlockInj(ps, block.Series.Get(i), true)
								if util.IsTruthy(ps.Res) {
									block.Series.S = append(block.Series.S[:i], block.Series.S[i+1:]...)
									i--
								}
								ps.Ser.Reset()
							}
							ps.Ser = ser
							ctx.Set(wrd.Index, block)
							return block
						}
					}
				}
			}
			return nil
		},
	},

	// map should at the end map over block, raw-map, etc ...
	// it should accept a block of code, a function and a builtin
	// it should use injected block so it doesn't need a variable defined like map [ 1 2 3 ] x [ add a 100 ]
	// map [ 1 2 3 ] { .add 3 }
	"map": {
		Argsn: 2,
		Doc:   "Maps values of a block to a new block by evaluating a block of code.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch list := arg0.(type) {
			case env.Block:
				switch block := arg1.(type) {
				case env.Block, env.Builtin:
					l := list.Series.Len()
					newl := make([]env.Object, l)
					switch block := block.(type) {
					case env.Block:
						ser := ps.Ser
						ps.Ser = block.Series
						for i := 0; i < l; i++ {
							ps = EvalBlockInj(ps, list.Series.Get(i), true)
							newl[i] = ps.Res
							ps.Ser.Reset()
						}
						ps.Ser = ser
					case env.Builtin:
						for i := 0; i < l; i++ {
							newl[i] = DirectlyCallBuiltin(ps, block, list.Series.Get(i), nil)
						}
					}
					return *env.NewBlock(*env.NewTSeries(newl))
				}
			case env.List:
				switch block := arg1.(type) {
				case env.Block, env.Builtin:
					l := len(list.Data)
					newl := make([]interface{}, l)
					switch block := block.(type) {
					case env.Block:
						ser := ps.Ser
						ps.Ser = block.Series
						for i := 0; i < l; i++ {
							ps = EvalBlockInj(ps, JsonToRye(list.Data[i]), true)
							newl[i] = ps.Res
							ps.Ser.Reset()
						}
						ps.Ser = ser
					case env.Builtin:
						for i := 0; i < l; i++ {
							newl[i] = DirectlyCallBuiltin(ps, block, JsonToRye(list.Data[i]), nil)
						}
					}
					return *env.NewList(newl)
				}
			}
			return nil
		},
	},

	// map should at the end map over block, raw-map, etc ...
	// it should accept a block of code, a function and a builtin
	// it should use injected block so it doesn't need a variable defined like map [ 1 2 3 ] x [ add a 100 ]
	// reduce [ 1 2 3 ] 'acc { + acc }
	"reduce": {
		Argsn: 3,
		Doc:   "Reduces values of a block to a new block by evaluating a block of code ...",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch list := arg0.(type) {
			case env.Block:
				switch accu := arg1.(type) {
				case env.Tagword:
					// ps.Ctx.Set(accu.Index)
					switch block := arg2.(type) {
					case env.Block, env.Builtin:
						l := len(list.Series.S)
						acc := list.Series.Get(0)
						switch block := block.(type) {
						case env.Block:
							ser := ps.Ser
							ps.Ser = block.Series
							for i := 1; i < l; i++ {
								ps.Ctx.Set(accu.Index, acc)
								ps = EvalBlockInj(ps, list.Series.Get(i), true)
								if ps.ErrorFlag {
									return ps.Res
								}
								acc = ps.Res
								ps.Ser.Reset()
							}
							ps.Ser = ser
						case env.Builtin:
							// TODO
							for i := 1; i < l; i++ {
								acc = DirectlyCallBuiltin(ps, block, acc, list.Series.Get(i))
							}
						}
						return acc
					}
				}
			case env.List:
				switch accu := arg1.(type) {
				case env.Tagword:
					// ps.Ctx.Set(accu.Index)
					switch block := arg2.(type) {
					case env.Block, env.Builtin:
						l := len(list.Data)
						acc := JsonToRye(list.Data[0])
						switch block := block.(type) {
						case env.Block:
							ser := ps.Ser
							ps.Ser = block.Series
							for i := 1; i < l; i++ {
								ps.Ctx.Set(accu.Index, acc)
								ps = EvalBlockInj(ps, JsonToRye(list.Data[i]), true)
								if ps.ErrorFlag {
									return ps.Res
								}
								acc = ps.Res
								ps.Ser.Reset()
							}
							ps.Ser = ser
						case env.Builtin:
							// TODO
							for i := 1; i < l; i++ {
								acc = DirectlyCallBuiltin(ps, block, acc, JsonToRye(list.Data[i]))
							}
						}
						return acc
					}
				}
			case env.String:
				switch accu := arg1.(type) {
				case env.Tagword:
					switch block := arg2.(type) {
					case env.Block, env.Builtin:
						input := []rune(list.Value)
						var acc env.Object
						acc = env.String{string(input[0])}
						switch block := block.(type) {
						case env.Block:
							ser := ps.Ser
							ps.Ser = block.Series
							for i := 1; i < len(input); i++ {
								ps.Ctx.Set(accu.Index, acc)
								ps = EvalBlockInj(ps, env.String{string(input[i])}, true)
								if ps.ErrorFlag {
									return ps.Res
								}
								acc = ps.Res
								ps.Ser.Reset()
							}
							ps.Ser = ser
						case env.Builtin:
						}
						return acc
					}
				}
			}
			return nil
		},
	},

	// map should at the end map over block, raw-map, etc ...
	// it should accept a block of code, a function and a builtin
	// it should use injected block so it doesn't need a variable defined like map [ 1 2 3 ] x [ add a 100 ]
	// reduce [ 1 2 3 ] 'acc { + acc }
	"fold": {
		Argsn: 4,
		Doc:   "Reduces values of a block to a new block by evaluating a block of code ...",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch list := arg0.(type) {
			case env.Block:
				switch accu := arg1.(type) {
				case env.Tagword:
					// ps.Ctx.Set(accu.Index)
					switch block := arg3.(type) {
					case env.Block, env.Builtin:
						l := len(list.Series.S)
						acc := arg2
						switch block := block.(type) {
						case env.Block:
							ser := ps.Ser
							ps.Ser = block.Series
							for i := 0; i < l; i++ {
								ps.Ctx.Set(accu.Index, acc)
								ps = EvalBlockInj(ps, list.Series.Get(i), true)
								if ps.ErrorFlag {
									return ps.Res
								}
								acc = ps.Res
								ps.Ser.Reset()
							}
							ps.Ser = ser
						case env.Builtin:
							// TODO
							for i := 1; i < l; i++ {
								acc = DirectlyCallBuiltin(ps, block, acc, list.Series.Get(i))
							}
						}
						return acc
					}
				}
			case env.List:
				switch accu := arg1.(type) {
				case env.Tagword:
					// ps.Ctx.Set(accu.Index)
					switch block := arg3.(type) {
					case env.Block, env.Builtin:
						l := len(list.Data)
						acc := arg2
						switch block := block.(type) {
						case env.Block:
							ser := ps.Ser
							ps.Ser = block.Series
							for i := 0; i < l; i++ {
								ps.Ctx.Set(accu.Index, acc)
								ps = EvalBlockInj(ps, JsonToRye(list.Data[i]), true)
								if ps.ErrorFlag {
									return ps.Res
								}
								acc = ps.Res
								ps.Ser.Reset()
							}
							ps.Ser = ser
						case env.Builtin:
							// TODO
							for i := 1; i < l; i++ {
								acc = DirectlyCallBuiltin(ps, block, acc, JsonToRye(list.Data[i]))
							}
						}
						return acc
					}
				}
			case env.String:
				switch accu := arg1.(type) {
				case env.Tagword:
					switch block := arg3.(type) {
					case env.Block, env.Builtin:
						input := []rune(list.Value)
						var acc env.Object
						acc = arg2
						switch block := block.(type) {
						case env.Block:
							ser := ps.Ser
							ps.Ser = block.Series
							for i := 0; i < len(input); i++ {
								ps.Ctx.Set(accu.Index, acc)
								ps = EvalBlockInj(ps, env.String{string(input[i])}, true)
								if ps.ErrorFlag {
									return ps.Res
								}
								acc = ps.Res
								ps.Ser.Reset()
							}
							ps.Ser = ser
						case env.Builtin:
						}
						return acc
					}
				}
			}
			return nil
		},
	},

	// map should at the end map over block, raw-map, etc ...
	// it should accept a block of code, a function and a builtin
	// it should use injected block so it doesn't need a variable defined like map [ 1 2 3 ] x [ add a 100 ]
	// reduce [ 1 2 3 ] 'acc { + acc }
	"sum-up": {
		Argsn: 2,
		Doc:   "Reduces values of a block to a new block by evaluating a block of code ...",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch list := arg0.(type) {
			case env.Block:
				switch block := arg1.(type) {
				case env.Block, env.Builtin:
					l := len(list.Series.S)
					var acc env.Object
					switch block := block.(type) {
					case env.Block:
						ser := ps.Ser
						ps.Ser = block.Series
						for i := 0; i < l; i++ {
							// ps.Ctx.Set(accu.Index, acc)
							ps = EvalBlockInj(ps, list.Series.Get(i), true)
							if ps.ErrorFlag {
								return ps.Res
							}
							switch res := ps.Res.(type) {
							case env.Integer:
								if acc == nil {
									acc = env.Integer{0}
								}
								switch acc_ := acc.(type) {
								case env.Integer:
									acc_.Value = acc_.Value + res.Value
									acc = acc_
								}
							}
							ps.Ser.Reset()
						}
						ps.Ser = ser
					case env.Builtin:
						// TODO
						for i := 1; i < l; i++ {
							acc = DirectlyCallBuiltin(ps, block, acc, list.Series.Get(i))
						}
					}
					return acc
				}
				return makeError(ps, "A2 not block")
			}
			return makeError(ps, "A1 not block")
		},
	},

	// map should at the end map over block, raw-map, etc ...
	// it should accept a block of code, a function and a builtin
	// it should use injected block so it doesn't need a variable defined like map [ 1 2 3 ] x [ add a 100 ]
	// map [ 1 2 3 ] { .add 3 }
	"partition": {
		Argsn: 2,
		Doc:   "Maps values of a block to a new block by evaluating a block of code.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch list := arg0.(type) {
			case env.Block:
				switch block := arg1.(type) {
				case env.Block, env.Builtin:
					l := list.Series.Len()
					newl := make([]env.Object, 0)
					subl := make([]env.Object, 0)
					var prevres env.Object
					switch block := block.(type) {
					case env.Block:
						ser := ps.Ser
						ps.Ser = block.Series
						for i := 0; i < l; i++ {
							curval := list.Series.Get(i)
							ps = EvalBlockInj(ps, curval, true)
							if ps.ErrorFlag {
								return ps.Res
							}
							if prevres == nil || equalValues(ps, ps.Res, prevres) {
								subl = append(subl, curval)
							} else {
								newl = append(newl, env.NewBlock(*env.NewTSeries(subl)))
								subl = make([]env.Object, 1)
								subl[0] = curval
							}
							prevres = ps.Res
							ps.Ser.Reset()
						}
						newl = append(newl, env.NewBlock(*env.NewTSeries(subl)))
						ps.Ser = ser
					case env.Builtin:
						for i := 0; i < l; i++ {
							newl[i] = DirectlyCallBuiltin(ps, block, list.Series.Get(i), nil)
						}
					}
					return *env.NewBlock(*env.NewTSeries(newl))
				}
			case env.String:
				switch block := arg1.(type) {
				case env.Block, env.Builtin:
					newl := make([]interface{}, 0)
					var subl strings.Builder
					var prevres env.Object
					switch block := block.(type) {
					case env.Block:
						ser := ps.Ser
						ps.Ser = block.Series
						for _, curval := range list.Value {
							ps = EvalBlockInj(ps, env.String{string(curval)}, true)
							if ps.ErrorFlag {
								return ps.Res
							}
							if prevres == nil || equalValues(ps, ps.Res, prevres) {
								subl.WriteRune(curval)
							} else {
								newl = append(newl, subl.String())
								subl.Reset()
								subl.WriteRune(curval)
							}
							prevres = ps.Res
							ps.Ser.Reset()
						}
						newl = append(newl, subl.String())
						ps.Ser = ser
					case env.Builtin:
					}
					return *env.NewList(newl)
				}
			}
			return nil
		},
	},

	"group": {
		Argsn: 2,
		Doc:   "",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch list := arg0.(type) {
			case env.Block:
				switch block := arg1.(type) {
				case env.Block, env.Builtin:
					l := list.Series.Len()
					newd := make(map[string]interface{})
					switch block := block.(type) {
					case env.Block:
						ser := ps.Ser
						ps.Ser = block.Series
						for i := 0; i < l; i++ {
							curval := list.Series.Get(i)
							ps = EvalBlockInj(ps, curval, true)
							if ps.ErrorFlag {
								return ps.Res
							}
							// TODO !!! -- currently only works if results are keys
							newkey := ps.Res.(env.String).Value
							entry, ok := newd[newkey]
							if !ok {
								newd[newkey] = env.NewList(make([]interface{}, 0))
								entry, ok = newd[newkey]
							}
							switch ee := entry.(type) {
							case *env.List:
								ee.Data = append(ee.Data, curval)
							default:
								return makeError(ps, "FAILURE TODO")
							}
							ps.Ser.Reset()
						}
						ps.Ser = ser
						return *env.NewDict(newd)
					}
				}
			}
			return nil
		},
	},

	// filter [ 1 2 3 ] { .add 3 }
	"filter": {
		Argsn: 2,
		Doc:   "Filters values from a seris based on return of a injected code block.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			var ll []interface{}
			var lo []env.Object
			var llen int
			modeObj := 0
			switch data := arg0.(type) {
			case env.Block:
				lo = data.Series.S
				llen = len(lo)
				modeObj = 2
			case env.List:
				ll = data.Data
				llen = len(ll)
				modeObj = 1
			}

			if modeObj == 0 {
				ps.FailureFlag = true
				return env.NewError("expects list or block")
			}

			switch block := arg1.(type) {
			case env.Block, env.Builtin:
				var newl []env.Object
				switch block := block.(type) {
				case env.Block:
					ser := ps.Ser
					ps.Ser = block.Series
					for i := 0; i < llen; i++ {
						var item interface{}
						if modeObj == 1 {
							item = ll[i]
						} else {
							item = lo[i]
						}
						ps = EvalBlockInj(ps, JsonToRye(item), true)
						if util.IsTruthy(ps.Res) { // todo -- move these to util or something
							newl = append(newl, JsonToRye(item))
						}
						ps.Ser.Reset()
					}
					ps.Ser = ser
				case env.Builtin:
					for i := 0; i < llen; i++ {
						var item interface{}
						if modeObj == 1 {
							item = ll[i]
						} else {
							item = lo[i]
						}
						res := DirectlyCallBuiltin(ps, block, JsonToRye(item), nil)
						if util.IsTruthy(res) { // todo -- move these to util or something
							newl = append(newl, JsonToRye(item))
						}
					}
				}
				return *env.NewBlock(*env.NewTSeries(newl))
			}
			return nil
		},
	},

	"seek": {
		Argsn: 2,
		Doc:   "Seek over a series until a Block of code returns True.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch list := arg0.(type) {
			case env.Block:
				switch block := arg1.(type) {
				case env.Block, env.Builtin:
					l := list.Series.Len()
					switch block := block.(type) {
					case env.Block:
						ser := ps.Ser
						ps.Ser = block.Series
						for i := 0; i < l; i++ {
							ps = EvalBlockInj(ps, list.Series.Get(i), true)
							if util.IsTruthy(ps.Res) { // todo -- move these to util or something
								return list.Series.Get(i)
							}
							ps.Ser.Reset()
						}
						ps.Ser = ser
					case env.Builtin:
						for i := 0; i < l; i++ {
							res := DirectlyCallBuiltin(ps, block, list.Series.Get(i), nil)
							if util.IsTruthy(res) { // todo -- move these to util or something
								return list.Series.Get(i)
							}
						}
					default:
						ps.ErrorFlag = true
						return env.NewError("Second argument should be block, builtin (or function).")
					}
				}
			}
			return nil
		},
	},

	// collections exploration functions

	"max": {
		Argsn: 1,
		Doc:   "Accepts a block of values and returns maximal value.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			var max env.Object
			switch block := arg0.(type) {
			case env.Block:
				l := block.Series.Len()
				for i := 0; i < l; i++ {
					if max == nil || greaterThan(ps, block.Series.Get(i), max) {
						max = block.Series.Get(i)
					}
				}
			}
			return max
		},
	},

	"min": {
		Argsn: 1,
		Doc:   "Accepts a block of values and returns maximal value.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			var max env.Object
			switch block := arg0.(type) {
			case env.Block:
				l := block.Series.Len()
				for i := 0; i < l; i++ {
					if max == nil || greaterThan(ps, max, block.Series.Get(i)) {
						max = block.Series.Get(i)
					}
				}
			}
			return max
		},
	},

	"avg": {
		Argsn: 1,
		Doc:   "Accepts a block of values and returns maximal value.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			var sum float64
			switch block := arg0.(type) {
			case env.Block:
				l := block.Series.Len()
				for i := 0; i < l; i++ {
					obj := block.Series.Get(i)
					switch val1 := obj.(type) {
					case env.Integer:
						sum += float64(val1.Value)
					case env.Decimal:
						sum += val1.Value
					}
				}
				return env.Decimal{sum / float64(l)}
			}
			return nil
		},
	},

	"sum": {
		Argsn: 1,
		Doc:   "Accepts a block of values and returns maximal value.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			var sum float64
			switch block := arg0.(type) {
			case env.Block:
				l := block.Series.Len()
				for i := 0; i < l; i++ {
					obj := block.Series.Get(i)
					switch val1 := obj.(type) {
					case env.Integer:
						sum += float64(val1.Value)
					case env.Decimal:
						sum += val1.Value
					}
				}
				return env.Decimal{sum}
			}
			return nil
		},
	},

	"sort!": {
		Argsn: 1,
		Doc:   "Accepts a block of values and returns maximal value.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch block := arg0.(type) {
			case env.Block:
				ss := block.Series.S
				sort.Sort(RyeBlockSort(ss))
				return *env.NewBlock(*env.NewTSeries(ss))
			}
			return nil
		},
	},

	"reverse!": {
		Argsn: 1,
		Doc:   "Accepts a block of values and returns maximal value.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch block := arg0.(type) {
			case env.Block:
				a := block.Series.S
				for left, right := 0, len(a)-1; left < right; left, right = left+1, right-1 {
					a[left], a[right] = a[right], a[left]
				}
				// sort.Sort(RyeBlockSort(ss))
				return *env.NewBlock(*env.NewTSeries(a))
			}
			return nil
		},
	},

	// add distinct? and count? functions
	// make functions work with list, which column and row can return

	// end of collections exploration

	//test if we can do recur similar to clojure one. Since functions in rejy are of fixed arity we would need recur1 recur2 recur3 and recur [ ] which is less optimal
	//otherwise word recur could somehow be bound to correct version or args depending on number of args of func. Try this at first.
	"recur-if\\1": { //recur1-if
		Argsn: 2,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch cond := arg0.(type) {
			case env.Integer:
				if cond.Value > 0 {
					switch arg := arg1.(type) {
					case env.Integer:
						ps.Ctx.Set(ps.Args[0], arg)
						ps.Ser.Reset()
						return nil
					}
				} else {
					return ps.Res
				}
			}
			return nil
		},
	},

	"recur-if\\2": { //recur1-if
		Argsn: 3,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			//arg0.Trace("a0")
			//arg1.Trace("a1")
			//arg2.Trace("a2")
			switch cond := arg0.(type) {
			case env.Integer:
				if cond.Value > 0 {
					switch argi1 := arg1.(type) {
					case env.Integer:
						switch argi2 := arg2.(type) {
						case env.Integer:
							ps.Ctx.Set(ps.Args[0], argi1)
							ps.Ctx.Set(ps.Args[1], argi2)
							ps.Ser.Reset()
							return ps.Res
						}
					}
				} else {
					return ps.Res
				}
			}
			return nil
		},
	},

	"recur-if\\3": { //recur1-if
		Argsn: 4,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			//arg0.Trace("a0")
			//arg1.Trace("a1")
			//arg2.Trace("a2")
			switch cond := arg0.(type) {
			case env.Integer:
				if cond.Value > 0 {
					switch argi1 := arg1.(type) {
					case env.Integer:
						switch argi2 := arg2.(type) {
						case env.Integer:
							switch argi3 := arg3.(type) {
							case env.Integer:
								ps.Ctx.Set(ps.Args[0], argi1)
								ps.Ctx.Set(ps.Args[1], argi2)
								ps.Ctx.Set(ps.Args[2], argi3)
								ps.Ser.Reset()
								return ps.Res
							}
						}
					}
				} else {
					return ps.Res
				}
			}
			return nil
		},
	},

	"does": {
		Argsn: 1,
		Doc:   "Creates a function without arguments.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch body := arg0.(type) {
			case env.Block:
				//spec := []env.Object{env.Word{aaaidx}}
				//body := []env.Object{env.Word{printidx}, env.Word{aaaidx}, env.Word{recuridx}, env.Word{greateridx}, env.Integer{99}, env.Word{aaaidx}, env.Word{incidx}, env.Word{aaaidx}}
				return *env.NewFunction(*env.NewBlock(*env.NewTSeries(make([]env.Object, 0))), body, false)
			}
			return nil
		},
	},

	"fn1": {
		Argsn: 1,
		Doc:   "Creates a function that accepts one anonymouse argument.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch body := arg0.(type) {
			case env.Block:
				spec := []env.Object{env.Word{1}}
				//body := []env.Object{env.Word{printidx}, env.Word{aaaidx}, env.Word{recuridx}, env.Word{greateridx}, env.Integer{99}, env.Word{aaaidx}, env.Word{incidx}, env.Word{aaaidx}}
				return *env.NewFunction(*env.NewBlock(*env.NewTSeries(spec)), body, false)
			}
			return nil
		},
	},

	"fn": {
		Argsn: 2,
		Doc:   "Creates a function.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch args := arg0.(type) {
			case env.Block:
				switch body := arg1.(type) {
				case env.Block:
					//spec := []env.Object{env.Word{aaaidx}}
					//body := []env.Object{env.Word{printidx}, env.Word{aaaidx}, env.Word{recuridx}, env.Word{greateridx}, env.Integer{99}, env.Word{aaaidx}, env.Word{incidx}, env.Word{aaaidx}}
					return *env.NewFunction(args, body, false)
				}
			}
			return nil
		},
	},

	"pfn": {
		Argsn: 2,
		Doc:   "Creates a pure function.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch args := arg0.(type) {
			case env.Block:
				switch body := arg1.(type) {
				case env.Block:
					return *env.NewFunction(args, body, true)
				}
			}
			return nil
		},
	},

	"fnc": {
		// a function with context	 bb: 10 add10 [ a ] context [ b: bb ] [ add a b ]
		// 							add10 [ a ] this [ add a b ]
		// later maybe			   add10 [ a ] [ b: b ] [ add a b ]
		//  						   add10 [ a ] [ 'b ] [ add a b ]
		Argsn: 3,
		Doc:   "Creates a function with specific context.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch args := arg0.(type) {
			case env.Block:
				switch ctx := arg1.(type) {
				case env.RyeCtx:
					switch body := arg2.(type) {
					case env.Block:
						return *env.NewFunctionC(args, body, &ctx, false)
					default:
						ps.ErrorFlag = true
						return env.NewError("Third arg should be Block")
					}
				default:
					ps.ErrorFlag = true
					return env.NewError("Second arg should be Context")
				}
			default:
				ps.ErrorFlag = true
				return env.NewError("First argument should be Block")
			}
			return nil
		},
	},

	"kind": {
		Argsn: 2,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Tagword:
				switch spec := arg1.(type) {
				case env.Block:
					return *env.NewKind(s1.ToWord(), spec)
				default:
					return env.NewError("2nd not block")
				}
			default:
				return env.NewError("first not lit-word")
			}
			return nil
		},
	},

	"_>>": {
		Argsn: 2,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch spec := arg1.(type) {
			case env.Kind:
				switch dict := arg0.(type) {
				case env.Dict:
					obj := BuiValidate(ps, dict, spec.Spec)
					switch obj1 := obj.(type) {
					case env.Dict:
						ctx := util.Dict2Context(ps, obj1)
						ctx.Kind = spec.Kind
						return ctx
					default:
						return obj
					}
				case env.RyeCtx:
					if spec.HasConverter(dict.Kind.Index) {
						obj := BuiConvert(ps, dict, spec.Converters[dict.Kind.Index])
						switch ctx := obj.(type) {
						case env.RyeCtx:
							ctx.Kind = spec.Kind
							return ctx
						default:
							return *env.NewError("2344nd xxxx   xxx sn't Dict")
						}
					}
					return *env.NewError("2nd xxxx   xxx sn't Dict")
				default:
					return *env.NewError("2nd isn't Dict")
				}
			default:
				return *env.NewError("1st isn't kind")
			}
			return *env.NewError("1st isn't kind xxxx")
		},
	},

	"_<<": {
		Argsn: 2,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch spec := arg0.(type) {
			case env.Kind:
				switch dict := arg1.(type) {
				case env.Dict:
					obj := BuiValidate(ps, dict, spec.Spec)
					switch obj1 := obj.(type) {
					case env.Dict:
						ctx := util.Dict2Context(ps, obj1)
						ctx.Kind = spec.Kind
						return ctx
					default:
						return env.NewError("2nd A isn't ")
					}
				case env.RyeCtx:
					if spec.HasConverter(dict.Kind.Index) {
						obj := BuiConvert(ps, dict, spec.Converters[dict.Kind.Index])
						switch ctx := obj.(type) {
						case env.RyeCtx:
							ctx.Kind = spec.Kind
							return ctx
						default:
							return env.NewError("2344nd xxxx   xxx sn't Dict")
						}
					}
				default:
					return env.NewError("2nd isn't Dict")
				}
			default:
				return env.NewError("1st isn't kind")
			}
			return nil
		},
	},

	"assure-kind": {
		Argsn: 2,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch spec := arg1.(type) {
			case env.Kind:
				switch dict := arg0.(type) {
				case env.Dict:
					obj := BuiValidate(ps, dict, spec.Spec)
					switch obj1 := obj.(type) {
					case env.Dict:
						ctx := util.Dict2Context(ps, obj1)
						ctx.Kind = spec.Kind
						return ctx
					default:
						return env.NewError("2nd A isn't ")
					}
				default:
					return env.NewError("2nd isn't Dict")
				}
			default:
				return env.NewError("1st isn't kind")
			}
			return nil
		},
	},

	// BASIC STRING FUNCTIONS

	"left": {
		Argsn: 2,
		Doc:   "Returns the left N characters of the String.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.String:
				switch s2 := arg1.(type) {
				case env.Integer:
					return env.String{s1.Value[0:s2.Value]}
				}
			}
			return nil
		},
	},

	"newline": {
		Argsn: 0,
		Doc:   "Returns the newline character.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return env.String{"\n"}
		},
	},

	"trim": {
		Argsn: 1,
		Doc:   "Trims the String of spacing characters.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.String:
				return env.String{strings.TrimSpace(s1.Value)}
			}
			return nil
		},
	},

	"replace": {
		Argsn: 3,
		Doc:   "Returns the string with all parts of the strings replaced.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.String:
				switch s2 := arg1.(type) {
				case env.String:
					switch s3 := arg2.(type) {
					case env.String:
						return env.String{strings.ReplaceAll(s1.Value, s2.Value, s3.Value)}
					}
				}
			}
			return nil
		},
	},

	"substring": {
		Argsn: 3,
		Doc:   "Returns part of the String between two positions.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.String:
				switch s2 := arg1.(type) {
				case env.Integer:
					switch s3 := arg2.(type) {
					case env.Integer:
						return env.String{s1.Value[s2.Value:s3.Value]}
					}
				}
			}
			return nil
		},
	},

	"contains": {
		Argsn: 2,
		Doc:   "Returns part of the String between two positions.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.String:
				switch s2 := arg1.(type) {
				case env.String:
					if strings.Contains(s1.Value, s2.Value) {
						return env.Integer{1}
					} else {
						return env.Integer{0}
					}
				}
			}
			return nil
		},
	},

	"index?": {
		Argsn: 2,
		Doc:   "Returns part of the String between two positions.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.String:
				switch s2 := arg1.(type) {
				case env.String:
					res := strings.Index(s2.Value, s1.Value)
					return env.Integer{int64(res)}
				}
			}
			return nil
		},
	},

	"position?": {
		Argsn: 2,
		Doc:   "Returns part of the String between two positions.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.String:
				switch s2 := arg1.(type) {
				case env.String:
					res := strings.Index(s2.Value, s1.Value)
					return env.Integer{int64(res + 1)}
				}
			}
			return nil
		},
	},

	"right": {
		Argsn: 2,
		Doc:   "Returns the N characters from the right of the String.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.String:
				switch s2 := arg1.(type) {
				case env.Integer:
					return env.String{s1.Value[len(s1.Value)-int(s2.Value):]}
				}
			}
			return nil
		},
	},

	"concat": {
		Argsn: 2,
		Doc:   "Joins two strings together.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Integer:
				switch s2 := arg1.(type) {
				case env.String:
					return env.String{strconv.Itoa(int(s1.Value)) + s2.Value}
				case env.Integer:
					return env.String{strconv.Itoa(int(s1.Value)) + strconv.Itoa(int(s2.Value))}
				}
			case env.String:
				switch s2 := arg1.(type) {
				case env.String:
					return env.String{s1.Value + s2.Value}
				case env.Integer:
					return env.String{s1.Value + strconv.Itoa(int(s2.Value))}
				}
			case env.Block:
				switch b2 := arg1.(type) {
				case env.Block:
					s := &s1.Series
					s1.Series = *s.AppendMul(b2.Series.GetAll())
					return s1
				case env.Object:
					s := &s1.Series
					s1.Series = *s.Append(b2)
					return s1
				}
			}
			return nil
		},
	},

	"intersect": {
		Argsn: 2,
		Doc:   "Finds the intersection of two values.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.String:
				switch s2 := arg1.(type) {
				case env.String:
					inter := util.IntersectStrings(s1.Value, s2.Value)
					return env.String{inter}
				default:
					return makeError(ps, "Arg 2 not String")
				}
			case env.Block:
				switch b2 := arg1.(type) {
				case env.Block:
					inter := util.IntersectLists(ps, s1.Series.S, b2.Series.S)
					return *env.NewBlock(*env.NewTSeries(inter))
				default:
					return makeError(ps, "Arg 2 not Block")
				}
			}
			return makeError(ps, "Arg 1 not Block or String")
		},
	},

	"str": {
		Argsn: 1,
		Doc:   "Turn Rye value to String.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Integer:
				return env.String{strconv.Itoa(int(s1.Value))}
			}
			return nil
		},
	},
	"capitalize": {
		Argsn: 1,
		Pure:  true,
		Doc:   "Takes a String and returns the same String, but  with first character turned to upper case.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.String:
				return env.String{strings.Title(s1.Value)}
			default:
				return env.NewError("first arg must be string")
			}
		},
	},
	"to-lower": {
		Argsn: 1,
		Pure:  true,
		Doc:   "Takes a String and returns the same String, but  with all characters turned to lower case.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.String:
				return env.String{strings.ToLower(s1.Value)}
			default:
				return env.NewError("first arg must be string")
			}
		},
	},
	"to-upper": {
		Argsn: 1,
		Pure:  true,
		Doc:   "Takes a String and returns the same String, but with all characters turned to upper case.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.String:
				return env.String{strings.ToUpper(s1.Value)}
			default:
				return env.NewError("first arg must be string")
			}
		},
	},

	"concat3": {
		Argsn: 3,
		Pure:  true,
		Doc:   "Joins 3 Rye values together.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.String:
				switch s2 := arg1.(type) {
				case env.String:
					switch s3 := arg2.(type) {
					case env.String:
						return env.String{s1.Value + s2.Value + s3.Value}
					}
				}
			}
			return nil
		},
	},

	"join": { // todo -- join\w data ","
		Argsn: 1,
		Pure:  true,
		Doc:   "Joins Block or list of values together.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.List:
				var str strings.Builder
				for _, c := range s1.Data {
					switch it := c.(type) {
					case string:
						str.WriteString(it)
					case env.String:
						str.WriteString(it.Value)
					case int:
						str.WriteString(strconv.Itoa(it))
					case env.Integer:
						str.WriteString(strconv.Itoa(int(it.Value)))
					}
				}
				return env.String{str.String()}
			case env.Block:
				var str strings.Builder
				for _, c := range s1.Series.S {
					switch it := c.(type) {
					case env.String:
						str.WriteString(it.Value)
					case env.Integer:
						str.WriteString(strconv.Itoa(int(it.Value)))
					}
				}
				return env.String{str.String()}
			}
			return nil
		},
	},

	"join-with": { // todo -- join\w data ","
		Argsn: 2,
		Pure:  true,
		Doc:   "Joins Block or list of values together.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.List:
				switch s2 := arg1.(type) {
				case env.String:
					var str strings.Builder
					for _, c := range s1.Data {
						switch it := c.(type) {
						case string:
							str.WriteString(it)
						case env.String:
							str.WriteString(it.Value)
						case int:
							str.WriteString(strconv.Itoa(it))
						case env.Integer:
							str.WriteString(strconv.Itoa(int(it.Value)))
						}
						str.WriteString(s2.Value)
					}
					return env.String{str.String()}
				}
			case env.Block:
				switch s2 := arg1.(type) {
				case env.String:
					var str strings.Builder
					for _, c := range s1.Series.S {
						switch it := c.(type) {
						case env.String:
							str.WriteString(it.Value)
						case env.Integer:
							str.WriteString(strconv.Itoa(int(it.Value)))
						}
						str.WriteString(s2.Value)
					}
					return env.String{str.String()}
				}
			}
			return nil
		},
	},

	"split-quoted": {
		Argsn: 3,
		Pure:  true,
		Doc:   "Splits a line of string into values by separator by respecting quotes",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch str := arg0.(type) {
			case env.String:
				switch sepa := arg1.(type) {
				case env.String:
					switch quote := arg2.(type) {
					case env.String:
						return util.StringToFieldsWithQuoted(str.Value, sepa.Value, quote.Value)
					default:
						return makeError(ps, "Quote character not a string.")
					}
				default:
					return makeError(ps, "Separator character not a string.")
				}
			default:
				return makeError(ps, "Input text not a string.")
			}
		},
	},

	"split": {
		Argsn: 2,
		Pure:  true,
		Doc:   "Splits a string into a block of values using a separator",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch str := arg0.(type) {
			case env.String:
				switch sepa := arg1.(type) {
				case env.String:
					spl := strings.Split(str.Value, sepa.Value) // util.StringToFieldsWithQuoted(str.Value, sepa.Value, quote.Value)
					spl2 := make([]env.Object, len(spl))
					for i, val := range spl {
						spl2[i] = env.String{val}
					}
					return *env.NewBlock(*env.NewTSeries(spl2))
				default:
					return makeError(ps, "Separator character not a string.")
				}
			default:
				return makeError(ps, "Input text not a string.")
			}
		},
	},

	"split\\many": {
		Argsn: 2,
		Pure:  true,
		Doc:   "Splits a string into a block of values using a separator",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch str := arg0.(type) {
			case env.String:
				switch sepa := arg1.(type) {
				case env.String:
					spl := util.SplitMulti(str.Value, sepa.Value) // util.StringToFieldsWithQuoted(str.Value, sepa.Value, quote.Value)
					spl2 := make([]env.Object, len(spl))
					for i, val := range spl {
						spl2[i] = env.String{val}
					}
					return *env.NewBlock(*env.NewTSeries(spl2))
				default:
					return makeError(ps, "Separator character not a string.")
				}
			default:
				return makeError(ps, "Input text not a string.")
			}
		},
	},

	"split-every": {
		Argsn: 2,
		Pure:  true,
		Doc:   "Splits a string into a block of values using a separator",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch str := arg0.(type) {
			case env.String:
				switch sepa := arg1.(type) {
				case env.Integer:
					spl := util.SplitEveryString(str.Value, int(sepa.Value))
					spl2 := make([]env.Object, len(spl))
					for i, val := range spl {
						spl2[i] = env.String{val}
					}
					return *env.NewBlock(*env.NewTSeries(spl2))
				default:
					return makeError(ps, "Separator character not a string.")
				}
			case env.Block:
				switch sepa := arg1.(type) {
				case env.Integer:
					spl := util.SplitEveryList(str.Series.S, int(sepa.Value))
					spl2 := make([]env.Object, len(spl))
					for i, val := range spl {
						spl2[i] = *env.NewBlock(*env.NewTSeries(val))
					}
					return *env.NewBlock(*env.NewTSeries(spl2))
				default:
					return makeError(ps, "Separator character not a string.")
				}
			default:
				return makeError(ps, "Input text not a string.")
			}
		},
	},

	"to-integer": {
		Argsn: 1,
		Doc:   "Splits a line of string into values by separator by respecting quotes",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch addr := arg0.(type) {
			case env.String:
				int, err := strconv.Atoi(addr.Value)
				if err != nil {
					return makeError(ps, err.Error())
				}
				return env.Integer{int64(int)}
			default:
				return makeError(ps, "Arg 1 should be String.")
			}
		},
	},

	// BASIC SERIES FUNCTIONS

	"first": {
		Argsn: 1,
		Doc:   "Accepts Block, List or String and returns the first item.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Block:
				if len(s1.Series.S) == 0 {
					return makeError(ps, "Block is empty.")
				}
				return s1.Series.Get(int(0))
			case env.List:
				if len(s1.Data) == 0 {
					return makeError(ps, "List is empty.")
				}
				return JsonToRye(s1.Data[int(0)])
			case env.String:
				str := []rune(s1.Value)
				if len(str) == 0 {
					return makeError(ps, "String is empty.")
				}
				return env.String{string(str[0])}
			case env.Spreadsheet:
				return s1.GetRow(ps, int(0))
			default:
				return env.NewError("Arg 1 not a Series.")
			}
		},
	},

	"rest": {
		Argsn: 1,
		Doc:   "Accepts Block, List or String and returns all but first items.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Block:
				if len(s1.Series.S) == 0 {
					return makeError(ps, "Block is empty.")
				}
				return *env.NewBlock(*env.NewTSeries(s1.Series.S[1:]))
			case env.List:
				if len(s1.Data) == 0 {
					return makeError(ps, "List is empty.")
				}
				return env.NewList(s1.Data[int(1):])
			case env.String:
				str := []rune(s1.Value)
				if len(str) < 1 {
					return makeError(ps, "String has only one element.")
				}
				return env.String{string(str[1:])}
			default:
				return env.NewError("Arg 1 not a Series.")
			}
		},
	},
	"tail": {
		Argsn: 2,
		Doc:   "Accepts Block, List or String and returns all but first items.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch num := arg1.(type) {
			case env.Integer:
				switch s1 := arg0.(type) {
				case env.Block:
					if len(s1.Series.S) == 0 {
						return makeError(ps, "Block is empty.")
					}
					return *env.NewBlock(*env.NewTSeries(s1.Series.S[len(s1.Series.S)-int(num.Value):]))
				case env.List:
					if len(s1.Data) == 0 {
						return makeError(ps, "List is empty.")
					}
					return env.NewList(s1.Data[len(s1.Data)-int(num.Value):])
				case env.String:
					str := []rune(s1.Value)
					if len(str) < 1 {
						return makeError(ps, "String has only one element.")
					}
					return env.String{string(str[len(str)-int(num.Value):])}
				default:
					return env.NewError("Arg 1 not a Series.")
				}
			default:
				return env.NewError("Arg 2 not a Integer.")
			}
		},
	},
	"second": {
		Argsn: 1,
		Doc:   "Accepts Block and returns the second value in it.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Block:
				if len(s1.Series.S) < 2 {
					return makeError(ps, "Block has no second element.")
				}
				return s1.Series.Get(1)
			case env.List:
				if len(s1.Data) < 2 {
					return makeError(ps, "List has no second element.")
				}
				return JsonToRye(s1.Data[1])
			case env.String:
				str := []rune(s1.Value)
				if len(str) < 2 {
					return makeError(ps, "String has no second element.")
				}
				return env.String{string(str[1])}
			default:
				return env.NewError("Arg 1 not a Series.")
			}
		},
	},
	"third": {
		Argsn: 1,
		Doc:   "Accepts Block and returns the third value in it.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Block:
				return s1.Series.Get(int(2))
			}
			return nil
		},
	},
	"last": {
		Argsn: 1,
		Doc:   "Accepts Block and returns the last value in it.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Block:
				return s1.Series.Get(s1.Series.Len() - 1)
			case env.String:
				return env.String{s1.Value[len(s1.Value)-1:]}
			}
			return nil
		},
	},

	"head": {
		Argsn: 2,
		Doc:   "Accepts a Block or a List and an Integer N. Returns first N values of the Block.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s2 := arg1.(type) {
			case env.Integer:
				switch s1 := arg0.(type) {
				case env.Block:
					return *env.NewBlock(*env.NewTSeries(s1.Series.S[0:s2.Value]))
				case env.List:
					return *env.NewList(s1.Data[0:s2.Value])
				default:
					return *env.NewError("not block or list")
				}
			default:
				return *env.NewError("not integer")
			}
		},
	},

	"nth": {
		Argsn: 2,
		Doc:   "Accepts Block and Integer N, returns the N-th value of the block.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Block:
				switch s2 := arg1.(type) {
				case env.Integer:
					return s1.Series.Get(int(s2.Value - 1))
				}
			}
			return nil
		},
	},
	"peek": {
		Argsn: 1,
		Doc:   "Accepts Block and returns the current value, without removing it.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Block:
				return s1.Series.Peek()
			}
			return nil
		},
	},
	"pop": {
		Argsn: 1,
		Doc:   "Accepts Block and returns the next value and removes it from the Block.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Block:
				return s1.Series.Pop()
			}
			return nil
		},
	},
	"pos": {
		Argsn: 1,
		Doc:   "Accepts Block and returns the position of it's carret.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Block:
				return env.Integer{int64(s1.Series.Pos())}
			}
			return nil
		},
	},
	"next": {
		Argsn: 1,
		Doc:   "Accepts Block and returns the next value from it.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Block:
				s1.Series.Next()
				return s1
			}
			return nil
		},
	},
	"remove-last!": {
		Argsn: 1,
		Pure:  false,
		Doc:   "Accepts Block and returns the next value and removes it from the Block.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch wrd := arg0.(type) {
			case env.Tagword:
				val, found, ctx := ps.Ctx.Get2(wrd.Index)
				if found {
					switch oldval := val.(type) {
					case env.Block:
						s := &oldval.Series
						oldval.Series = *s.RmLast()
						ctx.Set(wrd.Index, oldval)
						return oldval
					}
				}
			}
			return nil
		},
	},
	"append!": {
		Argsn: 2,
		Doc:   "Accepts Rye value and Tagword with a Block or String. Appends Rye value to Block/String in place, also returns it	.",
		Pure:  false,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch wrd := arg1.(type) {
			case env.Tagword:
				val, found, ctx := ps.Ctx.Get2(wrd.Index)
				if found {
					switch oldval := val.(type) {
					case env.String:
						var newval env.String
						switch s3 := arg0.(type) {
						case env.String:
							newval = env.String{oldval.Value + s3.Value}
						case env.Integer:
							newval = env.String{oldval.Value + strconv.Itoa(int(s3.Value))}
						}
						ctx.Set(wrd.Index, newval)
						return newval
					case *env.Block: // TODO
						fmt.Println(123)
						s := &oldval.Series
						oldval.Series = *s.Append(arg0)
						ctx.Set(wrd.Index, oldval)
						return oldval
					default:
						return makeError(ps, "Type of tagword is not String.")
					}
				}
				return makeError(ps, "Tagword not found.")
			case env.Block:
				s := &wrd.Series
				wrd.Series = *s.Append(arg0)
				//ctx.Set(wrd.Index, oldval)
				return nil
			default:
				return makeError(ps, "Value not tagword")
			}
		},
	},
	// FUNCTIONALITY AROUND GENERIC METHODS
	// generic <integer> <add> fn [ a b ] [ a + b ] // tagwords are temporary here
	"generic": {
		Argsn: 3,
		Doc:   "Registers a generic function.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Tagword:
				switch s2 := arg1.(type) {
				case env.Tagword:
					switch s3 := arg2.(type) {
					case env.Object:
						fmt.Println(s1.Index)
						fmt.Println(s2.Index)
						fmt.Println("Generic")

						registerGeneric(ps, s1.Index, s2.Index, s3)
						return s3
					}
				}
			}
			ps.ErrorFlag = true
			return env.NewError("Wrong args when creating generic function")
		},
	},

	"table": {
		Argsn: 1,
		Doc:   "Constructs an empty table, accepts a block of column names",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch bloc := arg0.(type) {
			case env.Block:
				vv := bloc.Series.Peek()
				switch vv.(type) {
				case env.String:
					cols := make([]string, bloc.Series.Len())
					for i := 0; i < bloc.Series.Len(); i++ {
						cols[i] = bloc.Series.Get(i).(env.String).Value
					}
					return *env.NewSpreadsheet(cols)

				case env.Tagword:
					// TODO
				}
				return nil
			}
			return nil
		},
	},

	"add-row": {
		Argsn: 2,
		Doc:   "Constructs an empty table, accepts a block of column names",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch table := arg0.(type) {
			case env.Spreadsheet:
				switch bloc := arg1.(type) {
				case env.Block:
					vals := make([]interface{}, bloc.Series.Len())
					for i := 0; i < bloc.Series.Len(); i++ {
						vals[i] = bloc.Series.Get(i)
					}
					table.AddRow(env.SpreadsheetRow{vals, &table})
					return table
				}
				return nil
			}
			return nil
		},
	},

	"dict": {
		Argsn: 1,
		Doc:   "Constructs a Dict from the Block of key and value pairs.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch bloc := arg0.(type) {
			case env.Block:
				return env.NewDictFromSeries(bloc.Series, ps.Idx)
			}
			return nil
		},
	},

	// BASIC ENV / Dict FUNCTIONS
	"_->": {
		Argsn: 2,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return getFrom(ps, arg0, arg1, false)
		},
	},
	"_<-": {
		Argsn: 2,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return getFrom(ps, arg1, arg0, false)
		},
	},
	/* "_<-": {
		Argsn: 2,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Block:
				switch s2 := arg1.(type) {
				case env.Integer:
					idx := s2.Value
					//					if posMode {
					// 	idx--
					//}
					v := s1.Series.PGet(int(idx))
					ok := true
					if ok {
						return v
					} else {
						ps.FailureFlag = true
						return env.NewError1(5) // NOT_FOUND
					}
				}
				//return getFrom(ps, arg1, arg0, false)
				return nil
			}
			return nil
		},
	},*/
	"_<~": {
		Argsn: 2,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return getFrom(ps, arg1, arg0, true)
		},
	},
	"_~>": {
		Argsn: 2,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return getFrom(ps, arg0, arg1, true)
		},
	},

	// return , error , failure functions
	"return": {
		Argsn: 1,
		Doc:   "Accepts one value and returns it.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			//fmt.Println("RETURN")
			ps.ReturnFlag = true
			return arg0
		},
	},

	"^fail": {
		Argsn: 1,
		Doc:   "Returning Fail.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			//fmt.Println("FAIL")
			ps.FailureFlag = true
			ps.ReturnFlag = true
			switch val := arg0.(type) {
			case env.String: // todo .. make Error type .. make error construction micro dialect, return the error wrapping error that caused it
				return env.NewError(val.Value)
			case env.Integer: // todo .. make Error type .. make error construction micro dialect, return the error wrapping error that caused it
				return env.NewError1(int(val.Value))
			}
			return arg0
		},
	},

	"fail": {
		Argsn: 1,
		Doc:   "Constructs and Fails with an Error object. Accepts String as message, Integer as code, or block for multiple parameters.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			//fmt.Println("FAIL")
			ps.FailureFlag = true
			return MakeRyeError(ps, arg0, nil)
		},
	},

	"new-error": {
		Argsn: 1,
		Doc:   "Constructs and Error object. Accepts String as message, Integer as code, or block for multiple parameters.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			//ps.ErrorFlag = true
			return MakeRyeError(ps, arg0, nil)
		},
	},

	"code?": {
		AcceptFailure: true,
		Argsn:         1,
		Doc:           "Returns the status code of the Error.", // TODO -- seems duplicate of status
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch er := arg0.(type) {
			case env.Error:
				return env.Integer{int64(er.Status)}
			case *env.Error:
				return env.Integer{int64(er.Status)}
			default:
				ps.FailureFlag = true
				return env.NewError("arg 0 not error")
			}
		},
	},

	"disarm": {
		AcceptFailure: true,
		Argsn:         1,
		Doc:           "Disarms the Error.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			ps.FailureFlag = false
			return arg0
		},
	},

	"failed?": {
		AcceptFailure: true,
		Argsn:         1,
		Doc:           "Checks if first argument is an Error. Returns a boolean.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			ps.FailureFlag = false
			switch arg0.(type) {
			case env.Error:
				return env.Integer{int64(1)}
			case *env.Error:
				return env.Integer{int64(1)}
			}
			return env.Integer{int64(0)}
		},
	},

	"check": {
		AcceptFailure: true,
		Argsn:         2,
		Doc:           "Check if Arg 1 is not failure, if it wraps it into another Failure (Arg 2), otherwise returns Arg 1.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			if ps.FailureFlag || arg0.Type() == env.ErrorType {
				ps.FailureFlag = false
				switch er := arg0.(type) {
				case *env.Error: // todo .. make Error type .. make error construction micro dialect, return the error wrapping error that caused it
					if er.Status == 0 && er.Message == "" {
						er = nil
					}
					return MakeRyeError(ps, arg1, er)
				}
			}
			return arg0
		},
	},

	"^check": {
		AcceptFailure: true,
		Argsn:         2,
		Doc:           "Returning Check.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			if ps.FailureFlag {
				ps.ReturnFlag = true
				switch er := arg0.(type) {
				case *env.Error: // todo .. make Error type .. make error construction micro dialect, return the error wrapping error that caused it
					if er.Status == 0 && er.Message == "" {
						er = nil
					}
					return MakeRyeError(ps, arg1, er)
				}
				return env.NewError("error 1")
			}
			return arg0
		},
	},

	"^require": {
		AcceptFailure: true,
		Argsn:         2,
		Doc:           "Returning Require.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch cond := arg0.(type) {
			case env.Object:
				if !util.IsTruthy(cond) {
					ps.FailureFlag = true
					ps.ReturnFlag = true
					return MakeRyeError(ps, arg1, nil)
				} else {
					return arg0
				}
				return arg0
			}
			return arg0
		},
	},

	"require": {
		AcceptFailure: true,
		Argsn:         2,
		Doc:           "Require that first argument is Truthy value, if not produce a failure based on second argument",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch cond := arg0.(type) {
			case env.Object:
				if !util.IsTruthy(cond) {
					ps.FailureFlag = true
					// ps.ReturnFlag = true
					return MakeRyeError(ps, arg1, nil)
				} else {
					return arg0
				}
				return arg0
			}
			return arg0
		},
	},

	"assert-equal": {
		Argsn: 2,
		Doc:   "Test if two values are equal. Fail if not.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			if arg0.GetKind() == arg1.GetKind() && arg0.Inspect(*ps.Idx) == arg1.Inspect(*ps.Idx) {
				return env.Integer{1}
			} else {
				return makeError(ps, "Values are not equal: "+arg0.Inspect(*ps.Idx)+" "+arg1.Inspect(*ps.Idx))
			}
		},
	},

	"fix": {
		AcceptFailure: true,
		Argsn:         2,
		Doc:           "If Arg 1 is a failure, do a block and return the result of it, otherwise return Arg 1.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			if ps.FailureFlag || arg0.Type() == env.ErrorType {
				ps.FailureFlag = false
				// TODO -- create function do_block and call in all cases
				switch bloc := arg1.(type) {
				case env.Block:
					ser := ps.Ser
					ps.Ser = bloc.Series
					EvalBlockInj(ps, arg0, true)
					ps.Ser = ser
					return ps.Res
				default:
					ps.FailureFlag = true
					return env.NewError("expecting block")
				}
			} else {
				return arg0
			}
		},
	},

	"^fix": {
		AcceptFailure: true,
		Argsn:         2,
		Doc:           "Fix as a returning function. If Arg 1 is failure, do the block and return to caller.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			if ps.FailureFlag || arg0.Type() == env.ErrorType {
				ps.FailureFlag = false
				// TODO -- create function do_block and call in all cases
				switch bloc := arg1.(type) {
				case env.Block:
					ser := ps.Ser
					ps.Ser = bloc.Series
					EvalBlockInj(ps, arg0, true)
					ps.Ser = ser
					ps.ReturnFlag = true
					return ps.Res
				default:
					ps.FailureFlag = true
					return env.NewError("expecting block")
				}
			} else {
				return arg0
			}
		},
	},

	"`fix": {
		AcceptFailure: true,
		Argsn:         2,
		Doc:           "Fix as a skipping function.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			if ps.FailureFlag || arg0.Type() == env.ErrorType {
				ps.FailureFlag = false
				// TODO -- create function do_block and call in all cases
				switch bloc := arg1.(type) {
				case env.Block:
					ser := ps.Ser
					ps.Ser = bloc.Series
					EvalBlockInj(ps, arg0, true)
					ps.Ser = ser
					ps.SkipFlag = true
					return ps.Res
				default:
					ps.FailureFlag = true
					return env.NewError("expecting block")
				}
			} else {
				return arg0
			}
		},
	},

	"fix\\either": {
		AcceptFailure: true,
		Argsn:         3,
		Doc:           "Fix also with else block.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			if ps.FailureFlag || arg0.Type() == env.ErrorType {
				ps.FailureFlag = false
				// TODO -- create function do_block and call in all cases
				switch bloc := arg1.(type) {
				case env.Block:
					ser := ps.Ser
					ps.Ser = bloc.Series
					EvalBlockInj(ps, arg0, true)
					ps.Ser = ser
					return ps.Res
				default:
					ps.FailureFlag = true
					return env.NewError("expecting block")
				}
			} else {
				switch bloc := arg2.(type) {
				case env.Block:
					ser := ps.Ser
					ps.Ser = bloc.Series
					EvalBlockInj(ps, arg0, true)
					ps.Ser = ser
					return ps.Res
				default:
					ps.FailureFlag = true
					return env.NewError("expecting block")
				}
			}
		},
	},

	"fix-else": {
		AcceptFailure: true,
		Argsn:         2,
		Doc:           "Do a block of code if Arg 1 is not a failure.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			if !(ps.FailureFlag || arg0.Type() == env.ErrorType) {
				ps.FailureFlag = false
				// TODO -- create function do_block and call in all cases
				switch bloc := arg1.(type) {
				case env.Block:
					ser := ps.Ser
					ps.Ser = bloc.Series
					EvalBlockInj(ps, arg0, true)
					ps.Ser = ser
					return ps.Res
				default:
					ps.FailureFlag = true
					return env.NewError("expecting block")
				}
			} else {
				return arg0
			}
		},
	},

	"load": {
		Argsn: 1,
		Doc:   "Loads a string into Rye values.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.String:
				block, _ := loader.LoadString(s1.Value, false)
				//ps = env.AddToProgramState(ps, block.Series, genv)
				return block
			case env.Uri:
				var str string
				fileIdx, _ := ps.Idx.GetIndex("file")
				if s1.Scheme.Index == fileIdx {
					b, err := ioutil.ReadFile(s1.GetPath())
					if err != nil {
						return makeError(ps, err.Error())
					}
					str = string(b) // convert content to a 'string'
				}
				block, _ := loader.LoadString(str, false)
				//ps = env.AddToProgramState(ps, block.Series, genv)
				return block
			default:
				ps.FailureFlag = true
				return env.NewError("Must be string or file TODO")
			}
		},
	},

	"load-sig": {
		Argsn: 1,
		Doc:   "Checks the signature, if OK then loads a string into Rye values.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.String:
				block, _ := loader.LoadString(s1.Value, true)
				//ps = env.AddToProgramState(ps, block.Series, genv)
				return block
			default:
				ps.FailureFlag = true
				return env.NewError("Must be string or file TODO")
			}
		},
	},

	// BASIC ENV / Dict FUNCTIONS
	"format": {
		Argsn: 1,
		Doc:   "Accepts a Dict and returns formatted presentation of it as a string.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			var r strings.Builder
			switch s1 := arg0.(type) {
			case env.Dict:
				for k, v := range s1.Data {
					r.WriteString(k)
					r.WriteString(":\n\t")
					r.WriteString(fmt.Sprintln(v))
				}
			default:
				return env.String{arg0.Probe(*ps.Idx)}
			}
			return env.String{r.String()}
		},
	},

	// date time functions
	"now": {
		Argsn: 0,
		Doc:   "Returns current Time.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return env.Time{time.Now()}
		},
	},

	// end of date time functions

	"to-context": {
		Argsn: 1,
		Doc:   "Takes a Dict and returns a Context with same names and values.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Dict:

				return util.Dict2Context(ps, s1)
				// make new context with no parent

			default:
				fmt.Println("Error")
			}
			return nil
		},
	},

	"length?": {
		Argsn: 1,
		Doc:   "Accepts a collection (String, Block, Dict, Spreadsheet) and returns it's length.", // TODO -- accept list, context also
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.String:
				return env.Integer{int64(len(s1.Value))}
			case env.Dict:
				return env.Integer{int64(len(s1.Data))}
			case env.Block:
				return env.Integer{int64(s1.Series.Len())}
			case env.Spreadsheet:
				return env.Integer{int64(len(s1.Rows))}
			case env.RyeCtx:
				return env.Integer{int64(s1.GetWords(*ps.Idx).Series.Len())}
			default:
				fmt.Println("Error")
			}
			return nil
		},
	},
	"ncols": {
		Doc:   "Accepts a Spreadsheet and returns number of columns.",
		Argsn: 1,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Dict:
			case env.Block:
			case env.Spreadsheet:
				return env.Integer{int64(len(s1.Cols))}
			default:
				fmt.Println("Error")
			}
			return nil
		},
	},
	"keys": {
		Argsn: 1,
		Doc:   "Accepts Dict or Spreadsheet and returns a Block of keys or column names.", // TODO -- accept context also
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Dict:
				keys := make([]env.Object, len(s1.Data))
				i := 0
				for k, _ := range s1.Data {
					keys[i] = env.String{k}
					i++
				}
				return *env.NewBlock(*env.NewTSeries(keys))
			case env.Spreadsheet:
				keys := make([]env.Object, len(s1.Cols))
				for i, k := range s1.Cols {
					keys[i] = env.String{k}
				}
				return *env.NewBlock(*env.NewTSeries(keys))
			default:
				fmt.Println("Error")
			}
			return nil
		},
	},

	"colsum": {
		Argsn: 2,
		Doc:   "Accepts a spreadsheet and a column name and returns a sum of a column.", // TODO -- let it accept a block and list also
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			var name string
			switch s1 := arg0.(type) {
			case env.Spreadsheet:
				switch s2 := arg1.(type) {
				case env.Tagword:
					name = ps.Idx.GetWord(s2.Index)
				case env.String:
					name = s2.Value
				default:
					ps.ErrorFlag = true
					return env.NewError("second arg not string")
				}
				r := s1.Sum(name)
				if r.Type() == env.ErrorType {
					ps.ErrorFlag = true
				}
				return r

			default:
				ps.ErrorFlag = true
				return env.NewError("first arg not spreadsheet")
			}
			return nil
		},
	},
	"A1": {
		Argsn: 1,
		Doc:   "Accepts a Spreadsheet and returns the first row first column cell.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s0 := arg0.(type) {
			case env.Spreadsheet:
				r := s0.Rows[0].Values[0]
				return JsonToRye(r)

			default:
				ps.ErrorFlag = true
				return env.NewError("first arg not spreadsheet")
			}
			return nil
		},
	},
	"B1": {
		Argsn: 1,
		Doc:   "Accepts a Spreadsheet and returns the first row second column cell.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s0 := arg0.(type) {
			case env.Spreadsheet:
				r := s0.Rows[0].Values[1]
				return JsonToRye(r)

			default:
				ps.ErrorFlag = true
				return env.NewError("first arg not spreadsheet")
			}
			return nil
		},
	},

	/* Terminal functions .. move to it's own later */

	"cmd": {
		Argsn: 1,
		Doc:   "",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s0 := arg0.(type) {
			case env.String:
				/*				cmd := exec.Command("date")
								err := cmd.Run()
								if err != nil {
									log.Fatal(err)
								}
								fmt.Println("out:", outb.String(), "err:", errb.String()) */

				r := exec.Command("/bin/bash", "-c", s0.Value)
				// stdout, stderr := r.Output()
				var outb, errb bytes.Buffer
				r.Stdout = &outb
				r.Stderr = &errb

				err := r.Run()
				if err != nil {
					fmt.Println("ERROR")
					fmt.Println(err)
				}
				fmt.Println("out:", outb.String(), "err:", errb.String())

				/*				if stderr != nil {
									fmt.Println(stderr.Error())
								}
								return JsonToRye(" "-----------" + string(stdout)) */
				//				return JsonToRye(string(stdout))
			default:
				return makeError(ps, "Arg 1 should be String")
			}
			return nil
		},
	},

	"rye": {
		Argsn: 0,
		Doc:   "",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewNative(ps.Idx, BuiltinNames, "Rye-itself")
		},
	},

	"Rye-itself//needs": {
		Argsn: 2,
		Doc:   "",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch mod := arg1.(type) {
			case env.Block:
				var str strings.Builder
				missing := make([]env.Object, 0)
				for i := 0; i < mod.Series.Len(); i++ {
					switch node := mod.Series.Get(i).(type) {
					case env.Word:
						name := ps.Idx.GetWord(node.Index)
						cnt, ok := BuiltinNames[name]
						// TODO -- distinguish between modules that aren't loaded or don't exists
						if ok && cnt > 0 {
							//							return env.Integer{1}
						} else {
							str.WriteString("\nBinding *" + name + "* is missing.")
							missing = append(missing, node)
							// v0 todo: Print mis
							// v1 todo: Print the instructions of what modules to go get in the project folder and reinstall
							// v2 todo: ge get modules and then recompile rye with these flags into current folder
							//							return env.Integer{0}
						}
					}
				}
				if len(missing) > 0 {
					return makeError(ps, str.String())
				} else {
					return env.Integer{1}
				}
				// return *env.NewBlock(*env.NewTSeries(missing))
			default:
				return makeError(ps, "Arg 1 should be Block of Tagwords.")
			}
		},
	},

	"Rye-itself//includes?": {
		Argsn: 1,
		Doc:   "",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			blts := make([]env.Object, 0)
			for i, v := range BuiltinNames {
				if v > 0 {
					idx := ps.Idx.IndexWord(i)
					blts = append(blts, env.Word{idx})
				}
			}
			return *env.NewBlock(*env.NewTSeries(blts))
		},
	},
	"Rye-itself//can-include?": {
		Argsn: 1,
		Doc:   "",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			blts := make([]env.Object, 0)
			for i, v := range BuiltinNames {
				if v == 0 {
					idx := ps.Idx.IndexWord(i)
					blts = append(blts, env.Word{idx})
				}
			}
			return *env.NewBlock(*env.NewTSeries(blts))
		},
	},
	"Rye-itself//args": {
		Argsn: 1,
		Doc:   "",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return util.StringToFieldsWithQuoted(strings.Join(os.Args, " "), " ", "\"")
			// block, _ := loader.LoadString(os.Args[0], false)
			// return block
		},
	},
}

/* Terminal functions .. move to it's own later */

/*
func isTruthy(arg env.Object) env.Object {
	switch cond := arg.(type) {
	case env.Integer:
		return cond.Value != 0
	case env.String:
		return cond.Value != ""
	default:
		// if it's neither we just return error for now
		ps.FailureFlag = true
		return env.NewError("Error determining if truty")
	}
}
*/

func RegisterBuiltins(ps *env.ProgramState) {
	BuiltinNames = make(map[string]int)
	RegisterBuiltins2(builtins, ps, "core")
	RegisterBuiltins2(Builtins_io, ps, "io")
	RegisterBuiltins2(Builtins_regexp, ps, "regexp")
	RegisterBuiltins2(Builtins_web, ps, "web")
	RegisterBuiltins2(Builtins_sxml, ps, "sxml")
	RegisterBuiltins2(Builtins_html, ps, "html")
	RegisterBuiltins2(Builtins_sqlite, ps, "sqlite")
	RegisterBuiltins2(Builtins_gtk, ps, "gtk")
	RegisterBuiltins2(Builtins_validation, ps, "validation")
	RegisterBuiltins2(Builtins_ps, ps, "ps")
	RegisterBuiltins2(Builtins_nats, ps, "nats")
	RegisterBuiltins2(Builtins_qframe, ps, "qframe")
	RegisterBuiltins2(Builtins_webview, ps, "webview")
	RegisterBuiltins2(Builtins_json, ps, "json")
	RegisterBuiltins2(Builtins_stackless, ps, "stackless")
	RegisterBuiltins2(Builtins_eyr, ps, "eyr")
	RegisterBuiltins2(Builtins_conversion, ps, "conversion")
	RegisterBuiltins2(Builtins_nng, ps, "nng")
	RegisterBuiltins2(Builtins_http, ps, "http")
	RegisterBuiltins2(Builtins_crypto, ps, "crypto")
	RegisterBuiltins2(Builtins_goroutines, ps, "gorourines")
	RegisterBuiltins2(Builtins_psql, ps, "psql")
	RegisterBuiltins2(Builtins_mysql, ps, "mysql")
	RegisterBuiltins2(Builtins_bcrypt, ps, "bcrypt")
	RegisterBuiltins2(Builtins_raylib, ps, "raylib")
	RegisterBuiltins2(Builtins_email, ps, "email")
	RegisterBuiltins2(Builtins_cayley, ps, "cayley")
	RegisterBuiltins2(Builtins_structures, ps, "structs")
	RegisterBuiltins2(Builtins_telegrambot, ps, "telegram")
	RegisterBuiltins2(Builtins_spreadsheet, ps, "spreadsheet")
}

var BuiltinNames map[string]int

func RegisterBuiltins2(builtins map[string]*env.Builtin, ps *env.ProgramState, name string) {
	BuiltinNames[name] = len(builtins)
	for k, v := range builtins {
		bu := env.NewBuiltin(v.Fn, v.Argsn, v.AcceptFailure, v.Pure, v.Doc)
		registerBuiltin(ps, k, *bu)
	}
}

func registerBuiltin(ps *env.ProgramState, word string, builtin env.Builtin) {
	// indexWord
	// TODO -- this with string separator is a temporary way of how we define generic builtins
	// in future a map will probably not be a map but an array and builtin will also support the Kind value

	idxk := 0
	if strings.Index(word, "//") > 0 {
		temp := strings.Split(word, "//")
		word = temp[1]
		idxk = ps.Idx.IndexWord(temp[0])
	}
	idxw := ps.Idx.IndexWord(word)
	// set global word with builtin
	if idxk == 0 {
		ps.Ctx.Set(idxw, builtin)
		if builtin.Pure {
			ps.PCtx.Set(idxw, builtin)
		}

	} else {
		ps.Gen.Set(idxk, idxw, builtin)
	}
}
