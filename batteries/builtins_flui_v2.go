package batteries

// builtins_flui_v2.go - Refactored TUI builtins using widget abstraction
// This file can replace builtins_flui.go once validated

import (
	"encoding/json"
	"os"

	goterm "golang.org/x/term"

	"github.com/refaktor/rye/env"
	"github.com/refaktor/rye/evaldo"
	"github.com/refaktor/rye/term"
)

var Builtins_flui_v2 = map[string]*env.Builtin{

	// =========================================================================
	// select - Interactive selection from Block, Dict, Table, List, TableRow
	// =========================================================================
	"select": {
		Argsn: 1,
		Doc:   "Interactively displays a selection interface. Returns the selected item. Ctrl+C cancels.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			items, ok := extractItems(arg0, ps.Idx)
			if !ok {
				return evaldo.MakeArgError(ps, 1, []env.Type{env.BlockType, env.DictType, env.ListType, env.TableType}, "select")
			}

			// Determine if we need pagination
			_, height, termErr := goterm.GetSize(int(os.Stdout.Fd()))
			if termErr != nil {
				height = 20
			}
			pageSize := height - 4

			term.SaveCurPos()

			var widget term.Widget
			if len(items) <= pageSize {
				widget = term.NewSelectWidget(items, ps.Idx)
			} else {
				widget = term.NewPaginatedSelectWidget(items, pageSize, ps.Idx)
			}

			result, canceled := term.RunWidget(widget)
			if canceled {
				ps.FailureFlag = true
				return env.NewError("canceled by user")
			}
			return result
		},
	},

	// =========================================================================
	// input - Single line text input
	// =========================================================================
	"input": {
		Argsn: 1,
		Doc:   "Interactively displays an input field. Takes max width, returns entered string. Ctrl+C cancels.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch width := arg0.(type) {
			case env.Integer:
				term.SaveCurPos()
				widget := term.NewInputWidget(int(width.Value), ps.Idx)
				result, canceled := term.RunWidget(widget)
				if canceled {
					ps.FailureFlag = true
					return env.NewError("canceled by user")
				}
				return result
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.IntegerType}, "input")
			}
		},
	},

	// =========================================================================
	// datepicker - Interactive date input (YYYY-MM-DD)
	// =========================================================================
	"datepicker": {
		Argsn: 1,
		Doc:   "Interactive date input. Use arrow keys to navigate/change values, or type digits. Ctrl+C cancels.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch initialDate := arg0.(type) {
			case env.String:
				term.SaveCurPos()
				widget := term.NewDateWidget(initialDate.Value, ps.Idx)
				result, canceled := term.RunWidget(widget)
				if canceled {
					ps.FailureFlag = true
					return env.NewError("canceled by user")
				}
				return result
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.StringType}, "datepicker")
			}
		},
	},

	// =========================================================================
	// textarea - Multiline text input
	// =========================================================================
	"textarea": {
		Argsn: 2,
		Doc:   "Interactive multiline text input. Takes width and height. Ctrl+D to submit, Ctrl+C cancels.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch width := arg0.(type) {
			case env.Integer:
				switch height := arg1.(type) {
				case env.Integer:
					term.SaveCurPos()
					widget := term.NewTextAreaWidget(int(width.Value), int(height.Value), "", ps.Idx)
					result, canceled := term.RunWidget(widget)
					if canceled {
						ps.FailureFlag = true
						return env.NewError("canceled by user")
					}
					return result
				default:
					return evaldo.MakeArgError(ps, 2, []env.Type{env.IntegerType}, "textarea")
				}
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.IntegerType}, "textarea")
			}
		},
	},

	// =========================================================================
	// textarea\ - Multiline text input with initial text (pipe-friendly)
	// =========================================================================
	"textarea\\": {
		Argsn: 3,
		Doc:   "Interactive multiline text input with initial text. Ctrl+D to submit, Ctrl+C cancels.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch text := arg0.(type) {
			case env.String:
				switch width := arg1.(type) {
				case env.Integer:
					switch height := arg2.(type) {
					case env.Integer:
						term.SaveCurPos()
						widget := term.NewTextAreaWidget(int(width.Value), int(height.Value), text.Value, ps.Idx)
						result, canceled := term.RunWidget(widget)
						if canceled {
							ps.FailureFlag = true
							return env.NewError("canceled by user")
						}
						return result
					default:
						return evaldo.MakeArgError(ps, 3, []env.Type{env.IntegerType}, "textarea\\")
					}
				default:
					return evaldo.MakeArgError(ps, 2, []env.Type{env.IntegerType}, "textarea\\")
				}
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.StringType}, "textarea\\")
			}
		},
	},

	// =========================================================================
	// displayx - JSON representation for Flutter UI
	// =========================================================================
	"displayx": {
		Argsn: 1,
		Doc:   "Returns a !!! prefixed JSON representation for Flutter UI rendering.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch val := arg0.(type) {
			case env.Block:
				return blockToDisplayJSON(val, ps.Idx)
			case *env.Block:
				return blockToDisplayJSON(*val, ps.Idx)
			case env.List:
				return blockToDisplayJSON(*listToBlock(val), ps.Idx)
			case *env.List:
				return blockToDisplayJSON(*listToBlock(*val), ps.Idx)
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.BlockType, env.ListType}, "displayx")
			}
		},
	},
}

// extractItems converts various Rye types to a slice of Objects for selection
func extractItems(arg env.Object, idx *env.Idxs) ([]env.Object, bool) {
	switch val := arg.(type) {
	case env.Block:
		return val.Series.S, true
	case *env.Block:
		return val.Series.S, true
	case env.List:
		return listToObjects(val), true
	case *env.List:
		return listToObjects(*val), true
	case env.Dict:
		return dictToObjects(val), true
	case *env.Dict:
		return dictToObjects(*val), true
	// Note: Table and TableRow would need special handling
	// to preserve their row structure
	}
	return nil, false
}

func listToObjects(list env.List) []env.Object {
	result := make([]env.Object, len(list.Data))
	for i, item := range list.Data {
		result[i] = env.ToRyeValue(item)
	}
	return result
}

func dictToObjects(dict env.Dict) []env.Object {
	result := make([]env.Object, 0, len(dict.Data))
	for k, v := range dict.Data {
		// Create a representation that shows key: value
		if obj, ok := v.(env.Object); ok {
			result = append(result, env.NewString(k+": "+obj.Print(env.Idxs{})))
		}
	}
	return result
}

func listToBlock(list env.List) *env.Block {
	bloc := env.NewBlock(*env.NewTSeries(make([]env.Object, len(list.Data))))
	for i, item := range list.Data {
		bloc.Series.S[i] = env.ToRyeValue(item)
	}
	return bloc
}

// blockToDisplayJSON converts a Block to a JSON string prefixed with !!! for Flutter UI
func blockToDisplayJSON_v2(block env.Block, idx *env.Idxs) env.Object {
	items := make([][]string, 0, block.Series.Len())
	for i := 0; i < block.Series.Len(); i++ {
		item := block.Series.Get(i)
		typeName := getTypeName_v2(item)
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
func getTypeName_v2(obj env.Object) string {
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
