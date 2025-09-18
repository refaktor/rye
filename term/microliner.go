package term

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"

	"github.com/mattn/go-runewidth"
	"github.com/refaktor/rye/env"
	// "github.com/refaktor/rye/term"
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

// HISTORY

// ReadHistory reads history entries from an io.Reader and adds them to the history buffer.
// It returns the number of entries successfully read and any error encountered.
// If an error occurs, some entries may have been added to the history buffer.
func (s *MLState) ReadHistory(r io.Reader) (num int, err error) {
	if r == nil {
		return 0, fmt.Errorf("reader cannot be nil")
	}

	s.historyMutex.Lock()
	defer s.historyMutex.Unlock()

	in := bufio.NewReader(r)
	num = 0
	for {
		line, part, err := in.ReadLine()
		if err == io.EOF {
			break
		}
		if err != nil {
			return num, fmt.Errorf("error reading history at line %d: %w", num+1, err)
		}
		if part {
			return num, fmt.Errorf("line %d is too long", num+1)
		}
		if !utf8.Valid(line) {
			return num, fmt.Errorf("invalid UTF-8 string at line %d", num+1)
		}
		num++
		s.history = append(s.history, string(line))
		if len(s.history) > HistoryLimit {
			s.history = s.history[1:]
		}
	}
	return num, nil
}

// WriteHistory writes scrollback history to w. Returns the number of lines
// successfully written, and any write error.
//
// Unlike the rest of liner's API, WriteHistory is safe to call
// from another goroutine while Prompt is in progress.
// This exception is to facilitate the saving of the history buffer
// during an unexpected exit (for example, due to Ctrl-C being invoked)
func (s *MLState) WriteHistory(w io.Writer) (num int, err error) {
	if w == nil {
		return 0, fmt.Errorf("writer cannot be nil")
	}

	s.historyMutex.RLock()
	defer s.historyMutex.RUnlock()

	for _, item := range s.history {
		_, err := fmt.Fprintln(w, item)
		if err != nil {
			return num, fmt.Errorf("error writing history at line %d: %w", num+1, err)
		}
		num++
	}
	return num, nil
}

// TAB COMPLETER

func (s *MLState) circularTabs(items []string) func(direction int) (string, int, error) {
	item := -1
	return func(direction int) (string, int, error) {
		if direction == 1 {
			if item < len(items)-1 {
				item++
			} else {
				item = 0
			}
		}
		return items[item], item, nil
	}
}

// displayTabSuggestions shows the tab completion suggestions with the current selection highlighted
func (s *MLState) displayTabSuggestions(items []string, currentIndex int) {
	if len(items) == 0 {
		return
	}

	// Save cursor position
	s.sendBack("\033[s")

	// Disable line wrapping to prevent suggestions from wrapping
	s.sendBack("\033[?7l")

	// Move cursor down to display suggestions
	s.sendBack("\n")
	s.sendBack("\033[K") // Clear line

	// Display suggestions with highlighting
	s.sendBack("\033[34m") // Magenta color
	s.sendBack("current: ")
	s.sendBack("\033[0m") // Reset color

	for i, item := range items {
		if i == currentIndex {
			// Highlight current selection with magenta background and black text
			s.sendBack("\033[45;30m") // Magenta background, black text
			s.sendBack(" ")
			s.sendBack(item)
			s.sendBack(" ")
			s.sendBack("\033[0m") // Reset formatting
		} else {
			// Non-selected items with magenta text
			s.sendBack("\033[36m") // Magenta text
			s.sendBack(" ")
			s.sendBack(item)
			s.sendBack(" ")
			s.sendBack("\033[0m") // Reset formatting
		}

		if i < len(items)-1 {
			s.sendBack("  ")
		}
	}

	// Display probe/preview of the currently selected item
	if currentIndex >= 0 && currentIndex < len(items) {
		s.sendBack("\n")
		s.sendBack("\033[K") // Clear line
		// Get a preview/description of the selected item
		probe := s.getItemProbe(items[currentIndex])
		s.sendBack("\033[38;5;247m")
		s.sendBack(probe)
		s.sendBack("\033[0m") // Reset color
	}

	// Re-enable line wrapping and restore cursor position
	s.sendBack("\033[?7h")
	s.sendBack("\033[u")
}

