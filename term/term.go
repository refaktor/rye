package term

import (
	"fmt"
	"rye/env"

	"github.com/pkg/term"
)

func DisplayBlock(bloc env.Block, idx *env.Idxs) (env.Object, bool) {
	HideCur()
	curr := 0
	moveUp := 0
	mode := 0 // 0 - human, 1 - dev
DODO:
	if moveUp > 0 {
		CurUp(moveUp)
	}
	SaveCurPos()
	for i, v := range bloc.Series.S {
		ClearLine()
		if i == curr {
			ColorOrange()
			fmt.Print("*")
		} else {
			fmt.Print(" ")
		}
		switch ob := v.(type) {
		case env.Object:
			if mode == 0 {
				fmt.Println(" " + ob.Probe(*idx) + " ")
			} else {
				fmt.Println(" " + ob.Inspect(*idx) + " ")
			}
		default:
			fmt.Println(" " + fmt.Sprint(ob) + " ")
		}
		CloseProps()
		// term.CurUp(1)
	}

	moveUp = bloc.Series.Len()

	defer func() {
		// Show cursor.
		fmt.Printf("\033[?25h")
	}()

	// RestoreCurPos()

	for {
		ascii, keyCode, err := GetChar()

		if (ascii == 3 || ascii == 27) || err != nil {
			fmt.Println()
			ShowCur()
			return nil, true
		}

		if ascii == 13 {
			fmt.Println()
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
			goto DODO
		} else if keyCode == 38 {
			curr--
			goto DODO
		}
	}
}

func ShowCur() {
	fmt.Print("\x1b[?25h")
}

func HideCur() {
	fmt.Print("\x1b[?25l")
}
func SaveCurPos() {
	fmt.Print("\x1b7")
}

func ClearLine() {
	fmt.Print("\x1b[0K")
}

func RestoreCurPos() {
	fmt.Print("\x1b8")
}

func ColorOrange() {
	fmt.Printf("\x1b[33m")
}

func CloseProps() {
	fmt.Printf("\x1b[0m")
}

func CurUp(n int) {
	fmt.Printf("\x1b[%dA", n)
}

// From https://github.com/paulrademacher/climenu/blob/master/getchar.go

func GetChar() (ascii int, keyCode int, err error) {
	t, _ := term.Open("/dev/tty")
	term.RawMode(t)
	bytes := make([]byte, 3)

	var numRead int
	numRead, err = t.Read(bytes)
	if err != nil {
		return
	}
	if numRead == 3 && bytes[0] == 27 && bytes[1] == 91 {
		// Three-character control sequence, beginning with "ESC-[".

		// Since there are no ASCII codes for arrow keys, we use
		// Javascript key codes.
		if bytes[2] == 65 {
			// Up
			keyCode = 38
		} else if bytes[2] == 66 {
			// Down
			keyCode = 40
		} else if bytes[2] == 67 {
			// Right
			keyCode = 39
		} else if bytes[2] == 68 {
			// Left
			keyCode = 37
		}
	} else if numRead == 1 {
		ascii = int(bytes[0])
	} else {
		// Two characters read??
	}
	t.Restore()
	t.Close()
	return
}
