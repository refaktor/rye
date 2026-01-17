package evaldo

import (
	"encoding/json"

	"github.com/refaktor/rye/env"
	"github.com/refaktor/rye/term"
)

var Builtins_flui = map[string]*env.Builtin{
	"select": {
		Argsn: 1,
		Doc:   "Interactively displays a selection interface for Block, Dict, Table, List, or TableRow. Returns the selected item. Ctrl+C cancels and returns a failure.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			term.SaveCurPos()
			var obj env.Object
			var esc bool
			switch val := arg0.(type) {
			case env.Block:
				obj, esc = term.DisplayBlock(val, ps.Idx)
			case *env.Block:
				obj, esc = term.DisplayBlock(*val, ps.Idx)
			case env.Dict:
				obj, esc = term.DisplayDict(val, ps.Idx)
			case *env.Dict:
				obj, esc = term.DisplayDict(*val, ps.Idx)
			case env.List:
				// Convert list to block for display
				bloc := env.NewBlock(*env.NewTSeries(make([]env.Object, len(val.Data))))
				for i, item := range val.Data {
					bloc.Series.S[i] = env.ToRyeValue(item)
				}
				obj, esc = term.DisplayBlock(*bloc, ps.Idx)
			case *env.List:
				// Convert list to block for display
				bloc := env.NewBlock(*env.NewTSeries(make([]env.Object, len(val.Data))))
				for i, item := range val.Data {
					bloc.Series.S[i] = env.ToRyeValue(item)
				}
				obj, esc = term.DisplayBlock(*bloc, ps.Idx)
			case env.Table:
				obj, esc = term.DisplayTable(val, ps.Idx)
			case *env.Table:
				obj, esc = term.DisplayTable(*val, ps.Idx)
			case env.TableRow:
				obj, esc = term.DisplayTableRow(val, ps.Idx)
			case *env.TableRow:
				obj, esc = term.DisplayTableRow(*val, ps.Idx)
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType, env.DictType, env.ListType, env.TableType}, "select")
			}
			if esc {
				ps.FailureFlag = true
				return env.NewError("canceled by user")
			}
			return obj
		},
	},
	"input": {
		Argsn: 1,
		Doc:   "Interactively displays an input field. Takes the max input width, returns the entered string. Ctrl+C cancels and returns a failure.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			// This is temporary implementation for experimenting what it would work like at all
			// later it should belong to the object (and the medium of display, terminal, html ..., it's part of the frontend)
			term.SaveCurPos()
			switch width := arg0.(type) {
			case env.Integer:
				obj, esc := term.DisplayInputField(0, int(width.Value))
				if esc {
					ps.FailureFlag = true
					return env.NewError("canceled by user")
				}
				return obj
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType}, "input")
			}
		},
	},
	"datepicker": {
		Argsn: 1,
		Doc:   "Interactive date input in format YYYY-MM-DD. Use arrow keys to navigate between year/month/day fields and increment/decrement values, or type digits. Returns date string. Ctrl+C cancels and returns a failure.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			term.SaveCurPos()
			switch initialDate := arg0.(type) {
			case env.String:
				obj, esc := term.DisplayDateInput(initialDate.Value, 0)
				if esc {
					ps.FailureFlag = true
					return env.NewError("canceled by user")
				}
				return obj
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "datepicker")
			}
		},
	},
	"textarea": {
		Argsn: 2,
		Doc:   "Interactive multiline text input. Takes width and height (number of lines). Use arrow keys to navigate, Enter for new line, Ctrl+D to submit. Returns string with newlines. Ctrl+C cancels.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			term.SaveCurPos()
			switch width := arg0.(type) {
			case env.Integer:
				switch height := arg1.(type) {
				case env.Integer:
					obj, esc := term.DisplayTextArea(int(width.Value), int(height.Value), "")
					if esc {
						ps.FailureFlag = true
						return env.NewError("canceled by user")
					}
					return obj
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "textarea")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType}, "textarea")
			}
		},
	},
	"textarea\\": {
		Argsn: 3,
		Doc:   "Interactive multiline text input with initial text. Takes width, height (number of lines), and initial text string. Use arrow keys to navigate, Enter for new line, Ctrl+D to submit. Returns string with newlines. Ctrl+C cancels.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			term.SaveCurPos()
			switch width := arg1.(type) {
			case env.Integer:
				switch height := arg2.(type) {
				case env.Integer:
					switch text := arg0.(type) {
					case env.String:
						obj, esc := term.DisplayTextArea(int(width.Value), int(height.Value), text.Value)
						if esc {
							ps.FailureFlag = true
							return env.NewError("canceled by user")
						}
						return obj
					default:
						return MakeArgError(ps, 3, []env.Type{env.StringType}, "textarea\\")
					}
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "textarea\\")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType}, "textarea\\")
			}
		},
	},
	"displayx": {
		Argsn: 1,
		Doc:   "Returns a special !!! prefixed JSON representation of a Block for Flutter UI rendering. The JSON format is [\"block\", [[\"type\", \"value\"], ...]]. Used for interactive display in Flutter clients.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch val := arg0.(type) {
			case env.Block:
				return blockToDisplayJSON(val, ps.Idx)
			case *env.Block:
				return blockToDisplayJSON(*val, ps.Idx)
			case env.List:
				// Convert list to block for display
				bloc := env.NewBlock(*env.NewTSeries(make([]env.Object, len(val.Data))))
				for i, item := range val.Data {
					bloc.Series.S[i] = env.ToRyeValue(item)
				}
				return blockToDisplayJSON(*bloc, ps.Idx)
			case *env.List:
				// Convert list to block for display
				bloc := env.NewBlock(*env.NewTSeries(make([]env.Object, len(val.Data))))
				for i, item := range val.Data {
					bloc.Series.S[i] = env.ToRyeValue(item)
				}
				return blockToDisplayJSON(*bloc, ps.Idx)
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType, env.ListType}, "displayx")
			}
		},
	},
}

// blockToDisplayJSON converts a Block to a JSON string prefixed with !!! for Flutter UI
func blockToDisplayJSON(block env.Block, idx *env.Idxs) env.Object {
	items := make([][]string, 0, block.Series.Len())
	for i := 0; i < block.Series.Len(); i++ {
		item := block.Series.Get(i)
		typeName := getTypeName(item)
		value := item.Print(*idx)
		items = append(items, []string{typeName, value})
	}
	result := []interface{}{"block", items}
	jsonBytes, err := json.Marshal(result)
	if err != nil {
		return env.NewError("failed to marshal to JSON: " + err.Error())
	}
	return *env.NewString("!!!" + string(jsonBytes))
}

// getTypeName returns the type name of a Rye object as a string
func getTypeName(obj env.Object) string {
	switch obj.(type) {
	case env.String:
		return "string"
	case env.Integer:
		return "integer"
	case env.Decimal:
		return "decimal"
	case env.Word:
		return "word"
	case env.Setword:
		return "setword"
	case env.Getword:
		return "getword"
	case env.Block:
		return "block"
	case env.Dict:
		return "dict"
	case env.List:
		return "list"
	case env.Error:
		return "error"
	default:
		return "unknown"
	}
}
