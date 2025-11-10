//go:build windows && !wasm

package runner

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

// setupGlobalSignalHandler sets up global signal handling for interrupting Rye programs (Windows)
func setupGlobalSignalHandler() {
	c := make(chan os.Signal, 1)
	// Windows only supports os.Interrupt and syscall.SIGTERM
	// SIGTSTP is not available on Windows
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		for sig := range c {
			programStateMutex.RLock()
			if currentPs != nil {
				currentPs.InterruptFlag = true
				fmt.Fprintf(os.Stderr, "\nReceived signal %v - interrupting operation...\n", sig)
			}
			programStateMutex.RUnlock()
		}
	}()
}
