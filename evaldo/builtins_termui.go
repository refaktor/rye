//go:build !no_termui
// +build !no_termui

package evaldo

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/refaktor/keyboard"
	"github.com/refaktor/rye/env"
	"github.com/refaktor/rye/term"
)

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

	// If already rendered, move cursor up to re-render in place
	if rendered {
		term.CurUp(height)
	}

	// Capture the render output
	var output string

	switch fn := renderFn.(type) {
	case env.Function:
		if app.ps != nil {
			// Call Rye render function with state
			// Create a temporary program state for the call
			psTemp := *app.ps
			psTemp.Res = nil

			// Create a new context for the function call
			// Use the program state's context as parent to access script-defined functions
			fnCtx := env.NewEnv(psTemp.Ctx)

			// Set the first argument (state)
			if fn.Spec.Series.Len() > 0 {
				argWord := fn.Spec.Series.Get(0)
				if word, ok := argWord.(env.Word); ok {
					fnCtx.Set(word.Index, state)
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
			}
			if psX.FailureFlag {
				fmt.Println("Render failure:", psX.Res.Inspect(*psX.Idx))
			}

			psTemp.Res = psX.Res
		}
	case *env.Native:
		// Check if it's a GoRenderFn
		if goFn, ok := fn.Value.(GoRenderFn); ok {
			output = goFn(state, 80, height) // TODO: get actual terminal width
			lines := strings.Split(output, "\n")
			for i := 0; i < height; i++ {
				term.ClearLine()
				if i < len(lines) {
					fmt.Println(lines[i])
				} else {
					fmt.Println()
				}
			}
		}
	}

	// If using Rye function that prints directly, ensure we have the right number of lines
	if _, ok := renderFn.(env.Function); ok {
		// The function handles its own printing
		// We assume it prints exactly `height` lines
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
	// equal { app: inline-app 1 , app .render fn { s } { print "test" } |type? } 'native
	// Args:
	// * app: InlineApp native object
	// * render-fn: Function that takes state Dict and renders output
	// Returns:
	// * the app object
	"inline-app//render": {
		Argsn: 2,
		Doc:   "Sets the render function for an inline app. The function receives the current state.",
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
}

func boolToInt64(b bool) int64 {
	if b {
		return 1
	}
	return 0
}
