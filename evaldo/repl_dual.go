//go:build !b_norepl && !wasm && !js

package evaldo

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	tcell "github.com/gdamore/tcell/v2"
	"github.com/refaktor/rye/env"
	"github.com/refaktor/rye/term"
)

// DualReplState represents the state of the dual REPL
type DualReplState struct {
	leftRepl     *Repl
	rightRepl    *Repl
	screen       tcell.Screen
	leftPos      int
	rightPos     int
	active       int // 0 for left, 1 for right
	quit         bool
	mutex        sync.Mutex
	leftBuffer   string      // Input buffer for left REPL
	rightBuffer  string      // Input buffer for right REPL
	leftCursor   int         // Cursor position for left REPL
	rightCursor  int         // Cursor position for right REPL
	leftInputCh  chan string // Channel for sending input to left REPL
	rightInputCh chan string // Channel for sending input to right REPL
	leftHistory  []string    // Command history for left REPL
	rightHistory []string    // Command history for right REPL
	leftOutput   string      // Current output from left REPL
	rightOutput  string      // Current output from right REPL
	leftLog      []string    // Log of inputs and outputs for left REPL
	rightLog     []string    // Log of inputs and outputs for right REPL
}

// Custom message handler for the dual REPL that writes to the tcell screen
func (d *DualReplState) leftReceiveMessage(message string) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	log.Println("Left Message: " + message)

	// Check for special control sequences
	if strings.HasPrefix(message, "\r") {
		// Carriage return - move cursor to start of line
		d.leftCursor = 0
		return
	}

	// Handle ANSI escape sequences
	if strings.HasPrefix(message, "\033") || strings.HasPrefix(message, "\x1b") {
		// Check for clear line sequences
		if strings.Contains(message, "\033[K") || strings.Contains(message, "\x1b[K") ||
			strings.Contains(message, "\033[2K") || strings.Contains(message, "\x1b[2K") {
			// Clear line
			d.leftBuffer = ""
			d.leftCursor = 0
			// Also clear the output buffer to prevent text accumulation
			d.leftOutput = ""
			return
		}

		// Check for cursor movement
		if strings.HasPrefix(message, "\033[") && strings.Contains(message, "C") {
			// Move cursor right
			parts := strings.Split(message, "\033[")
			for _, part := range parts {
				if strings.HasSuffix(part, "C") {
					numStr := strings.TrimSuffix(part, "C")
					if num, err := strconv.Atoi(numStr); err == nil {
						d.leftCursor += num
						if d.leftCursor > len(d.leftBuffer) {
							d.leftCursor = len(d.leftBuffer)
						}
					}
				}
			}
			return
		}

		// For other ANSI sequences, store them in the output buffer
		d.leftOutput += message
		return
	}

	// Regular text - insert at cursor position
	if d.leftCursor == len(d.leftBuffer) {
		d.leftBuffer += message
	} else {
		d.leftBuffer = d.leftBuffer[:d.leftCursor] + message + d.leftBuffer[d.leftCursor:]
	}
	d.leftCursor += len(message)
}

