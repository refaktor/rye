//go:build !profile
// +build !profile

package main

// startProfiling is a no-op when profiling is disabled
// Build with -tags profile to enable profiling
func startProfiling() func() {
	return func() {}
}
