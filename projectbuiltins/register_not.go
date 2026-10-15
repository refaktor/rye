//go:build !proj_builtins
// +build !proj_builtins

// Package projectbuiltins is a reserved, empty-by-default package into which
// the local Rye builder (bin/l.rye) drops a project's own Go builtins.
//
// In a normal build (without the "proj_builtins" build tag) Register is a
// no-op, so the package has no effect and the binary builds exactly as before.
//
// When a project supplies a project_builtins/ directory, l.rye copies its
// *.go files here and builds with the "proj_builtins" tag, which selects
// register_on.go instead of this file.
package projectbuiltins

import "github.com/refaktor/rye/env"

// Register registers project-specific builtins into the program state.
// This no-op implementation is used in normal builds.
func Register(ps *env.ProgramState) {}
