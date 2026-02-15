package evaldo

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/refaktor/rye/env"

	"github.com/refaktor/rye/util"
)

var builtins_types = map[string]*env.Builtin{

	//
	// ##### Types and Kinds ##### ""
	//
	// Tests:
	// equal { to-integer "1234" } 1234
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
				return *env.NewString(string(rune(value.Value)))
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType}, "to-char")
			}
		},
	},

	// Tests:
	// equal { list [ 1 2 3 ] |to-block |type? } 'block
	// equal  { list [ 1 2 3 ] |to-block |first } 1
	// equal { vector [ 1 2 3 ] |to-block } { 1.0 2.0 3.0 }
	// equal { table { 'a 'b } { 1 2 3 4 } -> 1 |to-block } { 1 2 }
	// Args:
	// * value: List, Vector, or TableRow to convert to a block
	// Returns:
	// * A block containing the same elements as the input
	"to-block": { // ***
		Argsn: 1,
		Doc:   "Turns a List, Vector, or TableRow to a Block",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch val := arg0.(type) {
			case env.List:
				return env.List2Block(ps, val)
			case env.Vector:
				items := make([]env.Object, len(val.Value))
				for i, v := range val.Value {
					items[i] = *env.NewDecimal(v)
				}
				return *env.NewBlock(*env.NewTSeries(items))
			case env.TableRow:
				items := make([]env.Object, len(val.Values))
				for i, v := range val.Values {
					items[i] = env.ToRyeValue(v)
				}
				return *env.NewBlock(*env.NewTSeries(items))
			default:
				return MakeArgError(ps, 1, []env.Type{env.ListType, env.VectorType, env.TableRowType}, "to-block")
			}
		},
	},

	// Tests:
	// equal   { dict [ "a" 1 "b" 2 "c" 3 ] |to-context |type? } 'context
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
				return MakeArgError(ps, 1, []env.Type{env.DictType}, "to-contextXX")
			}
		},
	},

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
			case env.Opword:
				return *env.NewWord(str.Index)
			case env.Pipeword:
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
				return MakeArgError(ps, 1, []env.Type{env.StringType, env.WordType, env.OpwordType, env.PipewordType, env.XwordType, env.EXwordType}, "to-word")
			}
		},
	},

	// Tests:
	// equal   { is-string "test" } true
	// equal   { is-string 'test } false
	// equal   { is-string 123 } false
	// Args:
	// * value: Any Rye value to test
	// Returns:
	// * Boolean true if the value is a string, false otherwise
	"is-string": { // ***
		Argsn: 1,
		Doc:   "Returns true if value is a string.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			if arg0.Type() == env.StringType {
				return *env.NewBoolean(true)
			} else {
				return *env.NewBoolean(false)
			}
		},
	},

	// Tests:
	// equal   { is-integer 123 } true
	// equal   { is-integer 123.4 } false
	// equal   { is-integer "123" } false
	// Args:
	// * value: Any Rye value to test
	// Returns:
	// * Boolean true if the value is an integer, false otherwise
	"is-integer": { // ***
		Argsn: 1,
		Doc:   "Returns true if value is an integer.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			if arg0.Type() == env.IntegerType {
				return *env.NewBoolean(true)
			} else {
				return *env.NewBoolean(false)
			}
		},
	},

	// Tests:
	// equal   { is-decimal 123.4 } true
	// equal   { is-decimal 123 } false
	// equal   { is-decimal "123.4" } false
	// Args:
	// * value: Any Rye value to test
	// Returns:
	// * Boolean true if the value is a decimal, false otherwise
	"is-decimal": { // ***
		Argsn: 1,
		Doc:   "Returns true if value is a decimal.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			if arg0.Type() == env.DecimalType {
				return *env.NewBoolean(true)
			} else {
				return *env.NewBoolean(false)
			}
		},
	},

	// Tests:
	// equal   { is-number 123 } true
	// equal   { is-number 123.4 } true
	// equal   { is-number "123" } false
	// Args:
	// * value: Any Rye value to test
	// Returns:
	// * Boolean true if the value is a number (integer or decimal), false otherwise
	"is-number": { // ***
		Argsn: 1,
		Doc:   "Returns true if value is a number (integer or decimal).",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			if arg0.Type() == env.IntegerType || arg0.Type() == env.DecimalType {
				return *env.NewBoolean(true)
			} else {
				return *env.NewBoolean(false)
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
	// equal   { kind? %file } 'file-uri
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
	// equal   { person: kind 'person { name: "" age: 0 } assure-kind dict { "name" "John" "age" 30 } person |type? } 'context
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
	// equal   { method 'integer 'add fn { a b } { a + b } |type? } 'function
	// Args:
	// * kind: Word representing the kind for which to register the function
	// * method: Word representing the method name
	// * function: Function to register for the kind and method
	// Returns:
	// * The registered function
	// method <integer> <add> fn [ a b ] [ a + b ] // tagwords are temporary here
	"method": {
		Argsn: 3,
		Doc:   "Registers a method (generic function).",
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
	// equal { secret "password" |type? } 'secret
	// equal { secret "password" |reveal } "password"
	// error { secret 123 }
	// Args:
	// * value: String value to convert to a secret
	// Returns:
	// * A secret value that hides the string from casual inspection
	"secret": {
		Argsn: 1,
		Doc:   "Creates a secret from a string.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch str := arg0.(type) {
			case env.String:
				return *env.NewSecret(str.Value)
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "secret")
			}
		},
	},

	// Tests:
	// equal { secret "password" |reveal } "password"
	// equal { secret "test123" |reveal } "test123"
	// error { reveal "not-a-secret" }
	// Args:
	// * secret: Secret value to reveal
	// Returns:
	// * The string value contained in the secret
	"reveal": {
		Argsn: 1,
		Doc:   "Reveals the secret value.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s := arg0.(type) {
			case env.Secret:
				return *env.NewString(s.Value)
			default:
				return MakeArgError(ps, 1, []env.Type{env.SecretType}, "reveal")
			}
		},
	},

	// Tests:
	// ; equal   { person: kind 'person { name: "" age: 0 } , { "name" "John" "age" 30 } .dict >> person |type? } 'ctx
	// Args:
	// * value: Dict or context to convert
	// * kind: Kind to convert the value to
	// Returns:
	// * A new context of the specified kind
	"lazy": {
		Argsn: 1,
		Doc:   "Creates a lazy value from a block. The block will only be evaluated when forced with _!.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch block := arg0.(type) {
			case env.Block:
				return env.NewLazyValue(block, ps.Ctx)
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType}, "lazy")
			}
		},
	},

	"_!": {
		Argsn: 1,
		Doc:   "Forces evaluation of a lazy value. If frozen evaluates the block and caches the result. If already thawed returns cached result.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch lv := arg0.(type) {
			case *env.LazyValue:
				switch lv.State {
				case 0: // Frozen - evaluate the block
					ser := ps.Ser
					ps.Ser = lv.Block.Series
					ps.Ser.Reset()
					EvalBlockInCtxInj(ps, lv.Ctx, nil, false)
					ps.Ser = ser
					if ps.ErrorFlag {
						lv.State = 2
						lv.Result = ps.Res
						return ps.Res
					}
					lv.State = 1
					lv.Result = ps.Res
					return lv.Result
				case 1: // Thawed - return cached result
					return lv.Result
				case 2: // Error
					return lv.Result
				default:
					return MakeBuiltinError(ps, "Unknown lazy value state.", "_!")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.LazyValueType}, "_!")
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
}
