//go:build !b_norepl && !wasm && !js

package main

import (
	"context"
	"fmt"
	"os"

	"github.com/cszczepaniak/keyboard"
	"github.com/refaktor/rye/term"
	"github.com/refaktor/rye/util"
)

// Clear screen
func Clear() {
	fmt.Print("\033[2J")
}

// Move cursor to given position
func MoveCursor(x int, y int) {
	fmt.Printf("\033[%d;%dH", y, x)
}

// Move cursor up relative the current position
func MoveCursorUp(bias int) {
	fmt.Printf("\033[%dA", bias)
}

// Move cursor down relative the current position
func MoveCursorDown(bias int) {
	fmt.Printf("\033[%dB", bias)
}

// Move cursor forward relative the current position
func MoveCursorForward(bias int) {
	fmt.Printf("\033[%dC", bias)
}

// Move cursor backward relative the current position
func MoveCursorBackward(bias int) {
	fmt.Printf("\033[%dD", bias)
}

type Repl struct {
	ml *term.MLState

	dialect string

	fullCode string
}

func (r *Repl) recieveMessage(message string) {
	fmt.Print(message)
}

func (r *Repl) recieveLine(line string) string {
	res := r.evalLine(line)
	return res
}

func (r *Repl) evalLine(code string) string {
	fmt.Println("******1")
	fmt.Println(code)
	return code
}

// constructKeyEvent maps a rune and keyboard.Key to a util.KeyEvent, which uses javascript key event codes
// only keys used in microliner are mapped
func constructKeyEvent(r rune, k keyboard.Key) term.KeyEvent {
	// fmt.Println(r)
	// fmt.Println(k)
	var ctrl bool
	alt := k == keyboard.KeyEsc
	var code int
	ch := string(r)
	switch k {
	case keyboard.KeyCtrlA:
		ch = "a"
		ctrl = true
	case keyboard.KeyCtrlS:
		ch = "s"
		ctrl = true
	case keyboard.KeyCtrlC:
		ch = "c"
		ctrl = true
	case keyboard.KeyCtrlB:
		ch = "b"
		ctrl = true
	case keyboard.KeyCtrlD:
		ch = "d"
		ctrl = true
	case keyboard.KeyCtrlE:
		ch = "e"
		ctrl = true
	case keyboard.KeyCtrlF:
		ch = "f"
		ctrl = true
	case keyboard.KeyCtrlK:
		ch = "k"
		ctrl = true
	case keyboard.KeyCtrlL:
		ch = "l"
		ctrl = true
	case keyboard.KeyCtrlN:
		ch = "n"
		ctrl = true
	case keyboard.KeyCtrlP:
		ch = "p"
		ctrl = true
	case keyboard.KeyCtrlU:
		ch = "u"
		ctrl = true
	case keyboard.KeyCtrlX:
		ch = "x"
		ctrl = true

	case keyboard.KeyEnter:
		code = 13
	case keyboard.KeyTab:
		code = 9
	case keyboard.KeyBackspace, keyboard.KeyBackspace2:
		code = 8
	case keyboard.KeyDelete:
		code = 46
	case keyboard.KeyArrowRight:
		code = 39
	case keyboard.KeyArrowLeft:
		code = 37
	case keyboard.KeyArrowUp:
		code = 38
	case keyboard.KeyArrowDown:
		code = 40
	case keyboard.KeyHome:
		code = 36
	case keyboard.KeyEnd:
		code = 35

	case keyboard.KeySpace:
		ch = " "
		code = 20
	}
	return term.NewKeyEvent(ch, code, ctrl, alt, false)
}

