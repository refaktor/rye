package term

import (
	"fmt"
	"os"
	"unicode/utf8"

	goterm "golang.org/x/term"
)

// TerminalState tracks the current state of the terminal for robust cursor management
type TerminalState struct {
	width           int
	height          int
	cursorRow       int
	cursorCol       int
	savedRow        int
	savedCol        int
	suggestionsRows int  // Number of rows used by suggestions
	atBottom        bool // Whether we're at the bottom of the terminal
}

// ImprovedMLState extends MLState with better terminal state management
type ImprovedMLState struct {
	*MLState
	termState *TerminalState
}

// NewImprovedMicroLiner creates a new improved microliner with terminal state tracking
func NewImprovedMicroLiner(ch chan KeyEvent, sb func(msg string), el func(line string) string) *ImprovedMLState {
	base := NewMicroLiner(ch, sb, el)

	width, height, err := goterm.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		width, height = 80, 24 // fallback
	}

	return &ImprovedMLState{
		MLState: base,
		termState: &TerminalState{
			width:  width,
			height: height,
		},
	}
}

// queryTerminalSize queries the terminal for its current dimensions
func (s *ImprovedMLState) queryTerminalSize() error {
	width, height, err := goterm.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return err
	}
	s.termState.width = width
	s.termState.height = height
	return nil
}

// queryCursorPosition queries the terminal for the current cursor position
func (s *ImprovedMLState) queryCursorPosition() error {
	// Send cursor position query
	s.sendBack("\033[6n")

	// Read response (format: \033[row;colR)
	// This is a simplified version - in practice you'd need to read from stdin
	// For now, we'll estimate based on our tracking
	return nil
}

// ensureSpaceForSuggestions ensures there's enough space at the bottom for suggestions
func (s *ImprovedMLState) ensureSpaceForSuggestions(suggestionsCount int) {
	s.queryTerminalSize()

	// Calculate how many lines we need for suggestions
	linesNeeded := suggestionsCount + 2 // +2 for current selection and probe

	// Check if we're too close to the bottom
	if s.termState.cursorRow > s.termState.height-linesNeeded {
		// We need to scroll up or move cursor up
		linesToMove := (s.termState.cursorRow + linesNeeded) - s.termState.height
		if linesToMove > 0 {
			// Move cursor up by the required amount
			s.sendBack(fmt.Sprintf("\033[%dA", linesToMove))
			s.termState.cursorRow -= linesToMove
		}
	}
}

// displayTabSuggestionsImproved shows tab completion suggestions with proper scroll handling
func (s *ImprovedMLState) displayTabSuggestionsImproved(items []string, currentIndex int) {
	if len(items) == 0 {
		return
	}

	// Ensure we have space for suggestions
	s.ensureSpaceForSuggestions(len(items))

	// Save current cursor position (our own tracking, not terminal's)
	s.termState.savedRow = s.termState.cursorRow
	s.termState.savedCol = s.termState.cursorCol

	// Clear any previous suggestions
	s.clearSuggestions()

	// Move to next line for suggestions
	s.sendBack("\n")
	s.termState.cursorRow++

	// Display suggestions line
	s.sendBack("\033[K")                   // Clear line
	s.sendBack("\033[34mcurrent: \033[0m") // Blue "current:" label

	for i, item := range items {
		if i == currentIndex {
			s.sendBack("\033[45;30m ") // Magenta background, black text
			s.sendBack(item)
			s.sendBack(" \033[0m") // Reset
		} else {
			s.sendBack("\033[36m ") // Cyan text
			s.sendBack(item)
			s.sendBack(" \033[0m") // Reset
		}

		if i < len(items)-1 {
			s.sendBack("  ")
		}
	}

	// Display probe/preview line
	if currentIndex >= 0 && currentIndex < len(items) {
		s.sendBack("\n")
		s.termState.cursorRow++
		s.sendBack("\033[K") // Clear line
		probe := s.getItemProbe(items[currentIndex])
		s.sendBack("\033[38;5;247m") // Gray color
		s.sendBack(probe)
		s.sendBack("\033[0m") // Reset
	}

	s.termState.suggestionsRows = 2 // We used 2 rows for suggestions

	// Return to original cursor position
	s.restoreCursorPosition()
}

// clearSuggestions clears previously displayed suggestions
func (s *ImprovedMLState) clearSuggestions() {
	if s.termState.suggestionsRows > 0 {
		// Move to where suggestions start
		s.sendBack(fmt.Sprintf("\033[%dB", 1)) // Move down to suggestions area

		// Clear each suggestion line
		for i := 0; i < s.termState.suggestionsRows; i++ {
			s.sendBack("\033[K") // Clear line
			if i < s.termState.suggestionsRows-1 {
				s.sendBack("\033[B") // Move down
			}
		}

		// Move back to original position
		s.sendBack(fmt.Sprintf("\033[%dA", s.termState.suggestionsRows))
		s.termState.suggestionsRows = 0
	}
}

// restoreCursorPosition restores cursor to the saved position
func (s *ImprovedMLState) restoreCursorPosition() {
	// Calculate the difference and move cursor accordingly
	rowDiff := s.termState.cursorRow - s.termState.savedRow
	colDiff := s.termState.cursorCol - s.termState.savedCol

	if rowDiff > 0 {
		s.sendBack(fmt.Sprintf("\033[%dA", rowDiff)) // Move up
	} else if rowDiff < 0 {
		s.sendBack(fmt.Sprintf("\033[%dB", -rowDiff)) // Move down
	}

	if colDiff > 0 {
		s.sendBack(fmt.Sprintf("\033[%dD", colDiff)) // Move left
	} else if colDiff < 0 {
		s.sendBack(fmt.Sprintf("\033[%dC", -colDiff)) // Move right
	}

	s.termState.cursorRow = s.termState.savedRow
	s.termState.cursorCol = s.termState.savedCol
}

