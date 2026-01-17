package term

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	goterm "golang.org/x/term"

	"github.com/refaktor/rye/env"
	"github.com/refaktor/rye/util"
)

// termPrint is a function that abstracts the terminal output
// In Unix/Windows, it uses fmt.Print directly
// In WASM, it uses the sendBack function
var termPrint func(string)
var termPrintln func(string)
var termPrintf func(string, ...interface{})

func init() {
	// Default to fmt.Print for Unix/Windows
	// This will be overridden in WASM by SetSB
	termPrint = func(s string) {
		fmt.Print(s)
	}
	termPrintln = func(s string) {
		fmt.Println(s)
	}
	termPrintf = func(format string, args ...interface{}) {
		fmt.Printf(format, args...)
	}
}

func DisplayBlock(bloc env.Block, idx *env.Idxs) (env.Object, bool) {
	mode := 0 // 0 - human, 1 - dev
	totalItems := bloc.Series.Len()
	_, height, err := goterm.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		height = 20 // Fallback default
	}
	pageSize := height - 4 // Reserve lines for prompts/instructions
	if pageSize < 1 {
		pageSize = 1
	}
	totalPages := (totalItems + pageSize - 1) / pageSize // Ceiling division
	if totalPages == 0 {
		totalPages = 1
	}

	// If totalPages <= 1, use inline interactive mode
	if totalPages <= 1 {
		HideCur()
		curr := 0
		moveUp := 0

		defer func() {
			ShowCur()
		}()

	INLINE_DODO:
		if moveUp > 0 {
			CurUp(moveUp)
		}
		SaveCurPos()

		totalLines := 0
		// Print all items with cursor highlighting
		for i, v := range bloc.Series.S {
			ClearLine()
			if i == curr {
				ColorBrGreen()
				Bold()
				termPrint("\u00bb ")
			} else {
				termPrint(" ")
			}
			var valueStr string
			switch ob := v.(type) {
			case env.Object:
				if mode == 0 {
					valueStr = ob.Print(*idx)
				} else {
					valueStr = ob.Inspect(*idx)
				}
			default:
				valueStr = fmt.Sprint(ob)
			}
			termPrintln(valueStr)
			// Count the actual number of lines this entry takes (including newlines in the value)
			totalLines += strings.Count(valueStr, "\n") + 1
			CloseProps()
		}

		moveUp = totalLines

		for {
			ascii, keyCode, err := GetChar()

			if (ascii == 3 || ascii == 27) || err != nil {
				return bloc, true // Return full block on Ctrl+C or Esc
			}

			if ascii == 13 {
				if curr < totalItems {
					return bloc.Series.Get(curr), false // Return selected item on Enter
				}
				return nil, true
			}

			if ascii == 77 || ascii == 109 { // 'm' or 'M' for mode toggle
				mode = 1 - mode
				goto INLINE_DODO
			}

			if keyCode == 40 { // Down arrow
				curr++
				if curr >= totalItems {
					curr = 0 // Wrap to top
				}
				goto INLINE_DODO
			} else if keyCode == 38 { // Up arrow
				curr--
				if curr < 0 {
					curr = totalItems - 1 // Wrap to bottom
				}
				goto INLINE_DODO
			}
		}
	}

	// Full-screen paginated mode for blocks that need pagination
	HideCur()
	currentPage := 0
	localCurr := 0
	moveUp := 0

DODO:
	if moveUp > 0 {
		CurUp(moveUp)
	}
	SaveCurPos()
	start := currentPage * pageSize
	end := start + pageSize
	if end > totalItems {
		end = totalItems
	}
	displayedItems := bloc.Series.S[start:end]
	displayLen := len(displayedItems)
	totalLines := 0
	for i := 0; i < pageSize; i++ {
		ClearLine()
		if i < displayLen {
			v := displayedItems[i]
			if i == localCurr {
				ColorBrGreen()
				Bold()
				termPrint("\u00bb ")
			} else {
				termPrint(" ")
			}
			var valueStr string
			switch ob := v.(type) {
			case env.Object:
				if mode == 0 {
					valueStr = ob.Print(*idx)
				} else {
					valueStr = ob.Inspect(*idx)
				}
			default:
				valueStr = fmt.Sprint(ob)
			}
			termPrintln(valueStr)
			// Count the actual number of lines this entry takes (including newlines in the value)
			totalLines += strings.Count(valueStr, "\n") + 1
			CloseProps()
		} else {
			termPrintln("")
			totalLines += 1
		}
	}
	termPrintln(fmt.Sprintf("Page %d/%d (n=next, p=prev, m=mode)", currentPage+1, totalPages))
	totalLines += 1 // +1 for footer
	moveUp = totalLines

	defer func() {
		// Show cursor.
		termPrint("\033[?25h")
	}()

	for {
		ascii, keyCode, err := GetChar()

		if (ascii == 3 || ascii == 27) || err != nil {
			ShowCur()
			return nil, true
		}

		if ascii == 13 {
			globalIndex := start + localCurr
			if globalIndex < totalItems {
				return bloc.Series.Get(globalIndex), false
			}
			return nil, true // Fallback if out of bounds
		}

		if ascii == 77 || ascii == 109 {
			mode = 1 - mode
			goto DODO
		}

		if ascii == 110 || ascii == 78 { // 'n' or 'N'
			if currentPage < totalPages-1 {
				currentPage++
				localCurr = 0
				goto DODO
			}
		} else if ascii == 112 || ascii == 80 { // 'p' or 'P'
			if currentPage > 0 {
				currentPage--
				localCurr = 0
				goto DODO
			}
		}

		if keyCode == 40 {
			localCurr++
			if localCurr >= displayLen {
				if currentPage < totalPages-1 {
					currentPage++
					localCurr = 0
				} else {
					currentPage = 0
					localCurr = 0
				}
			}
			goto DODO
		} else if keyCode == 38 {
			localCurr--
			if localCurr < 0 {
				if currentPage > 0 {
					currentPage--
					localCurr = pageSize - 1
					if localCurr >= (totalItems - currentPage*pageSize) {
						localCurr = (totalItems - currentPage*pageSize) - 1
					}
				} else {
					currentPage = totalPages - 1
					localCurr = (totalItems - 1) % pageSize
				}
			}
			goto DODO
		}
	}
}

