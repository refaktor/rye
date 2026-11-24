package term

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/refaktor/rye/util"

	"github.com/refaktor/rye/env"
)

const (
	// DisplayMaxValueWidth is the longest compact value shown in a table cell.
	// Values longer than this are truncated with an ellipsis.
	DisplayMaxValueWidth = 52

	// displayValueMaxItems keeps nested collections readable: show a few items,
	// then an ellipsis. The same rule is used by tables and by display/explore.
	displayValueMaxItems = 5
)

// DisplayValueText formats a cell or entry in human (Print) or developer
// (Inspect) mode. Collections are summarized so nested blocks, lists and dicts
// stay readable; both static output and the interactive views use it.
func DisplayValueText(value any, idx *env.Idxs, developer bool) string {
	if obj, ok := value.(env.Object); ok {
		switch v := obj.(type) {
		case env.Block:
			return util.TruncateString(displayBlockItems(v.Series.GetAll(), idx, developer), DisplayMaxValueWidth)
		case *env.Block:
			return util.TruncateString(displayBlockItems(v.Series.GetAll(), idx, developer), DisplayMaxValueWidth)
		case env.List:
			return util.TruncateString(displayListItems(v.Data, idx, developer), DisplayMaxValueWidth)
		case *env.List:
			return util.TruncateString(displayListItems(v.Data, idx, developer), DisplayMaxValueWidth)
		case env.Dict:
			return util.TruncateString(displayDictItems(v, idx, developer), DisplayMaxValueWidth)
		case *env.Dict:
			return util.TruncateString(displayDictItems(*v, idx, developer), DisplayMaxValueWidth)
		case env.Bunch:
			return util.TruncateString(displayBunchItems(v), DisplayMaxValueWidth)
		case *env.Bunch:
			return util.TruncateString(displayBunchItems(*v), DisplayMaxValueWidth)
		case env.Table:
			return v.Inspect(*idx)
		case *env.Table:
			return v.Inspect(*idx)
		}
		if developer {
			return obj.Inspect(*idx)
		}
		return obj.Print(*idx)
	}
	return fmt.Sprint(value)
}

// CompactValueText is DisplayValueText capped to the width used by table cells.
func CompactValueText(value any, idx *env.Idxs, developer bool) string {
	return util.TruncateString(DisplayValueText(value, idx, developer), DisplayMaxValueWidth)
}

func displayBlockItems(items []env.Object, idx *env.Idxs, developer bool) string {
	parts, truncated := displayItemParts(items, idx, developer)
	return joinDisplayParts(parts, truncated, "", "")
}

func displayListItems(items []any, idx *env.Idxs, developer bool) string {
	parts, truncated := displayAnyItemParts(items, idx, developer)
	return joinDisplayParts(parts, truncated, "L[", " ]")
}

