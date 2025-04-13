//go:build linux || darwin || windows
// +build linux darwin windows

package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/jwalton/go-supportscolor"
	"github.com/refaktor/rye/env"
	"github.com/refaktor/rye/runner"
	"github.com/refaktor/rye/security"
)

// Global variables to store the current security profiles
// These can be accessed from builtins to enforce additional restrictions
// These are now defined in the security package

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
	// Initialize security profiles
	// These are no-ops on non-Linux systems or when built without the appropriate tags
	// The actual configuration will be set in runner.DoMain based on command-line flags

	supportscolor.Stdout()
	runner.DoMain(func(ps *env.ProgramState) {
		// Initialize seccomp with configuration from command-line flags
		seccompConfig := security.SeccompConfig{
			Enabled: *runner.SeccompProfile != "",
			Profile: *runner.SeccompProfile,
			Action:  *runner.SeccompAction,
		}

		// Initialize landlock with configuration from command-line flags
		landlockConfig := security.LandlockConfig{
			Enabled: *runner.LandlockEnabled,
			Profile: *runner.LandlockProfile,
			Paths:   strings.Split(*runner.LandlockPaths, ","),
		}

		// Initialize seccomp if profile is set
		if seccompConfig.Enabled {
			// If using trap action, set up the trap handler
			if seccompConfig.Action == "trap" {
				security.SetupSeccompTrapHandler()
			}

			if err := security.InitSeccomp(seccompConfig); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: Failed to initialize seccomp: %v\n", err)
				// Continue execution even if seccomp initialization fails
				// This ensures the program can run without seccomp if needed
			} else {
				// DEBUG: fmt.Fprintf(os.Stderr, "Seccomp initialized with profile: %s\n", seccompConfig.Profile)
			}
		}

		// Initialize landlock if enabled
		if landlockConfig.Enabled {
			// Clean up empty paths that might result from splitting an empty string
			var cleanPaths []string
			for _, path := range landlockConfig.Paths {
				if path != "" {
					cleanPaths = append(cleanPaths, path)
				}
			}
			landlockConfig.Paths = cleanPaths

			if err := security.InitLandlock(landlockConfig); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: Failed to initialize landlock: %v\n", err)
				// Continue execution even if landlock initialization fails
				// This ensures the program can run without landlock if needed
			}
		}
	})
}

// ErrNoPath is returned when 'cd' was called without a second argument.
var ErrNoPath = errors.New("path required")
