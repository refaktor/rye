package evaldo

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"

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

func NameOfRyeType(t env.Type) string {
	if t < 0 || int(t) >= len(env.NativeTypes) {
		return "INVALID TYPE (" + strconv.FormatInt(int64(t), 10) + ")"
	}
	return env.NativeTypes[t]
}

func MakeArgErrorMessage(N int, allowedTypes []env.Type, fn string) string {
	types := ""
	for i, tt := range allowedTypes {
		if i > 0 {
			types += ", "
		}
		types += env.NativeTypes[tt-1]
	}
	return "builtin `" + fn + "` requires argument " + strconv.Itoa(N) + " to be: " + types + "."
}

func MakeArgError(env1 *env.ProgramState, N int, typ []env.Type, fn string) *env.Error {
	env1.FailureFlag = true
	return env.NewError(MakeArgErrorMessage(N, typ, fn))
}

func MakeNeedsThawedArgError(env1 *env.ProgramState, fn string) *env.Error {
	env1.FailureFlag = true
	return env.NewError("builtin `" + fn + "` requires a thawed table as the first argument")
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
		// TODONOW implement the dialect
		var code env.Object
		if len(val.Series.S) > 0 {
			code = val.Series.Get(0)
			if code.Type() == env.IntegerType {
				return env.NewError4(int(code.(env.Integer).Value), "", er, nil)
			}
		} else {
			return makeError(env1, "Empty error constructor block")
		}
		if len(val.Series.S) > 1 {
			message := val.Series.Get(1)
			if code.Type() == env.IntegerType && message.Type() == env.StringType {
				return env.NewError4(int(code.(env.Integer).Value), message.(env.String).Value, er, nil)
			}
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
			case env.Decimal:
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
			case env.Native:
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
			if idx < 0 {
				return makeError(ps, "Index too low")
			}
			if len(s1.Data) > int(idx) && idx >= 0 {
				v := s1.Data[idx]
				return env.ToRyeValue(v)
			} else {
				return makeError(ps, "Index larger than length")
			}
		}
	case *env.List:
		switch s2 := key.(type) {
		case env.Integer:
			idx := s2.Value
			if posMode {
				idx--
			}
			if idx < 0 {
				return makeError(ps, "Index too low")
			}
			if len(s1.Data) > int(idx) && idx >= 0 {
				v := s1.Data[idx]
				return env.ToRyeValue(v)
			} else {
				return makeError(ps, "Index larger than length")
			}
		}
	case env.Block:
		switch s2 := key.(type) {
		case env.Integer:
			idx := s2.Value
			if posMode {
				idx--
			}
			if idx < 0 {
				return makeError(ps, "Index too low")
			}
			if len(s1.Series.S) >= int(idx)+1 {
				v := s1.Series.Get(int(idx))
				return v
			} else {
				return makeError(ps, "Index larger than length")
			}
		}
	case env.Table:
		switch s2 := key.(type) {
		case env.Integer:
			idx := s2.Value
			if posMode {
				idx--
			}
			if idx < 0 {
				return makeError(ps, "Index too low")
			}
			v := s1.Rows[idx]
			ok := true
			if ok {
				return v
			} else {
				return makeError(ps, "Index larger than length")
			}
		}
	case env.TableRow:
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
				return makeError(ps, "Index larger than length")
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
				return makeError(ps, "Index larger than length")
			}
		}
	}
	// fmt.Printf("GETFROM: %#v %#v %#v\n", data, key, posMode)
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

// Sort list interface
type RyeStringSort []rune

func (s RyeStringSort) Len() int {
	return len(s)
}
func (s RyeStringSort) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s RyeStringSort) Less(i, j int) bool {
	return s[i] < s[j]
}

// Custom Sort object interface
type RyeBlockCustomSort struct {
	data []env.Object
	fn   env.Function
	ps   *env.ProgramState
}

func (s RyeBlockCustomSort) Len() int {
	return len(s.data)
}
func (s RyeBlockCustomSort) Swap(i, j int) {
	s.data[i], s.data[j] = s.data[j], s.data[i]
}
func (s RyeBlockCustomSort) Less(i, j int) bool {
	CallFunctionArgs2(s.fn, s.ps, s.data[i], s.data[j], nil)
	return util.IsTruthy(s.ps.Res)
}

// Custom Sort object interface
type RyeListCustomSort struct {
	data []any
	fn   env.Function
	ps   *env.ProgramState
}

func (s RyeListCustomSort) Len() int {
	return len(s.data)
}
func (s RyeListCustomSort) Swap(i, j int) {
	s.data[i], s.data[j] = s.data[j], s.data[i]
}
func (s RyeListCustomSort) Less(i, j int) bool {
	CallFunctionArgs2(s.fn, s.ps, env.ToRyeValue(s.data[i]), env.ToRyeValue(s.data[j]), nil)
	return util.IsTruthy(s.ps.Res)
}

func IntersectStringsCustom(a env.String, b env.String, ps *env.ProgramState, fn env.Function) string {
	set := make(map[rune]bool)
	var bu strings.Builder
	for _, ch := range a.Value {
		CallFunctionArgs2(fn, ps, b, *env.NewString(string(ch)), nil)
		res := util.IsTruthy(ps.Res)
		if res && !set[ch] {
			bu.WriteRune(ch)
			set[ch] = true
		}
	}
	return bu.String()
}

func IntersectBlocksCustom(a env.Block, b env.Block, ps *env.ProgramState, fn env.Function) []env.Object {
	set := make(map[string]bool)
	res := make([]env.Object, 0)
	for _, v := range a.Series.S {
		CallFunctionArgs2(fn, ps, b, v, nil)
		r := util.IsTruthy(ps.Res)
		strv := v.Inspect(*ps.Idx)
		if r && !set[strv] {
			res = append(res, v)
			set[strv] = true
		}
	}
	return res
}

func LoadScriptLocalFile(ps *env.ProgramState, s1 env.Uri) (env.Object, string) {
	var str string
	fileIdx, _ := ps.Idx.GetIndex("file")
	fullpath := filepath.Join(filepath.Dir(ps.ScriptPath), s1.GetPath())
	if s1.Scheme.Index == fileIdx {
		b, err := os.ReadFile(fullpath)
		if err != nil {
			return MakeBuiltinError(ps, err.Error(), "import"), ps.ScriptPath
		}
		str = string(b) // convert content to a 'string'
	}
	script_ := ps.ScriptPath
	ps.ScriptPath = fullpath
	block_ := loader.LoadStringNEW(str, false, ps)
	return block_, script_
}

func EvaluateLoadedValue(ps *env.ProgramState, block_ env.Object, script_ string, allowMod bool) env.Object {
	switch block := block_.(type) {
	case env.Block:
		ser := ps.Ser
		ps.Ser = block.Series
		ps.AllowMod = allowMod
		EvalBlock(ps)
		ps.AllowMod = false
		ps.Ser = ser
		return ps.Res
	case env.Error:
		ps.ScriptPath = script_
		ps.ErrorFlag = true
		return MakeBuiltinError(ps, block.Message, "import")
	default:
		fmt.Println(block)
		panic("Not block and not error in import builtin.") // TODO -- Think how best to handle this
		// return env.Void{}
	}
}

var ShowResults bool

