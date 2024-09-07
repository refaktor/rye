package util

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"unicode"

	"github.com/mattn/go-runewidth"
)

// These character classes are mostly zero width (when combined).
// A few might not be, depending on the user's font. Fixing this
// is non-trivial, given that some terminals don't support
// ANSI DSR/CPR
var zeroWidth = []*unicode.RangeTable{
	unicode.Mn,
	unicode.Me,
	unicode.Cc,
	unicode.Cf,
}

// countGlyphs considers zero-width characters to be zero glyphs wide,
// and members of Chinese, Japanese, and Korean scripts to be 2 glyphs wide.
func countGlyphs(s []rune) int {
	n := 0
	for _, r := range s {
		// speed up the common case
		if r < 127 {
			n++
			continue
		}

		n += runewidth.RuneWidth(r)
	}
	return n
}

func getPrefixGlyphs(s []rune, num int) []rune {
	p := 0
	for n := 0; n < num && p < len(s); p++ {
		// speed up the common case
		if s[p] < 127 {
			n++
			continue
		}
		if !unicode.IsOneOf(zeroWidth, s[p]) {
			n++
		}
	}
	for p < len(s) && unicode.IsOneOf(zeroWidth, s[p]) {
		p++
	}
	return s[:p]
}

func getSuffixGlyphs(s []rune, num int) []rune {
	p := len(s)
	for n := 0; n < num && p > 0; p-- {
		// speed up the common case
		if s[p-1] < 127 {
			n++
			continue
		}
		if !unicode.IsOneOf(zeroWidth, s[p-1]) {
			n++
		}
	}
	return s[p:]
}

type KeyEvent struct {
	Key   string
	Code  int
	Ctrl  bool
	Alt   bool
	Shift bool
}

func NewKeyEvent(key string, code int, ctrl bool, alt bool, shift bool) KeyEvent {
	return KeyEvent{key, code, ctrl, alt, shift}
}

func (s *MLState) cursorPos(x int) {
	// 'C' is "Cursor Forward (CUF)"
	s.sendBack("\r")
	if x > 0 {
		s.sendBack(fmt.Sprintf("\x1b[%dC", x))
	}
}

func (s *MLState) cursorPos2(x int, y int) {
	// 'C' is "Cursor Forward (CUF)"
	s.sendBack("\r")
	if x > 0 {
		trace("CURSOR POS:")
		trace(x)
		s.sendBack(fmt.Sprintf("\x1b[%dC", x))
	}
	if y > 0 {
		trace("CURSOR POS:")
		trace(y)
		s.sendBack(fmt.Sprintf("\x1b[%dA", y))
	}
}

func (s *MLState) eraseLine() {
	//str := fmt.Sprintf("\x1b[0K")
	// s.sendBack("\x1b[0K")
	s.sendBack("\x1b[2Kr")
}

func (s *MLState) doBeep() {
	//str := fmt.Sprintf("\x1b[0K")
	// s.sendBack("\x1b[0K")
	s.sendBack("\a")
}

func (s *MLState) eraseScreen() {
	str := "\x1b[H\x1b[2J"
	s.sendBack(str)
}

func (s *MLState) moveUp(lines int) {
	str := fmt.Sprintf("\x1b[%dA", lines)
	s.sendBack(str)
}

func (s *MLState) moveDown(lines int) {
	s.sendBack(fmt.Sprintf("\x1b[%dB", lines))
}

func (s *MLState) emitNewLine() {
	s.sendBack("\n")
}

// State represents an open terminal
type MLState struct {
	needRefresh    bool
	next           <-chan KeyEvent
	sendBack       func(msg string)
	enterLine      func(line string) string
	history        []string
	historyMutex   sync.RWMutex
	columns        int
	inString       bool
	lastLineString bool
	prevLines      int
	prevCursorLine int
	// killRing *ring.Ring
	//	completer         WordCompleter
	// pending     []rune
}

// NewLiner initializes a new *State, and sets the terminal into raw mode. To
// restore the terminal to its previous state, call State.Close().
func NewMicroLiner(ch chan KeyEvent, sb func(msg string), el func(line string) string) *MLState {
	var s MLState
	s.next = ch
	s.sendBack = sb
	s.enterLine = el
	//	s.r = bufio.NewReader(os.Stdin)
	return &s
}

func (s *MLState) getColumns() bool {
	s.columns = GetTerminalColumns()
	// fmt.Print("*getColumns* : ")
	// fmt.Println(s.columns)
	return true
}

// Redrawing input
// Called when it needs to redraw / refresh the current input, dispatches to single line and multiline

func (s *MLState) refresh(prompt []rune, buf []rune, pos int) error {
	s.needRefresh = false
	return s.refreshSingleLine(prompt, buf, pos)
}

