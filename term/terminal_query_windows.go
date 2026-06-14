//go:build windows
// +build windows

package term

import (
	"fmt"
	"os"

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
// Note: On Windows, this may not work in all terminal environments
func QueryCursorPosition() (CursorPosition, error) {
	// Windows implementation would need platform-specific code
	// For now, return an error indicating it's not supported
	return CursorPosition{}, fmt.Errorf("cursor position query not implemented for Windows")
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
