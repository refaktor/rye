package term

import (
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"

	"github.com/refaktor/rye/util"

	"github.com/refaktor/rye/env"
)

// DisplayValueText formats a cell or entry in human (Print) or developer
// (Inspect) mode. Both static output and the interactive views use it.
func DisplayValueText(value any, idx *env.Idxs, developer bool) string {
	if obj, ok := value.(env.Object); ok {
		if developer {
			return obj.Inspect(*idx)
		}
		return obj.Print(*idx)
	}
	return fmt.Sprint(value)
}

// TableDisplayWidths lays out static and interactive tables consistently.
// The interactive view retains its historical text width limit; static output
// expands columns to show complete values.
func TableDisplayWidths(table env.Table, idx *env.Idxs, full bool) []int {
	widths := make([]int, len(table.Cols))
	for i, col := range table.Cols {
		widths[i] = len(col) + 1
	}
	for _, row := range table.Rows {
		for i, value := range row.Values {
			if i >= len(widths) {
				continue
			}
			width := 5
			switch v := value.(type) {
			case string:
				width = len(v) + 2
				if width > 52 {
					width = 52
				}
			case int64:
				width = len(strconv.FormatInt(v, 10)) + 1
			case env.Integer:
				width = len(strconv.FormatInt(v.Value, 10)) + 1
			case float64:
				width = len(strconv.FormatFloat(v, 'f', 2, 64)) + 1
			case env.Decimal:
				width = len(strconv.FormatFloat(v.Value, 'f', 2, 64)) + 1
			case env.String, env.Vector:
				width = len(DisplayValueText(v, idx, false))
				if width > 52 {
					width = 52
				}
			}
			if full {
				width = max(width, len(DisplayValueText(value, idx, false)))
			}
			if widths[i] < width {
				widths[i] = width + 1
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
	default:
		fmt.Fprintln(w, DisplayValueText(value, idx, false))
	}
}
