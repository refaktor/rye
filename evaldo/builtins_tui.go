//go:build !no_tui
// +build !no_tui

package evaldo

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/mattn/go-runewidth"
	"github.com/refaktor/keyboard"
	"github.com/refaktor/rye/env"
	"github.com/refaktor/rye/term"
)

// TUI Library - Minimal declarative terminal UI
// Design principles (from ideas.md):
// - Widget constructors return blocks of values (pure data), not native objects
// - Structure: { 'widget-type { styles... } content }
// - Themes are dicts of named styles
// - Rendering is separate from construction
// - Inline (scroll) mode is the primary mode

// ## Style Helpers

// tuiApplyStyle applies ANSI styles based on a style dict
func tuiApplyStyle(ps *env.ProgramState, styleDict env.Dict, text string) string {
	if len(styleDict.Data) == 0 {
		return text
	}

	var prefix strings.Builder
	needsReset := false

	// Check for various style properties
	if bold, ok := styleDict.Data["bold"]; ok {
		if b, ok := bold.(env.Integer); ok && b.Value != 0 {
			prefix.WriteString("\x1b[1m")
			needsReset = true
		}
	}

	if italic, ok := styleDict.Data["italic"]; ok {
		if b, ok := italic.(env.Integer); ok && b.Value != 0 {
			prefix.WriteString("\x1b[3m")
			needsReset = true
		}
	}

	if underline, ok := styleDict.Data["underline"]; ok {
		if b, ok := underline.(env.Integer); ok && b.Value != 0 {
			prefix.WriteString("\x1b[4m")
			needsReset = true
		}
	}

	if dim, ok := styleDict.Data["dim"]; ok {
		if b, ok := dim.(env.Integer); ok && b.Value != 0 {
			prefix.WriteString("\x1b[2m")
			needsReset = true
		}
	}

	if inverse, ok := styleDict.Data["inverse"]; ok {
		if b, ok := inverse.(env.Integer); ok && b.Value != 0 {
			prefix.WriteString("\x1b[7m")
			needsReset = true
		}
	}

	// Color handling - support both names and ANSI codes
	if color, ok := styleDict.Data["color"]; ok {
		colorCode := tuiColorToAnsi(ps.Idx, color, false)
		if colorCode != "" {
			prefix.WriteString(colorCode)
			needsReset = true
		}
	}

	if bg, ok := styleDict.Data["background"]; ok {
		colorCode := tuiColorToAnsi(ps.Idx, bg, true)
		if colorCode != "" {
			prefix.WriteString(colorCode)
			needsReset = true
		}
	}

	if !needsReset {
		return text
	}

	return prefix.String() + text + "\x1b[0m"
}

// tuiColorToAnsi converts a color value to ANSI escape code
func tuiColorToAnsi(idx *env.Idxs, color any, background bool) string {
	var colorName string

	switch c := color.(type) {
	case env.Word:
		// Look up word index
		colorName = strings.ToLower(idx.GetWord(c.Index))
	case env.Tagword:
		// Look up tagword index
		colorName = strings.ToLower(idx.GetWord(c.Index))
	case env.String:
		// Direct ANSI code or color name
		if strings.HasPrefix(c.Value, "\x1b[") {
			return c.Value
		}
		colorName = strings.ToLower(c.Value)
	default:
		return ""
	}

	// Map color names to ANSI codes
	fgColors := map[string]string{
		"black":   "\x1b[30m",
		"red":     "\x1b[31m",
		"green":   "\x1b[32m",
		"yellow":  "\x1b[33m",
		"blue":    "\x1b[34m",
		"magenta": "\x1b[35m",
		"cyan":    "\x1b[36m",
		"white":   "\x1b[37m",
		"gray":    "\x1b[90m",
		"grey":    "\x1b[90m",
	}

	bgColors := map[string]string{
		"black":   "\x1b[40m",
		"red":     "\x1b[41m",
		"green":   "\x1b[42m",
		"yellow":  "\x1b[43m",
		"blue":    "\x1b[44m",
		"magenta": "\x1b[45m",
		"cyan":    "\x1b[46m",
		"white":   "\x1b[47m",
		"gray":    "\x1b[100m",
		"grey":    "\x1b[100m",
	}

	if background {
		return bgColors[colorName]
	}
	return fgColors[colorName]
}

// tuiGetPadding extracts padding from style dict
func tuiGetPadding(styleDict env.Dict) int {
	if padding, ok := styleDict.Data["padding"]; ok {
		if p, ok := padding.(env.Integer); ok {
			return int(p.Value)
		}
	}
	return 0
}

// ============================================================================
// Widget Block Helpers
// ============================================================================

// tuiMakeWidget creates a widget block: { 'type { styles } content }
func tuiMakeWidget(ps *env.ProgramState, widgetType string, styles env.Dict, content env.Object) env.Block {
	typeWord := env.NewWord(ps.Idx.IndexWord(widgetType))
	items := []env.Object{*typeWord, styles, content}
	return *env.NewBlock(*env.NewTSeries(items))
}

// tuiGetWidgetType extracts widget type from a widget block
func tuiGetWidgetType(ps *env.ProgramState, widget env.Block) (string, bool) {
	if widget.Series.Len() < 1 {
		return "", false
	}
	item := widget.Series.Get(0)
	switch w := item.(type) {
	case env.Word:
		return ps.Idx.GetWord(w.Index), true
	case env.Tagword:
		return ps.Idx.GetWord(w.Index), true
	}
	return "", false
}

// tuiGetWidgetStyles extracts styles from a widget block
func tuiGetWidgetStyles(widget env.Block) (env.Dict, bool) {
	if widget.Series.Len() < 2 {
		return *env.NewDict(nil), false
	}
	item := widget.Series.Get(1)
	if d, ok := item.(env.Dict); ok {
		return d, true
	}
	return *env.NewDict(nil), false
}

// tuiGetWidgetContent extracts content from a widget block
func tuiGetWidgetContent(widget env.Block) (env.Object, bool) {
	if widget.Series.Len() < 3 {
		return nil, false
	}
	return widget.Series.Get(2), true
}

// ## Rendering Engine

// tuiRenderWidget renders a single widget to lines of text
func tuiRenderWidget(ps *env.ProgramState, widget env.Block, width int, theme env.Dict) []string {
	widgetType, ok := tuiGetWidgetType(ps, widget)
	if !ok {
		return []string{}
	}

	styles, _ := tuiGetWidgetStyles(widget)
	content, _ := tuiGetWidgetContent(widget)

	// Merge theme styles with widget styles
	mergedStyles := tuiMergeStyles(theme, widgetType, styles)
	padding := tuiGetPadding(mergedStyles)

	switch widgetType {
	case "text":
		return tuiRenderText(ps, content, width, mergedStyles, false)
	case "block":
		return tuiRenderText(ps, content, width-padding*2, mergedStyles, true)
	case "hline":
		return tuiRenderHLine(ps.Idx, content, width, mergedStyles)
	case "vspace":
		return tuiRenderVSpace(content)
	case "vbox":
		return tuiRenderVBox(ps, content, width, theme, mergedStyles)
	case "select":
		return tuiRenderSelect(ps, content, width, mergedStyles)
	case "tabs":
		return tuiRenderTabs(ps, content, width, mergedStyles)
	case "input":
		return tuiRenderInput(ps, content, width, mergedStyles)
	case "field":
		// Unmanaged field - render using the same logic as input
		return tuiRenderInput(ps, content, width, mergedStyles)
	case "input-field":
		// Managed input widget - content is a string with cursor rendering already done
		if str, ok := content.(env.String); ok {
			return []string{str.Value}
		}
		return []string{""}
	default:
		return []string{fmt.Sprintf("[unknown widget: %s]", widgetType)}
	}
}

// tuiMergeStyles merges theme styles with widget-specific styles
func tuiMergeStyles(theme env.Dict, widgetType string, widgetStyles env.Dict) env.Dict {
	result := env.NewDict(make(map[string]any))

	// First apply theme styles for this widget type
	if themeStyle, ok := theme.Data[widgetType]; ok {
		if ts, ok := themeStyle.(env.Dict); ok {
			for k, v := range ts.Data {
				result.Data[k] = v
			}
		}
	}

	// Then override with widget-specific styles
	for k, v := range widgetStyles.Data {
		result.Data[k] = v
	}

	return *result
}

