//go:build !linux

package runner

import (
	"fmt"
	"os"
)

// UnshareConfig holds the namespace isolation options. On non-Linux systems
// these options are parsed but have no effect.
type UnshareConfig struct {
	Fs  bool
	Net bool
	Pid bool
	Uts bool
}

// IsUnshareChild always returns false on non-Linux systems.
func IsUnshareChild() bool {
	return false
}

// ReadUnshareChildConfig returns a zero-value config on non-Linux systems.
func ReadUnshareChildConfig() UnshareConfig {
	return UnshareConfig{}
}

// DoReexecInUnshare prints an error and exits on non-Linux systems because
// Linux namespaces are not available.
func DoReexecInUnshare(_ UnshareConfig) {
	fmt.Fprintf(os.Stderr, "rye --unshare: namespace isolation is only supported on Linux\n")
	os.Exit(1)
}

// SetupUnshareFilesystem is a no-op on non-Linux systems.
func SetupUnshareFilesystem() error {
	return nil
}
