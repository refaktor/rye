//go:build (linux || darwin) && !wasm

package runner

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

// setupGlobalSignalHandler sets up global signal handling for interrupting Rye programs (Unix/Linux)
func setupGlobalSignalHandler() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGTSTP)

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