func displayDictItems(dict env.Dict, idx *env.Idxs, developer bool) string {
	keys := make([]string, 0, len(dict.Data))
	for key := range dict.Data {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	limit := min(len(keys), displayValueMaxItems)
	parts := make([]string, 0, limit)
	for _, key := range keys[:limit] {
		parts = append(parts, key+": "+displayItemText(dict.Data[key], idx, developer))
	}
	return joinDisplayParts(parts, len(keys) > limit, "[", " ]")
}

func displayBunchItems(bunch env.Bunch) string {
	keys := bunch.SortedKeys()
	limit := min(len(keys), displayValueMaxItems)
	parts := make([]string, 0, limit)
	for _, key := range keys[:limit] {
		parts = append(parts, fmt.Sprintf("%q", key))
	}
	return joinDisplayParts(parts, len(keys) > limit, "Bunch[", " ]")
}

func displayItemParts(items []env.Object, idx *env.Idxs, developer bool) ([]string, bool) {
	limit := min(len(items), displayValueMaxItems)
	parts := make([]string, 0, limit)
	for _, item := range items[:limit] {
		if item == nil {
			parts = append(parts, "[NIL]")
			continue
		}
		parts = append(parts, displayItemText(item, idx, developer))
	}
	return parts, len(items) > limit
}

func displayAnyItemParts(items []any, idx *env.Idxs, developer bool) ([]string, bool) {
	limit := min(len(items), displayValueMaxItems)
	parts := make([]string, 0, limit)
	for _, item := range items[:limit] {
		parts = append(parts, displayItemText(item, idx, developer))
	}
	return parts, len(items) > limit
}

func displayItemText(value any, idx *env.Idxs, developer bool) string {
	// Recurse so nested collections get the same item limit.
	switch v := value.(type) {
	case map[string]any:
		return displayDictItems(*env.NewDict(v), idx, developer)
	case env.Block, *env.Block, env.List, *env.List, env.Dict, *env.Dict, env.Bunch, *env.Bunch, env.Table, *env.Table:
		return DisplayValueText(value, idx, developer)
	}
	if obj, ok := value.(env.Object); ok {
		if developer {
			return obj.Inspect(*idx)
		}
		return obj.Print(*idx)
	}
	return fmt.Sprint(value)
}

func joinDisplayParts(parts []string, truncated bool, open string, close string) string {
	text := strings.Join(parts, " ")
	if truncated {
		if text != "" {
			text += " "
		}
		text += "..."
	}
	if open == "" && close == "" {
		return text
	}
	if text == "" {
		return open + close
	}
	return open + " " + text + close
}

// TableDisplayWidths lays out static and interactive tables consistently.
// Interactive tables show compact values capped at DisplayMaxValueWidth; static
// output may expand complete scalar values, but collections remain summarized.
func TableDisplayWidths(table env.Table, idx *env.Idxs, full bool) []int {
	widths := make([]int, len(table.Cols))
	for i, col := range table.Cols {
		widths[i] = utf8.RuneCountInString(col) + 1
	}
	for _, row := range table.Rows {
		for i, value := range row.Values {
			if i >= len(widths) {
				continue
			}
			text := DisplayValueText(value, idx, false)
			if !full {
				text = util.TruncateString(text, DisplayMaxValueWidth)
			}
			width := utf8.RuneCountInString(text) + 1
			if widths[i] < width {
				widths[i] = width
			}
		}
	}
	return widths
}

// RenderTableHeader writes the column names and separator without terminal
// cursor effects, so both display and explore can use the same layout.
func RenderTableHeader(w io.Writer, columns []string, widths []int) {
	for i, name := range columns {
		fmt.Fprintf(w, "| %-*s", widths[i], name)
	}
	fmt.Fprintln(w, "|")
	fullWidth := 0
	for _, width := range widths {
		fullWidth += width + 2
	}
	fmt.Fprintln(w, "+"+strings.Repeat("-", max(0, fullWidth-1))+"+")
}

// RenderTableCells writes one row with the same widths and value formatting
// used by explore. Interactive mode adds its highlighting outside this helper.
func RenderTableCells(w io.Writer, values []any, widths []int, idx *env.Idxs, developer bool) {
	for i, value := range values {
		if i >= len(widths) {
			break
		}
		text := DisplayValueText(value, idx, developer)
		if !developer {
			text = util.TruncateString(text, widths[i])
		}
		fmt.Fprintf(w, "| %-*s", widths[i], text)
	}
	fmt.Fprintln(w, "|")
}

// RenderMarkdownItems prints every preformatted Markdown section without
// selecting or paging; the interactive view uses the same DisplayLines.
func RenderMarkdownItems(w io.Writer, items []interface{}) {
	for i, item := range items {
		if section, ok := item.(map[string]interface{}); ok {
			if lines, ok := section["DisplayLines"].([]string); ok {
				if i > 0 {
					fmt.Fprintln(w)
				}
				for _, line := range lines {
					fmt.Fprintln(w, line)
				}
			}
		}
	}
}

// RenderValue writes every entry without cursor control, pagination or input.
func RenderValue(w io.Writer, value env.Object, idx *env.Idxs) {
	switch v := value.(type) {
	case env.Block:
		for _, item := range v.Series.GetAll() {
			fmt.Fprintln(w, DisplayValueText(item, idx, false))
		}
	case *env.Block:
		RenderValue(w, *v, idx)
	case env.Dict:
		keys := make([]string, 0, len(v.Data))
		for key := range v.Data {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			fmt.Fprintf(w, "%s: %s\n", key, DisplayValueText(v.Data[key], idx, false))
		}
	case *env.Dict:
		RenderValue(w, *v, idx)
	case env.Table:
		widths := TableDisplayWidths(v, idx, true)
		RenderTableHeader(w, v.Cols, widths)
		for _, row := range v.Rows {
			RenderTableCells(w, row.Values, widths, idx, false)
		}
	case *env.Table:
		RenderValue(w, *v, idx)
	case env.TableRow:
		for i, key := range v.Uplink.GetColumnNames() {
			if i < len(v.Values) {
				fmt.Fprintf(w, "%s: %s\n", key, DisplayValueText(v.Values[i], idx, false))
			}
		}
	case *env.TableRow:
		RenderValue(w, *v, idx)
	case *env.Error:
		if idx.GetWord(v.Kind.Index) == "warning" {
			ColorRed()
		} else if idx.GetWord(v.Kind.Index) == "note" {
			ColorMagenta()
		} else {
			ColorBrRed()
		}
		fmt.Println(v.Print2(*idx, 3, true))
		CloseProps()
	default:
		fmt.Fprintln(w, DisplayValueText(value, idx, false))
	}
}
