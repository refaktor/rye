package term

import (
	"fmt"
	"strings"

	"github.com/refaktor/rye/env"
)

// Key constants for better readability
const (
	KeyCtrlC     = 3
	KeyCtrlD     = 4
	KeyEnter     = 13
	KeyEsc       = 27
	KeyBackspace = 127

	KeyCodeUp    = 38
	KeyCodeDown  = 40
	KeyCodeLeft  = 37
	KeyCodeRight = 39
)

// WidgetKey represents a keyboard input event for widgets
type WidgetKey struct {
	Char    string
	ASCII   int
	KeyCode int
	Err     error
}

// IsCancel returns true if the key is Ctrl+C or Escape
func (k WidgetKey) IsCancel() bool {
	return k.ASCII == KeyCtrlC || k.ASCII == KeyEsc || k.Err != nil
}

// IsSubmit returns true if the key is Enter
func (k WidgetKey) IsSubmit() bool {
	return k.ASCII == KeyEnter
}

// IsUp returns true if the key is Up arrow
func (k WidgetKey) IsUp() bool {
	return k.KeyCode == KeyCodeUp
}

// IsDown returns true if the key is Down arrow
func (k WidgetKey) IsDown() bool {
	return k.KeyCode == KeyCodeDown
}

// IsLeft returns true if the key is Left arrow
func (k WidgetKey) IsLeft() bool {
	return k.KeyCode == KeyCodeLeft
}

// IsRight returns true if the key is Right arrow
func (k WidgetKey) IsRight() bool {
	return k.KeyCode == KeyCodeRight
}

// Widget is the interface for all TUI components
type Widget interface {
	// Render draws the widget at the current cursor position
	Render()
	// HandleKey processes a key event, returns (done, canceled)
	HandleKey(key WidgetKey) (done bool, canceled bool)
	// GetValue returns the current value of the widget
	GetValue() env.Object
	// GetHeight returns the number of lines the widget occupies
	GetHeight() int
}

// Theme defines colors and styles for widgets
type Theme struct {
	Selected     func() // Function to set selected item style
	Normal       func() // Function to reset to normal style
	Header       func() // Function to set header style
	Cursor       string // Cursor character for selected items
	CursorNormal string // Space or empty for non-selected items
}

// DefaultTheme is the default color theme
var DefaultTheme = Theme{
	Selected: func() {
		ColorBrGreen()
		Bold()
	},
	Normal: func() {
		CloseProps()
	},
	Header: func() {
		Bold()
	},
	Cursor:       "» ",
	CursorNormal: "  ",
}

// BaseWidget provides common functionality for all widgets
type BaseWidget struct {
	Theme    Theme
	Idx      *env.Idxs
	Mode     int // 0 = human (Print), 1 = dev (Inspect)
	MoveUp   int // Lines to move up before redraw
}

// NewBaseWidget creates a new BaseWidget with default theme
func NewBaseWidget(idx *env.Idxs) BaseWidget {
	return BaseWidget{
		Theme: DefaultTheme,
		Idx:   idx,
		Mode:  0,
	}
}

// ToggleMode switches between human and dev display modes
func (b *BaseWidget) ToggleMode() {
	b.Mode = 1 - b.Mode
}

// FormatValue formats an object based on the current mode
func (b *BaseWidget) FormatValue(obj env.Object) string {
	if b.Mode == 0 {
		return obj.Print(*b.Idx)
	}
	return obj.Inspect(*b.Idx)
}

// PrepareRedraw moves cursor up and saves position for redraw
func (b *BaseWidget) PrepareRedraw() {
	if b.MoveUp > 0 {
		CurUp(b.MoveUp)
	}
	SaveCurPos()
}

// RunWidget runs a widget's event loop with proper cursor handling
func RunWidget(w Widget) (env.Object, bool) {
	HideCur()
	defer ShowCur()

	SaveCurPos()

	for {
		w.Render()

		ascii, keyCode, err := GetChar()
		key := WidgetKey{ASCII: ascii, KeyCode: keyCode, Err: err}

		done, canceled := w.HandleKey(key)
		if done {
			return w.GetValue(), canceled
		}
	}
}

