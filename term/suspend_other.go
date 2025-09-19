//go:build !unix

package term

import "fmt"

// Global variable to hold the terminal restoration function (no-op on non-Unix systems)
var terminalRestoreFunc func() error

// SetTerminalRestoreFunc allows external packages to register a function
// that will be called to restore terminal state after suspension (no-op on non-Unix systems)
func SetTerminalRestoreFunc(restoreFunc func() error) {
	// No-op on non-Unix systems since they don't support Unix-style process suspension
	terminalRestoreFunc = restoreFunc
}

// SuspendProcess implements fallback process suspension for non-Unix systems
func SuspendProcess() error {
	fmt.Println("[ Process suspension (Ctrl+Z) is not supported on Windows - Use Ctrl+D to exit ]")
	return nil // Return nil to continue running instead of exiting
}
