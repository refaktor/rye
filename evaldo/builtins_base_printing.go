package evaldo

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/refaktor/rye/env"
	"github.com/refaktor/rye/term"
	"golang.org/x/sync/errgroup"

	// JM 20230825	"github.com/refaktor/rye/term"

	"github.com/refaktor/rye/util"
)

// displayRyeValue handles the display of Rye values, supporting both interactive and non-interactive modes
func displayRyeValue(ps *env.ProgramState, arg0 env.Object, interactive bool) (env.Object, string) {
	if interactive {
		// Full interactive mode - use terminal display functions for navigation
		term.SaveCurPos()
		switch bloc := arg0.(type) {
		case env.Block:
			obj, esc := term.DisplayBlock(bloc, ps.Idx)
			if !esc {
				return obj, ""
			}
		case *env.Block:
			obj, esc := term.DisplayBlock(*bloc, ps.Idx)
			if !esc {
				return obj, ""
			}
		case env.Dict:
			obj, esc := term.DisplayDict(bloc, ps.Idx)
			if !esc {
				return obj, ""
			}
		case *env.Dict:
			obj, esc := term.DisplayDict(*bloc, ps.Idx)
			if !esc {
				return obj, ""
			}
		case env.Table:
			obj, esc := term.DisplayTable(bloc, ps.Idx)
			if !esc {
				return obj, ""
			}
		case *env.Table:
			obj, esc := term.DisplayTable(*bloc, ps.Idx)
			if !esc {
				return obj, ""
			}
		case env.TableRow:
			obj, esc := term.DisplayTableRow(bloc, ps.Idx)
			if !esc {
				return obj, ""
			}
		case *env.TableRow:
			obj, esc := term.DisplayTableRow(*bloc, ps.Idx)
			if !esc {
				return obj, ""
			}
		}
	}

	// Non-interactive mode or fallback - return formatted string representation
	// Use enhanced formatting for supported types
	p := ""
	if env.IsPointer(arg0) {
		p = "Ref"
	}

	switch obj := arg0.(type) {
	case env.Block:
		// For blocks, show a more readable format
		if len(obj.Series.GetAll()) <= 5 {
			// Short blocks - show inline
			return arg0, p + obj.Inspect(*ps.Idx)
		} else {
			// Long blocks - show with count
			return arg0, p + fmt.Sprintf("[Block with %d items: %s ... ]", len(obj.Series.GetAll()), obj.Series.GetAll()[0].Inspect(*ps.Idx))
		}
	case *env.Block:
		// For block pointers
		if len(obj.Series.GetAll()) <= 5 {
			return arg0, p + obj.Inspect(*ps.Idx)
		} else {
			return arg0, p + fmt.Sprintf("[Block with %d items: %s ... ]", len(obj.Series.GetAll()), obj.Series.GetAll()[0].Inspect(*ps.Idx))
		}
	case env.Table:
		// For tables, show dimensions and sample
		rows := len(obj.Rows)
		cols := len(obj.Cols)
		return arg0, p + fmt.Sprintf("[Table %dx%d: %v]", rows, cols, obj.Cols)
	case *env.Table:
		rows := len(obj.Rows)
		cols := len(obj.Cols)
		return arg0, p + fmt.Sprintf("[Table %dx%d: %v]", rows, cols, obj.Cols)
	case env.Dict:
		// For dicts, show key count and sample keys
		keys := make([]string, 0)
		for k := range obj.Data {
			keys = append(keys, k)
			if len(keys) >= 3 {
				break
			}
		}
		if len(obj.Data) <= 3 {
			return arg0, p + fmt.Sprintf("[Dict with keys: %v]", keys)
		} else {
			return arg0, p + fmt.Sprintf("[Dict with %d keys: %v ...]", len(obj.Data), keys)
		}
	case *env.Dict:
		keys := make([]string, 0)
		for k := range obj.Data {
			keys = append(keys, k)
			if len(keys) >= 3 {
				break
			}
		}
		if len(obj.Data) <= 3 {
			return arg0, p + fmt.Sprintf("[Dict with keys: %v]", keys)
		} else {
			return arg0, p + fmt.Sprintf("[Dict with %d keys: %v ...]", len(obj.Data), keys)
		}
	default:
		// For other types, use standard inspect
		return arg0, p + obj.Inspect(*ps.Idx)
	}
}

