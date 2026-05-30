package term

import (
	"strings"
)

// HistoryItem represents a history entry with its original index
type HistoryItem struct {
	Content     string // The history line content
	OriginalIdx int    // Index in the original s.history slice
}

// historyBrowseUnified is a unified function that replaces both historyBrowse and historyBrowseFiltered
func (s *MLState) historyBrowseUnified(p []rune, line []rune, pos int, items []HistoryItem) ([]rune, int, KeyEvent, error) {
	const numLines = 4

	if len(items) == 0 {
		return line, pos, KeyEvent{Code: 27}, nil
	}

	// Start at the most recent entry (last in the items slice)
	// For full history: this is the newest entry
	// For filtered results: this is the first search result
	selectedIdx := len(items) - 1

	s.reserveSuggestionSpace(numLines)
	defer s.clearSuggestionSpace()

	maxW := s.columns - 4
	if maxW < 10 {
		maxW = 10
	}

	// renderDisplay writes the 4 history lines into the reserved space.
	// content[0] → bottom line (selected)
	// content[3] → top line (oldest visible)
	renderDisplay := func() {
		content := make([]string, numLines)
		for i := 0; i < numLines; i++ {
			// i=0 is bottom (selected), i=numLines-1 is top (oldest visible)
			histIdx := selectedIdx - i
			if histIdx < 0 || histIdx >= len(items) {
				content[i] = ""
				continue
			}
			entry := items[histIdx].Content
			runes := []rune(entry)
			if len(runes) > maxW {
				entry = string(runes[:maxW-3]) + "..."
			}
			if histIdx == selectedIdx {
				content[i] = "\033[38;5;37m▶ " + entry + "\033[0m"
			} else {
				content[i] = "\033[90m  " + entry + "\033[0m"
			}
		}
		s.updateSuggestionContent(content)
	}

	// Load and display the initially selected entry
	selectedLine := []rune(items[selectedIdx].Content)
	selectedPos := len(selectedLine)
	if err := s.refresh(p, selectedLine, selectedPos); err != nil {
		return line, pos, KeyEvent{Code: 27}, err
	}
	renderDisplay()

	for {
		next := <-s.next

		switch {
		case next.Code == 38: // Up arrow → older entry (or previous search result)
			if selectedIdx > 0 {
				selectedIdx--
				selectedLine = []rune(items[selectedIdx].Content)
				selectedPos = len(selectedLine)
				s.refresh(p, selectedLine, selectedPos) //nolint:errcheck
				renderDisplay()
			} else {
				s.doBeep()
			}
		case next.Code == 40: // Down arrow → newer entry (or next search result)
			if selectedIdx < len(items)-1 {
				selectedIdx++
				selectedLine = []rune(items[selectedIdx].Content)
				selectedPos = len(selectedLine)
				s.refresh(p, selectedLine, selectedPos) //nolint:errcheck
				renderDisplay()
			} else {
				s.doBeep()
			}
		case next.Ctrl && strings.ToLower(next.Key) == "x": // Ctrl+X → delete selected entry
			// Delete from s.history using the tracked original index
			realIdx := items[selectedIdx].OriginalIdx
			s.DeleteHistoryAt(realIdx)

			// Adjust all OriginalIdx values that are above the deleted slot
			for i := range items {
				if items[i].OriginalIdx > realIdx {
					items[i].OriginalIdx--
				}
			}

			// Remove from local items
			items = append(items[:selectedIdx], items[selectedIdx+1:]...)

			if len(items) == 0 {
				// History is now empty – exit the browser
				s.refresh(p, line, pos) //nolint:errcheck
				return line, pos, KeyEvent{Code: 27}, nil
			}

			// Keep selectedIdx in bounds (clamp to new last entry when at end)
			if selectedIdx >= len(items) {
				selectedIdx = len(items) - 1
			}

			selectedLine = []rune(items[selectedIdx].Content)
			selectedPos = len(selectedLine)
			s.refresh(p, selectedLine, selectedPos) //nolint:errcheck
			renderDisplay()

		case next.Code == 27: // Escape → cancel, restore original input
			s.refresh(p, line, pos) //nolint:errcheck
			return line, pos, KeyEvent{Code: 27}, nil
		case next.Ctrl && strings.ToLower(next.Key) == "r": // Ctrl+R again → accept without executing
			return selectedLine, selectedPos, KeyEvent{Code: 27}, nil
		case next.Code == 13: // Enter → accept and execute
			return selectedLine, selectedPos, next, nil
		default:
			// Any other key: accept the selection and pass the key through for normal handling
			return selectedLine, selectedPos, next, nil
		}
	}
}

// Helper functions to prepare data for the unified browser

// prepareFullHistory prepares all history items for browsing
func (s *MLState) prepareFullHistory() []HistoryItem {
	s.historyMutex.RLock()
	defer s.historyMutex.RUnlock()

	items := make([]HistoryItem, len(s.history))
	for i, content := range s.history {
		items[i] = HistoryItem{
			Content:     content,
			OriginalIdx: i,
		}
	}
	return items
}

// prepareFilteredHistory prepares filtered history items for browsing
func (s *MLState) prepareFilteredHistory(filteredHistory []string) []HistoryItem {
	s.historyMutex.RLock()
	fullHistory := make([]string, len(s.history))
	copy(fullHistory, s.history)
	s.historyMutex.RUnlock()

	items := make([]HistoryItem, 0, len(filteredHistory))

	// Map filtered entries back to original indices
	for _, filteredEntry := range filteredHistory {
		for j, fullEntry := range fullHistory {
			if fullEntry == filteredEntry {
				items = append(items, HistoryItem{
					Content:     filteredEntry,
					OriginalIdx: j,
				})
				break
			}
		}
	}
	return items
}

// Updated function signatures for the existing functions:

// historyBrowse now becomes a simple wrapper
func (s *MLState) historyBrowse(p []rune, line []rune, pos int) ([]rune, int, KeyEvent, error) {
	items := s.prepareFullHistory()
	if len(items) == 0 {
		return line, pos, KeyEvent{Code: 27}, nil
	}
	return s.historyBrowseUnified(p, line, pos, items)
}

// historyBrowseFiltered now becomes a simple wrapper
func (s *MLState) historyBrowseFiltered(p []rune, line []rune, pos int, filteredHistory []string) ([]rune, int, KeyEvent, error) {
	items := s.prepareFilteredHistory(filteredHistory)
	if len(items) == 0 {
		return line, pos, KeyEvent{Code: 27}, nil
	}
	return s.historyBrowseUnified(p, line, pos, items)
}