func (d *DualReplState) rightReceiveMessage(message string) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	log.Println("Right Message: " + message)

	// Check for special control sequences
	if strings.HasPrefix(message, "\r") {
		// Carriage return - move cursor to start of line
		d.rightCursor = 0
		return
	}

	// Handle ANSI escape sequences
	if strings.HasPrefix(message, "\033") || strings.HasPrefix(message, "\x1b") {
		// Check for clear line sequences
		if strings.Contains(message, "\033[K") || strings.Contains(message, "\x1b[K") ||
			strings.Contains(message, "\033[2K") || strings.Contains(message, "\x1b[2K") {
			// Clear line
			d.rightBuffer = ""
			d.rightCursor = 0
			// Also clear the output buffer to prevent text accumulation
			d.rightOutput = ""
			return
		}

		// Check for cursor movement
		if strings.HasPrefix(message, "\033[") && strings.Contains(message, "C") {
			// Move cursor right
			parts := strings.Split(message, "\033[")
			for _, part := range parts {
				if strings.HasSuffix(part, "C") {
					numStr := strings.TrimSuffix(part, "C")
					if num, err := strconv.Atoi(numStr); err == nil {
						d.rightCursor += num
						if d.rightCursor > len(d.rightBuffer) {
							d.rightCursor = len(d.rightBuffer)
						}
					}
				}
			}
			return
		}

		// For other ANSI sequences, store them in the output buffer
		d.rightOutput += message
		return
	}

	// Regular text - insert at cursor position
	if d.rightCursor == len(d.rightBuffer) {
		d.rightBuffer += message
	} else {
		d.rightBuffer = d.rightBuffer[:d.rightCursor] + message + d.rightBuffer[d.rightCursor:]
	}
	d.rightCursor += len(message)
}

// NewDualReplState creates a new dual REPL state
func NewDualReplState(leftPs, rightPs *env.ProgramState, dialect string, showResults bool, leftChan chan term.KeyEvent, rightChan chan term.KeyEvent) (*DualReplState, error) {
	// Initialize tcell screen
	screen, err := tcell.NewScreen()
	if err != nil {
		return nil, fmt.Errorf("error creating screen: %v", err)
	}

	if err := screen.Init(); err != nil {
		return nil, fmt.Errorf("error initializing screen: %v", err)
	}

	log.Print("New Dual REPL State")

	// Create the dual state first so we can use its methods as callbacks
	dualState := &DualReplState{
		screen:       screen,
		active:       0, // Start with left panel active
		leftBuffer:   "",
		rightBuffer:  "",
		leftCursor:   0,
		rightCursor:  0,
		leftHistory:  make([]string, 0),
		rightHistory: make([]string, 0),
		leftOutput:   "Left REPL ready",
		rightOutput:  "Right REPL ready",
		leftLog:      make([]string, 0),
		rightLog:     make([]string, 0),
	}

	leftRepl := &Repl{
		ps:            leftPs,
		dialect:       dialect,
		showResults:   false,
		stack:         env.NewEyrStack(),
		captureStdout: true,
	}
	// Create a callback function to handle the output from receiveLine
	leftReceiveLineCallback := func(line string) string {
		output := leftRepl.recieveLine(line)
		// Add the output to the log
		if output != "" {
			dualState.mutex.Lock()
			dualState.leftLog = append(dualState.leftLog, output)
			dualState.leftOutput = output
			dualState.mutex.Unlock()
		}
		return output
	}
	mlL := term.NewMicroLiner(leftChan, dualState.leftReceiveMessage, leftReceiveLineCallback)
	leftRepl.ml = mlL

	rightRepl := &Repl{
		ps:            rightPs,
		dialect:       dialect,
		showResults:   false,
		stack:         env.NewEyrStack(),
		captureStdout: true,
	}
	// Create a callback function to handle the output from receiveLine
	rightReceiveLineCallback := func(line string) string {
		output := rightRepl.recieveLine(line)
		// Add the output to the log
		if output != "" {
			dualState.mutex.Lock()
			dualState.rightLog = append(dualState.rightLog, output)
			dualState.rightOutput = output
			dualState.mutex.Unlock()
		}
		return output
	}
	mlR := term.NewMicroLiner(rightChan, dualState.rightReceiveMessage, rightReceiveLineCallback)
	rightRepl.ml = mlR

	dualState.leftRepl = leftRepl
	dualState.rightRepl = rightRepl

	// Create input channels
	leftInputCh := make(chan string, 10)
	rightInputCh := make(chan string, 10)

	dualState.leftInputCh = leftInputCh
	dualState.rightInputCh = rightInputCh

	go func() {
		// Create context with timeout to prevent potential deadlocks
		ctx, cancel := context.WithTimeout(context.Background(), 24*time.Hour)
		defer cancel()
		// Run the REPL with improved error handling
		_, err = mlL.MicroPrompt("x> ", "", 0, ctx)
		if err != nil {
			log.Printf("MicroPrompt error: %v", err)
		}
	}()

	go func() {
		// Create context with timeout to prevent potential deadlocks
		ctx, cancel := context.WithTimeout(context.Background(), 24*time.Hour)
		defer cancel()
		// Run the REPL with improved error handling
		_, err = mlR.MicroPrompt("x> ", "", 0, ctx)
		if err != nil {
			log.Printf("MicroPrompt error: %v", err)
		}
	}()

	return dualState, nil
}

