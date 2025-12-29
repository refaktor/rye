package evaldo

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"

	"github.com/refaktor/rye/env"
	"github.com/refaktor/rye/term"

	"github.com/refaktor/rye/loader"
	// JM 20230825	"github.com/refaktor/rye/term"
	"strconv"
	"strings"
	"time"

	"github.com/refaktor/rye/util"

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
	return env.NewError(msg + " In builtin `" + fn + "`")
}

// docNameRegex matches a simple name in parentheses at the end of a docstring
// Supports Rye function names with: letters, digits, underscore, backslash, dash, plus, asterisk, question mark, exclamation
var docNameRegex = regexp.MustCompile(`\(([a-zA-Z0-9_\\/?!*+<>=.-]+)\)\s*$`)

// ExtractFunctionName extracts the function name from a docstring that ends with "(name)"
// If the pattern is found, returns the name; otherwise returns the full docstring
func ExtractFunctionName(doc string) (name string, found bool) {
	matches := docNameRegex.FindStringSubmatch(doc)
	if len(matches) >= 2 {
		return matches[1], true
	}
	return doc, false
}

// FormatBuiltinReference formats a builtin reference for error display
// If a name can be extracted from the docstring, it returns the emphasized name
// Otherwise returns the full docstring
func FormatBuiltinReference(doc string) string {
	name, found := ExtractFunctionName(doc)
	if found {
		return "\x1b[41m\x1b[30m " + name + " \x1b[0m\x1b[1;31m" // Red background, black text, then reset and restore bold red
	}
	return doc
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
		// Check if in bounds before accessing env.NativeTypes
		if tt > 0 && int(tt-1) < len(env.NativeTypes) {
			types += env.NativeTypes[tt-1]
		} else {
			types += "UNKNOWN_TYPE"
		}
	}
	// Format function name with red badge style
	formattedFn := "\x1b[41m\x1b[30m " + fn + " \x1b[0m"
	return "builtin " + formattedFn + " requires argument " + strconv.Itoa(N) + " to be: " + types + "."
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
	case env.Word: // todo .. make Error type .. make error construction micro dialect, return the error wrapping error that caused it
		return env.NewError5(val, 0, "", er, nil)
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
	case env.Time:
		switch vB := arg1.(type) {
		case env.Time:
			return vA.Value.After(vB.Value)
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
	case env.String:
		switch vB := arg1.(type) {
		case env.String:
			return vA.Value < vB.Value
		default:
			return false
		}
	case env.Time:
		switch vB := arg1.(type) {
		case env.Time:
			return vA.Value.Before(vB.Value)
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
	case env.Time:
		switch vB := arg1.(type) {
		case env.Time:
			return vA.Value.Before(vB.Value)
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
			case env.Boolean:
				return v1
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
				return env.NewError("Key is missing in dict")
			default:
				ps.FailureFlag = true
				return env.NewError("Unhandeled value, type: " + reflect.TypeOf(v1).String())
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
			if idx < int64(len(s1.Rows)) {
				v := s1.Rows[idx]
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
			columnNames := s1.Uplink.GetColumnNames()
			for i := 0; i < len(columnNames); i++ {
				if columnNames[i] == s2.Value {
					index = i
				}
			}
			v := s1.Values[index]
			if true {
				return env.ToRyeValue(v)
			} else {
				return makeError(ps, "Index larger than table row length")
			}
		case env.Integer:
			idx := s2.Value
			if posMode {
				idx--
			}
			if idx < int64(len(s1.Values)) {
				v := s1.Values[idx]
				return env.ToRyeValue(v)
			} else {
				return makeError(ps, "Index larger than table row length")
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
	// Add bounds checking
	if i < 0 || i >= len(s.data) || j < 0 || j >= len(s.data) {
		// If indices are out of bounds, maintain stable sort behavior
		return i < j
	}

	// Defer error recovery to prevent panic during sorting
	defer func() {
		if r := recover(); r != nil {
			// If comparison function panics, default to index-based comparison
			// This prevents the entire sort from failing
		}
	}()

	// Call the user's comparison function with error handling
	CallFunctionArgs2(s.fn, s.ps, s.data[i], s.data[j], nil)

	// Check for errors in the program state
	if s.ps.ErrorFlag || s.ps.FailureFlag {
		// Reset error flags to prevent cascading failures
		s.ps.ErrorFlag = false
		s.ps.FailureFlag = false
		// Default to index-based comparison on error
		return i < j
	}

	// TODO -- probably we should throw error if not boolean result #strictness
	// fmt.Println(s.ps.Res.Inspect(*s.ps.Idx))
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
	// Add bounds checking
	if i < 0 || i >= len(s.data) || j < 0 || j >= len(s.data) {
		// If indices are out of bounds, maintain stable sort behavior
		return i < j
	}

	// Defer error recovery to prevent panic during sorting
	defer func() {
		if r := recover(); r != nil {
			// If comparison function panics, default to index-based comparison
			// This prevents the entire sort from failing
		}
	}()

	// Safely convert values to Rye objects with error handling
	var val1, val2 env.Object

	// Convert first value with error handling
	func() {
		defer func() {
			if r := recover(); r != nil {
				// If conversion panics, create a safe fallback value
				val1 = *env.NewString(fmt.Sprintf("%v", s.data[i]))
			}
		}()
		val1 = env.ToRyeValue(s.data[i])
	}()

	// Convert second value with error handling
	func() {
		defer func() {
			if r := recover(); r != nil {
				// If conversion panics, create a safe fallback value
				val2 = *env.NewString(fmt.Sprintf("%v", s.data[j]))
			}
		}()
		val2 = env.ToRyeValue(s.data[j])
	}()

	// Call the user's comparison function with error handling
	CallFunctionArgs2(s.fn, s.ps, val1, val2, nil)

	// Check for errors in the program state
	if s.ps.ErrorFlag || s.ps.FailureFlag {
		// Reset error flags to prevent cascading failures
		s.ps.ErrorFlag = false
		s.ps.FailureFlag = false
		// Default to index-based comparison on error
		return i < j
	}

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
		// fmt.Println(block)
		panic("Not block and not error in import builtin.") // TODO -- Think how best to handle this
		// return env.Void{}
	}
}

var ShowResults bool

var builtins = map[string]*env.Builtin{
	/* "compile-fast": {
		Argsn: 1,
		Doc:   "Takes a block of code and compiles it to function pointers for the fast evaluator.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch bloc := arg0.(type) {
			case env.Block:
				ser := ps.Ser
				ps.Ser = bloc.Series
				program := Rye0_CompileBlock(ps)
				ps.Ser = ser
				return *env.NewNative(ps.Idx, program, "compiled-program")
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType}, "compile-fast")
			}
		},
	},

	"execute-fast": {
		Argsn: 1,
		Doc:   "Takes a compiled program and executes it using the fast evaluator.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch prog := arg0.(type) {
			case env.Native:
				if ps.Idx.GetWord(prog.Kind.Index) == "compiled-program" {
					// Create a VM
					vm := NewRye0VM(ps.Ctx, ps.PCtx, ps.Gen, ps.Idx)

					// Execute the program
					program := prog.Value.(*Program)
					result, err := vm.Execute(program)
					if err != nil {
						ps.ErrorFlag = true
						return env.NewError(err.Error())
					}
					return result
				}
				return MakeBuiltinError(ps, "Expected a compiled program", "execute-fast")
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "execute-fast")
			}
		},
	},

	"compile-fast\\debug": {
		Argsn: 1,
		Doc:   "Takes a block of code and compiles it to function pointers for the fast evaluator, printing debug info.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch bloc := arg0.(type) {
			case env.Block:
				// Save original position
				origPos := bloc.Series.GetPos()

				// Print header
				fmt.Println("\n=== Compiling to function pointers ===")
				fmt.Println("Block:", bloc.Print(*ps.Idx))

				// Iterate through each object in the block and print what's being compiled
				for i := 0; i < bloc.Series.Len(); i++ {
					obj := bloc.Series.Get(i)
					fmt.Printf("\nCompiling object %d: %s (type: %s)\n", i, obj.Print(*ps.Idx), NameOfRyeType(obj.Type()))

					// Create a temporary series with just this object
					tempSeries := *env.NewTSeries([]env.Object{obj})
					tempPS := env.NewProgramState(tempSeries, ps.Idx)
					tempPS.Ctx = ps.Ctx
					tempPS.Dialect = ps.Dialect

					// Compile just this object and print info about the instruction
					instr := Rye0_CompileExpression(tempPS)
					fmt.Printf("  -> Created function pointer: %T\n", instr)

					// Print what the instruction does (simplified description)
					switch obj.Type() {
					case env.IntegerType, env.DecimalType, env.StringType:
						fmt.Print("  -> Action: Push literal value onto stack\n")
					case env.WordType:
						fmt.Print("  -> Action: Evaluate word and push result onto stack\n")
					case env.BlockType:
						fmt.Print("  -> Action: Create block object and push onto stack\n")
					case env.SetwordType:
						fmt.Print("  -> Action: Set word to value\n")
					case env.ModwordType:
						fmt.Print("  -> Action: Modify word with value\n")
					default:
						fmt.Print("  -> Action: Handle object based on type\n")
					}
				}

				// Reset position
				bloc.Series.SetPos(origPos)

				// Compile the whole block
				ser := ps.Ser
				ps.Ser = bloc.Series
				program := Rye0_CompileBlock(ps)
				ps.Ser = ser

				fmt.Printf("\nCompiled %d instructions into program\n", len(program.code))
				fmt.Print("=== End compilation ===\n\n")

				return *env.NewNative(ps.Idx, program, "compiled-program")
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType}, "compile-fast/debug")
			}
		},
	},

	// Tests:
	// ; equal  { do-fast { 123 + 123 } } 246
	// ; error  { do-fast { 123 + } }
	// ; equal  { do-fast { _+ _+ 12 23 34 } } 69
	// ; equal  { do-fast { 12 * 23 |+ 34 } } 310
	// ; equal  { do-fast { ( 12 * 23 ) + 34 } } 310
	// ; equal  { do-fast { 12 * 23 | + 34 } } 310
	// ; equal  { do-fast { 12 * 23 :a + 34 } } 310
	"do-fast": {
		Argsn: 1,
		Doc:   "Takes a block of code and does (runs) it using the fast evaluator.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch bloc := arg0.(type) {
			case env.Block:
				ser := ps.Ser
				ps.Ser = bloc.Series
				ps.Dialect = env.Rye2Dialect
				Rye0_FastEvalBlock(ps)
				ps.Ser = ser
				return ps.Res
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType}, "do-fast")
			}
		},
	}, */

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

	//
	// ##### Other ##### "..."
	//
	// Tests:
	// equal   { var 'x 123 , change! 234 'x , x } 234
	// equal   { a:: 123 change! 333 'a a } 333
	// equal   { a:: 123 change! 124 'a } true
	// equal   { a:: 123 change! 123 'a } false
	// Args:
	// * value: New value to assign to the word
	// * word: Word whose value should be changed
	// Returns:
	// * Boolean true if the value changed, false if the new value is the same as the old value
	"change!": { // ***
		Argsn: 2,
		Doc:   "Searches for a word and changes it's value in-place. Only works on variables declared with var. If value changes returns true otherwise false",
		Pure:  false,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch arg := arg1.(type) {
			case env.Word:
				val, found, ctx := ps.Ctx.Get2(arg.Index)
				if found {
					// Attempt to modify the word
					if ok := ctx.Mod(arg.Index, arg0); !ok {
						ps.FailureFlag = true
						return env.NewError("Cannot modify constant '" + ps.Idx.GetWord(arg.Index) + "', use 'var' to declare it as a variable")
					}

					if arg0.GetKind() == val.GetKind() && arg0.Inspect(*ps.Idx) == val.Inspect(*ps.Idx) {
						return *env.NewBoolean(false)
					} else {
						// Trigger observers if the variable was successfully modified and the value actually changed
						// Use the correct context (ctx) where the variable was actually found and modified
						if ctx.IsVariable(arg.Index) {
							if ryeCtx, ok := ctx.(*env.RyeCtx); ok {
								TriggerObservers(ps, ryeCtx, arg.Index, val, arg0)
							}
						}
						return *env.NewBoolean(true)
					}
				}
				return MakeBuiltinError(ps, "Word not found in context.", "change!")
			default:
				return MakeArgError(ps, 1, []env.Type{env.WordType}, "change!")
			}
		},
	},

	// Tests:
	// equal   { var 'x 123 , modify! 234 'x , x } 234
	// equal   { a:: 123 modify! 333 'a a } 333
	// equal   { a:: 123 modify! 124 'a } true
	// equal   { a:: 123 modify! 123 'a } false
	// Args:
	// * value: New value to assign to the word
	// * word: Word whose value should be changed
	// Returns:
	// * Boolean true if the value changed, false if the new value is the same as the old value, will replace change!
	"modify!": { // ***
		Argsn: 2,
		Doc:   "Searches for a word and modifies	 it's value in-place. Only works on variables declared with var. If value changes returns true otherwise false",
		Pure:  false,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch arg := arg1.(type) {
			case env.Word:
				val, found, ctx := ps.Ctx.Get2(arg.Index)
				if found {
					// Attempt to modify the word
					if ok := ctx.Mod(arg.Index, arg0); !ok {
						ps.FailureFlag = true
						return env.NewError("Cannot modify constant '" + ps.Idx.GetWord(arg.Index) + "', use 'var' to declare it as a variable")
					}

					if arg0.GetKind() == val.GetKind() && arg0.Inspect(*ps.Idx) == val.Inspect(*ps.Idx) {
						return *env.NewBoolean(false)
					} else {
						// Trigger observers if the variable was successfully modified and the value actually changed
						// Use the correct context (ctx) where the variable was actually found and modified
						if ctx.IsVariable(arg.Index) {
							if ryeCtx, ok := ctx.(*env.RyeCtx); ok {
								TriggerObservers(ps, ryeCtx, arg.Index, val, arg0)
							}
						}
						return *env.NewBoolean(true)
					}
				}
				return MakeBuiltinError(ps, "Word not found in context.", "change!")
			default:
				return MakeArgError(ps, 1, []env.Type{env.WordType}, "change!")
			}
		},
	},

	// Tests:
	// equal   { set { 123 234 } { a b }  b } 234
	// Args:
	// * values: Value or block of values to assign to the word(s)
	// * words: Word or block of words to be set
	// Returns:
	// * The value or block of values that was assigned
	"set": { // ***
		Argsn: 2,
		Doc:   "Set word to value or words by deconstructing a block. Only works on variables declared with var.",
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
								return MakeBuiltinError(ps, "More words than values.", "set")
							}
							val := vals.Series.S[i]

							// Get old value for observer notification
							oldValue, exists := ps.Ctx.GetCurrent(word.Index)

							// if it exists then we set it to word from words
							if ok := ps.Ctx.Mod(word.Index, val); !ok {
								ps.FailureFlag = true
								return env.NewError("Cannot modify constant '" + ps.Idx.GetWord(word.Index) + "', use 'var' to declare it as a variable")
							} else {
								// Trigger observers if the variable was successfully modified
								if exists && ps.Ctx.IsVariable(word.Index) {
									// Only trigger if the value actually changed
									if oldValue == nil || !oldValue.Equal(val) {
										TriggerObservers(ps, ps.Ctx, word.Index, oldValue, val)
									}
								}
							}
						default:
							fmt.Println(word)
							return MakeBuiltinError(ps, "Only words in words block", "set")
						}
					}
					return arg0
				default:
					return MakeArgError(ps, 2, []env.Type{env.BlockType}, "set")
				}
			case env.Word:
				// Get old value and context for observer notification
				oldValue, found, ctx := ps.Ctx.Get2(words.Index)

				if ok := ctx.Mod(words.Index, arg0); !ok {
					ps.FailureFlag = true
					return env.NewError("Cannot modify constant '" + ps.Idx.GetWord(words.Index) + "', use 'var' to declare it as a variable")
				} else {
					// Trigger observers if the variable was successfully modified
					// Use the correct context (ctx) where the variable was actually found and modified
					if found && ctx.IsVariable(words.Index) {
						// Only trigger if the value actually changed
						if oldValue == nil || !oldValue.Equal(arg0) {
							if ryeCtx, ok := ctx.(*env.RyeCtx); ok {
								TriggerObservers(ps, ryeCtx, words.Index, oldValue, arg0)
							}
						}
					}
				}
				return arg0
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType}, "set")
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
	// ; TODO equal   { person: kind 'person { name: "" age: 0 } person << dict { "name" "John" "age" 30 } |type? } 'ctx
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
					return MakeArgError(ps, 2, []env.Type{env.DictType, env.ContextType}, "_<<")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.KindType}, "_<<")
			}
		},
	},

	// Tests:
	// equal { 5 _| } 5
	// equal { "hello" _| } "hello"
	// equal { true _| } true
	// equal { { 1 2 3 } _| } { 1 2 3 }
	// Args:
	// * value: Any value to be passed through unchanged
	// Returns:
	// * the original value (used in pipeline operations for explicit pass-through)
	"_|": {
		Argsn: 1,
		Doc:   "Pipeline operator that passes the value through unchanged (used with 'not' and other operations).",
		Pure:  true,
		Fn: func(env1 *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return arg0
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
	// ; equal  { save\current\secure |type? } 'integer
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
				env1.ErrorFlag = true
				return MakeArgError(env1, 1, []env.Type{env.ContextType, env.PersistentContextType}, "doc\\of?")
			}

		},
	},
	// Tests:
	// equal   { is-ref ref { 1 2 3 } } true
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
				return MakeArgError(ps, 1, []env.Type{env.TableType, env.DictType, env.ListType, env.BlockType, env.StringType, env.NativeType, env.ContextType}, "deref")
			}
		},
	},

	// Tests:
	// equal   { is-ref deref ref { 1 2 3 } } false
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
				return MakeArgError(ps, 1, []env.Type{env.TableType, env.DictType, env.ListType, env.BlockType, env.StringType, env.NativeType, env.ContextType}, "deref")
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
				return env.NewBoolean(true)
			} else {
				return env.NewBoolean(false)
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
		Pure:  true,
		Doc:   "Constructs a Dict from the Block of key and value pairs.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch bloc := arg0.(type) {
			case env.Block:
				// Validate that all keys are strings, tagwords, words, or setwords
				for i := 0; i < bloc.Series.Len(); i += 2 {
					if i >= bloc.Series.Len() {
						break
					}
					key := bloc.Series.Get(i)
					switch key.(type) {
					case env.String, env.Tagword, env.Word, env.Setword:
						// Valid key types
					default:
						// Invalid key type - return failure
						return MakeBuiltinError(ps, fmt.Sprintf("Dict keys must be strings, tagwords, words, or setwords, but got %s at position %d", key.Inspect(*ps.Idx), i), "dict")
					}
				}
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
		Pure:  true,
		Doc:   "Constructs a List from the Block of values.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch bloc := arg0.(type) {
			case env.Block:
				return env.NewListFromSeries(bloc.Series)
			}
			return nil
		},
	},

	/* JM 20230825 */
	"display-selection": {
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
	"display-input": {
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
	},
	"display-date-input": {
		Argsn: 2,
		Doc:   "Interactive date input in format YYYY-MM-DD. Use arrow keys to navigate between year/month/day fields and increment/decrement values, or type digits. Returns date string.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			term.SaveCurPos()
			switch initialDate := arg0.(type) {
			case env.String:
				obj, esc := term.DisplayDateInput(initialDate.Value, int(arg1.(env.Integer).Value))
				if !esc {
					return obj
				}
			}
			return nil
		},
	}, /* */

	// Tests:
	// ; import file://test.rye  ; imports and executes test.rye
	// Args:
	// * uri: URI of the file to import and execute
	// Returns:
	// * result of executing the imported file
	"file-uri//Import": { // **
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

	// Tests:
	// ; import\live file://test.rye  ; imports, executes, and watches test.rye for changes
	// Args:
	// * uri: URI of the file to import, execute, and watch for changes
	// Returns:
	// * result of executing the imported file
	"file-uri//Import\\live": { // **
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
	// Args:
	// * source: String containing Rye code or URI of file to load
	// Returns:
	// * Block containing the parsed Rye values
	"load": { // **
		Argsn: 1,
		Doc:   "Loads a string into Rye values.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.String:
				block := loader.LoadStringNEW(s1.Value, false, ps)
				//ps = env.AddToProgramState(ps, block.Series, genv)
				return block
			// Deprecated ... should only load strings, like reader / Reader functions
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

	// Tests:
	// ; equal  { load " 1 2 3 " |third } 3
	// ; equal  { load "{ 1 2 3 }" |first |third } 3
	// Args:
	// * source: String containing Rye code or URI of file to load
	// Returns:
	// * Block containing the parsed Rye values
	"file-uri//Load": { // **
		Argsn: 1,
		Doc:   "Loads a string into Rye values.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
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

	// Tests:
	// ; load\mod file://modifiable.rye  ; loads file with word modification allowed
	// Args:
	// * source: String containing Rye code or URI of file to load with modification allowed
	// Returns:
	// * Block containing the parsed Rye values
	"load\\mod": { // **
		Argsn: 1,
		Doc:   "Loads a string into Rye values. During load it allows modification of words.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.String:
				block, _ := loader.LoadStringNoPEG(s1.Value, false)
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

	// Tests:
	// ; load\live file://watched.rye  ; loads and watches file for changes
	// Args:
	// * source: String containing Rye code or URI of file to load with modification allowed and file watching
	// Returns:
	// * Block containing the parsed Rye values
	"load\\live": { // **
		Argsn: 1,
		Doc:   "Loads a string into Rye values. During load it allows modification of words.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.String:
				block, _ := loader.LoadStringNoPEG(s1.Value, false)
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

	// Tests:
	// ; load\sig "signed-code"  ; loads only if signature is valid
	// Args:
	// * source: String containing signed Rye code to verify and load
	// Returns:
	// * Block containing the parsed Rye values if signature is valid
	"load\\sig": {
		Argsn: 1,
		Doc:   "Checks the signature, if OK then loads a string into Rye values.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.String:
				block, _ := loader.LoadStringNoPEG(s1.Value, true)
				//ps = env.AddToProgramState(ps, block.Series, genv)
				return block
			default:
				ps.FailureFlag = true
				return env.NewError("Must be string or file TODO")
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
	// Args:
	// * block: Block of code to execute
	// Returns:
	// * result of executing the block
	"do": {
		Argsn: 1,
		Doc:   "Takes a block of code and does (runs) it.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch bloc := arg0.(type) {
			case env.Block:
				ser := ps.Ser
				ps.Ser = bloc.Series
				ps.BlockFile = bloc.FileName
				ps.BlockLine = bloc.Line
				EvalBlock(ps)
				MaybeDisplayFailureOrError(ps, ps.Idx, "for\\kv")
				ps.Ser = ser
				return ps.Res
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType}, "do")
			}
		},
	},

	// Tests:
	// equal  { with 100 { + 11 } } 111
	// equal  { with 100 { + 11 , * 3 } } 300
	// Args:
	// * value: Value to inject into the block's execution context
	// * block: Block of code to execute with the injected value
	// Returns:
	// * result of executing the block with the injected value
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
				MaybeDisplayFailureOrError(ps, ps.Idx, "with")
				ps.Ser = ser
				return ps.Res
			default:
				return MakeArgError(ps, 2, []env.Type{env.BlockType}, "with")
			}
		},
	},

	// Tests:
	// equal  { c: context { x: 100 } do\inside c { x * 9.99 } } 999.0
	// equal  { c: context { x:: 100 } do\inside c { inc! 'x } } 101
	// equal  { c: context { var 'x 100 } do\inside c { x:: 200 } c/x } 200
	// equal  { c: context { x:: 100 } do\inside c { x:: 200 , x } } 200
	// Args:
	// * context: Context in which to execute the block
	// * block: Block of code to execute within the specified context
	// Returns:
	// * result of executing the block within the given context
	"do\\inside": { // **
		Argsn: 2,
		Doc:   "Takes a Context and a Block. It Does a block inside a given Context.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch ctx := arg0.(type) {
			case env.RyeCtx:
				switch bloc := arg1.(type) {
				case env.Block:
					ser := ps.Ser
					ps.Ser = bloc.Series
					EvalBlockInCtxInj(ps, &ctx, nil, false)
					MaybeDisplayFailureOrError(ps, ps.Idx, "do\\inside")
					ps.Ser = ser
					return ps.Res
				default:
					ps.ErrorFlag = true
					return MakeArgError(ps, 2, []env.Type{env.BlockType}, "do\\inside")
				}
			case PersistentCtx:
				switch bloc := arg1.(type) {
				case env.Block:
					ser := ps.Ser
					ps.Ser = bloc.Series
					// Use a special evaluation function for PersistentCtx
					EvalBlockInPersistentCtx(ps, &ctx)
					MaybeDisplayFailureOrError(ps, ps.Idx, "do\\inside")
					ps.Ser = ser
					return ps.Res
				default:
					ps.ErrorFlag = true
					return MakeArgError(ps, 2, []env.Type{env.BlockType}, "do\\inside")
				}
			default:
				ps.ErrorFlag = true
				return MakeArgError(ps, 1, []env.Type{env.ContextType, env.PersistentContextType}, "do\\inside")
			}

		},
	},

	// Tests:
	// equal  { c: context { x: 100 } do\in c { x * 9.99 } } 999.0
	// equal  { c: context { x:: 100 } do\in c { inc! 'x } } 101
	// equal  { c: context { x: 100 } do\in c { x:: 200 , x } } 200
	// equal  { c: context { x: 100 } do\in c { x:: 200 } c/x } 100
	// Args:
	// * context: Context to use as parent context during execution
	// * block: Block of code to execute in current context with the specified parent context
	// Returns:
	// * result of executing the block with the modified parent context
	"do\\in": { // **
		Argsn: 2,
		Doc:   "Takes a Context and a Block. It Does a block in current context but with parent a given Context.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch ctx := arg0.(type) {
			case env.RyeCtx:
				switch bloc := arg1.(type) {
				case env.Block:
					ser := ps.Ser
					ps.Ser = bloc.Series
					// we take ARG CTX
					tempCtx := &ctx
					for {
						// fmt.Println("UP ->")
						// do we change parents globally (and we have to fix them back) or just in local copy of a value? check
						// If it's parent is the same as current context is
						// we set it's parent to our parent, skip our context
						if tempCtx.Parent == ps.Ctx {
							// fmt.Println("ISTI PARENT")
							// temp = ctx.Parent
							tempCtx.Parent = ps.Ctx.Parent
							break
						}
						if tempCtx.Parent != nil {
							// fmt.Println("še en parent")
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
					MaybeDisplayFailureOrError(ps, ps.Idx, "do\\in")
					// if temp != nil {
					ps.Ctx.Parent = temp
					// ctx.Parent = temp
					// }
					ps.Ser = ser
					return ps.Res
				default:
					ps.ErrorFlag = true
					return MakeArgError(ps, 2, []env.Type{env.BlockType}, "do\\in")
				}
			default:
				ps.ErrorFlag = true
				return MakeArgError(ps, 1, []env.Type{env.ContextType}, "do\\in")
			}

		},
	},

	// Tests:
	// equal  { c: context { x: 100 } do\in c { x * 9.99 } } 999.0
	// equal  { c: context { x:: 100 } do\in c { inc! 'x } } 101
	// equal  { c: context { x: 100 } do\in c { x:: 200 , x } } 200
	// equal  { c: context { x: 100 } do\in c { x:: 200 } c/x } 100
	// Args:
	// * context: Context to use as parent context during execution
	// * block: Block of code to execute in current context with the specified parent context
	// Returns:
	// * result of executing the block with the modified parent context
	"do\\inx": { // **
		Argsn: 2,
		Doc:   "Takes a Context and a Block. It Does a block in current context but with parent a given Context.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch ctx := arg0.(type) {
			case env.RyeCtx:
				switch bloc := arg1.(type) {
				case env.Block:
					ser := ps.Ser
					ps.Ser = bloc.Series
					// we take ARG CTX
					// tempCtx := &ctx
					// fmt.Println("UP ->")
					// do we change parents globally (and we have to fix them back) or just in local copy of a value? check
					// If it's parent is the same as current context is
					// we set it's parent to our parent, skip our context
					ctx.Parent = ps.Ctx.Parent

					//var temp *env.RyeCtx
					// set argument's parent context to current parent context
					// }
					// set argument context as parent
					temp := ps.Ctx.Parent
					ps.Ctx.Parent = &ctx
					EvalBlock(ps)
					MaybeDisplayFailureOrError(ps, ps.Idx, "for\\kv")
					// if temp != nil {
					ps.Ctx.Parent = temp
					// ctx.Parent = temp
					// }
					ps.Ser = ser
					return ps.Res
				default:
					ps.ErrorFlag = true
					return MakeArgError(ps, 2, []env.Type{env.BlockType}, "do\\in")
				}
			default:
				ps.ErrorFlag = true
				return MakeArgError(ps, 1, []env.Type{env.ContextType}, "do\\in")
			}

		},
	},

	// COLLECTION AND RETURNING  ... NOT DETERMINED IF WILL BE INCLUDED YET EXPERIMENTAL

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

	// END COLLECT EXPERIMENTAL

	"lk": {
		Argsn: 0,
		Doc:   "Lists available kinds",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			fmt.Println(ps.Gen.PreviewKinds(*ps.Idx, ""))
			return env.Void{}
		},
	},

	"lk\\": {
		Argsn: 1,
		Doc:   "Lists available kinds with string filter",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.String:
				fmt.Println(ps.Gen.PreviewKinds(*ps.Idx, s1.Value))
				return env.Void{}
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "lk\\")
			}
		},
	},

	"lg": {
		Argsn: 1,
		Doc:   "Lists generic words related to specific kind, or lists all matching methods across all kinds if given a string filter",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.String:
				// If argument is a string, show all matching methods across all kinds
				fmt.Println(ps.Gen.PreviewAllMatchingMethods(*ps.Idx, s1.Value))
				return env.Void{}
			case env.Word:
				fmt.Println(ps.Gen.PreviewMethods(*ps.Idx, s1.Index, ""))
				return env.Void{}
			default:
				kindIdx := arg0.GetKind()
				fmt.Println(ps.Gen.PreviewMethods(*ps.Idx, kindIdx, ""))
				return arg0
			}
		},
	},

	"lg\\": {
		Argsn: 2,
		Doc:   "Lists generic words related to specific kind with string filter",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch filter := arg1.(type) {
			case env.String:
				switch s1 := arg0.(type) {
				case env.Word:
					fmt.Println(ps.Gen.PreviewMethods(*ps.Idx, s1.Index, filter.Value))
					return env.Void{}
				default:
					kindIdx := arg0.GetKind()
					fmt.Println(ps.Gen.PreviewMethods(*ps.Idx, kindIdx, filter.Value))
					return env.Void{}
				}
			default:
				return MakeArgError(ps, 2, []env.Type{env.StringType}, "lg\\")
			}
		},
	},

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
				ps.SkipFlag = false // TODO --- make true
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
	// equal { scmd `echo "hello"` } 0
	// equal { scmd `exit 1` } 1
	// equal { scmd `exit 42` } 42
	"scmd": {
		Argsn: 1,
		Doc:   "Execute a shell command and return its exit status code.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s0 := arg0.(type) {
			case env.String:

				r := exec.Command("/bin/bash", "-c", s0.Value) //nolint: gosec
				// stdout, stderr := r.Output()
				r.Stdout = os.Stdout
				r.Stderr = os.Stderr

				err := r.Run()
				if err != nil {
					// Check if this is an exit error with a status code
					if exitError, ok := err.(*exec.ExitError); ok {
						// Return the actual exit code
						return *env.NewInteger(int64(exitError.ExitCode()))
					}
					// For other errors (like command not found), print error and return -1
					fmt.Println(err)
					return *env.NewInteger(-1)
				}
				// Command succeeded, return 0
				return *env.NewInteger(0)
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "scmd")
			}
		},
	},

	// Tests:
	// equal { scmd `echo "hello"` } "hello"
	"scmd\\capture": {
		Argsn: 1,
		Doc:   "Execute a shell command and capture the output, return it as string",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s0 := arg0.(type) {
			case env.String:
				r := exec.Command("/bin/bash", "-c", s0.Value) //nolint: gosec
				var outb, errb bytes.Buffer
				r.Stdout = &outb
				r.Stderr = &errb

				err := r.Run()
				if err != nil {
					// If command fails, return the error output as part of the result
					if errb.Len() > 0 {
						ps.FailureFlag = true
						return env.NewError("Command failed: " + errb.String())
					}
					// For exit errors, still return stdout if available
					if _, ok := err.(*exec.ExitError); ok && outb.Len() > 0 {
						return *env.NewString(outb.String())
					}
					ps.FailureFlag = true
					return env.NewError("Command failed: " + err.Error())
				}

				// Return the captured stdout as string
				return *env.NewString(outb.String())
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "scmd\\capture")
			}
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
	// equal { x:: 123 defer { x:: 345 } x } 123
	// stdout { ff:: fn { } { var 'x 123 defer { print 234 } x } , ff } "234\n"
	// equal { ff:: fn { } { x:: 123 defer { x:: 234 } x + 111 } , ff } 234 ; the result of defer expression is returned TODO, change this
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

	// Tests:
	// equal { x:: 0 defer\ 42 { + 1 } x } 0
	// stdout { ff:: fn { } { defer\ "hello" { .print } "done" } , ff } "hello\n"
	// Args:
	// * value: Value to inject into the deferred block
	// * block: Block to execute with the injected value when function exits
	// Returns:
	// * Void value
	"defer\\": {
		Argsn: 2,
		Doc:   "Registers a block of code with an injected value to be executed when the current function exits or the program terminates. Works like 'with' but deferred.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch block := arg1.(type) {
			case env.Block:
				// Create a new block that contains "arg0 .with arg1" equivalent
				// We need to create a block that will inject arg0 into block when executed

				// Create the objects for the deferred operation:
				// arg0 (the value) .with block
				withIdx := ps.Idx.IndexWord("with")
				withWord := *env.NewOpword(withIdx, 0) // .with as opword

				// Create a new series with: value, .with, block
				objects := make([]env.Object, 3)
				objects[0] = arg0     // the value to inject
				objects[1] = withWord // .with opword
				objects[2] = block    // the block to execute

				series := env.NewTSeries(objects)
				deferredBlock := env.NewBlock(*series)

				// Add the constructed block to the deferred blocks list
				ps.DeferBlocks = append(ps.DeferBlocks, *deferredBlock)
				return arg0
			case env.Word:
				// Create a new block that contains "arg0 .with arg1" equivalent
				// We need to create a block that will inject arg0 into block when executed

				// Create the objects for the deferred operation:
				// arg0 (the value) .with block
				withIdx := ps.Idx.IndexWord("with")
				withWord := *env.NewOpword(withIdx, 0) // .with as opword

				block1 := make([]env.Object, 1)
				block1[0] = *env.NewOpword(block.Index, 0) // the value to inject
				series1 := env.NewTSeries(block1)
				block1r := env.NewBlock(*series1)

				// Create a new series with: value, .with, block
				objects := make([]env.Object, 3)
				objects[0] = arg0     // the value to inject
				objects[1] = withWord // .with opword
				objects[2] = *block1r // the block to execute

				series := env.NewTSeries(objects)
				deferredBlock := env.NewBlock(*series)

				// Add the constructed block to the deferred blocks list
				ps.DeferBlocks = append(ps.DeferBlocks, *deferredBlock)
				return arg0
			default:
				return MakeArgError(ps, 2, []env.Type{env.BlockType}, "defer\\with")
			}
		},
	},

	// Tests:
	// equal  { assert { 1 + 2 } 3 } 3
	// equal  { assert { "hello" } "hello" } "hello"
	// error  { assert { 1 + 2 } 5 }
	// Args:
	// * block: Block of code to evaluate
	// * expected: Expected result value
	// Returns:
	// * The result value if assertion passes, error if it fails
	"assert": {
		Argsn: 2,
		Doc:   "Evaluates a block of code and asserts that its result equals the expected value. Returns the result if assertion passes, fails with error message if not.",
		Pure:  false,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch bloc := arg0.(type) {
			case env.Block:
				ser := ps.Ser
				ps.Ser = bloc.Series
				EvalBlock(ps)
				ps.Ser = ser

				// Check if evaluation produced an error
				if ps.ErrorFlag {
					return ps.Res
				}
				if ps.FailureFlag {
					return ps.Res
				}

				result := ps.Res
				expected := arg1

				// Compare result with expected value
				if result.Equal(expected) {
					return result
				}

				// Assertion failed - create error message
				resultStr := result.Inspect(*ps.Idx)
				expectedStr := expected.Inspect(*ps.Idx)

				ps.FailureFlag = true
				return env.NewError("Assertion failed: expected `" + expectedStr + "` but got `" + resultStr + "`")
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType}, "assert")
			}
		},
	},

	// Tests:
	// stdout { assert\display "check addition" { 1 + 2 } 3 } "\033[35mcheck addition\033[0m: \033[32mOK\033[0m\n"
	// equal  { assert\display "check addition" { 1 + 2 } 5 } 3
	// Args:
	// * explanation: String describing the test being performed
	// * block: Block of code to evaluate
	// * expected: Expected result value
	// Returns:
	// * The result value (displays status but continues evaluation even on failure)
	"assert\\display": {
		Argsn: 3,
		Doc:   "Evaluates a block of code and displays the result. Prints the explanation in magenta followed by OK in green on success, or error message in red on failure. Does not fail - continues evaluation.",
		Pure:  false,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch explanation := arg0.(type) {
			case env.String:
				switch bloc := arg1.(type) {
				case env.Block:
					ser := ps.Ser
					ps.Ser = bloc.Series
					EvalBlock(ps)
					ps.Ser = ser

					// ANSI color codes
					magenta := "\033[35m"
					green := "\033[32m"
					red := "\033[31m"
					reset := "\033[0m"

					// Check if evaluation produced an error - display but clear flags and continue
					if ps.ErrorFlag {
						fmt.Println(magenta + explanation.Value + reset + ": " + red + "FAILED: " + ps.Res.Inspect(*ps.Idx) + reset)
						ps.ErrorFlag = false
						ps.FailureFlag = false
						return ps.Res
					}
					if ps.FailureFlag {
						fmt.Println(magenta + explanation.Value + reset + ": " + red + "FAILED: " + ps.Res.Inspect(*ps.Idx) + reset)
						ps.FailureFlag = false
						return ps.Res
					}

					result := ps.Res
					expected := arg2

					// Compare result with expected value
					if result.Equal(expected) {
						fmt.Println(magenta + explanation.Value + reset + ": " + green + "OK" + reset)
						return result
					}

					// Assertion failed - display but continue (don't set failure flag)
					resultStr := result.Inspect(*ps.Idx)
					expectedStr := expected.Inspect(*ps.Idx)

					fmt.Println(magenta + explanation.Value + reset + ": " + red + "FAILED: expected `" + expectedStr + "` but got `" + resultStr + "`" + reset)
					return result
				default:
					return MakeArgError(ps, 2, []env.Type{env.BlockType}, "assert\\display")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "assert\\display")
			}
		},
	},

	// Tests:
	// Args:
	// * condition: Integer value, if > 0 triggers recursion reset
	// Returns:
	// * nil if recursion continues, ps.Res if recursion ends
	"recur-if": {
		Argsn: 2,
		Doc:   "Internal recursion control function. Resets function execution if condition > 0, otherwise returns result.",
		Pure:  false,
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
				return MakeArgError(ps, 1, []env.Type{env.IntegerType}, "recur-if")
			}
		},
	},

	// Tests:
	// Args:
	// * condition: Integer value, if > 0 triggers recursion with updated argument
	// * arg1: New value for the first function argument during recursion
	// Returns:
	// * nil if recursion continues, ps.Res if recursion ends
	"recur-if\\1": {
		Argsn: 2,
		Doc:   "Internal recursion control function for single-argument functions. Updates first argument and resets if condition > 0.",
		Pure:  false,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch cond := arg0.(type) {
			case env.Integer:
				if cond.Value > 0 {
					switch arg := arg1.(type) {
					case env.Integer:
						ps.Ctx.Mod(ps.Args[0], arg)
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

	// Tests:
	// Args:
	// * condition: Integer value, if > 0 triggers recursion with updated arguments
	// * arg1: New value for the first function argument during recursion
	// * arg2: New value for the second function argument during recursion
	// Returns:
	// * ps.Res (function result) regardless of condition
	"recur-if\\2": {
		Argsn: 3,
		Doc:   "Internal recursion control function for two-argument functions. Updates both arguments and resets if condition > 0.",
		Pure:  false,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
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

	// Tests:
	// Args:
	// * condition: Integer value, if > 0 triggers recursion with updated arguments
	// * arg1: New value for the first function argument during recursion
	// * arg2: New value for the second function argument during recursion
	// * arg3: New value for the third function argument during recursion
	// Returns:
	// * ps.Res (function result) regardless of condition, nil if recursion continues
	"recur-if\\3": {
		Argsn: 4,
		Doc:   "Internal recursion control function for three-argument functions. Updates all three arguments and resets if condition > 0.",
		Pure:  false,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
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
							default:
								return MakeArgError(ps, 4, []env.Type{env.IntegerType}, "recur-if\\3")
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
	RegisterBuiltins2(builtins_complex, ps, "base")
	RegisterBuiltins2(builtins_time, ps, "base")
	RegisterBuiltins2(builtins_string, ps, "base")
	RegisterBuiltins2(builtins_collection, ps, "base")
	RegisterBuiltins2(builtins_conditionals, ps, "base")
	RegisterBuiltins2(builtins_combinators, ps, "base")
	RegisterBuiltins2(builtins_printing, ps, "base")
	RegisterBuiltins2(builtins_types, ps, "base")
	RegisterBuiltins2(builtins_iteration, ps, "base")
	RegisterBuiltins2(builtins_contexts, ps, "base")
	RegisterBuiltins2(builtins_persistent_contexts, ps, "base")
	RegisterBuiltins2(builtins_functions, ps, "base")
	// already in base_functions RegisterBuiltins2(builtins_apply, ps, "base")
	RegisterBuiltins2(Builtins_error_creation, ps, "error-creation")
	RegisterBuiltins2(Builtins_error_inspection, ps, "error-inspection")
	RegisterBuiltins2(Builtins_error_handling, ps, "error-handling")
	RegisterBuiltins2(Builtins_match, ps, "match")
	RegisterBuiltins2(Builtins_table, ps, "table")
	RegisterBuiltins2(Builtins_vector, ps, "vector")
	RegisterBuiltins2(Builtins_io, ps, "io")
	RegisterBuiltins2(Builtins_cmd, ps, "cmd")
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
	RegisterBuiltins2(Builtins_msgdispatcher, ps, "msgdispatcher")
	RegisterBuiltins2(Builtins_http, ps, "http")
	RegisterBuiltinsInContext(Builtins_gin, ps, "gin")
	RegisterBuiltins2(Builtins_sqlite, ps, "sqlite")
	RegisterBuiltins2(Builtins_psql, ps, "psql")
	RegisterBuiltins2(Builtins_mysql, ps, "mysql")
	RegisterBuiltins2(Builtins_email, ps, "email")
	RegisterBuiltins2(Builtins_imap, ps, "imap")
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
	RegisterBuiltinsInContext(Builtins_termstr, ps, "termstr")
	RegisterBuiltinsInContext(Builtins_telegrambot, ps, "telegram")
	RegisterBuiltinsInContext(Builtins_mcp, ps, "mcp")
	RegisterBuiltins2(Builtins_mqtt, ps, "mqtt")
	RegisterBuiltins2(builtins_trees, ps, "trees")
	RegisterBuiltinsInContext(Builtins_git, ps, "git")
	RegisterBuiltinsInContext(Builtins_prometheus, ps, "prometheus")
	RegisterBuiltinsInContext(Builtins_echarts, ps, "echarts")
	RegisterErrorUtilsBuiltins(ps) // Register additional error handling utilities
	// ## Archived modules
	// RegisterBuiltins2(Builtins_gtk, ps, "gtk")
	// RegisterBuiltins2(Builtins_nats, ps, "nats")
	// RegisterBuiltins2(Builtins_qframe, ps, "qframe")
	// RegisterBuiltins2(Builtins_nng, ps, "nng")
	// RegisterBuiltins2(Builtins_raylib, ps, "raylib")
	// RegisterBuiltins2(Builtins_cayley, ps, "cayley")

	// Execute initialization code after everything is loaded but before any custom code is evaluated
	executeInitializationCode(ps)
}

// executeInitializationCode runs the injected initialization code after builtins are loaded
// but before any custom user code is evaluated. This works for both REPL and file execution modes.
func executeInitializationCode(ps *env.ProgramState) {
	// These are defined in the parent context of the actual script context. Which is what we want
	initCode := " collector: context { var 'collected ref { } , collect!: fn { v } { .append! 'collected } , collected?: does { deref collected } } " +
		" use: fn { c b } { .clone\\deep .do\\inx b }\n" +
		" cmdo: fn { b } { cmd b |Output } "

	// Save current state
	/* origSer := ps.Ser
	origRes := ps.Res
	origErrorFlag := ps.ErrorFlag
	origFailureFlag := ps.FailureFlag
	origReturnFlag := ps.ReturnFlag*/

	// Load and parse the initialization code
	block := loader.LoadStringNEW(initCode, false, ps)

	// Check if loading was successful
	switch loadedBlock := block.(type) {
	case env.Block:
		// Set the series to the loaded initialization code
		ps.Ser = loadedBlock.Series

		// Execute the initialization code
		EvalBlock(ps)

		// If there were errors, we silently ignore them for initialization
		// This prevents initialization issues from breaking the main program

	case env.Error:
		// Silently ignore loading errors for initialization code
		// This prevents malformed init code from breaking the interpreter
	}

	// Restore original state (except for any side effects the init code may have had)
	/* ps.Ser = origSer
	ps.Res = origRes
	ps.ErrorFlag = origErrorFlag
	ps.FailureFlag = origFailureFlag
	ps.ReturnFlag = origReturnFlag*/
}

func RegisterVarBuiltins(ps *env.ProgramState) {
	// RegisterVarBuiltins2(VarBuiltins_demo, ps, "test")
}

var BuiltinNames map[string]int // TODO --- this looks like some hanging global ... it should move to ProgramState, it doesn't even really work with contrib and external probably

func RegisterBuiltins2(builtins map[string]*env.Builtin, ps *env.ProgramState, name string) {
	BuiltinNames[name] = len(builtins)
	for k, v := range builtins {
		bu := env.NewBuiltin(v.Fn, v.Argsn, v.AcceptFailure, v.Pure, v.Doc+" ("+name+")")
		registerBuiltin(ps, k, *bu)
	}
}

func RegisterVarBuiltins2(builtins map[string]*env.VarBuiltin, ps *env.ProgramState, name string) {
	BuiltinNames[name] = len(builtins)
	for k, v := range builtins {
		bu := env.NewVarBuiltin(v.Fn, v.Argsn, v.AcceptFailure, v.Pure, v.Doc+" ("+name+")")
		registerVarBuiltin(ps, k, *bu)
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

func registerVarBuiltin(ps *env.ProgramState, word string, builtin env.VarBuiltin) {
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