func DisplaySelection(bloc env.Block, idx *env.Idxs, right int) (env.Object, bool) {
	HideCur()
	curr := 0
	moveUp := 0
	mode := 0 // 0 - human, 1 - dev

	// Check if first element is a string - if so, duplicate strings to create string-string pairs
	if bloc.Series.Len() > 0 {
		if _, ok := bloc.Series.Get(0).(env.String); ok {
			// Create a new block with duplicated strings (string-string pairs)
			newSeries := make([]env.Object, bloc.Series.Len()*2)
			for i := 0; i < bloc.Series.Len(); i++ {
				str := bloc.Series.Get(i)
				newSeries[i*2] = str   // First copy
				newSeries[i*2+1] = str // Second copy (for display)
			}
			bloc.Series = *env.NewTSeries(newSeries)
		}
	}

	len := bloc.Series.Len() / 2
DODO1:
	if moveUp > 0 {
		CurUp(moveUp)
	}
	SaveCurPos()
	idents := make([]env.Object, (bloc.Series.Len()/2)+3)
	//fmt.Println("---")
	//fmt.Println(bloc.Series.Len())
	for i := 0; i < bloc.Series.Len(); i += 2 {
		//fmt.Println(i)
		// Store the identifier (could be Word, String, or Integer)
		idents[i/2] = bloc.Series.Get(i)
		label := bloc.Series.Get(i + 1)
		// ClearLine()
		CurRight(right)
		if i/2 == curr {
			ColorGreen()
			Bold()
			termPrint("\u00bb ")
		} else {
			termPrint(" ")
		}
		switch ob := label.(type) {
		case env.String:
			termPrintln(ob.Value + " ")
		default:
			termPrintln("" + fmt.Sprint(ob) + "***")
		}
		CloseProps()
		// term.CurUp(1)
	}

	moveUp = bloc.Series.Len() / 2

	defer func() {
		// Show cursor.
		termPrint("\033[?25h")
	}()

	// RestoreCurPos()

	for {
		ascii, keyCode, err := GetChar()

		if (ascii == 3 || ascii == 27) || err != nil {
			//fmt.Println()
			ShowCur()
			return nil, true
		}

		if ascii == 13 {
			termPrintln("")
			return idents[curr], false
		}

		if ascii == 77 || ascii == 109 {
			if mode == 0 {
				mode = 1
			} else {
				mode = 0
			}
			goto DODO1
		}

		if keyCode == 40 {
			curr++
			if curr > len-1 {
				curr = 0
			}
			goto DODO1
		} else if keyCode == 38 {
			curr--
			if curr < 0 {
				curr = len - 1
			}
			goto DODO1
		}
	}
}

func DisplayInputField(right int, mlen int) (env.Object, bool) {
	// HideCur()
	//curr := 0
	moveUp := 0
	text := ""
	//DODO1:
	if moveUp > 0 {
		CurUp(moveUp)
	}

	defer func() {
		// Show cursor.
		// termPrint("\033[?25h")
	}()

	// RestoreCurPos()
	//termPrintln(".")
	//termPrintln(".")
	//CurUp(2)

	CurRight(right)

	SaveCurPos()

	for {
		letter, ascii, _, err := GetChar2()
		//		letter := fmt.Scan()
		// RestoreCurPos()
		//CurDown(1)
		//termPrint("-----------")
		//termPrint(ascii)
		//CurUp(1)
		if (ascii == 3 || ascii == 27) || err != nil {
			// ShowCur()
			return nil, true
		} else if (ascii == 127) || err != nil {
			// Ensure we don't cause a slice bounds error on empty text
			if len(text) > 0 {
				text = text[0 : len(text)-1]
			}
			RestoreCurPos()
			termPrint("                  ")
			RestoreCurPos()
			termPrint(text)
		} else if ascii == 13 {
			termPrintln("")
			termPrintln("")
			return *env.NewString(text), false
		} else {
			if len(text) < mlen {
				text += letter
				RestoreCurPos()
				termPrint(text)
			}
		}

		// if keyCode == 40 {
		// }
		// else if keyCode == 38 {
		// }
	}
}