// Draw draws the dual REPL UI
func (d *DualReplState) Draw() {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	// Clear the screen
	d.screen.Clear()

	// Get screen dimensions
	width, height := d.screen.Size()
	halfWidth := width / 2

	// Draw divider line
	for y := 0; y < height; y++ {
		d.screen.SetContent(halfWidth, y, '│', nil, tcell.StyleDefault)
	}

	// Draw header
	drawText(d.screen, 2, 0, halfWidth-4, "Left REPL", tcell.StyleDefault.Foreground(tcell.ColorWhite))
	drawText(d.screen, halfWidth+2, 0, halfWidth-4, "Right REPL", tcell.StyleDefault.Foreground(tcell.ColorWhite))

	// Highlight active panel
	if d.active == 0 {
		drawText(d.screen, 0, 0, 2, "▶ ", tcell.StyleDefault.Foreground(tcell.ColorGreen))
	} else {
		drawText(d.screen, halfWidth, 0, 2, "▶ ", tcell.StyleDefault.Foreground(tcell.ColorGreen))
	}

	// Draw REPL history and output area
	outputStyle := tcell.StyleDefault.Foreground(tcell.ColorYellow)
	inputStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite)
	promptStyle := tcell.StyleDefault.Foreground(tcell.ColorBlue)

	// Calculate available space for history
	historyHeight := height - 5 // Reserve space for header, input line, and status

	// Draw left panel history and current output
	if len(d.leftLog) > 0 {
		// Display as many log entries as will fit, starting from the most recent
		startIdx := len(d.leftLog)
		if startIdx > historyHeight {
			startIdx = len(d.leftLog) - historyHeight
		} else {
			startIdx = 0
		}

		// Draw the log entries
		for i := startIdx; i < len(d.leftLog); i++ {
			y := 2 + (i - startIdx)
			entry := d.leftLog[i]

			// Determine style based on whether it's an input or output
			style := outputStyle
			if strings.HasPrefix(entry, "×>") {
				style = promptStyle
			}

			drawText(d.screen, 2, y, halfWidth-4, entry, style)
		}
	} else if d.leftOutput != "" {
		// If no log entries yet, just show the current output
		drawText(d.screen, 2, height-5, halfWidth-4, d.leftOutput, outputStyle)
	}

	// Draw right panel history and current output
	if len(d.rightLog) > 0 {
		// Display as many log entries as will fit, starting from the most recent
		startIdx := len(d.rightLog)
		if startIdx > historyHeight {
			startIdx = len(d.rightLog) - historyHeight
		} else {
			startIdx = 0
		}

		// Draw the log entries
		for i := startIdx; i < len(d.rightLog); i++ {
			y := 2 + (i - startIdx)
			entry := d.rightLog[i]

			// Determine style based on whether it's an input or output
			style := outputStyle
			if strings.HasPrefix(entry, "×>") {
				style = promptStyle
			}

			drawText(d.screen, halfWidth+2, y, halfWidth-4, entry, style)
		}
	} else if d.rightOutput != "" {
		// If no log entries yet, just show the current output
		drawText(d.screen, halfWidth+2, height-5, halfWidth-4, d.rightOutput, outputStyle)
	}

	// Draw input prompts
	drawText(d.screen, 2, height-3, 2, "×> ", promptStyle)
	drawText(d.screen, halfWidth+2, height-3, 2, "×> ", promptStyle)

	// Draw input buffers
	drawText(d.screen, 5, height-3, halfWidth-7, d.leftBuffer, inputStyle)
	drawText(d.screen, halfWidth+5, height-3, halfWidth-7, d.rightBuffer, inputStyle)

	// Draw cursors
	if d.active == 0 {
		// Position cursor for left panel
		d.screen.ShowCursor(5+d.leftCursor, height-3)
	} else {
		// Position cursor for right panel
		d.screen.ShowCursor(halfWidth+5+d.rightCursor, height-3)
	}

	// Draw status line at the bottom
	statusText := "Tab: Switch panels | Enter: Execute | Ctrl+C: Quit"
	drawText(d.screen, 0, height-1, width, statusText, tcell.StyleDefault.Foreground(tcell.ColorYellow))

	// Show the screen
	d.screen.Show()
}