// HISTORY

// HistoryLimit is the maximum number of entries saved in the scrollback history.
const HistoryLimit = 1000

// AppendHistory appends an entry to the scrollback history. AppendHistory
// should be called iff Prompt returns a valid command.
func (s *MLState) AppendHistory(item string) {
	s.historyMutex.Lock()
	defer s.historyMutex.Unlock()
	if len(s.history) > 0 {
		if item == s.history[len(s.history)-1] {
			return
		}
	}
	s.history = append(s.history, item)
	if len(s.history) > HistoryLimit {
		s.history = s.history[1:]
	}
}

// ClearHistory clears the scrollback history.
func (s *MLState) ClearHistory() {
	s.historyMutex.Lock()
	defer s.historyMutex.Unlock()
	s.history = nil
}

// Returns the history lines starting with prefix
func (s *MLState) getHistoryByPrefix(prefix string) (ph []string) {
	for _, h := range s.history {
		if strings.HasPrefix(h, prefix) {
			ph = append(ph, h)
		}
	}
	return
}

// Returns the history lines matching the intelligent search
func (s *MLState) getHistoryByPattern(pattern string) (ph []string, pos []int) {
	if pattern == "" {
		return
	}
	for _, h := range s.history {
		if i := strings.Index(h, pattern); i >= 0 {
			ph = append(ph, h)
			pos = append(pos, i)
		}
	}
	return
}

// END HISTORY

/*
// addToKillRing adds some text to the kill ring. If mode is 0 it adds it to a
// new node in the end of the kill ring, and move the current pointer to the new
// node. If mode is 1 or 2 it appends or prepends the text to the current entry
// of the killRing.
func (s *MLState) addToKillRing(text []rune, mode int) {
	// Don't use the same underlying array as text
	killLine := make([]rune, len(text))
	copy(killLine, text)

	// Point killRing to a newNode, procedure depends on the killring state and
	// append mode.
	if mode == 0 { // Add new node to killRing
		if s.killRing == nil { // if killring is empty, create a new one
			s.killRing = ring.New(1)
		} else if s.killRing.Len() >= KillRingMax { // if killring is "full"
			s.killRing = s.killRing.Next()
		} else { // Normal case
			s.killRing.Link(ring.New(1))
			s.killRing = s.killRing.Next()
		}
	} else {
		if s.killRing == nil { // if killring is empty, create a new one
			s.killRing = ring.New(1)
			s.killRing.Value = []rune{}
		}
		if mode == 1 { // Append to last entry
			killLine = append(s.killRing.Value.([]rune), killLine...)
		} else if mode == 2 { // Prepend to last entry
			killLine = append(killLine, s.killRing.Value.([]rune)...)
		}
	}

	// Save text in the current killring node
	s.killRing.Value = killLine
}

func (s *MLState) yank(p []rune, text []rune, pos int) ([]rune, int, interface{}, error) {
	if s.killRing == nil {
		return text, pos, rune(esc), nil
	}

	lineStart := text[:pos]
	lineEnd := text[pos:]
	var line []rune

	for {
		value := s.killRing.Value.([]rune)
		line = make([]rune, 0)
		line = append(line, lineStart...)
		line = append(line, value...)
		line = append(line, lineEnd...)

		pos = len(lineStart) + len(value)
		err := s.refresh(p, line, pos)
		if err != nil {
			return line, pos, 0, err
		}

		next, err := s.readNext()
		if err != nil {
			return line, pos, next, err
		}

		switch v := next.(type) {
		case rune:
			return line, pos, next, nil
		case action:
			switch v {
			case altY:
				s.killRing = s.killRing.Prev()
			default:
				return line, pos, next, nil
			}
		}
	}
}
*/

func traceTop(t any, n int) {
	if false {
		// Save cursor position
		fmt.Printf("\033[s")

		// Move cursor to top
		fmt.Printf("\033[H")
		// Move down
		if n > 0 {
			fmt.Printf("\033[%dB", n)
		}
		// Move cursor to top
		fmt.Printf("\033[K")

		// Print text
		fmt.Println(t)

		// Restore cursor position
		fmt.Printf("\033[u")
	}
}

func trace(t any) {
	if false {
		fmt.Println(t)
	}
}

func trace2(t any) {
	if false {
		fmt.Println(t)
	}
}

func splitText(text string, splitLength int) []string {
	result := []string{}
	current := bytes.NewBufferString("")
	for _, char := range text {
		if current.Len() == splitLength {
			result = append(result, current.String())
			current.Reset()
		}
		current.WriteRune(char)
	}
	if current.Len() > 0 {
		result = append(result, current.String())
	}
	return result
}