// DisplayDateInput displays an interactive date input in format YYYY-MM-DD
// Allows arrow keys to navigate between fields and increment/decrement values
// Returns a string in format "YYYY-MM-DD"
func DisplayDateInput(initialDate string, right int) (env.Object, bool) {
	// Parse initial date or use current date
	year, month, day := 0, 0, 0
	if initialDate != "" {
		parsed := strings.Split(initialDate, "-")
		if len(parsed) == 3 {
			year, _ = strconv.Atoi(parsed[0])
			month, _ = strconv.Atoi(parsed[1])
			day, _ = strconv.Atoi(parsed[2])
		}
	}

	// If parsing failed, use current date
	if year == 0 || month == 0 || day == 0 {
		// Use a default date as fallback
		year, month, day = 2025, 1, 1
	}

	field := 0 // 0=year, 1=month, 2=day

	// Helper function to get days in month
	daysInMonth := func(y, m int) int {
		switch m {
		case 2:
			if (y%4 == 0 && y%100 != 0) || (y%400 == 0) {
				return 29
			}
			return 28
		case 4, 6, 9, 11:
			return 30
		default:
			return 31
		}
	}

	// Helper to format date string
	formatDate := func() string {
		return fmt.Sprintf("%04d-%02d-%02d", year, month, day)
	}

	// Helper to display with highlighting
	display := func() {
		RestoreCurPos()
		dateStr := formatDate()
		parts := strings.Split(dateStr, "-")

		for i, part := range parts {
			if i > 0 {
				termPrint("-")
			}
			if i == field {
				ColorBrGreen()
				Bold()
				termPrint(part)
				CloseProps()
			} else {
				termPrint(part)
			}
		}
	}

	CurRight(right)
	SaveCurPos()
	display()

	defer func() {
		ShowCur()
	}()

	for {
		ascii, keyCode, err := GetChar()

		if (ascii == 3 || ascii == 27) || err != nil {
			termPrintln("")
			return nil, true
		}

		if ascii == 13 {
			termPrintln("")
			return *env.NewString(formatDate()), false
		}

		// Left/Right arrows to switch fields
		if keyCode == 37 { // Left arrow
			field--
			if field < 0 {
				field = 2
			}
			display()
		} else if keyCode == 39 { // Right arrow
			field++
			if field > 2 {
				field = 0
			}
			display()
		} else if keyCode == 38 { // Up arrow - increment
			switch field {
			case 0: // Year
				year++
				if year > 9999 {
					year = 1
				}
			case 1: // Month
				month++
				if month > 12 {
					month = 1
				}
				// Adjust day if necessary
				maxDays := daysInMonth(year, month)
				if day > maxDays {
					day = maxDays
				}
			case 2: // Day
				day++
				maxDays := daysInMonth(year, month)
				if day > maxDays {
					day = 1
				}
			}
			display()
		} else if keyCode == 40 { // Down arrow - decrement
			switch field {
			case 0: // Year
				year--
				if year < 1 {
					year = 9999
				}
			case 1: // Month
				month--
				if month < 1 {
					month = 12
				}
				// Adjust day if necessary
				maxDays := daysInMonth(year, month)
				if day > maxDays {
					day = maxDays
				}
			case 2: // Day
				day--
				if day < 1 {
					day = daysInMonth(year, month)
				}
			}
			display()
		} else if ascii >= 48 && ascii <= 57 { // Digit keys 0-9
			digit := int(ascii - 48)
			switch field {
			case 0: // Year - allow typing
				yearStr := fmt.Sprintf("%04d", year)
				yearStr = yearStr[1:] + strconv.Itoa(digit)
				year, _ = strconv.Atoi(yearStr)
			case 1: // Month
				if digit >= 0 && digit <= 1 {
					month = month%10 + digit*10
					if month > 12 {
						month = digit
					}
					if month == 0 {
						month = 10
					}
				} else if month < 10 {
					month = digit
					if month == 0 {
						month = 1
					}
				}
				// Adjust day if necessary
				maxDays := daysInMonth(year, month)
				if day > maxDays {
					day = maxDays
				}
			case 2: // Day
				maxDays := daysInMonth(year, month)
				newDay := day%10 + digit*10
				if newDay > maxDays || newDay == 0 {
					newDay = digit
					if newDay == 0 {
						newDay = 1
					}
				}
				day = newDay
			}
			display()
		}
	}
}