// tcellEventToTermKeyEvent converts a tcell.EventKey to a term.KeyEvent
func tcellEventToTermKeyEvent(ev *tcell.EventKey) term.KeyEvent {
	// Map tcell key codes to term.KeyEvent fields
	var key string
	var code int
	var ctrl, alt, shift bool

	// Set modifiers
	mods := ev.Modifiers()
	ctrl = mods&tcell.ModCtrl != 0
	alt = mods&tcell.ModAlt != 0
	shift = mods&tcell.ModShift != 0

	// Handle key mapping
	switch ev.Key() {
	case tcell.KeyRune:
		key = string(ev.Rune())
		code = int(ev.Rune())
	case tcell.KeyEnter:
		code = 13
	case tcell.KeyTab:
		code = 9
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		code = 8
		key = "backspace"
	case tcell.KeyDelete:
		code = 46
	case tcell.KeyRight:
		code = 39
	case tcell.KeyLeft:
		code = 37
	case tcell.KeyUp:
		code = 38
	case tcell.KeyDown:
		code = 40
	case tcell.KeyHome:
		code = 36
	case tcell.KeyEnd:
		code = 35
	case tcell.KeyEscape:
		code = 27
	case tcell.KeyCtrlA:
		key = "a"
		ctrl = true
	case tcell.KeyCtrlB:
		key = "b"
		ctrl = true
	case tcell.KeyCtrlC:
		key = "c"
		ctrl = true
	case tcell.KeyCtrlD:
		key = "d"
		ctrl = true
	case tcell.KeyCtrlE:
		key = "e"
		ctrl = true
	case tcell.KeyCtrlF:
		key = "f"
		ctrl = true
	case tcell.KeyCtrlK:
		key = "k"
		ctrl = true
	case tcell.KeyCtrlL:
		key = "l"
		ctrl = true
	case tcell.KeyCtrlN:
		key = "n"
		ctrl = true
	case tcell.KeyCtrlP:
		key = "p"
		ctrl = true
	case tcell.KeyCtrlU:
		key = "u"
		ctrl = true
	}

	return term.NewKeyEvent(key, code, ctrl, alt, shift)
}

// HandleInput handles keyboard input for the dual REPL
func (d *DualReplState) HandleInput(ev *tcell.EventKey) bool {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	log.Println("HANDLE INPUT 1")

	// Handle global keys
	switch ev.Key() {
	case tcell.KeyTab:
		// Switch active panel
		d.active = 1 - d.active
		return true
	case tcell.KeyCtrlC:
		// Quit
		d.quit = true
		return false
	}

	log.Println("EV->KEYEVENT")
	log.Println(ev)
	// Convert tcell event to term.KeyEvent
	keyEvent := tcellEventToTermKeyEvent(ev)

	log.Println(keyEvent)

	// Send the key event to the appropriate microliner
	if d.active == 0 {
		// Send to left REPL
		select {
		case leftChan <- keyEvent:
			// Successfully sent
		default:
			// Channel full, could log an error here
		}
	} else {
		// Send to right REPL
		select {
		case rightChan <- keyEvent:
			// Successfully sent
		default:
			// Channel full, could log an error here
		}
	}

	return true
}

