//go:build wasm
// +build wasm

package term

import (
	"fmt"
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
// In WASM, we return default values since we don't have direct terminal access
func GetTerminalSize() (TerminalSize, error) {
	// In WASM environment, return reasonable defaults
	// These could be made configurable via external variables
	return TerminalSize{Width: 80, Height: 24}, nil
}

// GetTerminalHeight returns just the terminal height in rows
func GetTerminalHeight() int {
	size, _ := GetTerminalSize()
	return size.Height
}

// QueryCursorPosition queries the terminal for the current cursor position using ANSI escape sequences
// Returns the current row and column (1-based indexing as returned by terminal)
// Note: In WASM, this is not supported
func QueryCursorPosition() (CursorPosition, error) {
	// WASM doesn't have direct terminal access
	return CursorPosition{}, fmt.Errorf("cursor position query not supported in WASM environment")
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
