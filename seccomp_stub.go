//go:build !linux || !seccomp
// +build !linux !seccomp

package main

import (
	"os"
)

// SeccompConfig holds the configuration for seccomp filtering
type SeccompConfig struct {
	Enabled bool
	Profile string
	Action  string
}

// InitSeccomp is a stub implementation for systems where seccomp is not available
// or when the seccomp build tag is not enabled
func InitSeccomp(config SeccompConfig) error {
	// Do nothing on non-Linux systems or when seccomp is not enabled

	// Still set the global CurrentSeccompProfile variable
	if config.Enabled {
		CurrentSeccompProfile = config.Profile

		// Set an environment variable that can be checked by builtins
		os.Setenv("RYE_SECCOMP_PROFILE", config.Profile)
	} else {
		CurrentSeccompProfile = ""
		os.Setenv("RYE_SECCOMP_PROFILE", "")
	}

	return nil
}

// DisableSeccompForDebug is a stub implementation for systems where seccomp is not available
func DisableSeccompForDebug() {
	// Do nothing
}

// SetupSeccompTrapHandler is a stub implementation for systems where seccomp is not available
func SetupSeccompTrapHandler() {
	// Do nothing on non-Linux systems or when seccomp is not enabled
}
