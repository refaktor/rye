//go:build !windows

package security

import (
	"fmt"
	"os"
	"syscall"
)

// verifyFileSecure checks that a file has secure ownership and permissions (Unix version)
func verifyFileSecure(filePath string, info os.FileInfo) error {
	// Check permissions: must not be writable by group or others
	mode := info.Mode()
	if mode&0022 != 0 {
		return fmt.Errorf("file has insecure permissions %s (writable by group/others)", mode.String())
	}

	// On Unix systems, check if owned by root
	stat, ok := info.Sys().(*syscall.Stat_t)
	if ok && stat.Uid != 0 {
		return fmt.Errorf("file must be owned by root (current uid: %d)", stat.Uid)
	}

	return nil
}
