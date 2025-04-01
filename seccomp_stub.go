//go:build !linux || !seccomp
// +build !linux !seccomp

package main

// InitSeccomp is a stub implementation for systems where seccomp is not available
// or when the seccomp build tag is not enabled
func InitSeccomp() error {
	// Do nothing on non-Linux systems or when seccomp is not enabled
	return nil
}

// DisableSeccompForDebug is a stub implementation for systems where seccomp is not available
func DisableSeccompForDebug() {
	// Do nothing
}
