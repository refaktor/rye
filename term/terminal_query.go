//go:build !windows && !wasm
// +build !windows,!wasm

package term

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/term"
	goterm "golang.org/x/term"
)

// CursorPosition represents the current cursor position
type CursorPosition struct {
	Row int
	Col int
}

// TerminalSize represents the terminal dimensions
type TerminalSize struct {
	Width  int
	Height int
}

// GetTerminalSize returns the current terminal size (width and height in characters)
func GetTerminalSize() (TerminalSize, error) {
	width, height, err := goterm.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return TerminalSize{Width: 80, Height: 24}, err
	}
	return TerminalSize{Width: width, Height: height}, nil
}

// GetTerminalHeight returns just the terminal height in rows
func GetTerminalHeight() int {
	size, err := GetTerminalSize()
	if err != nil {
		return 24 // fallback
	}
	return size.Height
}

// QueryCursorPosition queries the terminal for the current cursor position using ANSI escape sequences
// Returns the current row and column (1-based indexing as returned by terminal)
func QueryCursorPosition() (CursorPosition, error) {
	t, err := term.Open("/dev/tty")
	if err != nil {
		return CursorPosition{}, fmt.Errorf("failed to open terminal: %w", err)
	}
	defer t.Close()

	// Set raw mode to read the response
	if err = term.RawMode(t); err != nil {
		return CursorPosition{}, fmt.Errorf("failed to set raw mode: %w", err)
	}
	defer t.Restore()

	// Send cursor position query (CPR - Cursor Position Report)
	_, err = t.Write([]byte("\033[6n"))
	if err != nil {
		return CursorPosition{}, fmt.Errorf("failed to write query: %w", err)
	}

	// Read the response with timeout
	response := make([]byte, 32)

	// Set a read timeout
	done := make(chan bool, 1)
	var numRead int
	var readErr error

	go func() {
		numRead, readErr = t.Read(response)
		done <- true
	}()

	select {
	case <-done:
		if readErr != nil {
			return CursorPosition{}, fmt.Errorf("failed to read response: %w", readErr)
		}
	case <-time.After(100 * time.Millisecond):
		return CursorPosition{}, fmt.Errorf("timeout waiting for cursor position response")
	}

	// Parse the response: ESC[{row};{col}R
	responseStr := string(response[:numRead])
	if !strings.HasPrefix(responseStr, "\033[") || !strings.HasSuffix(responseStr, "R") {
		return CursorPosition{}, fmt.Errorf("invalid response format: %s", responseStr)
	}

	// Extract the row;col part
	coords := responseStr[2 : len(responseStr)-1] // Remove ESC[ and R
	parts := strings.Split(coords, ";")
	if len(parts) != 2 {
		return CursorPosition{}, fmt.Errorf("invalid coordinate format: %s", coords)
	}

	row, err := strconv.Atoi(parts[0])
	if err != nil {
		return CursorPosition{}, fmt.Errorf("invalid row number: %s", parts[0])
	}

	col, err := strconv.Atoi(parts[1])
	if err != nil {
		return CursorPosition{}, fmt.Errorf("invalid column number: %s", parts[1])
	}

	return CursorPosition{Row: row, Col: col}, nil
}

// GetCurrentRow returns just the current row position
func GetCurrentRow() (int, error) {
	pos, err := QueryCursorPosition()
	if err != nil {
		return 0, err
	}
	return pos.Row, nil
}

// GetCurrentColumn returns just the current column position
func GetCurrentColumn() (int, error) {
	pos, err := QueryCursorPosition()
	if err != nil {
		return 0, err
	}
	return pos.Col, nil
}

// IsNearBottomOfTerminal checks if cursor is near the bottom of the terminal
// Returns true if within 'threshold' lines of the bottom
func IsNearBottomOfTerminal(threshold int) (bool, error) {
	pos, err := QueryCursorPosition()
	if err != nil {
		return false, err
	}

	size, err := GetTerminalSize()
	if err != nil {
		return false, err
	}

	return (size.Height - pos.Row) <= threshold, nil
}