// getItemProbe returns a preview/description of the given item by looking it up in the environment
func (s *MLState) getItemProbe(item string) string {
	// If we don't have access to the program state, fall back to simple categorization
	if s.programState == nil {
		return "xxx"
	}

	// Look up the word in the index
	wordIndex, found := s.programState.Idx.GetIndex(item)
	if !found {
		return fmt.Sprintf("undefined word: %s", item)
	}

	// Try to get the object from the context
	obj, exists := s.programState.Ctx.Get(wordIndex)
	if !exists {
		return fmt.Sprintf("unbound word: %s", item)
	}

	// Call the Inspect method on the object to get detailed information
	return obj.Inspect(*s.programState.Idx)
}

// getItemProbeSimple provides fallback categorization when program state is not available
func (s *MLState) getItemProbeSimple(item string) string {
	switch {
	case strings.HasSuffix(item, "?"):
		return "predicate function"
	case strings.Contains(item, "print"):
		return "output function"
	case strings.Contains(item, "get"):
		return "accessor function"
	case strings.Contains(item, "set"):
		return "mutator function"
	case strings.Contains(item, "new"):
		return "constructor function"
	case strings.Contains(item, "load"):
		return "loader function"
	case strings.Contains(item, "save"):
		return "persistence function"
	default:
		return fmt.Sprintf("word: %s", item)
	}
}

// tabComplete handles tab completion functionality.
// It returns the completed line, new cursor position, next key event, and any error.
func (s *MLState) tabComplete(p []rune, line []rune, pos int, mode int) ([]rune, int, KeyEvent, error) {
	// if no completer defined
	if s.completer == nil {
		return line, pos, KeyEvent{Code: 27}, nil
	}

	// Run the completer
	head, list, tail := s.completer(string(line), pos, mode)
	if len(list) <= 0 {
		return line, pos, KeyEvent{Code: 27}, nil
	}

	// If there is one result, use it immediately
	hl := utf8.RuneCountInString(head)
	if len(list) == 1 {
		completedLine := []rune(head + list[0] + tail)
		newPos := hl + utf8.RuneCountInString(list[0])
		err := s.refresh(p, completedLine, newPos)
		if err != nil {
			return line, pos, KeyEvent{Code: 27}, fmt.Errorf("failed to refresh display: %w", err)
		}
		return completedLine, newPos, KeyEvent{Code: 27}, nil
	}

	// Handle multiple completion options
	direction := 1
	tabPrinter := s.circularTabs(list)
	for {
		pick, currentIndex, err := tabPrinter(direction)
		if err != nil {
			return line, pos, KeyEvent{Code: 27}, fmt.Errorf("tab completion error: %w", err)
		}

		completedLine := []rune(head + pick + tail)
		newPos := hl + utf8.RuneCountInString(pick)
		err = s.refresh(p, completedLine, newPos)
		if err != nil {
			return line, pos, KeyEvent{Code: 27}, fmt.Errorf("failed to refresh display: %w", err)
		}

		// Display suggestions with current selection highlighted
		s.displayTabSuggestions(list, currentIndex)

		// Wait for next key input
		next := <-s.next

		if next.Code == 9 {
			direction = 1
			continue
		}
		if next.Code == 27 {
			return line, pos, KeyEvent{Code: 27}, nil
		}
		return completedLine, newPos, next, nil
	}

	/*
		direction := tabForward

		if s.tabStyle == TabPrints {
			tabPrinter = s.printedTabs(list)
		}

		for {
			pick, err := tabPrinter(direction)
			if err != nil {
				return line, pos, rune(27), err
			}
			err = s.refresh(p, []rune(head+pick+tail), hl+utf8.RuneCountInString(pick))
			if err != nil {
				return line, pos, rune(27), err
			}

			next, err := s.readNext()
			if err != nil {
				return line, pos, rune(27), err
			}
			if key, ok := next.(rune); ok {
				if key == 9 {
					direction = tabForward
					continue
				}
				if key == 27 {
					return line, pos, rune(27), nil
				}
			}
			if a, ok := next.(action); ok && a == shiftTab {
				direction = tabReverse
				continue
			}
			return []rune(head + pick + tail), hl + utf8.RuneCountInString(pick), next, nil
		} */
}