// Global channels for key events
var leftChan chan term.KeyEvent
var rightChan chan term.KeyEvent

// RunDualRepl runs the dual REPL
func RunDualRepl(leftPs, rightPs *env.ProgramState, dialect string, showResults bool) {

	// Create left and right REPLs
	leftChan = make(chan term.KeyEvent)
	rightChan = make(chan term.KeyEvent)

	// Create dual REPL state
	dualState, err := NewDualReplState(leftPs, rightPs, dialect, showResults, leftChan, rightChan)
	if err != nil {
		log.Fatalf("Failed to create dual REPL: %v", err)
	}

	// Ensure screen is closed when function exits
	defer dualState.screen.Fini()

	// Create channels for communication between goroutines
	leftResultCh := make(chan string)
	rightResultCh := make(chan string)
	quitCh := make(chan struct{})

	// Create context with cancel for goroutines
	//ctx, cancel := context.WithCancel(context.Background())
	// defer cancel()

	// Start goroutines for each REPL
	// go runReplInGoroutine(ctx, dualState.leftRepl, dualState.leftInputCh, leftResultCh, "Left")
	// go runReplInGoroutine(ctx, dualState.rightRepl, dualState.rightInputCh, rightResultCh, "Right")

	// Draw initial UI
	dualState.Draw()

	// Main event loop
	go func() {
		// Set up a ticker to redraw the screen periodically
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case result := <-leftResultCh:
				dualState.mutex.Lock()
				// Add the result to the log
				if len(dualState.leftBuffer) > 0 {
					// Add the input to the log first
					inputEntry := fmt.Sprintf("×> %s", dualState.leftBuffer)
					dualState.leftLog = append(dualState.leftLog, inputEntry)
					// Then add the result
					dualState.leftLog = append(dualState.leftLog, result)
					// Clear the input buffer
					dualState.leftBuffer = ""
					dualState.leftCursor = 0
				}
				dualState.leftOutput = result
				dualState.mutex.Unlock()
				dualState.Draw()
			case result := <-rightResultCh:
				dualState.mutex.Lock()
				// Add the result to the log
				if len(dualState.rightBuffer) > 0 {
					// Add the input to the log first
					inputEntry := fmt.Sprintf("×> %s", dualState.rightBuffer)
					dualState.rightLog = append(dualState.rightLog, inputEntry)
					// Then add the result
					dualState.rightLog = append(dualState.rightLog, result)
					// Clear the input buffer
					dualState.rightBuffer = ""
					dualState.rightCursor = 0
				}
				dualState.rightOutput = result
				dualState.mutex.Unlock()
				dualState.Draw()
			case <-ticker.C:
				// Redraw the screen periodically to ensure updates are visible
				dualState.Draw()
			default:
				// Poll for UI events
				ev := dualState.screen.PollEvent()
				switch ev := ev.(type) {
				case *tcell.EventKey:
					if !dualState.HandleInput(ev) {
						close(quitCh)
						return
					}
					// Force a redraw after each key press
					dualState.Draw()
				case *tcell.EventResize:
					dualState.screen.Sync()
					dualState.Draw()
				}
			}
		}
	}()

	// Wait for quit signal
	<-quitCh
}

