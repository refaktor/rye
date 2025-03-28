//go:build wasm
// +build wasm

package term

import (
	"fmt"
	"sync"
)

// KeyEvent is defined in microliner.go
// Using the KeyEvent struct from microliner.go to avoid duplicate declaration

var sendBack func(string)
var keyEventChan chan KeyEvent
var keyEventMutex sync.Mutex

func SetSB(fn func(string)) {
	sendBack = fn

	// Override the terminal output functions to use sendBack
	termPrint = func(s string) {
		sendBack(s)
	}
	termPrintln = func(s string) {
		sendBack(s + "\n")
	}
	termPrintf = func(format string, args ...interface{}) {
		sendBack(fmt.Sprintf(format, args...))
	}
}

// GetChar reads a character from the browser terminal and returns ASCII code, key code, and error
func GetChar() (ascii int, keyCode int, err error) {
	// Use NonBlockingGetChar to avoid deadlock in DisplayBlock
	return NonBlockingGetChar()
}

// NonBlockingGetChar reads a character from the browser terminal without blocking
// Returns default values if no key event is available
func NonBlockingGetChar() (ascii int, keyCode int, err error) {
	// Try to read from channel without blocking
	select {
	case event := <-keyEventChan:
		// Process the event normally
		ascii = 0
		keyCode = event.Code

		// Handle special keys
		if event.Key == "Enter" {
			ascii = 13
		} else if event.Key == "Escape" {
			ascii = 27
		} else if event.Key == "m" || event.Key == "M" {
			ascii = int(event.Key[0])
		} else if len(event.Key) == 1 {
			ascii = int(event.Key[0])
		}

		// Handle arrow keys
		if event.Key == "ArrowUp" {
			keyCode = 38
		} else if event.Key == "ArrowDown" {
			keyCode = 40
		} else if event.Key == "ArrowLeft" {
			keyCode = 37
		} else if event.Key == "ArrowRight" {
			keyCode = 39
		}

		return ascii, keyCode, nil
	default:
		// Return default values if no event is available
		return 27, 0, nil // Return ESC key to exit DisplayBlock
	}
}

// GetChar2 is similar to GetChar but returns the character as a string
func GetChar2() (letter string, ascii int, keyCode int, err error) {
	a, k, e := GetChar()
	if e != nil {
		return "", 0, 0, e
	}

	if a != 0 {
		letter = string(rune(a))
	}

	return letter, a, k, nil
}

// InitKeyEventChannel initializes the key event channel
func InitKeyEventChannel(ch chan KeyEvent) {
	keyEventMutex.Lock()
	defer keyEventMutex.Unlock()
	keyEventChan = ch
}

// Terminal size functions are defined in getcolumns_wasm.go

// Note: Display functions and terminal output functions are implemented in term.go
// The termPrint, termPrintln, and termPrintf functions are overridden in SetSB