// tuiRenderText renders text content
// Content can be:
// - String: plain text
// - Block of strings: joined with newlines
// - Block with mixed strings and text widgets: inline rendering with per-element styles
func tuiRenderText(ps *env.ProgramState, content env.Object, width int, styles env.Dict, wrap bool) []string {
	padding := tuiGetPadding(styles)
	paddingStr := strings.Repeat(" ", padding)

	var text string

	switch c := content.(type) {
	case env.String:
		text = c.Value
	case env.Block:
		// Check if this is a mixed content block (strings + inline widgets)
		// or just a block of strings
		hasMixedContent := false
		for i := 0; i < c.Series.Len(); i++ {
			item := c.Series.Get(i)
			if itemBlock, ok := item.(env.Block); ok {
				// Check if it's a text widget
				if wtype, ok := tuiGetWidgetType(ps, itemBlock); ok && wtype == "text" {
					hasMixedContent = true
					break
				}
			}
		}

		if hasMixedContent {
			// Render mixed content inline
			var inlineText strings.Builder
			for i := 0; i < c.Series.Len(); i++ {
				item := c.Series.Get(i)
				switch it := item.(type) {
				case env.String:
					// Apply parent styles to plain strings
					inlineText.WriteString(tuiApplyStyle(ps, styles, it.Value))
				case env.Block:
					// Check if it's a text widget
					if wtype, ok := tuiGetWidgetType(ps, it); ok && wtype == "text" {
						// Get the widget's styles and content
						widgetStyles, _ := tuiGetWidgetStyles(it)
						widgetContent, _ := tuiGetWidgetContent(it)

						// Merge parent styles with widget styles (widget wins)
						mergedStyles := env.NewDict(make(map[string]any))
						for k, v := range styles.Data {
							mergedStyles.Data[k] = v
						}
						for k, v := range widgetStyles.Data {
							mergedStyles.Data[k] = v
						}

						// Get the text content
						var itemText string
						if s, ok := widgetContent.(env.String); ok {
							itemText = s.Value
						} else if widgetContent != nil {
							itemText = widgetContent.Print(*ps.Idx)
						}

						// Apply merged styles
						inlineText.WriteString(tuiApplyStyle(ps, *mergedStyles, itemText))
					} else {
						// Not a text widget, just print it
						inlineText.WriteString(item.Print(*ps.Idx))
					}
				default:
					inlineText.WriteString(item.Print(*ps.Idx))
				}
			}
			text = inlineText.String()

			// For mixed content, wrapping and final styling is different
			// The text already has ANSI codes embedded, so we wrap but don't re-style
			var lines []string
			if wrap && width > 0 {
				lines = WrapText(text, width)
			} else {
				lines = strings.Split(text, "\n")
			}

			var result []string
			for _, line := range lines {
				result = append(result, paddingStr+line)
			}
			return result
		} else {
			// Block of strings - join them with newlines
			var parts []string
			for i := 0; i < c.Series.Len(); i++ {
				item := c.Series.Get(i)
				if s, ok := item.(env.String); ok {
					parts = append(parts, s.Value)
				} else {
					parts = append(parts, item.Print(*ps.Idx))
				}
			}
			text = strings.Join(parts, "\n")
		}
	default:
		text = content.Print(*ps.Idx)
	}

	var lines []string
	if wrap && width > 0 {
		lines = WrapText(text, width)
	} else {
		lines = strings.Split(text, "\n")
	}

	var result []string
	for _, line := range lines {
		styledLine := tuiApplyStyle(ps, styles, line)
		result = append(result, paddingStr+styledLine)
	}

	return result
}

// tuiRenderHLine renders a horizontal line
func tuiRenderHLine(idx *env.Idxs, content env.Object, width int, styles env.Dict) []string {
	char := "─"
	if c, ok := content.(env.String); ok && c.Value != "" {
		char = c.Value
	}

	// Calculate how many times to repeat the character to fill width
	charWidth := runewidth.StringWidth(char)
	if charWidth == 0 {
		charWidth = 1
	}
	repeatCount := width / charWidth
	line := strings.Repeat(char, repeatCount)

	// Apply style
	if styles.Data != nil && len(styles.Data) > 0 {
		line = tuiApplyStyleSimple(idx, styles, line)
	}

	return []string{line}
}

// tuiApplyStyleSimple applies styles with idx for color lookup
func tuiApplyStyleSimple(idx *env.Idxs, styles env.Dict, text string) string {
	var prefix strings.Builder
	needsReset := false

	if bold, ok := styles.Data["bold"]; ok {
		if b, ok := bold.(env.Integer); ok && b.Value != 0 {
			prefix.WriteString("\x1b[1m")
			needsReset = true
		}
	}

	if color, ok := styles.Data["color"]; ok {
		colorCode := tuiColorToAnsi(idx, color, false)
		if colorCode != "" {
			prefix.WriteString(colorCode)
			needsReset = true
		}
	}

	if !needsReset {
		return text
	}

	return prefix.String() + text + "\x1b[0m"
}

// tuiRenderVSpace renders vertical space
func tuiRenderVSpace(content env.Object) []string {
	height := 1
	if h, ok := content.(env.Integer); ok {
		height = int(h.Value)
	}

	result := make([]string, height)
	for i := range result {
		result[i] = ""
	}
	return result
}

// tuiRenderVBox renders a vertical box container with optional styles
func tuiRenderVBox(ps *env.ProgramState, content env.Object, width int, theme env.Dict, styles env.Dict) []string {
	block, ok := content.(env.Block)
	if !ok {
		return []string{}
	}

	padding := tuiGetPadding(styles)
	innerWidth := width - (padding * 2)
	if innerWidth < 1 {
		innerWidth = 1
	}

	// Render children
	var childLines []string
	for i := 0; i < block.Series.Len(); i++ {
		item := block.Series.Get(i)
		if childWidget, ok := item.(env.Block); ok {
			lines := tuiRenderWidget(ps, childWidget, innerWidth, theme)
			childLines = append(childLines, lines...)
		}
	}

	// Check if we have background/style to apply
	hasBackground := false
	var stylePrefix string

	if bg, ok := styles.Data["background"]; ok {
		bgCode := tuiColorToAnsi(ps.Idx, bg, true)
		if bgCode != "" {
			stylePrefix += bgCode
			hasBackground = true
		}
	}

	if color, ok := styles.Data["color"]; ok {
		colorCode := tuiColorToAnsi(ps.Idx, color, false)
		if colorCode != "" {
			stylePrefix += colorCode
			hasBackground = true
		}
	}

	if bold, ok := styles.Data["bold"]; ok {
		if b, ok := bold.(env.Integer); ok && b.Value != 0 {
			stylePrefix += "\x1b[1m"
			hasBackground = true
		}
	}

	// If no background styles, just add padding
	if !hasBackground && padding == 0 {
		return childLines
	}

	paddingStr := strings.Repeat(" ", padding)

	// Apply styles to each line
	var result []string
	for _, line := range childLines {
		if hasBackground {
			// Calculate visible width and pad to fill background
			visWidth := VisibleWidth(line)
			rightPad := innerWidth - visWidth
			if rightPad < 0 {
				rightPad = 0
			}
			fullLine := stylePrefix + paddingStr + line + strings.Repeat(" ", rightPad) + paddingStr + "\x1b[0m"
			result = append(result, fullLine)
		} else {
			// Just add padding
			result = append(result, paddingStr+line)
		}
	}

	return result
}

