//go:build linux || darwin || windows
// +build linux darwin windows

package main

import (
	"errors"

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
	runner.DoMain(func(ps *env.ProgramState) {})
}

// ErrNoPath is returned when 'cd' was called without a second argument.
var ErrNoPath = errors.New("path required")
