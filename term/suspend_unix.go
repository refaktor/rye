//go:build unix

package term

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

// Global variable to hold the terminal restoration function
var terminalRestoreFunc func() error

// SetTerminalRestoreFunc allows external packages to register a function
// that will be called to restore terminal state after suspension
func SetTerminalRestoreFunc(restoreFunc func() error) {
	terminalRestoreFunc = restoreFunc
}

// SuspendProcess implements Unix-style process suspension using SIGTSTP signal
// It properly handles terminal state restoration after resume to fix input issues
func SuspendProcess() error {
	fmt.Println("[ Process suspended with Ctrl+Z - Use '%NUMBER' to resume ]")

	// Create a channel to handle signals
	sigCh := make(chan os.Signal, 1)

	// Setup signal handler for SIGCONT (resume)
	signal.Notify(sigCh, syscall.SIGCONT)

	// Send SIGTSTP to current process (suspend)
	pid := os.Getpid()
	if err := syscall.Kill(pid, syscall.SIGTSTP); err != nil {
		signal.Stop(sigCh)
		close(sigCh)
		return fmt.Errorf("failed to suspend process: %w", err)
	}

	// Wait for SIGCONT (resume signal)
	<-sigCh

	// Clean up signal handler
	signal.Stop(sigCh)
	close(sigCh)

	// After resume, restore terminal state
	if err := restoreTerminalAfterResume(); err != nil {
		fmt.Printf("[ Process resumed - Warning: failed to restore terminal: %v ]\n", err)
	} else {
		fmt.Println("[ Process resumed - Terminal restored ]")
	}

	return nil
}

// restoreTerminalAfterResume attempts to restore terminal state after suspension
func restoreTerminalAfterResume() error {
	// First, try to use the registered restore function if available
	if terminalRestoreFunc != nil {
		if err := terminalRestoreFunc(); err != nil {
			// If the registered function fails, fall back to basic terminal reset
			basicTerminalReset()
			return fmt.Errorf("terminal restore function failed: %w", err)
		}
		return nil
	}

	// Fallback to basic terminal reset if no restore function is registered
	basicTerminalReset()
	return nil
}

// basicTerminalReset performs basic terminal state restoration
func basicTerminalReset() {
	// Send terminal reset sequences to restore proper functioning
	fmt.Print("\033c")     // Full terminal reset
	fmt.Print("\033[0m")   // Reset all attributes
	fmt.Print("\033[?25h") // Show cursor
	fmt.Print("\033[0J")   // Clear from cursor to end of screen
	fmt.Print("\r")        // Move cursor to beginning of line

	// Force a flush to ensure the sequences are sent
	os.Stdout.Sync()
}