// tuiRenderSelect renders a selection list
func tuiRenderSelect(ps *env.ProgramState, content env.Object, width int, styles env.Dict) []string {
	block, ok := content.(env.Block)
	if !ok {
		return []string{}
	}

	// Get selected index from styles
	selected := 0
	if s, ok := styles.Data["selected"]; ok {
		if si, ok := s.(env.Integer); ok {
			selected = int(si.Value)
		}
	}

	// Prefix styles
	prefix := "> "
	normalPrefix := "  "
	if p, ok := styles.Data["prefix"]; ok {
		if ps, ok := p.(env.String); ok {
			prefix = ps.Value
		}
	}

	var result []string
	for i := 0; i < block.Series.Len(); i++ {
		item := block.Series.Get(i)
		var text string
		if s, ok := item.(env.String); ok {
			text = s.Value
		} else {
			text = item.Print(*ps.Idx)
		}

		var line string
		if i == selected {
			line = prefix + text
			// Apply selected style
			if selStyle, ok := styles.Data["selected-style"]; ok {
				if ss, ok := selStyle.(env.Dict); ok {
					line = tuiApplyStyleSimple(ps.Idx, ss, line)
				}
			} else {
				// Default: inverse
				line = "\x1b[7m" + line + "\x1b[0m"
			}
		} else {
			line = normalPrefix + text
		}

		result = append(result, TruncateToWidth(line, width, "..."))
	}

	return result
}

// tuiRenderTabs renders horizontal tabs
func tuiRenderTabs(ps *env.ProgramState, content env.Object, width int, styles env.Dict) []string {
	block, ok := content.(env.Block)
	if !ok {
		return []string{}
	}

	// Get selected index from styles
	selected := 0
	if s, ok := styles.Data["selected"]; ok {
		if si, ok := s.(env.Integer); ok {
			selected = int(si.Value)
		}
	}

	separator := " | "
	if sep, ok := styles.Data["separator"]; ok {
		if ss, ok := sep.(env.String); ok {
			separator = ss.Value
		}
	}

	var parts []string
	for i := 0; i < block.Series.Len(); i++ {
		item := block.Series.Get(i)
		var text string
		if s, ok := item.(env.String); ok {
			text = s.Value
		} else {
			text = item.Print(*ps.Idx)
		}

		if i == selected {
			// Apply selected style
			if selStyle, ok := styles.Data["selected-style"]; ok {
				if ss, ok := selStyle.(env.Dict); ok {
					text = tuiApplyStyleSimple(ps.Idx, ss, text)
				}
			} else {
				// Default: bold + underline
				text = "\x1b[1;4m" + text + "\x1b[0m"
			}
		}

		parts = append(parts, text)
	}

	line := strings.Join(parts, separator)
	return []string{TruncateToWidth(line, width, "...")}
}

// tuiRenderInput renders a text input field
func tuiRenderInput(ps *env.ProgramState, content env.Object, width int, styles env.Dict) []string {
	// Content should be a dict with: value, cursor, placeholder, focused
	inputData, ok := content.(env.Dict)
	if !ok {
		return []string{"[input]"}
	}

	value := ""
	if v, ok := inputData.Data["value"]; ok {
		if vs, ok := v.(env.String); ok {
			value = vs.Value
		}
	}

	placeholder := ""
	if p, ok := inputData.Data["placeholder"]; ok {
		if ps, ok := p.(env.String); ok {
			placeholder = ps.Value
		}
	}

	focused := false
	if f, ok := inputData.Data["focused"]; ok {
		if fi, ok := f.(env.Integer); ok {
			focused = fi.Value != 0
		}
	}

	cursor := 0
	if c, ok := inputData.Data["cursor"]; ok {
		if ci, ok := c.(env.Integer); ok {
			cursor = int(ci.Value)
		}
	}

	inputWidth := width
	if w, ok := styles.Data["width"]; ok {
		if wi, ok := w.(env.Integer); ok {
			inputWidth = int(wi.Value)
		}
	}

	var display string
	if value == "" && !focused && placeholder != "" {
		display = "\x1b[90m" + placeholder + "\x1b[0m" // Gray placeholder
	} else if focused {
		// Show cursor
		if cursor < len(value) {
			before := value[:cursor]
			at := string(value[cursor])
			after := value[cursor+1:]
			display = before + "\x1b[7m" + at + "\x1b[0m" + after
		} else {
			display = value + "\x1b[7m \x1b[0m"
		}
	} else {
		display = value
	}

	display = TruncateToWidth(display, inputWidth, "")
	return []string{display}
}

// ## TUI App - Manages rendering and input

// TuiApp represents an inline terminal app
type TuiApp struct {
	theme          env.Dict   // Theme styles
	state          env.Dict   // Current state
	view           env.Object // View block or function
	keyHandlers    map[string]env.Object
	allKeysHandler env.Object // Handler for all keys (set via on-keys)
	running        bool
	stopChan       chan struct{}
	ps             *env.ProgramState
	height         int       // Lines rendered
	prevLines      []string  // Previous output for diff rendering
	focusedInput   *TuiInput // Currently focused input widget
	mu             sync.Mutex
}

// NewTuiApp creates a new TUI app
func NewTuiApp(theme env.Dict) *TuiApp {
	return &TuiApp{
		theme:       theme,
		state:       *env.NewDict(make(map[string]any)),
		keyHandlers: make(map[string]env.Object),
		stopChan:    make(chan struct{}),
	}
}

// SetView sets the view (block or function)
func (app *TuiApp) SetView(view env.Object) {
	app.mu.Lock()
	defer app.mu.Unlock()
	app.view = view
}

// SetState sets the state
func (app *TuiApp) SetState(state env.Dict) {
	app.mu.Lock()
	defer app.mu.Unlock()
	app.state = state
}

// GetState returns the state
func (app *TuiApp) GetState() env.Dict {
	app.mu.Lock()
	defer app.mu.Unlock()
	return app.state
}

// Update merges updates into state and re-renders
func (app *TuiApp) Update(updates env.Dict) {
	app.mu.Lock()
	for k, v := range updates.Data {
		app.state.Data[k] = v
	}
	app.mu.Unlock()
	app.Render()
}

// Render renders the current view
func (app *TuiApp) Render() {
	app.mu.Lock()
	view := app.view
	state := app.state
	theme := app.theme
	prevHeight := app.height
	ps := app.ps
	app.mu.Unlock()

	if view == nil || ps == nil {
		return
	}

	width := term.GetTerminalColumns()
	if width < 20 {
		width = 80
	}

	// Evaluate view if it's a function
	var widgetBlock env.Block
	switch v := view.(type) {
	case env.Block:
		widgetBlock = v
	case env.Function:
		// Call function with state
		psTemp := *ps
		fnCtx := env.NewEnv(psTemp.Ctx)

		// Set first argument (state)
		if v.Spec.Series.Len() > 0 {
			argWord := v.Spec.Series.Get(0)
			if word, ok := argWord.(env.Word); ok {
				fnCtx.Set(word.Index, state)
			}
		}

		psX := env.NewProgramState(v.Body.Series, psTemp.Idx)
		psX.Ctx = fnCtx
		psX.PCtx = psTemp.PCtx
		psX.Gen = psTemp.Gen
		psX.Ser.SetPos(0)
		EvalBlockInj(psX, state, true)

		if psX.ErrorFlag || psX.FailureFlag {
			fmt.Println("View error:", psX.Res.Inspect(*psX.Idx))
			return
		}

		if block, ok := psX.Res.(env.Block); ok {
			widgetBlock = block
		} else {
			return
		}
	default:
		return
	}

	// Render the widget tree
	lines := tuiRenderWidget(ps, widgetBlock, width, theme)

	// Move cursor up if we've rendered before
	if prevHeight > 0 {
		term.CurUp(prevHeight)
	}

	// Print lines with sync for flicker-free rendering
	fmt.Print("\x1b[?2026h") // Sync start
	for _, line := range lines {
		term.ClearLine()
		fmt.Println(line)
	}
	fmt.Print("\x1b[?2026l") // Sync end

	app.mu.Lock()
	app.height = len(lines)
	app.prevLines = lines
	app.mu.Unlock()
}

// OnKey registers a key handler for a specific key
func (app *TuiApp) OnKey(key string, handler env.Object) {
	app.mu.Lock()
	defer app.mu.Unlock()
	app.keyHandlers[key] = handler
}

// OnKeys registers a handler for all keys
func (app *TuiApp) OnKeys(handler env.Object) {
	app.mu.Lock()
	defer app.mu.Unlock()
	app.allKeysHandler = handler
}