func DisplayDict(bloc env.Dict, idx *env.Idxs) (env.Object, bool) {
	HideCur()
	curr := 0
	moveUp := 0
	mode := 0 // 0 - human, 1 - dev
	len1 := len(bloc.Data)
	// make a slice for keys
	keys := make([]string, len(bloc.Data))
	i := 0
	for k := range bloc.Data {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
DODO:
	if moveUp > 0 {
		CurUp(moveUp)
	}
	SaveCurPos()
	totalLines := 0
	for ii, k := range keys {
		// for k, v := range bloc.Data {
		v := bloc.Data[k]
		ClearLine()
		if ii == curr {
			ColorBrGreen()
			Bold()
			termPrint("\u00bb ")
		} else {
			termPrint(" ")
		}
		Bold()
		termPrint(k + ": ")
		ResetBold()
		var valueStr string
		switch ob := v.(type) {
		case env.Object:
			if mode == 0 {
				valueStr = ob.Print(*idx)
			} else {
				valueStr = ob.Inspect(*idx)
			}
		default:
			valueStr = fmt.Sprint(ob)
		}
		termPrintln(valueStr)
		// Count the actual number of lines this entry takes (including newlines in the value)
		totalLines += strings.Count(valueStr, "\n") + 1
		CloseProps()
		// term.CurUp(1)
	}

	moveUp = totalLines

	defer func() {
		// Show cursor.
		termPrint("\033[?25h")
	}()

	// RestoreCurPos()

	for {
		ascii, keyCode, err := GetChar()

		if (ascii == 3 || ascii == 27) || err != nil {
			//termPrintln()
			ShowCur()
			return nil, true
		}

		if ascii == 13 {
			//termPrintln()
			ret := ""
			for ii, k := range keys {
				if ii == curr {
					ret = k
				}
			}
			return *env.NewString(ret), false // bloc.Series.Get(curr), false
		}

		if ascii == 120 {
			//termPrintln()
			var ret env.Object
			for ii, k := range keys {
				if ii == curr {
					ret = bloc.Data[k].(env.Object)
				}
			}
			return ret, false // bloc.Series.Get(curr), false
		}

		if ascii == 77 || ascii == 109 {
			if mode == 0 {
				mode = 1
			} else {
				mode = 0
			}
			goto DODO
		}

		if keyCode == 40 {
			curr++
			if curr > len1-1 {
				curr = 0
			}
			goto DODO
		} else if keyCode == 38 {
			curr--
			if curr < 0 {
				curr = len1 - 1
			}
			goto DODO
		}
	}
}

func DisplayTableRow(bloc env.TableRow, idx *env.Idxs) (env.Object, bool) {
	HideCur()
	curr := 0
	moveUp := 0
	mode := 0 // 0 - human, 1 - dev
	len1 := len(bloc.Values)
	// make a slice for keys
	/* keys := make([]string, len(bloc.Data))
	i := 0
	for k, _ := range bloc.Data {
		keys[i] = k
		i++
	}
	sort.Strings(keys)*/
DODO:
	if moveUp > 0 {
		CurUp(moveUp)
	}
	SaveCurPos()
	totalLines := 0
	for ii, k := range bloc.Uplink.GetColumnNames() {
		// for k, v := range bloc.Data {
		v := bloc.Values[ii]
		ClearLine()
		if ii == curr {
			ColorBrGreen()
			Bold()
			termPrint("\u00bb ")
		} else {
			termPrint(" ")
		}
		Bold()
		termPrint(k + ": ")
		ResetBold()
		var valueStr string
		switch ob := v.(type) {
		case env.Object:
			if mode == 0 {
				valueStr = ob.Print(*idx)
			} else {
				valueStr = ob.Inspect(*idx)
			}
		default:
			valueStr = fmt.Sprint(ob)
		}
		termPrintln(valueStr)
		// Count the actual number of lines this entry takes (including newlines in the value)
		totalLines += strings.Count(valueStr, "\n") + 1
		CloseProps()
		// term.CurUp(1)
	}

	moveUp = totalLines

	defer func() {
		// Show cursor.
		termPrint("\033[?25h")
	}()

	// RestoreCurPos()

	for {
		ascii, keyCode, err := GetChar()

		if (ascii == 3 || ascii == 27) || err != nil {
			//termPrintln()
			ShowCur()
			return nil, true
		}

		if ascii == 13 {
			return env.ToRyeValue(bloc.Values[curr]), false
		}

		if ascii == 120 {
			//termPrintln()
			return env.String{Value: bloc.Uplink.GetColumnNames()[curr]}, false
		}

		if ascii == 77 || ascii == 109 {
			if mode == 0 {
				mode = 1
			} else {
				mode = 0
			}
			goto DODO
		}

		if keyCode == 40 {
			curr++
			if curr > len1-1 {
				curr = 0
			}
			goto DODO
		} else if keyCode == 38 {
			curr--
			if curr < 0 {
				curr = len1 - 1
			}
			goto DODO
		}
	}
}

func DisplayTable(bloc env.Table, idx *env.Idxs) (env.Object, bool) {
	mode := 0 // 0 - human, 1 - dev
	totalItems := len(bloc.Rows)
	_, height, err := goterm.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		height = 20 // Fallback default
	}
	pageSize := height - 5 // Reserve lines for header, separator, footer, prompts
	if pageSize < 1 {
		pageSize = 1
	}
	totalPages := (totalItems + pageSize - 1) / pageSize // Ceiling division
	if totalPages == 0 {
		totalPages = 1
	}

	// get the ideal widths of columns
	widths := make([]int, len(bloc.Cols))
	// check all col names
	for ic, col := range bloc.Cols {
		widths[ic] = len(col) + 1
	}
	// check all data
	for _, r := range bloc.Rows {
		for ic, v := range r.Values {
			ww := 5
			switch val := v.(type) {
			case string:
				ww = len(val) + 2
				if ww > 52 {
					ww = 52
				}
			case int64:
				ww = len(strconv.Itoa(int(val))) + 1
			case env.Integer:
				ww = len(strconv.Itoa(int(val.Value))) + 1
			case float64:
				ww = len(strconv.FormatFloat(val, 'f', 2, 64)) + 1
			case env.Decimal:
				ww = len(strconv.FormatFloat(val.Value, 'f', 2, 64)) + 1
			case env.String:
				ww = len(val.Print(*idx))
				if ww > 52 {
					ww = 52
				}
				//if ww > 60 {
				// ww = 60
				//}
			case env.Vector:
				ww = len(val.Print(*idx))
			}
			if len(widths) > ic && widths[ic] < ww {
				widths[ic] = ww + 1
			}
		}
	}
	fulwidth := 0
	for _, w := range widths {
		fulwidth += w + 2
	}

	// If totalPages <= 1, use inline interactive mode
	if totalPages <= 1 {
		HideCur()
		curr := 0
		moveUp := 0

		defer func() {
			ShowCur()
		}()

	INLINE_DODO:
		if moveUp > 0 {
			CurUp(moveUp)
		}
		SaveCurPos()

		// Print header
		for ic, cn := range bloc.Cols {
			Bold()
			termPrintf("| %-"+strconv.Itoa(widths[ic])+"s", cn)
			CloseProps()
		}
		termPrintln("|")
		termPrintln("+" + strings.Repeat("-", fulwidth-1) + "+")

		// Print all rows with cursor highlighting
		for i, r := range bloc.Rows {
			ClearLine()
			if i == curr {
				ColorBrGreen()
				termPrint("")
			} else {
				termPrint("")
			}
			for ic, v := range r.Values {
				if ic < len(widths) {
					switch ob := v.(type) {
					case env.Object:
						if mode == 0 {
							termPrintf("| %-"+strconv.Itoa(widths[ic])+"s", util.TruncateString(ob.Print(*idx), widths[ic]))
						} else {
							termPrintf("| %-"+strconv.Itoa(widths[ic])+"s", ob.Inspect(*idx))
						}
					default:
						termPrintf("| %-"+strconv.Itoa(widths[ic])+"s", fmt.Sprint(ob))
					}
				}
			}
			CloseProps()
			termPrintln("|")
		}

		moveUp = totalItems + 2 // rows + header + separator line

		for {
			ascii, keyCode, err := GetChar()

			if (ascii == 3 || ascii == 27) || err != nil {
				return bloc, true // Return full table on Ctrl+C or Esc
			}

			if ascii == 13 {
				if curr < totalItems {
					return bloc.GetRowNew(curr), false // Return selected row on Enter
				}
				return nil, true
			}

			if ascii == 77 || ascii == 109 { // 'm' or 'M' for mode toggle
				mode = 1 - mode
				goto INLINE_DODO
			}

			if keyCode == 40 { // Down arrow
				curr++
				if curr >= totalItems {
					curr = 0 // Wrap to top
				}
				goto INLINE_DODO
			} else if keyCode == 38 { // Up arrow
				curr--
				if curr < 0 {
					curr = totalItems - 1 // Wrap to bottom
				}
				goto INLINE_DODO
			}
		}
	}

	// Full-screen paginated mode for tables that need pagination
	HideCur()
	currentPage := 0
	localCurr := 0
	moveUp := 0

DODO:
	if moveUp > 0 {
		CurUp(moveUp)
	}
	SaveCurPos()
	start := currentPage * pageSize
	end := start + pageSize
	if end > totalItems {
		end = totalItems
	}
	displayedItems := end - start

	// Print header
	for ic, cn := range bloc.Cols {
		Bold()
		termPrintf("| %-"+strconv.Itoa(widths[ic])+"s", cn)
		CloseProps()
	}
	termPrintln("|")
	termPrintln("+" + strings.Repeat("-", fulwidth-1) + "+")

	// Print rows and clear extra lines
	for i := 0; i < pageSize; i++ {
		ClearLine()
		if i < displayedItems {
			r := bloc.Rows[start+i]
			if i == localCurr {
				ColorBrGreen()
				termPrint("")
			} else {
				termPrint("")
			}
			for ic, v := range r.Values {
				if ic < len(widths) {
					switch ob := v.(type) {
					case env.Object:
						if mode == 0 {
							termPrintf("| %-"+strconv.Itoa(widths[ic])+"s", util.TruncateString(ob.Print(*idx), widths[ic]))
						} else {
							termPrintf("| %-"+strconv.Itoa(widths[ic])+"s", ob.Inspect(*idx))
						}
					default:
						termPrintf("| %-"+strconv.Itoa(widths[ic])+"s", fmt.Sprint(ob))
					}
				}
			}
			CloseProps()
			termPrintln("|")
		} else {
			termPrintln("")
		}
	}

	// Print footer
	termPrintln(fmt.Sprintf("Page %d/%d (n=next, p=prev, m=mode)", currentPage+1, totalPages))

	moveUp = pageSize + 3 // rows + header + sep + footer

	defer func() {
		// Show cursor.
		termPrint("\033[?25h")
	}()

	for {
		ascii, keyCode, err := GetChar()

		if (ascii == 3 || ascii == 27) || err != nil {
			ShowCur()
			return nil, true
		}

		if ascii == 13 {
			globalIndex := start + localCurr
			if globalIndex < totalItems {
				return bloc.GetRowNew(globalIndex), false
			}
			return nil, true // Fallback if out of bounds
		}

		if ascii == 77 || ascii == 109 {
			mode = 1 - mode
			goto DODO
		}

		if ascii == 110 || ascii == 78 { // 'n' or 'N'
			if currentPage < totalPages-1 {
				currentPage++
				localCurr = 0
				goto DODO
			}
		} else if ascii == 112 || ascii == 80 { // 'p' or 'P'
			if currentPage > 0 {
				currentPage--
				localCurr = 0
				goto DODO
			}
		}

		if keyCode == 40 {
			localCurr++
			if localCurr >= displayedItems {
				if currentPage < totalPages-1 {
					currentPage++
					localCurr = 0
				} else {
					currentPage = 0
					localCurr = 0
				}
			}
			goto DODO
		} else if keyCode == 38 {
			localCurr--
			if localCurr < 0 {
				if currentPage > 0 {
					currentPage--
					localCurr = pageSize - 1
					if localCurr >= (totalItems - currentPage*pageSize) {
						localCurr = (totalItems - currentPage*pageSize) - 1
					}
				} else {
					currentPage = totalPages - 1
					localCurr = (totalItems - 1) % pageSize
				}
			}
			goto DODO
		}
	}
}