// runReplInGoroutine runs a REPL in a goroutine
func runReplInGoroutine(ctx context.Context, repl *Repl, inputCh chan string, resultCh chan string, name string) {
	// Set up a ticker to periodically check for input
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case input := <-inputCh:
			log.Println("INPUT CH")
			// Process input from UI
			if len(input) > 0 {
				// STDIO CAPTURE START
				// Define variables outside the if block
				var r1, w *os.File
				var err error
				var oldStdout *os.File

				// Create a pipe to capture stdout
				r1, w, err = os.Pipe()
				if err != nil {
					log.Printf("Failed to create pipe: %v", err)
					resultCh <- "Error: Failed to capture output"
					continue
				}

				// Save the original stdout
				oldStdout = os.Stdout
				// Replace stdout with our pipe writer
				os.Stdout = w

				// Create a channel for the captured output
				stdoutCh := make(chan string)

				// Start a goroutine to read from the pipe
				go func() {
					var buf bytes.Buffer
					_, err := io.Copy(&buf, r1)
					if err != nil {
						log.Printf("Error reading from pipe: %v", err)
					}
					stdoutCh <- buf.String()
				}()

				// Call evalLine which will now write to our pipe
				result := repl.evalLine(repl.ps, input)

				// Close the pipe writer to signal EOF to the reader
				w.Close()

				// Restore the original stdout
				os.Stdout = oldStdout

				// Get the captured stdout
				capturedOutput := <-stdoutCh

				log.Println("CAPTURED STDOUT")
				log.Println(capturedOutput)

				// Close the pipe reader
				r1.Close()

				// Send both the result and captured stdout to the result channel
				if capturedOutput != "" {
					// If we have captured output, send it along with the result
					resultCh <- capturedOutput + result
				} else {
					// If no captured output, just send the result
					resultCh <- result
				}
			}
		case <-ticker.C:
			// Periodic check (could be used for other tasks)
		}
	}
}

// Helper function to draw text on the screen
func drawText(s tcell.Screen, x, y, maxWidth int, text string, style tcell.Style) {
	// Parse ANSI escape sequences and convert to tcell styles
	currentStyle := style
	i := 0 // Position in the output

	for j := 0; j < len(text); {
		if j+1 < len(text) && (text[j] == '\033' || text[j] == '\x1b') {
			// Found an escape sequence
			if j+2 < len(text) && text[j+1] == '[' {
				// ANSI escape sequence
				end := j + 2
				for end < len(text) && ((text[end] >= '0' && text[end] <= '9') || text[end] == ';' || text[end] == '[' || text[end] == '?') {
					end++
				}

				if end < len(text) {
					// Process the escape sequence
					escapeCode := text[j+2 : end]
					command := text[end]

					switch command {
					case 'm': // SGR (Select Graphic Rendition)
						// Handle color and style codes
						codes := strings.Split(escapeCode, ";")
						for _, code := range codes {
							if code == "" {
								continue
							}

							codeNum, err := strconv.Atoi(code)
							if err != nil {
								continue
							}

							switch {
							case codeNum == 0: // Reset
								currentStyle = style
							case codeNum == 1: // Bold/Bright
								currentStyle = currentStyle.Bold(true)
							case codeNum == 2: // Dim
								currentStyle = currentStyle.Dim(true)
							case codeNum >= 30 && codeNum <= 37: // Foreground color
								color := tcell.Color(codeNum - 30)
								currentStyle = currentStyle.Foreground(color)
							case codeNum >= 40 && codeNum <= 47: // Background color
								color := tcell.Color(codeNum - 40)
								currentStyle = currentStyle.Background(color)
							}
						}
					}

					// Skip the escape sequence
					j = end + 1
					continue
				}
			}
		}

		// Regular character
		if i < maxWidth {
			r, size := utf8.DecodeRuneInString(text[j:])
			s.SetContent(x+i, y, r, nil, currentStyle)
			i++
			j += size
		} else {
			break
		}
	}
}

// DoRyeDualRepl is the entry point for the dual REPL mode
func DoRyeDualRepl(leftPs, rightPs *env.ProgramState, dialect string, showResults bool) {
	RunDualRepl(leftPs, rightPs, dialect, showResults)
}
