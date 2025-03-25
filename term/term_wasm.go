//go:build wasm
// +build wasm

package term

import (
	"fmt"

	"github.com/refaktor/rye/env"
)

var sendBack func(string)

func SetSB(fn func(string)) {
	sendBack = fn
}

func DisplayBlock(bloc env.Block, idx *env.Idxs) (env.Object, bool) {
	HideCur()
	curr := 0
	moveUp := 0
	mode := 0 // 0 - human, 1 - dev
	//	len := bloc.Series.Len()
	// DODO:
	if moveUp > 0 {
		CurUp(moveUp)
	}
	SaveCurPos()
	for i, v := range bloc.Series.S {
		ClearLine()
		if i == curr {
			ColorBrGreen()
			Bold()
			fmt.Print("\u00bb ")
		} else {
			fmt.Print(" ")
		}
		switch ob := v.(type) {
		case env.Object:
			if mode == 0 {
				fmt.Println("" + ob.Print(*idx) + "")
			} else {
				fmt.Println("" + ob.Inspect(*idx) + "")
			}
		default:
			fmt.Println("" + fmt.Sprint(ob) + "")
		}
		CloseProps()
		// term.CurUp(1)
	}

	moveUp = bloc.Series.Len()

	defer func() {
		// Show cursor.
		fmt.Println("\033[?25h")
	}()

	// RestoreCurPos()

	/* for {
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
	} */
	return bloc.Series.Get(curr), false
}

// DisplayBlock is dummy non-implementation for wasm for display builtin
// func DisplayBlock(bloc env.Block, idx *env.Idxs) (env.Object, bool) {
// 	panic("unimplemented")
// }

// DisplayDict is dummy non-implementation for wasm for display builtin
func DisplayDict(bloc env.Dict, idx *env.Idxs) (env.Object, bool) {
	panic("unimplemented")
}

// DisplayTableRow is dummy non-implementation for wasm for display builtin
func DisplayTableRow(bloc env.TableRow, idx *env.Idxs) (env.Object, bool) {
	panic("unimplemented")
}

// DisplayTable is dummy non-implementation for wasm for display builtin
func DisplayTable(bloc env.Table, idx *env.Idxs) (env.Object, bool) {
	panic("unimplemented")
}

func DisplayTableCustom(bloc env.Table, myfn func(row env.Object, iscurr env.Integer), idx *env.Idxs) (env.Object, bool) {
	panic("unimplemented")
}

func ShowCur() {
	sendBack("\x1b[?25h")
}

func HideCur() {
	sendBack("\x1b[?25l")
}
func SaveCurPos() {
	sendBack("\x1b7")
}

func ClearLine() {
	sendBack("\x1b[0K")
}

func RestoreCurPos() {
	sendBack("\x1b8")
}

// Standard colors
func ColorBlack() {
	sendBack("\x1b[30m")
}
func ColorRed() {
	sendBack("\x1b[31m")
}
func ColorGreen() {
	sendBack("\x1b[32m")
}
func ColorYellow() {
	sendBack("\x1b[33m")
}
func ColorBlue() {
	sendBack("\x1b[34m")
}
func ColorMagenta() {
	sendBack("\x1b[35m")
}
func ColorCyan() {
	sendBack("\x1b[36m")
}
func ColorWhite() {
	sendBack("\x1b[37m")
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
	sendBack("\x1b[30;1m")
}
func ColorBrRed() {
	sendBack("\x1b[31;1m")
}
func ColorBrGreen() {
	sendBack("\x1b[32;1m")
}
func ColorBrYellow() {
	sendBack("\x1b[33;1m")
}
func ColorBrBlue() {
	sendBack("\x1b[34;1m")
}
func ColorBrMagenta() {
	sendBack("\x1b[36;1m")
}
func ColorBrCyan() {
	sendBack("\x1b[37;1m")
}
func ColorBrWhite() {
	sendBack("\x1b[37;1m")
}

// Background
func ColorBgBlack() {
	sendBack("\x1b[40m")
}
func ColorBgRed() {
	sendBack("\x1b[41m")
}
func ColorBgGreen() {
	sendBack("\x1b[42m")
}
func ColorBgYellow() {
	sendBack("\x1b[43m")
}
func ColorBgBlue() {
	sendBack("\x1b[44m")
}
func ColorBgMagenta() {
	sendBack("\x1b[45m")
}
func ColorBgCyan() {
	sendBack("\x1b[46m")
}
func ColorBgWhite() {
	sendBack("\x1b[47m")
}

// Font style
func Bold() {
	sendBack("\x1b[1m")
}
func Italic() {
	sendBack("\x1b[3m")
}
func Underline() {
	sendBack("\x1b[4m")
}

func ResetBold() {
	sendBack("\x1b[22m")
}
func CloseProps() {
	sendBack("\x1b[0m")
}
func StrCloseProps() string {
	return "\x1b[0m"
}
func CurUp(n int) {
	sendBack(fmt.Sprintf("\x1b[%dA", n))
}
func CurDown(n int) {
	sendBack(fmt.Sprintf("\x1b[%dB", n))
}
func CurRight(n int) {
	sendBack(fmt.Sprintf("\x1b[%dC", n))
}
func CurLeft(n int) {
	sendBack(fmt.Sprintf("\x1b[%dD", n))
}