// ideation:
// .display\custom fn { x } { -> 'subject .elipsis 20 .red .prn , spacer 2 , -> 'score .align-right 10 .print }
func DisplayTableCustom(bloc env.Table, myfn func(row env.Object, iscurr env.Integer), idx *env.Idxs) (env.Object, bool) {
	HideCur()
	curr := 0
	moveUp := 0
	mode := 0 // 0 - human, 1 - dev
	// get the ideal widths of columns
	widths := make([]int, len(bloc.Cols))
	// check all col names
	for ic, col := range bloc.Cols {
		widths[ic] = len(col) + 1
	}
	// check all data
	for _, r := range bloc.Rows {
		for ic, v := range r.Values {
			ww := 5
			switch val := v.(type) {
			case string:
				ww = len(val) + 2
				if ww > 52 {
					ww = 52
				}
			case int64:
				ww = len(strconv.Itoa(int(val))) + 1
			case env.Integer:
				ww = len(strconv.Itoa(int(val.Value))) + 1
			case float64:
				ww = len(strconv.FormatFloat(val, 'f', 2, 64)) + 1
			case env.Decimal:
				ww = len(strconv.FormatFloat(val.Value, 'f', 2, 64)) + 1
			case env.String:
				ww = len(val.Print(*idx))
				if ww > 52 {
					ww = 52
				}
				//if ww > 60 {
				// ww = 60
				//}
			case env.Vector:
				ww = len(val.Print(*idx))
			}
			if len(widths) > ic && widths[ic] < ww {
				widths[ic] = ww + 1
			}
		}
	}
	fulwidth := 0
	for _, w := range widths {
		fulwidth += w + 2
	}

DODO:
	if moveUp > 0 {
		CurUp(moveUp)
	}
	SaveCurPos()
	/* for ic, cn := range bloc.Cols {
		Bold()
		termPrintf("| %-"+strconv.Itoa(widths[ic])+"s", cn)
		CloseProps()
	}
	termPrintln("|")
	termPrintln("+" + strings.Repeat("-", fulwidth-1) + "+")
	*/

	for range bloc.Rows {
		ClearLine()
	}
	for i, r := range bloc.Rows {
		iscurr := *env.NewInteger(0)
		if i == curr {
			iscurr = *env.NewInteger(1)
		}

		// call funtion with row and is-current value
		myfn(r, iscurr)

		//CloseProps()
		// termPrintln("|")
	}

	moveUp = len(bloc.Rows)

	defer func() {
		// Show cursor.
		termPrint("\033[?25h")
	}()

	// RestoreCurPos()

	for {
		ascii, keyCode, err := GetChar()

		if (ascii == 3 || ascii == 27) || err != nil {
			termPrintln("")
			ShowCur()
			return nil, true
		}

		if ascii == 13 {
			termPrintln("")
			return bloc.GetRowNew(curr), false // bloc.Series.Get(curr), false
		}

		if ascii == 77 || ascii == 109 {
			if mode == 0 {
				mode = 1
			} else {
				mode = 0
			}
			goto DODO
		}

		if keyCode == 40 {
			curr++
			goto DODO
		} else if keyCode == 38 {
			curr--
			goto DODO
		}
	}
}

