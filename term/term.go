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

func DisplayDict(bloc env.Dict, idx *env.Idxs) (env.Object, bool) {
	HideCur()
	curr := 0
	moveUp := 0
	mode := 0 // 0 - human, 1 - dev
DODO:
	if moveUp > 0 {
		CurUp(moveUp)
	}
	SaveCurPos()
	for k, v := range bloc.Data {
		ClearLine()
		if 0 == curr {
			ColorOrange()
			fmt.Print("*")
		} else {
			fmt.Print(" ")
		}
		fmt.Print(k + ": ")
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

	moveUp = len(bloc.Data)

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
			return nil, false // bloc.Series.Get(curr), false
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

func DisplayTable(bloc env.Spreadsheet, idx *env.Idxs) (env.Object, bool) {
	HideCur()
	curr := 0
	moveUp := 0
	mode := 0 // 0 - human, 1 - dev
DODO:
	if moveUp > 0 {
		CurUp(moveUp)
	}
	SaveCurPos()
	for _, cc := range bloc.Cols {
		fmt.Print("| " + cc + "\t")
	}
	fmt.Println()
	fmt.Println("---------------------------------")

	if bloc.RawMode {
		for range bloc.Rows {
			ClearLine()
		}
		for i, r := range bloc.RawRows {
			ClearLine()
			for _, v := range r {
				if i == curr {
					ColorOrange()
					// fmt.Print("*")
				} else {
					// fmt.Print(" ")
				}
				fmt.Print("| " + fmt.Sprint(v) + "\t")
				CloseProps()
				// term.CurUp(1)
			}
			fmt.Println()
		}

	} else {
		for range bloc.Rows {
			ClearLine()
		}
		for _, r := range bloc.Rows {
			for i, v := range r.Values {
				if i == curr {
					ColorOrange()
					fmt.Print("*")
				} else {
					fmt.Print(" ")
				}
				switch ob := v.(type) {
				case env.Object:
					if mode == 0 {
						fmt.Print("| " + ob.Probe(*idx) + "\t")
					} else {
						fmt.Print("| " + ob.Inspect(*idx) + "\t")
					}
				default:
					fmt.Print("| " + fmt.Sprint(ob) + "\t")
				}
				CloseProps()
				// term.CurUp(1)
			}
			fmt.Println()
		}

	}

	moveUp = len(bloc.Rows)

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
			return nil, false // bloc.Series.Get(curr), false
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
