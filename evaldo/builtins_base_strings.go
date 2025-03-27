package evaldo

import (
	"encoding/base64"
	"encoding/pem"

	"github.com/refaktor/rye/env"

	// JM 20230825	"github.com/refaktor/rye/term"
	"strconv"
	"strings"

	"github.com/refaktor/rye/util"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var builtins_string = map[string]*env.Builtin{

	//
	// ##### Strings ##### ""
	//
	// Tests:
	// equal { newline } "\n"
	// Args:
	// * none
	// Returns:
	// * a string containing a single newline character
	"newline": {
		Argsn: 0,
		Doc:   "Returns a string containing a single newline character.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewString("\n")
		},
	},

	// Tests:
	// equal { "123" .ln } "123\n"
	// equal { "hello" .ln } "hello\n"
	// Args:
	// * string: String to append a newline to
	// Returns:
	// * a new string with a newline character appended
	"ln": {
		Argsn: 1,
		Doc:   "Appends a newline character to the end of a string.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.String:
				return *env.NewString(s1.Value + "\n")
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "nl")
			}
		},
	},

	// Tests:
	// equal { trim " ASDF " } "ASDF"
	// equal { trim "   ASDF   " } "ASDF"
	// equal { trim "\t\nASDF\r\n" } "ASDF"
	// Args:
	// * string: String to trim
	// Returns:
	// * a new string with leading and trailing whitespace removed
	"trim": {
		Argsn: 1,
		Doc:   "Removes all leading and trailing whitespace characters from a string.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.String:
				return *env.NewString(strings.TrimSpace(s1.Value))
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "trim\\space")
			}
		},
	},

	// Tests:
	// equal { trim\ "__ASDF__" "_" } "ASDF"
	// equal { trim\ "##Hello##" "#" } "Hello"
	// Args:
	// * string: String to trim
	// * cutset: String containing the characters to trim
	// Returns:
	// * a new string with specified characters removed from both ends
	"trim\\": {
		Argsn: 2,
		Doc:   "Removes all leading and trailing occurrences of the specified characters from a string.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.String:
				switch s2 := arg1.(type) {
				case env.String:
					return *env.NewString(strings.Trim(s1.Value, s2.Value))
				default:
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "trim")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "trim")
			}
		},
	},

	// Tests:
	// equal { trim\right "__ASDF__" "_" } "__ASDF"
	// equal { trim\right "  ASDF  " " " } "  ASDF"
	// equal { trim\right "Hello!!!" "!" } "Hello"
	// Args:
	// * string: String to trim
	// * cutset: String containing the characters to trim
	// Returns:
	// * a new string with specified characters removed from the right end
	"trim\\right": {
		Argsn: 2,
		Doc:   "Removes all trailing occurrences of the specified characters from a string.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.String:
				switch s2 := arg1.(type) {
				case env.String:
					return *env.NewString(strings.TrimRight(s1.Value, s2.Value))
				default:
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "trim\\right")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "trim\\right")
			}
		},
	},
	// Tests:
	// equal { trim\left "___ASDF__" "_" } "ASDF__"
	// equal { trim\left "  ASDF  " " " } "ASDF  "
	// equal { trim\left "###Hello" "#" } "Hello"
	// Args:
	// * string: String to trim
	// * cutset: String containing the characters to trim
	// Returns:
	// * a new string with specified characters removed from the left end
	"trim\\left": {
		Argsn: 2,
		Doc:   "Removes all leading occurrences of the specified characters from a string.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.String:
				switch s2 := arg1.(type) {
				case env.String:
					return *env.NewString(strings.TrimLeft(s1.Value, s2.Value))
				default:
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "trim\\left")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "trim\\left")
			}
		},
	},
	// Tests:
	// equal { replace "...xo..." "xo" "LoL" } "...LoL..."
	// equal { replace "...xoxo..." "xo" "LoL" } "...LoLLoL..."
	// equal { replace "hello world" "world" "everyone" } "hello everyone"
	// Args:
	// * string: Original string
	// * old: Substring to replace
	// * new: Replacement string
	// Returns:
	// * a new string with all occurrences of old replaced by new
	"replace": {
		Argsn: 3,
		Doc:   "Replaces all occurrences of a substring with another string.",
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
	// Todo: this could be a general slice function that also works with blocks and
	// Tests:
	// equal { substring "xoxo..." 0 4 } "xoxo"
	// equal { substring "...xoxo..." 3 7 } "xoxo"
	// equal { substring "hello world" 6 11 } "world"
	// Args:
	// * string: String to extract from
	// * start: Starting position (0-based, inclusive)
	// * end: Ending position (0-based, exclusive)
	// Returns:
	// * a new string containing the specified substring
	"substring": {
		Argsn: 3,
		Doc:   "Extracts a portion of a string between the specified start and end positions.",
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
	// Tests:
	// equal { contains "...xoxo..." "xo"  } 1
	// equal { contains "...xoxo..." "lol" } 0
	// equal { contains { ".." "xoxo" ".." } "xoxo" } 1
	// equal { contains { ".." "xoxo" ".." } "lol"  } 0
	// equal { contains list { 1 2 3 } 2 } 1
	// Args:
	// * collection: String, block or list to search in
	// * value: Value to search for
	// Returns:
	// * integer 1 if the collection contains the value, 0 otherwise
	"contains": {
		Argsn: 2,
		Doc:   "Checks if a string, block or list contains a specific value, returning 1 if found or 0 if not found.",
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
				case env.Object:
					if util.ContainsVal(ps, s1.Series.S, value) {
						return *env.NewInteger(1)
					} else {
						return *env.NewInteger(0)
					}
				default:
					return MakeArgError(ps, 2, []env.Type{}, "contains")
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
	// Tests:
	// equal { has-suffix "xoxo..." "xoxo" } 0
	// equal { has-suffix "...xoxo" "xoxo" } 1
	// equal { has-suffix "hello.txt" ".txt" } 1
	// Args:
	// * string: String to check
	// * suffix: Suffix to look for
	// Returns:
	// * integer 1 if the string ends with the suffix, 0 otherwise
	"has-suffix": {
		Argsn: 2,
		Doc:   "Checks if a string ends with a specific suffix, returning 1 if true or 0 if false.",
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
	// Tests:
	// equal { has-prefix "xoxo..." "xoxo" } 1
	// equal { has-prefix "...xoxo" "xoxo" } 0
	// equal { has-prefix "http://example.com" "http://" } 1
	// Args:
	// * string: String to check
	// * prefix: Prefix to look for
	// Returns:
	// * integer 1 if the string starts with the prefix, 0 otherwise
	"has-prefix": {
		Argsn: 2,
		Doc:   "Checks if a string starts with a specific prefix, returning 1 if true or 0 if false.",
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

	// Todos:
	// - fail if not found
	// Tests:
	// equal { index? "...xo..." "xo" } 3
	// equal { index? "xo..." "xo" } 0
	// equal { index? { "xo" ".." } "xo" } 0
	// equal { index? { ".." "xo" ".." } "xo" } 1
	// Args:
	// * collection: String or block to search in
	// * value: Value to search for
	// Returns:
	// * integer index (0-based) of the first occurrence of the value, or -1 if not found
	"index?": {
		Argsn: 2,
		Doc:   "Finds the 0-based index of the first occurrence of a value in a string or block.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.String:
				switch s2 := arg1.(type) {
				case env.String:
					res := strings.Index(s1.Value, s2.Value)
					return *env.NewInteger(int64(res))
				default:
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "index?")
				}
			case env.Block:
				res := util.IndexOfSlice(ps, s1.Series.S, arg1)
				if res == -1 {
					return MakeBuiltinError(ps, "not found", "index?")
				}
				return *env.NewInteger(int64(res))
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType, env.StringType}, "index?")
			}
		},
	},
	// Todos:
	// - fail if not found
	// Tests:
	// equal { position? "...xo..." "xo" } 4
	// equal { position? "xo..." "xo" } 1
	// equal { position? { "xo" ".." } "xo" } 1
	// equal { position? { ".." "xo" ".." } "xo" } 2
	// Args:
	// * collection: String or block to search in
	// * value: Value to search for
	// Returns:
	// * integer position (1-based) of the first occurrence of the value, or error if not found
	"position?": {
		Argsn: 2,
		Doc:   "Finds the 1-based position of the first occurrence of a value in a string or block.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.String:
				switch s2 := arg1.(type) {
				case env.String:
					res := strings.Index(s1.Value, s2.Value)
					return *env.NewInteger(int64(res + 1))
				default:
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "position?")
				}
			case env.Block:
				res := util.IndexOfSlice(ps, s1.Series.S, arg1)
				if res == -1 {
					return MakeBuiltinError(ps, "not found", "position?")
				}
				return *env.NewInteger(int64(res + 1))
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType, env.BlockType}, "position?")
			}
		},
	},

	// Tests:
	// equal { encode-to\base64 "abcd" } "YWJjZA=="
	// equal { encode-to\base64 "hello world" } "aGVsbG8gd29ybGQ="
	// Args:
	// * data: String or native bytes/pem-block to encode
	// Returns:
	// * base64-encoded string
	"encode-to\\base64": {
		Argsn: 1,
		Doc:   "Encodes a string or binary data as a base64 string.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Native:
				switch ps.Idx.GetWord(s1.GetKind()) {
				case "bytes":
					bb, ok := s1.Value.([]byte)
					if ok {
						ata := base64.StdEncoding.EncodeToString(bb)
						return *env.NewString(string(ata))
					}
					return MakeError(ps, "Native not of kind 1: bytes")
				case "pem-block":
					bb, ok := s1.Value.(*pem.Block)
					if ok {
						ata := base64.StdEncoding.EncodeToString(bb.Bytes)
						return *env.NewString(string(ata))
					}
					return MakeError(ps, "Native not of kind 2: bytes")
				}
				return MakeError(ps, "Native not of kind 3: bytes")
			case env.String:
				ata := base64.StdEncoding.EncodeToString([]byte(s1.Value))
				return *env.NewString(string(ata))
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType, env.NativeType}, "base64-encode")
			}
		},
	},

	// Tests:
	// equal { decode\base64 encode\base64 "abcd" } "abcd"
	// equal { decode\base64 "aGVsbG8gd29ybGQ=" } "hello world"
	// Args:
	// * string: Base64-encoded string to decode
	// Returns:
	// * decoded string
	"decode\\base64": {
		Argsn: 1,
		Doc:   "Decodes a base64-encoded string back to its original form.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.String:
				ata, err := base64.StdEncoding.DecodeString(s1.Value)
				if err != nil {
					return MakeBuiltinError(ps, err.Error(), "base64-decode")
				}
				return *env.NewString(string(ata))
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "base64-encode")
			}
		},
	},

	// Tests:
	// equal { "abcd" .space } "abcd "
	// equal { "" .space } " "
	// Args:
	// * string: String to append a space to
	// Returns:
	// * a new string with a space character appended
	"space": {
		Argsn: 1,
		Doc:   "Appends a space character to the end of a string.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.String:
				return *env.NewString(s1.Value + " ")
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "space")
			}
		},
	},

	// Tests:
	// equal { capitalize "abcde" } "Abcde"
	// equal { capitalize "hello world" } "Hello World"
	// equal { capitalize "HELLO" } "Hello"
	// Args:
	// * string: String to capitalize
	// Returns:
	// * a new string with the first character converted to uppercase
	"capitalize": { // **
		Argsn: 1,
		Pure:  true,
		Doc:   "Converts the first character of a string to uppercase, leaving the rest unchanged.",
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

	// Tests:
	// equal { to-lower "ABCDE" } "abcde"
	// equal { to-lower "Hello World" } "hello world"
	// equal { to-lower "123ABC" } "123abc"
	// Args:
	// * string: String to convert
	// Returns:
	// * a new string with all characters converted to lowercase
	"to-lower": { // **
		Argsn: 1,
		Pure:  true,
		Doc:   "Converts all characters in a string to lowercase.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.String:
				return *env.NewString(strings.ToLower(s1.Value))
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "to-lower")
			}
		},
	},

	// Tests:
	// equal { to-upper "abcde" } "ABCDE"
	// equal { to-upper "Hello World" } "HELLO WORLD"
	// equal { to-upper "123abc" } "123ABC"
	// Args:
	// * string: String to convert
	// Returns:
	// * a new string with all characters converted to uppercase
	"to-upper": { // **
		Argsn: 1,
		Pure:  true,
		Doc:   "Converts all characters in a string to uppercase.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case *env.String:
				return *env.NewString(strings.ToUpper(s1.Value))
			case env.String:
				return *env.NewString(strings.ToUpper(s1.Value))
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "to-upper")
			}
		},
	},

	// Tests:
	// equal { concat3 "aa" "BB" "cc" } "aaBBcc"
	// equal { concat3 "hello" " " "world" } "hello world"
	// Args:
	// * string1: First string
	// * string2: Second string
	// * string3: Third string
	// Returns:
	// * a new string containing all three strings concatenated together
	"concat3": {
		Argsn: 3,
		Pure:  true,
		Doc:   "Concatenates three strings together into a single string.",
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

	// Tests:
	// equal { join { "Mary" "Anne" } } "MaryAnne"
	// equal { join { "Spot" "Fido" "Rex" } } "SpotFidoRex"
	// equal { join { 1 2 3 } } "123"
	// Args:
	// * collection: Block or list of strings or numbers to join
	// Returns:
	// * a single string with all values concatenated together
	"join": { // **
		Argsn: 1,
		Pure:  true,
		Doc:   "Concatenates all strings or numbers in a block or list into a single string with no separator.",
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

	// Tests:
	// equal { join\with { "Mary" "Anne" } " " } "Mary Anne"
	// equal { join\with { "Spot" "Fido" "Rex" } "/" } "Spot/Fido/Rex"
	// equal { join\with { 1 2 3 } "-" } "1-2-3"
	// Args:
	// * collection: Block or list of strings or numbers to join
	// * delimiter: String to insert between each value
	// Returns:
	// * a single string with all values concatenated with the delimiter between them
	"join\\with": { // **
		Argsn: 2,
		Pure:  true,
		Doc:   "Concatenates all strings or numbers in a block or list into a single string with a specified delimiter between values.",
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

	// Tests:
	// equal { split "a,b,c" "," } { "a" "b" "c" }
	// equal { split "hello world" " " } { "hello" "world" }
	// equal { split "one::two::three" "::" } { "one" "two" "three" }
	// Args:
	// * string: String to split
	// * separator: String that separates values
	// Returns:
	// * a block of strings resulting from splitting the input string
	"split": { // **
		Argsn: 2,
		Pure:  true,
		Doc:   "Splits a string into a block of strings using a separator string.",
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

	// Tests:
	// equal { split\quoted "`a,b`,c,d" "," "`" } { "`a,b`" "c" "d" }
	// equal { split\quoted "'hello, world',foo,bar" "," "'" } { "'hello, world'" "foo" "bar" }
	// Args:
	// * string: String to split
	// * separator: String that separates values
	// * quote: String that marks quoted sections that should not be split
	// Returns:
	// * a block of strings resulting from splitting the input string, preserving quoted sections
	"split\\quoted": { // **
		Argsn: 3,
		Pure:  true,
		Doc:   "Splits a string into a block of strings using a separator, while respecting quoted sections that should remain intact.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch str := arg0.(type) {
			case env.String:
				switch sepa := arg1.(type) {
				case env.String:
					switch quote := arg2.(type) {
					case env.String:
						return util.StringToFieldsWithQuoted(str.Value, sepa.Value, quote.Value)
					default:
						return MakeArgError(ps, 3, []env.Type{env.StringType}, "split\\quoted")
					}
				default:
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "split\\quoted")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "split\\quoted")
			}
		},
	},

	// Tests:
	// equal { split\many "192.0.0.1" "." } { "192" "0" "0" "1" }
	// equal { split\many "abcd..e.q|bar" ".|" } { "abcd" "e" "q" "bar" }
	// equal { split\many "a;b,c:d" ";,:" } { "a" "b" "c" "d" }
	// Args:
	// * string: String to split
	// * separators: String containing all characters that should be treated as separators
	// Returns:
	// * a block of strings resulting from splitting the input string at any of the separator characters
	"split\\many": {
		Argsn: 2,
		Pure:  true,
		Doc:   "Splits a string into a block of strings using any character in the separators string as a delimiter.",
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

	// Tests:
	// equal { split\every "abcdefg" 3 } { "abc" "def" "g" }
	// equal { split\every "abcdefg" 2 } { "ab" "cd" "ef" "g" }
	// equal { split\every "abcdef" 2 } { "ab" "cd" "ef" }
	// equal { split\every { 1 2 3 4 5 } 2 } { { 1 2 } { 3 4 } { 5 } }
	// Args:
	// * collection: String or block to split
	// * size: Number of elements in each chunk
	// Returns:
	// * a block of strings or blocks, each containing at most 'size' elements
	"split\\every": { // **
		Argsn: 2,
		Pure:  true,
		Doc:   "Splits a string or block into chunks of the specified size, with any remaining elements in the last chunk.",
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
					return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "split\\every")
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
					return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "split\\every")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType, env.BlockType}, "split\\every")
			}
		},
	},
}
