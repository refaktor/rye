//go:build linux || darwin || windows
// +build linux darwin windows

package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/jwalton/go-supportscolor"
	"github.com/refaktor/rye/env"
	"github.com/refaktor/rye/runner"
)

// Global variable to store the current seccomp profile
// This can be accessed from builtins to enforce additional restrictions
var CurrentSeccompProfile string

type TagType int
type RjType int
type Series []any

type anyword struct {
	kind RjType
	idx  int
}

type node struct {
	kind  RjType
	value any
}

var CODE []any

//
// main function. Dispatches to appropriate mode function
//

// NEW FLASGS HANDLING

func main() {
	// Initialize seccomp profile
	// This is a no-op on non-Linux systems or when built without the seccomp tag
	// The actual configuration will be set in runner.DoMain based on command-line flags

	supportscolor.Stdout()
	runner.DoMain(func(ps *env.ProgramState) {
		// Initialize seccomp with configuration from command-line flags
		config := SeccompConfig{
			Enabled: *runner.SeccompProfile != "",
			Profile: *runner.SeccompProfile,
			Action:  *runner.SeccompAction,
		}

		// Initialize seccomp if profile is set
		if config.Enabled {
			// If using trap action, set up the trap handler
			if config.Action == "trap" {
				SetupSeccompTrapHandler()
			}

			if err := InitSeccomp(config); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: Failed to initialize seccomp: %v\n", err)
				// Continue execution even if seccomp initialization fails
				// This ensures the program can run without seccomp if needed
			} else {
				// DEBUG: fmt.Fprintf(os.Stderr, "Seccomp initialized with profile: %s\n", config.Profile)
			}
		}
	})
}

// ErrNoPath is returned when 'cd' was called without a second argument.
var ErrNoPath = errors.New("path required")