func splitText2(text string, splitLength int) []string {
	result := []string{}
	current := bytes.NewBufferString("")
	for _, char := range text {
		current.WriteRune(char)
		if current.Len() == splitLength {
			result = append(result, current.String())
			traceTop("appending", 3)
			current.Reset()
		}
	}
	//if current.Len() > 0 {
	result = append(result, current.String())
	//}
	return result
}

func (s *MLState) refreshSingleLine(prompt []rune, buf []rune, pos int) error {
	traceTop(pos, 0)
	// s.sendBack("\033[?25l") // hide cursors
	/// s.cursorPos(0)
	/// s.sendBack("\033[K")
	/// s.cursorPos(0)
	// s.sendBack(string(prompt))

	pLen := countGlyphs(prompt)
	// bLen := countGlyphs(buf)
	// on some OS / terminals extra column is needed to place the cursor char
	///// pos = countGlyphs(buf[:pos])

	// bLen := countGlyphs(buf)
	// on some OS / terminals extra column is needed to place the cursor char
	/*	if cursorColumn {
		bLen++
	}*/
	cols := s.columns - 6
	text := string(buf)
	texts := splitText2(text, cols)

	traceTop(len(texts), 0)

	// inString := false
	// text2 := wordwrap.String(text, 5)
	s.cursorPos(0)
	s.sendBack("\033[K") // delete line
	for i := 0; i < s.prevLines; i++ {
		//if i > 0 && len(texts[i]) > 1 {
		if i > 0 {
			s.cursorPos(0)
			s.sendBack("\033[A")
			s.sendBack("\033[K") // delete line
		}
	}
	// s.sendBack("\033[s")
	s.sendBack(string(prompt))
	for i, tt := range texts {
		if i > 0 {
			s.sendBack("\nx  ")
		}
		// tt2, inString := tt, false // RyeHighlight(tt, s.lastLineString, 6)
		tt2, inString := RyeHighlight(tt, s.lastLineString, cols)
		s.sendBack(tt2)
		s.inString = inString
	}

	s.prevLines = len(texts)

	/* text, inString := RyeHighlight(string(buf), s.lastLineString)
	s.sendBack(text)
	trace("*************** IN STRING: ******************++")dlk
	trace(inString)
	s.inString = inString
	trace(pLen + pos)
	trace("SETTING CURSOR POS AFER HIGHLIGHT") */
	curLineN := pos / cols
	curLines := len(text) / cols
	traceTop(curLineN, 1)
	traceTop(curLines, 2)
	curUp := curLineN - s.prevCursorLine
	curLeft := pLen + ((pos) % cols)
	// traceTop("---", 3)
	traceTop(curUp, 4)
	traceTop(curLeft, 5)
	// s.sendBack("\033[u")

	s.cursorPos2(curLeft, 0) // s.prevCursorLine-curLineN)
	s.prevCursorLine = curLineN
	//s.sendBack("\033[?25h") // show cursor
	return nil
}