func main() { // here because of some odd options we were experimentally adding
	fmt.Println("*puffcat*")
	err := keyboard.Open()
	if err != nil {
		fmt.Println(err)
		return
	}

	c := make(chan term.KeyEvent)
	r := Repl{
		dialect: "",
	}
	ml := term.NewMicroLiner(c, r.recieveMessage, r.recieveLine)
	r.ml = ml

	/* ml.SetCompleter(func(line string, mode int) (c []string) {
	// #IMPROV #IDEA words defined in current context should be bold
	// #IMPROV #Q how would we cycle just through words in current context?
	// #TODO don't display more than N words
	// #TODO make current word bold

	// # TRICK: we don't have the cursor position, but the caller code handles that already so we can suggest in the 	middle
	suggestions := make([]string, 0)
	var wordpart string
	spacePos := strings.LastIndex(line, " ")
	var prefix string
	if spacePos < 0 {
		fmt.Println("*")
		wordpart = line
		prefix = ""
	} else {
		wordpart = strings.TrimSpace(line[spacePos:])
		fmt.Print("=(")
		fmt.Print(wordpart)
		fmt.Print(")=")
		prefix = line[0:spacePos] + " "
		if wordpart == "" { // we are probably 1 space after last word
			fmt.Println("+")
			return
		}
	}

	fmt.Print("=[")
	fmt.Print(wordpart)
	fmt.Print("]=")
	switch mode {
	case 0:
		for i := 0; i < es.Idx.GetWordCount(); i++ {
			// fmt.Print(es.Idx.GetWord(i))
			if strings.HasPrefix(es.Idx.GetWord(i), strings.ToLower(wordpart)) {
				c = append(c, prefix+es.Idx.GetWord(i))
				suggestions = append(suggestions, es.Idx.GetWord(i))
			} else if strings.HasPrefix("."+es.Idx.GetWord(i), strings.ToLower(wordpart)) {
				c = append(c, prefix+"."+es.Idx.GetWord(i))
				suggestions = append(suggestions, es.Idx.GetWord(i))
			} else if strings.HasPrefix("|"+es.Idx.GetWord(i), strings.ToLower(wordpart)) {
				c = append(c, prefix+"|"+es.Idx.GetWord(i))
				suggestions = append(suggestions, es.Idx.GetWord(i))
			}
		}
	case 1:
		for key := range es.Ctx.GetState() {
			// fmt.Print(es.Idx.GetWord(i))
			if strings.HasPrefix(es.Idx.GetWord(key), strings.ToLower(wordpart)) {
				c = append(c, prefix+es.Idx.GetWord(key))
				suggestions = append(suggestions, es.Idx.GetWord(key))
			} else if strings.HasPrefix("."+es.Idx.GetWord(key), strings.ToLower(wordpart)) {
				c = append(c, prefix+"."+es.Idx.GetWord(key))
				suggestions = append(suggestions, es.Idx.GetWord(key))
			} else if strings.HasPrefix("|"+es.Idx.GetWord(key), strings.ToLower(wordpart)) {
				c = append(c, prefix+"|"+es.Idx.GetWord(key))
				suggestions = append(suggestions, es.Idx.GetWord(key))
			}
		}
	}
	*/

	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()

	// ctx := context.Background()
	// defer os.Exit(0)
	// defer ctx.Done()
	defer keyboard.Close()
	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				// fmt.Println("Done")
				return
			default:
				// fmt.Println("Select default")
				r, k, keyErr := keyboard.GetKey()
				if err != nil {
					fmt.Println(keyErr)
					break
				}
				if k == keyboard.KeyCtrlC {
					// fmt.Println("Ctrl C 1")
					cancel()
					err1 := util.KillProcess(os.Getpid())
					// err1 := syscall.Kill(os.Getpid(), syscall.SIGINT)
					if err1 != nil {
						fmt.Println(err.Error()) // TODO -- temprorary just printed
					}
					//ctx.Done()
					// fmt.Println("")
					// return
					//break
					//					os.Exit(0)
				}
				c <- constructKeyEvent(r, k)
			}
		}
	}(ctx)

	// fmt.Println("MICRO")
	_, err = ml.MicroPrompt("", "", 0, ctx)
	if err != nil {
		fmt.Println(err)
	}
}
