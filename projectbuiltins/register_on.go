//go:build proj_builtins
// +build proj_builtins

package projectbuiltins

import "github.com/refaktor/rye/env"

// Register registers project-specific builtins by delegating to
// RegisterProjectBuiltins, which must be defined in one of the Go files that
// l.rye copied from the project's project_builtins/ directory.
func Register(ps *env.ProgramState) {
	RegisterProjectBuiltins(ps)
}