// signals end-of-file by pressing Ctrl-D.
func (s *MLState) MicroPrompt(prompt string, text string, pos int, ctx1 context.Context) (string, error) {
	// history related
	historyEnd := ""
	var historyPrefix []string
	historyPos := 0
	historyStale := true
	// historyAction := false // used to mark history related actions
	// killAction := 0        // used to mark kill related actions
	multiline := false
startOfHere:

	var p []rune
	var line = []rune(text)
	if !multiline {
		s.sendBack(prompt)
		p = []rune(prompt)
	} else {
		s.sendBack("   ")
		p = []rune("   ")
		multiline = false
	}

	// defer s.stopPrompt()

	// if negative or past end put to the end
	if pos < 0 || len(line) < pos {
		pos = len(line)
	}
	// if len of line is > 0 then refresh
	if len(line) > 0 {
		err := s.refresh(p, line, pos)
		if err != nil {
			return "", err
		}
	}
	// var next string

	// LBL restart:
	//	s.startPrompt()
	//	s.getColumns()
	s.getColumns()
	traceTop(strconv.Itoa(s.columns)+"**", 0)

	histPrev := func() {
		if historyStale {
			historyPrefix = s.getHistoryByPrefix(string(line))
			historyPos = len(historyPrefix)
			historyStale = false
		}
		if historyPos > 0 {
			if historyPos == len(historyPrefix) {
				historyEnd = string(line)
			}
			historyPos--
			line = []rune(historyPrefix[historyPos])
			pos = len(line)
			s.needRefresh = true
		} else {
			s.doBeep()
		}
	}

	histNext := func() {
		if historyStale {
			historyPrefix = s.getHistoryByPrefix(string(line))
			historyPos = len(historyPrefix)
			historyStale = false
		}
		if historyPos < len(historyPrefix) {
			historyPos++
			if historyPos == len(historyPrefix) {
				line = []rune(historyEnd)
			} else {
				line = []rune(historyPrefix[historyPos])
			}
			pos = len(line)
			s.needRefresh = true
		} else {
			s.doBeep()
		}
	}

	// JM
	//	s_instr := 0

	// mainLoop:
	for {
		select {
		case <-ctx1.Done():
			// fmt.Println("Exitin due to coancelation")
			return "", nil
		default:
			trace("POS: ")
			trace(pos)
			// receive next character from channel
			next := <-s.next
			// s.sendBack(next)
			// err := nil
			// LBL haveNext:
			/* if err != nil {
				if s.shouldRestart != nil && s.shouldRestart(err) {
					goto restart
				}
				return "", err
			}*/

			// historyAction = false
			//switch v := next.(type) {
			//case string:
			//}
			/* if pos == len(line) && !s.multiLineMode &&
			len(p)+len(line) < s.columns*4 && // Avoid countGlyphs on large lines
			countGlyphs(p)+countGlyphs(line) < s.columns-1 {*/
			///// pLen := countGlyphs(p)
			if next.Ctrl {
				switch strings.ToLower(next.Key) {
				case "c":
					/* return "", ErrPromptAborted
					line = line[:0]
					pos = 0
					s.restartPrompt() */
					// fmt.Print("case C")
					return "", nil
				case "a":
					pos = 0
					// s.needRefresh = true
				case "e":
					pos = len(line)
					// s.needRefresh = true
				case "b":
					if pos > 0 {
						pos -= len(getSuffixGlyphs(line[:pos], 1))
						//s.needRefresh = true
					} else {
						s.doBeep()
					}
				case "f": // right
					if pos < len(line) {
						pos += len(getPrefixGlyphs(line[pos:], 1))
						// s.needRefresh = true
					} else {
						s.doBeep()
					}
				case "k": // delete remainder of line
					if pos >= len(line) {
						// s.doBeep()
					} else {
						// if killAction > 0 {
						//	s.addToKillRing(line[pos:], 1) // Add in apend mode
						// } else {
						//	s.addToKillRing(line[pos:], 0) // Add in normal mode
						// }
						// killAction = 2 // Mark that there was a kill action
						line = line[:pos]
						s.needRefresh = true
					}
				case "l":
					s.eraseScreen()
					s.needRefresh = true
				case "u": // delete to beginning of line
					if pos == 0 {
						s.doBeep()
					} else {
						line = line[pos:]
						pos = 0
						s.needRefresh = true
					}
				case "n":
					histNext()
				case "p":
					histPrev()
				}
			} else if next.Alt {
				switch strings.ToLower(next.Key) {
				case "b":
					if pos > 0 {
						var spaceHere, spaceLeft, leftKnown bool
						for {
							pos--
							if pos == 0 {
								break
							}
							if leftKnown {
								spaceHere = spaceLeft
							} else {
								spaceHere = unicode.IsSpace(line[pos])
							}
							spaceLeft, leftKnown = unicode.IsSpace(line[pos-1]), true
							if !spaceHere && spaceLeft {
								break
							}
						}
					} else {
						s.doBeep()
					}
				case "f":
					if pos < len(line) {
						var spaceHere, spaceLeft, hereKnown bool
						for {
							pos++
							if pos == len(line) {
								break
							}
							if hereKnown {
								spaceLeft = spaceHere
							} else {
								spaceLeft = unicode.IsSpace(line[pos-1])
							}
							spaceHere, hereKnown = unicode.IsSpace(line[pos]), true
							if spaceHere && !spaceLeft {
								break
							}
						}
					} else {
						s.doBeep()
					}
				case "d": // Delete next word
					if pos == len(line) {
						s.doBeep()
						break
					}
					// Remove whitespace to the right
					var buf []rune // Store the deleted chars in a buffer
					for {
						if pos == len(line) || !unicode.IsSpace(line[pos]) {
							break
						}
						buf = append(buf, line[pos])
						line = append(line[:pos], line[pos+1:]...)
					}
					// Remove non-whitespace to the right
					for {
						if pos == len(line) || unicode.IsSpace(line[pos]) {
							break
						}
						buf = append(buf, line[pos])
						line = append(line[:pos], line[pos+1:]...)
						trace(buf)
					}
					s.needRefresh = true
					// Save the result on the killRing
					/*if killAction > 0 {
						s.addToKillRing(buf, 2) // Add in prepend mode
					} else {
						s.addToKillRing(buf, 0) // Add in normal mode
					} */
					// killAction = 2 // Mark that there was some killing
					//			case "bs": // Erase word
					//				pos, line, killAction = s.eraseWord(pos, line, killAction)
				}
			} else {
				switch next.Code {
				case 13: // Enter
					historyStale = true
					s.lastLineString = false
					// trace2("NL")
					if len(line) > 0 && unicode.IsSpace(line[len(line)-1]) {
						s.sendBack(fmt.Sprintf("%s⏎\n%s", color_emph, reset))
						if s.inString {
							s.lastLineString = true
						}
					} else {
						s.sendBack("\n")
					}
					xx := s.enterLine(string(line))
					pos = 0
					if xx == "next line" {
						multiline = true
					} else {
						s.sendBack("") // WW?
					}
					line = make([]rune, 0)
					trace(line)
					goto startOfHere
				case 8: // Backspace
					if pos <= 0 {
						s.doBeep()
					} else {
						// pos += 1
						n := len(getSuffixGlyphs(line[:pos], 1))
						trace("<---line--->")
						trace(line[:pos-n])
						trace(line[pos:])
						trace(n)
						trace(pos)
						trace(line)
						// line = append(line[:pos-n], ' ')
						line = append(line[:pos-n], line[pos:]...)
						//						line = line[:pos-1]
						trace(line)
						// line = append(line[:pos-n], line[pos:]...)
						pos -= n
						s.needRefresh = true
					}
				case 46: // Del
					if pos >= len(line) {
						s.doBeep()
					} else {
						n := len(getPrefixGlyphs(line[pos:], 1))
						line = append(line[:pos], line[pos+n:]...)
						s.needRefresh = true
					}
				case 39: // Right
					if pos < len(line) {
						pos += len(getPrefixGlyphs(line[pos:], 1))
					} else {
						s.doBeep()
					}
				case 37: // Left
					if pos > 0 {
						pos -= len(getSuffixGlyphs(line[:pos], 1))
						traceTop(pos, 3)
					} else {
						s.doBeep()
					}
				case 38: // Up
					histPrev()
				case 40: // Down
					histNext()
				case 36: // Home
					pos = 0
				case 35: // End
					pos = len(line)
				default:
					trace("***************************** ALARM *******************")
					vs := []rune(next.Key)
					v := vs[0]

					if pos >= countGlyphs(p)+countGlyphs(line) {
						line = append(line, v)
						//s.sendBack(fmt.Sprintf("%c", v))
						s.needRefresh = true // JM ---
						pos++
					} else {
						line = append(line[:pos], append([]rune{v}, line[pos:]...)...)
						pos++
						s.needRefresh = true
					}
				}
			}

			/* } else {
				line = append(line[:pos], append([]rune{v}, line[pos:]...)...)
				pos++
				s.needRefresh = true
			} */

			/* case rune:
				switch v {
				case cr, lf:
					if s.needRefresh {
						err := s.refresh(p, line, pos)
						if err != nil {
							return "", err
						}
					}
					if s.multiLineMode {
						s.resetMultiLine(p, line, pos)
					}
					trace()
					break mainLoop
				case ctrlA: // Start of line
					pos = 0
					s.needRefresh = true
				case ctrlE: // End of line
					pos = len(line)
					s.needRefresh = true
				case ctrlB: // left
					if pos > 0 {
						pos -= len(getSuffixGlyphs(line[:pos], 1))
						s.needRefresh = true
					} else {
						s.doBeep()
					}
				case ctrlF: // right
					if pos < len(line) {
						pos += len(getPrefixGlyphs(line[pos:], 1))
						s.needRefresh = true
					} else {
						s.doBeep()
					}
				case ctrlD: // del
					if pos == 0 && len(line) == 0 {
						// exit
						return "", io.EOF
					}

					// ctrlD is a potential EOF, so the rune reader shuts down.
					// Therefore, if it isn't actually an EOF, we must re-startPrompt.
					s.restartPrompt()

					if pos >= len(line) {
						s.doBeep()
					} else {
						n := len(getPrefixGlyphs(line[pos:], 1))
						line = append(line[:pos], line[pos+n:]...)
						s.needRefresh = true
					}
				case ctrlK: // delete remainder of line
					if pos >= len(line) {
						s.doBeep()
					} else {
						if killAction > 0 {
							s.addToKillRing(line[pos:], 1) // Add in apend mode
						} else {
							s.addToKillRing(line[pos:], 0) // Add in normal mode
						}

						killAction = 2 // Mark that there was a kill action
						line = line[:pos]
						s.needRefresh = true
					}
				case ctrlP: // up
					historyAction = true
					if historyStale {
						historyPrefix = s.getHistoryByPrefix(string(line))
						historyPos = len(historyPrefix)
						historyStale = false
					}
					if historyPos > 0 {
						if historyPos == len(historyPrefix) {
							historyEnd = string(line)
						}
						historyPos--
						line = []rune(historyPrefix[historyPos])
						pos = len(line)
						s.needRefresh = true
					} else {
						s.doBeep()
					}
				case ctrlN: // down
					historyAction = true
					if historyStale {
						historyPrefix = s.getHistoryByPrefix(string(line))
						historyPos = len(historyPrefix)
						historyStale = false
					}
					if historyPos < len(historyPrefix) {
						historyPos++
						if historyPos == len(historyPrefix) {
							line = []rune(historyEnd)
						} else {
							line = []rune(historyPrefix[historyPos])
						}
						pos = len(line)
						s.needRefresh = true
					} else {
						s.doBeep()
					}
				case ctrlT: // transpose prev glyph with glyph under cursor
					if len(line) < 2 || pos < 1 {
						s.doBeep()
					} else {
						if pos == len(line) {
							pos -= len(getSuffixGlyphs(line, 1))
						}
						prev := getSuffixGlyphs(line[:pos], 1)
						next := getPrefixGlyphs(line[pos:], 1)
						scratch := make([]rune, len(prev))
						copy(scratch, prev)
						copy(line[pos-len(prev):], next)
						copy(line[pos-len(prev)+len(next):], scratch)
						pos += len(next)
						s.needRefresh = true
					}
				case ctrlL: // clear screen
					s.eraseScreen()
					s.needRefresh = true
				case ctrlC: // reset
					trace("^C")
					if s.multiLineMode {
						s.resetMultiLine(p, line, pos)
					}
					if s.ctrlCAborts {
						return "", ErrPromptAborted
					}
					line = line[:0]
					pos = 0
					fmt.Print(prompt)
					s.restartPrompt()
				case ctrlH, bs: // Backspace
					if pos <= 0 {
						s.doBeep()
					} else {
						n := len(getSuffixGlyphs(line[:pos], 1))
						line = append(line[:pos-n], line[pos:]...)
						pos -= n
						s.needRefresh = true
					}
				case ctrlU: // Erase line before cursor
					if killAction > 0 {
						s.addToKillRing(line[:pos], 2) // Add in prepend mode
					} else {
						s.addToKillRing(line[:pos], 0) // Add in normal mode
					}

					killAction = 2 // Mark that there was some killing
					line = line[pos:]
					pos = 0
					s.needRefresh = true
				case ctrlW: // Erase word
					pos, line, killAction = s.eraseWord(pos, line, killAction)
				case ctrlY: // Paste from Yank buffer
					line, pos, next, err = s.yank(p, line, pos)
					goto haveNext
				case ctrlR: // Reverse Search
					line, pos, next, err = s.reverseISearch(line, pos)
					s.needRefresh = true
					goto haveNext
				case tab: // Tab completion
					line, pos, next, err = s.tabComplete(p, line, pos)
					goto haveNext
				// Catch keys that do nothing, but you don't want them to beep
				case esc:
					// DO NOTHING
				// Unused keys
				case ctrlG:
					// JM experimenting 20200108
					//for _, l := range codelines {
					//		MoveCursorDown(2)
					//		fmt.Print(l)
					//	}
					//MoveCursorDown(len(codelines))
					return "", ErrJMCodeUp
					//line = []rune(codelines[len(codelines)-1])
					//pos = len(line)
					s.needRefresh = true
				case ctrlS, ctrlO, ctrlQ, ctrlV, ctrlX, ctrlZ:
					fallthrough
				// Catch unhandled control codes (anything <= 31)
				case 0, 28, 29, 30, 31:
					s.doBeep()
				default:
					if pos == len(line) && !s.multiLineMode &&
						len(p)+len(line) < s.columns*4 && // Avoid countGlyphs on large lines
						countGlyphs(p)+countGlyphs(line) < s.columns-1 {
						line = append(line, v)
						fmt.Printf("%c", v)
						s.needRefresh = true // JM ---
						pos++

					} else {
						line = append(line[:pos], append([]rune{v}, line[pos:]...)...)
						pos++
						s.needRefresh = true
					}
					if s_instr == 2 && string(v) == "\"" {
						s_instr = 0
					}
				}
			case action:
				switch v {
				case del:
					if pos >= len(line) {
						s.doBeep()
					} else {
						n := len(getPrefixGlyphs(line[pos:], 1))
						line = append(line[:pos], line[pos+n:]...)
					}
				case left:
					if pos > 0 {
						pos -= len(getSuffixGlyphs(line[:pos], 1))
					} else {
						s.doBeep()
					}
				case wordLeft, altB:
					if pos > 0 {
						var spaceHere, spaceLeft, leftKnown bool
						for {
							pos--
							if pos == 0 {
								break
							}
							if leftKnown {
								spaceHere = spaceLeft
							} else {
								spaceHere = unicode.IsSpace(line[pos])
							}
							spaceLeft, leftKnown = unicode.IsSpace(line[pos-1]), true
							if !spaceHere && spaceLeft {
								break
							}
						}
					} else {
						s.doBeep()
					}
				case right:
					if pos < len(line) {
						pos += len(getPrefixGlyphs(line[pos:], 1))
					} else {
						s.doBeep()
					}
				case wordRight, altF:
					if pos < len(line) {
						var spaceHere, spaceLeft, hereKnown bool
						for {
							pos++
							if pos == len(line) {
								break
							}
							if hereKnown {
								spaceLeft = spaceHere
							} else {
								spaceLeft = unicode.IsSpace(line[pos-1])
							}
							spaceHere, hereKnown = unicode.IsSpace(line[pos]), true
							if spaceHere && !spaceLeft {
								break
							}
						}
					} else {
						s.doBeep()
					}
				case up:
					historyAction = true
					if historyStale {
						historyPrefix = s.getHistoryByPrefix(string(line))
						historyPos = len(historyPrefix)
						historyStale = false
					}
					if historyPos > 0 {
						if historyPos == len(historyPrefix) {
							historyEnd = string(line)
						}
						historyPos--
						line = []rune(historyPrefix[historyPos])
						pos = len(line)
					} else {
						s.doBeep()
					}
				case down:
					historyAction = true
					if historyStale {
						historyPrefix = s.getHistoryByPrefix(string(line))
						historyPos = len(historyPrefix)
						historyStale = false
					}
					if historyPos < len(historyPrefix) {
						historyPos++
						if historyPos == len(historyPrefix) {
							line = []rune(historyEnd)
						} else {
							line = []rune(historyPrefix[historyPos])
						}
						pos = len(line)
					} else {
						s.doBeep()
					}
				case home: // Start of line
					pos = 0
				case end: // End of line
					pos = len(line)
				case altD: // Delete next word
					if pos == len(line) {
						s.doBeep()
						break
					}
					// Remove whitespace to the right
					var buf []rune // Store the deleted chars in a buffer
					for {
						if pos == len(line) || !unicode.IsSpace(line[pos]) {
							break
						}
						buf = append(buf, line[pos])
						line = append(line[:pos], line[pos+1:]...)
					}
					// Remove non-whitespace to the right
					for {
						if pos == len(line) || unicode.IsSpace(line[pos]) {
							break
						}
						buf = append(buf, line[pos])
						line = append(line[:pos], line[pos+1:]...)
					}
					// Save the result on the killRing
					if killAction > 0 {
						s.addToKillRing(buf, 2) // Add in prepend mode
					} else {
						s.addToKillRing(buf, 0) // Add in normal mode
					}
					killAction = 2 // Mark that there was some killing
				case altBs: // Erase word
					pos, line, killAction = s.eraseWord(pos, line, killAction)
				case winch: // Window change
					if s.multiLineMode {
						if s.maxRows-s.cursorRows > 0 {
							s.moveDown(s.maxRows - s.cursorRows)
						}
						for i := 0; i < s.maxRows-1; i++ {
							s.cursorPos(0)
							s.eraseLine()
							s.moveUp(1)
						}
						s.maxRows = 1
						s.cursorRows = 1
					}
				}
				s.needRefresh = true
			} */
			//if true || s.needRefresh { //&& !s.inputWaiting() {
			// ALWAYS REFRESH SO WE HAVE JUST ONE TRUTH
			err := s.refresh(p, line, pos)
			if err != nil {
				return "", err
			}
			// } else {
			///// s.cursorPos(pLen + pos)
			// }
			/*if !historyAction {
				historyStale = true
			}
			if killAction > 0 {
				killAction--
			}*/
		}
	}
	// return string(line), nil
}

