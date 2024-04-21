package evaldo

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"io"
	"math"
	"math/big"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"

	"github.com/refaktor/rye/env"
	"github.com/refaktor/rye/term"

	"github.com/refaktor/rye/loader"
	// JM 20230825	"github.com/refaktor/rye/term"
	"strconv"
	"strings"
	"time"

	"github.com/refaktor/rye/util"

	"golang.org/x/sync/errgroup"
	goterm "golang.org/x/term"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func ss() {
	fmt.Print(1)
}

func MakeError(env1 *env.ProgramState, msg string) *env.Error {
	env1.FailureFlag = true
	return env.NewError(msg)
}

func MakeBuiltinError(env1 *env.ProgramState, msg string, fn string) *env.Error {
	env1.FailureFlag = true
	return env.NewError(msg + " in builtin " + fn + ".")
}

func MakeArgError(env1 *env.ProgramState, N int, typ []env.Type, fn string) *env.Error {
	env1.FailureFlag = true
	types := ""
	for i, tt := range typ {
		if i > 0 {
			types += ", "
		}
		types += env.NativeTypes[tt-1]
	}
	return env.NewError("Function " + fn + " requires argument " + strconv.Itoa(N) + " to be of	: " + types + ".")
}

func MakeNativeArgError(env1 *env.ProgramState, N int, knd []string, fn string) *env.Error {
	env1.FailureFlag = true
	kinds := strings.Join(knd, ", ")
	return env.NewError("Function " + fn + " requires native argument " + strconv.Itoa(N) + " to be of kind	: " + kinds + ".")
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
	case env.String:
		switch vB := arg1.(type) {
		case env.String:
			return vA.Value > vB.Value
		default:
			return false
		}
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

func lesserThanNew(arg0 env.Object, arg1 env.Object) bool {
	var valA float64
	var valB float64
	switch vA := arg0.(type) {
	case env.Integer:
		valA = float64(vA.Value)
	case env.Decimal:
		valA = vA.Value
	case env.String:
		switch vB := arg1.(type) {
		case env.String:
			return vA.Value < vB.Value
		default:
			return false
		}
	}
	switch vB := arg1.(type) {
	case env.Integer:
		valB = float64(vB.Value)
	case env.Decimal:
		valB = vB.Value
	}
	return valA < valB
}

func getFrom(ps *env.ProgramState, data any, key any, posMode bool) env.Object {
	switch s1 := data.(type) {
	case env.Dict:
		switch s2 := key.(type) {
		case env.String:
			v := s1.Data[s2.Value]
			switch v1 := v.(type) {
			case int, int64, float64, string, []any, map[string]any:
				return env.ToRyeValue(v1)
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
			case *env.List:
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
		case env.Word:
			v, ok := s1.Get(s2.Index)
			if ok {
				return v
			} else {
				return makeError(ps, "Not found in context")
			}
		case env.Tagword:
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
				return env.ToRyeValue(v)
			} else {
				ps.FailureFlag = true
				return env.NewError1(5) // NOT_FOUND
			}
		}
	case *env.List:
		switch s2 := key.(type) {
		case env.Integer:
			idx := s2.Value
			if posMode {
				idx--
			}
			v := s1.Data[idx]
			ok := true
			if ok {
				return env.ToRyeValue(v)
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
				return env.ToRyeValue(v)
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
				return env.ToRyeValue(v)
			} else {
				ps.FailureFlag = true
				return env.NewError1(5) // NOT_FOUND
			}
		}
	}
	return makeError(ps, "Wrong type or missing key for get-arrow")
}

// Sort object interface
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

// Sort list interface
type RyeListSort []any

func (s RyeListSort) Len() int {
	return len(s)
}
func (s RyeListSort) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s RyeListSort) Less(i, j int) bool {
	return greaterThanNew(env.ToRyeValue(s[j]), env.ToRyeValue(s[i]))
}

var ShowResults bool

