//go:build !no_termui
// +build !no_termui

package evaldo

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/mattn/go-runewidth"
	"github.com/refaktor/keyboard"
	"github.com/refaktor/rye/env"
	"github.com/refaktor/rye/term"
)

// ansiRegex matches ANSI escape sequences
var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]|\x1b\]8;;[^\x1b]*\x1b\\`)

// VisibleWidth returns the display width of a string, ignoring ANSI escape codes
func VisibleWidth(s string) int {
	// Strip ANSI codes
	stripped := ansiRegex.ReplaceAllString(s, "")
	return runewidth.StringWidth(stripped)
}

// TruncateToWidth truncates a string to fit within the given width,
// preserving ANSI escape codes and optionally adding an ellipsis
func TruncateToWidth(s string, width int, ellipsis string) string {
	if width <= 0 {
		return ""
	}

	visWidth := VisibleWidth(s)
	if visWidth <= width {
		return s
	}

	ellipsisWidth := VisibleWidth(ellipsis)
	targetWidth := width - ellipsisWidth
	if targetWidth <= 0 {
		return ellipsis[:width] // Edge case: ellipsis longer than width
	}

	var result strings.Builder
	currentWidth := 0
	inEscape := false
	escapeStart := 0
	hasAnsi := false // Track if we encountered any ANSI codes

	runes := []rune(s)
	for i := 0; i < len(runes); i++ {
		r := runes[i]

		// Detect start of ANSI escape
		if r == '\x1b' {
			inEscape = true
			hasAnsi = true
			escapeStart = i
			result.WriteRune(r)
			continue
		}

		if inEscape {
			result.WriteRune(r)
			// Check for end of escape sequence
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				inEscape = false
			}
			// Also handle OSC sequences (end with ST or BEL)
			if i > escapeStart && runes[escapeStart+1] == ']' {
				if r == '\\' && i > 0 && runes[i-1] == '\x1b' {
					inEscape = false
				}
			}
			continue
		}

		// Regular character - check if it fits
		charWidth := runewidth.RuneWidth(r)
		if currentWidth+charWidth > targetWidth {
			break
		}
		result.WriteRune(r)
		currentWidth += charWidth
	}

	// Only add reset code if we had ANSI codes (to close any open formatting)
	if hasAnsi {
		result.WriteString("\x1b[0m")
	}
	result.WriteString(ellipsis)
	return result.String()
}

// WrapText wraps text to the given width, preserving ANSI codes across lines
func WrapText(s string, width int) []string {
	if width <= 0 {
		return []string{}
	}

	var lines []string
	var currentLine strings.Builder
	currentWidth := 0
	var activeStyles strings.Builder // Track active ANSI styles

	runes := []rune(s)
	for i := 0; i < len(runes); i++ {
		r := runes[i]

		// Handle newlines
		if r == '\n' {
			lines = append(lines, currentLine.String())
			currentLine.Reset()
			// Reapply active styles to new line
			currentLine.WriteString(activeStyles.String())
			currentWidth = 0
			continue
		}

		// Detect ANSI escape sequences
		if r == '\x1b' && i+1 < len(runes) && runes[i+1] == '[' {
			// Find end of escape sequence
			escEnd := i + 2
			for escEnd < len(runes) && !((runes[escEnd] >= 'a' && runes[escEnd] <= 'z') || (runes[escEnd] >= 'A' && runes[escEnd] <= 'Z')) {
				escEnd++
			}
			if escEnd < len(runes) {
				escEnd++ // Include the letter
				escape := string(runes[i:escEnd])
				currentLine.WriteString(escape)

				// Track style changes
				if escape == "\x1b[0m" {
					activeStyles.Reset() // Reset clears all styles
				} else {
					activeStyles.WriteString(escape)
				}
				i = escEnd - 1
				continue
			}
		}

		// Regular character
		charWidth := runewidth.RuneWidth(r)
		if currentWidth+charWidth > width {
			// Wrap to next line
			lines = append(lines, currentLine.String())
			currentLine.Reset()
			// Reapply active styles to new line
			currentLine.WriteString(activeStyles.String())
			currentWidth = 0
		}
		currentLine.WriteRune(r)
		currentWidth += charWidth
	}

	// Don't forget the last line
	if currentLine.Len() > 0 {
		lines = append(lines, currentLine.String())
	}

	return lines
}

// ============================================================================
// Theming System
// ============================================================================

// Theme holds color/style definitions for TUI components
type Theme struct {
	colors map[string]string
	mu     sync.RWMutex
}

// DefaultTheme returns a theme with sensible defaults
func DefaultTheme() *Theme {
	return &Theme{
		colors: map[string]string{
			// Basic colors
			"text":     "",
			"accent":   "\x1b[36m", // cyan
			"muted":    "\x1b[90m", // gray
			"dim":      "\x1b[2m",  // dim
			"bold":     "\x1b[1m",
			"italic":   "\x1b[3m",
			"underline": "\x1b[4m",

			// Status colors
			"success": "\x1b[32m", // green
			"error":   "\x1b[31m", // red
			"warning": "\x1b[33m", // yellow
			"info":    "\x1b[34m", // blue

			// UI elements
			"border":       "\x1b[90m", // gray
			"borderAccent": "\x1b[36m", // cyan
			"selected":     "\x1b[7m",  // inverse
			"selectedBg":   "\x1b[44m", // blue bg
			"cursor":       "\x1b[7m",  // inverse

			// Component-specific
			"title":       "\x1b[1;36m", // bold cyan
			"placeholder": "\x1b[90m",   // gray
			"scrollbar":   "\x1b[90m",   // gray
		},
	}
}

// NewTheme creates an empty theme
func NewTheme() *Theme {
	return &Theme{
		colors: make(map[string]string),
	}
}

// Set sets a color/style
func (t *Theme) Set(name, value string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.colors[name] = value
}

// Get returns a color/style, or empty string if not found
func (t *Theme) Get(name string) string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.colors[name]
}

// Apply applies a style to text and adds reset
func (t *Theme) Apply(name, text string) string {
	style := t.Get(name)
	if style == "" {
		return text
	}
	return style + text + "\x1b[0m"
}

// Reset returns the ANSI reset code
func (t *Theme) Reset() string {
	return "\x1b[0m"
}

// Global default theme
var globalTheme = DefaultTheme()

// GetGlobalTheme returns the global theme
func GetGlobalTheme() *Theme {
	return globalTheme
}

// SetGlobalTheme sets the global theme
func SetGlobalTheme(t *Theme) {
	globalTheme = t
}

// ============================================================================
// Terminal Output Helpers
// ============================================================================

// tuiSyncStart begins synchronized output (prevents flicker)
func tuiSyncStart() {
	fmt.Print("\x1b[?2026h")
}

// tuiSyncEnd ends synchronized output
func tuiSyncEnd() {
	fmt.Print("\x1b[?2026l")
}

// tuiMoveCursor moves cursor to row, col (1-indexed)
func tuiMoveCursor(row, col int) {
	fmt.Printf("\x1b[%d;%dH", row, col)
}

// tuiMoveToColumn moves cursor to column (1-indexed)
func tuiMoveToColumn(col int) {
	fmt.Printf("\x1b[%dG", col)
}

// tuiSaveCursor saves cursor position
func tuiSaveCursor() {
	fmt.Print("\x1b[s")
}

// tuiRestoreCursor restores cursor position
func tuiRestoreCursor() {
	fmt.Print("\x1b[u")
}

// tuiHideCursor hides the cursor
func tuiHideCursor() {
	fmt.Print("\x1b[?25l")
}

// tuiShowCursor shows the cursor
func tuiShowCursor() {
	fmt.Print("\x1b[?25h")
}

// tuiClearScreen clears the entire screen
func tuiClearScreen() {
	fmt.Print("\x1b[2J")
}

// tuiClearToEnd clears from cursor to end of screen
func tuiClearToEnd() {
	fmt.Print("\x1b[J")
}

// tuiEnableAltScreen switches to alternate screen buffer
func tuiEnableAltScreen() {
	fmt.Print("\x1b[?1049h")
}

// tuiDisableAltScreen switches back to main screen buffer
func tuiDisableAltScreen() {
	fmt.Print("\x1b[?1049l")
}

// tuiEnableMouse enables mouse tracking
func tuiEnableMouse() {
	fmt.Print("\x1b[?1000h\x1b[?1006h")
}

// tuiDisableMouse disables mouse tracking
func tuiDisableMouse() {
	fmt.Print("\x1b[?1000l\x1b[?1006l")
}

// ============================================================================
// Component System
// ============================================================================

// UIComponent is the interface for all TUI components
type UIComponent interface {
	Render(width int) []string
	Invalidate()
}

// FocusableComponent is a component that can receive focus
type FocusableComponent interface {
	UIComponent
	SetFocused(focused bool)
	IsFocused() bool
	HandleKey(key string) bool // Returns true if key was handled
}

// Container holds child components and renders them vertically
type Container struct {
	children    []UIComponent
	cachedWidth int
	cachedLines []string
	mu          sync.Mutex
}

// NewContainer creates a new container
func NewContainer() *Container {
	return &Container{
		children: make([]UIComponent, 0),
	}
}

// AddChild adds a component to the container
func (c *Container) AddChild(child UIComponent) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.children = append(c.children, child)
	c.invalidate()
}

// RemoveChild removes a component from the container
func (c *Container) RemoveChild(child UIComponent) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for i, ch := range c.children {
		if ch == child {
			c.children = append(c.children[:i], c.children[i+1:]...)
			break
		}
	}
	c.invalidate()
}

// Clear removes all children
func (c *Container) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.children = make([]UIComponent, 0)
	c.invalidate()
}

// Children returns the list of children
func (c *Container) Children() []UIComponent {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.children
}

// Render renders all children vertically
func (c *Container) Render(width int) []string {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.cachedLines != nil && c.cachedWidth == width {
		return c.cachedLines
	}

	var lines []string
	for _, child := range c.children {
		childLines := child.Render(width)
		lines = append(lines, childLines...)
	}

	c.cachedWidth = width
	c.cachedLines = lines
	return lines
}

// Invalidate clears the cache and invalidates all children
func (c *Container) Invalidate() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.invalidate()
}

func (c *Container) invalidate() {
	c.cachedWidth = 0
	c.cachedLines = nil
	for _, child := range c.children {
		child.Invalidate()
	}
}

// TextComponent displays text with optional wrapping
type TextComponent struct {
	text        string
	wrap        bool
	paddingX    int
	paddingY    int
	style       string // ANSI style prefix
	cachedWidth int
	cachedLines []string
	mu          sync.Mutex
}

// NewTextComponent creates a new text component
func NewTextComponent(text string) *TextComponent {
	return &TextComponent{
		text: text,
		wrap: true,
	}
}

// SetText updates the text
func (t *TextComponent) SetText(text string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.text = text
	t.invalidate()
}

// SetWrap enables/disables word wrapping
func (t *TextComponent) SetWrap(wrap bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.wrap = wrap
	t.invalidate()
}

// SetPadding sets horizontal and vertical padding
func (t *TextComponent) SetPadding(x, y int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.paddingX = x
	t.paddingY = y
	t.invalidate()
}

// SetStyle sets the ANSI style (e.g., "\x1b[1m" for bold)
func (t *TextComponent) SetStyle(style string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.style = style
	t.invalidate()
}

// Render renders the text
func (t *TextComponent) Render(width int) []string {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.cachedLines != nil && t.cachedWidth == width {
		return t.cachedLines
	}

	contentWidth := width - (t.paddingX * 2)
	if contentWidth < 1 {
		contentWidth = 1
	}

	var lines []string

	// Top padding
	for i := 0; i < t.paddingY; i++ {
		lines = append(lines, "")
	}

	// Content
	var contentLines []string
	if t.wrap {
		contentLines = WrapText(t.text, contentWidth)
	} else {
		contentLines = []string{TruncateToWidth(t.text, contentWidth, "...")}
	}

	padding := strings.Repeat(" ", t.paddingX)
	for _, line := range contentLines {
		styledLine := line
		if t.style != "" {
			styledLine = t.style + line + "\x1b[0m"
		}
		lines = append(lines, padding+styledLine)
	}

	// Bottom padding
	for i := 0; i < t.paddingY; i++ {
		lines = append(lines, "")
	}

	t.cachedWidth = width
	t.cachedLines = lines
	return lines
}

// Invalidate clears the cache
func (t *TextComponent) Invalidate() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.invalidate()
}

func (t *TextComponent) invalidate() {
	t.cachedWidth = 0
	t.cachedLines = nil
}

// SpacerComponent adds empty lines
type SpacerComponent struct {
	height int
}

// NewSpacerComponent creates a spacer with given height
func NewSpacerComponent(height int) *SpacerComponent {
	return &SpacerComponent{height: height}
}

// Render returns empty lines
func (s *SpacerComponent) Render(width int) []string {
	lines := make([]string, s.height)
	for i := range lines {
		lines[i] = ""
	}
	return lines
}

// Invalidate does nothing for spacer
func (s *SpacerComponent) Invalidate() {}

// BoxComponent wraps content with a border
type BoxComponent struct {
	child       UIComponent
	title       string
	borderStyle int // 0=single, 1=double, 2=rounded, 3=ascii
	paddingX    int
	paddingY    int
	cachedWidth int
	cachedLines []string
	mu          sync.Mutex
}

// Border characters for different styles
var borderChars = []struct {
	tl, tr, bl, br, h, v string
}{
	{"┌", "┐", "└", "┘", "─", "│"}, // single
	{"╔", "╗", "╚", "╝", "═", "║"}, // double
	{"╭", "╮", "╰", "╯", "─", "│"}, // rounded
	{"+", "+", "+", "+", "-", "|"}, // ascii
}

// NewBoxComponent creates a box around a child component
func NewBoxComponent(child UIComponent) *BoxComponent {
	return &BoxComponent{
		child:       child,
		borderStyle: 0,
		paddingX:    1,
		paddingY:    0,
	}
}

// SetTitle sets the box title
func (b *BoxComponent) SetTitle(title string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.title = title
	b.invalidate()
}

// SetBorderStyle sets the border style (0=single, 1=double, 2=rounded, 3=ascii)
func (b *BoxComponent) SetBorderStyle(style int) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if style >= 0 && style < len(borderChars) {
		b.borderStyle = style
	}
	b.invalidate()
}

// SetPadding sets internal padding
func (b *BoxComponent) SetPadding(x, y int) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.paddingX = x
	b.paddingY = y
	b.invalidate()
}

// SetChild sets the child component
func (b *BoxComponent) SetChild(child UIComponent) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.child = child
	b.invalidate()
}

// Render renders the box with border
func (b *BoxComponent) Render(width int) []string {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.cachedLines != nil && b.cachedWidth == width {
		return b.cachedLines
	}

	bc := borderChars[b.borderStyle]
	innerWidth := width - 2 - (b.paddingX * 2) // 2 for borders
	if innerWidth < 1 {
		innerWidth = 1
	}

	var lines []string
	padding := strings.Repeat(" ", b.paddingX)

	// Top border with optional title
	topLine := bc.tl
	if b.title != "" {
		titlePart := " " + TruncateToWidth(b.title, width-4, "...") + " "
		remaining := width - 2 - VisibleWidth(titlePart)
		if remaining > 0 {
			topLine += titlePart + strings.Repeat(bc.h, remaining)
		} else {
			topLine += strings.Repeat(bc.h, width-2)
		}
	} else {
		topLine += strings.Repeat(bc.h, width-2)
	}
	topLine += bc.tr
	lines = append(lines, topLine)

	// Top padding
	for i := 0; i < b.paddingY; i++ {
		lines = append(lines, bc.v+strings.Repeat(" ", width-2)+bc.v)
	}

	// Content
	if b.child != nil {
		childLines := b.child.Render(innerWidth)
		for _, cl := range childLines {
			// Pad child line to inner width
			visWidth := VisibleWidth(cl)
			padRight := innerWidth - visWidth
			if padRight < 0 {
				padRight = 0
			}
			line := bc.v + padding + cl + strings.Repeat(" ", padRight) + padding + bc.v
			lines = append(lines, line)
		}
	}

	// Bottom padding
	for i := 0; i < b.paddingY; i++ {
		lines = append(lines, bc.v+strings.Repeat(" ", width-2)+bc.v)
	}

	// Bottom border
	bottomLine := bc.bl + strings.Repeat(bc.h, width-2) + bc.br
	lines = append(lines, bottomLine)

	b.cachedWidth = width
	b.cachedLines = lines
	return lines
}

// Invalidate clears cache and child cache
func (b *BoxComponent) Invalidate() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.invalidate()
}

func (b *BoxComponent) invalidate() {
	b.cachedWidth = 0
	b.cachedLines = nil
	if b.child != nil {
		b.child.Invalidate()
	}
}

// HorizontalLineComponent draws a horizontal line
type HorizontalLineComponent struct {
	char  string
	style string
}

// NewHorizontalLineComponent creates a horizontal line
func NewHorizontalLineComponent(char string) *HorizontalLineComponent {
	if char == "" {
		char = "─"
	}
	return &HorizontalLineComponent{char: char}
}

// SetStyle sets the ANSI style
func (h *HorizontalLineComponent) SetStyle(style string) {
	h.style = style
}

// Render renders the horizontal line
func (h *HorizontalLineComponent) Render(width int) []string {
	line := strings.Repeat(h.char, width)
	if h.style != "" {
		line = h.style + line + "\x1b[0m"
	}
	return []string{line}
}

// Invalidate does nothing for horizontal line
func (h *HorizontalLineComponent) Invalidate() {}

// SelectListComponent is an interactive selection list
type SelectListComponent struct {
	items        []string
	selected     int
	maxVisible   int
	scrollOffset int
	prefix       string
	prefixNormal string
	style        string // Style for selected item
	cachedWidth  int
	cachedLines  []string
	mu           sync.Mutex
}

// NewSelectListComponent creates a new selection list
func NewSelectListComponent(items []string, maxVisible int) *SelectListComponent {
	if maxVisible <= 0 {
		maxVisible = 10
	}
	return &SelectListComponent{
		items:        items,
		maxVisible:   maxVisible,
		prefix:       "> ",
		prefixNormal: "  ",
	}
}

// SetItems updates the list items
func (s *SelectListComponent) SetItems(items []string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items = items
	if s.selected >= len(items) {
		s.selected = len(items) - 1
	}
	if s.selected < 0 {
		s.selected = 0
	}
	s.invalidate()
}

// SetSelected sets the selected index
func (s *SelectListComponent) SetSelected(index int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if index >= 0 && index < len(s.items) {
		s.selected = index
		s.updateScroll()
		s.invalidate()
	}
}

// Selected returns the currently selected index
func (s *SelectListComponent) Selected() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.selected
}

// SelectedItem returns the currently selected item
func (s *SelectListComponent) SelectedItem() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.selected >= 0 && s.selected < len(s.items) {
		return s.items[s.selected]
	}
	return ""
}

// MoveUp moves selection up
func (s *SelectListComponent) MoveUp() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.selected > 0 {
		s.selected--
		s.updateScroll()
		s.invalidate()
	}
}

// MoveDown moves selection down
func (s *SelectListComponent) MoveDown() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.selected < len(s.items)-1 {
		s.selected++
		s.updateScroll()
		s.invalidate()
	}
}

// SetPrefix sets the selected/normal prefixes
func (s *SelectListComponent) SetPrefix(selected, normal string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.prefix = selected
	s.prefixNormal = normal
	s.invalidate()
}

// SetStyle sets the style for selected item
func (s *SelectListComponent) SetStyle(style string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.style = style
	s.invalidate()
}

func (s *SelectListComponent) updateScroll() {
	// Ensure selected item is visible
	if s.selected < s.scrollOffset {
		s.scrollOffset = s.selected
	} else if s.selected >= s.scrollOffset+s.maxVisible {
		s.scrollOffset = s.selected - s.maxVisible + 1
	}
}

// Render renders the selection list
func (s *SelectListComponent) Render(width int) []string {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.cachedLines != nil && s.cachedWidth == width {
		return s.cachedLines
	}

	var lines []string
	endIdx := s.scrollOffset + s.maxVisible
	if endIdx > len(s.items) {
		endIdx = len(s.items)
	}

	for i := s.scrollOffset; i < endIdx; i++ {
		var line string
		if i == s.selected {
			line = s.prefix + s.items[i]
			if s.style != "" {
				line = s.style + line + "\x1b[0m"
			}
		} else {
			line = s.prefixNormal + s.items[i]
		}
		lines = append(lines, TruncateToWidth(line, width, "..."))
	}

	// Scroll indicator
	if len(s.items) > s.maxVisible {
		indicator := fmt.Sprintf(" [%d-%d of %d]", s.scrollOffset+1, endIdx, len(s.items))
		lines = append(lines, indicator)
	}

	s.cachedWidth = width
	s.cachedLines = lines
	return lines
}

// Invalidate clears the cache
func (s *SelectListComponent) Invalidate() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.invalidate()
}

func (s *SelectListComponent) invalidate() {
	s.cachedWidth = 0
	s.cachedLines = nil
}

// ============================================================================
// End Component System
// ============================================================================

// InlineApp represents a reactive inline terminal UI component
// It renders N lines in place, re-rendering when state changes
type InlineApp struct {
	height         int                   // Number of lines this app occupies
	state          env.Dict              // Current state
	renderFn       env.Object            // Render function (Function or Native GoRenderFn)
	keyHandlers    map[string]env.Object // Key -> handler function/block
	defaultHandler env.Object            // Default handler for unregistered keys
	running        bool                  // Is the app running
	stopChan       chan struct{}         // Channel to signal stop
	mu             sync.Mutex            // Protects state and running
	ps             *env.ProgramState     // Program state for Rye function calls
	rendered       bool                  // Has rendered at least once
	updateChan     chan env.Dict         // Channel for state updates
	prevLines      []string              // Previous render output for diff rendering
	startRow       int                   // Row where app starts (for cursor positioning)
}

// GoRenderFn is the signature for Go-native render functions
type GoRenderFn func(state env.Dict, width, height int) string

// NewInlineApp creates a new inline app with the given height
func NewInlineApp(height int) *InlineApp {
	return &InlineApp{
		height:      height,
		state:       *env.NewDict(make(map[string]any)),
		keyHandlers: make(map[string]env.Object),
		running:     false,
		stopChan:    make(chan struct{}),
		updateChan:  make(chan env.Dict, 100),
		prevLines:   make([]string, height),
	}
}

// SetState sets the app state
func (app *InlineApp) SetState(state env.Dict) {
	app.mu.Lock()
	defer app.mu.Unlock()
	app.state = state
}

// GetState returns a copy of the current state
func (app *InlineApp) GetState() env.Dict {
	app.mu.Lock()
	defer app.mu.Unlock()
	return app.state
}

// SetRenderFn sets the render function
func (app *InlineApp) SetRenderFn(fn env.Object) {
	app.mu.Lock()
	defer app.mu.Unlock()
	app.renderFn = fn
}

// SetKeyHandler registers a key handler
func (app *InlineApp) SetKeyHandler(key string, handler env.Object) {
	app.mu.Lock()
	defer app.mu.Unlock()
	app.keyHandlers[key] = handler
}

// renderLines renders lines to the terminal with differential rendering
func (app *InlineApp) renderLines(lines []string, height int) {
	// Pad lines to height
	for len(lines) < height {
		lines = append(lines, "")
	}

	// Start synchronized output for flicker-free rendering
	tuiSyncStart()
	defer tuiSyncEnd()

	// Differential rendering: only update changed lines
	for i := 0; i < height; i++ {
		newLine := ""
		if i < len(lines) {
			newLine = lines[i]
		}

		// Check if line changed
		if i < len(app.prevLines) && app.prevLines[i] == newLine {
			// Line unchanged, just move down
			fmt.Println()
			continue
		}

		// Line changed, clear and redraw
		term.ClearLine()
		fmt.Println(newLine)
	}

	// Update previous lines cache
	app.prevLines = make([]string, len(lines))
	copy(app.prevLines, lines)
}

// Render renders the app to the terminal
func (app *InlineApp) Render() {
	app.mu.Lock()
	renderFn := app.renderFn
	state := app.state
	height := app.height
	rendered := app.rendered
	app.mu.Unlock()

	if renderFn == nil {
		return
	}

	// Get actual terminal width
	width := term.GetTerminalColumns()
	if width < 20 {
		width = 80 // Fallback
	}

	// If already rendered, move cursor up to re-render in place
	if rendered {
		term.CurUp(height)
	}

	switch fn := renderFn.(type) {
	case env.Function:
		if app.ps != nil {
			// Call Rye render function with state and width
			psTemp := *app.ps
			psTemp.Res = nil

			// Create a new context for the function call
			fnCtx := env.NewEnv(psTemp.Ctx)

			// Set the first argument (state)
			if fn.Spec.Series.Len() > 0 {
				argWord := fn.Spec.Series.Get(0)
				if word, ok := argWord.(env.Word); ok {
					fnCtx.Set(word.Index, state)
				}
			}

			// Set the second argument (width) if function accepts it
			if fn.Spec.Series.Len() > 1 {
				argWord := fn.Spec.Series.Get(1)
				if word, ok := argWord.(env.Word); ok {
					fnCtx.Set(word.Index, *env.NewInteger(int64(width)))
				}
			}

			// Create new program state for the function
			psX := env.NewProgramState(fn.Body.Series, psTemp.Idx)
			psX.Ctx = fnCtx
			psX.PCtx = psTemp.PCtx
			psX.Gen = psTemp.Gen

			// Execute the function body with injection
			psX.Ser.SetPos(0)
			EvalBlockInj(psX, state, true)

			// Check for errors
			if psX.ErrorFlag {
				fmt.Println("Render error:", psX.Res.Inspect(*psX.Idx))
				app.mu.Lock()
				app.rendered = true
				app.mu.Unlock()
				return
			}
			if psX.FailureFlag {
				fmt.Println("Render failure:", psX.Res.Inspect(*psX.Idx))
				app.mu.Lock()
				app.rendered = true
				app.mu.Unlock()
				return
			}

			// Handle result: String, Block of strings, or assume direct printing
			switch res := psX.Res.(type) {
			case env.String:
				// Single string - split into lines
				lines := strings.Split(res.Value, "\n")
				app.renderLines(lines, height)
			case env.Block:
				// Block of strings
				lines := make([]string, 0, res.Series.Len())
				for i := 0; i < res.Series.Len(); i++ {
					item := res.Series.Get(i)
					if s, ok := item.(env.String); ok {
						lines = append(lines, s.Value)
					} else {
						// Convert non-string items to their string representation
						lines = append(lines, item.Print(*psX.Idx))
					}
				}
				app.renderLines(lines, height)
			default:
				// Function printed directly (legacy behavior)
				// We assume it printed exactly `height` lines
			}
		}
	case *env.Native:
		// Check if it's a GoRenderFn
		if goFn, ok := fn.Value.(GoRenderFn); ok {
			output := goFn(state, width, height)
			lines := strings.Split(output, "\n")
			app.renderLines(lines, height)
		}
	}

	app.mu.Lock()
	app.rendered = true
	app.mu.Unlock()
}

// Update updates the state and triggers a re-render
func (app *InlineApp) Update(updates env.Dict) {
	app.mu.Lock()
	// Merge updates into state
	for k, v := range updates.Data {
		app.state.Data[k] = v
	}
	app.mu.Unlock()

	// Trigger re-render
	app.Render()
}

// Start starts the app's event loop
func (app *InlineApp) Start(ps *env.ProgramState) error {
	app.mu.Lock()
	if app.running {
		app.mu.Unlock()
		return fmt.Errorf("app already running")
	}
	app.running = true
	app.ps = ps
	app.stopChan = make(chan struct{})
	app.mu.Unlock()

	// Initial render
	app.Render()

	// Start keyboard listener
	if err := keyboard.Open(); err != nil {
		return err
	}

	go app.eventLoop()

	return nil
}

// Stop stops the app
func (app *InlineApp) Stop() {
	app.mu.Lock()
	if !app.running {
		app.mu.Unlock()
		return
	}
	app.running = false
	close(app.stopChan)
	app.mu.Unlock()

	keyboard.Close()
}

// IsRunning returns whether the app is running
func (app *InlineApp) IsRunning() bool {
	app.mu.Lock()
	defer app.mu.Unlock()
	return app.running
}

// eventLoop handles keyboard events and updates
func (app *InlineApp) eventLoop() {
	for {
		select {
		case <-app.stopChan:
			return
		default:
			// Check for keyboard input with timeout
			char, key, err := keyboard.GetKey()
			if err != nil {
				continue
			}

			keyStr := ""
			switch key {
			case keyboard.KeyArrowUp:
				keyStr = "up"
			case keyboard.KeyArrowDown:
				keyStr = "down"
			case keyboard.KeyArrowLeft:
				keyStr = "left"
			case keyboard.KeyArrowRight:
				keyStr = "right"
			case keyboard.KeyEnter:
				keyStr = "enter"
			case keyboard.KeyEsc:
				keyStr = "escape"
			case keyboard.KeyBackspace, keyboard.KeyBackspace2:
				keyStr = "backspace"
			case keyboard.KeyTab:
				keyStr = "tab"
			case keyboard.KeySpace:
				keyStr = "space"
			case keyboard.KeyCtrlC:
				keyStr = "ctrl-c"
				// Default behavior: stop the app
				app.Stop()
				return
			default:
				if char != 0 {
					keyStr = string(char)
				}
			}

			if keyStr != "" {
				app.handleKey(keyStr)
			}
		}
	}
}

// SetDefaultHandler sets the default handler for unregistered keys
func (app *InlineApp) SetDefaultHandler(handler env.Object) {
	app.mu.Lock()
	defer app.mu.Unlock()
	app.defaultHandler = handler
}

// handleKey handles a key press
func (app *InlineApp) handleKey(key string) {
	app.mu.Lock()
	handler, exists := app.keyHandlers[key]
	if !exists {
		// Use default handler if no specific handler registered
		handler = app.defaultHandler
		exists = handler != nil
	}
	ps := app.ps
	app.mu.Unlock()

	if !exists || ps == nil {
		return
	}

	switch h := handler.(type) {
	case env.Function:
		// Call the handler function with 2 arguments (state, key)
		psTemp := *ps
		psTemp.Res = nil

		// Create a new context for the function call
		// Use program state's context as parent to access script-defined functions
		fnCtx := env.NewEnv(psTemp.Ctx)

		// Set the first argument (state)
		currentState := app.GetState()
		if h.Spec.Series.Len() > 0 {
			argWord := h.Spec.Series.Get(0)
			if word, ok := argWord.(env.Word); ok {
				fnCtx.Set(word.Index, currentState)
			}
		}

		// Set the second argument (key) if function accepts it
		if h.Spec.Series.Len() > 1 {
			argWord := h.Spec.Series.Get(1)
			if word, ok := argWord.(env.Word); ok {
				fnCtx.Set(word.Index, *env.NewString(key))
			}
		}

		// Create new program state for the function
		psX := env.NewProgramState(h.Body.Series, psTemp.Idx)
		psX.Ctx = fnCtx
		psX.PCtx = psTemp.PCtx
		psX.Gen = psTemp.Gen

		// Execute the function body with injection
		psX.Ser.SetPos(0)
		EvalBlockInj(psX, currentState, true)

		// Check for errors
		if psX.ErrorFlag {
			fmt.Println("Key handler error:", psX.Res.Inspect(*psX.Idx))
			return
		}
		if psX.FailureFlag {
			fmt.Println("Key handler failure:", psX.Res.Inspect(*psX.Idx))
			return
		}

		psTemp.Res = psX.Res

		// If handler returned a Dict, use it to update state
		switch d := psTemp.Res.(type) {
		case env.Dict:
			app.Update(d) // Update already calls Render
		case *env.Dict:
			app.Update(*d) // Update already calls Render
		default:
			app.Render() // Only render if no state update
		}
	case env.Block:
		// Evaluate the block
		psTemp := *ps
		psTemp.Ser = h.Series
		psTemp.Ser.Reset()
		EvalBlock(&psTemp)

		// Check for errors
		if psTemp.ErrorFlag {
			fmt.Println("Block handler error:", psTemp.Res.Inspect(*psTemp.Idx))
			return
		}
		if psTemp.FailureFlag {
			fmt.Println("Block handler failure:", psTemp.Res.Inspect(*psTemp.Idx))
			return
		}

		// If block result is a Dict, use it to update state
		switch d := psTemp.Res.(type) {
		case env.Dict:
			app.Update(d) // Update already calls Render
		case *env.Dict:
			app.Update(*d) // Update already calls Render
		default:
			app.Render() // Only render if no state update
		}
	}
}

// WaitForStop blocks until the app stops
func (app *InlineApp) WaitForStop() {
	for app.IsRunning() {
		time.Sleep(50 * time.Millisecond)
	}
}

// ============================================================================
// Input Component (Focusable)
// ============================================================================

// InputComponent is a single-line text input field (UIComponent)
type InputComponent struct {
	value       string
	cursor      int
	placeholder string
	width       int
	focused     bool
	theme       *Theme
	onChange    func(string)
	onSubmit    func(string)
	cachedWidth int
	cachedLines []string
	mu          sync.Mutex
}

// NewInputComponent creates a new input component
func NewInputComponent(placeholder string, width int) *InputComponent {
	return &InputComponent{
		placeholder: placeholder,
		width:       width,
		theme:       globalTheme,
	}
}

// SetValue sets the text value
func (inp *InputComponent) SetValue(value string) {
	inp.mu.Lock()
	defer inp.mu.Unlock()
	inp.value = value
	if inp.cursor > len(inp.value) {
		inp.cursor = len(inp.value)
	}
	inp.invalidate()
}

// GetValue returns the current value
func (inp *InputComponent) GetValue() string {
	inp.mu.Lock()
	defer inp.mu.Unlock()
	return inp.value
}

// SetPlaceholder sets the placeholder text
func (inp *InputComponent) SetPlaceholder(placeholder string) {
	inp.mu.Lock()
	defer inp.mu.Unlock()
	inp.placeholder = placeholder
	inp.invalidate()
}

// SetTheme sets the theme
func (inp *InputComponent) SetTheme(theme *Theme) {
	inp.mu.Lock()
	defer inp.mu.Unlock()
	inp.theme = theme
	inp.invalidate()
}

// SetOnChange sets the change callback
func (inp *InputComponent) SetOnChange(fn func(string)) {
	inp.mu.Lock()
	defer inp.mu.Unlock()
	inp.onChange = fn
}

// SetOnSubmit sets the submit callback
func (inp *InputComponent) SetOnSubmit(fn func(string)) {
	inp.mu.Lock()
	defer inp.mu.Unlock()
	inp.onSubmit = fn
}

// SetFocused sets the focus state (FocusableComponent)
func (inp *InputComponent) SetFocused(focused bool) {
	inp.mu.Lock()
	defer inp.mu.Unlock()
	inp.focused = focused
	inp.invalidate()
}

// IsFocused returns whether the input is focused (FocusableComponent)
func (inp *InputComponent) IsFocused() bool {
	inp.mu.Lock()
	defer inp.mu.Unlock()
	return inp.focused
}

// HandleKey handles a key press (FocusableComponent)
func (inp *InputComponent) HandleKey(key string) bool {
	inp.mu.Lock()
	defer inp.mu.Unlock()

	if !inp.focused {
		return false
	}

	changed := false

	switch key {
	case "backspace":
		if inp.cursor > 0 {
			inp.value = inp.value[:inp.cursor-1] + inp.value[inp.cursor:]
			inp.cursor--
			changed = true
		}
	case "delete":
		if inp.cursor < len(inp.value) {
			inp.value = inp.value[:inp.cursor] + inp.value[inp.cursor+1:]
			changed = true
		}
	case "left":
		if inp.cursor > 0 {
			inp.cursor--
		}
	case "right":
		if inp.cursor < len(inp.value) {
			inp.cursor++
		}
	case "home":
		inp.cursor = 0
	case "end":
		inp.cursor = len(inp.value)
	case "enter":
		if inp.onSubmit != nil {
			inp.mu.Unlock()
			inp.onSubmit(inp.value)
			inp.mu.Lock()
		}
		return true
	default:
		// Regular character
		if len(key) == 1 && key[0] >= 32 {
			inp.value = inp.value[:inp.cursor] + key + inp.value[inp.cursor:]
			inp.cursor++
			changed = true
		}
	}

	if changed {
		inp.invalidate()
		if inp.onChange != nil {
			inp.mu.Unlock()
			inp.onChange(inp.value)
			inp.mu.Lock()
		}
	}

	return true // Input handles all keys when focused
}

// Render renders the input (UIComponent)
func (inp *InputComponent) Render(width int) []string {
	inp.mu.Lock()
	defer inp.mu.Unlock()

	if inp.cachedLines != nil && inp.cachedWidth == width {
		return inp.cachedLines
	}

	displayWidth := inp.width
	if displayWidth <= 0 || displayWidth > width {
		displayWidth = width
	}

	var display string

	if inp.value == "" && inp.placeholder != "" && !inp.focused {
		// Show placeholder
		display = inp.theme.Apply("placeholder", inp.placeholder)
	} else {
		// Show value with cursor
		if inp.focused {
			if inp.cursor < len(inp.value) {
				before := inp.value[:inp.cursor]
				at := string(inp.value[inp.cursor])
				after := inp.value[inp.cursor+1:]
				display = before + inp.theme.Apply("cursor", at) + after
			} else {
				display = inp.value + inp.theme.Apply("cursor", " ")
			}
		} else {
			display = inp.value
		}
	}

	// Truncate to width
	display = TruncateToWidth(display, displayWidth, "")

	inp.cachedWidth = width
	inp.cachedLines = []string{display}
	return inp.cachedLines
}

// Invalidate clears the cache (UIComponent)
func (inp *InputComponent) Invalidate() {
	inp.mu.Lock()
	defer inp.mu.Unlock()
	inp.invalidate()
}

func (inp *InputComponent) invalidate() {
	inp.cachedWidth = 0
	inp.cachedLines = nil
}

// ============================================================================
// Built-in Go Components
// ============================================================================

// SpinnerComponent is a Go-native animated spinner
type SpinnerComponent struct {
	style    string
	label    string
	frame    int
	running  bool
	stopChan chan struct{}
	mu       sync.Mutex
}

// SpinnerFrames holds animation frames for different styles
var SpinnerFrames = map[string][]string{
	"line":   {"|", "/", "-", "\\"},
	"dots":   {"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
	"circle": {"◐", "◓", "◑", "◒"},
	"bounce": {"⠁", "⠂", "⠄", "⠂"},
	"arc":    {"◜", "◠", "◝", "◞", "◡", "◟"},
}

// NewSpinner creates a new spinner component
func NewSpinner(style, label string) *SpinnerComponent {
	if _, ok := SpinnerFrames[style]; !ok {
		style = "line"
	}
	return &SpinnerComponent{
		style:    style,
		label:    label,
		frame:    0,
		stopChan: make(chan struct{}),
	}
}

// Render returns the current spinner frame
func (s *SpinnerComponent) Render() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	frames := SpinnerFrames[s.style]
	return fmt.Sprintf("%s %s", frames[s.frame%len(frames)], s.label)
}

// Start starts the spinner animation
func (s *SpinnerComponent) Start() {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}
	s.running = true
	s.stopChan = make(chan struct{})
	s.mu.Unlock()

	// Print initial frame
	fmt.Print(s.Render())

	go func() {
		ticker := time.NewTicker(80 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-s.stopChan:
				return
			case <-ticker.C:
				s.mu.Lock()
				s.frame++
				s.mu.Unlock()

				// Move to start of line and reprint
				fmt.Print("\r")
				term.ClearLine()
				fmt.Print(s.Render())
			}
		}
	}()
}

// Stop stops the spinner animation
func (s *SpinnerComponent) Stop() {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}
	s.running = false
	close(s.stopChan)
	s.mu.Unlock()

	// Clear the spinner line
	fmt.Print("\r")
	term.ClearLine()
}

// SetLabel updates the spinner label
func (s *SpinnerComponent) SetLabel(label string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.label = label
}

// ProgressBarComponent is a Go-native progress bar
type ProgressBarComponent struct {
	width    int
	complete float64 // 0.0 to 1.0
	style    string  // "block", "line", "ascii"
	label    string
	mu       sync.Mutex
}

// NewProgressBar creates a new progress bar
func NewProgressBar(width int, style, label string) *ProgressBarComponent {
	return &ProgressBarComponent{
		width:    width,
		complete: 0,
		style:    style,
		label:    label,
	}
}

// SetProgress sets the progress (0.0 to 1.0)
func (p *ProgressBarComponent) SetProgress(complete float64) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if complete < 0 {
		complete = 0
	}
	if complete > 1 {
		complete = 1
	}
	p.complete = complete
}

// Render returns the progress bar string
func (p *ProgressBarComponent) Render() string {
	p.mu.Lock()
	defer p.mu.Unlock()

	filled := int(float64(p.width) * p.complete)
	empty := p.width - filled

	var bar string
	switch p.style {
	case "block":
		bar = strings.Repeat("█", filled) + strings.Repeat("░", empty)
	case "line":
		bar = strings.Repeat("━", filled) + strings.Repeat("─", empty)
	default: // ascii
		bar = strings.Repeat("#", filled) + strings.Repeat("-", empty)
	}

	pct := int(p.complete * 100)
	if p.label != "" {
		return fmt.Sprintf("%s [%s] %3d%%", p.label, bar, pct)
	}
	return fmt.Sprintf("[%s] %3d%%", bar, pct)
}

// TextInputComponent is a reactive text input field
type TextInputComponent struct {
	value       string     // Current text value
	cursor      int        // Cursor position
	placeholder string     // Placeholder text
	maxLen      int        // Maximum length (0 = unlimited)
	width       int        // Display width
	focused     bool       // Is the input focused
	onChange    env.Object // Handler called when value changes
	onSubmit    env.Object // Handler called when Enter is pressed
	mu          sync.Mutex
}

// NewTextInput creates a new text input component
func NewTextInput(placeholder string, width int) *TextInputComponent {
	return &TextInputComponent{
		value:       "",
		cursor:      0,
		placeholder: placeholder,
		maxLen:      0,
		width:       width,
		focused:     false,
	}
}

// Value returns the current text value
func (t *TextInputComponent) Value() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.value
}

// SetValue sets the text value
func (t *TextInputComponent) SetValue(value string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.maxLen > 0 && len(value) > t.maxLen {
		value = value[:t.maxLen]
	}
	t.value = value
	if t.cursor > len(t.value) {
		t.cursor = len(t.value)
	}
}

// SetMaxLen sets the maximum length
func (t *TextInputComponent) SetMaxLen(maxLen int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.maxLen = maxLen
}

// Focus focuses the input
func (t *TextInputComponent) Focus() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.focused = true
}

// Blur unfocuses the input
func (t *TextInputComponent) Blur() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.focused = false
}

// IsFocused returns whether the input is focused
func (t *TextInputComponent) IsFocused() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.focused
}

// HandleKey handles a key press and returns true if the value changed
func (t *TextInputComponent) HandleKey(key string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.focused {
		return false
	}

	changed := false

	switch key {
	case "backspace":
		if t.cursor > 0 {
			t.value = t.value[:t.cursor-1] + t.value[t.cursor:]
			t.cursor--
			changed = true
		}
	case "delete":
		if t.cursor < len(t.value) {
			t.value = t.value[:t.cursor] + t.value[t.cursor+1:]
			changed = true
		}
	case "left":
		if t.cursor > 0 {
			t.cursor--
		}
	case "right":
		if t.cursor < len(t.value) {
			t.cursor++
		}
	case "home":
		t.cursor = 0
	case "end":
		t.cursor = len(t.value)
	default:
		// Regular character
		if len(key) == 1 {
			if t.maxLen == 0 || len(t.value) < t.maxLen {
				t.value = t.value[:t.cursor] + key + t.value[t.cursor:]
				t.cursor++
				changed = true
			}
		}
	}

	return changed
}

// Render returns the rendered text input as a string
func (t *TextInputComponent) Render() string {
	t.mu.Lock()
	defer t.mu.Unlock()

	// If empty and has placeholder, show placeholder
	if t.value == "" && t.placeholder != "" && !t.focused {
		return fmt.Sprintf("\033[90m%s\033[0m", t.placeholder) // Gray placeholder
	}

	// Build the display string
	display := t.value

	// Pad or truncate to width
	if t.width > 0 {
		if len(display) > t.width {
			display = display[:t.width]
		} else {
			display = display + strings.Repeat(" ", t.width-len(display))
		}
	}

	// If focused, show cursor
	if t.focused {
		if t.cursor < len(t.value) {
			// Cursor in middle - highlight character at cursor
			beforeCursor := t.value[:t.cursor]
			atCursor := string(t.value[t.cursor])
			afterCursor := t.value[t.cursor+1:]
			display = beforeCursor + "\033[7m" + atCursor + "\033[0m" + afterCursor
		} else {
			// Cursor at end - show block cursor
			display = t.value + "\033[7m \033[0m"
		}

		// Pad to width
		if t.width > 0 && len(t.value) < t.width-1 {
			display = display + strings.Repeat(" ", t.width-len(t.value)-1)
		}
	}

	return display
}

// RenderWithBorder returns the input with a border
func (t *TextInputComponent) RenderWithBorder() string {
	content := t.Render()
	return fmt.Sprintf("[%s]", content)
}

// ============================================================================
// Builtins
// ============================================================================

var Builtins_termui = map[string]*env.Builtin{

	// ##### Inline Terminal UI ##### "Reactive inline terminal UI components"

	// Tests:
	// equal { inline-app 3 |type? } 'native
	// Args:
	// * height: Integer specifying the number of lines this app occupies
	// Returns:
	// * a new InlineApp native object
	"inline-app": {
		Argsn: 1,
		Doc:   "Creates a new inline terminal app that occupies the specified number of lines.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch height := arg0.(type) {
			case env.Integer:
				app := NewInlineApp(int(height.Value))
				return *env.NewNative(ps.Idx, app, "inline-app")
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType}, "inline-app")
			}
		},
	},

	// Tests:
	// equal { app: inline-app 1 , app .render fn { s w } { "test" } |type? } 'native
	// Args:
	// * app: InlineApp native object
	// * render-fn: Function that takes (state Dict, width Integer) and returns String or Block of strings
	// Returns:
	// * the app object
	// Example:
	// ; Return a single string (will be split on newlines)
	// app .render fn { state width } { "Line 1\nLine 2" }
	// ; Or return a block of strings (one per line)
	// app .render fn { state width } { { "Line 1" "Line 2" } }
	// ; Legacy: functions that print directly still work
	"inline-app//render": {
		Argsn: 2,
		Doc:   "Sets the render function for an inline app. Function receives (state, width) and should return String or Block of strings.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch native := arg0.(type) {
			case env.Native:
				app, ok := native.Value.(*InlineApp)
				if !ok {
					return MakeBuiltinError(ps, "Expected InlineApp", "inline-app//render")
				}
				switch fn := arg1.(type) {
				case env.Function:
					app.SetRenderFn(fn)
					return arg0
				case env.Block:
					// Wrap block in a function
					app.SetRenderFn(fn)
					return arg0
				default:
					return MakeArgError(ps, 2, []env.Type{env.FunctionType, env.BlockType}, "inline-app//render")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "inline-app//render")
			}
		},
	},

	// Tests:
	// equal { app: inline-app 1 , app .state dict { "x" 1 } , app .get-state -> "x" } 1
	// Args:
	// * app: InlineApp native object
	// * state: Dict containing the initial/new state
	// Returns:
	// * the app object
	"inline-app//state": {
		Argsn: 2,
		Doc:   "Sets the state for an inline app.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch native := arg0.(type) {
			case env.Native:
				app, ok := native.Value.(*InlineApp)
				if !ok {
					return MakeBuiltinError(ps, "Expected InlineApp", "inline-app//state")
				}
				switch state := arg1.(type) {
				case env.Dict:
					app.SetState(state)
					return arg0
				default:
					return MakeArgError(ps, 2, []env.Type{env.DictType}, "inline-app//state")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "inline-app//state")
			}
		},
	},

	// Tests:
	// equal { app: inline-app 1 , app .state dict { "x" 1 } , app .get-state -> "x" } 1
	// Args:
	// * app: InlineApp native object
	// Returns:
	// * the current state Dict
	"inline-app//State?": {
		Argsn: 1,
		Doc:   "Gets the current state from an inline app.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch native := arg0.(type) {
			case env.Native:
				app, ok := native.Value.(*InlineApp)
				if !ok {
					return MakeBuiltinError(ps, "Expected InlineApp", "inline-app//get-state")
				}
				return app.GetState()
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "inline-app//get-state")
			}
		},
	},

	// Tests:
	// ; inline-app 1 .on-key "q" fn { s } { print "quit" }
	// Args:
	// * app: InlineApp native object
	// * key: String representing the key (e.g., "q", "enter", "up")
	// * handler: Function or block to execute when key is pressed
	// Returns:
	// * the app object
	"inline-app//on-key": {
		Argsn: 3,
		Doc:   "Registers a key handler for the inline app.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch native := arg0.(type) {
			case env.Native:
				app, ok := native.Value.(*InlineApp)
				if !ok {
					return MakeBuiltinError(ps, "Expected InlineApp", "inline-app//on-key")
				}
				switch key := arg1.(type) {
				case env.String:
					switch handler := arg2.(type) {
					case env.Function, env.Block:
						app.SetKeyHandler(key.Value, handler)
						return arg0
					default:
						return MakeArgError(ps, 3, []env.Type{env.FunctionType, env.BlockType}, "inline-app//on-key")
					}
				default:
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "inline-app//on-key")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "inline-app//on-key")
			}
		},
	},

	// Tests:
	// ; inline-app 1 .on-any-key fn { s key } { print key }
	// Args:
	// * app: InlineApp native object
	// * handler: Function that takes (state, key) to execute for any unhandled key
	// Returns:
	// * the app object
	"inline-app//on-any-key": {
		Argsn: 2,
		Doc:   "Registers a default handler for any key not specifically handled. Function receives (state, key).",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch native := arg0.(type) {
			case env.Native:
				app, ok := native.Value.(*InlineApp)
				if !ok {
					return MakeBuiltinError(ps, "Expected InlineApp", "inline-app//on-any-key")
				}
				switch handler := arg1.(type) {
				case env.Function, env.Block:
					app.SetDefaultHandler(handler)
					return arg0
				default:
					return MakeArgError(ps, 2, []env.Type{env.FunctionType, env.BlockType}, "inline-app//on-any-key")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "inline-app//on-any-key")
			}
		},
	},

	// Tests:
	// ; app: inline-app 1 , app .render fn { } { print "hi" } , app .start
	// Args:
	// * app: InlineApp native object
	// Returns:
	// * the app object
	"inline-app//start": {
		Argsn: 1,
		Doc:   "Starts the inline app, entering raw mode and beginning the render/event loop.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch native := arg0.(type) {
			case env.Native:
				app, ok := native.Value.(*InlineApp)
				if !ok {
					return MakeBuiltinError(ps, "Expected InlineApp", "inline-app//start")
				}
				err := app.Start(ps)
				if err != nil {
					return MakeBuiltinError(ps, err.Error(), "inline-app//start")
				}
				return arg0
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "inline-app//start")
			}
		},
	},

	// Tests:
	// ; app .stop
	// Args:
	// * app: InlineApp native object
	// Returns:
	// * the app object
	"inline-app//stop": {
		Argsn: 1,
		Doc:   "Stops the inline app and restores the terminal.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch native := arg0.(type) {
			case env.Native:
				app, ok := native.Value.(*InlineApp)
				if !ok {
					return MakeBuiltinError(ps, "Expected InlineApp", "inline-app//stop")
				}
				app.Stop()
				return arg0
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "inline-app//stop")
			}
		},
	},

	// Tests:
	// ; app .update dict { "x" 2 }
	// Args:
	// * app: InlineApp native object
	// * updates: Dict containing state updates to merge
	// Returns:
	// * the app object
	"inline-app//update": {
		Argsn: 2,
		Doc:   "Updates the inline app state and triggers a re-render.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch native := arg0.(type) {
			case env.Native:
				app, ok := native.Value.(*InlineApp)
				if !ok {
					return MakeBuiltinError(ps, "Expected InlineApp", "inline-app//update")
				}
				switch updates := arg1.(type) {
				case env.Dict:
					app.Update(updates)
					return arg0
				default:
					return MakeArgError(ps, 2, []env.Type{env.DictType}, "inline-app//update")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "inline-app//update")
			}
		},
	},

	// Tests:
	// ; app .wait
	// Args:
	// * app: InlineApp native object
	// Returns:
	// * the app object (after it stops)
	"inline-app//wait": {
		Argsn: 1,
		Doc:   "Blocks until the inline app stops.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch native := arg0.(type) {
			case env.Native:
				app, ok := native.Value.(*InlineApp)
				if !ok {
					return MakeBuiltinError(ps, "Expected InlineApp", "inline-app//wait")
				}
				app.WaitForStop()
				return arg0
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "inline-app//wait")
			}
		},
	},

	// ##### Go-Native Components ##### "Pre-built Go components for common UI patterns"

	// Tests:
	// equal { spinner "dots" "Loading" |type? } 'native
	// Args:
	// * style: String spinner style ("line", "dots", "circle", "bounce", "arc")
	// * label: String label to display next to spinner
	// Returns:
	// * a Spinner native object
	"spinner": {
		Argsn: 2,
		Doc:   "Creates a new animated spinner with the given style and label.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch style := arg0.(type) {
			case env.String:
				switch label := arg1.(type) {
				case env.String:
					spinner := NewSpinner(style.Value, label.Value)
					return *env.NewNative(ps.Idx, spinner, "spinner")
				default:
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "spinner")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "spinner")
			}
		},
	},

	// Tests:
	// ; spin: spinner "dots" "Loading" , spin .start
	// Args:
	// * spinner: Spinner native object
	// Returns:
	// * the spinner object
	"spinner//start": {
		Argsn: 1,
		Doc:   "Starts the spinner animation.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch native := arg0.(type) {
			case env.Native:
				spinner, ok := native.Value.(*SpinnerComponent)
				if !ok {
					return MakeBuiltinError(ps, "Expected Spinner", "spinner//start")
				}
				spinner.Start()
				return arg0
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "spinner//start")
			}
		},
	},

	// Tests:
	// ; spin .stop
	// Args:
	// * spinner: Spinner native object
	// Returns:
	// * the spinner object
	"spinner//stop": {
		Argsn: 1,
		Doc:   "Stops the spinner animation and clears the line.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch native := arg0.(type) {
			case env.Native:
				spinner, ok := native.Value.(*SpinnerComponent)
				if !ok {
					return MakeBuiltinError(ps, "Expected Spinner", "spinner//stop")
				}
				spinner.Stop()
				return arg0
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "spinner//stop")
			}
		},
	},

	// Tests:
	// ; spin .label "New label"
	// Args:
	// * spinner: Spinner native object
	// * label: String new label
	// Returns:
	// * the spinner object
	"spinner//label": {
		Argsn: 2,
		Doc:   "Updates the spinner label.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch native := arg0.(type) {
			case env.Native:
				spinner, ok := native.Value.(*SpinnerComponent)
				if !ok {
					return MakeBuiltinError(ps, "Expected Spinner", "spinner//label")
				}
				switch label := arg1.(type) {
				case env.String:
					spinner.SetLabel(label.Value)
					return arg0
				default:
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "spinner//label")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "spinner//label")
			}
		},
	},

	// Tests:
	// equal { progress-bar 20 "block" "Progress" |type? } 'native
	// Args:
	// * width: Integer width of the progress bar
	// * style: String style ("block", "line", "ascii")
	// * label: String label
	// Returns:
	// * a ProgressBar native object
	"progress-bar": {
		Argsn: 3,
		Doc:   "Creates a new progress bar with the given width, style, and label.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch width := arg0.(type) {
			case env.Integer:
				switch style := arg1.(type) {
				case env.String:
					switch label := arg2.(type) {
					case env.String:
						bar := NewProgressBar(int(width.Value), style.Value, label.Value)
						return *env.NewNative(ps.Idx, bar, "progress-bar")
					default:
						return MakeArgError(ps, 3, []env.Type{env.StringType}, "progress-bar")
					}
				default:
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "progress-bar")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType}, "progress-bar")
			}
		},
	},

	// Tests:
	// ; bar .progress 0.5
	// Args:
	// * bar: ProgressBar native object
	// * value: Decimal progress value (0.0 to 1.0)
	// Returns:
	// * the progress bar object
	"progress-bar//Progress": {
		Argsn: 2,
		Doc:   "Sets the progress value (0.0 to 1.0).",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch native := arg0.(type) {
			case env.Native:
				bar, ok := native.Value.(*ProgressBarComponent)
				if !ok {
					return MakeBuiltinError(ps, "Expected ProgressBar", "progress-bar//progress")
				}
				switch val := arg1.(type) {
				case env.Decimal:
					bar.SetProgress(val.Value)
					return arg0
				case env.Integer:
					bar.SetProgress(float64(val.Value))
					return arg0
				default:
					return MakeArgError(ps, 2, []env.Type{env.DecimalType, env.IntegerType}, "progress-bar//progress")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "progress-bar//progress")
			}
		},
	},

	// Tests:
	// ; bar .print
	// Args:
	// * bar: ProgressBar native object
	// Returns:
	// * the progress bar object
	"progress-bar//Print": {
		Argsn: 1,
		Doc:   "Prints the progress bar (in place, clearing line first).",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch native := arg0.(type) {
			case env.Native:
				bar, ok := native.Value.(*ProgressBarComponent)
				if !ok {
					return MakeBuiltinError(ps, "Expected ProgressBar", "progress-bar//print")
				}
				fmt.Print("\r")
				term.ClearLine()
				fmt.Print(bar.Render())
				return arg0
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "progress-bar//print")
			}
		},
	},

	// Tests:
	// equal { progress-bar 20 "block" "Test" .render } "Test [                    ]   0%"
	// Args:
	// * bar: ProgressBar native object
	// Returns:
	// * String representation of the progress bar
	"progress-bar//render": {
		Argsn: 1,
		Doc:   "Returns the string representation of the progress bar.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch native := arg0.(type) {
			case env.Native:
				bar, ok := native.Value.(*ProgressBarComponent)
				if !ok {
					return MakeBuiltinError(ps, "Expected ProgressBar", "progress-bar//render")
				}
				return *env.NewString(bar.Render())
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "progress-bar//render")
			}
		},
	},

	// ##### Text Input Component ##### "Reactive text input field"

	// Tests:
	// equal { text-input "Enter name" 20 |type? } 'native
	// Args:
	// * placeholder: String placeholder text
	// * width: Integer display width
	// Returns:
	// * a TextInput native object
	"text-input": {
		Argsn: 2,
		Doc:   "Creates a new text input component with placeholder and width.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch placeholder := arg0.(type) {
			case env.String:
				switch width := arg1.(type) {
				case env.Integer:
					input := NewTextInput(placeholder.Value, int(width.Value))
					return *env.NewNative(ps.Idx, input, "text-input")
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "text-input")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "text-input")
			}
		},
	},

	// Tests:
	// equal { text-input "" 20 .value? } ""
	// Args:
	// * input: TextInput native object
	// Returns:
	// * String current value
	"text-input//value?": {
		Argsn: 1,
		Doc:   "Returns the current text value.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch native := arg0.(type) {
			case env.Native:
				input, ok := native.Value.(*TextInputComponent)
				if !ok {
					return MakeBuiltinError(ps, "Expected TextInput", "text-input//value?")
				}
				return *env.NewString(input.Value())
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "text-input//value?")
			}
		},
	},

	// Tests:
	// equal { text-input "" 20 .set-value "hello" .value? } "hello"
	// Args:
	// * input: TextInput native object
	// * value: String new value
	// Returns:
	// * the input object
	"text-input//set-value": {
		Argsn: 2,
		Doc:   "Sets the text value.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch native := arg0.(type) {
			case env.Native:
				input, ok := native.Value.(*TextInputComponent)
				if !ok {
					return MakeBuiltinError(ps, "Expected TextInput", "text-input//set-value")
				}
				switch val := arg1.(type) {
				case env.String:
					input.SetValue(val.Value)
					return arg0
				default:
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "text-input//set-value")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "text-input//set-value")
			}
		},
	},

	// Args:
	// * input: TextInput native object
	// * maxLen: Integer maximum length
	// Returns:
	// * the input object
	"text-input//max-len": {
		Argsn: 2,
		Doc:   "Sets the maximum length.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch native := arg0.(type) {
			case env.Native:
				input, ok := native.Value.(*TextInputComponent)
				if !ok {
					return MakeBuiltinError(ps, "Expected TextInput", "text-input//max-len")
				}
				switch maxLen := arg1.(type) {
				case env.Integer:
					input.SetMaxLen(int(maxLen.Value))
					return arg0
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "text-input//max-len")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "text-input//max-len")
			}
		},
	},

	// Args:
	// * input: TextInput native object
	// Returns:
	// * the input object
	"text-input//focus": {
		Argsn: 1,
		Doc:   "Focuses the text input (shows cursor).",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch native := arg0.(type) {
			case env.Native:
				input, ok := native.Value.(*TextInputComponent)
				if !ok {
					return MakeBuiltinError(ps, "Expected TextInput", "text-input//focus")
				}
				input.Focus()
				return arg0
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "text-input//focus")
			}
		},
	},

	// Args:
	// * input: TextInput native object
	// Returns:
	// * the input object
	"text-input//blur": {
		Argsn: 1,
		Doc:   "Unfocuses the text input (hides cursor).",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch native := arg0.(type) {
			case env.Native:
				input, ok := native.Value.(*TextInputComponent)
				if !ok {
					return MakeBuiltinError(ps, "Expected TextInput", "text-input//blur")
				}
				input.Blur()
				return arg0
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "text-input//blur")
			}
		},
	},

	// Args:
	// * input: TextInput native object
	// Returns:
	// * Boolean whether focused
	"text-input//focused?": {
		Argsn: 1,
		Doc:   "Returns whether the input is focused.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch native := arg0.(type) {
			case env.Native:
				input, ok := native.Value.(*TextInputComponent)
				if !ok {
					return MakeBuiltinError(ps, "Expected TextInput", "text-input//focused?")
				}
				return *env.NewInteger(boolToInt64(input.IsFocused()))
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "text-input//focused?")
			}
		},
	},

	// Args:
	// * input: TextInput native object
	// * key: String key name (e.g., "a", "backspace", "left")
	// Returns:
	// * Integer 1 if value changed, 0 otherwise
	"text-input//handle-key": {
		Argsn: 2,
		Doc:   "Handles a key press. Returns 1 if value changed.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch native := arg0.(type) {
			case env.Native:
				input, ok := native.Value.(*TextInputComponent)
				if !ok {
					return MakeBuiltinError(ps, "Expected TextInput", "text-input//handle-key")
				}
				switch key := arg1.(type) {
				case env.String:
					changed := input.HandleKey(key.Value)
					return *env.NewInteger(boolToInt64(changed))
				default:
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "text-input//handle-key")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "text-input//handle-key")
			}
		},
	},

	// Args:
	// * input: TextInput native object
	// Returns:
	// * String rendered input
	"text-input//render": {
		Argsn: 1,
		Doc:   "Returns the rendered text input string with cursor.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch native := arg0.(type) {
			case env.Native:
				input, ok := native.Value.(*TextInputComponent)
				if !ok {
					return MakeBuiltinError(ps, "Expected TextInput", "text-input//render")
				}
				return *env.NewString(input.Render())
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "text-input//render")
			}
		},
	},

	// Args:
	// * input: TextInput native object
	// Returns:
	// * String rendered input with border
	"text-input//render-bordered": {
		Argsn: 1,
		Doc:   "Returns the rendered text input with a border.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch native := arg0.(type) {
			case env.Native:
				input, ok := native.Value.(*TextInputComponent)
				if !ok {
					return MakeBuiltinError(ps, "Expected TextInput", "text-input//render-bordered")
				}
				return *env.NewString(input.RenderWithBorder())
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "text-input//render-bordered")
			}
		},
	},

	// ##### Theming ##### "Color and style theming"

	// Tests:
	// equal { ui-theme |type? } 'native
	// Returns:
	// * a new Theme with default colors
	"ui-theme": {
		Argsn: 0,
		Doc:   "Creates a new theme with default colors.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			theme := DefaultTheme()
			return *env.NewNative(ps.Idx, theme, "ui-theme")
		},
	},

	// Args:
	// * theme: Theme native object
	// * name: String color name
	// * value: String ANSI escape code
	// Returns:
	// * the theme object
	"ui-theme//set-color": {
		Argsn: 3,
		Doc:   "Sets a color/style in the theme.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch native := arg0.(type) {
			case env.Native:
				theme, ok := native.Value.(*Theme)
				if !ok {
					return MakeBuiltinError(ps, "Expected Theme", "ui-theme//set-color")
				}
				switch name := arg1.(type) {
				case env.String:
					switch value := arg2.(type) {
					case env.String:
						theme.Set(name.Value, value.Value)
						return arg0
					default:
						return MakeArgError(ps, 3, []env.Type{env.StringType}, "ui-theme//set-color")
					}
				default:
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "ui-theme//set-color")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "ui-theme//set-color")
			}
		},
	},

	// Args:
	// * theme: Theme native object
	// * name: String color name
	// Returns:
	// * String ANSI escape code or empty string
	"ui-theme//get-color": {
		Argsn: 2,
		Doc:   "Gets a color/style from the theme.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch native := arg0.(type) {
			case env.Native:
				theme, ok := native.Value.(*Theme)
				if !ok {
					return MakeBuiltinError(ps, "Expected Theme", "ui-theme//get-color")
				}
				switch name := arg1.(type) {
				case env.String:
					return *env.NewString(theme.Get(name.Value))
				default:
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "ui-theme//get-color")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "ui-theme//get-color")
			}
		},
	},

	// Args:
	// * theme: Theme native object
	// * name: String color name
	// * text: String text to style
	// Returns:
	// * String styled text with reset
	"ui-theme//style": {
		Argsn: 3,
		Doc:   "Applies a style to text and adds reset code.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch native := arg0.(type) {
			case env.Native:
				theme, ok := native.Value.(*Theme)
				if !ok {
					return MakeBuiltinError(ps, "Expected Theme", "ui-theme//style")
				}
				switch name := arg1.(type) {
				case env.String:
					switch text := arg2.(type) {
					case env.String:
						return *env.NewString(theme.Apply(name.Value, text.Value))
					default:
						return MakeArgError(ps, 3, []env.Type{env.StringType}, "ui-theme//style")
					}
				default:
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "ui-theme//style")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "ui-theme//style")
			}
		},
	},

	// Returns:
	// * the global default theme
	"ui-global-theme": {
		Argsn: 0,
		Doc:   "Returns the global default theme.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewNative(ps.Idx, GetGlobalTheme(), "ui-theme")
		},
	},

	// Args:
	// * theme: Theme native object
	// Returns:
	// * the theme (now set as global)
	"ui-set-global-theme": {
		Argsn: 1,
		Doc:   "Sets the global default theme.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch native := arg0.(type) {
			case env.Native:
				theme, ok := native.Value.(*Theme)
				if !ok {
					return MakeBuiltinError(ps, "Expected Theme", "ui-set-global-theme")
				}
				SetGlobalTheme(theme)
				return arg0
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "ui-set-global-theme")
			}
		},
	},

	// ##### Component System ##### "Composable UI components"

	// Tests:
	// equal { ui-container |type? } 'native
	// Returns:
	// * a new Container native object
	"ui-container": {
		Argsn: 0,
		Doc:   "Creates a new container for composing UI components vertically.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			container := NewContainer()
			return *env.NewNative(ps.Idx, container, "ui-container")
		},
	},

	// Args:
	// * container: Container native object
	// * child: Component native object
	// Returns:
	// * the container object
	"ui-container//add-child": {
		Argsn: 2,
		Doc:   "Adds a child component to the container.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch native := arg0.(type) {
			case env.Native:
				container, ok := native.Value.(*Container)
				if !ok {
					return MakeBuiltinError(ps, "Expected Container", "ui-container//add-child")
				}
				switch childNative := arg1.(type) {
				case env.Native:
					if child, ok := childNative.Value.(UIComponent); ok {
						container.AddChild(child)
						return arg0
					}
					return MakeBuiltinError(ps, "Expected UIComponent", "ui-container//add-child")
				default:
					return MakeArgError(ps, 2, []env.Type{env.NativeType}, "ui-container//add-child")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "ui-container//add-child")
			}
		},
	},

	// Args:
	// * container: Container native object
	// * child: Component native object
	// Returns:
	// * the container object
	"ui-container//remove-child": {
		Argsn: 2,
		Doc:   "Removes a child component from the container.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch native := arg0.(type) {
			case env.Native:
				container, ok := native.Value.(*Container)
				if !ok {
					return MakeBuiltinError(ps, "Expected Container", "ui-container//remove-child")
				}
				switch childNative := arg1.(type) {
				case env.Native:
					if child, ok := childNative.Value.(UIComponent); ok {
						container.RemoveChild(child)
						return arg0
					}
					return MakeBuiltinError(ps, "Expected UIComponent", "ui-container//remove-child")
				default:
					return MakeArgError(ps, 2, []env.Type{env.NativeType}, "ui-container//remove-child")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "ui-container//remove-child")
			}
		},
	},

	// Args:
	// * container: Container native object
	// Returns:
	// * the container object
	"ui-container//clear": {
		Argsn: 1,
		Doc:   "Removes all children from the container.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch native := arg0.(type) {
			case env.Native:
				container, ok := native.Value.(*Container)
				if !ok {
					return MakeBuiltinError(ps, "Expected Container", "ui-container//clear")
				}
				container.Clear()
				return arg0
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "ui-container//clear")
			}
		},
	},

	// Args:
	// * container: Container native object
	// * width: Integer width
	// Returns:
	// * Block of strings (rendered lines)
	"ui-container//render": {
		Argsn: 2,
		Doc:   "Renders the container and all children to a block of strings.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch native := arg0.(type) {
			case env.Native:
				container, ok := native.Value.(*Container)
				if !ok {
					return MakeBuiltinError(ps, "Expected Container", "ui-container//render")
				}
				switch w := arg1.(type) {
				case env.Integer:
					lines := container.Render(int(w.Value))
					items := make([]env.Object, len(lines))
					for i, line := range lines {
						items[i] = *env.NewString(line)
					}
					return *env.NewBlock(*env.NewTSeries(items))
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "ui-container//render")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "ui-container//render")
			}
		},
	},

	// Args:
	// * container: Container native object
	// Returns:
	// * the container object
	"ui-container//invalidate": {
		Argsn: 1,
		Doc:   "Invalidates the container cache and all children.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch native := arg0.(type) {
			case env.Native:
				container, ok := native.Value.(*Container)
				if !ok {
					return MakeBuiltinError(ps, "Expected Container", "ui-container//invalidate")
				}
				container.Invalidate()
				return arg0
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "ui-container//invalidate")
			}
		},
	},

	// Tests:
	// equal { ui-text "Hello" |type? } 'native
	// Args:
	// * text: String content
	// Returns:
	// * a new Text component
	"ui-text": {
		Argsn: 1,
		Doc:   "Creates a text component with the given content.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s := arg0.(type) {
			case env.String:
				text := NewTextComponent(s.Value)
				return *env.NewNative(ps.Idx, text, "ui-text")
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "ui-text")
			}
		},
	},

	// Args:
	// * text: Text component
	// * content: String new content
	// Returns:
	// * the text component
	"ui-text//set-text": {
		Argsn: 2,
		Doc:   "Updates the text content.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch native := arg0.(type) {
			case env.Native:
				text, ok := native.Value.(*TextComponent)
				if !ok {
					return MakeBuiltinError(ps, "Expected TextComponent", "ui-text//set-text")
				}
				switch s := arg1.(type) {
				case env.String:
					text.SetText(s.Value)
					return arg0
				default:
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "ui-text//set-text")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "ui-text//set-text")
			}
		},
	},

	// Args:
	// * text: Text component
	// * paddingX: Integer horizontal padding
	// * paddingY: Integer vertical padding
	// Returns:
	// * the text component
	"ui-text//set-padding": {
		Argsn: 3,
		Doc:   "Sets the padding around the text.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch native := arg0.(type) {
			case env.Native:
				text, ok := native.Value.(*TextComponent)
				if !ok {
					return MakeBuiltinError(ps, "Expected TextComponent", "ui-text//set-padding")
				}
				switch px := arg1.(type) {
				case env.Integer:
					switch py := arg2.(type) {
					case env.Integer:
						text.SetPadding(int(px.Value), int(py.Value))
						return arg0
					default:
						return MakeArgError(ps, 3, []env.Type{env.IntegerType}, "ui-text//set-padding")
					}
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "ui-text//set-padding")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "ui-text//set-padding")
			}
		},
	},

	// Args:
	// * text: Text component
	// * wrap: Integer 1 to enable wrapping, 0 to disable
	// Returns:
	// * the text component
	"ui-text//set-wrap": {
		Argsn: 2,
		Doc:   "Enables or disables word wrapping.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch native := arg0.(type) {
			case env.Native:
				text, ok := native.Value.(*TextComponent)
				if !ok {
					return MakeBuiltinError(ps, "Expected TextComponent", "ui-text//set-wrap")
				}
				switch w := arg1.(type) {
				case env.Integer:
					text.SetWrap(w.Value != 0)
					return arg0
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "ui-text//set-wrap")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "ui-text//set-wrap")
			}
		},
	},

	// Args:
	// * text: Text component
	// * width: Integer width
	// Returns:
	// * Block of strings
	"ui-text//render": {
		Argsn: 2,
		Doc:   "Renders the text component.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch native := arg0.(type) {
			case env.Native:
				text, ok := native.Value.(*TextComponent)
				if !ok {
					return MakeBuiltinError(ps, "Expected TextComponent", "ui-text//render")
				}
				switch w := arg1.(type) {
				case env.Integer:
					lines := text.Render(int(w.Value))
					items := make([]env.Object, len(lines))
					for i, line := range lines {
						items[i] = *env.NewString(line)
					}
					return *env.NewBlock(*env.NewTSeries(items))
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "ui-text//render")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "ui-text//render")
			}
		},
	},

	// Tests:
	// equal { ui-spacer 2 |type? } 'native
	// Args:
	// * height: Integer number of empty lines
	// Returns:
	// * a new Spacer component
	"ui-spacer": {
		Argsn: 1,
		Doc:   "Creates a spacer component with the given height.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch h := arg0.(type) {
			case env.Integer:
				spacer := NewSpacerComponent(int(h.Value))
				return *env.NewNative(ps.Idx, spacer, "ui-spacer")
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType}, "ui-spacer")
			}
		},
	},

	// Args:
	// * spacer: Spacer component
	// * width: Integer width
	// Returns:
	// * Block of empty strings
	"ui-spacer//render": {
		Argsn: 2,
		Doc:   "Renders the spacer component.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch native := arg0.(type) {
			case env.Native:
				spacer, ok := native.Value.(*SpacerComponent)
				if !ok {
					return MakeBuiltinError(ps, "Expected SpacerComponent", "ui-spacer//render")
				}
				switch w := arg1.(type) {
				case env.Integer:
					lines := spacer.Render(int(w.Value))
					items := make([]env.Object, len(lines))
					for i, line := range lines {
						items[i] = *env.NewString(line)
					}
					return *env.NewBlock(*env.NewTSeries(items))
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "ui-spacer//render")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "ui-spacer//render")
			}
		},
	},

	// Tests:
	// equal { ui-box ui-text "Hi" |type? } 'native
	// Args:
	// * child: Component to wrap in box
	// Returns:
	// * a new Box component
	"ui-box": {
		Argsn: 1,
		Doc:   "Creates a box component that wraps a child with a border.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch native := arg0.(type) {
			case env.Native:
				if child, ok := native.Value.(UIComponent); ok {
					box := NewBoxComponent(child)
					return *env.NewNative(ps.Idx, box, "ui-box")
				}
				return MakeBuiltinError(ps, "Expected UIComponent", "ui-box")
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "ui-box")
			}
		},
	},

	// Args:
	// * box: Box component
	// * title: String title
	// Returns:
	// * the box component
	"ui-box//set-title": {
		Argsn: 2,
		Doc:   "Sets the box title.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch native := arg0.(type) {
			case env.Native:
				box, ok := native.Value.(*BoxComponent)
				if !ok {
					return MakeBuiltinError(ps, "Expected BoxComponent", "ui-box//set-title")
				}
				switch s := arg1.(type) {
				case env.String:
					box.SetTitle(s.Value)
					return arg0
				default:
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "ui-box//set-title")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "ui-box//set-title")
			}
		},
	},

	// Args:
	// * box: Box component
	// * style: Integer 0=single, 1=double, 2=rounded, 3=ascii
	// Returns:
	// * the box component
	"ui-box//set-border-style": {
		Argsn: 2,
		Doc:   "Sets the border style: 0=single, 1=double, 2=rounded, 3=ascii.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch native := arg0.(type) {
			case env.Native:
				box, ok := native.Value.(*BoxComponent)
				if !ok {
					return MakeBuiltinError(ps, "Expected BoxComponent", "ui-box//set-border-style")
				}
				switch s := arg1.(type) {
				case env.Integer:
					box.SetBorderStyle(int(s.Value))
					return arg0
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "ui-box//set-border-style")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "ui-box//set-border-style")
			}
		},
	},

	// Args:
	// * box: Box component
	// * paddingX: Integer horizontal padding
	// * paddingY: Integer vertical padding
	// Returns:
	// * the box component
	"ui-box//set-padding": {
		Argsn: 3,
		Doc:   "Sets internal padding of the box.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch native := arg0.(type) {
			case env.Native:
				box, ok := native.Value.(*BoxComponent)
				if !ok {
					return MakeBuiltinError(ps, "Expected BoxComponent", "ui-box//set-padding")
				}
				switch px := arg1.(type) {
				case env.Integer:
					switch py := arg2.(type) {
					case env.Integer:
						box.SetPadding(int(px.Value), int(py.Value))
						return arg0
					default:
						return MakeArgError(ps, 3, []env.Type{env.IntegerType}, "ui-box//set-padding")
					}
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "ui-box//set-padding")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "ui-box//set-padding")
			}
		},
	},

	// Args:
	// * box: Box component
	// * width: Integer width
	// Returns:
	// * Block of strings
	"ui-box//render": {
		Argsn: 2,
		Doc:   "Renders the box component.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch native := arg0.(type) {
			case env.Native:
				box, ok := native.Value.(*BoxComponent)
				if !ok {
					return MakeBuiltinError(ps, "Expected BoxComponent", "ui-box//render")
				}
				switch w := arg1.(type) {
				case env.Integer:
					lines := box.Render(int(w.Value))
					items := make([]env.Object, len(lines))
					for i, line := range lines {
						items[i] = *env.NewString(line)
					}
					return *env.NewBlock(*env.NewTSeries(items))
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "ui-box//render")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "ui-box//render")
			}
		},
	},

	// Tests:
	// equal { ui-hline |type? } 'native
	// Returns:
	// * a new HorizontalLine component
	"ui-hline": {
		Argsn: 0,
		Doc:   "Creates a horizontal line component.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			hline := NewHorizontalLineComponent("─")
			return *env.NewNative(ps.Idx, hline, "ui-hline")
		},
	},

	// Args:
	// * char: String character to use for the line
	// Returns:
	// * a new HorizontalLine component
	"ui-hline\\char": {
		Argsn: 1,
		Doc:   "Creates a horizontal line component with custom character.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch c := arg0.(type) {
			case env.String:
				hline := NewHorizontalLineComponent(c.Value)
				return *env.NewNative(ps.Idx, hline, "ui-hline")
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "ui-hline\\char")
			}
		},
	},

	// Args:
	// * hline: HorizontalLine component
	// * width: Integer width
	// Returns:
	// * Block of strings (single line)
	"ui-hline//render": {
		Argsn: 2,
		Doc:   "Renders the horizontal line.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch native := arg0.(type) {
			case env.Native:
				hline, ok := native.Value.(*HorizontalLineComponent)
				if !ok {
					return MakeBuiltinError(ps, "Expected HorizontalLineComponent", "ui-hline//render")
				}
				switch w := arg1.(type) {
				case env.Integer:
					lines := hline.Render(int(w.Value))
					items := make([]env.Object, len(lines))
					for i, line := range lines {
						items[i] = *env.NewString(line)
					}
					return *env.NewBlock(*env.NewTSeries(items))
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "ui-hline//render")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "ui-hline//render")
			}
		},
	},

	// Tests:
	// equal { ui-select-list { "a" "b" "c" } 5 |type? } 'native
	// Args:
	// * items: Block of strings
	// * maxVisible: Integer max visible items
	// Returns:
	// * a new SelectList component
	"ui-select-list": {
		Argsn: 2,
		Doc:   "Creates an interactive selection list.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch b := arg0.(type) {
			case env.Block:
				switch mv := arg1.(type) {
				case env.Integer:
					items := make([]string, b.Series.Len())
					for i := 0; i < b.Series.Len(); i++ {
						item := b.Series.Get(i)
						if s, ok := item.(env.String); ok {
							items[i] = s.Value
						} else {
							items[i] = item.Print(*ps.Idx)
						}
					}
					sl := NewSelectListComponent(items, int(mv.Value))
					return *env.NewNative(ps.Idx, sl, "ui-select-list")
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "ui-select-list")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType}, "ui-select-list")
			}
		},
	},

	// Args:
	// * list: SelectList component
	// Returns:
	// * the list component
	"ui-select-list//move-up": {
		Argsn: 1,
		Doc:   "Moves the selection up.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch native := arg0.(type) {
			case env.Native:
				sl, ok := native.Value.(*SelectListComponent)
				if !ok {
					return MakeBuiltinError(ps, "Expected SelectListComponent", "ui-select-list//move-up")
				}
				sl.MoveUp()
				return arg0
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "ui-select-list//move-up")
			}
		},
	},

	// Args:
	// * list: SelectList component
	// Returns:
	// * the list component
	"ui-select-list//move-down": {
		Argsn: 1,
		Doc:   "Moves the selection down.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch native := arg0.(type) {
			case env.Native:
				sl, ok := native.Value.(*SelectListComponent)
				if !ok {
					return MakeBuiltinError(ps, "Expected SelectListComponent", "ui-select-list//move-down")
				}
				sl.MoveDown()
				return arg0
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "ui-select-list//move-down")
			}
		},
	},

	// Args:
	// * list: SelectList component
	// Returns:
	// * Integer selected index
	"ui-select-list//selected": {
		Argsn: 1,
		Doc:   "Returns the selected index.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch native := arg0.(type) {
			case env.Native:
				sl, ok := native.Value.(*SelectListComponent)
				if !ok {
					return MakeBuiltinError(ps, "Expected SelectListComponent", "ui-select-list//selected")
				}
				return *env.NewInteger(int64(sl.Selected()))
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "ui-select-list//selected")
			}
		},
	},

	// Args:
	// * list: SelectList component
	// Returns:
	// * String selected item
	"ui-select-list//selected-item": {
		Argsn: 1,
		Doc:   "Returns the selected item.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch native := arg0.(type) {
			case env.Native:
				sl, ok := native.Value.(*SelectListComponent)
				if !ok {
					return MakeBuiltinError(ps, "Expected SelectListComponent", "ui-select-list//selected-item")
				}
				return *env.NewString(sl.SelectedItem())
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "ui-select-list//selected-item")
			}
		},
	},

	// Args:
	// * list: SelectList component
	// * index: Integer new selected index
	// Returns:
	// * the list component
	"ui-select-list//set-selected": {
		Argsn: 2,
		Doc:   "Sets the selected index.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch native := arg0.(type) {
			case env.Native:
				sl, ok := native.Value.(*SelectListComponent)
				if !ok {
					return MakeBuiltinError(ps, "Expected SelectListComponent", "ui-select-list//set-selected")
				}
				switch idx := arg1.(type) {
				case env.Integer:
					sl.SetSelected(int(idx.Value))
					return arg0
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "ui-select-list//set-selected")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "ui-select-list//set-selected")
			}
		},
	},

	// Args:
	// * list: SelectList component
	// * items: Block of strings
	// Returns:
	// * the list component
	"ui-select-list//set-items": {
		Argsn: 2,
		Doc:   "Updates the list items.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch native := arg0.(type) {
			case env.Native:
				sl, ok := native.Value.(*SelectListComponent)
				if !ok {
					return MakeBuiltinError(ps, "Expected SelectListComponent", "ui-select-list//set-items")
				}
				switch b := arg1.(type) {
				case env.Block:
					items := make([]string, b.Series.Len())
					for i := 0; i < b.Series.Len(); i++ {
						item := b.Series.Get(i)
						if s, ok := item.(env.String); ok {
							items[i] = s.Value
						} else {
							items[i] = item.Print(*ps.Idx)
						}
					}
					sl.SetItems(items)
					return arg0
				default:
					return MakeArgError(ps, 2, []env.Type{env.BlockType}, "ui-select-list//set-items")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "ui-select-list//set-items")
			}
		},
	},

	// Args:
	// * list: SelectList component
	// * width: Integer width
	// Returns:
	// * Block of strings
	"ui-select-list//render": {
		Argsn: 2,
		Doc:   "Renders the selection list.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch native := arg0.(type) {
			case env.Native:
				sl, ok := native.Value.(*SelectListComponent)
				if !ok {
					return MakeBuiltinError(ps, "Expected SelectListComponent", "ui-select-list//render")
				}
				switch w := arg1.(type) {
				case env.Integer:
					lines := sl.Render(int(w.Value))
					items := make([]env.Object, len(lines))
					for i, line := range lines {
						items[i] = *env.NewString(line)
					}
					return *env.NewBlock(*env.NewTSeries(items))
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "ui-select-list//render")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "ui-select-list//render")
			}
		},
	},

	// Args:
	// * list: SelectList component
	// Returns:
	// * the list component
	"ui-select-list//invalidate": {
		Argsn: 1,
		Doc:   "Invalidates the component cache.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch native := arg0.(type) {
			case env.Native:
				sl, ok := native.Value.(*SelectListComponent)
				if !ok {
					return MakeBuiltinError(ps, "Expected SelectListComponent", "ui-select-list//invalidate")
				}
				sl.Invalidate()
				return arg0
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "ui-select-list//invalidate")
			}
		},
	},

	// ##### Input Component ##### "Focusable text input"

	// Tests:
	// equal { ui-input "Name" 20 |type? } 'native
	// Args:
	// * placeholder: String placeholder text
	// * width: Integer display width
	// Returns:
	// * a new Input component
	"ui-input": {
		Argsn: 2,
		Doc:   "Creates a focusable text input component.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch placeholder := arg0.(type) {
			case env.String:
				switch width := arg1.(type) {
				case env.Integer:
					input := NewInputComponent(placeholder.Value, int(width.Value))
					return *env.NewNative(ps.Idx, input, "ui-input")
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "ui-input")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "ui-input")
			}
		},
	},

	// Args:
	// * input: Input component
	// * value: String new value
	// Returns:
	// * the input component
	"ui-input//set-value": {
		Argsn: 2,
		Doc:   "Sets the input value.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch native := arg0.(type) {
			case env.Native:
				input, ok := native.Value.(*InputComponent)
				if !ok {
					return MakeBuiltinError(ps, "Expected InputComponent", "ui-input//set-value")
				}
				switch value := arg1.(type) {
				case env.String:
					input.SetValue(value.Value)
					return arg0
				default:
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "ui-input//set-value")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "ui-input//set-value")
			}
		},
	},

	// Args:
	// * input: Input component
	// Returns:
	// * String current value
	"ui-input//value": {
		Argsn: 1,
		Doc:   "Gets the input value.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch native := arg0.(type) {
			case env.Native:
				input, ok := native.Value.(*InputComponent)
				if !ok {
					return MakeBuiltinError(ps, "Expected InputComponent", "ui-input//value")
				}
				return *env.NewString(input.GetValue())
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "ui-input//value")
			}
		},
	},

	// Args:
	// * input: Input component
	// Returns:
	// * the input component
	"ui-input//focus": {
		Argsn: 1,
		Doc:   "Focuses the input.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch native := arg0.(type) {
			case env.Native:
				input, ok := native.Value.(*InputComponent)
				if !ok {
					return MakeBuiltinError(ps, "Expected InputComponent", "ui-input//focus")
				}
				input.SetFocused(true)
				return arg0
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "ui-input//focus")
			}
		},
	},

	// Args:
	// * input: Input component
	// Returns:
	// * the input component
	"ui-input//blur": {
		Argsn: 1,
		Doc:   "Unfocuses the input.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch native := arg0.(type) {
			case env.Native:
				input, ok := native.Value.(*InputComponent)
				if !ok {
					return MakeBuiltinError(ps, "Expected InputComponent", "ui-input//blur")
				}
				input.SetFocused(false)
				return arg0
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "ui-input//blur")
			}
		},
	},

	// Args:
	// * input: Input component
	// Returns:
	// * Integer 1 if focused, 0 otherwise
	"ui-input//focused?": {
		Argsn: 1,
		Doc:   "Returns whether the input is focused.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch native := arg0.(type) {
			case env.Native:
				input, ok := native.Value.(*InputComponent)
				if !ok {
					return MakeBuiltinError(ps, "Expected InputComponent", "ui-input//focused?")
				}
				return *env.NewInteger(boolToInt64(input.IsFocused()))
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "ui-input//focused?")
			}
		},
	},

	// Args:
	// * input: Input component
	// * key: String key name
	// Returns:
	// * Integer 1 if key was handled, 0 otherwise
	"ui-input//handle-key": {
		Argsn: 2,
		Doc:   "Handles a key press. Returns 1 if handled.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch native := arg0.(type) {
			case env.Native:
				input, ok := native.Value.(*InputComponent)
				if !ok {
					return MakeBuiltinError(ps, "Expected InputComponent", "ui-input//handle-key")
				}
				switch key := arg1.(type) {
				case env.String:
					handled := input.HandleKey(key.Value)
					return *env.NewInteger(boolToInt64(handled))
				default:
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "ui-input//handle-key")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "ui-input//handle-key")
			}
		},
	},

	// Args:
	// * input: Input component
	// * width: Integer width
	// Returns:
	// * Block of strings (single line)
	"ui-input//render": {
		Argsn: 2,
		Doc:   "Renders the input component.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch native := arg0.(type) {
			case env.Native:
				input, ok := native.Value.(*InputComponent)
				if !ok {
					return MakeBuiltinError(ps, "Expected InputComponent", "ui-input//render")
				}
				switch w := arg1.(type) {
				case env.Integer:
					lines := input.Render(int(w.Value))
					items := make([]env.Object, len(lines))
					for i, line := range lines {
						items[i] = *env.NewString(line)
					}
					return *env.NewBlock(*env.NewTSeries(items))
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "ui-input//render")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "ui-input//render")
			}
		},
	},

	// Args:
	// * input: Input component
	// Returns:
	// * the input component
	"ui-input//invalidate": {
		Argsn: 1,
		Doc:   "Invalidates the component cache.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch native := arg0.(type) {
			case env.Native:
				input, ok := native.Value.(*InputComponent)
				if !ok {
					return MakeBuiltinError(ps, "Expected InputComponent", "ui-input//invalidate")
				}
				input.Invalidate()
				return arg0
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "ui-input//invalidate")
			}
		},
	},

	// ##### Text Utilities ##### "ANSI-aware text manipulation"

	// Tests:
	// equal { visible-width "Hello" } 5
	// equal { visible-width "\x1b[31mHello\x1b[0m" } 5
	// Args:
	// * text: String to measure
	// Returns:
	// * Integer visible width (ignoring ANSI codes)
	"visible-width": {
		Argsn: 1,
		Doc:   "Returns the visible display width of a string, ignoring ANSI escape codes.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s := arg0.(type) {
			case env.String:
				return *env.NewInteger(int64(VisibleWidth(s.Value)))
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "visible-width")
			}
		},
	},

	// Tests:
	// equal { truncate-to-width "Hello World" 8 } "Hello..."
	// equal { truncate-to-width "Hello" 10 } "Hello"
	// Args:
	// * text: String to truncate
	// * width: Integer maximum visible width
	// Returns:
	// * String truncated to fit width with ellipsis if needed
	"truncate-to-width": {
		Argsn: 2,
		Doc:   "Truncates a string to fit within the given visible width, adding '...' if truncated. Preserves ANSI codes.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s := arg0.(type) {
			case env.String:
				switch w := arg1.(type) {
				case env.Integer:
					return *env.NewString(TruncateToWidth(s.Value, int(w.Value), "..."))
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "truncate-to-width")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "truncate-to-width")
			}
		},
	},

	// Tests:
	// equal { truncate-to-width\ellipsis "Hello World" 8 ".." } "Hello .."
	// Args:
	// * text: String to truncate
	// * width: Integer maximum visible width
	// * ellipsis: String to append when truncating
	// Returns:
	// * String truncated to fit width with custom ellipsis
	"truncate-to-width\\ellipsis": {
		Argsn: 3,
		Doc:   "Truncates a string to fit within the given visible width with custom ellipsis. Preserves ANSI codes.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s := arg0.(type) {
			case env.String:
				switch w := arg1.(type) {
				case env.Integer:
					switch e := arg2.(type) {
					case env.String:
						return *env.NewString(TruncateToWidth(s.Value, int(w.Value), e.Value))
					default:
						return MakeArgError(ps, 3, []env.Type{env.StringType}, "truncate-to-width\\ellipsis")
					}
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "truncate-to-width\\ellipsis")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "truncate-to-width\\ellipsis")
			}
		},
	},

	// Tests:
	// equal { wrap-text "Hello World" 6 |length? } 2
	// Args:
	// * text: String to wrap
	// * width: Integer maximum width per line
	// Returns:
	// * Block of strings, one per line
	"wrap-text": {
		Argsn: 2,
		Doc:   "Wraps text to fit within the given width, preserving ANSI codes across line breaks.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s := arg0.(type) {
			case env.String:
				switch w := arg1.(type) {
				case env.Integer:
					lines := WrapText(s.Value, int(w.Value))
					items := make([]env.Object, len(lines))
					for i, line := range lines {
						items[i] = *env.NewString(line)
					}
					return *env.NewBlock(*env.NewTSeries(items))
				default:
					return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "wrap-text")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "wrap-text")
			}
		},
	},

	// Tests:
	// equal { terminal-width > 0 } 1
	// Returns:
	// * Integer current terminal width in columns
	"terminal-width": {
		Argsn: 0,
		Doc:   "Returns the current terminal width in columns.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewInteger(int64(term.GetTerminalColumns()))
		},
	},
}

func boolToInt64(b bool) int64 {
	if b {
		return 1
	}
	return 0
}