var builtins_printing = map[string]*env.Builtin{

	//
	// ##### Printing ##### "Functions for displaying and formatting values"
	//
	// Tests:
	// stdout { prns "xy" } "xy "
	// Args:
	// * value: Any value to print
	// Returns:
	// * the input value (for chaining)
	"prns": { // **
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

	// Tests:
	// stdout { prn "xy" } "xy"
	// Args:
	// * value: Any value to print
	// Returns:
	// * the input value (for chaining)
	"prn": { // **
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

	// Tests:
	// stdout { print "xy" } "xy\n"
	// Args:
	// * value: Any value to print
	// Returns:
	// * the input value (for chaining)
	"print": { // **
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

	// Tests:
	// stdout { print2 "hello" "world" } "hello world\n"
	// Args:
	// * value1: First value to print
	// * value2: Second value to print
	// Returns:
	// * the second input value
	"print2": { // **
		Argsn: 2,
		Doc:   "Prints two values separated by a space and followed by a newline, returning the second value.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch arg0 := arg0.(type) {
			case env.String:
				fmt.Print(arg0.Value)
			default:
				fmt.Print(arg0.Print(*ps.Idx))
			}
			fmt.Print(" ")
			switch arg1 := arg1.(type) {
			case env.String:
				fmt.Println(arg1.Value)
			default:
				fmt.Println(arg1.Print(*ps.Idx))
			}
			return arg1
		},
	},

	// Tests:
	// stdout { prn2 "hello" "world" } "hello world"
	// Args:
	// * value1: First value to print
	// * value2: Second value to print
	// Returns:
	// * the second input value
	"prn2": { // **
		Argsn: 2,
		Doc:   "Prints two values separated by a space without adding a newline, returning the second value.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch arg0 := arg0.(type) {
			case env.String:
				fmt.Print(arg0.Value)
			default:
				fmt.Print(arg0.Print(*ps.Idx))
			}
			fmt.Print(" ")
			switch arg1 := arg1.(type) {
			case env.String:
				fmt.Print(arg1.Value)
			default:
				fmt.Print(arg1.Print(*ps.Idx))
			}
			return arg1
		},
	},

	// Tests:
	// stdout { prns2 "hello" "world" } "hello world "
	// Args:
	// * value1: First value to print
	// * value2: Second value to print
	// Returns:
	// * the second input value
	"prns2": { // **
		Argsn: 2,
		Doc:   "Prints two values separated by a space and followed by a space, returning the second value.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch arg0 := arg0.(type) {
			case env.String:
				fmt.Print(arg0.Value)
			default:
				fmt.Print(arg0.Print(*ps.Idx))
			}
			fmt.Print(" ")
			switch arg1 := arg1.(type) {
			case env.String:
				fmt.Print(arg1.Value + " ")
			default:
				fmt.Print(arg1.Print(*ps.Idx) + " ")
			}
			return arg1
		},
	},

	// Tests:
	// equal { format 123 "num: %d" } "num: 123"
	// equal { format "hello" "%s world" } "hello world"
	// Args:
	// * value: Value to format (string, integer, or decimal)
	// * format: String containing Go's sprintf format specifiers
	// Returns:
	// * formatted string
	"format": {
		Argsn: 2,
		Doc:   "Formats a value according to Go's sprintf format specifiers, returning the formatted string.",
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
				case env.Block:
					series := val.Series
					vals := make([]interface{}, series.Len())
					for i, obj := range series.GetAll() {
						switch v := obj.(type) {
						case env.Integer:
							vals[i] = v.Value
						case env.String:
							vals[i] = v.Value
						case env.Decimal:
							vals[i] = v.Value
						default:
							vals[i] = v.Print(*ps.Idx)
						}
					}
					res = fmt.Sprintf(arg.Value, vals...)
				default:
					return MakeArgError(ps, 1, []env.Type{env.StringType, env.DecimalType, env.IntegerType, env.BlockType}, "format")
				}
				return *env.NewString(res)
			default:
				return MakeArgError(ps, 2, []env.Type{env.StringType}, "format")
			}
		},
	},

	// Tests:
	// stdout { prnf 123 "num: %d" } "num: 123"
	// Args:
	// * value: Value to format (string, integer, or decimal)
	// * format: String containing Go's sprintf format specifiers
	// Returns:
	// * the input value (for chaining)
	"prnf": { // **
		Argsn: 2,
		Doc:   "Formats a value according to Go's sprintf format specifiers and prints it without a newline, returning the input value.",
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
	// equal { embed 101 "val {}" } "val 101"
	// equal { embed "world" "hello {}" } "hello world"
	// Args:
	// * value: Value to embed
	// * template: String or URI containing {} as a placeholder
	// Returns:
	// * string or URI with the placeholder replaced by the value
	"embed": { // **
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

	// Tests:
	// stdout { prnv 101 "val {}" } "val 101"
	// Args:
	// * value: Value to embed
	// * template: String containing {} as a placeholder
	// Returns:
	// * the input value (for chaining)
	"prnv": { // **
		Argsn: 2,
		Doc:   "Embeds a value into a string by replacing {} placeholder and prints it without a newline, returning the input value.",
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
	// stdout { printv 101 "val {}" } "val 101\n"
	// Args:
	// * value: Value to embed
	// * template: String containing {} as a placeholder
	// Returns:
	// * the input value (for chaining)
	"printv": { // **
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

	// Tests:
	// stdout { print\ssv { 101 "asd" } } "101 asd\n"
	// Args:
	// * block: Block of values to format as space-separated values
	// Returns:
	// * the input block (for chaining)
	"print\\ssv": {
		Argsn: 1,
		Doc:   "Prints a block of values as space-separated values followed by a newline, returning the input block.",
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
	// stdout { print\csv { 101 "asd" } } "101,asd\n"
	// Args:
	// * block: Block of values to format as comma-separated values
	// Returns:
	// * the input block (for chaining)
	"print\\csv": { //
		Argsn: 1,
		Doc:   "Prints a block of values as comma-separated values followed by a newline, returning the input block.",
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
	// stdout { probe 101 } "[Integer: 101]\n"
	// Args:
	// * value: Any value to inspect
	// Returns:
	// * the input value (for chaining)
	"probe": { // **
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

	// Tests:
	// stdout { probe\ 101 "value:" } "value: [Integer: 101]\n"
	// Args:
	// * value: Any value to inspect
	// * prefix: String to print before the probed value
	// Returns:
	// * the input value (for chaining)
	"probe\\": { // **
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

	// Tests:
	// equal { inspect 101 } "[Integer: 101]"
	// Args:
	// * value: Any value to inspect
	// Returns:
	// * string containing detailed type and value information
	"inspect": { // **
		Argsn: 1,
		Doc:   "Returns a string containing detailed type and value information about a value.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewString(arg0.Inspect(*ps.Idx))
		},
	},

	// Tests:
	// ; equal { esc "[33m" } "\033[33m"   ; we can't represent hex or octal in strings yet
	// Args:
	// * sequence: String to append to the escape character
	// Returns:
	// * string with escape sequence
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

	// Tests:
	// ; equal { esc-val "[33m" "Error" } "\033[33mError"  ; we can't represent hex or octal in strings yet
	// Args:
	// * value: Value to embed
	// * template: String containing {} as a placeholder
	// Returns:
	// * string with escape sequence and embedded value
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
	// Tests:
	// ; display { 1 2 3 }  ; interactive display, can't test in automated tests
	// Args:
	// * value: Block, Dict, Table, or TableRow to display interactively
	// Returns:
	// * the input value or selected item from interactive display
	"display": {
		Pure:  true,
		Argsn: 1,
		Doc:   "Interactively displays a value (Block, Dict, Table, or TableRow) in the terminal with navigation capabilities.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			result, _ := displayRyeValue(ps, arg0, true)
			return result
		},
	},
	// Tests:
	// ; _.. { 1 2 3 }  ; interactive display, can't test in automated tests
	// Args:
	// * value: Block, Dict, Table, or TableRow to display interactively
	// Returns:
	// * the input value or selected item from interactive display
	"_..": {
		Argsn: 1,
		Doc:   "Shorthand alias for 'display' - interactively displays a value in the terminal with navigation capabilities.",
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
	// Tests:
	// ; display\custom my-table { |row curr| ... }  ; interactive display with custom renderer
	// Args:
	// * value: Table to display interactively
	// * renderer: Function that takes a row and current position indicator
	// Returns:
	// * the input value or selected item from interactive display
	"display\\custom": {
		Argsn: 2,
		Doc:   "Interactively displays a Table in the terminal with a custom rendering function for each row.",
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

	// Tests:
	// equal { capture-stdout { print "hello" } } "hello\n"
	// equal { capture-stdout { loop 3 { prns "x" } } } "x x x "
	// Args:
	// * block: Block of code to execute with captured stdout
	// Returns:
	// * string containing all output captured during block execution
	"capture-stdout": { // **
		Argsn: 1,
		Doc:   "Executes a block of code while capturing all output to stdout, returning the captured output as a string.",
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
				MaybeDisplayFailureOrError(ps, ps.Idx, "capture-stdout")
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
}
