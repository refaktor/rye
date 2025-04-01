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
	if err := InitSeccomp(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to initialize seccomp: %v\n", err)
		// Continue execution even if seccomp initialization fails
		// This ensures the program can run without seccomp if needed
	}

	supportscolor.Stdout()
	runner.DoMain(func(ps *env.ProgramState) {})
}

// ErrNoPath is returned when 'cd' was called without a second argument.
var ErrNoPath = errors.New("path required")
