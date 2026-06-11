package evaldo

import (
	"encoding/hex"
	"strings"
	"unicode"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"

	"github.com/refaktor/rye/env"
)

var Builtins_encoding = map[string]*env.Builtin{

	//
	// ##### Encoding and Text Processing ##### "Functions for encoding/decoding text and hex strings"
	//

	// Tests:
	// equal { "48656c6c6f20576f726c64" |hex\\decode-string } "Hello World"
	// equal { "48656c6c6f20576f726c64" |hex\\decode-string |type? } 'string
	// equal { "invalid" |hex\\decode-string |disarm |type? } 'error
	// Args:
	// * hex-string: hexadecimal string to decode
	// Returns:
	// * string containing the decoded bytes as a UTF-8 string
	"decode\\hex-string": {
		Argsn: 1,
		Doc:   "Decodes a hexadecimal string to a UTF-8 string.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch hexStr := arg0.(type) {
			case env.String:
				decoded, err := hex.DecodeString(hexStr.Value)
				if err != nil {
					ps.FailureFlag = true
					return MakeBuiltinError(ps, "Failed to decode hex string: "+err.Error(), "hex\\decode-string")
				}
				return *env.NewString(string(decoded))
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "hex\\decode-string")
			}
		},
	},

	// Tests:
	// equal { "Hello World" |hex\\encode-string } "48656c6c6f20576f726c64"
	// equal { "Hello World" |hex\\encode-string |type? } 'string
	// equal { "" |hex\\encode-string } ""
	// Args:
	// * string: string to encode
	// Returns:
	// * hexadecimal string representation
	"encode\\hex-string": {
		Argsn: 1,
		Doc:   "Encodes a string to its hexadecimal representation.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch str := arg0.(type) {
			case env.String:
				encoded := hex.EncodeToString([]byte(str.Value))
				return *env.NewString(encoded)
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "hex\\encode-string")
			}
		},
	},

	// Tests:
	// equal { "  70 44 01 4b  \n  55 50 4e 51  " |hex\\clean-string } "7044014b55504e51"
	// equal { "70 44 01 4b\t55 50 4e 51\r\n52 0a" |hex\\clean-string } "7044014b55504e51520a"
	// equal { "7044014b55504e51520a" |hex\\clean-string } "7044014b55504e51520a"
	// Args:
	// * hex-string: raw hex string with possible whitespace
	// Returns:
	// * cleaned hex string with whitespace removed and even length
	"clean\\hex-string": {
		Argsn: 1,
		Doc:   "Removes all whitespace from a hex string and ensures even length.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch hexStr := arg0.(type) {
			case env.String:
				var sb strings.Builder
				for _, r := range hexStr.Value {
					if !unicode.IsSpace(r) {
						sb.WriteRune(r)
					}
				}

				// Ensure even length for valid hex string
				result := sb.String()
				if len(result)%2 != 0 {
					result = result[:len(result)-1]
				}
				return *env.NewString(result)
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "hex\\clean-string")
			}
		},
	},

	// Tests:
	// equal { charmap\\windows-1250 |type? } 'native
	// equal { charmap\\windows-1250 |kind? } 'charmap-encoding
	// Args:
	// * none
	// Returns:
	// * Windows-1250 encoding as a native value
	"charmap\\windows-1250": {
		Argsn: 0,
		Doc:   "Returns the Windows-1250 (CP1250) character encoding for Central European languages.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewNative(ps.Idx, charmap.Windows1250, "charmap-encoding")
		},
	},

	// Tests:
	// equal { charmap\\iso-8859-1 |type? } 'native
	// equal { charmap\\iso-8859-1 |kind? } 'charmap-encoding
	// Args:
	// * none
	// Returns:
	// * ISO-8859-1 encoding as a native value
	"charmap\\iso-8859-1": {
		Argsn: 0,
		Doc:   "Returns the ISO-8859-1 (Latin-1) character encoding.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewNative(ps.Idx, charmap.ISO8859_1, "charmap-encoding")
		},
	},

	// Tests:
	// equal { charmap\\iso-8859-2 |type? } 'native
	// equal { charmap\\iso-8859-2 |kind? } 'charmap-encoding
	// Args:
	// * none
	// Returns:
	// * ISO-8859-2 encoding as a native value
	"charmap\\iso-8859-2": {
		Argsn: 0,
		Doc:   "Returns the ISO-8859-2 (Latin-2) character encoding for Central European languages.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewNative(ps.Idx, charmap.ISO8859_2, "charmap-encoding")
		},
	},

	// Tests:
	// equal { charmap\\windows-1252 |type? } 'native
	// equal { charmap\\windows-1252 |kind? } 'charmap-encoding
	// Args:
	// * none
	// Returns:
	// * Windows-1252 encoding as a native value
	"charmap\\windows-1252": {
		Argsn: 0,
		Doc:   "Returns the Windows-1252 character encoding for Western European languages.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewNative(ps.Idx, charmap.Windows1252, "charmap-encoding")
		},
	},

	// Tests:
	// equal { charmap\\code-page-437 |type? } 'native
	// equal { charmap\\code-page-437 |kind? } 'charmap-encoding
	// Args:
	// * none
	// Returns:
	// * Code Page 437 encoding as a native value
	"charmap\\code-page-437": {
		Argsn: 0,
		Doc:   "Returns the Code Page 437 (DOS/IBM PC) character encoding.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewNative(ps.Idx, charmap.CodePage437, "charmap-encoding")
		},
	},

	// Tests:
	// equal { charmap\\code-page-850 |type? } 'native
	// equal { charmap\\code-page-850 |kind? } 'charmap-encoding
	// Args:
	// * none
	// Returns:
	// * Code Page 850 encoding as a native value
	"charmap\\code-page-850": {
		Argsn: 0,
		Doc:   "Returns the Code Page 850 (DOS Latin-1) character encoding.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewNative(ps.Idx, charmap.CodePage850, "charmap-encoding")
		},
	},

	// Tests:
	// equal { charmap\\koi8r |type? } 'native
	// equal { charmap\\koi8r |kind? } 'charmap-encoding
	// Args:
	// * none
	// Returns:
	// * KOI8-R encoding as a native value
	"charmap\\koi8r": {
		Argsn: 0,
		Doc:   "Returns the KOI8-R character encoding for Russian.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewNative(ps.Idx, charmap.KOI8R, "charmap-encoding")
		},
	},

	// Tests:
	// equal { charmap\\windows-1250 |charmap-encoding\\new-decoder |type? } 'native
	// equal { charmap\\windows-1250 |charmap-encoding\\new-decoder |kind? } 'text-decoder
	// Args:
	// * encoding: charmap encoding as a native value
	// Returns:
	// * text decoder as a native value
	"charmap-encoding//Decoder": {
		Argsn: 1,
		Doc:   "Creates a new text decoder for the given character encoding.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch encoding := arg0.(type) {
			case env.Native:
				if ps.Idx.GetWord(encoding.GetKind()) != "charmap-encoding" {
					ps.FailureFlag = true
					return MakeArgError(ps, 1, []env.Type{env.NativeType}, "charmap-encoding\\new-decoder")
				}
				enc := encoding.Value.(*charmap.Charmap)
				decoder := enc.NewDecoder()
				return *env.NewNative(ps.Idx, decoder, "text-decoder")
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "charmap-encoding\\new-decoder")
			}
		},
	},

	// Tests:
	// equal { charmap\\windows-1250 |charmap-encoding\\new-encoder |type? } 'native
	// equal { charmap\\windows-1250 |charmap-encoding\\new-encoder |kind? } 'text-encoder
	// Args:
	// * encoding: charmap encoding as a native value
	// Returns:
	// * text encoder as a native value
	"charmap-encoding//Encoder": {
		Argsn: 1,
		Doc:   "Creates a new text encoder for the given character encoding.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch encoding := arg0.(type) {
			case env.Native:
				if ps.Idx.GetWord(encoding.GetKind()) != "charmap-encoding" {
					ps.FailureFlag = true
					return MakeArgError(ps, 1, []env.Type{env.NativeType}, "charmap-encoding\\new-encoder")
				}
				enc := encoding.Value.(*charmap.Charmap)
				encoder := enc.NewEncoder()
				return *env.NewNative(ps.Idx, encoder, "text-encoder")
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "charmap-encoding\\new-encoder")
			}
		},
	},

	// Tests:
	// equal { charmap\\windows-1250 |charmap-encoding\\new-decoder "Plačilo računa" |text-decoder\\string |contains? "č" } 1
	// equal { charmap\\windows-1250 |charmap-encoding\\new-decoder "Hello" |text-decoder\\string } "Hello"
	// equal { charmap\\windows-1250 |charmap-encoding\\new-decoder "" |text-decoder\\string } ""
	// Args:
	// * decoder: text decoder as a native value
	// * input: string to decode
	// Returns:
	// * decoded string
	"text-decoder//Decode": {
		Argsn: 2,
		Doc:   "Decodes a string using the given text decoder.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch decoder := arg0.(type) {
			case env.Native:
				if ps.Idx.GetWord(decoder.GetKind()) != "text-decoder" {
					ps.FailureFlag = true
					return MakeArgError(ps, 1, []env.Type{env.NativeType}, "text-decoder\\string")
				}
				switch input := arg1.(type) {
				case env.String:
					dec := decoder.Value.(*encoding.Decoder)
					decoded, err := dec.String(input.Value)
					if err != nil {
						ps.FailureFlag = true
						return MakeBuiltinError(ps, "Failed to decode string: "+err.Error(), "text-decoder\\string")
					}
					return *env.NewString(decoded)
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "text-decoder\\string")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "text-decoder\\string")
			}
		},
	},

	// Tests:
	// equal { charmap\\windows-1250 |charmap-encoding\\new-encoder "Hello" |text-encoder\\string } "Hello"
	// equal { charmap\\windows-1250 |charmap-encoding\\new-encoder "Plačilo" |text-encoder\\string |hex\\encode-string } "506c61e8696c6f"
	// Args:
	// * encoder: text encoder as a native value
	// * input: string to encode
	// Returns:
	// * encoded string
	"text-encoder//Encode": {
		Argsn: 2,
		Doc:   "Encodes a string using the given text encoder.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch encoder := arg0.(type) {
			case env.Native:
				if ps.Idx.GetWord(encoder.GetKind()) != "text-encoder" {
					ps.FailureFlag = true
					return MakeArgError(ps, 1, []env.Type{env.NativeType}, "text-encoder\\string")
				}
				switch input := arg1.(type) {
				case env.String:
					enc := encoder.Value.(*encoding.Encoder)
					encoded, err := enc.String(input.Value)
					if err != nil {
						ps.FailureFlag = true
						return MakeBuiltinError(ps, "Failed to encode string: "+err.Error(), "text-encoder\\string")
					}
					return *env.NewString(encoded)
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "text-encoder\\string")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "text-encoder\\string")
			}
		},
	},

	// Tests:
	// equal { " " |is-space? } 1
	// equal { "\t" |is-space? } 1
	// equal { "\n" |is-space? } 1
	// equal { "\r" |is-space? } 1
	// equal { "a" |is-space? } 0
	// equal { "A" |is-space? } 0
	// equal { "1" |is-space? } 0
	// Args:
	// * input: string to check (should be single character)
	// Returns:
	// * integer 1 if character is whitespace, 0 otherwise
	"is-space?": {
		Argsn: 1,
		Doc:   "Checks if a single character string is a whitespace character.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch str := arg0.(type) {
			case env.String:
				if len(str.Value) == 0 {
					return *env.NewInteger(0)
				}
				// Check first rune of the string
				for _, r := range str.Value {
					if unicode.IsSpace(r) {
						return *env.NewInteger(1)
					}
					return *env.NewInteger(0)
				}
				return *env.NewInteger(0)
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "is-space?")
			}
		},
	},

	// Tests:
	// equal { "Hello World" |remove-all-space } "HelloWorld"
	// equal { "  Hello  World  " |remove-all-space } "HelloWorld"
	// equal { "\t\n Hello \r\n World \t" |remove-all-space } "HelloWorld"
	// equal { "" |remove-all-space } ""
	// Args:
	// * input: string to process
	// Returns:
	// * string with all whitespace characters removed
	"remove-all-space": {
		Argsn: 1,
		Doc:   "Removes all whitespace characters from a string.",
		Pure:  true,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch str := arg0.(type) {
			case env.String:
				var sb strings.Builder
				for _, r := range str.Value {
					if !unicode.IsSpace(r) {
						sb.WriteRune(r)
					}
				}
				return *env.NewString(sb.String())
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "remove-all-space")
			}
		},
	},
}
