//go:build no_baseio
// +build no_baseio

package evaldo

// builtins_base_printing_not.go — lightweight printing builtins used when the
// no_baseio build tag is active.  The terminal-interactive builtins (display,
// _.. , display\custom) and the CSV/SSV printers are omitted so that the embed
// module has no dependency on the term / util / keyboard packages.

import (
	"fmt"
	"strings"

	"github.com/refaktor/rye/env"
)

// DisplayRyeValue is a non-interactive stub: it just returns the value as-is.
// The full interactive version (using the term package) lives in
// builtins_base_printing.go which is excluded under no_baseio.
func DisplayRyeValue(ps *env.ProgramState, arg0 env.Object, interactive bool) (env.Object, string) {
	return arg0, ""
}

var builtins_printing = map[string]*env.Builtin{

	//
	// ##### Printing ##### "Functions for displaying and formatting values"
	//

	"prns": {
		Argsn: 1,
		Doc:   "Prints a value followed by a space, returning the input value.",
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

	"prn": {
		Argsn: 1,
		Doc:   "Prints a value without adding a newline, returning the input value.",
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

	"print": {
		Argsn: 1,
		Doc:   "Prints a value followed by a newline, returning the input value.",
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

	"print2": {
		Argsn: 2,
		Doc:   "Prints two values separated by a space and followed by a newline, returning the second value.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch a := arg0.(type) {
			case env.String:
				fmt.Print(a.Value)
			default:
				fmt.Print(arg0.Print(*ps.Idx))
			}
			fmt.Print(" ")
			switch b := arg1.(type) {
			case env.String:
				fmt.Println(b.Value)
			default:
				fmt.Println(arg1.Print(*ps.Idx))
			}
			return arg1
		},
	},

	"prn2": {
		Argsn: 2,
		Doc:   "Prints two values separated by a space without a newline, returning the second value.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch a := arg0.(type) {
			case env.String:
				fmt.Print(a.Value)
			default:
				fmt.Print(arg0.Print(*ps.Idx))
			}
			fmt.Print(" ")
			switch b := arg1.(type) {
			case env.String:
				fmt.Print(b.Value)
			default:
				fmt.Print(arg1.Print(*ps.Idx))
			}
			return arg1
		},
	},

	"prns2": {
		Argsn: 2,
		Doc:   "Prints two values each followed by a space, returning the second value.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch a := arg0.(type) {
			case env.String:
				fmt.Print(a.Value + " ")
			default:
				fmt.Print(arg0.Print(*ps.Idx) + " ")
			}
			switch b := arg1.(type) {
			case env.String:
				fmt.Print(b.Value + " ")
			default:
				fmt.Print(arg1.Print(*ps.Idx) + " ")
			}
			return arg1
		},
	},

	"format": {
		Argsn: 2,
		Doc:   "Formats a block of values using a format string with {} placeholders, returning the resulting string.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch tmpl := arg0.(type) {
			case env.String:
				switch vals := arg1.(type) {
				case env.Block:
					result := tmpl.Value
					for _, v := range vals.Series.S {
						result = strings.Replace(result, "{}", v.Print(*ps.Idx), 1)
					}
					return *env.NewString(result)
				default:
					return MakeArgError(ps, 2, []env.Type{env.BlockType}, "format")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "format")
			}
		},
	},

	"prnf": {
		Argsn: 2,
		Doc:   "Formats and prints a value by replacing {} in the template string, without a newline.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch tmpl := arg1.(type) {
			case env.String:
				vals := arg0.Print(*ps.Idx)
				news := strings.ReplaceAll(tmpl.Value, "{}", vals)
				fmt.Print(news)
				return arg0
			default:
				return MakeArgError(ps, 2, []env.Type{env.StringType}, "prnf")
			}
		},
	},

	"printf": {
		Argsn: 2,
		Doc:   "Formats and prints a value by replacing {} in the template string, followed by a newline.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch tmpl := arg1.(type) {
			case env.String:
				vals := arg0.Print(*ps.Idx)
				news := strings.ReplaceAll(tmpl.Value, "{}", vals)
				fmt.Println(news)
				return arg0
			default:
				return MakeArgError(ps, 2, []env.Type{env.StringType}, "printf")
			}
		},
	},

	"embed": {
		Argsn: 2,
		Doc:   "Embeds a value into a string or URI by replacing {} placeholder with the string representation of the value.",
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

	"prnv": {
		Argsn: 2,
		Doc:   "Embeds a value into a string by replacing {} placeholder and prints it without a newline, returning the input value.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch arg := arg1.(type) {
			case env.String:
				vals := arg0.Print(*ps.Idx)
				news := strings.ReplaceAll(arg.Value, "{}", vals)
				fmt.Print(news)
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "prnv")
			}
			return arg0
		},
	},

	"printv": {
		Argsn: 2,
		Doc:   "Embeds a value into a string by replacing {} placeholder and prints it followed by a newline, returning the input value.",
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

	"probe": {
		Argsn: 1,
		Doc:   "Prints detailed type and value information about a value, followed by a newline, returning the input value.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			p := ""
			if env.IsPointer(arg0) {
				p = "REF"
			}
			fmt.Println(p + arg0.Inspect(*ps.Idx))
			return arg0
		},
	},

	"probe\\": {
		Argsn: 2,
		Doc:   "Prints a prefix string followed by detailed type and value information about a value, followed by a newline, returning the input value.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch prefix := arg1.(type) {
			case env.String:
				p := ""
				if env.IsPointer(arg0) {
					p = "REF"
				}
				fmt.Println(prefix.Value + " " + p + arg0.Inspect(*ps.Idx))
				return arg0
			default:
				return MakeArgError(ps, 2, []env.Type{env.StringType}, "probe\\")
			}
		},
	},

	"inspect": {
		Argsn: 1,
		Doc:   "Returns a string containing detailed type and value information about a value.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewString(arg0.Inspect(*ps.Idx))
		},
	},

	"esc": {
		Argsn: 1,
		Doc:   "Creates an ANSI escape sequence by prepending the escape character (\\033) to the input string.",
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
		Doc:   "Creates an ANSI escape sequence with an embedded value by replacing {} placeholder and prepending the escape character (\\033).",
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

	// display, _.. , display\custom and print\ssv / print\csv are omitted
	// under no_baseio because they require the term / util packages.
}