const bright = "\x1b[1m"
const dim = "\x1b[2m"
const black = "\x1b[30m"
const red = "\x1b[31m"
const green = "\x1b[32m"
const yellow = "\x1b[33m"
const blue = "\x1b[34m"
const magenta = "\x1b[35m"
const cyan = "\x1b[36m"
const white = "\x1b[37m"
const reset = "\x1b[0m"
const reset2 = "\033[39;49m"

const color_word1 = cyan
const color_word2 = yellow
const color_num2 = magenta
const color_string2 = green
const color_comment = dim + white
const color_emph = bright

type HighlightedStringBuilder struct {
	b strings.Builder
}

func (h *HighlightedStringBuilder) WriteRune(c rune) {
	h.b.WriteRune(c)
}

func (h *HighlightedStringBuilder) String() string {
	return h.b.String()
}

func (h *HighlightedStringBuilder) ColoredString(inStr bool) string {
	return h.getColor(inStr) + h.b.String() + reset
}

func (h *HighlightedStringBuilder) Reset() {
	h.b.Reset()
}

func (h *HighlightedStringBuilder) getColor(inStr bool) string {
	s := h.b.String()
	if len(s) == 0 {
		return ""
	}
	if strings.HasPrefix(s, ";") {
		return color_comment
	}
	if inStr || hasPrefixMultiple(s, "\"", "`") {
		return color_string2
	}
	if strings.HasPrefix(s, "%") && len(s) != 1 {
		return color_string2
	}
	if hasPrefixMultiple(s, "?", "~", "|", "\\", ".", "'", "<") {
		if len(s) != 1 {
			return color_word2
		}
	}
	if strings.HasPrefix(s, ":") {
		if strings.HasPrefix(s, "::") {
			if len(s) != 2 {
				return color_emph + color_word1
			}
		} else if len(s) != 1 {
			return color_word1
		}
	}
	if strings.HasSuffix(s, ":") {
		if strings.HasSuffix(s, "::") {
			if len(s) != 2 {
				return color_emph + color_word1
			}
		} else if len(s) != 1 {
			return color_word1
		}
	}
	if unicode.IsNumber(rune(s[0])) {
		return color_num2
	}
	if unicode.IsLetter(rune(s[0])) {
		if strings.Contains(s, "://") {
			return color_string2
		}
		if strings.HasSuffix(s, "!") || strings.HasPrefix(s, "set-") {
			return color_emph + color_word2
		}
		return color_word2
	}
	return ""
}

