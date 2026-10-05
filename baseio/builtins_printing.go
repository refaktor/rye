package baseio

import (
	"fmt"
	"os"

	"github.com/refaktor/rye/env"
	"github.com/refaktor/rye/evaldo"
	"github.com/refaktor/rye/term"
	"github.com/refaktor/rye/util"
)

// DisplayRyeValue handles the display of Rye values, supporting both interactive and non-interactive modes.
// Exported so the console package can reference it via baseio.DisplayRyeValue.
func DisplayRyeValue(ps *env.ProgramState, arg0 env.Object, interactive bool) (env.Object, string) {
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
		case env.Markdown:
			items := evaldo.BatteryMarkdownDisplayHook(bloc.Value)
			if len(items) == 0 {
				return bloc, ""
			}
			obj, esc := term.DisplayMarkdownItems(items, ps.Idx)
			if !esc {
				return obj, ""
			}
		case *env.Markdown:
			items := evaldo.BatteryMarkdownDisplayHook(bloc.Value)
			if len(items) == 0 {
				return bloc, ""
			}
			obj, esc := term.DisplayMarkdownItems(items, ps.Idx)
			if !esc {
				return obj, ""
			}
		case *env.Error:
			obj, esc := term.DisplayError(bloc, ps.Idx)
			if !esc {
				return obj, ""
			}
		case env.Error:
			obj, esc := term.DisplayError(&bloc, ps.Idx)
			if !esc {
				return obj, ""
			}
		}
	}

	// Non-interactive mode or fallback - return formatted string representation
	p := ""
	if env.IsPointer(arg0) {
		p = "Ref"
	}

	switch obj := arg0.(type) {
	case env.Block:
		if len(obj.Series.GetAll()) <= 5 {
			return arg0, p + obj.Inspect(*ps.Idx)
		} else {
			return arg0, p + fmt.Sprintf("[Block with %d items: %s ... ]", len(obj.Series.GetAll()), obj.Series.GetAll()[0].Inspect(*ps.Idx))
		}
	case *env.Block:
		if len(obj.Series.GetAll()) <= 5 {
			return arg0, p + obj.Inspect(*ps.Idx)
		} else {
			return arg0, p + fmt.Sprintf("[Block with %d items: %s ... ]", len(obj.Series.GetAll()), obj.Series.GetAll()[0].Inspect(*ps.Idx))
		}
	case env.Table:
		rows := len(obj.Rows)
		cols := len(obj.Cols)
		return arg0, p + fmt.Sprintf("[Table %dx%d: %v]", rows, cols, obj.Cols)
	case *env.Table:
		rows := len(obj.Rows)
		cols := len(obj.Cols)
		return arg0, p + fmt.Sprintf("[Table %dx%d: %v]", rows, cols, obj.Cols)
	case env.Dict:
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
		return arg0, p + obj.Inspect(*ps.Idx)
	}
}

// builtins_printing_extra contains only the printing builtins that
// require the term / util packages (interactive display, CSV/SSV output).
// The basic printing builtins (prns, print, probe, inspect, etc.) are
// registered by evaldo.RegisterBaseBuiltins and do not require these deps.
var builtins_printing_extra = map[string]*env.Builtin{

	// Prints the entire value without input or pagination and passes it through.
	"display": {
		Argsn: 1,
		Doc:   "Displays all entries without interaction or pagination and returns the original value.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			term.RenderValue(os.Stdout, arg0, ps.Idx)
			return arg0
		},
	},

	// Example: [1 2 3] |explore  ; choose an item with the arrow keys
	// Selection and cancellation behave like the former display builtin.
	"explore": {
		Argsn: 1,
		Doc:   "Interactively explores a value; returns the selected item or the original value when cancelled.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			result, _ := DisplayRyeValue(ps, arg0, true)
			return result
		},
	},

	// Example:
	// _.. [1 2 3]
	// Args:
	// * value: Block, Dict, Table, TableRow, Markdown, or Error to display interactively
	// Returns:
	// * the selected value or the original value when user exits
	"_..": {
		Argsn: 1,
		Doc:   "Shorthand alias for explore: interactively select a value.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			result, _ := DisplayRyeValue(ps, arg0, true)
			return result
		},
	},

	// Example:
	// table { "n" } { 1 2 3 } |explore\custom fn { row is-curr } { if is-curr > 0 { print "*" } print row }
	// Args:
	// * table: Table to display
	// * renderer: Function called for each row with args (row is-current)
	// Returns:
	// * the selected row or original table when user exits
	"explore\\custom": {
		Argsn: 2,
		Doc:   "Interactively displays a Table in the terminal with a custom rendering function for each row.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			term.SaveCurPos()
			switch fnc := arg1.(type) {
			case env.Function:
				switch bloc := arg0.(type) {
				case env.Table:
					obj, esc := term.DisplayTableCustom(
						bloc,
						func(row env.Object, iscurr env.Integer) { evaldo.CallFunctionArgsN(fnc, ps, ps.Ctx, row, iscurr) },
						ps.Idx)
					if !esc {
						return obj
					}
				case *env.Table:
					obj, esc := term.DisplayTableCustom(
						*bloc,
						func(row env.Object, iscurr env.Integer) { evaldo.CallFunctionArgsN(fnc, ps, ps.Ctx, row, iscurr) },
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
	// stdout { print\ssv { 1 2 "a" } } "1 2 a\n"
	// Args:
	// * values: Block to format as space-separated values
	// Returns:
	// * returns the input block after printing
	"print\\ssv": {
		Argsn: 1,
		Doc:   "Prints a block of values as space-separated values followed by a newline, returning the input block.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch arg := arg0.(type) {
			case env.Object:
				fmt.Println(util.FormatSsv(arg, *ps.Idx))
			default:
				return evaldo.MakeBuiltinError(ps, "Not Rye object.", "print-ssv")
			}
			return arg0
		},
	},

	// Tests:
	// stdout { print\csv { 1 2 "a" } } "1,2,a\n"
	// Args:
	// * values: Block to format as comma-separated values
	// Returns:
	// * returns the input block after printing
	"print\\csv": {
		Argsn: 1,
		Doc:   "Prints a block of values as comma-separated values followed by a newline, returning the input block.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch arg := arg0.(type) {
			case env.Object:
				fmt.Println(util.FormatCsv(arg, *ps.Idx))
			default:
				return evaldo.MakeBuiltinError(ps, "Not Rye object.", "print-csv")
			}
			return arg0
		},
	},
}