// Start starts the app
func (app *TuiApp) Start(ps *env.ProgramState) error {
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
func (app *TuiApp) Stop() {
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
func (app *TuiApp) IsRunning() bool {
	app.mu.Lock()
	defer app.mu.Unlock()
	return app.running
}

// Wait blocks until the app stops
func (app *TuiApp) Wait() {
	for app.IsRunning() {
		time.Sleep(50 * time.Millisecond)
	}
}

// eventLoop handles keyboard events
func (app *TuiApp) eventLoop() {
	for {
		select {
		case <-app.stopChan:
			return
		default:
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
				keyStr = " "
			case keyboard.KeyCtrlC:
				keyStr = "ctrl-c"
				app.Stop()
				return
			case keyboard.KeyCtrlA:
				keyStr = "ctrl-a"
			case keyboard.KeyCtrlE:
				keyStr = "ctrl-e"
			case keyboard.KeyCtrlW:
				keyStr = "ctrl-w"
			case keyboard.KeyCtrlU:
				keyStr = "ctrl-u"
			case keyboard.KeyCtrlK:
				keyStr = "ctrl-k"
			case keyboard.KeyDelete:
				keyStr = "delete"
			case keyboard.KeyHome:
				keyStr = "home"
			case keyboard.KeyEnd:
				keyStr = "end"
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

// handleKey handles a key press
func (app *TuiApp) handleKey(key string) {
	app.mu.Lock()
	focusedInput := app.focusedInput
	ps := app.ps
	app.mu.Unlock()

	// First, route to focused input if any
	if focusedInput != nil && ps != nil {
		handled := focusedInput.HandleKey(key, ps)
		if handled {
			app.Render()
			return
		}
	}

	app.mu.Lock()
	handler, exists := app.keyHandlers[key]
	// If no specific handler, use the all-keys handler
	if !exists && app.allKeysHandler != nil {
		handler = app.allKeysHandler
		exists = true
	}
	app.mu.Unlock()

	if !exists || ps == nil {
		return
	}

	switch h := handler.(type) {
	case env.Function:
		psTemp := *ps
		fnCtx := env.NewEnv(psTemp.Ctx)
		currentState := app.GetState()

		// Set first argument (state)
		if h.Spec.Series.Len() > 0 {
			argWord := h.Spec.Series.Get(0)
			if word, ok := argWord.(env.Word); ok {
				fnCtx.Set(word.Index, currentState)
			}
		}

		// Set second argument (key)
		if h.Spec.Series.Len() > 1 {
			argWord := h.Spec.Series.Get(1)
			if word, ok := argWord.(env.Word); ok {
				fnCtx.Set(word.Index, *env.NewString(key))
			}
		}

		psX := env.NewProgramState(h.Body.Series, psTemp.Idx)
		psX.Ctx = fnCtx
		psX.PCtx = psTemp.PCtx
		psX.Gen = psTemp.Gen
		psX.Ser.SetPos(0)
		EvalBlockInj(psX, currentState, true)

		if psX.ErrorFlag {
			fmt.Println("Key handler error:", psX.Res.Inspect(*psX.Idx))
			return
		}

		// If result is a Dict, update state
		switch d := psX.Res.(type) {
		case env.Dict:
			app.Update(d)
		case *env.Dict:
			app.Update(*d)
		default:
			app.Render()
		}

	case env.Block:
		psTemp := *ps
		psTemp.Ser = h.Series
		psTemp.Ser.Reset()
		EvalBlock(&psTemp)

		if psTemp.ErrorFlag {
			fmt.Println("Block handler error:", psTemp.Res.Inspect(*psTemp.Idx))
			return
		}

		switch d := psTemp.Res.(type) {
		case env.Dict:
			app.Update(d)
		case *env.Dict:
			app.Update(*d)
		default:
			app.Render()
		}
	}
}

// ## TUI Input - Managed text input widget with built-in key handling

// TuiInput represents a text input field with cursor and selection
type TuiInput struct {
	value        []rune     // Current value as runes for proper unicode handling
	cursor       int        // Cursor position (in runes)
	placeholder  string     // Placeholder text
	scrollOffset int        // Horizontal scroll offset for long text
	width        int        // Display width
	focused      bool       // Whether input is focused
	onSubmit     env.Object // Callback when Enter is pressed
	onChange     env.Object // Callback when value changes
	ps           *env.ProgramState
}

// NewTuiInput creates a new input widget
func NewTuiInput(placeholder string) *TuiInput {
	return &TuiInput{
		value:       []rune{},
		cursor:      0,
		placeholder: placeholder,
		width:       40,
		focused:     true,
	}
}

// Type returns the type name for env.Native interface
func (inp *TuiInput) Type() env.Type {
	return env.NativeType
}

// Inspect returns a string representation
func (inp *TuiInput) Inspect(idxs env.Idxs) string {
	return fmt.Sprintf("[TuiInput: \"%s\" cursor:%d]", string(inp.value), inp.cursor)
}

// GetKind returns the kind string
func (inp *TuiInput) GetKind() int {
	return int(env.NativeType)
}

// Equal checks equality
func (inp *TuiInput) Equal(other env.Object) bool {
	if o, ok := other.(*TuiInput); ok {
		return inp == o
	}
	return false
}

// Dump returns a printable representation
func (inp *TuiInput) Dump(e env.Idxs) string {
	return inp.Inspect(e)
}

// Trace returns a trace representation
func (inp *TuiInput) Trace(msg string) {
	fmt.Println(msg)
}

// Print returns printable string
func (inp *TuiInput) Print(e env.Idxs) string {
	return inp.Inspect(e)
}

// SetValue sets the input value
func (inp *TuiInput) SetValue(val string) {
	inp.value = []rune(val)
	if inp.cursor > len(inp.value) {
		inp.cursor = len(inp.value)
	}
}

// GetValue returns the current value
func (inp *TuiInput) GetValue() string {
	return string(inp.value)
}

// HandleKey processes a key event, returns true if handled
func (inp *TuiInput) HandleKey(key string, ps *env.ProgramState) bool {
	inp.ps = ps
	oldValue := string(inp.value)
	handled := true

	switch key {
	// Submit
	case "enter":
		if inp.onSubmit != nil {
			inp.callCallback(inp.onSubmit, string(inp.value))
		}
		return true

	// Navigation
	case "left":
		if inp.cursor > 0 {
			inp.cursor--
		}
	case "right":
		if inp.cursor < len(inp.value) {
			inp.cursor++
		}
	case "ctrl-a", "home":
		inp.cursor = 0
	case "ctrl-e", "end":
		inp.cursor = len(inp.value)
	case "ctrl-left", "alt-left":
		inp.cursor = inp.findWordBoundaryLeft()
	case "ctrl-right", "alt-right":
		inp.cursor = inp.findWordBoundaryRight()

	// Deletion
	case "backspace":
		if inp.cursor > 0 {
			inp.value = append(inp.value[:inp.cursor-1], inp.value[inp.cursor:]...)
			inp.cursor--
		}
	case "delete":
		if inp.cursor < len(inp.value) {
			inp.value = append(inp.value[:inp.cursor], inp.value[inp.cursor+1:]...)
		}
	case "ctrl-w", "alt-backspace":
		// Delete word backwards
		newPos := inp.findWordBoundaryLeft()
		inp.value = append(inp.value[:newPos], inp.value[inp.cursor:]...)
		inp.cursor = newPos
	case "ctrl-u":
		// Delete to start of line
		inp.value = inp.value[inp.cursor:]
		inp.cursor = 0
	case "ctrl-k":
		// Delete to end of line
		inp.value = inp.value[:inp.cursor]

	default:
		// Check if it's a printable character (single rune)
		runes := []rune(key)
		if len(runes) == 1 && runes[0] >= ' ' && runes[0] <= '~' {
			// Insert character at cursor
			newValue := make([]rune, len(inp.value)+1)
			copy(newValue, inp.value[:inp.cursor])
			newValue[inp.cursor] = runes[0]
			copy(newValue[inp.cursor+1:], inp.value[inp.cursor:])
			inp.value = newValue
			inp.cursor++
		} else {
			handled = false
		}
	}

	// Call onChange if value changed
	newValue := string(inp.value)
	if newValue != oldValue && inp.onChange != nil {
		inp.callCallback(inp.onChange, newValue)
	}

	return handled
}

// findWordBoundaryLeft finds the position of the previous word boundary
func (inp *TuiInput) findWordBoundaryLeft() int {
	if inp.cursor == 0 {
		return 0
	}
	pos := inp.cursor - 1
	// Skip any spaces
	for pos > 0 && inp.value[pos] == ' ' {
		pos--
	}
	// Skip word characters
	for pos > 0 && inp.value[pos-1] != ' ' {
		pos--
	}
	return pos
}

// findWordBoundaryRight finds the position of the next word boundary
func (inp *TuiInput) findWordBoundaryRight() int {
	if inp.cursor >= len(inp.value) {
		return len(inp.value)
	}
	pos := inp.cursor
	// Skip word characters
	for pos < len(inp.value) && inp.value[pos] != ' ' {
		pos++
	}
	// Skip spaces
	for pos < len(inp.value) && inp.value[pos] == ' ' {
		pos++
	}
	return pos
}

// callCallback calls a Rye function with a value
func (inp *TuiInput) callCallback(callback env.Object, value string) {
	if inp.ps == nil {
		return
	}

	switch fn := callback.(type) {
	case env.Function:
		psTemp := *inp.ps
		fnCtx := env.NewEnv(psTemp.Ctx)

		// Set first argument (value)
		if fn.Spec.Series.Len() > 0 {
			argWord := fn.Spec.Series.Get(0)
			if word, ok := argWord.(env.Word); ok {
				fnCtx.Set(word.Index, *env.NewString(value))
			}
		}

		psX := env.NewProgramState(fn.Body.Series, psTemp.Idx)
		psX.Ctx = fnCtx
		psX.PCtx = psTemp.PCtx
		EvalBlock(psX)
	}
}

// Render returns the widget block for this input
func (inp *TuiInput) Render() env.Block {
	// Build display string with cursor
	displayWidth := inp.width
	if displayWidth < 10 {
		displayWidth = 40
	}

	var display string
	if len(inp.value) == 0 && !inp.focused {
		display = inp.placeholder
	} else {
		// Calculate scroll offset to keep cursor visible
		if inp.cursor < inp.scrollOffset {
			inp.scrollOffset = inp.cursor
		} else if inp.cursor >= inp.scrollOffset+displayWidth-1 {
			inp.scrollOffset = inp.cursor - displayWidth + 2
		}

		// Get visible portion
		start := inp.scrollOffset
		end := start + displayWidth
		if end > len(inp.value) {
			end = len(inp.value)
		}

		visibleValue := string(inp.value[start:end])

		if inp.focused {
			// Insert cursor character
			cursorPos := inp.cursor - inp.scrollOffset
			if cursorPos < 0 {
				cursorPos = 0
			}
			if cursorPos > len([]rune(visibleValue)) {
				cursorPos = len([]rune(visibleValue))
			}

			runes := []rune(visibleValue)
			if cursorPos < len(runes) {
				// Cursor on a character - show inverted
				before := string(runes[:cursorPos])
				cursorChar := string(runes[cursorPos])
				after := string(runes[cursorPos+1:])
				display = before + "\x1b[7m" + cursorChar + "\x1b[0m" + after
			} else {
				// Cursor at end - show block
				display = visibleValue + "\x1b[7m \x1b[0m"
			}
		} else {
			display = visibleValue
		}
	}

	// Create widget block - just return the text with cursor, styles applied by renderer
	content := *env.NewString(display)
	styles := env.NewDict(map[string]any{})

	// Use the ps.Idx for proper word indexing
	if inp.ps != nil {
		return tuiMakeWidgetWithIdx(inp.ps.Idx, "input-field", *styles, content)
	}
	// Fallback - shouldn't happen in normal use
	return tuiMakeWidgetWithIdx(env.NewIdxs(), "input-field", *styles, content)
}

// tuiMakeWidgetWithIdx creates a widget block using the given Idxs
func tuiMakeWidgetWithIdx(idxs *env.Idxs, widgetType string, styles env.Dict, content env.Object) env.Block {
	typeWord := env.NewWord(idxs.IndexWord(widgetType))
	stylesBlock := env.NewBlock(*env.NewTSeries([]env.Object{styles}))
	return *env.NewBlock(*env.NewTSeries([]env.Object{*typeWord, *stylesBlock, content}))
}

// ## Builtins

var Builtins_tui = map[string]*env.Builtin{

	//
	// ##### TUI library ##### ""
	//
	// text - inline text without wrapping
	// Tests:
	// equal { text "hello" |type? } 'block
	// Args:
	// * content: String text content
	// Returns:
	// * block representing text widget
	"text": {
		Argsn: 1,
		Doc:   "Creates a text widget (no wrapping). Returns { 'text { } content }",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return tuiMakeWidget(ps, "text", *env.NewDict(nil), arg0)
		},
	},

	// text\ - inline text with style
	// Args:
	// * style: Word naming a theme style or Dict of styles
	// * content: String text content
	// Returns:
	// * block representing styled text widget
	"text\\": {
		Argsn: 2,
		Doc:   "Creates a styled text widget. Returns { 'text { styles } content }",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			var styles env.Dict
			switch s := arg0.(type) {
			case env.Dict:
				styles = s
			case env.Word:
				// Style name - create a dict with the name for theme lookup
				styles = *env.NewDict(map[string]any{"_style": ps.Idx.GetWord(s.Index)})
			default:
				styles = *env.NewDict(nil)
			}
			return tuiMakeWidget(ps, "text", styles, arg1)
		},
	},

	// block - text block that wraps
	// Tests:
	// equal { tui-block "hello world" |type? } 'block
	// Args:
	// * content: String or block of strings
	// Returns:
	// * block representing block widget
	"block": {
		Argsn: 1,
		Doc:   "Creates a text block widget that wraps. Returns { 'block { } content }",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return tuiMakeWidget(ps, "block", *env.NewDict(nil), arg0)
		},
	},

	// block\ - text block with style
	// Args:
	// * style: Dict of styles
	// * content: String or block of strings
	// Returns:
	// * block representing styled block widget
	"block\\": {
		Argsn: 2,
		Doc:   "Creates a styled text block widget. Returns { 'block { styles } content }",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			var styles env.Dict
			switch s := arg0.(type) {
			case env.Dict:
				styles = s
			default:
				styles = *env.NewDict(nil)
			}
			return tuiMakeWidget(ps, "block", styles, arg1)
		},
	},

	// hline - horizontal line
	// Tests:
	// equal { tui-hline |type? } 'block
	// Returns:
	// * block representing hline widget
	"hline": {
		Argsn: 0,
		Doc:   "Creates a horizontal line widget. Returns { 'hline { } \"─\" }",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return tuiMakeWidget(ps, "hline", *env.NewDict(nil), *env.NewString("─"))
		},
	},

	// hline\char - horizontal line with custom char
	// Args:
	// * char: String character to use
	// Returns:
	// * block representing hline widget
	"hline\\char": {
		Argsn: 1,
		Doc:   "Creates a horizontal line with custom character. Returns { 'hline { } char }",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return tuiMakeWidget(ps, "hline", *env.NewDict(nil), arg0)
		},
	},

	// vspace - vertical space
	// Tests:
	// equal { tui-vspace 2 |type? } 'block
	// Args:
	// * height: Integer number of empty lines
	// Returns:
	// * block representing vspace widget
	"vspace": {
		Argsn: 1,
		Doc:   "Creates a vertical spacer widget. Returns { 'vspace { } height }",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return tuiMakeWidget(ps, "vspace", *env.NewDict(nil), arg0)
		},
	},

	// vbox - vertical container
	// Tests:
	// equal { tui-vbox { } |type? } 'block
	// Args:
	// * children: Block of widget expressions (will be evaluated)
	// Returns:
	// * block representing vbox widget
	"vbox": {
		Argsn: 1,
		Doc:   "Creates a vertical box container. Evaluates the block and collects widget results. Returns { 'vbox { } children }",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch block := arg0.(type) {
			case env.Block:
				// Evaluate the block to get widget children
				// Each expression in the block should return a widget block
				var children []env.Object

				// Create a temporary program state for evaluation
				psTemp := *ps
				psTemp.Ser = block.Series
				psTemp.Ser.Reset()

				for psTemp.Ser.Pos() < psTemp.Ser.Len() {
					EvalExpression(&psTemp, nil, false, false)
					if psTemp.ErrorFlag || psTemp.FailureFlag {
						return psTemp.Res
					}
					if psTemp.Res != nil {
						// Check if it's a widget block (should have type as first element)
						if _, ok := psTemp.Res.(env.Block); ok {
							children = append(children, psTemp.Res)
						}
					}
				}

				childrenBlock := *env.NewBlock(*env.NewTSeries(children))
				return tuiMakeWidget(ps, "vbox", *env.NewDict(nil), childrenBlock)
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType}, "vbox")
			}
		},
	},

	// box - styled container with background
	// Tests:
	// equal { box dict { } { } |type? } 'block
	// vbox\style - styled vertical box container
	// Args:
	// * styles: Dict with background, color, padding, etc.
	// * children: Block of widget expressions (will be evaluated)
	// Returns:
	// * block representing styled vbox widget
	"vbox\\": {
		Argsn: 2,
		Doc:   "Creates a styled vertical box container. Supports background, color, padding, bold.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			styles, ok := arg0.(env.Dict)
			if !ok {
				return MakeArgError(ps, 1, []env.Type{env.DictType}, "vbox\\")
			}

			switch block := arg1.(type) {
			case env.Block:
				// Evaluate the block to get widget children
				var children []env.Object

				psTemp := *ps
				psTemp.Ser = block.Series
				psTemp.Ser.Reset()

				for psTemp.Ser.Pos() < psTemp.Ser.Len() {
					EvalExpression(&psTemp, nil, false, false)
					if psTemp.ErrorFlag || psTemp.FailureFlag {
						return psTemp.Res
					}
					if psTemp.Res != nil {
						if _, ok := psTemp.Res.(env.Block); ok {
							children = append(children, psTemp.Res)
						}
					}
				}

				childrenBlock := *env.NewBlock(*env.NewTSeries(children))
				return tuiMakeWidget(ps, "vbox", styles, childrenBlock)
			default:
				return MakeArgError(ps, 2, []env.Type{env.BlockType}, "vbox\\")
			}
		},
	},

	// select - vertical selection list
	// Tests:
	// equal { tui-select { "a" "b" "c" } |type? } 'block
	// Args:
	// * items: Block of strings
	// Returns:
	// * block representing select widget
	"select": {
		Argsn: 1,
		Doc:   "Creates a selection list widget. Returns { 'select { selected: 0 } items }",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			styles := env.NewDict(map[string]any{"selected": *env.NewInteger(0)})
			return tuiMakeWidget(ps, "select", *styles, arg0)
		},
	},

	// select\selected - selection list with initial selection
	// Args:
	// * selected: Integer selected index
	// * items: Block of strings
	// Returns:
	// * block representing select widget
	"select\\selected": {
		Argsn: 2,
		Doc:   "Creates a selection list with initial selection. Returns { 'select { selected: n } items }",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			idx, ok := arg0.(env.Integer)
			if !ok {
				return MakeArgError(ps, 1, []env.Type{env.IntegerType}, "select\\selected")
			}
			styles := env.NewDict(map[string]any{"selected": idx})
			return tuiMakeWidget(ps, "select", *styles, arg1)
		},
	},

	// tabs - horizontal tabs
	// Tests:
	// equal { tui-tabs { "Tab1" "Tab2" } |type? } 'block
	// Args:
	// * items: Block of strings
	// Returns:
	// * block representing tabs widget
	"tabs": {
		Argsn: 1,
		Doc:   "Creates a horizontal tabs widget. Returns { 'tabs { selected: 0 } items }",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			styles := env.NewDict(map[string]any{"selected": *env.NewInteger(0)})
			return tuiMakeWidget(ps, "tabs", *styles, arg0)
		},
	},

	// tabs\selected - tabs with initial selection
	// Args:
	// * selected: Integer selected index
	// * items: Block of strings
	// Returns:
	// * block representing tabs widget
	"tabs\\selected": {
		Argsn: 2,
		Doc:   "Creates tabs with initial selection. Returns { 'tabs { selected: n } items }",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			idx, ok := arg0.(env.Integer)
			if !ok {
				return MakeArgError(ps, 1, []env.Type{env.IntegerType}, "tabs\\selected")
			}
			styles := env.NewDict(map[string]any{"selected": idx})
			return tuiMakeWidget(ps, "tabs", *styles, arg1)
		},
	},

	// input - managed text input field with built-in key handling
	// Args:
	// * placeholder: String placeholder text
	// Returns:
	// * Native TuiInput object
	"input": {
		Argsn: 1,
		Doc:   "Creates a managed text input widget with built-in key handling. Returns native TuiInput.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			placeholder, ok := arg0.(env.String)
			if !ok {
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "input")
			}
			inp := NewTuiInput(placeholder.Value)
			inp.ps = ps
			return *env.NewNative(ps.Idx, inp, "tui-input")
		},
	},

	// field - unmanaged text field (pure data block, manual key handling)
	// Args:
	// * placeholder: String placeholder text
	// Returns:
	// * block representing field widget { 'field { } { value: "" placeholder: ... } }
	"field": {
		Argsn: 1,
		Doc:   "Creates an unmanaged text field widget (pure data). Handle keys manually. Returns { 'field { } { value: '' placeholder: ... } }",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			placeholder, ok := arg0.(env.String)
			if !ok {
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "field")
			}
			content := env.NewDict(map[string]any{
				"value":       *env.NewString(""),
				"placeholder": placeholder,
				"cursor":      *env.NewInteger(0),
			})
			return tuiMakeWidget(ps, "field", *env.NewDict(nil), *content)
		},
	},

	// field\value - create field with initial value
	"field\\value": {
		Argsn: 2,
		Doc:   "Creates an unmanaged text field with initial value.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			value, ok := arg0.(env.String)
			if !ok {
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "field\\value")
			}
			placeholder, ok := arg1.(env.String)
			if !ok {
				return MakeArgError(ps, 2, []env.Type{env.StringType}, "field\\value")
			}
			content := env.NewDict(map[string]any{
				"value":       value,
				"placeholder": placeholder,
				"cursor":      *env.NewInteger(int64(len(value.Value))),
			})
			return tuiMakeWidget(ps, "field", *env.NewDict(nil), *content)
		},
	},

	// input//On-submit - set submit callback
	"tui-input//On-submit": {
		Argsn: 2,
		Doc:   "Sets the callback function called when Enter is pressed. Callback receives (value).",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch inp := arg0.(type) {
			case env.Native:
				if tinp, ok := inp.Value.(*TuiInput); ok {
					tinp.onSubmit = arg1
					tinp.ps = ps
					return arg0
				}
				return MakeBuiltinError(ps, "Expected TuiInput", "tui-input//On-submit")
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "tui-input//On-submit")
			}
		},
	},

	// input//On-change - set change callback
	"tui-input//On-change": {
		Argsn: 2,
		Doc:   "Sets the callback function called when value changes. Callback receives (value).",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch inp := arg0.(type) {
			case env.Native:
				if tinp, ok := inp.Value.(*TuiInput); ok {
					tinp.onChange = arg1
					tinp.ps = ps
					return arg0
				}
				return MakeBuiltinError(ps, "Expected TuiInput", "tui-input//On-change")
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "tui-input//On-change")
			}
		},
	},

	// input//Value? - get current value
	"tui-input//Value?": {
		Argsn: 1,
		Doc:   "Returns the current value of the input.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch inp := arg0.(type) {
			case env.Native:
				if tinp, ok := inp.Value.(*TuiInput); ok {
					return *env.NewString(tinp.GetValue())
				}
				return MakeBuiltinError(ps, "Expected TuiInput", "tui-input//Value?")
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "tui-input//Value?")
			}
		},
	},

	// input//Set-value - set value
	"tui-input//Set-value": {
		Argsn: 2,
		Doc:   "Sets the value of the input.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch inp := arg0.(type) {
			case env.Native:
				if tinp, ok := inp.Value.(*TuiInput); ok {
					if val, ok := arg1.(env.String); ok {
						tinp.SetValue(val.Value)
						return arg0
					}
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "tui-input//Set-value")
				}
				return MakeBuiltinError(ps, "Expected TuiInput", "tui-input//Set-value")
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "tui-input//Set-value")
			}
		},
	},

	// input//Focus - set focus state
	"tui-input//Focus": {
		Argsn: 2,
		Doc:   "Sets the focus state of the input (1 = focused, 0 = not focused).",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch inp := arg0.(type) {
			case env.Native:
				if tinp, ok := inp.Value.(*TuiInput); ok {
					if val, ok := arg1.(env.Integer); ok {
						tinp.focused = val.Value != 0
						return arg0
					}
					return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "tui-input//Focus")
				}
				return MakeBuiltinError(ps, "Expected TuiInput", "tui-input//Focus")
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "tui-input//Focus")
			}
		},
	},

	// input//Width - set display width
	"tui-input//Width": {
		Argsn: 2,
		Doc:   "Sets the display width of the input.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch inp := arg0.(type) {
			case env.Native:
				if tinp, ok := inp.Value.(*TuiInput); ok {
					if val, ok := arg1.(env.Integer); ok {
						tinp.width = int(val.Value)
						return arg0
					}
					return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "tui-input//Width")
				}
				return MakeBuiltinError(ps, "Expected TuiInput", "tui-input//Width")
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "tui-input//Width")
			}
		},
	},

	// input//Handle-key - process a key event
	"tui-input//Handle-key": {
		Argsn: 2,
		Doc:   "Processes a key event. Returns 1 if handled, 0 if not.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch inp := arg0.(type) {
			case env.Native:
				if tinp, ok := inp.Value.(*TuiInput); ok {
					if key, ok := arg1.(env.String); ok {
						handled := tinp.HandleKey(key.Value, ps)
						if handled {
							return *env.NewInteger(1)
						}
						return *env.NewInteger(0)
					}
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "tui-input//Handle-key")
				}
				return MakeBuiltinError(ps, "Expected TuiInput", "tui-input//Handle-key")
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "tui-input//Handle-key")
			}
		},
	},

	// input//Widget - get the widget block for rendering
	"tui-input//Widget": {
		Argsn: 1,
		Doc:   "Returns the widget block for rendering this input.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch inp := arg0.(type) {
			case env.Native:
				if tinp, ok := inp.Value.(*TuiInput); ok {
					tinp.ps = ps // Ensure ps is set for proper widget creation
					return tinp.Render()
				}
				return MakeBuiltinError(ps, "Expected TuiInput", "tui-input//Widget")
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "tui-input//Widget")
			}
		},
	},

	// ## Rendering
	// tui-render - render a widget tree to the terminal
	// Args:
	// * widget: Block widget tree
	// Returns:
	// * Integer number of lines rendered
	"render": {
		Argsn: 1,
		Doc:   "Renders a widget tree to the terminal and returns number of lines.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			widget, ok := arg0.(env.Block)
			if !ok {
				return MakeArgError(ps, 1, []env.Type{env.BlockType}, "render")
			}

			width := term.GetTerminalColumns()
			if width < 20 {
				width = 80
			}

			theme := *env.NewDict(nil)
			lines := tuiRenderWidget(ps, widget, width, theme)

			for _, line := range lines {
				fmt.Println(line)
			}

			return *env.NewInteger(int64(len(lines)))
		},
	},

	// tui-render\theme - render with theme
	// Args:
	// * theme: Dict of theme styles
	// * widget: Block widget tree
	// Returns:
	// * Integer number of lines rendered
	"render\\theme": {
		Argsn: 2,
		Doc:   "Renders a widget tree with a theme.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			theme, ok := arg0.(env.Dict)
			if !ok {
				return MakeArgError(ps, 1, []env.Type{env.DictType}, "render\\theme")
			}
			widget, ok := arg1.(env.Block)
			if !ok {
				return MakeArgError(ps, 2, []env.Type{env.BlockType}, "render\\theme")
			}

			width := term.GetTerminalColumns()
			if width < 20 {
				width = 80
			}

			lines := tuiRenderWidget(ps, widget, width, theme)

			for _, line := range lines {
				fmt.Println(line)
			}

			return *env.NewInteger(int64(len(lines)))
		},
	},

	// tui-render\to-string - render to string (not terminal)
	// Args:
	// * widget: Block widget tree
	// * width: Integer width
	// Returns:
	// * String rendered output
	"render\\to-string": {
		Argsn: 2,
		Doc:   "Renders a widget tree to a string.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			widget, ok := arg0.(env.Block)
			if !ok {
				return MakeArgError(ps, 1, []env.Type{env.BlockType}, "render\\to-string")
			}
			width, ok := arg1.(env.Integer)
			if !ok {
				return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "render\\to-string")
			}

			theme := *env.NewDict(nil)
			lines := tuiRenderWidget(ps, widget, int(width.Value), theme)

			return *env.NewString(strings.Join(lines, "\n"))
		},
	},

	// ## TUI App

	// tui-app - create a TUI app
	// Tests:
	// equal { tui-app dict { } |type? } 'native
	// Args:
	// * theme: Dict of theme styles
	// Returns:
	// * TuiApp native object
	"app": {
		Argsn: 1,
		Doc:   "Creates a new TUI app with the given theme.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			theme, ok := arg0.(env.Dict)
			if !ok {
				return MakeArgError(ps, 1, []env.Type{env.DictType}, "app")
			}
			app := NewTuiApp(theme)
			return *env.NewNative(ps.Idx, app, "tui-app")
		},
	},

	// tui-app//View - set the view
	// Args:
	// * app: TuiApp native
	// * view: Block widget or Function returning widget
	// Returns:
	// * the app
	"tui-app//View": {
		Argsn: 2,
		Doc:   "Sets the view for the TUI app. Can be a widget block or a function that returns one.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			native, ok := arg0.(env.Native)
			if !ok {
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "tui-app//View")
			}
			app, ok := native.Value.(*TuiApp)
			if !ok {
				return MakeBuiltinError(ps, "Expected TuiApp", "tui-app//View")
			}
			app.SetView(arg1)
			return arg0
		},
	},

	// tui-app//State - set the state
	// Args:
	// * app: TuiApp native
	// * state: Dict state
	// Returns:
	// * the app
	"tui-app//State": {
		Argsn: 2,
		Doc:   "Sets the state for the TUI app.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			native, ok := arg0.(env.Native)
			if !ok {
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "tui-app//State")
			}
			app, ok := native.Value.(*TuiApp)
			if !ok {
				return MakeBuiltinError(ps, "Expected TuiApp", "tui-app//State")
			}
			state, ok := arg1.(env.Dict)
			if !ok {
				return MakeArgError(ps, 2, []env.Type{env.DictType}, "tui-app//State")
			}
			app.SetState(state)
			return arg0
		},
	},

	// tui-app//State? - get the state
	// Args:
	// * app: TuiApp native
	// Returns:
	// * Dict current state
	"tui-app//State?": {
		Argsn: 1,
		Doc:   "Gets the current state of the TUI app.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			native, ok := arg0.(env.Native)
			if !ok {
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "tui-app//State?")
			}
			app, ok := native.Value.(*TuiApp)
			if !ok {
				return MakeBuiltinError(ps, "Expected TuiApp", "tui-app//State?")
			}
			return app.GetState()
		},
	},

	// tui-app//Update - update state and re-render
	// Args:
	// * app: TuiApp native
	// * updates: Dict state updates
	// Returns:
	// * the app
	"tui-app//Update": {
		Argsn: 2,
		Doc:   "Updates the state and re-renders.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			native, ok := arg0.(env.Native)
			if !ok {
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "tui-app//Update")
			}
			app, ok := native.Value.(*TuiApp)
			if !ok {
				return MakeBuiltinError(ps, "Expected TuiApp", "tui-app//Update")
			}
			updates, ok := arg1.(env.Dict)
			if !ok {
				return MakeArgError(ps, 2, []env.Type{env.DictType}, "tui-app//Update")
			}
			app.Update(updates)
			return arg0
		},
	},

	// tui-app//On-key - register key handler
	// Args:
	// * app: TuiApp native
	// * key: String key name (or "*" for default)
	// * handler: Function or Block
	// Returns:
	// * the app
	"tui-app//On-key": {
		Argsn: 3,
		Doc:   "Registers a handler for a specific key. Key can be 'up', 'down', 'enter', 'escape', 'q', etc.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			native, ok := arg0.(env.Native)
			if !ok {
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "tui-app//On-key")
			}
			app, ok := native.Value.(*TuiApp)
			if !ok {
				return MakeBuiltinError(ps, "Expected TuiApp", "tui-app//On-key")
			}
			key, ok := arg1.(env.String)
			if !ok {
				return MakeArgError(ps, 2, []env.Type{env.StringType}, "tui-app//On-key")
			}
			switch arg2.(type) {
			case env.Function, env.Block:
				app.OnKey(key.Value, arg2)
				return arg0
			default:
				return MakeArgError(ps, 3, []env.Type{env.FunctionType, env.BlockType}, "tui-app//On-key")
			}
		},
	},

	// tui-app//On-keys - register a handler for all keys
	// Args:
	// * app: TuiApp native
	// * handler: Function that takes (state, key)
	// Returns:
	// * the app
	"tui-app//On-keys": {
		Argsn: 2,
		Doc:   "Registers a handler for ALL keys. Function receives (state, key). Use switch on key to handle different keys.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			native, ok := arg0.(env.Native)
			if !ok {
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "tui-app//On-keys")
			}
			app, ok := native.Value.(*TuiApp)
			if !ok {
				return MakeBuiltinError(ps, "Expected TuiApp", "tui-app//On-keys")
			}
			switch arg1.(type) {
			case env.Function, env.Block:
				app.OnKeys(arg1)
				return arg0
			default:
				return MakeArgError(ps, 2, []env.Type{env.FunctionType, env.BlockType}, "tui-app//On-keys")
			}
		},
	},

	// app//Focus - set focused input widget
	// Args:
	// * app: TuiApp native
	// * input: TuiInput native (or 0 to clear focus)
	// Returns:
	// * the app
	"tui-app//Focus": {
		Argsn: 2,
		Doc:   "Sets the focused input widget. Keys are routed to focused input first.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			native, ok := arg0.(env.Native)
			if !ok {
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "tui-app//Focus")
			}
			app, ok := native.Value.(*TuiApp)
			if !ok {
				return MakeBuiltinError(ps, "Expected TuiApp", "tui-app//Focus")
			}

			// Allow clearing focus with 0 or none
			if i, ok := arg1.(env.Integer); ok && i.Value == 0 {
				app.mu.Lock()
				app.focusedInput = nil
				app.mu.Unlock()
				return arg0
			}

			inputNative, ok := arg1.(env.Native)
			if !ok {
				return MakeArgError(ps, 2, []env.Type{env.NativeType}, "tui-app//Focus")
			}
			inp, ok := inputNative.Value.(*TuiInput)
			if !ok {
				return MakeBuiltinError(ps, "Expected TuiInput", "tui-app//Focus")
			}

			app.mu.Lock()
			app.focusedInput = inp
			inp.focused = true
			inp.ps = ps
			app.mu.Unlock()
			return arg0
		},
	},

	// tui-app//Start - start the app
	// Args:
	// * app: TuiApp native
	// Returns:
	// * the app
	"tui-app//Start": {
		Argsn: 1,
		Doc:   "Starts the TUI app, entering raw mode and beginning the event loop.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			native, ok := arg0.(env.Native)
			if !ok {
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "tui-app//Start")
			}
			app, ok := native.Value.(*TuiApp)
			if !ok {
				return MakeBuiltinError(ps, "Expected TuiApp", "tui-app//Start")
			}
			err := app.Start(ps)
			if err != nil {
				return MakeBuiltinError(ps, err.Error(), "tui-app//Start")
			}
			return arg0
		},
	},

	// tui-app//Stop - stop the app
	// Args:
	// * app: TuiApp native
	// Returns:
	// * the app
	"tui-app//Stop": {
		Argsn: 1,
		Doc:   "Stops the TUI app.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			native, ok := arg0.(env.Native)
			if !ok {
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "tui-app//Stop")
			}
			app, ok := native.Value.(*TuiApp)
			if !ok {
				return MakeBuiltinError(ps, "Expected TuiApp", "tui-app//Stop")
			}
			app.Stop()
			return arg0
		},
	},

	// tui-app//Wait - wait for app to stop
	// Args:
	// * app: TuiApp native
	// Returns:
	// * the app
	"tui-app//Wait": {
		Argsn: 1,
		Doc:   "Blocks until the TUI app stops.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			native, ok := arg0.(env.Native)
			if !ok {
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "tui-app//Wait")
			}
			app, ok := native.Value.(*TuiApp)
			if !ok {
				return MakeBuiltinError(ps, "Expected TuiApp", "tui-app//Wait")
			}
			app.Wait()
			return arg0
		},
	},

	// tui-app//Render - force a render
	// Args:
	// * app: TuiApp native
	// Returns:
	// * the app
	"tui-app//Redraw": {
		Argsn: 1,
		Doc:   "Forces a re-render of the TUI app.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			native, ok := arg0.(env.Native)
			if !ok {
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "tui-app//Redraw")
			}
			app, ok := native.Value.(*TuiApp)
			if !ok {
				return MakeBuiltinError(ps, "Expected TuiApp", "tui-app//Redraw")
			}
			// Ensure ps is set for rendering
			app.mu.Lock()
			if app.ps == nil {
				app.ps = ps
			}
			app.mu.Unlock()
			app.Render()
			return arg0
		},
	},

	// ## Helpers

	// tui-style - create a style dict
	// Args:
	// * props: Block of key value pairs
	// Returns:
	// * Dict style
	"style": {
		Argsn: 1,
		Doc:   "Creates a style dict from a block. Example: tui-style { bold: 1 color: 'blue }",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch b := arg0.(type) {
			case env.Block:
				// Convert block to dict
				data := make(map[string]any)
				for i := 0; i < b.Series.Len(); i += 2 {
					if i+1 >= b.Series.Len() {
						break
					}
					key := b.Series.Get(i)
					val := b.Series.Get(i + 1)

					var keyStr string
					switch k := key.(type) {
					case env.Tagword:
						keyStr = ps.Idx.GetWord(k.Index)
					case env.Word:
						keyStr = ps.Idx.GetWord(k.Index)
					case env.String:
						keyStr = k.Value
					default:
						continue
					}

					data[keyStr] = val
				}
				return *env.NewDict(data)
			case env.Dict:
				return b
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType, env.DictType}, "style")
			}
		},
	},

	// tui-theme - create a theme dict
	// Args:
	// * defs: Block of widget-type style pairs
	// Returns:
	// * Dict theme
	"theme": {
		Argsn: 1,
		Doc:   "Creates a theme dict. Example: tui-theme { text { color: 'blue } hline { color: 'gray } }",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch b := arg0.(type) {
			case env.Block:
				data := make(map[string]any)
				for i := 0; i < b.Series.Len(); i += 2 {
					if i+1 >= b.Series.Len() {
						break
					}
					key := b.Series.Get(i)
					val := b.Series.Get(i + 1)

					var keyStr string
					switch k := key.(type) {
					case env.Tagword:
						keyStr = ps.Idx.GetWord(k.Index)
					case env.Word:
						keyStr = ps.Idx.GetWord(k.Index)
					case env.String:
						keyStr = k.Value
					default:
						continue
					}

					// Value should be a style block
					switch v := val.(type) {
					case env.Block:
						// Convert to dict
						styleData := make(map[string]any)
						for j := 0; j < v.Series.Len(); j += 2 {
							if j+1 >= v.Series.Len() {
								break
							}
							sk := v.Series.Get(j)
							sv := v.Series.Get(j + 1)

							var skStr string
							switch skk := sk.(type) {
							case env.Tagword:
								skStr = ps.Idx.GetWord(skk.Index)
							case env.Word:
								skStr = ps.Idx.GetWord(skk.Index)
							case env.String:
								skStr = skk.Value
							default:
								continue
							}
							styleData[skStr] = sv
						}
						data[keyStr] = *env.NewDict(styleData)
					case env.Dict:
						data[keyStr] = v
					}
				}
				return *env.NewDict(data)
			case env.Dict:
				return b
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType, env.DictType}, "theme")
			}
		},
	},
}
