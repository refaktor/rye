package term

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

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
	HideCur()
	curr := 0
	moveUp := 0
	mode := 0 // 0 - human, 1 - dev
	len := bloc.Series.Len()
DODO:
	if moveUp > 0 {
		CurUp(moveUp)
	}
	SaveCurPos()
	for i, v := range bloc.Series.S {
		ClearLine()
		if i == curr {
			ColorBrGreen()
			Bold()
			termPrint("\u00bb ")
		} else {
			termPrint(" ")
		}
		switch ob := v.(type) {
		case env.Object:
			if mode == 0 {
				termPrintln("" + ob.Print(*idx) + "")
			} else {
				termPrintln("" + ob.Inspect(*idx) + "")
			}
		default:
			termPrintln("" + fmt.Sprint(ob) + "")
		}
		CloseProps()
		// term.CurUp(1)
	}

	moveUp = bloc.Series.Len()

	defer func() {
		// Show cursor.
		termPrint("\033[?25h")
	}()

	// RestoreCurPos()

	// In WASM environment, we use a non-blocking GetChar that may return ESC (27)
	// to avoid deadlock when no key events are available
	for {
		ascii, keyCode, err := GetChar()

		if (ascii == 3 || ascii == 27) || err != nil {
			//fmt.Println()
			ShowCur()
			return nil, true
		}

		if ascii == 13 {
			//fmt.Println()
			return bloc.Series.Get(curr), false
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
			if curr > len-1 {
				curr = 0
			}
			goto DODO
		} else if keyCode == 38 {
			curr--
			if curr < 0 {
				curr = len - 1
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
	len := bloc.Series.Len() / 2
DODO1:
	if moveUp > 0 {
		CurUp(moveUp)
	}
	SaveCurPos()
	idents := make([]int, (bloc.Series.Len()/2)+3)
	//fmt.Println("---")
	//fmt.Println(bloc.Series.Len())
	for i := 0; i < bloc.Series.Len(); i += 2 {
		//fmt.Println(i)
		// todo check if it's word
		idents[i/2] = bloc.Series.Get(i).(env.Word).Index
		label := bloc.Series.Get(i + 1)
		// ClearLine()
		CurRight(right)
		if i/2 == curr {
			ColorMagenta()
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
			return *env.NewWord(idents[curr]), false
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
			text = text[0 : len(text)-1]
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
		switch ob := v.(type) {
		case env.Object:
			if mode == 0 {
				termPrintln("" + ob.Print(*idx) + "")
			} else {
				termPrintln("" + ob.Inspect(*idx) + "")
			}
		default:
			termPrintln(" " + fmt.Sprint(ob) + " ")
		}
		CloseProps()
		// term.CurUp(1)
	}

	moveUp = len(bloc.Data)

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
	for ii, k := range bloc.Uplink.Cols {
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
		switch ob := v.(type) {
		case env.Object:
			if mode == 0 {
				termPrintln("" + ob.Print(*idx) + "")
			} else {
				termPrintln("" + ob.Inspect(*idx) + "")
			}
		default:
			termPrintln(" " + fmt.Sprint(ob) + " ")
		}
		CloseProps()
		// term.CurUp(1)
	}

	moveUp = len(bloc.Values)

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
			return env.String{Value: bloc.Uplink.Cols[curr]}, false
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
	for ic, cn := range bloc.Cols {
		Bold()
		termPrintf("| %-"+strconv.Itoa(widths[ic])+"s", cn)
		CloseProps()
	}
	termPrintln("|")
	termPrintln("+" + strings.Repeat("-", fulwidth-1) + "+")

	for range bloc.Rows {
		ClearLine()
	}
	for i, r := range bloc.Rows {
		if i == curr {
			ColorBrGreen()
			termPrint("")
		} else {
			termPrint("")
		}
		for ic, v := range r.Values {
			if ic < len(widths) {
				// termPrintln(v)
				switch ob := v.(type) {
				case env.Object:
					if mode == 0 {
						termPrintf("| %-"+strconv.Itoa(widths[ic])+"s", util.TruncateString(ob.Print(*idx), widths[ic]))
						//termPrint("| " +  + "\t")
					} else {
						termPrintf("| %-"+strconv.Itoa(widths[ic])+"s", ob.Inspect(*idx))
						//termPrint("| " +  + "\t")
					}
				default:
					termPrintf("| %-"+strconv.Itoa(widths[ic])+"s", fmt.Sprint(ob))
					///termPrint("| " + +"\t")
				}
			}
			// term.CurUp(1)
		}
		CloseProps()
		termPrintln("|")
	}

	moveUp = len(bloc.Rows) + 2

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

// GetChar and GetChar2 functions are implemented in platform-specific files:
// - term_unix.go for Unix/Linux systems
// - term_windows.go for Windows systems