func itoa(i int) {
	panic("unimplemented")
}

func ShowCur() {
	termPrint("\x1b[?25h")
}

func HideCur() {
	termPrint("\x1b[?25l")
}
func SaveCurPos() {
	termPrint("\x1b7")
}

func ClearLine() {
	termPrint("\x1b[0K")
}

func RestoreCurPos() {
	termPrint("\x1b8")
}

// Standard colors
func ColorBlack() {
	termPrint("\x1b[30m")
}
func ColorRed() {
	termPrint("\x1b[31m")
}
func ColorGreen() {
	termPrint("\x1b[32m")
}
func ColorYellow() {
	termPrint("\x1b[33m")
}
func ColorBlue() {
	termPrint("\x1b[34m")
}
func ColorMagenta() {
	termPrint("\x1b[35m")
}
func ColorCyan() {
	termPrint("\x1b[36m")
}
func ColorWhite() {
	termPrint("\x1b[37m")
}

// Standard colors returned
func StrColorBlack() string {
	return "\x1b[30m"
}
func StrColorRed() string {
	return "\x1b[31m"
}
func StrColorGreen() string {
	return "\x1b[32m"
}
func StrColorYellow() string {
	return "\x1b[33m"
}
func StrColorBlue() string {
	return "\x1b[34m"
}
func StrColorMagenta() string {
	return "\x1b[35m"
}
func StrColorCyan() string {
	return "\x1b[36m"
}
func StrColorWhite() string {
	return "\x1b[37m"
}