// Completer takes the currently edited line content at the left of the cursor
// to the completer which may return {"Hello, world", "Hello, Word"} to have "Hello, world!!!".
type Completer func(line string, mode int) []string

// WordCompleter takes the currently edited line with the cursor position and
// to the completer which may returns ("Hello, ", {"world", "Word"}, "!!!") to have "Hello, world!!!".
type WordCompleter func(line string, pos int, mode int) (head string, completions []string, tail string)

// SetCompleter sets the completion function that Liner will call to
// fetch completion candidates when the user presses tab.
func (s *MLState) SetCompleter(f Completer) {
	if f == nil {
		s.completer = nil
		return
	}

	s.completer = func(line string, pos int, mode int) (string, []string, string) {
		if pos < 0 || pos > len(line) {
			// Handle invalid position safely
			pos = 0
			if len(line) > 0 {
				pos = len(line)
			}
		}

		// Convert to runes to handle multi-byte characters correctly
		runes := []rune(line)
		if pos > len(runes) {
			pos = len(runes)
		}

		// Call the user-provided completer function
		completions := f(string(runes[:pos]), mode)

		return "", completions, string(runes[pos:])
	}
}

// END COMPLETER

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
	inBlock        bool
	lastLineString bool
	prevLines      int
	prevCursorLine int
	completer      WordCompleter
	lines          []string                                                       // added for the multiline behaviour
	currline       int                                                            // same
	programState   *env.ProgramState                                              // added for environment access
	displayValue   func(*env.ProgramState, env.Object, bool) (env.Object, string) // callback for displaying values
	// pending     []rune
	// killRing *ring.Ring
}

// NewMicroLiner initializes a new *MLState with the provided event channel,
// output function, and line handler function.
func NewMicroLiner(ch chan KeyEvent, sb func(msg string), el func(line string) string) *MLState {
	if ch == nil {
		panic("KeyEvent channel cannot be nil")
	}
	if sb == nil {
		panic("sendBack function cannot be nil")
	}
	if el == nil {
		panic("enterLine function cannot be nil")
	}

	var s MLState
	s.next = ch
	s.sendBack = sb
	s.enterLine = el
	s.columns = 80 // Default value, will be updated by getColumns()

	return &s
}

// SetProgramState sets the program state for environment access during tab completion
func (s *MLState) SetProgramState(ps *env.ProgramState) {
	s.programState = ps
}

// SetDisplayValueFunc sets the callback function used to display values
func (s *MLState) SetDisplayValueFunc(fn func(*env.ProgramState, env.Object, bool) (env.Object, string)) {
	s.displayValue = fn
}

func (s *MLState) getColumns() bool {
	s.columns = GetTerminalColumns()
	// fmt.Print("*getColumns* : ")
	// fmt.Println(s.columns)
	return true
}

func (s *MLState) GetKeyChan() <-chan KeyEvent {
	return s.next
}

func (s *MLState) SetColumns(cols int) bool {
	s.columns = cols
	//	fmt.Print("*setColumns* : ")
	//	fmt.Println(s.columns)
	return true
}

// Redrawing input
// Called when it needs to redraw / refresh the current input, dispatches to single line and multiline

