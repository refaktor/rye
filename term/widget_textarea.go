package term

import (
	"strings"

	"github.com/refaktor/rye/env"
)

// =============================================================================
// TextAreaWidget - Multiline text input
// =============================================================================

// TextAreaWidget handles multiline text input
type TextAreaWidget struct {
	BaseWidget
	Lines  []string
	Width  int
	Height int
	CurRow int
	CurCol int
}

// NewTextAreaWidget creates a new textarea widget
func NewTextAreaWidget(width, height int, initialText string, idx *env.Idxs) *TextAreaWidget {
	w := &TextAreaWidget{
		BaseWidget: NewBaseWidget(idx),
		Lines:      make([]string, height),
		Width:      width,
		Height:     height,
		CurRow:     0,
		CurCol:     0,
	}

	// Initialize empty lines
	for i := range w.Lines {
		w.Lines[i] = ""
	}

	// Parse initial text if provided
	if initialText != "" {
		inputLines := strings.Split(initialText, "\n")
		for i := 0; i < len(inputLines) && i < height; i++ {
			line := inputLines[i]
			if len(line) > width {
				line = line[:width]
			}
			w.Lines[i] = line
		}
	}

	return w
}

func (w *TextAreaWidget) Render() {
	RestoreCurPos()

	for i := 0; i < w.Height; i++ {
		ClearLine()
		line := w.Lines[i]

		if i == w.CurRow {
			// Render line with cursor
			w.renderLineWithCursor(line)
		} else {
			// Pad line to width
			padded := line + strings.Repeat(" ", w.Width-len(line))
			termPrint(padded)
		}
		termPrintln("")
	}

	// Footer
	ClearLine()
	ColorMagenta()
	termPrint("─" + strings.Repeat(" ", w.Width-1) + "┘\n")
	termPrint("ctrl+d to submit, ctrl+c to cancel")
	CloseProps()
	termPrintln("")
}

func (w *TextAreaWidget) renderLineWithCursor(line string) {
	// Characters before cursor
	var pre string
	if w.CurCol > 0 && w.CurCol <= len(line) {
		pre = line[:w.CurCol]
	} else if w.CurCol > len(line) {
		pre = line + strings.Repeat(" ", w.CurCol-len(line))
	}

	// Character at cursor
	cursorChar := " "
	if w.CurCol < len(line) {
		cursorChar = string(line[w.CurCol])
	}

	// Characters after cursor
	var post string
	if w.CurCol < len(line) {
		post = line[w.CurCol+1:]
	}

	termPrint(pre)
	ColorBgGreen()
	ColorBlack()
	termPrint(cursorChar)
	CloseProps()
	termPrint(post)

	// Pad remaining width
	remaining := w.Width - len(line)
	if remaining > 0 {
		termPrint(strings.Repeat(" ", remaining))
	}
}

func (w *TextAreaWidget) HandleKey(key WidgetKey) (done bool, canceled bool) {
	if key.IsCancel() {
		termPrintln("")
		return true, true
	}

	// Ctrl+D to submit
	if key.ASCII == KeyCtrlD {
		termPrintln("")
		return true, false
	}

	// Enter - move to next line
	if key.IsSubmit() {
		if w.CurRow < w.Height-1 {
			w.CurRow++
			w.CurCol = 0
		}
		return false, false
	}

	// Backspace
	if key.ASCII == KeyBackspace {
		w.handleBackspace()
		return false, false
	}

	// Arrow navigation
	if key.IsLeft() {
		w.moveLeft()
		return false, false
	}
	if key.IsRight() {
		w.moveRight()
		return false, false
	}
	if key.IsUp() {
		w.moveUp()
		return false, false
	}
	if key.IsDown() {
		w.moveDown()
		return false, false
	}

	// Regular character input
	if key.ASCII >= 32 && key.ASCII < 127 {
		w.insertChar(string(rune(key.ASCII)))
		return false, false
	}

	return false, false
}

func (w *TextAreaWidget) handleBackspace() {
	if w.CurCol > 0 {
		// Delete character before cursor (UTF-8 safe)
		line := w.Lines[w.CurRow]
		if w.CurCol <= len(line) {
			runes := []rune(line)
			if w.CurCol <= len(runes) {
				w.Lines[w.CurRow] = string(runes[:w.CurCol-1]) + string(runes[w.CurCol:])
			}
		}
		w.CurCol--
	} else if w.CurRow > 0 {
		// Merge with previous line
		prevLen := len(w.Lines[w.CurRow-1])
		w.Lines[w.CurRow-1] += w.Lines[w.CurRow]

		// Shift lines up
		for i := w.CurRow; i < w.Height-1; i++ {
			w.Lines[i] = w.Lines[i+1]
		}
		w.Lines[w.Height-1] = ""

		w.CurRow--
		w.CurCol = prevLen
	}
}

func (w *TextAreaWidget) moveLeft() {
	if w.CurCol > 0 {
		w.CurCol--
	} else if w.CurRow > 0 {
		w.CurRow--
		w.CurCol = len(w.Lines[w.CurRow])
	}
}

func (w *TextAreaWidget) moveRight() {
	if w.CurCol < len(w.Lines[w.CurRow]) {
		w.CurCol++
	} else if w.CurRow < w.Height-1 {
		w.CurRow++
		w.CurCol = 0
	}
}

func (w *TextAreaWidget) moveUp() {
	if w.CurRow > 0 {
		w.CurRow--
		if w.CurCol > len(w.Lines[w.CurRow]) {
			w.CurCol = len(w.Lines[w.CurRow])
		}
	}
}

func (w *TextAreaWidget) moveDown() {
	if w.CurRow < w.Height-1 {
		w.CurRow++
		if w.CurCol > len(w.Lines[w.CurRow]) {
			w.CurCol = len(w.Lines[w.CurRow])
		}
	}
}

func (w *TextAreaWidget) insertChar(ch string) {
	line := w.Lines[w.CurRow]
	if len(line) < w.Width {
		if w.CurCol >= len(line) {
			w.Lines[w.CurRow] = line + ch
		} else {
			w.Lines[w.CurRow] = line[:w.CurCol] + ch + line[w.CurCol:]
		}
		w.CurCol++
	}
}

func (w *TextAreaWidget) GetValue() env.Object {
	result := strings.Join(w.Lines, "\n")
	result = strings.TrimRight(result, "\n ")
	return *env.NewString(result)
}

func (w *TextAreaWidget) GetHeight() int {
	return w.Height + 2 // +2 for footer
}