var builtins = map[string]*env.Builtin{

	// There is require with arity 2 below which makes more sense
	// error { 1 = 0 |require |type? }
	/* "require": {
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
	}, */

	// COMBINATORS

	// Tests:
	// equal  { 101 .pass { 202 } } 101
	// equal  { 101 .pass { 202 + 303 } } 101
	// Args:
	// * value: Any value that will be passed to the block and returned
	// * block: Block of code to execute with the value injected
	// Returns:
	// * The original value, regardless of what the block returns
	"pass": { // **
		Argsn: 2,
		Doc:   "Accepts a value and a block. It does the block, with value injected, and returns (passes on) the initial value.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch bloc := arg1.(type) {
			case env.Block:
				res := arg0
				ser := ps.Ser
				ps.Ser = bloc.Series
				EvalBlockInjMultiDialect(ps, arg0, true)
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

	// Tests:
	// stdout { wrap { prn "*" } { prn "x" } } "*x*"
	// Args:
	// * wrapper: Block of code to execute before and after the main block
	// * block: Main block of code to execute between wrapper executions
	// Returns:
	// * The result of the main block execution
	"wrap": { // **
		Argsn: 2,
		Doc:   "Executes a wrapper block before and after executing a main block, returning the result of the main block.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch wrap := arg0.(type) {
			case env.Block:
				switch bloc := arg1.(type) {
				case env.Block:
					ser := ps.Ser
					ps.Ser = wrap.Series
					EvalBlockInjMultiDialect(ps, arg0, true)
					if ps.ReturnFlag {
						return ps.Res
					}

					ps.Ser = bloc.Series
					EvalBlockInjMultiDialect(ps, arg0, true)
					if ps.ReturnFlag {
						return ps.Res
					}
					res := ps.Res

					ps.Ser = wrap.Series
					EvalBlockInjMultiDialect(ps, arg0, true)
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

	// Tests:
	// equal  { 20 .keep { + 202 } { + 101 } } 222
	// Args:
	// * value: Value to be injected into both blocks
	// * block1: First block whose result will be returned
	// * block2: Second block to execute after the first one
	// Returns:
	// * The result of the first block, ignoring the result of the second block
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
					EvalBlockInjMultiDialect(ps, arg0, true)
					if ps.ErrorFlag {
						return ps.Res
					}
					res := ps.Res
					ps.Ser = b2.Series
					EvalBlockInjMultiDialect(ps, arg0, true)
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

	// Tests:
	// equal   { x: 123 , change! 234 'x , x } 234
	// equal   { a:: 123 change! 333 'a a } 333
	// equal   { a:: 123 change! 124 'a } 1
	// equal   { a:: 123 change! 123 'a } 0
	// Args:
	// * value: New value to assign to the word
	// * word: Word whose value should be changed
	// Returns:
	// * Integer 1 if the value changed, 0 if the new value is the same as the old value
	"change!": { // ***
		Argsn: 2,
		Doc:   "Searches for a word and changes it's value in-place. If value changes returns true otherwise false",
		Pure:  false,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch arg := arg1.(type) {
			case env.Word:
				val, found, ctx := ps.Ctx.Get2(arg.Index)
				if found {
					ctx.Mod(arg.Index, arg0)
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

	// Tests:
	// equal   { set! { 123 234 } { a b }  b } 234
	// Args:
	// * values: Value or block of values to assign to the word(s)
	// * words: Word or block of words to be set
	// Returns:
	// * The value or block of values that was assigned
	"set!": { // ***
		Argsn: 2,
		Doc:   "Set word to value or words by deconstructing a block",
		Pure:  false,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch words := arg1.(type) {
			case env.Block:
				switch vals := arg0.(type) {
				case env.Block:
					for i, word_ := range words.Series.S {
						switch word := word_.(type) {
						case env.Word:
							// get nth value from values
							if len(vals.Series.S) < i {
								return MakeBuiltinError(ps, "More words than values.", "set!")
							}
							val := vals.Series.S[i]
							// if it exists then we set it to word from words
							ps.Ctx.Mod(word.Index, val)
							/* if res.Type() == env.ErrorType {
								return MakeBuiltinError(ps, res.(env.Error).Message, "set")
							}*/
						default:
							fmt.Println(word)
							return MakeBuiltinError(ps, "Only words in words block", "set!")
						}
					}
					return arg0
				default:
					return MakeArgError(ps, 2, []env.Type{env.BlockType}, "set!")
				}
			case env.Word:
				ps.Ctx.Mod(words.Index, arg0)
				return arg0
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType}, "set!")
			}
		},
	},

	// Tests:
	// equal   { x: 1 unset! 'x x: 2 } 2 ; otherwise would produce an error
	// Args:
	// * word: Word to be unset from the current context
	// Returns:
	// * Void value
	"unset!": { // ***
		Argsn: 1,
		Doc:   "Unset a word in current context, only meant to be used in console",
		Pure:  false,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch word := arg0.(type) {
			case env.Word:
				return ps.Ctx.Unset(word.Index, ps.Idx)
			default:
				return MakeArgError(ps, 1, []env.Type{env.WordType}, "unset!")
			}
		},
	},

	// Tests:
	// equal   { x: 123 val 'x } 123
	// equal   { x: 123 y: 'x val y } 123
	// Args:
	// * word: Word whose value should be retrieved
	// Returns:
	// * The value associated with the word in the current context
	"val": {
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
					return MakeBuiltinError(ps, "Word not found in contexts	", "get_")
				}
			case env.Opword:
				object, found := ps.Ctx.Get(w.Index)
				if found {
					return object
				} else {
					return MakeBuiltinError(ps, "Word not found in contexts	", "get_")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.WordType}, "get_")
			}
		},
	},

	// Tests:
	// equal   { kind 'person { name: "" age: 0 } |type? } 'kind
	// Args:
	// * name: Word that will be the name of the kind
	// * spec: Block containing the specification for the kind
	// Returns:
	// * A new kind object
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

	// Tests:
	// equal   { person: kind 'person { name: "" age: 0 } dict { "name" "John" "age" 30 } _>> person |type? } 'ctx
	// Args:
	// * value: Dict or context to convert
	// * kind: Kind to convert the value to
	// Returns:
	// * A new context of the specified kind
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

	// Tests:
	// equal   { person: kind 'person { name: "" age: 0 } person _<< dict { "name" "John" "age" 30 } |type? } 'ctx
	// Args:
	// * kind: Kind to convert the value to
	// * value: Dict or context to convert
	// Returns:
	// * A new context of the specified kind
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

	// Tests:
	// equal   { person: kind 'person { name: "" age: 0 } assure-kind dict { "name" "John" "age" 30 } person |type? } 'ctx
	// Args:
	// * value: Dict to convert to a specific kind
	// * kind: Kind to convert the value to
	// Returns:
	// * A new context of the specified kind
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

	// FUNCTIONALITY AROUND GENERIC METHODS
	// Tests:
	// equal   { generic 'integer 'add fn { a b } { a + b } |type? } 'function
	// Args:
	// * kind: Word representing the kind for which to register the function
	// * method: Word representing the method name
	// * function: Function to register for the kind and method
	// Returns:
	// * The registered function
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

	//
	// ##### Values and Types ##### ""
	//
	// Tests:
	// equal { to-word "test" } 'test
	// error { to-word 123 }
	// Args:
	// * value: String or word-like value to convert to a word
	// Returns:
	// * A word with the same name as the input value
	"to-word": {
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

	// Tests:
	// equal { to-integer "123" } 123
	// ; equal { to-integer "123.4" } 123
	// ; equal { to-integer "123.6" } 123
	// ; equal { to-integer "123.4" } 123
	// error { to-integer "abc" }
	// Args:
	// * value: String or decimal value to convert to an integer
	// Returns:
	// * An integer value
	"to-integer": {
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
			case env.Decimal:
				return *env.NewInteger(int64(addr.Value))
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "to-integer")
			}
		},
	},

	// Tests:
	// equal { to-decimal "123.4" } 123.4
	// error { to-decimal "abc" }
	// Args:
	// * value: String value to convert to a decimal
	// Returns:
	// * A decimal value
	"to-decimal": {
		Argsn: 1,
		Doc:   "Tries to change a Rye value (like string) to decimal.",
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
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "to-decimal")
			}
		},
	},

	// Tests:
	// equal { to-string 'test } "test"
	// equal { to-string 123 } "123"
	// equal { to-string 123.4 } "123.400000"
	// equal { to-string "test" } "test"
	// Args:
	// * value: Any Rye value to convert to a string
	// Returns:
	// * A string representation of the value
	"to-string": { // ***
		Argsn: 1,
		Doc:   "Tries to turn a Rye value to string.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewString(arg0.Print(*ps.Idx))
		},
	},

	// Tests:
	// equal { to-char 42 } "*"
	// error { to-char "*" }
	// Args:
	// * value: Integer representing an ASCII code point
	// Returns:
	// * A string containing the character corresponding to the ASCII code
	"to-char": { // ***
		Argsn: 1,
		Doc:   "Tries to turn a Rye value (like integer) to ascii character.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch value := arg0.(type) {
			case env.Integer:
				return *env.NewString(string(value.Value))
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType}, "to-char")
			}
		},
	},

	// Tests:
	// equal { list [ 1 2 3 ] |to-block |type? } 'block
	// equal  { list [ 1 2 3 ] |to-block |first } 1
	// Args:
	// * list: List value to convert to a block
	// Returns:
	// * A block containing the same elements as the input list
	"to-block": { // ***
		Argsn: 1,
		Doc:   "Turns a List to a Block",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch list := arg0.(type) {
			case env.List:
				return env.List2Block(ps, list)
			default:
				return MakeArgError(ps, 1, []env.Type{env.DictType}, "to-context")
			}
		},
	},

	// Tests:
	// equal   { dict [ "a" 1 "b" 2 "c" 3 ] |to-context |type? } 'ctx   ; TODO - rename ctx to context in Rye
	// ; equal   { dict [ "a" 1 ] |to-context do\in { a } } '1
	// Args:
	// * dict: Dict value to convert to a context
	// Returns:
	// * A context with the same keys and values as the input dict
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

	// Tests:
	// equal   { is-string "test" } 1
	// equal   { is-string 'test } 0
	// equal   { is-string 123 } 0
	// Args:
	// * value: Any Rye value to test
	// Returns:
	// * Integer 1 if the value is a string, 0 otherwise
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

	// Tests:
	// equal   { is-integer 123 } 1
	// equal   { is-integer 123.4 } 0
	// equal   { is-integer "123" } 0
	// Args:
	// * value: Any Rye value to test
	// Returns:
	// * Integer 1 if the value is an integer, 0 otherwise
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

	// Tests:
	// equal   { is-decimal 123.0 } 1
	// equal   { is-decimal 123 } 0
	// equal   { is-decimal "123.4" } 0
	// Args:
	// * value: Any Rye value to test
	// Returns:
	// * Integer 1 if the value is a decimal, 0 otherwise
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

	// Tests:
	// equal   { is-number 123 } 1
	// equal   { is-number 123.4 } 1
	// equal   { is-number "123" } 0
	// Args:
	// * value: Any Rye value to test
	// Returns:
	// * Integer 1 if the value is a number (integer or decimal), 0 otherwise
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

	// Tests:
	// equal   { to-uri "https://example.com" } https://example.com
	// ; error { to-uri "not-uri" }
	"to-uri": { // ** TODO-FIXME: return possible failures
		Argsn: 1,
		Doc:   "Tries to change Rye value to an URI.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch val := arg0.(type) {
			case env.String:
				return *env.NewUri1(ps.Idx, val.Value) // TODO turn to switch
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "to-uri")
			}

		},
	},

	// Tests:
	// equal   { to-file "example.txt" } %example.txt
	// equal { to-file 123 } %123
	"to-file": { // **  TODO-FIXME: return possible failures
		Argsn: 1,
		Doc:   "Tries to change Rye value to a file.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch val := arg0.(type) {
			case env.Integer:
				return *env.NewFileUri(ps.Idx, strconv.Itoa(int(val.Value))) // TODO turn to switch
			case env.String:
				return *env.NewFileUri(ps.Idx, val.Value) // TODO turn to switch
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "to-file")
			}
		},
	},
	// Tests:
	// equal   { type? "test" } 'string
	// equal   { type? 123.4 } 'decimal
	// Args:
	// * value: Any Rye value to check the type of
	// Returns:
	// * A word representing the type of the value
	"type?": { // ***
		Argsn: 1,
		Doc:   "Returns the type of Rye value as a word.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewWord(int(arg0.Type()))
		},
	},

	// Tests:
	// equal   { kind? %file } 'file-schema
	// Args:
	// * value: Any Rye value to check the kind of
	// Returns:
	// * A word representing the kind of the value
	"kind?": { // ***
		Argsn: 1,
		Doc:   "Returns the kind of Rye value as a word.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewWord(int(arg0.GetKind()))
		},
	},

	// Tests:
	// equal   { types? { "test" 123 } } { string integer }
	// Args:
	// * collection: A block, table, or table row containing values to check types of
	// Returns:
	// * A block of words representing the types of each value in the collection
	"types?": { // TODO
		Argsn: 1,
		Doc:   "Returns the types of Rye values in a block or table row as a block of words.",
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
			case env.Table:
				l := len(list.Rows[0].Values)
				newl := make([]env.Object, l)
				for i := 0; i < l; i++ {
					newl[i] = *env.NewWord(int(env.ToRyeValue(list.Rows[0].Values[i]).Type()))
				}
				return *env.NewBlock(*env.NewTSeries(newl))
			case env.TableRow:
				l := len(list.Values)
				newl := make([]env.Object, l)
				for i := 0; i < l; i++ {
					newl[i] = *env.NewWord(int(env.ToRyeValue(list.Values[i]).Type()))
				}
				return *env.NewBlock(*env.NewTSeries(newl))
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType, env.TableType, env.TableRowType}, "types?")
			}
		},
	},

	// Tests:
	// equal { dump 123 } "123"
	// equal { dump "string" } `"string"`
	// equal { does { 1 } |dump } "fn { } { 1 }"
	// Args:
	// * value: Any Rye value to dump as code
	// Returns:
	// * A string containing the Rye code representation of the value
	"dump": { // *** currently a concept in testing ... for getting a code of a function, maybe same would be needed for context?
		Argsn: 1,
		Doc:   "Returns (dumps) Rye code representing the object.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewString(arg0.Dump(*ps.Idx))
		},
	},

	// Tests:
	// equal  { mold 123 } "123"
	// equal  { mold { 123 } } "{ 123 }"
	// Args:
	// * value: Any Rye value to convert to a string representation
	// Returns:
	// * A string containing the representation of the value
	"mold": { // **
		Argsn: 1,
		Doc:   "Turn value to it's string representation.",
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			// fmt.Println()
			return *env.NewString(arg0.Dump(*env1.Idx))
		},
	},

	// Tests:
	// equal  { mold\nowrap 123 } "123"
	// equal  { mold\nowrap { 123 } } "123"
	// equal  { mold\nowrap { 123 234 } } "123 234"
	// Args:
	// * value: Any Rye value to convert to a string representation
	// Returns:
	// * A string containing the representation of the value without block wrapping
	"mold\\nowrap": { // **
		Argsn: 1,
		Doc:   "Turn value to it's string representation. Doesn't wrap the blocks",
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			// fmt.Println(arg0)
			str := arg0.Dump(*env1.Idx)
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

	// Tests:
	// equal  { save\current |type? } 'integer
	// Args:
	// * None
	// Returns:
	// * Integer 1 on success
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

	// Tests:
	// equal  { save\current\secure |type? } 'integer
	// Args:
	// * None
	// Returns:
	// * Integer 1 on success
	"save\\current\\secure": {
		Argsn: 0,
		Doc:   "Saves current state of the program to a file with password protection.",
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

	// Tests:
	// equal   { x: private { doc! "some doc" doc? } } "some doc"
	// Args:
	// * doc: String to set as the docstring for the current context
	// Returns:
	// * Integer 1 on success
	"doc!": { // ***
		Argsn: 1,
		Doc:   "Sets docstring of the current context.",
		Pure:  true,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch d := arg0.(type) {
			case env.String:
				env1.Ctx.Doc = d.Value
				return *env.NewInteger(1)
			default:
				return MakeArgError(env1, 1, []env.Type{env.StringType}, "doc!")
			}
		},
	},

	// Tests:
	// equal   { x: private { doc! "some doc" doc? } } "some doc"
	// Args:
	// * None
	// Returns:
	// * String containing the docstring of the current context
	"doc?": { // ***
		Argsn: 0,
		Doc:   "Gets docstring of the current context.",
		Pure:  true,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewString(env1.Ctx.Doc)
		},
	},

	// Tests:
	// equal   { x: context { doc! "some doc" } doc\of? x } "some doc"
	// Args:
	// * value: Function, builtin, or context to get the docstring from
	// Returns:
	// * String containing the docstring of the provided value
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
	// Tests:
	// equal   { is-ref ref { 1 2 3 } } 1
	// Args:
	// * value: Value to make mutable
	// Returns:
	// * A mutable reference to the value
	"ref": {
		Argsn: 1,
		Doc:   "Makes a value mutable instead of immutable",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch sp := arg0.(type) {
			case env.Table:
				return &sp
			case *env.Table:
				return sp
			case env.Dict:
				return &sp
			case env.List:
				return &sp
			case env.Block:
				return &sp
			case env.String:
				return &sp
			case env.RyeCtx:
				return &sp
			case env.Native:
				sp.Value = env.ReferenceAny(sp.Value)
				return &sp
			default:
				return MakeArgError(ps, 1, []env.Type{env.TableType, env.DictType, env.ListType, env.BlockType, env.StringType, env.NativeType, env.CtxType}, "deref")
			}
		},
	},

	// Tests:
	// equal   { is-ref deref ref { 1 2 3 } } 0
	// Args:
	// * value: Mutable reference to make immutable
	// Returns:
	// * An immutable copy of the value
	"deref": {
		Argsn: 1,
		Doc:   "Makes a value again immutable",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch sp := arg0.(type) {
			case env.Table:
				return sp
			case *env.Table:
				return *sp
			case *env.Dict:
				return *sp
			case *env.List:
				return *sp
			case *env.Block:
				return *sp
			case *env.RyeCtx:
				return *sp
			case *env.String:
				return *sp
			case *env.Native:
				sp.Value = env.DereferenceAny(sp.Value)
				return *sp
			default:
				return MakeArgError(ps, 1, []env.Type{env.TableType, env.DictType, env.ListType, env.BlockType, env.StringType, env.NativeType, env.CtxType}, "deref")
			}
		},
	},

	// Tests:
	// equal  { ref { } |is-ref } true
	// equal  { { } |is-ref } false
	// Args:
	// * value: Any value to check if it's a reference
	// Returns:
	// * Integer 1 if the value is a reference, 0 otherwise
	"is-ref": { // **
		Argsn: 1,
		Doc:   "Checks if a value is a mutable reference.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			// fmt.Println(arg0.Inspect(*ps.Idx))
			if env.IsPointer(arg0) {
				return env.NewInteger(1)
			} else {
				return env.NewInteger(0)
			}
		},
	},

	// Tests:
	// equal { dict { "a" 123 } -> "a" } 123
	// Args:
	// * block: Block containing alternating keys and values
	// Returns:
	// * A new Dict with the specified keys and values
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

	// Tests:
	// equal { list { "a" 123 } -> 0 } "a"
	// Args:
	// * block: Block containing values to put in the list
	// Returns:
	// * A new List with the values from the block
	"list": {
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

	//
	// ##### Printing ##### ""
	//
	// Tests:
	// stdout { prns "xy" } "xy "
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

	// Tests:
	// stdout { prn "xy" } "xy"
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

	// Tests:
	// stdout { print "xy" } "xy\n"
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

	// Tests:
	// equal { format 123  "num: %d" } "num: 123"
	"format": {
		Argsn: 2,
		Doc:   "Formats a value according to Go-s sprintf format",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			var res string
			switch arg := arg1.(type) {
			case env.String:
				switch val := arg0.(type) {
				case env.String:
					res = fmt.Sprintf(arg.Value, val.Value)
				case env.Integer:
					res = fmt.Sprintf(arg.Value, val.Value)
				case env.Decimal:
					res = fmt.Sprintf(arg.Value, val.Value)
					// TODO make option with multiple values and block as second arg
				default:
					return MakeArgError(ps, 1, []env.Type{env.StringType, env.DecimalType, env.IntegerType}, "format")
				}
				return *env.NewString(res)
			default:
				return MakeArgError(ps, 2, []env.Type{env.StringType}, "format")
			}
		},
	},

	// Tests:
	// stdout { prnf 123 "num: %d" } "num: 123"
	"prnf": { // **
		Argsn: 2,
		Doc:   "Formats a value according to Go-s sprintf format and prn-s it",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch arg := arg1.(type) {
			case env.String:
				switch val := arg0.(type) {
				case env.String:
					fmt.Printf(arg.Value, val.Value)
				case env.Integer:
					fmt.Printf(arg.Value, val.Value)
				case env.Decimal:
					fmt.Printf(arg.Value, val.Value)
					// TODO make option with multiple values and block as second arg
				default:
					return MakeArgError(ps, 1, []env.Type{env.StringType, env.DecimalType, env.IntegerType}, "prnf")
				}
				return arg0
			default:
				return MakeArgError(ps, 2, []env.Type{env.StringType}, "prnf")
			}
		},
	},

	// Tests:
	// equal   { embed 101 "val {}" } "val 101"
	"embed": { // **
		Argsn: 2,
		Doc:   "Embeds a value into a string with {} placeholder.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch val := arg1.(type) {
			case env.String:
				vals := arg0.Print(*ps.Idx)
				news := strings.ReplaceAll(val.Value, "{}", vals)
				return *env.NewString(news)
			case env.Uri:
				vals := arg0.Print(*ps.Idx)
				news := strings.ReplaceAll(val.Path, "{}", vals)
				return *env.NewUri(ps.Idx, val.Scheme, news)
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType, env.UriType}, "embed")
			}
		},
	},

	// Tests:
	// stdout  { prnv 101 "val {}" } "val 101"
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

	// Tests:
	// stdout  { printv 101 "val {}" } "val 101\n"
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
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "printv")
			}
			return arg0
		},
	},

	// Tests:
	// stdout  { print\ssv { 101 "asd" } } "101 asd\n"
	"print\\ssv": {
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

	// Tests:
	// stdout  { print\csv { 101 "asd" } } "101,asd\n"
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

	// Tests:
	// stdout  { probe 101 } "[Integer: 101]\n"
	"probe": { // **
		Argsn: 1,
		Doc:   "Prints information about a value.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			p := ""
			if env.IsPointer(arg0) {
				p = "REF"
			}
			fmt.Println(p + arg0.Inspect(*ps.Idx))
			return arg0
		},
	},

	// Tests:
	// equal  { inspect 101 } "[Integer: 101]"
	"inspect": { // **
		Argsn: 1,
		Doc:   "Returs information about a value.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewString(arg0.Inspect(*ps.Idx))
		},
	},

	// Tests:
	// ; equal  { esc "[33m" } "\033[33m"   ; we can't represent hex or octal in strings yet
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

	// Tests:
	// ; equal  { esc-val "[33m" "Error" } "\033[33mError"  ; we can't represent hex or octal in strings yet
	"esc-val": {
		Argsn: 2,
		Doc:   "Escapes a value and adds a newline.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch base := arg1.(type) {
			case env.String:
				vals := arg0.Print(*ps.Idx)
				news := strings.ReplaceAll(base.Value, "{}", vals)
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
			case *env.Block:
				obj, esc := term.DisplayBlock(*bloc, ps.Idx)
				if !esc {
					return obj
				}
			case env.Dict:
				obj, esc := term.DisplayDict(bloc, ps.Idx)
				if !esc {
					return obj
				}
			case *env.Dict:
				obj, esc := term.DisplayDict(*bloc, ps.Idx)
				if !esc {
					return obj
				}
			case env.Table:
				obj, esc := term.DisplayTable(bloc, ps.Idx)
				if !esc {
					return obj
				}
			case *env.Table:
				obj, esc := term.DisplayTable(*bloc, ps.Idx)
				if !esc {
					return obj
				}
			case env.TableRow:
				obj, esc := term.DisplayTableRow(bloc, ps.Idx)
				if !esc {
					return obj
				}
			case *env.TableRow:
				obj, esc := term.DisplayTableRow(*bloc, ps.Idx)
				if !esc {
					return obj
				}
			}
			return arg0
		},
	},
	"_..": {
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
			case *env.Block:
				obj, esc := term.DisplayBlock(*bloc, ps.Idx)
				if !esc {
					return obj
				}
			case env.Dict:
				obj, esc := term.DisplayDict(bloc, ps.Idx)
				if !esc {
					return obj
				}
			case *env.Dict:
				obj, esc := term.DisplayDict(*bloc, ps.Idx)
				if !esc {
					return obj
				}
			case env.Table:
				obj, esc := term.DisplayTable(bloc, ps.Idx)
				if !esc {
					return obj
				}
			case *env.Table:
				obj, esc := term.DisplayTable(*bloc, ps.Idx)
				if !esc {
					return obj
				}
			case env.TableRow:
				obj, esc := term.DisplayTableRow(bloc, ps.Idx)
				if !esc {
					return obj
				}
			case *env.TableRow:
				obj, esc := term.DisplayTableRow(*bloc, ps.Idx)
				if !esc {
					return obj
				}
			}
			return arg0
		},
	},
	"display\\custom": {
		Argsn: 2,
		Doc:   "Work in progress Interactively displays a value.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			// This is temporary implementation for experimenting what it would work like at all
			// later it should belong to the object (and the medium of display, terminal, html ..., it's part of the frontend)
			term.SaveCurPos()
			switch fnc := arg1.(type) {
			case env.Function:
				switch bloc := arg0.(type) {
				case env.Table:
					obj, esc := term.DisplayTableCustom(
						bloc,
						func(row env.Object, iscurr env.Integer) { CallFunctionArgsN(fnc, ps, ps.Ctx, row, iscurr) },
						ps.Idx)
					if !esc {
						return obj
					}
				case *env.Table:
					obj, esc := term.DisplayTableCustom(
						*bloc,
						func(row env.Object, iscurr env.Integer) { CallFunctionArgsN(fnc, ps, ps.Ctx, row, iscurr) },
						ps.Idx)
					if !esc {
						return obj
					}
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

	"import": { // **
		Argsn: 1,
		Doc:   "Imports a file, loads and does it from script local path.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Uri:
				block_, script_ := LoadScriptLocalFile(ps, s1)
				/*
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
				*/
				ps.Res = EvaluateLoadedValue(ps, block_, script_, false)
				/* switch block := block_.(type) {
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
				} */
				ps.ScriptPath = script_
				//ps = env.AddToProgramState(ps, block.Series, genv)
				return ps.Res
			default:
				return MakeArgError(ps, 1, []env.Type{env.UriType}, "import")
			}
		},
	},

	"import\\live": { // **
		Argsn: 1,
		Doc:   "Imports a file, loads and does it from script local path.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Uri:
				block_, script_ := LoadScriptLocalFile(ps, s1)
				ps.Res = EvaluateLoadedValue(ps, block_, script_, false)
				ps.LiveObj.Add(s1.GetPath()) // add to watcher
				ps.ScriptPath = script_
				//ps = env.AddToProgramState(ps, block.Series, genv)
				return ps.Res
			default:
				return MakeArgError(ps, 1, []env.Type{env.UriType}, "import\\live")
			}
		},
	},

	// Tests:
	// equal  { load " 1 2 3 " |third } 3
	// equal  { load "{ 1 2 3 }" |first |third } 3
	"load": { // **
		Argsn: 1,
		Doc:   "Loads a string into Rye values.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.String:
				block := loader.LoadStringNEW(s1.Value, false, ps)
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
				return MakeArgError(ps, 1, []env.Type{env.StringType, env.UriType}, "import\\live")
			}
		},
	},

	// TODO -- refactor load variants so they use common function LoadString and LoadFile

	"load\\mod": { // **
		Argsn: 1,
		Doc:   "Loads a string into Rye values. During load it allows modification of words.",
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
				ps.AllowMod = true
				ps.ScriptPath = s1.GetPath()
				block := loader.LoadStringNEW(str, false, ps)
				ps.AllowMod = false
				ps.ScriptPath = scrip
				//ps = env.AddToProgramState(ps, block.Series, genv)
				return block
			default:
				ps.FailureFlag = true
				return env.NewError("Must be string or file TODO")
			}
		},
	},

	"load\\live": { // **
		Argsn: 1,
		Doc:   "Loads a string into Rye values. During load it allows modification of words.",
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
					ps.LiveObj.Add(s1.GetPath()) // add to watcher
					if err != nil {
						return makeError(ps, err.Error())
					}
					str = string(b) // convert content to a 'string'
				}
				scrip := ps.ScriptPath
				ps.AllowMod = true
				ps.ScriptPath = s1.GetPath()
				block := loader.LoadStringNEW(str, false, ps)
				ps.AllowMod = false
				ps.ScriptPath = scrip
				//ps = env.AddToProgramState(ps, block.Series, genv)
				return block
			default:
				ps.FailureFlag = true
				return env.NewError("Must be string or file TODO")
			}
		},
	},

	"load\\sig": {
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

	//
	// ##### Flow control ##### ""
	//
	// Tests:
	// equal  { if true { 222 } } 222
	// equal  { if false { 333 } } false
	// error  { if 1 { 222 } }
	// error  { if 0 { 333 } }
	"if": { // **
		Argsn: 2,
		Doc:   "Basic conditional. Takes a boolean condition and a block of code.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			// function accepts 2 args. arg0 is a boolean value, arg1 is a block of code

			// Check if the first argument is a boolean
			switch cond := arg0.(type) {
			case env.Boolean:
				// we switch on the type of second argument, so far it should be block (later we could accept native and function)
				switch bloc := arg1.(type) {
				case env.Block:
					// if cond.Value is true, execute the block
					if cond.Value {
						// we store current series (block of code with position we are at) to temp 'ser'
						ser := ps.Ser
						// we set ProgramStates series to series ob the block
						ps.Ser = bloc.Series
						// we eval the block (current context / scope stays the same as it was in parent block)
						// Inj means we inject the condition value into the block, because it costs us very little. we could do "if name { .print }"
						EvalBlockInjMultiDialect(ps, arg0, true)
						// we set temporary series back to current program state
						ps.Ser = ser
						// we return the last return value (the return value of executing the block)
						return ps.Res
					}
					return *env.NewBoolean(false)
				default:
					// if it's not a block we return error
					ps.FailureFlag = true
					return MakeArgError(ps, 2, []env.Type{env.BlockType}, "if")
				}
			default:
				// if it's not a boolean we return error
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.BooleanType}, "if")
			}
		},
	},

	// Tests:
	// equal  { x: does { ^if 1 { 222 } 555 } x } 222
	// equal  { x: does { ^if 0 { 333 } 444 } x } 444
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

	// Tests:
	// equal  { either true { 222 } { 333 } } 222
	// equal  { either false { 222 } { 333 } } 333
	// error  { either 1 { 222 } { 333 } }
	// error  { either 0 { 222 } { 333 } }
	"either": { // **
		Argsn: 3,
		Doc:   "The if/else conditional. Takes a boolean condition and true and false blocks of code.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			// Check if the first argument is a boolean
			switch cond := arg0.(type) {
			case env.Boolean:
				switch bloc1 := arg1.(type) {
				case env.Block:
					switch bloc2 := arg2.(type) {
					case env.Block:
						ser := ps.Ser
						if cond.Value {
							ps.Ser = bloc1.Series
							ps.Ser.Reset()
						} else {
							ps.Ser = bloc2.Series
							ps.Ser.Reset()
						}
						EvalBlockInjMultiDialect(ps, arg0, true)
						ps.Ser = ser
						return ps.Res
					default:
						return MakeArgError(ps, 3, []env.Type{env.BlockType}, "either")
					}
				case env.Object:
					switch bloc2 := arg2.(type) {
					case env.Object: // If true value is not block then also false value will be treated as literal
						if cond.Value {
							return bloc1
						} else {
							return bloc2
						}
					default:
						return MakeBuiltinError(ps, "Third argument must be Object Type.", "either")
					}
				default:
					return MakeBuiltinError(ps, "Second argument must be Block or Object Type.", "either")
				}
			default:
				// If it's not a boolean, return an error
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.BooleanType}, "either")
			}
		},
	},

	// ; equal  { fail 404 |^tidy\switch { 404 { "ER1" } 305 { "ER2" } } } "ER1"
	"^tidy\\switch": {
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
						EvalBlockInjMultiDialect(ps, arg0, true)
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

	// Tests:
	// equal  { switch 101 { 101 { 111 } 202 { 222 } } } 111
	// equal  { switch 202 { 101 { 111 } 202 { 222 } } } 222
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
						EvalBlockInjMultiDialect(ps, arg0, true)
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

	// Tests:
	// equal  { cases 0 { { 1 > 0 } { + 100 } { 2 > 1 } { + 1000 } } } 1100
	// equal  { cases 0 { { 1 > 0 } { + 100 } { 2 < 1 } { + 1000 } } } 100
	// equal  { cases 0 { { 1 < 0 } { + 100 } { 2 > 1 } { + 1000 } } } 1000
	// equal  { cases 0 { { 1 < 0 } { + 100 } { 2 < 1 } { + 1000 } } } 0
	// equal  { cases 1 { { 1 > 0 } { + 100 } { 2 < 1 } { + 1000 } _ { * 3 } } } 101
	// equal  { cases 1 { { 1 < 0 } { + 100 } { 2 > 1 } { + 1000 } _ { * 3 } } } 1001
	// equal  { cases 1 { { 1 < 0 } { + 100 } { 2 < 1 } { + 1000 } _ { * 3 } } } 3
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
						EvalBlockInjMultiDialect(ps, cumul, true)
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

	// Tests:
	// equal  { do { 123 + 123 } } 246
	// error  { do { 123 + } }
	// equal  { do { _+ _+ 12 23 34 } } 69
	// equal  { do { 12 * 23 |+ 34 } } 310
	// equal  { do { ( 12 * 23 ) + 34 } } 310
	// equal  { do { 12 * 23 | + 34 } } 310
	// equal  { do { 12 * 23 :a + 34 } } 310
	// equal  { do { 12 * 23 :a a + 34 } } 310
	"do": {
		Argsn: 1,
		Doc:   "Takes a block of code and does (runs) it.",
		Pure:  true,
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

	// Tests:
	// equal  { try { 123 + 123 } } 246
	// equal  { try { 123 + "asd" } \type? } 'error
	// equal  { try { 123 + } \type? } 'error
	"try": { // **
		Argsn: 1,
		Doc:   "Takes a block of code and does (runs) it.",
		Pure:  true,
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
				return MakeArgError(ps, 1, []env.Type{env.BlockType}, "try")
			}
		},
	},

	// Tests:
	// equal  { with 100 { + 11 } } 111
	// equal  { with 100 { + 11 , * 3 } } 300
	"with": { // **
		AcceptFailure: true,
		Doc:           "Takes a value and a block of code. It does the code with the value injected.",
		Argsn:         2,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch bloc := arg1.(type) {
			case env.Block:
				ser := ps.Ser
				ps.Ser = bloc.Series
				EvalBlockInjMultiDialect(ps, arg0, true)
				ps.Ser = ser
				return ps.Res
			default:
				return MakeArgError(ps, 2, []env.Type{env.BlockType}, "with")
			}
		},
	},

	// Tests:
	// equal  { c: context { x: 100 } do\in c { x * 9.99 } } 999.0
	// equal  { c: context { x: 100 } do\in c { inc! 'x } } 101
	// equal  { c: context { x: 100 } do\in c { x:: 200 } c/x } 200
	// equal  { c: context { x: 100 } do\in c { x:: 200 , x } } 200
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
					return MakeArgError(ps, 2, []env.Type{env.BlockType}, "do\\in")
				}
			default:
				ps.ErrorFlag = true
				return MakeArgError(ps, 1, []env.Type{env.CtxType}, "do\\in")
			}

		},
	},

	// Tests:
	// equal  { c: context { x: 100 } try\in c { x * 9.99 } } 999.0
	// equal  { c: context { x: 100 } try\in c { inc! 'x } } 101
	// equal  { c: context { x: 100 } try\in c { x:: 200 , x } } 200
	// equal  { c: context { x: 100 } try\in c { x:: 200 } c/x } 200
	// equal  { c: context { x: 100 } try\in c { inc! 'y } |type? } 'error
	"try\\in": { // **
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
					return MakeArgError(ps, 2, []env.Type{env.BlockType}, "try\\in")
				}
			default:
				ps.ErrorFlag = true
				return MakeArgError(ps, 1, []env.Type{env.CtxType}, "try\\in")
			}

		},
	},

	// Tests:
	// equal  { c: context { x: 100 } do\par c { x * 9.99 } } 999.0
	// equal  { c: context { x: 100 } do\par c { inc! 'x } } 101
	// equal  { c: context { x: 100 } do\par c { x:: 200 , x } } 200
	// equal  { c: context { x: 100 } do\par c { x:: 200 } c/x } 100
	"do\\par": { // **
		Argsn: 2,
		Doc:   "Takes a Context and a Block. It Does a block in current context but with parent a given Context.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch ctx := arg0.(type) {
			case env.RyeCtx:
				switch bloc := arg1.(type) {
				case env.Block:
					ser := ps.Ser
					ps.Ser = bloc.Series
					tempCtx := &ctx
					for {
						// fmt.Println("UP ->")
						// do we change parents globally (and we have to fix them back) or just in local copy of a value? check
						if tempCtx.Parent == ps.Ctx {
							// fmt.Println("ISTI PARENT")
							// temp = ctx.Parent
							tempCtx.Parent = ps.Ctx.Parent
							break
						}
						if tempCtx.Parent != nil {
							tempCtx = tempCtx.Parent
						} else {
							break
						}
					}
					//var temp *env.RyeCtx
					// set argument's parent context to current parent context
					// }
					// set argument context as parent
					temp := ps.Ctx.Parent
					ps.Ctx.Parent = &ctx
					EvalBlock(ps)
					// if temp != nil {
					ps.Ctx.Parent = temp
					// ctx.Parent = temp
					// }
					ps.Ser = ser
					return ps.Res
				default:
					ps.ErrorFlag = true
					return MakeArgError(ps, 2, []env.Type{env.BlockType}, "do\\par")
				}
			default:
				ps.ErrorFlag = true
				return MakeArgError(ps, 1, []env.Type{env.CtxType}, "do\\par")
			}

		},
	},

	// Tests:
	// equal { capture-stdout { print "hello" } } "hello\n"
	// equal { capture-stdout { loop 3 { prns "x" } } } "x x x "
	"capture-stdout": { // **
		Argsn: 1,
		Doc:   "Takes a block of code and does (runs) it.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch bloc := arg0.(type) {
			case env.Block:

				old := os.Stdout // keep backup of the real stdout
				r, w, _ := os.Pipe()
				os.Stdout = w

				outC := make(chan string, 1000)
				g := errgroup.Group{}
				// copy the output in a separate goroutine so printing can't block indefinitely
				g.Go(func() error {
					/* var buf bytes.Buffer
					reader := bufio.NewReader(r)
					for {
						line, err := reader.ReadString('\n')
						if err != nil {
							if err == io.EOF {
								break
							}
							// Handle error
							fmt.Println(err)
							break
						}
						buf.WriteString(line)
					}
					outC <- buf.String()
					*/
					var buf bytes.Buffer
					_, err := io.Copy(&buf, r)
					if err != nil {
						w.Close()
						os.Stdout = old // restoring the real stdout
						fmt.Println(err.Error())
						return err
					}
					outC <- buf.String()
					return nil
				})

				ser := ps.Ser
				ps.Ser = bloc.Series
				EvalBlock(ps)
				ps.Ser = ser

				// back to normal state
				w.Close()
				os.Stdout = old // restoring the real stdout

				if err := g.Wait(); err != nil {
					return MakeBuiltinError(ps, fmt.Sprintf("Error reading stdout: %v", err), "capture-stdout")
				}
				out := <-outC

				if ps.ErrorFlag {
					return ps.Res
				}
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

	// Tests:
	// equal { time-it { sleep 100 } } 100
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

	// Tests:
	// equal { x: 1 y: 2 vals { x y } } { 1 2 }
	// equal { x: 1 y: 2 vals { 1 y } } { 1 2 }
	// equal { x: 1 y: 2 try { vals { z y } } |type? } 'error
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
					if checkErrorReturnFlag(ps) {
						return ps.Res
					}
					res = append(res, ps.Res)
					// check and raise the flags if needed if true (error) return
					//if checkFlagsAfterBlock(ps, 101) {
					//	return ps
					//}
					// if return flag was raised return ( errorflag I think would return in previous if anyway)
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

	// Tests:
	// equal { x: 1 y: 2 vals\with 10 { + x , * y } } { 11 20 }
	// equal { x: 1 y: 2 vals\with 100 { + 10 , * 8.9 } } { 110 890.0 }
	"vals\\with": {
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

	// EXPERIMENTAL: COLLECTION AND RETURNING FUNCTIONS ... NOT DETERMINED IF WILL BE INCLUDED YET

	"returns!": {
		Argsn: 1,
		Doc:   "Sets up a value to return at the end of function.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			ps.ForcedResult = arg0
			return arg0
		},
	},

	"collect!": {
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

	"collect-key!": {
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

	"collect-update-key!": {
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

	"pop-collected!": {
		Argsn: 0,
		Doc:   "Returns the implicit collected data structure and resets it.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			result := ps.ForcedResult
			ps.ForcedResult = nil
			return result
		},
	},

	// CONTEXT

	// Tests:
	// equal { c: context { x: 9999 , incr: fn\in { } current { x:: inc x } } c/incr c/x } 10000
	"current": { // **
		Argsn: 0,
		Doc:   "Returns current context.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *ps.Ctx
		},
	},

	// Tests:
	// equal { y: 99 c: context { incr: fn\in { } parent { y:: inc y } } c/incr y } 100
	"parent": { // **
		Argsn: 0,
		Doc:   "Returns parent context of the current context.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *ps.Ctx.Parent
		},
	},

	// Tests:
	// equal { ct: context { p: 123 } parent\of ct |= current } 1
	"parent\\of": {
		Argsn: 1,
		Doc:   "Returns parent context of the current context.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch c := arg0.(type) {
			case env.RyeCtx:
				return *c.Parent
			case *env.RyeCtx:
				return *c.Parent
			default:
				return MakeArgError(ps, 1, []env.Type{env.CtxType}, "parent?")
			}
		},
	},

	"lc": {
		Argsn: 0,
		Doc:   "Lists words in current context",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			fmt.Println(ps.Ctx.Preview(*ps.Idx, ""))
			return env.Void{}
		},
	},

	"lc\\data": {
		Argsn: 0,
		Doc:   "Lists words in current context",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return ps.Ctx.GetWords(*ps.Idx)
		},
	},

	"lc\\data\\": {
		Argsn: 1,
		Doc:   "Lists words in current context",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch c := arg0.(type) {
			case env.RyeCtx:
				return c.GetWords(*ps.Idx)
			case *env.RyeCtx:
				return c.GetWords(*ps.Idx)
			default:
				return MakeArgError(ps, 1, []env.Type{env.CtxType}, "parent?")
			}
		},
	},

	"lcp": {
		Argsn: 0,
		Doc:   "Lists words in current context",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			if ps.Ctx.Parent != nil {
				fmt.Println(ps.Ctx.Parent.Preview(*ps.Idx, ""))
			} else {
				fmt.Println("No parent")
			}
			return env.Void{}
		},
	},

	"lc\\": {
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

	"lcp\\": {
		Argsn: 1,
		Doc:   "Lists words in current context with string filter",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.String:
				if ps.Ctx.Parent != nil {
					fmt.Println(ps.Ctx.Parent.Preview(*ps.Idx, s1.Value))
				} else {
					fmt.Println("No parent")
				}
				return env.Void{}
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "lsp\\")
			}
		},
	},

	"cc": {
		Argsn: 1,
		Doc:   "Change to context",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.RyeCtx:
				// s1.Parent = ps.Ctx // TODO ... this is temporary so ccp works, but some other method must be figured out as changing the parent is not OK
				ps.Ctx = &s1
				return s1
			case *env.RyeCtx:
				// s1.Parent = ps.Ctx // TODO ... this is temporary so ccp works, but some other method must be figured out as changing the parent is not OK
				ps.Ctx = s1
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
				ret := ps.Ctx.Set(word.Index, newctx)
				s, ok := ret.(env.Error)
				if ok {
					return s
				}
				ctx := ps.Ctx
				ps.Ctx = newctx // make new context with current par
				return *ctx
			default:
				return MakeArgError(ps, 1, []env.Type{env.WordType}, "mkcc")
			}
		},
	},

	"lk": {
		Argsn: 0,
		Doc:   "Lists available kinds",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			fmt.Println(ps.Gen.PreviewKinds(*ps.Idx, ""))
			return env.Void{}
		},
	},

	"lg": {
		Argsn: 1,
		Doc:   "Lists generic words related to specific kind",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Word:
				fmt.Println(ps.Gen.PreviewMethods(*ps.Idx, s1.Index, ""))
				return env.Void{}
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "ls\\")
			}
		},
	},

	//
	// ##### Other ##### "functions related to date and time"
	//
	// return , error , failure functions
	// Tests:
	// equal { x: fn { } { return 101 202 } x } 101
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
			// close the keyboard opened for terminal
			// fmt.Println("Closing keyboard in Exit")
			// keyboard.Close()
			util.BeforeExit()
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

	/* Terminal functions .. move to it's own later */

	// Tests:
	// equal { cmd `echo "hello"` } 1
	"cmd": {
		Argsn: 1,
		Doc:   "Execute a shell command.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s0 := arg0.(type) {
			case env.String:

				r := exec.Command("/bin/bash", "-c", s0.Value) //nolint: gosec
				// stdout, stderr := r.Output()
				r.Stdout = os.Stdout
				r.Stderr = os.Stderr

				err := r.Run()
				if err != nil {
					fmt.Println(err)
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "cmd\\capture")
			}
			return env.NewInteger(1)
		},
	},

	"cmd\\capture": {
		Argsn: 1,
		Doc:   "Execute a shell command and capture the output, return it as string",
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
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "cmd\\capture")
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

	// Tests:
	// equal { rye .type? } 'native
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
			var firstArg int
			if ps.Embedded {
				firstArg = 1
			} else {
				firstArg = 2
			}
			return util.StringToFieldsWithQuoted(strings.Join(os.Args[firstArg:], " "), " ", "\"")
			// block, _ := loader.LoadString(os.Args[0], false)
			// return block
		},
	},
	"Rye-itself//args\\raw": {
		Argsn: 1,
		Doc:   "",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			var firstArg int
			if ps.Embedded {
				firstArg = 1
			} else {
				firstArg = 2
			}
			if len(os.Args) > 1 {
				return *env.NewString(strings.Join(os.Args[firstArg:], " "))
			} else {
				return *env.NewString("")
			}
			// block, _ := loader.LoadString(os.Args[0], false)
			// return block
		},
	},

	// Tests:
	// equal { x: 0 defer { x:: 1 } x } 0
	// equal { fn { } { x: 0 defer { x:: 1 } x } } 0
	// equal { fn { } { x: 0 defer { x:: 1 } } x } 1
	"defer": {
		Argsn: 1,
		Doc:   "Registers a block of code to be executed when the current function exits or the program terminates.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch block := arg0.(type) {
			case env.Block:
				// Add the block to the deferred blocks list
				ps.DeferBlocks = append(ps.DeferBlocks, block)
				return env.Void{}
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType}, "defer")
			}
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
	RegisterBuiltins2(builtins, ps, "base")
	RegisterBuiltins2(builtins_boolean, ps, "base")
	RegisterBuiltins2(builtins_numbers, ps, "base")
	RegisterBuiltins2(builtins_time, ps, "base")
	RegisterBuiltins2(builtins_string, ps, "base")
	RegisterBuiltins2(builtins_collection, ps, "base")
	RegisterBuiltins2(builtins_iteration, ps, "base")
	RegisterBuiltins2(builtins_contexts, ps, "base")
	RegisterBuiltins2(builtins_functions, ps, "base")
	RegisterBuiltins2(Builtins_failure, ps, "failure")
	RegisterBuiltins2(Builtins_table, ps, "table")
	RegisterBuiltins2(Builtins_vector, ps, "vector")
	RegisterBuiltins2(Builtins_io, ps, "io")
	RegisterBuiltins2(Builtins_regexp, ps, "regexp")
	RegisterBuiltins2(Builtins_validation, ps, "validation")
	RegisterBuiltins2(Builtins_conversion, ps, "conversion")
	RegisterBuiltins2(Builtins_web, ps, "web")
	RegisterBuiltins2(Builtins_markdown, ps, "markdown")
	RegisterBuiltins2(Builtins_sxml, ps, "sxml")
	RegisterBuiltins2(Builtins_html, ps, "html")
	RegisterBuiltins2(Builtins_json, ps, "json")
	RegisterBuiltins2(Builtins_bson, ps, "bson")
	RegisterBuiltins2(Builtins_stackless, ps, "stackless")
	RegisterBuiltins2(Builtins_eyr, ps, "eyr")
	RegisterBuiltins2(Builtins_goroutines, ps, "goroutines")
	RegisterBuiltins2(Builtins_http, ps, "http")
	RegisterBuiltins2(Builtins_sqlite, ps, "sqlite")
	RegisterBuiltins2(Builtins_psql, ps, "psql")
	RegisterBuiltins2(Builtins_mysql, ps, "mysql")
	RegisterBuiltins2(Builtins_email, ps, "email")
	RegisterBuiltins2(Builtins_structures, ps, "structs")
	RegisterBuiltins2(Builtins_smtpd, ps, "smtpd")
	RegisterBuiltins2(Builtins_mail, ps, "mail")
	RegisterBuiltins2(Builtins_ssh, ps, "ssh")
	RegisterBuiltins2(Builtins_bcrypt, ps, "bcrypt")
	RegisterBuiltins2(Builtins_console, ps, "console")
	RegisterBuiltinsInContext(Builtins_crypto, ps, "crypto")
	RegisterBuiltinsInContext(Builtins_math, ps, "math")
	RegisterBuiltinsInContext(Builtins_os, ps, "os")
	RegisterBuiltinsInContext(Builtins_pipes, ps, "pipes")
	RegisterBuiltinsInContext(Builtins_term, ps, "term")
	RegisterBuiltinsInContext(Builtins_telegrambot, ps, "telegram")
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
		bu := env.NewBuiltin(v.Fn, v.Argsn, v.AcceptFailure, v.Pure, v.Doc+" ("+name+")")
		registerBuiltin(ps, k, *bu)
	}
}