// refresh updates the display with the current input buffer.
// It returns any error encountered during the refresh operation.
func (s *MLState) refresh(prompt []rune, buf []rune, pos int) error {
	if prompt == nil {
		prompt = []rune{}
	}
	if buf == nil {
		buf = []rune{}
	}
	if pos < 0 {
		pos = 0
	}
	if pos > len(buf) {
		pos = len(buf)
	}

	s.needRefresh = false
	err := s.refreshSingleLine_NO_WRAP(prompt, buf, pos)
	if err != nil {
		return fmt.Errorf("refresh failed: %w", err)
	}
	return nil
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

func (s *MLState) refreshSingleLine_WITH_WRAP_HALFMADE(prompt []rune, buf []rune, pos int) error {
	// traceTop(pos, 0)
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

	// traceTop(len(texts), 0)

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

// refreshSingleLineWithWrap updates the display for input that may wrap across multiple terminal lines.
// It properly detects line wrapping, clears all affected lines, and redraws content with correct cursor positioning.
func (s *MLState) refreshSingleLineWithWrap(prompt []rune, buf []rune, pos int) error {
	if s.columns <= 6 {
		return fmt.Errorf("terminal width too small: %d columns", s.columns)
	}

	pLen := countGlyphs(prompt)
	bLen := countGlyphs(buf)
	text := string(buf)

	// Calculate how many terminal lines the content will occupy
	totalWidth := pLen + bLen
	linesNeeded := (totalWidth + s.columns - 1) / s.columns // Ceiling division
	if linesNeeded < 1 {
		linesNeeded = 1
	}

	// Hide cursor before redrawing to prevent blinking
	s.sendBack("\033[?25l")

	// Move cursor to beginning of current line
	s.sendBack("\r")

	// Clear current line and any additional lines from previous wrapped content
	s.sendBack("\033[K") // Clear current line

	// Clear additional lines if content was previously wrapped
	// We need to clear MORE lines than we think we need, because the previous
	// content might have used more lines than the current content will use
	linesToClear := s.prevLines
	if linesToClear < linesNeeded {
		linesToClear = linesNeeded
	}

	// Clear down to eliminate any residual wrapped lines
	for i := 1; i < linesToClear; i++ {
		s.sendBack("\033[B") // Move down one line
		s.sendBack("\033[K") // Clear line
	}

	// Move back up to the original line position
	if linesToClear > 1 {
		s.sendBack(fmt.Sprintf("\033[%dA", linesToClear-1))
	}

	// Position cursor at start of line
	s.sendBack("\r")

	// Write prompt
	s.sendBack(string(prompt))

	// Apply syntax highlighting and write buffer
	tt2, inString := RyeHighlight(text, s.lastLineString, s.columns)
	s.sendBack(tt2)
	s.inString = inString

	// Calculate cursor position accounting for line wrapping
	cursorTotalPos := pLen + pos
	cursorLine := cursorTotalPos / s.columns
	cursorCol := cursorTotalPos % s.columns

	// Move cursor to correct position
	if cursorLine > 0 {
		// Move down to the correct line
		s.sendBack(fmt.Sprintf("\033[%dB", cursorLine))
	}

	// Move to correct column
	s.sendBack("\r")
	if cursorCol > 0 {
		s.sendBack(fmt.Sprintf("\033[%dC", cursorCol))
	}

	// Update state for next refresh
	s.prevLines = linesNeeded

	// Show cursor after redrawing is complete
	s.sendBack("\033[?25h")

	return nil
}

// refreshSingleLine_NO_WRAP updates the display for a single line input.
// It handles cursor positioning and syntax highlighting, with proper line wrap handling.
func (s *MLState) refreshSingleLine_NO_WRAP(prompt []rune, buf []rune, pos int) error {
	if s.columns <= 6 {
		return fmt.Errorf("terminal width too small: %d columns", s.columns)
	}

	pLen := countGlyphs(prompt)
	bLen := countGlyphs(buf)
	text := string(buf)

	// Calculate how many terminal lines the content will occupy
	totalWidth := pLen + bLen
	linesNeeded := (totalWidth + s.columns - 1) / s.columns // Ceiling division
	if linesNeeded < 1 {
		linesNeeded = 1
	}

	// Hide cursor before redrawing to prevent blinking
	s.sendBack("\033[?25l")

	// Move cursor to beginning of current line
	s.sendBack("\r")

	// Always clear enough lines to handle any previous wrapped content
	// This is the key fix - we clear MORE lines than we might need to ensure
	// we don't leave any residual content from previous redraws
	maxLinesToClear := s.prevLines
	if maxLinesToClear < linesNeeded {
		maxLinesToClear = linesNeeded
	}
	if maxLinesToClear < 3 {
		maxLinesToClear = 3 // Always clear at least 3 lines to be safe
	}

	// Clear current line first
	s.sendBack("\033[K")

	// Clear additional lines by moving down and clearing each one
	for i := 1; i < maxLinesToClear; i++ {
		s.sendBack("\033[B") // Move down one line
		s.sendBack("\033[K") // Clear line
	}

	// Move back up to the original line position
	if maxLinesToClear > 1 {
		s.sendBack(fmt.Sprintf("\033[%dA", maxLinesToClear-1))
	}

	// Position cursor at start of line
	s.sendBack("\r")

	// Write prompt
	s.sendBack(string(prompt))

	// Apply syntax highlighting and write buffer
	tt2, inString := RyeHighlight(text, s.lastLineString, s.columns)
	s.sendBack(tt2)
	s.inString = inString

	// Calculate cursor position accounting for line wrapping
	cursorTotalPos := pLen + pos
	cursorLine := cursorTotalPos / s.columns
	cursorCol := cursorTotalPos % s.columns

	// Move cursor to correct position
	if cursorLine > 0 {
		// Move down to the correct line
		s.sendBack(fmt.Sprintf("\033[%dB", cursorLine))
	}

	// Move to correct column
	s.sendBack("\r")
	if cursorCol > 0 {
		s.sendBack(fmt.Sprintf("\033[%dC", cursorCol))
	}

	// Update state for next refresh
	s.prevLines = linesNeeded

	// Show cursor after redrawing is complete
	s.sendBack("\033[?25h")

	return nil
}

func getLengthOfLastLine(input string) (int, bool) {
	if !strings.Contains(input, "\n") {
		return len(input), false
	}

	lines := strings.Split(input, "\n")
	lastLine := lines[len(lines)-1]
	return len(lastLine) - 3, true // for the prefix because currently string isn't padded on left line TODO unify this
}

// checkIncompleteBlock checks if the accumulated text has unbalanced braces/brackets/parens
func (s *MLState) checkIncompleteBlock(text string) bool {
	openBraces := 0
	openBrackets := 0
	openParens := 0
	inString := false
	stringChar := ' '

	for _, char := range text {
		// Handle string literals to avoid counting brackets inside strings
		if (char == '"' || char == '`') && (stringChar == ' ' || stringChar == char) {
			if !inString {
				inString = true
				stringChar = char
			} else {
				inString = false
				stringChar = ' '
			}
			continue
		}

		if !inString {
			switch char {
			case '{':
				openBraces++
			case '}':
				openBraces--
			case '[':
				openBrackets++
			case ']':
				openBrackets--
			case '(':
				openParens++
			case ')':
				openParens--
			}
		}
	}

	// If any delimiters are unbalanced, consider it an incomplete block
	return openBraces > 0 || openBrackets > 0 || openParens > 0
}

// calculateIndentLevel calculates the indentation level based on open braces/brackets/parens
func (s *MLState) calculateIndentLevel(text string) int {
	openBraces := 0
	openBrackets := 0
	openParens := 0
	inString := false
	stringChar := ' '

	for _, char := range text {
		// Handle string literals to avoid counting brackets inside strings
		if (char == '"' || char == '`') && (stringChar == ' ' || stringChar == char) {
			if !inString {
				inString = true
				stringChar = char
			} else {
				inString = false
				stringChar = ' '
			}
			continue
		}

		if !inString {
			switch char {
			case '{':
				openBraces++
			case '}':
				openBraces--
			case '[':
				openBrackets++
			case ']':
				openBrackets--
			case '(':
				openParens++
			case ')':
				openParens--
			}
		}
	}

	// Return total nesting level (each level = 2 spaces)
	return (openBraces + openBrackets + openParens) * 2
}

// MicroPrompt displays a prompt and handles user input with editing capabilities.
// It returns the final input string or an error if the operation was canceled or failed.
// The prompt is displayed with the given text and cursor position.
// The context can be used to cancel the operation.
func (s *MLState) MicroPrompt(prompt string, text string, pos int, ctx1 context.Context) (string, error) {
	if ctx1 == nil {
		return "", fmt.Errorf("context cannot be nil")
	}
	lastIndentLevel := 0
	// history related
	refreshAllLines := false
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

	if s.currline == 0 {
		p = []rune(prompt)
	} else {
		indent := strings.Repeat(" ", lastIndentLevel)
		p = []rune(indent + ".. ")
	}

	// defer s.stopPrompt()

	// if negative or past end put to the end
	if pos < 0 || len(line) < pos {
		pos = len(line)
	}
	// if len of line is > 0 then refresh
	// if len(line) > 0 {
	err := s.refresh(p, line, pos)
	if err != nil {
		return "", err
	}
	// }
	// var next string

	// LBL restart:
	//	s.startPrompt()
	//	s.getColumns()
	s.getColumns()
	// traceTop(strconv.Itoa(s.columns)+"**", 0)

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
			pos, multiline = getLengthOfLastLine(string(line)) // TODO
			if multiline {
				s.lines = strings.Split(string(line), "\n")
				s.currline = len(s.lines) - 1
				line = []rune(s.lines[len(s.lines)-1])
				refreshAllLines = true
				s.inString = false
			}
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
			pos, multiline = getLengthOfLastLine(string(line))
			if multiline {
				s.lines = strings.Split(string(line), "\n")
				s.currline = len(s.lines) - 1
			}
			s.needRefresh = true
		} else {
			s.doBeep()
		}
	}

	// JM
	//	s_instr := 0

	tabCompletionWasActive := false

	log.Println("MicroPrompt started")

	// mainLoop:
	for {
		select {
		case <-ctx1.Done():
			log.Println("Context canceled")
			fmt.Println("Exiting due to cancelation")
			return "", fmt.Errorf("operation canceled by context")
		default:
			trace("POS: ")
			trace(pos)
			// receive next character from channel
			next := <-s.next
			if s.next == nil {
				return "", fmt.Errorf("event channel is nil")
			}

			log.Printf("Received key event: Key='%s', Code=%d, Ctrl=%t, Alt=%t, Shift=%t", next.Key, next.Code, next.Ctrl, next.Alt, next.Shift)

			// Debug: Check for Ctrl+Z specifically
			if next.Ctrl && (strings.ToLower(next.Key) == "z" || next.Code == 26) {
				log.Println("Detected Ctrl+Z!")
			}

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
		haveNext:
			if next.Ctrl {
				switch strings.ToLower(next.Key) {
				// next line0,
				case "n":
					historyStale = true
					s.lastLineString = false
					s.sendBack(fmt.Sprintf("%s\n%s", color_emph, reset)) // ⏎
					if s.inString {
						s.lastLineString = true
					}
					// DONT SEND LINE BACK BUT STORE IT
					// s.enterLine(string(line) + " ")
					s.currline += 1
					s.lines = append(s.lines, string(line))
					pos = 0
					multiline = true
					line = make([]rune, 0)
					trace(line)
					goto startOfHere
				case "c":
					/* return "", ErrPromptAborted
					line = line[:0]
					pos = 0
					s.restartPrompt() */
					fmt.Println("[ Ctrl+C detected in Microliner , Use Ctrl+D to Exit ]")
					// return "", nil
				case "d":
					/* return "", ErrPromptAborted
					line = line[:0]
					pos = 0
					s.restartPrompt() */
					fmt.Println("Ctrl+D detected in Microliner")
					return "", fmt.Errorf("input canceled with Ctrl+D")
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
				//case "n":
				//	histNext()
				// case "p":
				//	histPrev()
				case "s": // seek in context #experimental
					fmt.Print("*")
					line, pos, next, _ = s.tabComplete(p, line, pos, 1)
					tabCompletionWasActive = true
					goto haveNext
				case "x": // display last returned value interactively
					if s.programState != nil && s.programState.Res != nil && s.displayValue != nil {
						// Move to a new line
						s.sendBack("\n")

						// Call displayValue with interactive=true to show the interactive display
						returnedObj, _ := s.displayValue(s.programState, s.programState.Res, true)

						// If a selection was made (not escaped), insert it into the current line
						if returnedObj != nil {
							// Convert the returned object to its string representation
							//objStr := returnedObj.Print(*s.programState.Idx)

							// Insert the selected value at the current cursor position
							/* objRunes := []rune(objStr)
							line = append(line[:pos], append(objRunes, line[pos:]...)...)
							pos += len(objRunes)
							s.needRefresh = true */
							s.programState.Res = returnedObj
							//fmt.Println(returnedObj.Inspect(*s.programState.Idx))
							fmt.Println(&s.programState)
							p := ""
							if env.IsPointer(s.programState.Res) {
								p = "Ref"
							}
							resultStr := s.programState.Res.Inspect(*s.programState.Idx)
							fmt.Print("\033[38;5;37m" + p + resultStr + "\x1b[0m")
						} else {
							fmt.Println("NIL RETURNED")
						}

						// Force refresh to redraw the prompt
						s.needRefresh = true
					} else {
						s.doBeep() // No result to display
					}
				case "z": // suspend process (Ctrl+Z)
					if err := SuspendProcess(); err != nil {
						return "", err
					}
					// If we reach here, the process was resumed
					fmt.Println("Process is resumed")
					s.needRefresh = true
					// Add new case for Ctrl+Backspace
				case "backspace": // or check `next.Code == 8` if needed
					if pos <= 0 {
						s.doBeep()
					} else {
						// Find the start of the current word
						newPos := pos
						// Skip trailing whitespace
						for newPos > 0 && unicode.IsSpace(line[newPos-1]) {
							newPos--
						}
						// Skip non-whitespace (the word itself)
						for newPos > 0 && !unicode.IsSpace(line[newPos-1]) {
							newPos--
						}
						// Delete from newPos to pos
						line = append(line[:newPos], line[pos:]...)
						pos = newPos
						s.needRefresh = true
					}
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

				case string(0x7f): // or check `next.Code == 8` if needed
					if pos <= 0 {
						s.doBeep()
					} else {
						// Find the start of the current word
						newPos := pos
						// Skip trailing whitespace
						for newPos > 0 && unicode.IsSpace(line[newPos-1]) {
							newPos--
						}
						// Skip non-whitespace (the word itself)
						for newPos > 0 && !unicode.IsSpace(line[newPos-1]) {
							newPos--
						}
						// Delete from newPos to pos
						line = append(line[:newPos], line[pos:]...)
						pos = newPos
						s.needRefresh = true
					}
				}
			} else {
				// Check for Ctrl+Z by ASCII code (26) regardless of other flags
				if next.Code == 26 {
					log.Println("Detected Ctrl+Z by ASCII code 26!")
					fmt.Println("*****")
					if err := SuspendProcess(); err != nil {
						return "", err
					}
					// If we reach here, the process was resumed
					s.needRefresh = true
					continue
				}

				switch next.Code {
				case 13: // Enter Newline
					// Check if we should continue in multiline mode
					allText := strings.Join(s.lines, "\n") + string(line)
					inIncompleteBlock := s.checkIncompleteBlock(allText)

					if s.inString || inIncompleteBlock {
						// This is copy from ctrl+x code above ... deduplicate and systemize TODO
						historyStale = true
						s.lastLineString = false
						s.sendBack(fmt.Sprintf("%s\n%s", color_emph, reset)) //
						if s.inString {
							s.lastLineString = true
						}
						// DONT SEND LINE BACK BUT STORE IT
						// s.enterLine(string(line) + " ")
						s.lines = append(s.lines, string(line))
						pos = 0
						multiline = true
						s.currline += 1
						line = make([]rune, 0)
						trace(line)
						goto startOfHere
					}
					if tabCompletionWasActive {
						// TODO --- make it into a function - deduplicate
						fmt.Println("")
						ClearLine()
						CurUp(1)
					}
					historyStale = true
					s.lastLineString = false
					s.sendBack("\n")
					xx := ""
					if multiline {
						// fmt.Println(s.currline)
						// fmt.Println(len(s.lines))
						if s.currline > len(s.lines)-1 {
							s.lines = append(s.lines, string(line))
						} else {
							s.lines[s.currline] = string(line)
							if len(s.lines) > s.currline+1 {
								CurDown(len(s.lines) - s.currline)
								fmt.Println("")
							}
						}
						xx = s.enterLine(strings.Join(s.lines, "\n"))
					} else {
						xx = s.enterLine(string(line))

					}
					pos = 0
					multiline = false
					if xx == "next line" {
						multiline = true
					} else {
						s.sendBack("") // WW?
					}
					s.currline = 0
					s.lines = make([]string, 0)
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
				case 127: // Alt+Backspace (Delete word)
					if pos <= 0 {
						s.doBeep()
					} else {
						// Find the start of the current word
						newPos := pos
						// Skip trailing whitespace
						for newPos > 0 && unicode.IsSpace(line[newPos-1]) {
							newPos--
						}
						// Skip non-whitespace (the word itself)
						for newPos > 0 && !unicode.IsSpace(line[newPos-1]) {
							newPos--
						}
						// Delete from newPos to pos
						line = append(line[:newPos], line[pos:]...)
						pos = newPos
						s.needRefresh = true
					}
				case 9: // Tab completion
					line, pos, next, _ = s.tabComplete(p, line, pos, 0)
					tabCompletionWasActive = true
					goto haveNext
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
					if multiline {
						if s.currline > 0 { //  len(s.lines) {
							CurUp(1)
							// append the last line -- only when in last line but ok for now
							if s.currline > len(s.lines)-1 {
								s.lines = append(s.lines, string(line))
							} else {
								s.lines[s.currline] = string(line)
							}
							s.currline -= 1                    // later increment
							line = []rune(s.lines[s.currline]) // + "⏎")
							if pos > len(line) {
								pos = len(line) - 1
							}
						}
					} else {
						histPrev()
					}
				case 40: // Down
					if multiline {
						if s.currline < len(s.lines)-1 {
							CurDown(1)
							// append the last line -- only when in last line but ok for now
							s.lines[s.currline] = string(line)
							s.currline += 1                    // later increment
							line = []rune(s.lines[s.currline]) // + "⏎...")
							if pos > len(line) {
								pos = len(line) - 1
							}
						}
					} else {
						histNext()
					}
				case 36: // Home
					pos = 0
				case 35: // End
					pos = len(line)
				case 27: // Escape

				default:
					if (next.Key == " ") && tabCompletionWasActive {
						// TODO --- make it into a function - deduplicate
						// CurDown(1)
						fmt.Println("")
						ClearLine()
						fmt.Println("")
						ClearLine()
						CurUp(2)
					}

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
					fmt.Println("Exiting due to error at refreshAllLines")
					return "", fmt.Errorf("refresh error: %w", err)
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
					}					// Save the result on the killRing
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
			log.Println("MICROPROMPT 2")
			if refreshAllLines {
				log.Println("REFRESH LLL ")
				for i, line1 := range s.lines {
					if i == 0 {
						//	s.sendBack(prompt)
						p = []rune(prompt)
					} else {
						//	s.sendBack(".. ")
						fmt.Println("") // turn to sendback
						// Check if we're in an incomplete block vs incomplete string
						allText := strings.Join(s.lines[:i+1], "\n")
						if s.checkIncompleteBlock(allText) {
							// Use distinctive prompt for incomplete blocks with indentation
							lastIndentLevel = s.calculateIndentLevel(allText)
							indent := strings.Repeat(" ", lastIndentLevel)
							p = []rune(indent + " > ")
							s.inBlock = true
						} else {
							// Use regular multiline prompt for strings etc
							p = []rune("-> ")
							s.inBlock = false
						}
					}
					err := s.refresh(p, []rune(line1), pos)
					if err != nil {
						fmt.Println("Exiting due to error at refreshAllLines")
						return "", err
					}
				}
				refreshAllLines = false
			} else {
				log.Println("ELSE REFRESH LLL ")
				if s.currline == 0 {
					//	s.sendBack(prompt)
					p = []rune(prompt)
				} else {
					// Check if we're in an incomplete block vs incomplete string
					allText := strings.Join(s.lines, "\n") + string(line)
					if s.checkIncompleteBlock(allText) {
						// Use distinctive prompt for incomplete blocks with indentation
						lastIndentLevel = s.calculateIndentLevel(allText)
						indent := strings.Repeat(" ", lastIndentLevel)
						p = []rune(indent + " > ")
						s.inBlock = true
					} else {
						// Use regular multiline prompt for strings etc
						p = []rune("-> ")
						s.inBlock = false
					}
				}
				err := s.refresh(p, line, pos)
				if err != nil {
					fmt.Println("Exiting due to error at refresh")
					return "", fmt.Errorf("refresh error: %w", err)
				}
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
