//go:build !linux || !seccomp
// +build !linux !seccomp

package main

// Seccomp2Config holds the configuration for seccomp2 filtering
type Seccomp2Config struct {
	Enabled bool
	Profile string
	Action  string
}

// InitSeccomp2 is a stub implementation for systems where seccomp is not available
// or when the seccomp build tag is not enabled
func InitSeccomp2(config Seccomp2Config) error {
	// Do nothing on non-Linux systems or when seccomp is not enabled
	return nil
}

// DisableSeccomp2ForDebug is a stub implementation for systems where seccomp is not available
func DisableSeccomp2ForDebug() {
	// Do nothing
}