func RegisterBuiltinsInContext(builtins map[string]*env.Builtin, ps *env.ProgramState, name string) *env.RyeCtx {
	BuiltinNames[name] = len(builtins)

	ctx := ps.Ctx
	ps.Ctx = env.NewEnv(ps.Ctx) // make new context with no parent

	for k, v := range builtins {
		bu := env.NewBuiltin(v.Fn, v.Argsn, v.AcceptFailure, v.Pure, v.Doc+" ("+k+")")
		registerBuiltin(ps, k, *bu)
	}
	newctx := ps.Ctx
	ps.Ctx = ctx

	wordIdx := ps.Idx.IndexWord(name)
	ps.Ctx.Mod(wordIdx, *newctx)

	return newctx
}

func RegisterBuiltinsInSubContext(builtins map[string]*env.Builtin, ps *env.ProgramState, parent *env.RyeCtx, name string) *env.RyeCtx {
	BuiltinNames[name] = len(builtins)

	ctx := ps.Ctx
	ps.Ctx = env.NewEnv(parent) // make new context with no parent

	for k, v := range builtins {
		bu := env.NewBuiltin(v.Fn, v.Argsn, v.AcceptFailure, v.Pure, v.Doc+" ("+k+")")
		registerBuiltin(ps, k, *bu)
	}
	newctx := ps.Ctx
	ps.Ctx = ctx

	wordIdx := ps.Idx.IndexWord(name)
	ps.Ctx.Mod(wordIdx, *newctx)

	return newctx
}

func registerBuiltin(ps *env.ProgramState, word string, builtin env.Builtin) {
	// indexWord
	// TODO -- this with string separator is a temporary way of how we define generic builtins
	// in future a map will probably not be a map but an array and builtin will also support the Kind value
	builtin.Doc = builtin.Doc + " (" + word + ")"
	idxk := 0
	if word != "_//" && strings.Index(word, "//") > 0 {
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