// =============================================================================
// SelectWidget - Selection from a list of items
// =============================================================================

// SelectWidget allows selecting an item from a block
type SelectWidget struct {
	BaseWidget
	Items   []env.Object
	Current int
}

// NewSelectWidget creates a new selection widget
func NewSelectWidget(items []env.Object, idx *env.Idxs) *SelectWidget {
	return &SelectWidget{
		BaseWidget: NewBaseWidget(idx),
		Items:      items,
		Current:    0,
	}
}

func (w *SelectWidget) Render() {
	w.PrepareRedraw()

	totalLines := 0
	for i, item := range w.Items {
		ClearLine()
		if i == w.Current {
			w.Theme.Selected()
			termPrint(w.Theme.Cursor)
		} else {
			termPrint(w.Theme.CursorNormal)
		}

		valueStr := w.FormatValue(item)
		termPrintln(valueStr)
		totalLines += strings.Count(valueStr, "\n") + 1
		w.Theme.Normal()
	}

	w.MoveUp = totalLines
}

func (w *SelectWidget) HandleKey(key WidgetKey) (done bool, canceled bool) {
	if key.IsCancel() {
		return true, true
	}

	if key.IsSubmit() {
		return true, false
	}

	// Mode toggle
	if key.ASCII == 'm' || key.ASCII == 'M' {
		w.ToggleMode()
		return false, false
	}

	// Navigation
	if key.IsDown() {
		w.Current++
		if w.Current >= len(w.Items) {
			w.Current = 0
		}
	} else if key.IsUp() {
		w.Current--
		if w.Current < 0 {
			w.Current = len(w.Items) - 1
		}
	}

	return false, false
}

func (w *SelectWidget) GetValue() env.Object {
	if w.Current >= 0 && w.Current < len(w.Items) {
		return w.Items[w.Current]
	}
	return env.NewError("no selection")
}

func (w *SelectWidget) GetHeight() int {
	return len(w.Items)
}

// =============================================================================
// InputWidget - Single line text input
// =============================================================================

// InputWidget handles single-line text input
type InputWidget struct {
	BaseWidget
	Text     string
	MaxLen   int
	Position int // cursor position within text
}

// NewInputWidget creates a new input widget
func NewInputWidget(maxLen int, idx *env.Idxs) *InputWidget {
	return &InputWidget{
		BaseWidget: NewBaseWidget(idx),
		Text:       "",
		MaxLen:     maxLen,
		Position:   0,
	}
}

func (w *InputWidget) Render() {
	RestoreCurPos()
	// Clear and redraw
	termPrint(strings.Repeat(" ", w.MaxLen+2))
	RestoreCurPos()
	termPrint(w.Text)
}

func (w *InputWidget) HandleKey(key WidgetKey) (done bool, canceled bool) {
	if key.IsCancel() {
		return true, true
	}

	if key.IsSubmit() {
		termPrintln("")
		termPrintln("")
		return true, false
	}

	// Backspace
	if key.ASCII == KeyBackspace {
		if len(w.Text) > 0 {
			// Handle UTF-8 properly
			runes := []rune(w.Text)
			w.Text = string(runes[:len(runes)-1])
		}
		return false, false
	}

	// Regular character
	if key.ASCII >= 32 && key.ASCII < 127 {
		if len(w.Text) < w.MaxLen {
			w.Text += string(rune(key.ASCII))
		}
	}

	return false, false
}

func (w *InputWidget) GetValue() env.Object {
	return *env.NewString(w.Text)
}

func (w *InputWidget) GetHeight() int {
	return 1
}

// =============================================================================
// PaginatedSelectWidget - Selection with pagination for large lists
// =============================================================================

// PaginatedSelectWidget handles selection from large lists with pagination
type PaginatedSelectWidget struct {
	BaseWidget
	Items       []env.Object
	Current     int // Global index
	PageSize    int
	CurrentPage int
}

