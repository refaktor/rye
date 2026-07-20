//go:build !wasm && no_baseio
// +build !wasm,no_baseio

package util

// BeforeExit is a no-op stub when the no_baseio build tag is active.
// The keyboard package is not imported, so there is nothing to close.
func BeforeExit() {}