func hasPrefixMultiple(s string, prefixes ...string) bool {
	for _, p := range prefixes {
		if strings.HasPrefix(s, p) {
			return true
		}
	}
	return false
}

func RyeHighlight(s string, inStrX bool, columns int) (string, bool) {
	var fullB strings.Builder
	var hb HighlightedStringBuilder

	var inComment, inStr1, inStr2 bool
	inStr1 = inStrX

	for _, c := range s {
		//if (i+2)%columns == 0 {
		//	hb.WriteRune('\n')
		// hb.WriteRune('\r')
		// }
		if inComment {
			hb.WriteRune(c)
		} else if c == ';' && !inStr1 && !inStr2 {
			inComment = true
			hb.WriteRune(c)
		} else if c == '"' {
			hb.WriteRune(c)
			if inStr1 {
				// trace2(".")
				fullB.WriteString(hb.ColoredString(inStr1))
				inStr1 = false
				hb.Reset()
			} else {
				inStr1 = true
			}
		} else if c == '`' {
			hb.WriteRune(c)
			if inStr2 {
				inStr2 = false
				fullB.WriteString(hb.ColoredString(inStr1))
				hb.Reset()
			} else {
				inStr2 = true
			}
		} else if unicode.IsSpace(c) && !inComment && !inStr1 && !inStr2 {
			fullB.WriteString(hb.ColoredString(inStr1))
			hb.Reset()

			fullB.WriteRune(c)
		} else {
			hb.WriteRune(c)
		}
	}
	fullB.WriteString(hb.ColoredString(inStr1))
	hb.Reset()
	return fullB.String(), inStr1
}