// NewPaginatedSelectWidget creates a new paginated selection widget
func NewPaginatedSelectWidget(items []env.Object, pageSize int, idx *env.Idxs) *PaginatedSelectWidget {
	return &PaginatedSelectWidget{
		BaseWidget:  NewBaseWidget(idx),
		Items:       items,
		Current:     0,
		PageSize:    pageSize,
		CurrentPage: 0,
	}
}

func (w *PaginatedSelectWidget) totalPages() int {
	pages := (len(w.Items) + w.PageSize - 1) / w.PageSize
	if pages == 0 {
		return 1
	}
	return pages
}

func (w *PaginatedSelectWidget) localIndex() int {
	return w.Current - (w.CurrentPage * w.PageSize)
}

func (w *PaginatedSelectWidget) Render() {
	w.PrepareRedraw()

	start := w.CurrentPage * w.PageSize
	end := start + w.PageSize
	if end > len(w.Items) {
		end = len(w.Items)
	}

	totalLines := 0
	for i := 0; i < w.PageSize; i++ {
		ClearLine()
		globalIdx := start + i
		if globalIdx < end {
			item := w.Items[globalIdx]
			if globalIdx == w.Current {
				w.Theme.Selected()
				termPrint(w.Theme.Cursor)
			} else {
				termPrint(w.Theme.CursorNormal)
			}
			valueStr := w.FormatValue(item)
			termPrintln(valueStr)
			totalLines += strings.Count(valueStr, "\n") + 1
			w.Theme.Normal()
		} else {
			termPrintln("")
			totalLines++
		}
	}

	// Footer
	termPrintln(fmt.Sprintf("Page %d/%d (n=next, p=prev, m=mode)", w.CurrentPage+1, w.totalPages()))
	totalLines++

	w.MoveUp = totalLines
}

func (w *PaginatedSelectWidget) HandleKey(key WidgetKey) (done bool, canceled bool) {
	if key.IsCancel() {
		return true, true
	}

	if key.IsSubmit() {
		return true, false
	}

	// Mode toggle
	if key.ASCII == 'm' || key.ASCII == 'M' {
		w.ToggleMode()
		return false, false
	}

	// Page navigation
	if key.ASCII == 'n' || key.ASCII == 'N' {
		if w.CurrentPage < w.totalPages()-1 {
			w.CurrentPage++
			w.Current = w.CurrentPage * w.PageSize
		}
		return false, false
	}
	if key.ASCII == 'p' || key.ASCII == 'P' {
		if w.CurrentPage > 0 {
			w.CurrentPage--
			w.Current = w.CurrentPage * w.PageSize
		}
		return false, false
	}

	// Item navigation
	pageStart := w.CurrentPage * w.PageSize
	pageEnd := pageStart + w.PageSize
	if pageEnd > len(w.Items) {
		pageEnd = len(w.Items)
	}

	if key.IsDown() {
		w.Current++
		if w.Current >= pageEnd {
			if w.CurrentPage < w.totalPages()-1 {
				w.CurrentPage++
				w.Current = w.CurrentPage * w.PageSize
			} else {
				w.CurrentPage = 0
				w.Current = 0
			}
		}
	} else if key.IsUp() {
		w.Current--
		if w.Current < pageStart {
			if w.CurrentPage > 0 {
				w.CurrentPage--
				w.Current = (w.CurrentPage+1)*w.PageSize - 1
				if w.Current >= len(w.Items) {
					w.Current = len(w.Items) - 1
				}
			} else {
				w.CurrentPage = w.totalPages() - 1
				w.Current = len(w.Items) - 1
			}
		}
	}

	return false, false
}

func (w *PaginatedSelectWidget) GetValue() env.Object {
	if w.Current >= 0 && w.Current < len(w.Items) {
		return w.Items[w.Current]
	}
	return env.NewError("no selection")
}

func (w *PaginatedSelectWidget) GetHeight() int {
	return w.PageSize + 1 // +1 for footer
}
