package baseio

import (
	"fmt"

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

	"display": {
		Pure:  true,
		Argsn: 1,
		Doc:   "Interactively displays a value (Block, Dict, Table, TableRow, or Markdown) in the terminal with navigation capabilities.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			result, _ := DisplayRyeValue(ps, arg0, true)
			return result
		},
	},

	"_..": {
		Argsn: 1,
		Doc:   "Shorthand alias for 'display' - interactively displays a value in the terminal with navigation capabilities.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
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
			case *env.Error:
				obj, esc := term.DisplayError(bloc, ps.Idx)
				if !esc {
					return obj
				}
			case env.Error:
				obj, esc := term.DisplayError(&bloc, ps.Idx)
				if !esc {
					return obj
				}
			}
			return arg0
		},
	},

	"display\\custom": {
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
