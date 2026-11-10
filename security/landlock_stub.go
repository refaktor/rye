//go:build !linux || !landlock
// +build !linux !landlock

package security

import (
	"fmt"
	"os"
	"runtime"
)

// LandlockConfig holds the configuration for landlock filesystem access control
// This is a stub implementation for non-Linux systems, or Linux builds without
// the landlock build tag.
type LandlockConfig struct {
	Enabled bool
	Profile string
	Paths   []string
}

// CurrentLandlockProfile stores the active landlock profile
var CurrentLandlockProfile string

// InitLandlock initializes the landlock filesystem access control for Rye.
// Requesting landlock when it is unavailable is an error: silently continuing
// without the requested sandbox would be fail-open.
func InitLandlock(config LandlockConfig) error {
	CurrentLandlockProfile = ""
	os.Setenv("RYE_LANDLOCK_PROFILE", "")

	if !config.Enabled {
		return nil
	}
	if runtime.GOOS != "linux" {
		return fmt.Errorf("landlock is only available on Linux")
	}
	return fmt.Errorf("landlock support is not compiled into this binary; rebuild with -tags landlock")
}