func RyeHighlight_OLD1(s string, inStrX bool) (string, bool) {
	var fullB strings.Builder
	var hb HighlightedStringBuilder

	var inComment, inStr1, inStr2 bool
	inStr1 = inStrX

	for _, c := range s {
		if inComment {
			hb.WriteRune(c)
		} else if c == ';' && !inStr1 && !inStr2 {
			inComment = true
			hb.WriteRune(c)
		} else if c == '"' {
			hb.WriteRune(c)
			if inStr1 {
				// trace2(".")
				fullB.WriteString(hb.ColoredString(inStr1))
				inStr1 = false
				hb.Reset()
			} else {
				inStr1 = true
			}
		} else if c == '`' {
			hb.WriteRune(c)
			if inStr2 {
				inStr2 = false
				fullB.WriteString(hb.ColoredString(inStr1))
				hb.Reset()
			} else {
				inStr2 = true
			}
		} else if unicode.IsSpace(c) && !inComment && !inStr1 && !inStr2 {
			fullB.WriteString(hb.ColoredString(inStr1))
			hb.Reset()

			fullB.WriteRune(c)
		} else {
			hb.WriteRune(c)
		}
	}
	fullB.WriteString(hb.ColoredString(inStr1))
	hb.Reset()
	return fullB.String(), inStr1
}