// updateCursorPosition updates our tracking of cursor position
func (s *ImprovedMLState) updateCursorPosition(row, col int) {
	s.termState.cursorRow = row
	s.termState.cursorCol = col
}

// Alternative approach using absolute positioning
func (s *ImprovedMLState) displayTabSuggestionsAbsolute(items []string, currentIndex int) {
	if len(items) == 0 {
		return
	}

	// Query current terminal size
	s.queryTerminalSize()

	// Calculate where to place suggestions (always at bottom of screen)
	suggestionsStartRow := s.termState.height - 2 // Reserve 2 lines at bottom

	// Save current cursor position
	s.sendBack("\033[s")

	// Move to suggestions area (absolute positioning)
	s.sendBack(fmt.Sprintf("\033[%d;1H", suggestionsStartRow)) // Move to row, column 1

	// Clear the suggestion lines
	s.sendBack("\033[K")       // Clear current line
	s.sendBack("\033[B\033[K") // Move down and clear next line

	// Move back to start of suggestions area
	s.sendBack(fmt.Sprintf("\033[%d;1H", suggestionsStartRow))

	// Display suggestions
	s.sendBack("\033[34mcurrent: \033[0m") // Blue "current:" label

	for i, item := range items {
		if i == currentIndex {
			s.sendBack("\033[45;30m ") // Magenta background, black text
			s.sendBack(item)
			s.sendBack(" \033[0m") // Reset
		} else {
			s.sendBack("\033[36m ") // Cyan text
			s.sendBack(item)
			s.sendBack(" \033[0m") // Reset
		}

		if i < len(items)-1 {
			s.sendBack("  ")
		}
	}

	// Display probe on next line
	if currentIndex >= 0 && currentIndex < len(items) {
		s.sendBack(fmt.Sprintf("\033[%d;1H", suggestionsStartRow+1)) // Move to next line
		s.sendBack("\033[K")                                         // Clear line
		probe := s.getItemProbe(items[currentIndex])
		s.sendBack("\033[38;5;247m") // Gray color
		s.sendBack(probe)
		s.sendBack("\033[0m") // Reset
	}

	// Restore original cursor position
	s.sendBack("\033[u")
}

// Enhanced refresh method that tracks cursor position
func (s *ImprovedMLState) refreshImproved(prompt []rune, buf []rune, pos int) error {
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

	// Update terminal size
	s.queryTerminalSize()

	// Clear suggestions before refresh
	s.clearSuggestions()

	err := s.refreshSingleLineImproved(prompt, buf, pos)
	if err != nil {
		return fmt.Errorf("refresh failed: %w", err)
	}
	return nil
}

// Enhanced single line refresh with cursor tracking
func (s *ImprovedMLState) refreshSingleLineImproved(prompt []rune, buf []rune, pos int) error {
	if s.columns <= 6 {
		return fmt.Errorf("terminal width too small: %d columns", s.columns)
	}

	pLen := countGlyphs(prompt)
	text := string(buf)
	cols := s.columns - 6

	// Hide cursor before redrawing
	s.sendBack("\033[?25l")

	// Position cursor at start of line and clear it
	s.sendBack("\r")     // Carriage return
	s.sendBack("\033[K") // Clear line
	s.termState.cursorCol = 0

	// Write prompt
	s.sendBack(string(prompt))

	// Apply syntax highlighting and write buffer
	tt2, inString1, inString2 := RyeHighlight(text, s.lastLineString, s.lastLineBacktick, cols)
	s.sendBack(tt2)
	s.lastLineString = inString1
	s.lastLineBacktick = inString2
	s.inString = inString1 || inString2

	// Position cursor correctly
	curLeft := pLen + pos
	if curLeft < 0 {
		curLeft = 0
	}

	// Move cursor to correct position
	if curLeft > 0 {
		s.sendBack(fmt.Sprintf("\033[%dC", curLeft))
	}
	s.termState.cursorCol = curLeft

	// Show cursor after redrawing is complete
	s.sendBack("\033[?25h")

	return nil
}

// Method to handle terminal resize events
func (s *ImprovedMLState) handleResize() {
	s.queryTerminalSize()
	s.columns = s.termState.width
	// Clear any suggestions that might be misplaced after resize
	s.clearSuggestions()
}

// Enhanced tab completion with improved suggestion display
func (s *ImprovedMLState) tabCompleteImproved(p []rune, line []rune, pos int, mode int) ([]rune, int, KeyEvent, error) {
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
		err := s.refreshImproved(p, completedLine, newPos)
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
		err = s.refreshImproved(p, completedLine, newPos)
		if err != nil {
			return line, pos, KeyEvent{Code: 27}, fmt.Errorf("failed to refresh display: %w", err)
		}

		// Display suggestions with current selection highlighted
		// Use the absolute positioning method for more reliability
		s.displayTabSuggestionsAbsolute(list, currentIndex)

		// Wait for next key input
		next := <-s.next

		if next.Code == 9 {
			direction = 1
			continue
		}
		if next.Code == 27 {
			// Clear suggestions before returning
			s.clearSuggestions()
			return line, pos, KeyEvent{Code: 27}, nil
		}

		// Clear suggestions before returning with new key
		s.clearSuggestions()
		return completedLine, newPos, next, nil
	}
}