func StrColorBrBlack() string {
	return "\x1b[30;1m"
}
func StrColorBrRed() string {
	return "\x1b[31;1m"
}
func StrColorBrGreen() string {
	return "\x1b[32;1m"
}
func StrColorBrYellow() string {
	return "\x1b[33;1m"
}
func StrColorBrBlue() string {
	return "\x1b[34;1m"
}
func StrColorBrMagenta() string {
	return "\x1b[35;1m"
}
func StrColorBrCyan() string {
	return "\x1b[36;1m"
}
func StrColorBrWhite() string {
	return "\x1b[37;1m"
}

// Bright colors
func ColorBrBlack() {
	termPrint("\x1b[30;1m")
}
func ColorBrRed() {
	termPrint("\x1b[31;1m")
}
func ColorBrGreen() {
	termPrint("\x1b[32;1m")
}
func ColorBrYellow() {
	termPrint("\x1b[33;1m")
}
func ColorBrBlue() {
	termPrint("\x1b[34;1m")
}
func ColorBrMagenta() {
	termPrint("\x1b[36;1m")
}
func ColorBrCyan() {
	termPrint("\x1b[37;1m")
}
func ColorBrWhite() {
	termPrint("\x1b[37;1m")
}

// Background string functions
func StrColorBgBlack() string {
	return "\x1b[40m"
}
func StrColorBgRed() string {
	return "\x1b[41m"
}
func StrColorBgGreen() string {
	return "\x1b[42m"
}
func StrColorBgYellow() string {
	return "\x1b[43m"
}
func StrColorBgBlue() string {
	return "\x1b[44m"
}
func StrColorBgMagenta() string {
	return "\x1b[45m"
}
func StrColorBgCyan() string {
	return "\x1b[46m"
}
func StrColorBgWhite() string {
	return "\x1b[47m"
}

// Background
func ColorBgBlack() {
	termPrint("\x1b[40m")
}
func ColorBgRed() {
	termPrint("\x1b[41m")
}
func ColorBgGreen() {
	termPrint("\x1b[42m")
}
func ColorBgYellow() {
	termPrint("\x1b[43m")
}
func ColorBgBlue() {
	termPrint("\x1b[44m")
}
func ColorBgMagenta() {
	termPrint("\x1b[45m")
}
func ColorBgCyan() {
	termPrint("\x1b[46m")
}
func ColorBgWhite() {
	termPrint("\x1b[47m")
}

// Font style string functions
func StrBold() string {
	return "\x1b[1m"
}
func StrUnderline() string {
	return "\x1b[4m"
}
func StrResetBold() string {
	return "\x1b[22m"
}

// Font style
func Bold() {
	termPrint("\x1b[1m")
}
func Italic() {
	termPrint("\x1b[3m")
}
func Underline() {
	termPrint("\x1b[4m")
}
func ResetBold() {
	termPrint("\x1b[22m")
}
func CloseProps() {
	termPrint("\x1b[0m")
}
func StrCloseProps() string {
	return "\x1b[0m"
}
func CurUp(n int) {
	termPrintf("\x1b[%dA", n)
}
func CurDown(n int) {
	termPrintf("\x1b[%dB", n)
}
func CurRight(n int) {
	termPrintf("\x1b[%dC", n)
}
func CurLeft(n int) {
	termPrintf("\x1b[%dD", n)
}