var builtins = map[string]*env.Builtin{

	"to-word": { // ***
		Argsn: 1,
		Doc:   "Tries to change a Rye value to a word with same name.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch str := arg0.(type) {
			case env.String:
				idx := ps.Idx.IndexWord(str.Value)
				return *env.NewWord(idx)
			case env.Word:
				return *env.NewWord(str.Index)
			case env.Xword:
				return *env.NewWord(str.Index)
			case env.EXword:
				return *env.NewWord(str.Index)
			case env.Tagword:
				return *env.NewWord(str.Index)
			case env.Setword:
				return *env.NewWord(str.Index)
			case env.LSetword:
				return *env.NewWord(str.Index)
			case env.Getword:
				return *env.NewWord(str.Index)
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType, env.WordType}, "to-word")
			}
		},
	},

	"to-integer": { // ***
		Argsn: 1,
		Doc:   "Tries to change a Rye value (like string) to integer.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch addr := arg0.(type) {
			case env.String:
				iValue, err := strconv.Atoi(addr.Value)
				if err != nil {
					return MakeBuiltinError(ps, err.Error(), "to-integer")
				}
				return *env.NewInteger(int64(iValue))
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "to-integer")
			}
		},
	},

	"to-decimal": { // ***
		Argsn: 1,
		Doc:   "Tries to change a Rye value (like string) to integer.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch addr := arg0.(type) {
			case env.String:
				floatVal, err := strconv.ParseFloat(addr.Value, 64)

				if err != nil {
					// Handle the error if the conversion fails (e.g., invalid format)
					return MakeBuiltinError(ps, err.Error(), "to-decimal")
				}
				return *env.NewDecimal(floatVal)
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "to-integer")
			}
		},
	},

	"to-string": { // ***
		Argsn: 1,
		Doc:   "Tries to turn a Rye value to string.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewString(arg0.Print(*ps.Idx))
		},
	},

	"to-context": { // ***
		Argsn: 1,
		Doc:   "Takes a Dict and returns a Context with same names and values.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Dict:

				return util.Dict2Context(ps, s1)
				// make new context with no parent

			default:
				return MakeArgError(ps, 1, []env.Type{env.DictType}, "to-context")
			}
		},
	},

	"is-string": { // ***
		Argsn: 1,
		Doc:   "Returns true if value is a string.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			if arg0.Type() == env.StringType {
				return *env.NewInteger(1)
			} else {
				return *env.NewInteger(0)
			}
		},
	},

	"is-integer": { // ***
		Argsn: 1,
		Doc:   "Returns true if value is an integer.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			if arg0.Type() == env.IntegerType {
				return *env.NewInteger(1)
			} else {
				return *env.NewInteger(0)
			}
		},
	},

	"is-decimal": { // ***
		Argsn: 1,
		Doc:   "Returns true if value is a decimal.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			if arg0.Type() == env.DecimalType {
				return *env.NewInteger(1)
			} else {
				return *env.NewInteger(0)
			}
		},
	},

	"is-number": { // ***
		Argsn: 1,
		Doc:   "Returns true if value is a number (integer or decimal).",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			if arg0.Type() == env.IntegerType || arg0.Type() == env.DecimalType {
				return *env.NewInteger(1)
			} else {
				return *env.NewInteger(0)
			}
		},
	},

	"to-uri": { // ** TODO-FIXME: return possible failures
		Argsn: 1,
		Doc:   "Tries to change Rye value to an URI.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch val := arg0.(type) {
			case env.String:
				return *env.NewUri1(ps.Idx, val.Value) // TODO turn to switch
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "to-file")
			}

		},
	},

	"to-file": { // **  TODO-FIXME: return possible failures
		Argsn: 1,
		Doc:   "Tries to change Rye value to a file.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch val := arg0.(type) {
			case env.String:
				return *env.NewUri1(ps.Idx, "file://"+val.Value) // TODO turn to switch
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "to-file")
			}
		},
	},

	"type?": { // ***
		Argsn: 1,
		Doc:   "Returns the type of Rye value as a word.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewWord(int(arg0.Type()))
		},
	},

	"types?": { // TODO
		Argsn: 1,
		Doc:   "Returns the types of Rye values in a block or spreadsheet row as a block of words.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch list := arg0.(type) {
			case env.Block:
				l := list.Series.Len()
				newl := make([]env.Object, l)
				for i := 0; i < l; i++ {
					newl[i] = *env.NewWord(int(list.Series.S[i].Type()))
				}
				return *env.NewBlock(*env.NewTSeries(newl))
			case env.Spreadsheet:
				l := len(list.Rows[0].Values)
				newl := make([]env.Object, l)
				for i := 0; i < l; i++ {
					newl[i] = *env.NewWord(int(env.ToRyeValue(list.Rows[0].Values[i]).Type()))
				}
				return *env.NewBlock(*env.NewTSeries(newl))
			case env.SpreadsheetRow:
				l := len(list.Values)
				newl := make([]env.Object, l)
				for i := 0; i < l; i++ {
					newl[i] = *env.NewWord(int(env.ToRyeValue(list.Values[i]).Type()))
				}
				return *env.NewBlock(*env.NewTSeries(newl))
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType, env.SpreadsheetType, env.SpreadsheetRowType}, "types?")
			}
		},
	},

	// NUMBERS

	"inc": { // ***
		Argsn: 1,
		Doc:   "Returns integer value incremented by 1.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch arg := arg0.(type) {
			case env.Integer:
				return *env.NewInteger(1 + arg.Value)
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType}, "inc")
			}
		},
	},

	"is-positive": { // ***
		Argsn: 1,
		Doc:   "Returns true if integer is positive.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch arg := arg0.(type) {
			case env.Integer:
				if arg.Value > 0 {
					return *env.NewInteger(1)
				} else {
					return *env.NewInteger(0)
				}
			case env.Decimal:
				if arg.Value > 0 {
					return *env.NewInteger(1)
				} else {
					return *env.NewInteger(0)
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType, env.DecimalType}, "is-positive")
			}
		},
	},

	"is-zero": { // ***
		Argsn: 1,
		Doc:   "Returns true if integer is zero.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) (res env.Object) {
			switch arg := arg0.(type) {
			case env.Integer:
				if arg.Value == 0 {
					return *env.NewInteger(1)
				} else {
					return *env.NewInteger(0)
				}
			case env.Decimal:
				if arg.Value == 0 {
					return *env.NewInteger(1)
				} else {
					return *env.NewInteger(0)
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType, env.DecimalType}, "is-zero")
			}
		},
	},

	"inc!": { // ***
		Argsn: 1,
		Doc:   "Searches for a word and increments it's integer value in-place.",
		Pure:  false,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch arg := arg0.(type) {
			case env.Word:
				intval, found, ctx := ps.Ctx.Get2(arg.Index)
				if found {
					switch iintval := intval.(type) {
					case env.Integer:
						ctx.Set(arg.Index, *env.NewInteger(1 + iintval.Value))
						return *env.NewInteger(1 + iintval.Value)
					default:
						return MakeBuiltinError(ps, "Value in word is not integer.", "inc!")
					}
				}
				return MakeBuiltinError(ps, "Word not found in context.", "inc!")

			default:
				return MakeArgError(ps, 1, []env.Type{env.WordType}, "inc!")
			}
		},
	},

	"dec!": { // ***
		Argsn: 1,
		Doc:   "Searches for a word and increments it's integer value in-place.",
		Pure:  false,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch arg := arg0.(type) {
			case env.Word:
				intval, found, ctx := ps.Ctx.Get2(arg.Index)
				if found {
					switch iintval := intval.(type) {
					case env.Integer:
						ctx.Set(arg.Index, *env.NewInteger(iintval.Value - 1))
						return *env.NewInteger(1 + iintval.Value)
					default:
						return MakeBuiltinError(ps, "Value in word is not integer.", "inc!")
					}
				}
				return MakeBuiltinError(ps, "Word not found in context.", "inc!")

			default:
				return MakeArgError(ps, 1, []env.Type{env.WordType}, "inc!")
			}
		},
	},

	"change!": { // ***
		Argsn: 2,
		Doc:   "Searches for a word and changes it's value in-place. If value changes returns true otherwise false",
		Pure:  false,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch arg := arg1.(type) {
			case env.Word:
				val, found, ctx := ps.Ctx.Get2(arg.Index)
				if found {
					ctx.Set(arg.Index, arg0)
					var res int64
					if arg0.GetKind() == val.GetKind() && arg0.Inspect(*ps.Idx) == val.Inspect(*ps.Idx) {
						res = 0
					} else {
						res = 1
					}
					return *env.NewInteger(res)
				}
				return MakeBuiltinError(ps, "Word not found in context.", "change!")
			default:
				return MakeArgError(ps, 1, []env.Type{env.WordType}, "change!")
			}
		},
	},

	"set": { // ***
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
								return MakeBuiltinError(ps, "More words than values.", "set")
							}
							val := vals.Series.S[i]
							// if it exists then we set it to word from words
							ps.Ctx.Set(word.Index, val)
						default:
							fmt.Println(word)
							return MakeBuiltinError(ps, "Only words in words block", "set")
						}
					}
					return arg0
				default:
					return MakeArgError(ps, 2, []env.Type{env.WordType}, "set")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.WordType}, "set")
			}
		},
	},

	"get_": { // *** find a name or decide on order of naming with generic words clashes with
		Argsn: 1,
		Doc:   "Returns value of the word in context",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch w := arg0.(type) {
			case env.Word:
				object, found := ps.Ctx.Get(w.Index)
				if found {
					return object
				} else {
					return MakeBuiltinError(ps, "Word not found in contexts	", "get")
				}
			case env.Opword:
				object, found := ps.Ctx.Get(w.Index)
				if found {
					return object
				} else {
					return MakeBuiltinError(ps, "Word not found in contexts	", "get")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.WordType}, "set")
			}
		},
	},

	// CONTINUE WORK HERE - SYSTEMATISATION

	"dump": { // *** currently a concept in testing ... for getting a code of a function, maybe same would be needed for context?
		Argsn: 1,
		Doc:   "Returns (dumps) Rye code representing the object.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return env.NewString(arg0.Dump(*ps.Idx))
		},
	},

	// TODO -- make save\\context ctx %file
	"save\\current": {
		Argsn: 0,
		Doc:   "Saves current state of the program to a file.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) (res env.Object) {
			s := ps.Dump()
			fileName := fmt.Sprintf("console_%s.rye", time.Now().Format("060102_150405"))

			err := os.WriteFile(fileName, []byte(s), 0600)
			if err != nil {
				ps.FailureFlag = true
				return MakeBuiltinError(ps, fmt.Sprintf("error writing state: %s", err.Error()), "save\\state")
			}
			fmt.Println("State current context to \033[1m" + fileName + "\033[0m.")
			return *env.NewInteger(1)
		},
	},

	// TODO -- make save\\context ctx %file
	"save\\current\\secure": {
		Argsn: 0,
		Doc:   "Saves current state of the program to a file.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) (res env.Object) {
			s := ps.Dump()
			fileName := fmt.Sprintf("console_%s.rye.enc", time.Now().Format("060102_150405"))

			fmt.Print("Enter Password: ")
			bytePassword, err := goterm.ReadPassword(int(os.Stdin.Fd()))
			if err != nil {
				panic(err)
			}
			password := string(bytePassword)

			util.SaveSecure(s, fileName, password)
			/*  err != nil {
				ps.FailureFlag = true
				return MakeBuiltinError(ps, fmt.Sprintf("error writing state: %s", err.Error()), "save\\state")
			}*/
			fmt.Println("State current context to \033[1m" + fileName + "\033[0m.")
			return *env.NewInteger(1)
		},
	},

	"doc": { // ***
		Argsn: 1,
		Doc:   "Sets docstring of the current context.",
		Pure:  true,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch d := arg0.(type) {
			case env.String:
				env1.Ctx.Doc = d.Value
				return *env.NewInteger(1)
			default:
				return MakeArgError(env1, 1, []env.Type{env.StringType}, "doc")
			}
		},
	},

	"doc?": { // ***
		Argsn: 0,
		Doc:   "Gets docstring of the current context.",
		Pure:  true,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewString(env1.Ctx.Doc)
		},
	},

	"doc\\of?": { // **
		Argsn: 1,
		Doc:   "Get docstring of the passed context.",
		Pure:  true,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch d := arg0.(type) {
			case env.Function:
				return *env.NewString(d.Doc)
			case env.Builtin:
				return *env.NewString(d.Doc)
			case env.RyeCtx:
				return *env.NewString(d.Doc)
			default:
				return MakeArgError(env1, 1, []env.Type{env.FunctionType, env.CtxType}, "doc\\of?")
			}

		},
	},

	// VALUES

	"dict": { // ***
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

	"list": { // ***
		Argsn: 1,
		Doc:   "Constructs a List from the Block of values.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch bloc := arg0.(type) {
			case env.Block:
				return env.NewListFromSeries(bloc.Series)
			}
			return nil
		},
	},

	"true": { // ***
		Argsn: 0,
		Doc:   "Returns a truthy value.",
		Pure:  true,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewInteger(1)
		},
	},

	"false": { // ***
		Argsn: 0,
		Doc:   "Returns a falsy value.",
		Pure:  true,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewInteger(0)
		},
	},

	"not": { // ***
		Argsn: 1,
		Doc:   "Turns a truthy value to a non-truthy and reverse.",
		Pure:  true,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			if util.IsTruthy(arg0) {
				return *env.NewInteger(0)
			} else {
				return *env.NewInteger(1)
			}
		},
	},

	"and": {
		Argsn: 2,
		Doc:   "Bitwise AND operation between two values.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Integer:
				switch s2 := arg1.(type) {
				case env.Integer:
					return *env.NewInteger(s1.Value & s2.Value)
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "and")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType}, "and")
			}
		},
	},
	"or": {
		Argsn: 2,
		Doc:   "Bitwise OR operation between two values.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Integer:
				switch s2 := arg1.(type) {
				case env.Integer:
					return *env.NewInteger(s1.Value | s2.Value)
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "or")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType}, "or")
			}
		},
	},
	"xor": {
		Argsn: 2,
		Doc:   "Bitwise XOR operation between two values.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Integer:
				switch s2 := arg1.(type) {
				case env.Integer:
					return *env.NewInteger(s1.Value ^ s2.Value)
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "xor")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType}, "xor")
			}
		},
	},

	"require_": { //
		Argsn: 1,
		Doc:   "Requite a truthy value or produce a failure.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			if util.IsTruthy(arg0) {
				return *env.NewInteger(1)
			} else {
				return MakeBuiltinError(ps, "Requirement failed.", "require_")
			}
		},
	},

	// BASIC FUNCTIONS WITH NUMBERS

	"multiple-of": { // ***
		Argsn: 2,
		Doc:   "Checks if a first argument is a factor of second.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch a := arg0.(type) {
			case env.Integer:
				switch b := arg1.(type) {
				case env.Integer:
					if a.Value%b.Value == 0 {
						return *env.NewInteger(1)
					} else {
						return *env.NewInteger(0)
					}
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "multiple-of")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType}, "multiple-of")
			}
		},
	},
	"odd": { // ***
		Argsn: 1,
		Doc:   "Checks if a number is odd.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch a := arg0.(type) {
			case env.Integer:
				if a.Value%2 != 0 {
					return *env.NewInteger(1)
				} else {
					return *env.NewInteger(0)
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType}, "odd")
			}
		},
	},
	"even": { // ***
		Argsn: 1,
		Doc:   "Checks if a number is even.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch a := arg0.(type) {
			case env.Integer:
				if a.Value%2 == 0 {
					return *env.NewInteger(1)
				} else {
					return *env.NewInteger(0)
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType}, "odd")
			}
		},
	},

	"mod": { // ***
		Argsn: 2,
		Doc:   "Calculates module (remainder) of two integers.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch a := arg0.(type) {
			case env.Integer:
				switch b := arg1.(type) {
				case env.Integer:
					return *env.NewInteger(a.Value % b.Value)
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "mod")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType}, "mod")
			}
		},
	},

	"_+": { // **
		Argsn: 2,
		Doc:   "Adds or joins two values together (Integers, Strings, Uri-s and Blocks)",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Integer:
				switch s2 := arg1.(type) {
				case env.Integer:
					return *env.NewInteger(s1.Value + s2.Value)
				case env.Decimal:
					return *env.NewDecimal(float64(s1.Value) + s2.Value)
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType, env.DecimalType}, "_+")
				}
			case env.Decimal:
				switch s2 := arg1.(type) {
				case env.Integer:
					return *env.NewDecimal(s1.Value + float64(s2.Value))
				case env.Decimal:
					return *env.NewDecimal(s1.Value + s2.Value)
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType, env.DecimalType}, "_+")
				}
			case env.String:
				switch s2 := arg1.(type) {
				case env.String:
					return *env.NewString(s1.Value + s2.Value)
				case env.Integer:
					return *env.NewString(s1.Value + strconv.Itoa(int(s2.Value)))
				case env.Decimal:
					return *env.NewString(s1.Value + strconv.FormatFloat(s2.Value, 'f', -1, 64))
				default:
					return MakeArgError(ps, 2, []env.Type{env.StringType, env.IntegerType, env.DecimalType}, "_+")
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
						case env.Word:
							_, err := str.WriteString(sepa + ps.Idx.GetWord(node.Index) + "=")
							if err != nil {
								return MakeBuiltinError(ps, "WriteString failed for Word type.", "_+")
							}
							sepa = "&"
						case env.String:
							_, err := str.WriteString(node.Value)
							if err != nil {
								return MakeBuiltinError(ps, "WriteString failed for String type.", "_+")
							}
						case env.Integer:
							_, err := str.WriteString(strconv.Itoa(int(node.Value)))
							if err != nil {
								return MakeBuiltinError(ps, "WriteString failed for Integer type.", "_+")
							}
						case env.Uri:
							_, err := str.WriteString(node.GetPath())
							if err != nil {
								return MakeBuiltinError(ps, "WriteString failed for Uri type.", "_+")
							}
						default:
							return MakeBuiltinError(ps, "Value in node is not word, string, int or uri type.", "_+")
						}
					}
					return *env.NewUri(ps.Idx, s1.Scheme, s1.Path+str.String())
				default:
					return MakeArgError(ps, 2, []env.Type{env.StringType, env.IntegerType, env.BlockType}, "_+")
				}
			case env.Block:
				switch b2 := arg1.(type) {
				case env.Block:
					s := &s1.Series
					s1.Series = *s.AppendMul(b2.Series.GetAll())
					return s1
				default:
					return MakeBuiltinError(ps, "Value in Block is not block type.", "_+")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType, env.IntegerType, env.BlockType, env.DecimalType, env.UriType}, "_+")
			}
		},
	},

	"_-": { // **
		Argsn: 2,
		Doc:   "Subtracts two numbers.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch a := arg0.(type) {
			case env.Integer:
				switch b := arg1.(type) {
				case env.Integer:
					return *env.NewInteger(a.Value - b.Value)
				case env.Decimal:
					return *env.NewDecimal(float64(a.Value) - b.Value)
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType, env.DecimalType}, "_-")
				}
			case env.Decimal:
				switch b := arg1.(type) {
				case env.Integer:
					return *env.NewDecimal(a.Value - float64(b.Value))
				case env.Decimal:
					return *env.NewDecimal(a.Value - b.Value)
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType, env.DecimalType}, "_-")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType, env.DecimalType}, "_-")
			}
		},
	},

	"_*": { // **
		Argsn: 2,
		Doc:   "Multiplies two numbers.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch a := arg0.(type) {
			case env.Integer:
				switch b := arg1.(type) {
				case env.Integer:
					return *env.NewInteger(a.Value * b.Value)
				case env.Decimal:
					return *env.NewDecimal(float64(a.Value) * b.Value)
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType, env.DecimalType}, "_*")
				}
			case env.Decimal:
				switch b := arg1.(type) {
				case env.Integer:
					return *env.NewDecimal(a.Value * float64(b.Value))
				case env.Decimal:
					return *env.NewDecimal(a.Value * b.Value)
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType, env.DecimalType}, "_*")
				}
			default:
				return MakeArgError(ps, 2, []env.Type{env.IntegerType, env.DecimalType}, "_*")
			}
		},
	},
	"_/": { // **
		Argsn: 2,
		Doc:   "Divided two numbers.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch a := arg0.(type) {
			case env.Integer:
				switch b := arg1.(type) {
				case env.Integer:
					if b.Value == 0 {
						ps.FailureFlag = true
						return MakeBuiltinError(ps, "Can't divide by Zero.", "_/")
					}
					return *env.NewDecimal(float64(a.Value) / float64(b.Value))
				case env.Decimal:
					if b.Value == 0.0 {
						ps.FailureFlag = true
						return MakeBuiltinError(ps, "Can't divide by Zero.", "_/")
					}
					return *env.NewDecimal(float64(a.Value) / b.Value)
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType, env.DecimalType}, "_/")
				}
			case env.Decimal:
				switch b := arg1.(type) {
				case env.Integer:
					if b.Value == 0 {
						ps.FailureFlag = true
						return MakeBuiltinError(ps, "Can't divide by Zero.", "_/")
					}
					return *env.NewDecimal(a.Value / float64(b.Value))
				case env.Decimal:
					if b.Value == 0.0 {
						ps.FailureFlag = true
						return MakeBuiltinError(ps, "Can't divide by Zero.", "_/")
					}
					return *env.NewDecimal(a.Value / b.Value)
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType, env.DecimalType}, "_/")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType, env.DecimalType}, "_/")
			}
		},
	},
	"_=": { // ***
		Argsn: 2,
		Doc:   "Checks if two Rye values are equal.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			var res int64
			if arg0.Equal(arg1) {
				res = 1
			} else {
				res = 0
			}
			return *env.NewInteger(res)
		},
	},

	"_>": { // *** // ENDED FIXING DOCSTRINGS HERE
		Argsn: 2,
		Doc:   "Checks if first argument is greater than the second.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			if greaterThan(ps, arg0, arg1) {
				return *env.NewInteger(1)
			} else {
				return *env.NewInteger(0)
			}
		},
	},
	"_>=": { // * *
		Argsn: 2,
		Doc:   "Checks if first argument is greater or equal than the second.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			if arg0.Equal(arg1) || greaterThan(ps, arg0, arg1) {
				return *env.NewInteger(1)
			} else {
				return *env.NewInteger(0)
			}
		},
	},
	"_<": { // **
		Argsn: 2,
		Doc:   "Tests if Arg1 is lesser than Arg 2.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			if lesserThan(ps, arg0, arg1) {
				return *env.NewInteger(1)
			} else {
				return *env.NewInteger(0)
			}
		},
	},
	"_<=": {
		Argsn: 2,
		Doc:   "Tests if Arg1 is lesser than Arg 2.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			if arg0.Equal(arg1) || lesserThan(ps, arg0, arg1) {
				return *env.NewInteger(1)
			} else {
				return *env.NewInteger(0)
			}
		},
	},

	// BASIC GENERAL FUNCTIONS

	"prns": { // **
		Argsn: 1,
		Doc:   "Prints a value and adds a space.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch arg := arg0.(type) {
			case env.String:
				fmt.Print(arg.Value + " ")
			default:
				fmt.Print(arg0.Print(*ps.Idx) + " ")
			}
			return arg0
		},
	},
	"prn": { // **
		Argsn: 1,
		Doc:   "Prints a value without newline.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch arg := arg0.(type) {
			case env.String:
				fmt.Print(arg.Value)
			default:
				fmt.Print(arg0.Print(*ps.Idx))
			}
			return arg0
		},
	},
	"print": { // **
		Argsn: 1,
		Doc:   "Prints a value and adds a newline.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch arg := arg0.(type) {
			case env.String:
				fmt.Println(arg.Value)
			default:
				fmt.Println(arg0.Print(*ps.Idx))
			}
			return arg0
		},
	},
	"embed": { // **
		Argsn: 2,
		Doc:   "Embeds a value in string.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch arg := arg1.(type) {
			case env.String:
				vals := arg0.Print(*ps.Idx)
				news := strings.ReplaceAll(arg.Value, "{}", vals)
				return *env.NewString(news)
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "esc")
			}
		},
	},
	"prnv": { // **
		Argsn: 2,
		Doc:   "Prints a value.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch arg := arg1.(type) {
			case env.String:
				vals := arg0.Print(*ps.Idx)
				news := strings.ReplaceAll(arg.Value, "{}", vals)
				fmt.Print(news)
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "esc")
			}
			return arg0
		},
	},
	"printv": { // **
		Argsn: 2,
		Doc:   "Prints a value and adds a newline.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch arg := arg1.(type) {
			case env.String:
				vals := arg0.Print(*ps.Idx)
				news := strings.ReplaceAll(arg.Value, "{}", vals)
				fmt.Println(news)
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "esc")
			}
			return arg0
		},
	},
	"print\\ssv": { //
		Argsn: 1,
		Doc:   "Prints a value and adds a newline.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch arg := arg0.(type) {
			case env.Object:
				fmt.Println(util.FormatSsv(arg, *ps.Idx))
			default:
				return MakeBuiltinError(ps, "Not Rye object.", "print-ssv")
			}
			return arg0
		},
	},
	"print\\csv": { //
		Argsn: 1,
		Doc:   "Prints a value and adds a newline.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch arg := arg0.(type) {
			case env.Object:
				fmt.Println(util.FormatCsv(arg, *ps.Idx))
			default:
				return MakeBuiltinError(ps, "Not Rye object.", "print-csv")
			}
			return arg0
		},
	},
	"print\\json": { //
		Argsn: 1,
		Doc:   "Prints a value and adds a newline.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch arg := arg0.(type) {
			case env.Object:
				fmt.Println(util.FormatJson(arg, *ps.Idx))
			default:
				return MakeBuiltinError(ps, "Not Rye object.", "print-json")
			}
			return arg0
		},
	},
	"probe": { // **
		Argsn: 1,
		Doc:   "Prints information about a value.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			fmt.Println(arg0.Inspect(*ps.Idx))
			return arg0
		},
	},
	"inspect": { // **
		Argsn: 1,
		Doc:   "Returs information about a value.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewString(arg0.Inspect(*ps.Idx))
		},
	},
	"esc": {
		Argsn: 1,
		Doc:   "Creates an escape sequence \033{}",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch arg := arg0.(type) {
			case env.String:
				return *env.NewString("\033" + arg.Value)
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "esc")
			}
		},
	},
	"esc-val": {
		Argsn: 2,
		Doc:   "Escapes a value and adds a newline.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch base := arg1.(type) {
			case env.String:
				vals := arg0.Print(*ps.Idx)
				news := strings.ReplaceAll(base.Value, "(*)", vals)
				return *env.NewString("\033" + news)
			default:
				return MakeArgError(ps, 2, []env.Type{env.StringType}, "esc-val")
			}
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
			case env.SpreadsheetRow:
				obj, esc := term.DisplaySpreadsheetRow(bloc, ps.Idx)
				if !esc {
					return obj
				}
			}

			return arg0
		},
	},
	/* JM 20230825
	"tui\\selection": {
		Argsn: 2,
		Doc:   "Work in progress Interactively displays a value.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			// This is temporary implementation for experimenting what it would work like at all
			// later it should belong to the object (and the medium of display, terminal, html ..., it's part of the frontend)
			term.SaveCurPos()
			switch bloc := arg0.(type) {
			case env.Block:
				obj, esc := term.DisplaySelection(bloc, ps.Idx, int(arg1.(env.Integer).Value))
				if !esc {
					return obj
				}
			}
			return nil
		},
	},
	"tui\\input": {
		Argsn: 2,
		Doc:   "Work in progress Interactively displays a value.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			// This is temporary implementation for experimenting what it would work like at all
			// later it should belong to the object (and the medium of display, terminal, html ..., it's part of the frontend)
			term.SaveCurPos()
			switch bloc := arg0.(type) {
			case env.Integer:
				obj, esc := term.DisplayInputField(int(bloc.Value), int(arg1.(env.Integer).Value))
				if !esc {
					return obj
				}
			}
			return nil
		},
	}, */

	"random-integer": {
		Argsn: 1,
		Doc:   "Accepts an integer n and eturns a random integer between 0 and n in the half-open interval [0,n).",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch arg := arg0.(type) {
			case env.Integer:
				val, err := rand.Int(rand.Reader, big.NewInt(arg.Value))
				if err != nil {
					return MakeBuiltinError(ps, err.Error(), "random-integer")
				}
				return *env.NewInteger(val.Int64())
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType}, "random-integer")
			}
		},
	},

	"import": { // **
		Argsn: 1,
		Doc:   "Imports a file, loads and does it from script local path.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Uri:
				var str string
				fileIdx, _ := ps.Idx.GetIndex("file")
				fullpath := filepath.Join(filepath.Dir(ps.ScriptPath), s1.GetPath())
				if s1.Scheme.Index == fileIdx {
					b, err := os.ReadFile(fullpath)
					if err != nil {
						return MakeBuiltinError(ps, err.Error(), "import")
					}
					str = string(b) // convert content to a 'string'
				}
				script_ := ps.ScriptPath
				ps.ScriptPath = fullpath
				block_ := loader.LoadStringNEW(str, false, ps)
				switch block := block_.(type) {
				case env.Block:
					ser := ps.Ser
					ps.Ser = block.Series
					EvalBlock(ps)
					ps.Ser = ser
				case env.Error:
					ps.ScriptPath = script_
					ps.ErrorFlag = true
					return MakeBuiltinError(ps, block.Message, "import")
				default:
					fmt.Println(block)
					panic("Not block and not error in import builtin.") // TODO -- Think how best to handle this
				}
				ps.ScriptPath = script_
				//ps = env.AddToProgramState(ps, block.Series, genv)
				return ps.Res
			default:
				return MakeArgError(ps, 1, []env.Type{env.UriType}, "import")
			}
		},
	},

	"load": { // **
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
					b, err := os.ReadFile(s1.GetPath())
					if err != nil {
						return makeError(ps, err.Error())
					}
					str = string(b) // convert content to a 'string'
				}
				scrip := ps.ScriptPath
				ps.ScriptPath = s1.GetPath()
				block := loader.LoadStringNEW(str, false, ps)
				ps.ScriptPath = scrip
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

	"mold": { // **
		Argsn: 1,
		Doc:   "Turn value to it's string representation.",
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			// fmt.Println()
			return *env.NewString(arg0.Print(*env1.Idx))
		},
	},

	"mold\\nowrap": { // **
		Argsn: 1,
		Doc:   "Turn value to it's string representation.",
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			// fmt.Println()
			str := arg0.Print(*env1.Idx)
			if len(str) > 0 {
				if str[0] == '{' || str[0] == '[' {
					str = str[1 : len(str)-1]
				}
			}
			str = strings.ReplaceAll(str, "._", "")  // temporary solution for special op-words
			str = strings.ReplaceAll(str, "|_", "|") // temporary solution for special op-words
			return *env.NewString(strings.Trim(str, " "))
		},
	},

	// CONTROL WORDS

	"otherwise": { // **
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
				return *env.NewInteger(0)
				// else {
				//	return MakeBuiltinError(ps, "Truthiness condition is not correct.", "otherwise")
				// }
			default:
				return MakeArgError(ps, 2, []env.Type{env.BlockType}, "otherwise")
			}
		},
	},

	"if": { // **
		Argsn: 2,
		Doc:   "Basic conditional. Takes a condition and a block of code.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			// function accepts 2 args. arg0 is a "boolean" value, arg1 is a block of code
			// we set bloc to block of code
			// (we don't have boolean type yet, because it's not crucial to important part of design, neither is truthiness ... this will be decided later
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
				return *env.NewInteger(0)
				// else {
				//	return MakeBuiltinError(ps, "Truthiness condition is not correct.", "if")
				// }
			default:
				// if it's not a block we return error for now
				ps.FailureFlag = true
				return MakeArgError(ps, 2, []env.Type{env.BlockType}, "if")
			}
		},
	},

	"^if": { // **
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
				return *env.NewInteger(0)
				// else {
				//	return MakeBuiltinError(ps, "Truthiness condition is not correct.", "^if")
				// }
			default:
				return MakeArgError(ps, 2, []env.Type{env.BlockType}, "^if")
			}
		},
	},

	"^otherwise": { // **
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
				return *env.NewInteger(0)
				// else {
				// 	return MakeBuiltinError(ps, "Truthiness condition is not correct.", "^otherwise")
				// }
			default:
				return MakeArgError(ps, 2, []env.Type{env.BlockType}, "^otherwise")
			}
		},
	},

	"either": { // **
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
						return MakeArgError(ps, 1, []env.Type{env.IntegerType, env.StringType}, "either")
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
				default:
					return MakeArgError(ps, 3, []env.Type{env.BlockType}, "either")
				}
			case env.Object:
				switch bloc2 := arg2.(type) {
				case env.Object: // TODO , if true value is not block then also false value will be treated as literal for now
					switch cond := arg0.(type) {
					case env.Integer:
						cond1 = cond.Value != 0
					case env.String:
						cond1 = cond.Value != ""
					default:
						return MakeArgError(ps, 1, []env.Type{env.IntegerType, env.StringType}, "either")
					}
					if cond1 {
						return bloc1
					} else {
						return bloc2
					}
				default:
					return MakeBuiltinError(ps, "Third argument must be Object Type.", "either")
				}
			default:
				//Note - ObjectType is not available so using MakeBuiltinError instead of MakeArgError
				return MakeBuiltinError(ps, "Second argument must be Block or Object Type.", "either")
			}
		},
	},

	"^tidy-switch": {
		Argsn:         2,
		AcceptFailure: true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			fmt.Println("FLAGS")

			ps.FailureFlag = false

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
							return MakeBuiltinError(ps, "Switch block malformed.", "^tidy-switch")
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
						default:
							return MakeBuiltinError(ps, "Invalid type in block series.", "^tidy-switch")
						}
					}
					switch cc := code.(type) {
					case env.Block:
						fmt.Println(code.Print(*ps.Idx))
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
						return MakeBuiltinError(ps, "Malformed switch block.", "^tidy-switch")
					}
				default:
					// if it's not a block we return error for now
					ps.FailureFlag = true
					return MakeArgError(ps, 2, []env.Type{env.BlockType}, "^tidy-switch")
				}
			default:
				return arg0
			}
		},
	},

	"switch": { // **
		Argsn:         2,
		Doc:           "Classic switch function. Takes a word and multiple possible values and block of code to do.",
		AcceptFailure: true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch bloc := arg1.(type) {
			case env.Block:

				var code env.Object

				any_found := false

				for i := 0; i < bloc.Series.Len(); i += 2 {

					if i > bloc.Series.Len()-2 {
						return MakeBuiltinError(ps, "Switch block malformed.", "switch")
					}

					ev := bloc.Series.Get(i)
					if arg0.GetKind() == ev.GetKind() && arg0.Inspect(*ps.Idx) == ev.Inspect(*ps.Idx) {
						any_found = true
						code = bloc.Series.Get(i + 1)
					}
					if ev.Type() == env.VoidType {
						if !any_found {
							code = bloc.Series.Get(i + 1)
							any_found = true
						}
					}
				}
				if any_found {
					switch cc := code.(type) {
					case env.Block:
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
						return MakeBuiltinError(ps, "Malformed switch block.", "switch")
					}
				}
				return arg0
			default:
				// if it's not a block we return error for now
				ps.FailureFlag = true
				return MakeArgError(ps, 2, []env.Type{env.BlockType}, "switch")
			}
		},
	},

	"cases": { // ** , TODO-FIXME: error handling
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
						if !foundany {
							doblk = true
						}
					default:
						return MakeBuiltinError(ps, "Invalid block series type.", "cases")
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
				return MakeArgError(ps, 2, []env.Type{env.BlockType}, "cases")
			}
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
							fmt.Println(ps.Ctx.Print(*ps.Idx))
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
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "enter-console")
			}
		},
	},

	/* TEMP FOR WASM 20250116
	"get-input": {
		Argsn: 1,
		Doc:   "Stops execution and gives you a Rye console, to test the code inside environment.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch name := arg0.(type) {
			case env.String:
				// fmt.Print("Get Input: \033[1m" + name.Value + "\033[0m")
				DoGeneralInput(ps, name.Value)
				return ps.Res
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "get-input")
			}
		},
	},

	"input-field": {
		Argsn: 1,
		Doc:   "Stops execution and gives you a Rye console, to test the code inside environment.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch name := arg0.(type) {
			case env.String:
				// fmt.Print("Get Input: \033[1m" + name.Value + "\033[0m")
				DoGeneralInputField(ps, name.Value)
				return ps.Res
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "input-field")
			}
		},
	}, */

	// DOERS

	"do": { // **
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
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType}, "do")
			}
		},
	},

	"try": { // **
		Argsn: 1,
		Doc:   "Takes a block of code and does (runs) it.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch bloc := arg0.(type) {
			case env.Block:
				ser := ps.Ser
				ps.Ser = bloc.Series
				EvalBlock(ps)

				// TODO -- probably shouldn't just display error ... but we return it and then handle it / display it
				// MaybeDisplayFailureOrError(ps, ps.Idx)

				ps.ReturnFlag = false
				ps.ErrorFlag = false
				ps.FailureFlag = false

				ps.Ser = ser
				return ps.Res
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType}, "do")
			}
		},
	},

	"with": { // **
		AcceptFailure: true,
		Doc:           "Takes a value and a block of code. It does the code with the value injected.",
		Argsn:         2,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch bloc := arg1.(type) {
			case env.Block:
				ser := ps.Ser
				ps.Ser = bloc.Series
				EvalBlockInj(ps, arg0, true)
				ps.Ser = ser
				return ps.Res
			default:
				return MakeArgError(ps, 2, []env.Type{env.BlockType}, "with")
			}
		},
	},

	"do\\in": { // **
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
					return MakeArgError(ps, 2, []env.Type{env.BlockType}, "do-in")
				}
			default:
				ps.ErrorFlag = true
				return MakeArgError(ps, 1, []env.Type{env.CtxType}, "do-in")
			}

		},
	},

	"do\\in\\try": { // **
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

					// TODO -- probably shouldn't just display error ... but we return it and then handle it / display it
					// MaybeDisplayFailureOrError(ps, ps.Idx)

					ps.ReturnFlag = false
					ps.ErrorFlag = false
					ps.FailureFlag = false

					ps.Ser = ser
					return ps.Res
				default:
					ps.ErrorFlag = true
					return MakeArgError(ps, 2, []env.Type{env.BlockType}, "do-in")
				}
			default:
				ps.ErrorFlag = true
				return MakeArgError(ps, 1, []env.Type{env.CtxType}, "do-in")
			}

		},
	},

	"capture-stdout": { // **
		Argsn: 1,
		Doc:   "Takes a block of code and does (runs) it.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch bloc := arg0.(type) {
			case env.Block:

				old := os.Stdout // keep backup of the real stdout
				r, w, _ := os.Pipe()
				os.Stdout = w

				ser := ps.Ser
				ps.Ser = bloc.Series
				EvalBlock(ps)
				ps.Ser = ser

				outC := make(chan string, 1)
				g := errgroup.Group{}
				// copy the output in a separate goroutine so printing can't block indefinitely
				g.Go(func() error {
					var buf bytes.Buffer
					_, err := io.Copy(&buf, r)
					if err != nil {
						return err
					}
					outC <- buf.String()
					return nil
				})

				// back to normal state
				w.Close()
				os.Stdout = old // restoring the real stdout

				if err := g.Wait(); err != nil {
					return MakeBuiltinError(ps, fmt.Sprintf("Error reading stdout: %v", err), "capture-stdout")
				}
				out := <-outC

				// reading our temp stdout
				// fmt.Println("previous output:")
				// fmt.Print(out)

				return *env.NewString(out)
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType}, "capture-stdout")
			}
		},
	},

	//

	"time-it": { // **
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
				return *env.NewInteger(elapsed.Nanoseconds() / 1000000)
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType}, "time-it")
			}
		},
	},

	"vals": { // **
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
				return MakeBuiltinError(ps, "Value not found.", "vals")
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType, env.WordType}, "vals")
			}
		},
	},

	"vals\\with": { // **
		Argsn: 2,
		Doc:   "Evaluate a block with injecting the first argument.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch bloc := arg1.(type) {
			case env.Block:
				ser := ps.Ser
				ps.Ser = bloc.Series
				res := make([]env.Object, 0)
				injnow := true
				for ps.Ser.Pos() < ps.Ser.Len() {
					// ps, injnow = EvalExpressionInj(ps, inj, injnow)
					//20231203 EvalExpressionInjectedVALS(ps, arg0, true)
					ps, injnow = EvalExpressionInj(ps, arg0, injnow)
					res = append(res, ps.Res)
					// check and raise the flags if needed if true (error) return
					//if checkFlagsAfterBlock(ps, 101) {
					//	return ps
					//}
					// if return flag was raised return ( errorflag I think would return in previous if anyway)
					//if checkErrorReturnFlag(ps) {
					//	return ps
					//}
					ps, injnow = MaybeAcceptComma(ps, arg0, injnow)
				}
				ps.Ser = ser
				return *env.NewBlock(*env.NewTSeries(res))
			default:
				return MakeArgError(ps, 2, []env.Type{env.BlockType}, "vals\\with")
			}
		},
	},

	"all": { // **
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
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType}, "all")
			}
		},
	},

	"any": { // **
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
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType}, "any")
			}
		},
	},

	"any\\with": { // TODO-FIXME error handling, halts on multiple expressions
		Argsn: 2,
		Doc:   "Takes a block, if any of the values or expressions are truthy, then it returns that one, in none false.",
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
			default:
				return MakeArgError(ps, 2, []env.Type{env.BlockType}, "any\\with")
			}
		},
	},

	// EXPERIMENTAL: COLLECTION AND RETURNING FUNCTIONS ... NOT DETERMINED IF WILL BE INCLUDED YET

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
		Doc:   "Collects key value pars to implicit block.",
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
		Doc:   "Collects key value pars to implicit block.",
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
		Doc:   "Returns the implicit data structure that we collected.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return ps.ForcedResult
		},
	},

	"pop-collected": {
		Argsn: 0,
		Doc:   "Returns the implicit collected data structure and resets it.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			result := ps.ForcedResult
			ps.ForcedResult = nil
			return result
		},
	},

	// CONTEXT

	"current-ctx": { // **
		Argsn: 0,
		Doc:   "Returns current context.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *ps.Ctx
		},
	},

	"parent-ctx": { // **
		Argsn: 0,
		Doc:   "Returns parent context of the current context.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *ps.Ctx.Parent
		},
	},

	"ls": {
		Argsn: 0,
		Doc:   "Lists words in current context",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			fmt.Println(ps.Ctx.Preview(*ps.Idx, ""))
			return env.Void{}
		},
	},

	"lsp": {
		Argsn: 0,
		Doc:   "Lists words in current context",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			fmt.Println(ps.Ctx.Parent.Preview(*ps.Idx, ""))
			return env.Void{}
		},
	},

	"ls\\": {
		Argsn: 1,
		Doc:   "Lists words in current context with string filter",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.String:
				fmt.Println(ps.Ctx.Preview(*ps.Idx, s1.Value))
				return env.Void{}
			case env.RyeCtx:
				fmt.Println(s1.Preview(*ps.Idx, ""))
				return env.Void{}
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "ls\\")
			}
		},
	},
	"lsp\\": {
		Argsn: 1,
		Doc:   "Lists words in current context with string filter",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.String:
				fmt.Println(ps.Ctx.Parent.Preview(*ps.Idx, s1.Value))
				return env.Void{}
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "ls\\")
			}
		},
	},

	"cc": {
		Argsn: 1,
		Doc:   "Change to context",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.RyeCtx:
				ps.Ctx = &s1
				return s1
			default:
				return MakeArgError(ps, 1, []env.Type{env.CtxType}, "cc")
			}
		},
	},

	"ccp": {
		Argsn: 0,
		Doc:   "Change to context",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			cc := ps.Ctx
			ps.Ctx = ps.Ctx.Parent
			return *cc
		},
	},

	"mkcc": {
		Argsn: 1,
		Doc:   "Make context with current as parent and change to it.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch word := arg0.(type) {
			case env.Word:
				newctx := env.NewEnv(ps.Ctx)
				ps.Ctx.Set(word.Index, *newctx)
				ctx := ps.Ctx
				ps.Ctx = newctx // make new context with current par
				return *ctx
			default:
				return MakeArgError(ps, 1, []env.Type{env.WordType}, "mkcc")
			}
		},
	},

	"raw-context": { // **
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
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType}, "raw-context")
			}
		},
	},

	"isolate": { // **
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
				rctx.Kind = *env.NewWord(-1)
				ps.Ctx = ctx
				ps.Ser = ser
				if ps.ErrorFlag {
					return ps.Res
				}
				return *rctx // return the resulting context
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType}, "isolate")
			}
		},
	},

	"context": { // **
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
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType}, "context")
			}
		},
	},

	"private": { // **
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
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType}, "private")
			}
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
					return MakeArgError(ps, 2, []env.Type{env.BlockType}, "private\\")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "private\\")
			}
		},
	},

	"extend": { // ** exclamation mark, because it as it is now extends/changes the source context too .. in place
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
				default:
					return MakeArgError(ps, 2, []env.Type{env.BlockType}, "extend")
				}
			default:
				ps.ErrorFlag = true
				return MakeArgError(ps, 1, []env.Type{env.CtxType}, "extend")
			}
		},
	},

	"bind": { // **
		Argsn: 2,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch swCtx1 := arg0.(type) {
			case env.RyeCtx:
				switch swCtx2 := arg1.(type) {
				case env.RyeCtx:
					swCtx1.Parent = &swCtx2
					return swCtx1
				default:
					return MakeArgError(ps, 2, []env.Type{env.CtxType}, "bind")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.CtxType}, "bind")
			}
		},
	},

	"unbind": { // **
		Argsn: 1,
		Doc:   "Accepts a Context and unbinds it from it's parent Context.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch swCtx1 := arg0.(type) {
			case env.RyeCtx:
				swCtx1.Parent = nil
				return swCtx1
			default:
				return MakeArgError(ps, 1, []env.Type{env.CtxType}, "unbind")
			}
		},
	},

	// COMBINATORS

	"pass": { // **
		Argsn: 2,
		Doc:   "Accepts a value and a block. It does the block, with value injected, and returns (passes on) the initial value.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch bloc := arg1.(type) {
			case env.Block:
				res := arg0
				ser := ps.Ser
				ps.Ser = bloc.Series
				EvalBlockInj(ps, arg0, true)
				if ps.ErrorFlag {
					return ps.Res
				}
				ps.Ser = ser
				if ps.ReturnFlag {
					return ps.Res
				}
				return res
			default:
				return MakeArgError(ps, 2, []env.Type{env.BlockType}, "pass")
			}
		},
	},

	"wrap": { // **
		Argsn: 2,
		Doc:   "Accepts a value and a block. It does the block, with value injected, and returns (passes on) the initial value.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch wrap := arg0.(type) {
			case env.Block:
				switch bloc := arg1.(type) {
				case env.Block:
					ser := ps.Ser
					ps.Ser = wrap.Series
					EvalBlockInj(ps, arg0, true)
					if ps.ReturnFlag {
						return ps.Res
					}

					ps.Ser = bloc.Series
					EvalBlockInj(ps, arg0, true)
					if ps.ReturnFlag {
						return ps.Res
					}
					res := ps.Res

					ps.Ser = wrap.Series
					EvalBlockInj(ps, arg0, true)
					ps.Ser = ser
					if ps.ReturnFlag {
						return ps.Res
					}
					return res
				default:
					return MakeArgError(ps, 2, []env.Type{env.BlockType}, "wrap")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType}, "wrap")
			}
		},
	},

	"keep": { // **
		Argsn: 3,
		Doc:   "Do the first block, then the second one but return the result of the first one.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch b1 := arg1.(type) {
			case env.Block:
				switch b2 := arg2.(type) {
				case env.Block:
					ser := ps.Ser
					ps.Ser = b1.Series
					EvalBlockInj(ps, arg0, true)
					if ps.ErrorFlag {
						return ps.Res
					}
					res := ps.Res
					ps.Ser = b2.Series
					EvalBlockInj(ps, arg0, true)
					ps.Ser = ser
					return res
				default:
					return MakeArgError(ps, 3, []env.Type{env.BlockType}, "keep")
				}
			default:
				return MakeArgError(ps, 2, []env.Type{env.BlockType}, "keep")
			}
		},
	},

	// LOOPING

	"loop": { // **
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
						ps = EvalBlockInj(ps, *env.NewInteger(int64(i + 1)), true)
						if ps.ErrorFlag {
							return ps.Res
						}
						ps.Ser.Reset()
					}
					ps.Ser = ser
					return ps.Res
				default:
					return MakeArgError(ps, 2, []env.Type{env.BlockType}, "loop")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType}, "loop")
			}
		},
	},

	"produce": { // **
		Argsn: 3,
		Doc:   "Accepts a number, initial value and a block of code. Does the block of code number times, injecting the number.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch cond := arg0.(type) {
			case env.Integer:
				switch bloc := arg2.(type) {
				case env.Block:
					acc := arg1
					ser := ps.Ser
					ps.Ser = bloc.Series
					ps.Res = arg1
					for i := 0; int64(i) < cond.Value; i++ {
						ps = EvalBlockInj(ps, acc, true)
						if ps.ErrorFlag {
							return ps.Res
						}
						ps.Ser.Reset()
						acc = ps.Res
					}
					ps.Ser = ser
					return ps.Res
				default:
					return MakeArgError(ps, 3, []env.Type{env.BlockType}, "produce")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType}, "produce")
			}
		},
	},

	"produce\\while": { // **
		Argsn: 3,
		Doc:   "Accepts a while condition, initial value and a block of code. Does the block of code number times, injecting the number first and then result of block.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch cond := arg0.(type) {
			case env.Block:
				switch bloc := arg2.(type) {
				case env.Block:
					acc := arg1
					ser := ps.Ser
					for {
						ps.Ser = cond.Series
						ps = EvalBlockInj(ps, acc, true)
						if ps.ErrorFlag {
							return ps.Res
						}
						if !util.IsTruthy(ps.Res) {
							ps.Ser.Reset()
							ps.Ser = ser
							return acc
						}
						ps.Ser.Reset()
						ps.Ser = bloc.Series
						ps = EvalBlockInj(ps, acc, true)
						if ps.ErrorFlag {
							return ps.Res
						}
						ps.Ser = ser
						ps.Ser.Reset()
						acc = ps.Res
					}
				default:
					return MakeArgError(ps, 3, []env.Type{env.BlockType}, "produce")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType}, "produce")
			}
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
					case env.Word:
						acc := arg1
						ps.Ctx.Set(accu.Index, acc)
						ser := ps.Ser
						ps.Ser = bloc.Series
						for i := 0; int64(i) < cond.Value; i++ {
							ps = EvalBlockInj(ps, acc, true)
							if ps.ErrorFlag {
								return ps.Res
							}
							ps.Ser.Reset()
							acc = ps.Res
						}
						ps.Ser = ser
						val, _ := ps.Ctx.Get(accu.Index)
						return val
					default:
						return MakeArgError(ps, 3, []env.Type{env.WordType}, "produce\\")
					}
				default:
					return MakeArgError(ps, 4, []env.Type{env.BlockType}, "produce\\")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType}, "produce\\")
			}
		},
	},

	"forever": { // **
		Argsn: 1,
		Doc:   "Accepts a block and does it forever.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch bloc := arg0.(type) {
			case env.Block:
				ser := ps.Ser
				ps.Ser = bloc.Series
				for {
					ps = EvalBlock(ps)
					if ps.ErrorFlag {
						return ps.Res
					}
					if ps.ReturnFlag {
						ps.ReturnFlag = false
						break
					}
					ps.Ser.Reset()
				}
				ps.Ser = ser
				return ps.Res
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.BlockType}, "forever")
			}
		},
	},
	"forever\\with": { // **
		Argsn: 2,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch bloc := arg1.(type) {
			case env.Block:
				ser := ps.Ser
				ps.Ser = bloc.Series
				for {
					EvalBlockInj(ps, arg0, true)
					if ps.ErrorFlag {
						return ps.Res
					}
					if ps.ReturnFlag {
						ps.ReturnFlag = false
						break
					}
					ps.Ser.Reset()
				}
				ps.Ser = ser
				return ps.Res
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.BlockType}, "forever\\with")
			}
		},
	},

	"for": { // **
		Argsn: 2,
		Doc:   "Accepts a block of values and a block of code, does the code for each of the values, injecting them.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch block := arg0.(type) {
			case env.String:
				switch code := arg1.(type) {
				case env.Block:
					ser := ps.Ser
					ps.Ser = code.Series
					for _, ch := range block.Value {
						ps = EvalBlockInj(ps, *env.NewString(string(ch)), true)
						if ps.ErrorFlag {
							return ps.Res
						}
						ps.Ser.Reset()
					}
					ps.Ser = ser
					return ps.Res
				default:
					return MakeArgError(ps, 2, []env.Type{env.BlockType}, "for")
				}
			case env.Block:
				switch code := arg1.(type) {
				case env.Block:
					ser := ps.Ser
					ps.Ser = code.Series
					for i := 0; i < block.Series.Len(); i++ {
						ps = EvalBlockInj(ps, block.Series.Get(i), true)
						if ps.ErrorFlag {
							return ps.Res
						}
						ps.Ser.Reset()
					}
					ps.Ser = ser
					return ps.Res
				default:
					return MakeArgError(ps, 2, []env.Type{env.BlockType}, "for")
				}
			case env.List:
				switch code := arg1.(type) {
				case env.Block:
					ser := ps.Ser
					ps.Ser = code.Series
					for i := 0; i < len(block.Data); i++ {
						ps = EvalBlockInj(ps, env.ToRyeValue(block.Data[i]), true)
						if ps.ErrorFlag {
							return ps.Res
						}
						ps.Ser.Reset()
					}
					ps.Ser = ser
					return ps.Res
				default:
					return MakeArgError(ps, 2, []env.Type{env.BlockType}, "for")
				}
			case env.Spreadsheet:
				switch code := arg1.(type) {
				case env.Block:
					ser := ps.Ser
					ps.Ser = code.Series
					for i := 0; i < len(block.Rows); i++ {
						row := block.Rows[i]
						row.Uplink = &block
						ps = EvalBlockInj(ps, row, true)
						if ps.ErrorFlag {
							return ps.Res
						}
						ps.Ser.Reset()
					}
					ps.Ser = ser
					return ps.Res
				default:
					return MakeArgError(ps, 2, []env.Type{env.BlockType}, "for")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType, env.BlockType, env.SpreadsheetType}, "for")
			}
		},
	},

	"walk": { // **
		Argsn: 2,
		Doc:   "Accepts a block of values and a block of code, does the code for each of the values, injecting them.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch block := arg0.(type) {
			case env.Block:
				switch code := arg1.(type) {
				case env.Block:
					ser := ps.Ser
					ps.Ser = code.Series

					for block.Series.GetPos() < block.Series.Len() {
						ps = EvalBlockInj(ps, block, true)
						if ps.ErrorFlag {
							return ps.Res
						}
						if ps.ReturnFlag {
							return ps.Res
						}
						block1, ok := ps.Res.(env.Block) // TODO ... switch and throw error if not block
						if ok {
							block = block1
						} else {
							fmt.Println("ERROR 1231241")
						}
						ps.Ser.Reset()
					}
					ps.Ser = ser
					return ps.Res
				default:
					return MakeArgError(ps, 2, []env.Type{env.BlockType}, "for")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType}, "for")
			}
		},
	},

	// Higher order functions

	"purge": { // TODO ... doesn't fully work
		Argsn: 2,
		Doc:   "Purges values from a series based on return of a injected code block.",
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
						if ps.ErrorFlag {
							return ps.Res
						}
						if util.IsTruthy(ps.Res) {
							block.Series.S = append(block.Series.S[:i], block.Series.S[i+1:]...)
							i--
						}
						ps.Ser.Reset()
					}
					ps.Ser = ser
					return block
				default:
					return MakeArgError(ps, 2, []env.Type{env.BlockType}, "purge")
				}
			case env.List:
				switch code := arg1.(type) {
				case env.Block:
					ser := ps.Ser
					ps.Ser = code.Series
					for i := 0; i < len(block.Data); i++ {
						ps = EvalBlockInj(ps, env.ToRyeValue(block.Data[i]), true)
						if ps.ErrorFlag {
							return ps.Res
						}
						if util.IsTruthy(ps.Res) {
							block.Data = append(block.Data[:i], block.Data[i+1:]...)
							i--
						}
						ps.Ser.Reset()
					}
					ps.Ser = ser
					return block
				default:
					return MakeArgError(ps, 2, []env.Type{env.BlockType}, "purge")
				}
			case env.String:
				switch code := arg1.(type) {
				case env.Block:
					input := []rune(block.Value)
					var newl []env.Object
					ser := ps.Ser
					ps.Ser = code.Series
					for i := 0; i < len(input); i++ {
						ps = EvalBlockInj(ps, env.ToRyeValue(input[i]), true)
						if ps.ErrorFlag {
							return ps.Res
						}
						if !util.IsTruthy(ps.Res) {
							newl = append(newl, *env.NewString(string(input[i])))
						}
						ps.Ser.Reset()
					}
					ps.Ser = ser
					return *env.NewBlock(*env.NewTSeries(newl))
				default:
					return MakeArgError(ps, 2, []env.Type{env.BlockType}, "purge")
				}
			case env.Spreadsheet:
				switch code := arg1.(type) {
				case env.Block:
					ser := ps.Ser
					ps.Ser = code.Series
					for i := 0; i < len(block.Rows); i++ {
						ps = EvalBlockInj(ps, block.Rows[i], true)
						if ps.ErrorFlag {
							return ps.Res
						}
						if util.IsTruthy(ps.Res) {
							block.Rows = append(block.Rows[:i], block.Rows[i+1:]...)
							i--
						}
						ps.Ser.Reset()
					}
					ps.Ser = ser
					return block
				default:
					return MakeArgError(ps, 2, []env.Type{env.BlockType}, "purge")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType, env.ListType, env.StringType, env.SpreadsheetType}, "purge")
			}
		},
	},

	"purge!": { // TODO ... doesn't fully work
		Argsn: 2,
		Doc:   "Purges values from a series based on return of a injected code block.",
		Pure:  false,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch wrd := arg1.(type) {
			case env.Word:
				val, found, ctx := ps.Ctx.Get2(wrd.Index)
				if found {
					switch block := val.(type) {
					case env.Block:
						switch code := arg0.(type) {
						case env.Block:
							ser := ps.Ser
							ps.Ser = code.Series
							purged := make([]env.Object, 0)
							for i := 0; i < block.Series.Len(); i++ {
								ps = EvalBlockInj(ps, block.Series.Get(i), true)
								if ps.ErrorFlag {
									return ps.Res
								}
								if util.IsTruthy(ps.Res) {
									purged = append(purged, block.Series.S[i])
									block.Series.S = append(block.Series.S[:i], block.Series.S[i+1:]...)
									i--
								}
								ps.Ser.Reset()
							}
							ps.Ser = ser
							ctx.Set(wrd.Index, block)
							return env.NewBlock(*env.NewTSeries(purged))
						default:
							return MakeArgError(ps, 1, []env.Type{env.BlockType}, "purge!")
						}
					default:
						return MakeBuiltinError(ps, "Context value should be block type.", "purge!")
					}
				} else {
					return MakeBuiltinError(ps, "Word not found in context.", "purge!")
				}
			default:
				return MakeArgError(ps, 2, []env.Type{env.WordType}, "purge!")
			}
		},
	},

	// map should at the end map over block, raw-map, etc ...
	// it should accept a block of code, a function and a builtin
	// it should use injected block so it doesn't need a variable defined like map [ 1 2 3 ] x [ add a 100 ]
	// map [ 1 2 3 ] { .add 3 }
	"map": { // **
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
							if ps.ErrorFlag {
								return ps.Res
							}
							newl[i] = ps.Res
							ps.Ser.Reset()
						}
						ps.Ser = ser
					case env.Builtin:
						for i := 0; i < l; i++ {
							newl[i] = DirectlyCallBuiltin(ps, block, list.Series.Get(i), nil)
						}
					default:
						return MakeBuiltinError(ps, "Block value should be builtin or block type.", "map")
					}
					return *env.NewBlock(*env.NewTSeries(newl))
				default:
					return MakeArgError(ps, 2, []env.Type{env.BlockType, env.BuiltinType}, "map")
				}
			case env.List:
				switch block := arg1.(type) {
				case env.Block, env.Builtin:
					l := len(list.Data)
					newl := make([]any, l)
					switch block := block.(type) {
					case env.Block:
						ser := ps.Ser
						ps.Ser = block.Series
						for i := 0; i < l; i++ {
							ps = EvalBlockInj(ps, env.ToRyeValue(list.Data[i]), true)
							if ps.ErrorFlag {
								return ps.Res
							}
							newl[i] = env.RyeToRaw(ps.Res)
							ps.Ser.Reset()
						}
						ps.Ser = ser
					case env.Builtin:
						for i := 0; i < l; i++ {
							newl[i] = DirectlyCallBuiltin(ps, block, env.ToRyeValue(list.Data[i]), nil)
						}
					default:
						return MakeBuiltinError(ps, "Block value should be builtin or block type.", "map")
					}
					return *env.NewList(newl)
				default:
					return MakeArgError(ps, 2, []env.Type{env.BlockType, env.BuiltinType}, "map")
				}
			case env.String:
				input := []rune(list.Value)
				l := len(input)
				switch block := arg1.(type) {
				case env.Block, env.Builtin:
					newl := make([]env.Object, l)
					switch block := block.(type) {
					case env.Block:
						ser := ps.Ser
						ps.Ser = block.Series
						for i := 0; i < l; i++ {
							ps = EvalBlockInj(ps, *env.NewString(string(input[i])), true)
							if ps.ErrorFlag {
								return ps.Res
							}
							newl[i] = ps.Res

							ps.Ser.Reset()
						}
						ps.Ser = ser
					case env.Builtin:
						for i := 0; i < l; i++ {
							newl[i] = DirectlyCallBuiltin(ps, block, *env.NewString(string(input[i])), nil)
						}
					default:
						return MakeBuiltinError(ps, "Block value should be builtin or block type.", "map")
					}
					return *env.NewBlock(*env.NewTSeries(newl))
				default:
					return MakeArgError(ps, 2, []env.Type{env.BlockType, env.BuiltinType}, "map")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType, env.ListType, env.StringType}, "map")
			}
		},
	},

	"map\\pos": { // TODO -- deduplicate map\pos and map\idx
		Argsn: 3,
		Doc:   "Maps values of a block to a new block by evaluating a block of code.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch list := arg0.(type) {
			case env.Block:
				switch accu := arg1.(type) {
				case env.Word:
					switch block := arg2.(type) {
					case env.Block:
						l := list.Series.Len()
						newl := make([]env.Object, l)
						ser := ps.Ser
						ps.Ser = block.Series
						for i := 0; i < l; i++ {
							ps.Ctx.Set(accu.Index, *env.NewInteger(int64(i + 1)))
							ps = EvalBlockInj(ps, list.Series.Get(i), true)
							if ps.ErrorFlag {
								return ps.Res
							}
							newl[i] = ps.Res
							ps.Ser.Reset()
						}
						ps.Ser = ser
						return *env.NewBlock(*env.NewTSeries(newl))
					default:
						return MakeArgError(ps, 3, []env.Type{env.BlockType}, "map\\pos")
					}
				default:
					return MakeArgError(ps, 2, []env.Type{env.WordType}, "map\\pos")
				}
			case env.List:
				switch block := arg2.(type) {
				case env.Block:
					l := len(list.Data)
					newl := make([]any, l)
					switch accu := arg1.(type) {
					case env.Word:
						ser := ps.Ser
						ps.Ser = block.Series
						for i := 0; i < l; i++ {
							ps.Ctx.Set(accu.Index, *env.NewInteger(int64(i + 1)))
							ps = EvalBlockInj(ps, env.ToRyeValue(list.Data[i]), true)
							if ps.ErrorFlag {
								return ps.Res
							}
							newl[i] = env.RyeToRaw(ps.Res)
							ps.Ser.Reset()
						}
						ps.Ser = ser
						return *env.NewList(newl)
					default:
						return MakeArgError(ps, 2, []env.Type{env.WordType}, "map\\pos")
					}
				default:
					return MakeArgError(ps, 3, []env.Type{env.BlockType}, "map\\pos")
				}
			case env.String:
				input := []rune(list.Value)
				l := len(input)
				switch accu := arg1.(type) {
				case env.Word:
					switch block := arg2.(type) {
					case env.Block:
						newl := make([]env.Object, l)
						ser := ps.Ser
						ps.Ser = block.Series
						for i := 0; i < l; i++ {
							ps.Ctx.Set(accu.Index, *env.NewInteger(int64(i + 1)))
							ps = EvalBlockInj(ps, *env.NewString(string(input[i])), true)
							if ps.ErrorFlag {
								return ps.Res
							}
							newl[i] = ps.Res

							ps.Ser.Reset()
						}
						ps.Ser = ser

						return *env.NewBlock(*env.NewTSeries(newl))
					default:
						return MakeArgError(ps, 3, []env.Type{env.BlockType}, "map\\pos")
					}
				default:
					return MakeArgError(ps, 2, []env.Type{env.WordType}, "map\\pos")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType, env.ListType, env.StringType}, "map\\pos")
			}
		},
	},

	"map\\idx": { // TODO -- deduplicate map\pos and map\idx
		Argsn: 3,
		Doc:   "Maps values of a block to a new block by evaluating a block of code.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch list := arg0.(type) {
			case env.Block:
				switch accu := arg1.(type) {
				case env.Word:
					switch block := arg2.(type) {
					case env.Block:
						l := list.Series.Len()
						newl := make([]env.Object, l)
						ser := ps.Ser
						ps.Ser = block.Series
						for i := 0; i < l; i++ {
							ps.Ctx.Set(accu.Index, *env.NewInteger(int64(i)))
							ps = EvalBlockInj(ps, list.Series.Get(i), true)
							if ps.ErrorFlag {
								return ps.Res
							}
							newl[i] = ps.Res
							ps.Ser.Reset()
						}
						ps.Ser = ser
						return *env.NewBlock(*env.NewTSeries(newl))
					default:
						return MakeArgError(ps, 3, []env.Type{env.BlockType}, "map\\pos")
					}
				default:
					return MakeArgError(ps, 2, []env.Type{env.WordType}, "map\\pos")
				}
			case env.List:
				switch block := arg2.(type) {
				case env.Block:
					l := len(list.Data)
					newl := make([]any, l)
					switch accu := arg1.(type) {
					case env.Word:
						ser := ps.Ser
						ps.Ser = block.Series
						for i := 0; i < l; i++ {
							ps.Ctx.Set(accu.Index, *env.NewInteger(int64(i)))
							ps = EvalBlockInj(ps, env.ToRyeValue(list.Data[i]), true)
							if ps.ErrorFlag {
								return ps.Res
							}
							newl[i] = env.RyeToRaw(ps.Res)
							ps.Ser.Reset()
						}
						ps.Ser = ser
						return *env.NewList(newl)
					default:
						return MakeArgError(ps, 2, []env.Type{env.WordType}, "map\\pos")
					}
				default:
					return MakeArgError(ps, 3, []env.Type{env.BlockType}, "map\\pos")
				}
			case env.String:
				input := []rune(list.Value)
				l := len(input)
				switch accu := arg1.(type) {
				case env.Word:
					switch block := arg2.(type) {
					case env.Block:
						newl := make([]env.Object, l)
						ser := ps.Ser
						ps.Ser = block.Series
						for i := 0; i < l; i++ {
							ps.Ctx.Set(accu.Index, *env.NewInteger(int64(i)))
							ps = EvalBlockInj(ps, *env.NewString(string(input[i])), true)
							if ps.ErrorFlag {
								return ps.Res
							}
							newl[i] = ps.Res

							ps.Ser.Reset()
						}
						ps.Ser = ser

						return *env.NewBlock(*env.NewTSeries(newl))
					default:
						return MakeArgError(ps, 3, []env.Type{env.BlockType}, "map\\pos")
					}
				default:
					return MakeArgError(ps, 2, []env.Type{env.WordType}, "map\\pos")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType, env.ListType, env.StringType}, "map\\pos")
			}
		},
	},

	"reduce": { // **
		Argsn: 3,
		Doc:   "Reduces values of a block to a new block by evaluating a block of code ...",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch list := arg0.(type) {
			case env.Block:
				l := len(list.Series.S)
				if l == 0 {
					return MakeBuiltinError(ps, "Block is empty.", "reduce")
				}
				switch accu := arg1.(type) {
				case env.Word:
					// ps.Ctx.Set(accu.Index)
					switch block := arg2.(type) {
					case env.Block:
						acc := list.Series.Get(0)
						ser := ps.Ser
						ps.Ser = block.Series
						for i := 1; i < l; i++ {
							ps.Ctx.Set(accu.Index, acc)
							ps = EvalBlockInj(ps, list.Series.Get(i), true)
							if ps.ErrorFlag {
								return ps.Res
							}
							if ps.ErrorFlag {
								return ps.Res
							}
							acc = ps.Res
							ps.Ser.Reset()
						}
						ps.Ser = ser
						return acc
					default:
						return MakeArgError(ps, 3, []env.Type{env.BlockType}, "reduce")
					}
				default:
					return MakeArgError(ps, 2, []env.Type{env.WordType}, "reduce")
				}
			case env.List:
				l := len(list.Data)
				if l == 0 {
					return MakeBuiltinError(ps, "List is empty.", "reduce")
				}
				switch accu := arg1.(type) {
				case env.Word:
					// ps.Ctx.Set(accu.Index)
					switch block := arg2.(type) {
					case env.Block:
						acc := env.ToRyeValue(list.Data[0])
						ser := ps.Ser
						ps.Ser = block.Series
						for i := 1; i < l; i++ {
							ps.Ctx.Set(accu.Index, acc)
							ps = EvalBlockInj(ps, env.ToRyeValue(list.Data[i]), true)
							if ps.ErrorFlag {
								return ps.Res
							}
							acc = ps.Res
							ps.Ser.Reset()
						}
						ps.Ser = ser
						return acc
					default:
						return MakeArgError(ps, 3, []env.Type{env.BlockType}, "reduce")
					}
				default:
					return MakeArgError(ps, 2, []env.Type{env.WordType}, "reduce")
				}
			case env.String:
				if len(list.Value) == 0 {
					return MakeBuiltinError(ps, "String is empty.", "reduce")
				}
				switch accu := arg1.(type) {
				case env.Word:
					switch block := arg2.(type) {
					case env.Block:
						input := []rune(list.Value)
						var acc env.Object
						acc = *env.NewString(string(input[0]))
						ser := ps.Ser
						ps.Ser = block.Series
						for i := 1; i < len(input); i++ {
							ps.Ctx.Set(accu.Index, acc)
							ps = EvalBlockInj(ps, *env.NewString(string(input[i])), true)
							if ps.ErrorFlag {
								return ps.Res
							}
							acc = ps.Res
							ps.Ser.Reset()
						}
						ps.Ser = ser
						return acc
					default:
						return MakeArgError(ps, 3, []env.Type{env.BlockType}, "reduce")
					}
				default:
					return MakeArgError(ps, 2, []env.Type{env.WordType}, "reduce")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType, env.ListType, env.StringType}, "reduce")
			}
		},
	},

	// map should at the end map over block, raw-map, etc ...
	// it should accept a block of code, a function and a builtin
	// it should use injected block so it doesn't need a variable defined like map [ 1 2 3 ] x [ add a 100 ]
	// reduce [ 1 2 3 ] 'acc { + acc }
	"fold": { // **
		Argsn: 4,
		Doc:   "Reduces values of a block to a new block by evaluating a block of code ...",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch list := arg0.(type) {
			case env.Block:
				switch accu := arg1.(type) {
				case env.Word:
					// ps.Ctx.Set(accu.Index)
					switch block := arg3.(type) {
					case env.Block:
						l := len(list.Series.S)
						acc := arg2
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
						return acc
					default:
						return MakeArgError(ps, 4, []env.Type{env.BlockType}, "fold")
					}
				default:
					return MakeArgError(ps, 2, []env.Type{env.WordType}, "fold")
				}
			case env.List:
				switch accu := arg1.(type) {
				case env.Word:
					// ps.Ctx.Set(accu.Index)
					switch block := arg3.(type) {
					case env.Block:
						l := len(list.Data)
						acc := arg2
						ser := ps.Ser
						ps.Ser = block.Series
						for i := 0; i < l; i++ {
							ps.Ctx.Set(accu.Index, acc)
							ps = EvalBlockInj(ps, env.ToRyeValue(list.Data[i]), true)
							if ps.ErrorFlag {
								return ps.Res
							}
							acc = ps.Res
							ps.Ser.Reset()
						}
						ps.Ser = ser
						return acc
					default:
						return MakeArgError(ps, 4, []env.Type{env.BlockType}, "fold")
					}
				default:
					return MakeArgError(ps, 2, []env.Type{env.WordType}, "fold")
				}
			case env.String:
				switch accu := arg1.(type) {
				case env.Word:
					switch block := arg3.(type) {
					case env.Block:
						input := []rune(list.Value)
						var acc env.Object
						acc = arg2
						ser := ps.Ser
						ps.Ser = block.Series
						for i := 0; i < len(input); i++ {
							ps.Ctx.Set(accu.Index, acc)
							ps = EvalBlockInj(ps, *env.NewString(string(input[i])), true)
							if ps.ErrorFlag {
								return ps.Res
							}
							acc = ps.Res
							ps.Ser.Reset()
						}
						ps.Ser = ser
						return acc
					default:
						return MakeArgError(ps, 4, []env.Type{env.BlockType}, "fold")
					}
				default:
					return MakeArgError(ps, 2, []env.Type{env.WordType}, "fold")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType, env.ListType, env.StringType}, "fold")
			}
		},
	},

	"sum-up": { // **
		Argsn: 2,
		Doc:   "Reduces values of a block or list by evaluating a block of code and summing the values.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			var ll []any
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
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType, env.ListType}, "sum-up")
			}

			switch block := arg1.(type) {
			case env.Block, env.Builtin:
				acc := *env.NewDecimal(0)
				onlyInts := true
				switch block := block.(type) {
				case env.Block:
					ser := ps.Ser
					ps.Ser = block.Series
					for i := 0; i < llen; i++ {
						var item any
						if modeObj == 1 {
							item = ll[i]
						} else {
							item = lo[i]
						}
						// ps.Ctx.Set(accu.Index, acc)
						ps = EvalBlockInj(ps, env.ToRyeValue(item), true)
						if ps.ErrorFlag {
							return ps.Res
						}
						switch res := ps.Res.(type) {
						case env.Integer:
							acc.Value += float64(res.Value)
						case env.Decimal:
							onlyInts = false
							acc.Value += res.Value
						default:
							return MakeBuiltinError(ps, "Block should return integer or decimal.", "sum-up")
						}
						ps.Ser.Reset()
					}
					ps.Ser = ser
				case env.Builtin:
					for i := 0; i < llen; i++ {
						var item any
						if modeObj == 1 {
							item = ll[i]
						} else {
							item = lo[i]
						}
						res := DirectlyCallBuiltin(ps, block, env.ToRyeValue(item), nil)
						switch res := res.(type) {
						case env.Integer:
							acc.Value += float64(res.Value)
						case env.Decimal:
							onlyInts = false
							acc.Value += res.Value
						default:
							return MakeBuiltinError(ps, "Block should return integer or decimal.", "sum-up")
						}
					}
				default:
					return MakeBuiltinError(ps, "Block type should be Builtin or Block.", "sum-up")
				}
				if onlyInts {
					return *env.NewInteger(int64(acc.Value))
				} else {
					return acc
				}
			default:
				return MakeArgError(ps, 2, []env.Type{env.BlockType, env.BuiltinType}, "sum-up")
			}
		},
	},

	"partition": { // **
		Argsn: 2,
		Doc:   "Partitions a series by evaluating a block of code.",
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
							if prevres == nil || ps.Res.Equal(prevres) {
								subl = append(subl, curval)
							} else {
								newl = append(newl, *env.NewBlock(*env.NewTSeries(subl)))
								subl = []env.Object{curval}
							}
							prevres = ps.Res
							ps.Ser.Reset()
						}
						newl = append(newl, *env.NewBlock(*env.NewTSeries(subl)))
						ps.Ser = ser
					case env.Builtin:
						for i := 0; i < l; i++ {
							curval := list.Series.Get(i)
							res := DirectlyCallBuiltin(ps, block, curval, nil)
							if prevres == nil || res.Equal(prevres) {
								subl = append(subl, curval)
							} else {
								newl = append(newl, *env.NewBlock(*env.NewTSeries(subl)))
								subl = []env.Object{curval}
							}
							prevres = res
						}
						newl = append(newl, *env.NewBlock(*env.NewTSeries(subl)))
					default:
						return MakeBuiltinError(ps, "Block type should be Builtin or Block.", "partition")
					}
					return *env.NewBlock(*env.NewTSeries(newl))
				default:
					return MakeArgError(ps, 2, []env.Type{env.BlockType, env.BuiltinType}, "partition")
				}
			case env.List:
				switch block := arg1.(type) {
				case env.Block, env.Builtin:
					l := len(list.Data)
					newl := make([]any, 0)
					subl := make([]any, 0)
					var prevres env.Object
					switch block := block.(type) {
					case env.Block:
						ser := ps.Ser
						ps.Ser = block.Series
						for i := 0; i < l; i++ {
							curval := list.Data[i]
							curvalRye := env.ToRyeValue(list.Data[i])
							ps = EvalBlockInj(ps, curvalRye, true)
							if ps.ErrorFlag {
								return ps.Res
							}
							if prevres == nil || ps.Res.Equal(prevres) {
								subl = append(subl, curval)
							} else {
								newl = append(newl, env.NewList(subl))
								subl = []any{curval}
							}
							prevres = ps.Res
							ps.Ser.Reset()
						}
						newl = append(newl, env.NewList(subl))
						ps.Ser = ser
					case env.Builtin:
						for i := 0; i < l; i++ {
							curval := list.Data[i]
							curvalRye := env.ToRyeValue(list.Data[i])
							res := DirectlyCallBuiltin(ps, block, curvalRye, nil)
							if prevres == nil || res.Equal(prevres) {
								subl = append(subl, curval)
							} else {
								newl = append(newl, env.NewList(subl))
								subl = []any{curval}
							}
							prevres = res
						}
						newl = append(newl, env.NewList(subl))
					default:
						return MakeBuiltinError(ps, "Block type should be Builtin or Block.", "partition")
					}
					return *env.NewList(newl)
				default:
					return MakeArgError(ps, 2, []env.Type{env.BlockType, env.BuiltinType}, "partition")
				}
			case env.String:
				switch block := arg1.(type) {
				case env.Block, env.Builtin:
					newl := make([]any, 0)
					var subl strings.Builder
					var prevres env.Object
					switch block := block.(type) {
					case env.Block:
						ser := ps.Ser
						ps.Ser = block.Series
						for _, curval := range list.Value {
							ps = EvalBlockInj(ps, *env.NewString(string(curval)), true)
							if ps.ErrorFlag {
								return ps.Res
							}
							if prevres == nil || ps.Res.Equal(prevres) {
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
						for _, curval := range list.Value {
							res := DirectlyCallBuiltin(ps, block, env.ToRyeValue(curval), nil)
							if prevres == nil || res.Equal(prevres) {
								subl.WriteRune(curval)
							} else {
								newl = append(newl, subl.String())
							}
						}
						newl = append(newl, subl.String())
					default:
						return MakeBuiltinError(ps, "Block type should be Builtin or Block.", "partition")
					}
					return *env.NewList(newl)
				default:
					return MakeArgError(ps, 2, []env.Type{env.BlockType, env.BuiltinType}, "partition")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType, env.ListType, env.StringType}, "partition")
			}
		},
	},

	"group": { // **
		Argsn: 2,
		Doc:   "Groups a block or list of values given condition.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			var ll []any
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
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType, env.ListType}, "group")
			}

			switch block := arg1.(type) {
			case env.Block, env.Builtin:
				newd := make(map[string]any)
				switch block := block.(type) {
				case env.Block:
					ser := ps.Ser
					ps.Ser = block.Series
					for i := 0; i < llen; i++ {
						var curval env.Object
						if modeObj == 1 {
							curval = env.ToRyeValue(ll[i])
						} else {
							curval = lo[i]
						}
						ps = EvalBlockInj(ps, curval, true)
						if ps.ErrorFlag {
							return ps.Res
						}
						// TODO !!! -- currently only works if results are strings
						newkeyStr, ok := ps.Res.(env.String)
						if !ok {
							return MakeBuiltinError(ps, "Grouping key should be string.", "group")
						}
						newkey := newkeyStr.Value
						entry, ok := newd[newkey]
						if !ok {
							newd[newkey] = env.NewList(make([]any, 0))
							entry, ok = newd[newkey]
							if !ok {
								return MakeBuiltinError(ps, "Key not found in List.", "group")
							}
						}
						switch ee := entry.(type) { // list in dict is a pointer
						case *env.List:
							ee.Data = append(ee.Data, env.RyeToRaw(curval))
						default:
							return MakeBuiltinError(ps, "Entry type should be List.", "group")
						}
						ps.Ser.Reset()
					}
					ps.Ser = ser
				case env.Builtin:
					for i := 0; i < llen; i++ {
						var curval env.Object
						if modeObj == 1 {
							curval = env.ToRyeValue(ll[i])
						} else {
							curval = lo[i]
						}
						res := DirectlyCallBuiltin(ps, block, curval, nil)
						// TODO !!! -- currently only works if results are strings
						newkeyStr, ok := res.(env.String)
						if !ok {
							return MakeBuiltinError(ps, "Grouping key should be string.", "group")
						}
						newkey := newkeyStr.Value
						entry, ok := newd[newkey]
						if !ok {
							newd[newkey] = env.NewList(make([]any, 0))
							entry, ok = newd[newkey]
							if !ok {
								return MakeBuiltinError(ps, "Key not found in List.", "group")
							}
						}
						switch ee := entry.(type) { // list in dict is a pointer
						case *env.List:
							ee.Data = append(ee.Data, env.RyeToRaw(curval))
						default:
							return MakeBuiltinError(ps, "Entry type should be List.", "group")
						}
					}
				default:
					return MakeBuiltinError(ps, "Block must be type of Block or builtin.", "group")
				}
				return *env.NewDict(newd)
			default:
				return MakeArgError(ps, 2, []env.Type{env.BlockType, env.BuiltinType}, "group")
			}
		},
	},

	// filter [ 1 2 3 ] { .add 3 }
	"filter": { // **
		Argsn: 2,
		Doc:   "Filters values from a seris based on return of a injected code block.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			var ll []any
			var lo []env.Object
			var ls []rune
			var llen int
			modeObj := 0
			switch data := arg0.(type) {
			case env.String:
				ls = []rune(data.Value)
				llen = len(ls)
				modeObj = 3
			case env.Block:
				lo = data.Series.S
				llen = len(lo)
				modeObj = 2
			case env.List:
				ll = data.Data
				llen = len(ll)
				modeObj = 1
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType, env.ListType, env.StringType}, "filter")
			}

			switch block := arg1.(type) {
			case env.Block, env.Builtin:
				var newlo []env.Object
				var newll []any
				switch block := block.(type) {
				case env.Block:
					ser := ps.Ser
					ps.Ser = block.Series
					for i := 0; i < llen; i++ {
						var item any
						if modeObj == 1 {
							item = ll[i]
						} else if modeObj == 2 {
							item = lo[i]
						} else {
							item = env.ToRyeValue(ls[i])
						}
						ps = EvalBlockInj(ps, env.ToRyeValue(item), true)
						if ps.ErrorFlag {
							return ps.Res
						}
						if util.IsTruthy(ps.Res) { // todo -- move these to util or something
							if modeObj == 1 {
								newll = append(newll, ll[i])
							} else if modeObj == 2 {
								newlo = append(newlo, lo[i])
							} else {
								newlo = append(newlo, item.(env.Object))
							}
						}
						ps.Ser.Reset()
					}
					ps.Ser = ser
				case env.Builtin:
					for i := 0; i < llen; i++ {
						var item any
						if modeObj == 1 {
							item = ll[i]
						} else if modeObj == 2 {
							item = lo[i]
						} else {
							item = env.ToRyeValue(ls[i])
						}
						res := DirectlyCallBuiltin(ps, block, env.ToRyeValue(item), nil)
						if util.IsTruthy(res) { // todo -- move these to util or something
							if modeObj == 1 {
								newll = append(newll, ll[i])
							} else if modeObj == 2 {
								newlo = append(newlo, lo[i])
							} else {
								newlo = append(newlo, item.(env.Object))
							}
						}
					}
				default:
					return MakeBuiltinError(ps, "Block type should be Builtin or Block.", "filter")
				}
				if modeObj == 1 {
					return *env.NewList(newll)
				} else if modeObj == 2 {
					return *env.NewBlock(*env.NewTSeries(newlo))
				} else {
					return *env.NewBlock(*env.NewTSeries(newlo))
				}

			default:
				return MakeArgError(ps, 2, []env.Type{env.BlockType, env.BuiltinType}, "filter")
			}
		},
	},

	"seek": { // **
		Argsn: 2,
		Doc:   "Seek over a series until a Block of code returns True and return the value.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			var ll []any
			var lo []env.Object
			var ls []rune
			var llen int
			modeObj := 0
			switch data := arg0.(type) {
			case env.String:
				ls = []rune(data.Value)
				llen = len(ls)
				modeObj = 3
			case env.Block:
				lo = data.Series.S
				llen = len(lo)
				modeObj = 2
			case env.List:
				ll = data.Data
				llen = len(ll)
				modeObj = 1
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType, env.ListType, env.StringType}, "seek")
			}
			switch block := arg1.(type) {
			case env.Block, env.Builtin:
				switch block := block.(type) {
				case env.Block:
					ser := ps.Ser
					ps.Ser = block.Series
					for i := 0; i < llen; i++ {
						var item any
						if modeObj == 1 {
							item = ll[i]
						} else if modeObj == 2 {
							item = lo[i]
						} else {
							item = *env.NewString(string(ls[i]))
						}
						ps = EvalBlockInj(ps, env.ToRyeValue(item), true)
						if ps.ErrorFlag {
							return ps.Res
						}
						if util.IsTruthy(ps.Res) { // todo -- move these to util or something
							return env.ToRyeValue(item)
						}
						ps.Ser.Reset()
					}
					ps.Ser = ser
				case env.Builtin:
					for i := 0; i < llen; i++ {
						var item any
						if modeObj == 1 {
							item = ll[i]
						} else if modeObj == 2 {
							item = lo[i]
						} else {
							item = *env.NewString(string(ls[i]))
						}
						res := DirectlyCallBuiltin(ps, block, env.ToRyeValue(item), nil)
						if util.IsTruthy(res) { // todo -- move these to util or something
							return env.ToRyeValue(item)
						}
					}
				default:
					ps.ErrorFlag = true
					return MakeBuiltinError(ps, "Second argument should be block, builtin (or function).", "seek")
				}
			default:
				return MakeArgError(ps, 2, []env.Type{env.BlockType, env.BuiltinType}, "seek")
			}
			return MakeBuiltinError(ps, "No element found.", "seek")
		},
	},

	// collections exploration functions

	"max": { // **
		Argsn: 1,
		Doc:   "Accepts a Block or List of values and returns the maximal value.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch data := arg0.(type) {
			case env.Block:
				var max env.Object
				l := data.Series.Len()
				if l == 0 {
					return MakeBuiltinError(ps, "Block is empty.", "max")
				}
				for i := 0; i < l; i++ {
					if max == nil || greaterThan(ps, data.Series.Get(i), max) {
						max = data.Series.Get(i)
					}
				}
				return max
			case env.List:
				max := math.SmallestNonzeroFloat64
				l := len(data.Data)
				if l == 0 {
					return MakeBuiltinError(ps, "List is empty.", "max")
				}
				var isMaxInt bool
				for i := 0; i < l; i++ {
					switch val1 := data.Data[i].(type) {
					case int64:
						if float64(val1) > max {
							max = float64(val1)
							isMaxInt = true
						}
					case float64:
						if val1 > max {
							max = val1
							isMaxInt = false
						}
					default:
						return MakeBuiltinError(ps, "List type should be Integer or Decimal.", "max")
					}
				}
				if isMaxInt {
					return *env.NewInteger(int64(max))
				} else {
					return *env.NewDecimal(max)
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType, env.ListType}, "max")
			}
		},
	},

	"min": { // **
		Argsn: 1,
		Doc:   "Accepts a Block or List of values and returns the minimal value.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch data := arg0.(type) {
			case env.Block:
				var min env.Object
				l := data.Series.Len()
				if l == 0 {
					return MakeBuiltinError(ps, "Block is empty.", "min")
				}
				for i := 0; i < l; i++ {
					if min == nil || greaterThan(ps, min, data.Series.Get(i)) {
						min = data.Series.Get(i)
					}
				}
				return min
			case env.List:
				l := len(data.Data)
				if l == 0 {
					return MakeBuiltinError(ps, "List is empty.", "min")
				}
				var isMinInt bool
				min := math.MaxFloat64
				for i := 0; i < l; i++ {
					switch val1 := data.Data[i].(type) {
					case int64:
						if float64(val1) < min {
							min = float64(val1)
							isMinInt = true
						}
					case float64:
						if val1 < min {
							min = val1
							isMinInt = false
						}
					case env.Integer: // TODO -- think about what values really List should hold and when / how it should be used
						if float64(val1.Value) < min {
							min = float64(val1.Value)
							isMinInt = true
						}
					case *env.Integer: // TODO -- think about what values really List should hold and when / how it should be used
						if float64(val1.Value) < min {
							min = float64(val1.Value)
							isMinInt = true
						}
					default:
						fmt.Println(data.Data[i])
						fmt.Printf("t1: %T\n", data.Data[i])
						return MakeBuiltinError(ps, "List type should be Integer or Decimal.", "min")
					}
				}
				if isMinInt {
					return *env.NewInteger(int64(min))
				} else {
					return *env.NewDecimal(min)
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType}, "min")
			}
		},
	},

	"avg": { // **
		Argsn: 1,
		Doc:   "Accepts a Block or List of values and returns the average value.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			var sum float64
			switch block := arg0.(type) {
			case env.Block:
				l := block.Series.Len()
				if l == 0 {
					return MakeBuiltinError(ps, "Block is empty.", "avg")
				}
				for i := 0; i < l; i++ {
					obj := block.Series.Get(i)
					switch val1 := obj.(type) {
					case env.Integer:
						sum += float64(val1.Value)
					case env.Decimal:
						sum += val1.Value
					default:
						return MakeBuiltinError(ps, "Block type should be Integer or Decimal.", "avg")
					}
				}
				return *env.NewDecimal(sum / float64(l))
			case env.List:
				l := len(block.Data)
				if l == 0 {
					return MakeBuiltinError(ps, "List is empty.", "avg")
				}
				for i := 0; i < l; i++ {
					obj := block.Data[i]
					switch val1 := obj.(type) {
					case int64:
						sum += float64(val1)
					case float64:
						sum += val1
					default:
						return MakeBuiltinError(ps, "List type should be Integer or Decimal.", "avg")
					}
				}
				return *env.NewDecimal(sum / float64(l))
			case env.Vector:
				return *env.NewDecimal(block.Value.Mean())
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType, env.VectorType}, "avg")
			}
		},
	},

	"sum": { // **
		Argsn: 1,
		Doc:   "Accepts a Block or List of values and returns the sum.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			var sum float64
			switch block := arg0.(type) {
			case env.Block:
				l := block.Series.Len()
				onlyInts := true
				for i := 0; i < l; i++ {
					obj := block.Series.Get(i)
					switch val1 := obj.(type) {
					case env.Integer:
						sum += float64(val1.Value)
					case env.Decimal:
						sum += val1.Value
						onlyInts = false
					default:
						return MakeBuiltinError(ps, "Block type should be Integer or Decimal.", "sum")
					}
				}
				if onlyInts {
					return *env.NewInteger(int64(sum))
				} else {
					return *env.NewDecimal(sum)
				}
			case env.List:
				l := len(block.Data)
				onlyInts := true
				for i := 0; i < l; i++ {
					obj := block.Data[i]
					switch val1 := obj.(type) {
					case int64:
						sum += float64(val1)
					case float64:
						sum += val1
						onlyInts = false
					default:
						return MakeBuiltinError(ps, "List type should be Integer or Decimal.", "sum")
					}
				}
				if onlyInts {
					return *env.NewInteger(int64(sum))
				} else {
					return *env.NewDecimal(sum)
				}

			case env.Vector:
				return *env.NewDecimal(block.Value.Sum())
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType, env.VectorType}, "sum")
			}
		},
	},

	"mul": { // **
		Argsn: 1,
		Doc:   "Accepts a Block or List of values and returns the sum.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			var sum float64 = 1
			switch block := arg0.(type) {
			case env.Block:
				l := block.Series.Len()
				onlyInts := true
				for i := 0; i < l; i++ {
					obj := block.Series.Get(i)
					switch val1 := obj.(type) {
					case env.Integer:
						sum *= float64(val1.Value)
					case env.Decimal:
						sum *= val1.Value
						onlyInts = false
					default:
						return MakeBuiltinError(ps, "Block type should be Integer or Decimal.", "sum")
					}
				}
				if onlyInts {
					return *env.NewInteger(int64(sum))
				} else {
					return *env.NewDecimal(sum)
				}
			case env.List:
				l := len(block.Data)
				onlyInts := true
				for i := 0; i < l; i++ {
					obj := block.Data[i]
					switch val1 := obj.(type) {
					case int64:
						sum *= float64(val1)
					case float64:
						sum *= val1
						onlyInts = false
					default:
						return MakeBuiltinError(ps, "List type should be Integer or Decimal.", "sum")
					}
				}
				if onlyInts {
					return *env.NewInteger(int64(sum))
				} else {
					return *env.NewDecimal(sum)
				}

			case env.Vector:
				return *env.NewDecimal(block.Value.Sum())
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType, env.VectorType}, "sum")
			}
		},
	},

	"sort!": { // **
		Argsn: 1,
		Doc:   "Accepts a block or list and sorts in place in ascending order and returns it.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch block := arg0.(type) {
			case env.Block:
				ss := block.Series.S
				sort.Sort(RyeBlockSort(ss))
				return *env.NewBlock(*env.NewTSeries(ss))
			case env.List:
				ss := block.Data
				sort.Sort(RyeListSort(ss))
				return *env.NewList(ss)
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType, env.ListType}, "sort!")
			}
		},
	},

	"unique": { // **
		Argsn: 1,
		Doc:   "Accepts a block or list of values and returns only unique values.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch block := arg0.(type) {
			case env.List:
				ss := block.Data

				// Create a map to store the unique values.
				// uniqueValues := make(map[string]bool)
				uniqueValues := make(map[any]bool)

				// Iterate over the slice and add the elements to the map.
				for _, element := range ss {
					// uniqueValues[env.ToRyeValue(element).Print(*ps.Idx)] = true
					uniqueValues[element] = true
				}

				// Create a new slice to store the unique values.
				uniqueSlice := make([]any, 0, len(uniqueValues))

				// Iterate over the map and add the keys to the new slice.
				for key := range uniqueValues {
					uniqueSlice = append(uniqueSlice, key)
				}
				return *env.NewList(uniqueSlice)
			case env.Block:
				uniqueList := util.RemoveDuplicate(ps, block.Series.S)
				return *env.NewBlock(*env.NewTSeries(uniqueList))
			case env.String:
				strSlice := make([]env.Object, 0)
				// create string to object slice
				for _, value := range block.Value {
					// if want to block  space then we can add here condition
					strSlice = append(strSlice, env.ToRyeValue(value))
				}
				uniqueStringSlice := util.RemoveDuplicate(ps, strSlice)
				uniqueStr := ""
				// converting object to string and append final
				for _, value := range uniqueStringSlice {
					uniqueStr = uniqueStr + env.RyeToRaw(value).(string)
				}
				return *env.NewString(uniqueStr)
			default:
				return MakeArgError(ps, 1, []env.Type{env.ListType, env.BlockType, env.StringType}, "unique")
			}
		},
	},

	"reverse!": { // **
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
			case env.String:
				s := block.Value
				reversed := ""
				for i := len(s) - 1; i >= 0; i-- {
					reversed += string(s[i])
				}
				return *env.NewString(reversed)
			case env.List:
				// Create slice of env.Object
				dataSlice := make([]env.Object, 0)
				for _, v := range block.Data {
					dataSlice = append(dataSlice, env.ToRyeValue(v))
				}
				// Reverse slice data
				for left, right := 0, len(dataSlice)-1; left < right; left, right = left+1, right-1 {
					dataSlice[left], dataSlice[right] = dataSlice[right], dataSlice[left]
				}
				// Create list frol slice data
				reverseList := make([]any, 0, len(dataSlice))
				for _, value := range dataSlice {
					reverseList = append(reverseList, value)
				}
				return *env.NewList(reverseList)
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType, env.StringType, env.ListType}, "reverse!")
			}
		},
	},

	// add distinct? and count? functions
	// make functions work with list, which column and row can return

	// end of collections exploration

	"recur-if": { //recur1-if
		Argsn: 2,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch cond := arg0.(type) {
			case env.Integer:
				if cond.Value > 0 {
					ps.Ser.Reset()
					return nil
				} else {
					return ps.Res
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType}, "recur-if\\1")
			}
		},
	},
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
					default:
						return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "recur-if\\1")
					}
				} else {
					return ps.Res
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType}, "recur-if\\1")
			}
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
						default:
							return MakeArgError(ps, 3, []env.Type{env.IntegerType}, "recur-if\\2")
						}
					default:
						return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "recur-if\\2")
					}
				} else {
					return ps.Res
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType}, "recur-if\\2")
			}
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
						default:
							return MakeArgError(ps, 3, []env.Type{env.IntegerType}, "recur-if\\3")
						}
					default:
						return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "recur-if\\3")
					}
				} else {
					return ps.Res
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType}, "recur-if\\3")
			}
			return nil
		},
	},

	"does": { // **
		Argsn: 1,
		Doc:   "Creates a function without arguments.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch body := arg0.(type) {
			case env.Block:
				//spec := []env.Object{*env.NewWord(aaaidx)}
				//body := []env.Object{*env.NewWord(printidx), *env.NewWord(aaaidx), *env.NewWord(recuridx), *env.NewWord(greateridx), *env.NewInteger(99), *env.NewWord(aaaidx), *env.NewWord(incidx), *env.NewWord(aaaidx)}
				return *env.NewFunction(*env.NewBlock(*env.NewTSeries(make([]env.Object, 0))), body, false)
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType}, "does")
			}
		},
	},

	"fn1": { // **
		Argsn: 1,
		Doc:   "Creates a function that accepts one anonymouse argument.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch body := arg0.(type) {
			case env.Block:
				spec := []env.Object{*env.NewWord(1)}
				//body := []env.Object{*env.NewWord(printidx), *env.NewWord(aaaidx), *env.NewWord(recuridx), *env.NewWord(greateridx), *env.NewInteger(99), *env.NewWord(aaaidx), *env.NewWord(incidx), *env.NewWord(aaaidx)}
				return *env.NewFunction(*env.NewBlock(*env.NewTSeries(spec)), body, false)
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType}, "fn1")
			}
		},
	},

	"fn": {
		Argsn: 2,
		Doc:   "Creates a function.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch args := arg0.(type) {
			case env.Block:
				var doc string
				if args.Series.Len() > 0 {
					var hasDoc bool
					switch a := args.Series.S[len(args.Series.S)-1].(type) {
					case env.String:
						doc = a.Value
						hasDoc = true
						//fmt.Println("DOC DOC")
						// default:
						//return MakeBuiltinError(ps, "Series type should be string.", "fn")
					}
					for i, o := range args.Series.GetAll() {
						if i == len(args.Series.S)-1 && hasDoc {
							break
						}
						if o.Type() != env.WordType {
							return MakeBuiltinError(ps, "Function arguments should be words", "fn")
						}
					}
				}
				switch body := arg1.(type) {
				case env.Block:
					//spec := []env.Object{*env.NewWord(aaaidx)}
					//body := []env.Object{*env.NewWord(printidx), *env.NewWord(aaaidx), *env.NewWord(recuridx), *env.NewWord(greateridx), *env.NewInteger(99), *env.NewWord(aaaidx), *env.NewWord(incidx), *env.NewWord(aaaidx)}
					// fmt.Println(doc)
					return *env.NewFunctionDoc(args, body, false, doc)
				default:
					return MakeArgError(ps, 2, []env.Type{env.BlockType}, "fn")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType}, "fn")
			}
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
				default:
					return MakeArgError(ps, 2, []env.Type{env.BlockType}, "pfn")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType}, "pfn")
			}
		},
	},

	"fnc": { // TODO -- fnc will maybe become fn\par context is set as parrent, fn\in will be executed directly in context
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
						return *env.NewFunctionC(args, body, &ctx, false, false)
					default:
						ps.ErrorFlag = true
						return MakeArgError(ps, 3, []env.Type{env.BlockType}, "fnc")
					}
				default:
					ps.ErrorFlag = true
					return MakeArgError(ps, 2, []env.Type{env.CtxType}, "fnc")
				}
			default:
				ps.ErrorFlag = true
				return MakeArgError(ps, 1, []env.Type{env.BlockType}, "fnc")
			}
		},
	},

	"fn\\cc": {
		// a function with context	 bb: 10 add10 [ a ] context [ b: bb ] [ add a b ]
		// 							add10 [ a ] this [ add a b ]
		// later maybe			   add10 [ a ] [ b: b ] [ add a b ]
		//  						   add10 [ a ] [ 'b ] [ add a b ]
		Argsn: 2,
		Doc:   "Creates a function with specific context.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch args := arg0.(type) {
			case env.Block:
				switch body := arg1.(type) {
				case env.Block:
					return *env.NewFunctionC(args, body, ps.Ctx, false, false)
				default:
					ps.ErrorFlag = true
					return MakeArgError(ps, 2, []env.Type{env.BlockType}, "fnc")
				}
			default:
				ps.ErrorFlag = true
				return MakeArgError(ps, 1, []env.Type{env.BlockType}, "fnc")
			}
		},
	},

	"fn\\par": {
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
						return *env.NewFunctionC(args, body, &ctx, false, false)
					default:
						ps.ErrorFlag = true
						return MakeArgError(ps, 3, []env.Type{env.BlockType}, "fnc")
					}
				default:
					ps.ErrorFlag = true
					return MakeArgError(ps, 2, []env.Type{env.CtxType}, "fnc")
				}
			default:
				ps.ErrorFlag = true
				return MakeArgError(ps, 1, []env.Type{env.BlockType}, "fnc")
			}
		},
	},

	"fn\\in": {
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
						return *env.NewFunctionC(args, body, &ctx, false, true)
					default:
						ps.ErrorFlag = true
						return MakeArgError(ps, 3, []env.Type{env.BlockType}, "fnc")
					}
				default:
					ps.ErrorFlag = true
					return MakeArgError(ps, 2, []env.Type{env.CtxType}, "fnc")
				}
			default:
				ps.ErrorFlag = true
				return MakeArgError(ps, 1, []env.Type{env.BlockType}, "fnc")
			}
		},
	},

	"closure": {
		// a function with context	 bb: 10 add10 [ a ] context [ b: bb ] [ add a b ]
		// 							add10 [ a ] this [ add a b ]
		// later maybe			   add10 [ a ] [ b: b ] [ add a b ]
		//  						   add10 [ a ] [ 'b ] [ add a b ]
		Argsn: 2,
		Doc:   "Creates a function with specific context.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			ctx := ps.Ctx
			switch args := arg0.(type) {
			case env.Block:
				switch body := arg1.(type) {
				case env.Block:
					return *env.NewFunctionC(args, body, ctx, false, false)
				default:
					ps.ErrorFlag = true
					return MakeArgError(ps, 2, []env.Type{env.BlockType}, "fnc")
				}
			default:
				ps.ErrorFlag = true
				return MakeArgError(ps, 1, []env.Type{env.BlockType}, "closure")
			}
		},
	},

	"kind": {
		Argsn: 2,
		Doc:   "Creates new kind.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Word:
				switch spec := arg1.(type) {
				case env.Block:
					return *env.NewKind(s1, spec)
				default:
					return MakeArgError(ps, 2, []env.Type{env.BlockType}, "kind")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.WordType}, "kind")
			}
		},
	},

	"_>>": {
		Argsn: 2,
		Doc:   "Converts first argument to a specific kind.",
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
							return MakeBuiltinError(ps, "Conversion value isn't Dict.", "_>>")
						}
					}
					return MakeBuiltinError(ps, "Conversion value isn't Dict.", "_>>")
				default:
					return MakeArgError(ps, 1, []env.Type{env.DictType, env.KindType}, "_>>")
				}
			default:
				return MakeArgError(ps, 2, []env.Type{env.KindType}, "_>>")
			}
		},
	},

	"_<<": {
		Argsn: 2,
		Doc:   "Converts a value to specific kind (R to L)",
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
						return MakeBuiltinError(ps, "Conversion value isn't Dict.", "_<<")
					}
				case env.RyeCtx:
					if spec.HasConverter(dict.Kind.Index) {
						obj := BuiConvert(ps, dict, spec.Converters[dict.Kind.Index])
						switch ctx := obj.(type) {
						case env.RyeCtx:
							ctx.Kind = spec.Kind
							return ctx
						default:
							return MakeBuiltinError(ps, "Conversion value isn't Dict.", "_<<")
						}
					}
					return MakeBuiltinError(ps, "Conversion value isn't Dict.", "_<<")
				default:
					return MakeArgError(ps, 2, []env.Type{env.DictType, env.CtxType}, "_<<")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.KindType}, "_<<")
			}
		},
	},

	"assure-kind": {
		Argsn: 2,
		Doc:   "Assuring kind.",
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
						return MakeBuiltinError(ps, "Object type is not Dict.", "assure-kind")
					}
				default:
					return MakeArgError(ps, 1, []env.Type{env.DictType}, "assure-kind")
				}
			default:
				return MakeArgError(ps, 2, []env.Type{env.KindType}, "assure-kind")
			}
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
					return *env.NewString(s1.Value[0:s2.Value])
				default:
					return MakeArgError(ps, 1, []env.Type{env.IntegerType}, "left")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "left")
			}
		},
	},

	"newline": {
		Argsn: 0,
		Doc:   "Returns the newline character.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewString("\n")
		},
	},

	"nl": {
		Argsn: 1,
		Doc:   "Returns the argument 1 a d a newline character.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.String:
				return *env.NewString(s1.Value + "\n")
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "left")
			}
		},
	},

	"pink": {
		Argsn: 1,
		Doc:   "Returns the argument 1 a d a newline character.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.String:
				return *env.NewString("\033[35m" + s1.Value + "\033[0m")
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "left")
			}
		},
	},

	"trim": {
		Argsn: 1,
		Doc:   "Trims the String of spacing characters.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.String:
				return *env.NewString(strings.TrimSpace(s1.Value))
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "trim")
			}
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
						return *env.NewString(strings.ReplaceAll(s1.Value, s2.Value, s3.Value))
					default:
						return MakeArgError(ps, 3, []env.Type{env.StringType}, "replace")
					}
				default:
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "replace")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "replace")
			}
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
						return *env.NewString(s1.Value[s2.Value:s3.Value])
					default:
						return MakeArgError(ps, 3, []env.Type{env.IntegerType}, "substring")
					}
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "substring")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "substring")
			}
		},
	},

	"contains": {
		Argsn: 2,
		Doc:   "Returns true if argument 2 contains argument 1",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			//contains with string
			case env.String:
				switch s2 := arg1.(type) {
				case env.String:
					if strings.Contains(s1.Value, s2.Value) {
						return *env.NewInteger(1)
					} else {
						return *env.NewInteger(0)
					}
				default:
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "contains")
				}
			//contains block
			case env.Block:
				switch value := arg1.(type) {
				case env.Integer:
					if util.ContainsVal(ps, s1.Series.S, value) {
						return *env.NewInteger(1)
					} else {
						return *env.NewInteger(0)
					}
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "contains")
				}
			// contains list
			case env.List:
				switch value := arg1.(type) {
				case env.Integer:
					isListContains := false
					for i := 0; i < len(s1.Data); i++ {
						if s1.Data[i] == value.Value {
							isListContains = true
							break
						}
					}
					if isListContains {
						return *env.NewInteger(1)
					} else {
						return *env.NewInteger(0)
					}
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "contains")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType, env.BlockType, env.ListType}, "contains")
			}
		},
	},

	"has-suffix": {
		Argsn: 2,
		Doc:   "Returns part of the String between two positions.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.String:
				switch s2 := arg1.(type) {
				case env.String:
					if strings.HasSuffix(s1.Value, s2.Value) {
						return *env.NewInteger(1)
					} else {
						return *env.NewInteger(0)
					}
				default:
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "has-suffix")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "has-suffix")
			}
		},
	},

	"has-prefix": {
		Argsn: 2,
		Doc:   "Returns part of the String between two positions.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.String:
				switch s2 := arg1.(type) {
				case env.String:
					if strings.HasPrefix(s1.Value, s2.Value) {
						return *env.NewInteger(1)
					} else {
						return *env.NewInteger(0)
					}
				default:
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "has-prefix")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "has-prefix")
			}
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
					return *env.NewInteger(int64(res))
				default:
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "index?")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "index?")
			}
		},
	},

	"position?": {
		Argsn: 2,
		Doc:   "Returns part of the String between two positions.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.String: // TODO-FIX ... let s1 be any type if s2 is block, and string only if s2 is string .. reverse nesting in switches
				switch s2 := arg1.(type) {
				case env.String:
					res := strings.Index(s2.Value, s1.Value)
					return *env.NewInteger(int64(res + 1))
				case env.Block:
					res := util.IndexOfSlice(ps, s2.Series.S, s1)
					if res == -1 {
						return MakeBuiltinError(ps, "not found", "position?")
					}
					return *env.NewInteger(int64(res + 1))
				default:
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "position?")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "position?")
			}
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
					return *env.NewString(s1.Value[len(s1.Value)-int(s2.Value):])
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "right")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "right")
			}
		},
	},

	"space": {
		Argsn: 1,
		Doc:   "Adds space to the end of argument",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.String:
				return *env.NewString(s1.Value + " ")
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "concat")
			}
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
					return *env.NewString(strconv.Itoa(int(s1.Value)) + s2.Value)
				case env.Integer:
					return *env.NewString(strconv.Itoa(int(s1.Value)) + strconv.Itoa(int(s2.Value)))
				default:
					return MakeArgError(ps, 2, []env.Type{env.StringType, env.IntegerType}, "concat")
				}
			case env.String:
				switch s2 := arg1.(type) {
				case env.String:
					return *env.NewString(s1.Value + s2.Value)
				case env.Integer:
					return *env.NewString(s1.Value + strconv.Itoa(int(s2.Value)))
				default:
					return MakeArgError(ps, 2, []env.Type{env.StringType, env.IntegerType}, "concat")
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
				default:
					return MakeBuiltinError(ps, "If Arg 1 is Block then Arg 2 should be Block or Object type.", "concat")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType, env.StringType, env.BlockType}, "concat")
			}
		},
	},

	"union": { // **
		Argsn: 2,
		Doc:   "Accepts a block or list of values and returns only unique values.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Block:
				switch b2 := arg1.(type) {
				case env.Block:
					union := util.UnionOfBlocks(ps, s1, b2)
					return *env.NewBlock(*env.NewTSeries(union))
				default:
					return MakeArgError(ps, 2, []env.Type{env.BlockType}, "union")
				}
			case env.List:
				switch l2 := arg1.(type) {
				case env.List:
					union := util.UnionOfLists(ps, s1, l2)
					return *env.NewList(union)
				default:
					return MakeArgError(ps, 2, []env.Type{env.ListType}, "union")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType, env.ListType}, "union")
			}
		},
	},

	"intersection": {
		Argsn: 2,
		Doc:   "Finds the intersection of two values.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.String:
				switch s2 := arg1.(type) {
				case env.String:
					inter := util.IntersectStrings(s1.Value, s2.Value)
					return *env.NewString(inter)
				default:
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "intersect")
				}
			case env.Block:
				switch b2 := arg1.(type) {
				case env.Block:
					inter := util.IntersectBlocks(ps, s1, b2)
					return *env.NewBlock(*env.NewTSeries(inter))
				default:
					return MakeArgError(ps, 2, []env.Type{env.BlockType}, "intersect")
				}
			case env.List:
				switch l2 := arg1.(type) {
				case env.List:
					inter := util.IntersectLists(ps, s1, l2)
					return *env.NewList(inter)
				default:
					return MakeArgError(ps, 2, []env.Type{env.ListType}, "intersect")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType, env.BlockType, env.ListType}, "intersect")
			}
		},
	},

	"difference": {
		Argsn: 2,
		Doc:   "Finds the difference (values in first but not in second) of two values.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.String:
				switch s2 := arg1.(type) {
				case env.String:
					diff := util.DiffStrings(s1.Value, s2.Value)
					return *env.NewString(diff)
				default:
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "difference")
				}
			case env.Block:
				switch b2 := arg1.(type) {
				case env.Block:
					diff := util.DiffBlocks(ps, s1, b2)
					return *env.NewBlock(*env.NewTSeries(diff))
				default:
					return MakeArgError(ps, 2, []env.Type{env.BlockType}, "difference")
				}
			case env.List:
				switch l2 := arg1.(type) {
				case env.List:
					diff := util.DiffLists(ps, s1, l2)
					return *env.NewList(diff)
				default:
					return MakeArgError(ps, 2, []env.Type{env.ListType}, "difference")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType, env.BlockType, env.ListType}, "difference")
			}
		},
	},

	"str": {
		Argsn: 1,
		Doc:   "Turn Rye value to String.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Integer:
				return *env.NewString(strconv.Itoa(int(s1.Value)))
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType}, "str")
			}
		},
	},

	"capitalize": { // **
		Argsn: 1,
		Pure:  true,
		Doc:   "Takes a String and returns the same String, but  with first character turned to upper case.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.String:

				english := cases.Title(language.English)
				return *env.NewString(english.String(s1.Value))
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "capitalize")
			}
		},
	},

	"to-lower": { // **
		Argsn: 1,
		Pure:  true,
		Doc:   "Takes a String and returns the same String, but  with all characters turned to lower case.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.String:
				return *env.NewString(strings.ToLower(s1.Value))
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "to-lower")
			}
		},
	},

	"to-upper": { // **
		Argsn: 1,
		Pure:  true,
		Doc:   "Takes a String and returns the same String, but with all characters turned to upper case.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.String:
				return *env.NewString(strings.ToUpper(s1.Value))
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "to-upper")
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
						return *env.NewString(s1.Value + s2.Value + s3.Value)
					default:
						return MakeArgError(ps, 3, []env.Type{env.StringType}, "concat3")
					}
				default:
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "concat3")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "concat3")
			}
		},
	},

	"join": { // **
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
					default:
						return MakeBuiltinError(ps, "List data should me integer or string.", "join")
					}
				}
				return *env.NewString(str.String())
			case env.Block:

				ser := ps.Ser
				ps.Ser = s1.Series
				res := make([]env.Object, 0)
				for ps.Ser.Pos() < ps.Ser.Len() {
					// ps, injnow = EvalExpressionInj(ps, inj, injnow)
					EvalExpression2(ps, false)
					res = append(res, ps.Res)
					if ps.ErrorFlag {
						return ps.Res
					}
					//ps.Ser = ser
					if ps.ReturnFlag {
						return ps.Res
					}
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
				bloc := *env.NewBlock(*env.NewTSeries(res))

				var str strings.Builder
				for _, c := range bloc.Series.S {
					switch it := c.(type) {
					case env.String:
						str.WriteString(it.Value)
					case env.Integer:
						str.WriteString(strconv.Itoa(int(it.Value)))
					default:
						return MakeBuiltinError(ps, "Block series data should be string or integer.", "join")
					}
				}
				return *env.NewString(str.String())
			default:
				return MakeArgError(ps, 1, []env.Type{env.ListType, env.BlockType}, "join")
			}
		},
	},

	"join\\with": { // **
		Argsn: 2,
		Pure:  true,
		Doc:   "Joins Block or list of values together.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.List:
				switch s2 := arg1.(type) {
				case env.String:
					var str strings.Builder
					for i, c := range s1.Data {
						if i > 0 {
							str.WriteString(s2.Value)
						}
						switch it := c.(type) {
						case string:
							str.WriteString(it)
						case env.String:
							str.WriteString(it.Value)
						case int:
							str.WriteString(strconv.Itoa(it))
						case env.Integer:
							str.WriteString(strconv.Itoa(int(it.Value)))
						default:
							return MakeBuiltinError(ps, "Data should be string or integer.", "join\\with")
						}
					}
					return *env.NewString(str.String())
				default:
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "join\\with")
				}
			case env.Block:
				switch s2 := arg1.(type) {
				case env.String:
					var str strings.Builder
					for i, c := range s1.Series.S {
						if i > 0 {
							str.WriteString(s2.Value)
						}
						switch it := c.(type) {
						case env.String:
							str.WriteString(it.Value)
						case env.Integer:
							str.WriteString(strconv.Itoa(int(it.Value)))
						default:
							return MakeBuiltinError(ps, "Block series data should be string or integer.", "join\\with")
						}
					}
					return *env.NewString(str.String())
				default:
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "join\\with")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.ListType, env.BlockType}, "join\\with")
			}
		},
	},

	"split": { // **
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
						spl2[i] = *env.NewString(val)
					}
					return *env.NewBlock(*env.NewTSeries(spl2))
				default:
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "split")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "split")
			}
		},
	},

	"split\\quoted": { // **
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
						return MakeArgError(ps, 3, []env.Type{env.StringType}, "split-quoted")
					}
				default:
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "split-quoted")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "split-quoted")
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
						spl2[i] = *env.NewString(val)
					}
					return *env.NewBlock(*env.NewTSeries(spl2))
				default:
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "split\\many")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "split\\many")
			}
		},
	},

	"split\\every": { // **
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
						spl2[i] = *env.NewString(val)
					}
					return *env.NewBlock(*env.NewTSeries(spl2))
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "split-every")
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
					return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "split-every")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType, env.BlockType}, "split-every")
			}
		},
	},

	// BASIC SERIES FUNCTIONS

	"first": { // **
		Argsn: 1,
		Doc:   "Accepts a Block, List or String and returns the first item.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Block:
				if len(s1.Series.S) == 0 {
					return MakeBuiltinError(ps, "Block is empty.", "first")
				}
				return s1.Series.Get(int(0))
			case env.List:
				if len(s1.Data) == 0 {
					return MakeBuiltinError(ps, "List is empty.", "first")
				}
				return env.ToRyeValue(s1.Data[int(0)])
			case env.String:
				str := []rune(s1.Value)
				if len(str) == 0 {
					return MakeBuiltinError(ps, "String is empty.", "first")
				}
				return *env.NewString(string(str[0]))
			case env.Spreadsheet:
				return s1.GetRow(ps, int(0))
			default:
				return MakeArgError(ps, 1, []env.Type{env.SpreadsheetType, env.BlockType, env.StringType, env.ListType}, "first")
			}
		},
	},

	"rest": { // **
		Argsn: 1,
		Doc:   "Accepts a Block, List or String and returns all but first item.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Block:
				if len(s1.Series.S) == 0 {
					return MakeBuiltinError(ps, "Block is empty.", "rest")
				}
				return *env.NewBlock(*env.NewTSeries(s1.Series.S[1:]))
			case env.List:
				if len(s1.Data) == 0 {
					return MakeBuiltinError(ps, "Block is empty.", "rest")
				}
				return env.NewList(s1.Data[int(1):])
			case env.String:
				str := []rune(s1.Value)
				if len(str) < 1 {
					return MakeBuiltinError(ps, "String has only one element.", "rest")
				}
				return *env.NewString(string(str[1:]))
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType, env.StringType, env.ListType}, "rest")
			}
		},
	},

	"rest\\from": { // **
		Argsn: 2,
		Doc:   "Accepts a Block, List or String and an Integer N, returns all but first N items.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch num := arg1.(type) {
			case env.Integer:
				switch s1 := arg0.(type) {
				case env.Block:
					if len(s1.Series.S) == 0 {
						return MakeBuiltinError(ps, "Block is empty.", "rest\\from")
					}
					if len(s1.Series.S) <= int(num.Value) {
						return MakeBuiltinError(ps, fmt.Sprintf("Block has less than %d elements.", num.Value+1), "rest\\from")
					}
					return *env.NewBlock(*env.NewTSeries(s1.Series.S[int(num.Value):]))
				case env.List:
					if len(s1.Data) == 0 {
						return MakeBuiltinError(ps, "List is empty.", "rest\\from")
					}
					if len(s1.Data) <= int(num.Value) {
						return MakeBuiltinError(ps, fmt.Sprintf("List has less than %d elements.", num.Value+1), "rest\\from")
					}
					return env.NewList(s1.Data[int(num.Value):])
				case env.String:
					str := []rune(s1.Value)
					if len(str) == 0 {
						return MakeBuiltinError(ps, "String is empty.", "rest\\from")
					}
					if len(str) <= int(num.Value) {
						return MakeBuiltinError(ps, fmt.Sprintf("String has less than %d elements.", num.Value+1), "rest\\from")
					}
					return *env.NewString(string(str[int(num.Value):]))
				default:
					return MakeArgError(ps, 1, []env.Type{env.BlockType, env.ListType, env.StringType}, "rest\\from")
				}
			default:
				return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "rest\\from")
			}
		},
	},

	"tail": { // **
		Argsn: 2,
		Doc:   "Accepts a Block, List or String and Integer N, returns the last N items.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch num := arg1.(type) {
			case env.Integer:
				numVal := int(num.Value)
				switch s1 := arg0.(type) {
				case env.Block:
					if len(s1.Series.S) == 0 {
						return *env.NewBlock(*env.NewTSeries([]env.Object{}))
					}
					if len(s1.Series.S) < numVal {
						numVal = len(s1.Series.S)
					}
					return *env.NewBlock(*env.NewTSeries(s1.Series.S[len(s1.Series.S)-numVal:]))
				case env.List:
					if len(s1.Data) == 0 {
						return *env.NewList([]any{})
					}
					if len(s1.Data) < numVal {
						numVal = len(s1.Data)
					}
					return *env.NewList(s1.Data[len(s1.Data)-numVal:])
				case env.String:
					str := []rune(s1.Value)
					if len(str) == 0 {
						return *env.NewString("")
					}
					if len(str) < numVal {
						numVal = len(str)
					}
					return *env.NewString(string(str[len(str)-numVal:]))
				default:
					return MakeArgError(ps, 1, []env.Type{env.BlockType, env.ListType, env.StringType}, "tail")
				}
			default:
				return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "tail")
			}
		},
	},

	"second": { // **
		Argsn: 1,
		Doc:   "Accepts a Block, List or String and returns the second value in it.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Block:
				if len(s1.Series.S) < 2 {
					return MakeBuiltinError(ps, "Block has no second element.", "second")
				}
				return s1.Series.Get(1)
			case env.List:
				if len(s1.Data) < 2 {
					return MakeBuiltinError(ps, "List has no second element.", "second")
				}
				return env.ToRyeValue(s1.Data[1])
			case env.String:
				str := []rune(s1.Value)
				if len(str) < 2 {
					return MakeBuiltinError(ps, "String has no second element.", "second")
				}
				return *env.NewString(string(str[1]))
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType, env.ListType, env.StringType}, "second")
			}
		},
	},

	"third": { // **
		Argsn: 1,
		Doc:   "Accepts a Block, List or String and returns the third value in it.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Block:
				if len(s1.Series.S) < 3 {
					return MakeBuiltinError(ps, "Block has no third element.", "third")
				}
				return s1.Series.Get(int(2))
			case env.List:
				if len(s1.Data) < 3 {
					return MakeBuiltinError(ps, "List has no third element.", "third")
				}
				return env.ToRyeValue(s1.Data[2])
			case env.String:
				str := []rune(s1.Value)
				if len(str) < 3 {
					return MakeBuiltinError(ps, "String has no third element.", "third")
				}
				return *env.NewString(string(str[2]))
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType, env.ListType, env.StringType}, "third")
			}
		},
	},

	"last": { // **
		Argsn: 1,
		Doc:   "Accepts a Block, List or String and returns the last value in it.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Block:
				if len(s1.Series.S) == 0 {
					return MakeBuiltinError(ps, "Block is empty.", "last")
				}
				return s1.Series.Get(s1.Series.Len() - 1)
			case env.List:
				if len(s1.Data) == 0 {
					return MakeBuiltinError(ps, "List is empty.", "last")
				}
				return env.ToRyeValue(s1.Data[len(s1.Data)-1])
			case env.String:
				if len(s1.Value) == 0 {
					return MakeBuiltinError(ps, "String is empty.", "last")
				}
				return *env.NewString(s1.Value[len(s1.Value)-1:])
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType, env.ListType, env.StringType}, "last")
			}
		},
	},

	"head": { // **
		Argsn: 2,
		Doc:   "Accepts a Block, List or String and an Integer N, returns the first N values.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch num := arg1.(type) {
			case env.Integer:
				numVal := int(num.Value)
				switch s1 := arg0.(type) {
				case env.Block:
					if len(s1.Series.S) == 0 {
						return *env.NewBlock(*env.NewTSeries([]env.Object{}))
					}
					if len(s1.Series.S) < numVal {
						numVal = len(s1.Series.S)
					}
					if numVal < 0 {
						numVal = len(s1.Series.S) + numVal // warn: numVal is negative so we must add
					}
					return *env.NewBlock(*env.NewTSeries(s1.Series.S[0:numVal]))
				case env.List:
					if len(s1.Data) == 0 {
						return *env.NewList([]any{})
					}
					if len(s1.Data) < numVal {
						numVal = len(s1.Data)
					}
					return *env.NewList(s1.Data[0:numVal])
				case env.String:
					str := []rune(s1.Value)
					if len(str) == 0 {
						return *env.NewString("")
					}
					if len(str) < numVal {
						numVal = len(str)
					}
					return *env.NewString(string(str[0:numVal]))
				default:
					return MakeArgError(ps, 1, []env.Type{env.BlockType, env.ListType, env.StringType}, "head")
				}
			default:
				return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "head")
			}
		},
	},

	"nth": { // **
		Argsn: 2,
		Doc:   "Accepts a Block, List or String and Integer N, returns the N-th value.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch num := arg1.(type) {
			case env.Integer:
				switch s1 := arg0.(type) {
				case env.Block:
					if num.Value > int64(s1.Series.Len()) {
						return MakeBuiltinError(ps, fmt.Sprintf("Block has less than %d elements.", num.Value), "nth")
					}
					return s1.Series.Get(int(num.Value - 1))
				case env.List:
					if num.Value > int64(len(s1.Data)) {
						return MakeBuiltinError(ps, fmt.Sprintf("List has less than %d elements.", num.Value), "nth")
					}
					return env.ToRyeValue(s1.Data[int(num.Value-1)])
				case env.String:
					str := []rune(s1.Value)
					if num.Value > int64(len(str)) {
						return MakeBuiltinError(ps, fmt.Sprintf("String has less than %d elements.", num.Value), "nth")
					}
					return *env.NewString(string(str[num.Value-1 : num.Value]))
				default:
					return MakeArgError(ps, 1, []env.Type{env.BlockType}, "nth")
				}
			default:
				return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "nth")
			}
		},
	},

	"values": { // **
		Argsn: 1,
		Doc:   "Accepts a Dict and returns a List of just values.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch dict := arg0.(type) {
			case env.Dict:
				newl := make([]any, 0)
				for _, v := range dict.Data {
					newl = append(newl, v)
				}
				return *env.NewList(newl)
			default:
				return MakeArgError(ps, 1, []env.Type{env.DictType}, "values")
			}
		},
	},

	// These are rebol like functions for blocks with carret ... I'm not sure yet it they will be included in long term
	// a carret is an imperative concept, doint blocks on the otheh hand requires a carret

	"peek": {
		Argsn: 1,
		Doc:   "Accepts Block and returns the current value, without removing it.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Block:
				return s1.Series.Peek()
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType}, "peek")
			}
		},
	},
	"pop": {
		Argsn: 1,
		Doc:   "Accepts Block and returns the next value and removes it from the Block.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Block:
				return s1.Series.Pop()
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType}, "pop")
			}
		},
	},
	"pos": {
		Argsn: 1,
		Doc:   "Accepts Block and returns the position of it's carret.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Block:
				return *env.NewInteger(int64(s1.Series.Pos()))
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType}, "pos")
			}
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
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType}, "next")
			}
		},
	},
	//
	"remove-last!": { // **
		Argsn: 1,
		Pure:  false,
		Doc:   "Accepts Block and returns the next value and removes it from the Block.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch wrd := arg0.(type) {
			case env.Word:
				val, found, ctx := ps.Ctx.Get2(wrd.Index)
				if found {
					switch oldval := val.(type) {
					case env.Block:
						s := &oldval.Series
						oldval.Series = *s.RmLast()
						ctx.Set(wrd.Index, oldval)
						return oldval
					default:
						return MakeBuiltinError(ps, "Old value should be Block type.", "remove-last!")
					}
				} else {
					return MakeBuiltinError(ps, "Word not found in context.", "remove-last!")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.WordType}, "remove-last!")
			}
		},
	},
	"append!": { // **
		Argsn: 2,
		Doc:   "Accepts Rye value and Tagword with a Block or String. Appends Rye value to Block/String in place, also returns it	.",
		Pure:  false,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch wrd := arg1.(type) {
			case env.Word:
				val, found, ctx := ps.Ctx.Get2(wrd.Index)
				if found {
					switch oldval := val.(type) {
					case env.String:
						var newval env.String
						switch s3 := arg0.(type) {
						case env.String:
							newval = *env.NewString(oldval.Value + s3.Value)
						case env.Integer:
							newval = *env.NewString(oldval.Value + strconv.Itoa(int(s3.Value)))
						}
						ctx.Set(wrd.Index, newval)
						return newval
					case env.Block: // TODO
						// 	fmt.Println(123)
						s := &oldval.Series
						oldval.Series = *s.Append(arg0)
						ctx.Set(wrd.Index, oldval)
						return oldval
					case env.List:
						dataSlice := make([]any, 0)
						switch listData := arg0.(type) {
						case env.List:
							for _, v1 := range oldval.Data {
								dataSlice = append(dataSlice, env.ToRyeValue(v1))
							}
							for _, v2 := range listData.Data {
								dataSlice = append(dataSlice, env.ToRyeValue(v2))
							}
						default:
							return makeError(ps, "Need to pass List of data")
						}
						combineList := make([]any, 0, len(dataSlice))
						for _, v := range dataSlice {
							combineList = append(combineList, env.ToRyeValue(v))
						}
						finalList := *env.NewList(combineList)
						ctx.Set(wrd.Index, finalList)
						return finalList
					default:
						return makeError(ps, "Type of tagword is not String or Block")
					}
				}
				return makeError(ps, "Tagword not found.")
			case env.Block:
				dataSlice := make([]env.Object, 0)
				switch blockData := arg0.(type) {
				case env.Block:
					for _, v1 := range wrd.Series.S {
						dataSlice = append(dataSlice, env.ToRyeValue(v1))
					}
					for _, v2 := range blockData.Series.S {
						dataSlice = append(dataSlice, env.ToRyeValue(v2))
					}
				default:
					return makeError(ps, "Need to pass block of data")
				}
				return *env.NewBlock(*env.NewTSeries(dataSlice))
			case env.String:
				finalStr := ""
				switch str := arg0.(type) {
				case env.String:
					finalStr = wrd.Value + str.Value
				case env.Integer:
					finalStr = wrd.Value + strconv.Itoa(int(str.Value))
				}
				return *env.NewString(finalStr)
			default:
				return makeError(ps, "Value not tagword")
			}
		},
	},

	"change\\nth!": { // **
		Argsn: 3,
		Doc:   "Accepts a Block or List, Integer n and a value. Changes the n-th value in the Block in place. Also returns the new series.",
		Pure:  false,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch num := arg1.(type) {
			case env.Integer:
				switch s1 := arg0.(type) {
				case env.Block:
					if num.Value > int64(s1.Series.Len()) {
						return MakeBuiltinError(ps, fmt.Sprintf("Block has less than %d elements.", num.Value), "change\\nth!")
					}
					s1.Series.S[num.Value-1] = arg2
					return s1
				case env.List:
					if num.Value > int64(len(s1.Data)) {
						return MakeBuiltinError(ps, fmt.Sprintf("List has less than %d elements.", num.Value), "change\\nth!")
					}
					s1.Data[num.Value-1] = env.RyeToRaw(arg2)
					return s1
				default:
					return MakeArgError(ps, 1, []env.Type{env.BlockType}, "change\\nth!")
				}
			default:
				return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "change\\nth!")
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
			case env.Word:
				switch s2 := arg1.(type) {
				case env.Word:
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

	/* "table": {
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

				case env.Word:
					// TODO
				}
				return nil
			}
			return nil
		},
	}, */

	"add-row": {
		Argsn: 2,
		Doc:   "Constructs an empty table, accepts a block of column names",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch table := arg0.(type) {
			case env.Spreadsheet:
				switch bloc := arg1.(type) {
				case env.Block:
					vals := make([]any, bloc.Series.Len())
					for i := 0; i < bloc.Series.Len(); i++ {
						vals[i] = bloc.Series.Get(i)
					}
					table.AddRow(*env.NewSpreadsheetRow(vals, &table))
					return table
				}
				return nil
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
	"return": { // **
		Argsn: 1,
		Doc:   "Accepts one value and returns it.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			ps.ReturnFlag = true
			ps.Res = arg0
			return arg0
		},
	},

	"or-return": { // **
		Argsn: 1,
		Doc:   "Accepts one value and returns from evaluation if true.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			//fmt.Println("RETURN")
			if util.IsTruthy(arg0) {
				// ps.ReturnFlag = true
				ps.SkipFlag = true
			}
			ps.Res = arg0
			return arg0
		},
	},

	// return , error , failure functions
	"exit": { // **
		Argsn: 1,
		Doc:   "Accepts one value and returns it.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch code := arg0.(type) {
			case env.Integer:
				os.Exit(int(code.Value))
				return nil
			default:
				fmt.Println(code.Inspect(*ps.Idx))
				os.Exit(0)
				return nil
			}
		},
	},

	"^fail": {
		Argsn: 1,
		Doc:   "Returning Fail.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			//fmt.Println("FAIL")
			ps.FailureFlag = true
			ps.ReturnFlag = true
			return MakeRyeError(ps, arg0, nil)
		},
	},

	"fail": { // **
		Argsn: 1,
		Doc:   "Constructs and Fails with an Error object. Accepts String as message, Integer as code, or block for multiple parameters.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			//fmt.Println("FAIL")
			ps.FailureFlag = true
			return MakeRyeError(ps, arg0, nil)
		},
	},

	"failure": { // **
		Argsn: 1,
		Doc:   "Constructs and Error object. Accepts String as message, Integer as code, or block for multiple parameters.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			//ps.ErrorFlag = true
			return MakeRyeError(ps, arg0, nil)
		},
	},

	"wrap\\failure": {
		Argsn: 2,
		Doc:   "Wraps an Error with another Error. Accepts String as message, Integer as code, or block for multiple parameters and Error as arguments.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch er := arg1.(type) {
			case *env.Error:
				return MakeRyeError(ps, arg0, er)
			default:
				return MakeArgError(ps, 2, []env.Type{env.ErrorType}, "wrap\\error")
			}
		},
	},

	"code?": { // **
		AcceptFailure: true,
		Argsn:         1,
		Doc:           "Returns the status code of the Error.", // TODO -- seems duplicate of status
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch er := arg0.(type) {
			case env.Error:
				return *env.NewInteger(int64(er.Status))
			case *env.Error:
				return *env.NewInteger(int64(er.Status))
			default:
				ps.FailureFlag = true
				return env.NewError("arg 0 not error")
			}
		},
	},

	"disarm": { // **
		AcceptFailure: true,
		Argsn:         1,
		Doc:           "Disarms the Error.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			ps.FailureFlag = false
			return arg0
		},
	},

	"failed?": { // **
		AcceptFailure: true,
		Argsn:         1,
		Doc:           "Checks if first argument is an Error. Returns a boolean.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			ps.FailureFlag = false
			switch arg0.(type) {
			case env.Error:
				return *env.NewInteger(int64(1))
			case *env.Error:
				return *env.NewInteger(int64(1))
			}
			return *env.NewInteger(int64(0))
		},
	},

	"check": { // **
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
			}
			return arg0
		},
	},

	"require": { // **
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
			}
			return arg0
		},
	},

	"assert-equal": { // **
		Argsn: 2,
		Doc:   "Test if two values are equal. Fail if not.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			if arg0.GetKind() == arg1.GetKind() && arg0.Inspect(*ps.Idx) == arg1.Inspect(*ps.Idx) {
				return *env.NewInteger(1)
			} else {
				return makeError(ps, "Values are not equal: "+arg0.Inspect(*ps.Idx)+" "+arg1.Inspect(*ps.Idx))
			}
		},
	},

	"fix": { // **
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

	"fix\\else": {
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
				return *env.NewString(arg0.Print(*ps.Idx))
			}
			return *env.NewString(r.String())
		},
	},

	// date time functions

	"date": {
		Argsn: 1,
		Doc:   "Accepts a String and returns a Date object.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) (res env.Object) {
			switch s1 := arg0.(type) {
			case env.String:
				t, err := time.Parse("2006-01-02", s1.Value)
				if err != nil {
					return MakeBuiltinError(ps, err.Error(), "date")
				}
				return *env.NewDate(t)
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "date")
			}
		},
	},

	"datetime": {
		Argsn: 1,
		Doc:   "Accepts a String and returns a Date object.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) (res env.Object) {
			switch s1 := arg0.(type) {
			case env.String:
				t, err := time.Parse("2006-01-02T15:04:05", s1.Value)
				if err != nil {
					return MakeBuiltinError(ps, err.Error(), "datetime")
				}
				return *env.NewTime(t)
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "datetime")
			}
		},
	},

	"now": {
		Argsn: 0,
		Doc:   "Returns current Time.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewTime(time.Now())
		},
	},

	// end of date time functions

	"range": { // **
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
						objs[idx] = *env.NewInteger(i)
						idx += 1
					}
					return *env.NewBlock(*env.NewTSeries(objs))
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "range")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType}, "range")
			}
		},
	},

	"length?": { // **
		Argsn: 1,
		Doc:   "Accepts a collection (String, Block, Dict, Spreadsheet) and returns it's length.", // TODO -- accept list, context also
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.String:
				return *env.NewInteger(int64(len(s1.Value)))
			case env.Dict:
				return *env.NewInteger(int64(len(s1.Data)))
			case env.List:
				return *env.NewInteger(int64(len(s1.Data)))
			case env.Block:
				return *env.NewInteger(int64(s1.Series.Len()))
			case env.Spreadsheet:
				return *env.NewInteger(int64(len(s1.Rows)))
			case env.RyeCtx:
				return *env.NewInteger(int64(s1.GetWords(*ps.Idx).Series.Len()))
			case env.Vector:
				return *env.NewInteger(int64(s1.Value.Len()))
			default:
				fmt.Println(s1)
				return MakeArgError(ps, 2, []env.Type{env.StringType, env.DictType, env.ListType, env.BlockType, env.SpreadsheetType, env.VectorType}, "range")
			}
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
				return *env.NewInteger(int64(len(s1.Cols)))
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
				for k := range s1.Data {
					keys[i] = *env.NewString(k)
					i++
				}
				return *env.NewBlock(*env.NewTSeries(keys))
			case env.Spreadsheet:
				keys := make([]env.Object, len(s1.Cols))
				for i, k := range s1.Cols {
					keys[i] = *env.NewString(k)
				}
				return *env.NewBlock(*env.NewTSeries(keys))
			default:
				fmt.Println("Error")
			}
			return nil
		},
	},

	"col-sum": {
		Argsn: 2,
		Doc:   "Accepts a spreadsheet and a column name and returns a sum of a column.", // TODO -- let it accept a block and list also
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			var name string
			switch s1 := arg0.(type) {
			case env.Spreadsheet:
				switch s2 := arg1.(type) {
				case env.Word:
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
		},
	},

	"col-avg": {
		Argsn: 2,
		Doc:   "Accepts a spreadsheet and a column name and returns a sum of a column.", // TODO -- let it accept a block and list also
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			var name string
			switch s1 := arg0.(type) {
			case env.Spreadsheet:
				switch s2 := arg1.(type) {
				case env.Word:
					name = ps.Idx.GetWord(s2.Index)
				case env.String:
					name = s2.Value
				default:
					ps.ErrorFlag = true
					return env.NewError("second arg not string")
				}
				r, err := s1.Sum_Just(name)
				if err != nil {
					ps.ErrorFlag = true
					return env.NewError(err.Error())
				}
				n := s1.NRows()
				return *env.NewDecimal(r / float64(n))

			default:
				ps.ErrorFlag = true
				return env.NewError("first arg not spreadsheet")
			}
		},
	},

	"A1": {
		Argsn: 1,
		Doc:   "Accepts a Spreadsheet and returns the first row first column cell.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s0 := arg0.(type) {
			case env.Spreadsheet:
				r := s0.Rows[0].Values[0]
				return env.ToRyeValue(r)

			default:
				ps.ErrorFlag = true
				return env.NewError("first arg not spreadsheet")
			}
		},
	},
	"B1": {
		Argsn: 1,
		Doc:   "Accepts a Spreadsheet and returns the first row second column cell.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s0 := arg0.(type) {
			case env.Spreadsheet:
				r := s0.Rows[0].Values[1]
				return env.ToRyeValue(r)

			default:
				ps.ErrorFlag = true
				return env.NewError("first arg not spreadsheet")
			}
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

				r := exec.Command("/bin/bash", "-c", s0.Value) //nolint: gosec
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
								return env.ToRyeValue(" "-----------" + string(stdout)) */
				//				return env.ToRyeValue(string(stdout))
			default:
				return makeError(ps, "Arg 1 should be String")
			}
			return nil
		},
	},

	/* TODO: * whitelist only to http and https prefix
	         * os\open is temp name, figure out where it belongs, maybe os module and subcontext
			 * figure out which other functions would belong in os module and if it makes sense
	 "os\\open": { // todo -- variation is not meant for grouping inside context ... just for function variations ... just temp to test
		Argsn: 1,
		Doc:   "",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s0 := arg0.(type) {
			case env.Uri:

				var err error
				url := s0.GetFullUri(*ps.Idx)

				switch runtime.GOOS {
				case "linux":
					err = exec.Command("xdg-open", url).Start()
				case "windows":
					err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
				case "darwin":
					err = exec.Command("open", url).Start()
				default:
					err = fmt.Errorf("unsupported platform")
				}
				if err != nil {
					log.Fatal(err)
				}
			default:
				return makeError(ps, "Arg 1 should be String") // TODO - make proper error
			}
			return nil
		},
	}, */

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
							//							return *env.NewInteger(1)
						} else {
							str.WriteString("\nBinding *" + name + "* is missing.")
							missing = append(missing, node)
							// v0 todo: Print mis
							// v1 todo: Print the instructions of what modules to go get in the project folder and reinstall
							// v2 todo: ge get modules and then recompile rye with these flags into current folder
							//							return *env.NewInteger(0)
						}
					}
				}
				if len(missing) > 0 {
					return makeError(ps, str.String())
				} else {
					return *env.NewInteger(1)
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
					blts = append(blts, *env.NewWord(idx))
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
					blts = append(blts, *env.NewWord(idx))
				}
			}
			return *env.NewBlock(*env.NewTSeries(blts))
		},
	},
	"Rye-itself//args": {
		Argsn: 1,
		Doc:   "",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return util.StringToFieldsWithQuoted(strings.Join(os.Args[2:], " "), " ", "\"")
			// block, _ := loader.LoadString(os.Args[0], false)
			// return block
		},
	},
	"Rye-itself//args\\raw": {
		Argsn: 1,
		Doc:   "",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewString(strings.Join(os.Args[2:], " "))
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
	RegisterBuiltins2(Builtins_validation, ps, "validation")
	RegisterBuiltins2(Builtins_json, ps, "json")
	RegisterBuiltins2(Builtins_stackless, ps, "stackless")
	RegisterBuiltins2(Builtins_eyr, ps, "eyr")
	RegisterBuiltins2(Builtins_conversion, ps, "conversion")
	RegisterBuiltins2(Builtins_http, ps, "http")
	RegisterBuiltins2(Builtins_crypto, ps, "crypto")
	RegisterBuiltins2(Builtins_goroutines, ps, "goroutines")
	RegisterBuiltins2(Builtins_psql, ps, "psql")
	RegisterBuiltins2(Builtins_mysql, ps, "mysql")
	RegisterBuiltins2(Builtins_bcrypt, ps, "bcrypt")
	RegisterBuiltins2(Builtins_email, ps, "email")
	RegisterBuiltins2(Builtins_structures, ps, "structs")
	RegisterBuiltins2(Builtins_telegrambot, ps, "telegram")
	RegisterBuiltins2(Builtins_spreadsheet, ps, "spreadsheet")
	RegisterBuiltins2(Builtins_vector, ps, "vector")
	RegisterBuiltins2(Builtins_bson, ps, "bson")
	RegisterBuiltins2(Builtins_smtpd, ps, "smtpd")
	RegisterBuiltins2(Builtins_mail, ps, "mail")
	RegisterBuiltins2(Builtins_ssh, ps, "ssh")
	RegisterBuiltinsInContext(Builtins_math, ps, "math")
	RegisterBuiltinsInContext(Builtins_devops, ps, "devops")
	// ## Archived modules
	// RegisterBuiltins2(Builtins_gtk, ps, "gtk")
	// RegisterBuiltins2(Builtins_nats, ps, "nats")
	// RegisterBuiltins2(Builtins_qframe, ps, "qframe")
	// RegisterBuiltins2(Builtins_nng, ps, "nng")
	// RegisterBuiltins2(Builtins_raylib, ps, "raylib")
	// RegisterBuiltins2(Builtins_cayley, ps, "cayley")
}

var BuiltinNames map[string]int // TODO --- this looks like some hanging global ... it should move to ProgramState, it doesn't even really work with contrib and external probably

func RegisterBuiltins2(builtins map[string]*env.Builtin, ps *env.ProgramState, name string) {
	BuiltinNames[name] = len(builtins)
	for k, v := range builtins {
		bu := env.NewBuiltin(v.Fn, v.Argsn, v.AcceptFailure, v.Pure, v.Doc)
		registerBuiltin(ps, k, *bu)
	}
}

func RegisterBuiltinsInContext(builtins map[string]*env.Builtin, ps *env.ProgramState, name string) {
	BuiltinNames[name] = len(builtins)

	ctx := ps.Ctx
	ps.Ctx = env.NewEnv(ps.Ctx) // make new context with no parent

	for k, v := range builtins {
		bu := env.NewBuiltin(v.Fn, v.Argsn, v.AcceptFailure, v.Pure, v.Doc)
		registerBuiltin(ps, k, *bu)
	}
	newctx := ps.Ctx
	ps.Ctx = ctx

	wordIdx := ps.Idx.IndexWord(name)
	ps.Ctx.Set(wordIdx, *newctx)
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
