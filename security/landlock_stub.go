//go:build !linux || !landlock
// +build !linux !landlock

package security

import (
	"fmt"
	"os"
)

// LandlockConfig holds the configuration for landlock filesystem access control
// This is a stub implementation for non-Linux systems
type LandlockConfig struct {
	Enabled bool
	Profile string
	Paths   []string
}

// CurrentLandlockProfile stores the active landlock profile
var CurrentLandlockProfile string

// InitLandlock initializes the landlock filesystem access control for Rye
// This is a stub implementation for non-Linux systems
func InitLandlock(config LandlockConfig) error {
	if config.Enabled {
		fmt.Println("Warning: Landlock is only supported on Linux systems")
	}

	// Still set the global CurrentLandlockProfile variable
	if config.Enabled {
		CurrentLandlockProfile = config.Profile
		// Set an environment variable that can be checked by builtins
		os.Setenv("RYE_LANDLOCK_PROFILE", config.Profile)
	} else {
		CurrentLandlockProfile = ""
		os.Setenv("RYE_LANDLOCK_PROFILE", "")
	}

	return nil
}