// DisplayTextArea displays an interactive multiline text input
// width is max characters per line, height is number of lines
// text is optional initial text that will be split into lines
// Returns the text as a string with newlines between lines
func DisplayTextArea(width, height int, text string) (env.Object, bool) {
	HideCur()
	// Initialize lines as empty strings
	lines := make([]string, height)
	for i := range lines {
		lines[i] = ""
	}

	// If text is provided, split it into lines and fill the lines variable
	if text != "" {
		inputLines := strings.Split(text, "\n")
		// Copy input lines to lines, truncating if more lines than height
		for i := 0; i < len(inputLines) && i < height; i++ {
			// Also truncate line width if needed
			if len(inputLines[i]) > width {
				lines[i] = inputLines[i][:width]
			} else {
				lines[i] = inputLines[i]
			}
		}
	}

	curRow := 0 // Current row (0 to height-1)
	curCol := 0 // Current column position

	// Helper to display all lines with cursor
	display := func() {
		RestoreCurPos()
		for i := 0; i < height; i++ {
			ClearLine()
			line := lines[i]
			// Pad line to width for visual consistency
			if len(line) < width {
				line = line + strings.Repeat(" ", width-len(line))
			}
			// Display line with cursor highlight
			if i == curRow {
				// Show cursor position with underscore or highlight
				pre := ""
				if curCol > 0 && curCol <= len(lines[i]) {
					pre = lines[i][:curCol]
				} else if curCol > len(lines[i]) {
					pre = lines[i] + strings.Repeat(" ", curCol-len(lines[i]))
				}
				post := ""
				if curCol < len(lines[i]) {
					post = lines[i][curCol:]
				}
				cursorChar := " "
				if curCol < len(lines[i]) {
					cursorChar = string(lines[i][curCol])
				}
				termPrint(pre)
				ColorBgGreen()
				ColorBlack()
				termPrint(cursorChar)
				CloseProps()
				if len(post) > 1 {
					termPrint(post[1:])
				}
				// Pad rest
				remaining := width - len(lines[i])
				if remaining > 0 {
					termPrint(strings.Repeat(" ", remaining))
				}
			} else {
				termPrint(line)
			}
			termPrintln("")
		}
		// Display border below textarea with corner on right (dim color)
		ClearLine()
		ColorMagenta()
		termPrint("─" + strings.Repeat(" ", width-1) + "┘\n")
		termPrint("ctrl+d to submit, ctrl+c to cancel")
		CloseProps()
		termPrintln("")
	}

	SaveCurPos()
	display()

	defer func() {
		ShowCur()
	}()

	for {
		letter, ascii, keyCode, err := GetChar2()

		if (ascii == 3 || ascii == 27) || err != nil {
			termPrintln("")
			return nil, true
		}

		// Regular Enter - move to next line
		if ascii == 13 {
			if curRow < height-1 {
				curRow++
				curCol = 0
			}
			display()
			continue
		}

		// Ctrl+D to submit (standard Unix "end of input")
		if ascii == 4 {
			// Join lines with newlines and return
			result := strings.Join(lines, "\n")
			// Trim trailing empty lines
			result = strings.TrimRight(result, "\n ")
			termPrintln("")
			return *env.NewString(result), false
		}

		// Backspace
		if ascii == 127 {
			if curCol > 0 {
				// Delete character before cursor
				line := lines[curRow]
				if curCol <= len(line) {
					lines[curRow] = line[:curCol-1] + line[curCol:]
				}
				curCol--
			} else if curRow > 0 {
				// At start of line, merge with previous line
				prevLen := len(lines[curRow-1])
				lines[curRow-1] = lines[curRow-1] + lines[curRow]
				// Shift remaining lines up
				for i := curRow; i < height-1; i++ {
					lines[i] = lines[i+1]
				}
				lines[height-1] = ""
				curRow--
				curCol = prevLen
			}
			display()
			continue
		}

		// Arrow keys
		if keyCode == 37 { // Left
			if curCol > 0 {
				curCol--
			} else if curRow > 0 {
				curRow--
				curCol = len(lines[curRow])
			}
			display()
			continue
		}
		if keyCode == 39 { // Right
			if curCol < len(lines[curRow]) {
				curCol++
			} else if curRow < height-1 {
				curRow++
				curCol = 0
			}
			display()
			continue
		}
		if keyCode == 38 { // Up
			if curRow > 0 {
				curRow--
				if curCol > len(lines[curRow]) {
					curCol = len(lines[curRow])
				}
			}
			display()
			continue
		}
		if keyCode == 40 { // Down
			if curRow < height-1 {
				curRow++
				if curCol > len(lines[curRow]) {
					curCol = len(lines[curRow])
				}
			}
			display()
			continue
		}

		// Regular character input
		if ascii >= 32 && ascii < 127 {
			line := lines[curRow]
			if len(line) < width {
				// Insert character at cursor position
				if curCol >= len(line) {
					lines[curRow] = line + letter
				} else {
					lines[curRow] = line[:curCol] + letter + line[curCol:]
				}
				curCol++
			}
			display()
		}
	}
}

// GetChar and GetChar2 functions are implemented in platform-specific files:
// - term_unix.go for Unix/Linux systems
// - term_windows.go for Windows systems
