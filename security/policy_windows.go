//go:build windows

package security

import (
	"fmt"
	"os"
)

// verifyFileSecure checks that a file has secure permissions (Windows version)
// Note: Windows doesn't have Unix-style ownership, so we only check permissions
func verifyFileSecure(filePath string, info os.FileInfo) error {
	// Check permissions: must not be writable by group or others
	mode := info.Mode()
	if mode&0022 != 0 {
		return fmt.Errorf("file has insecure permissions %s (writable by group/others)", mode.String())
	}

	// Windows doesn't have Unix-style uid/gid ownership
	// Additional Windows-specific security checks could be added here
	// (e.g., checking ACLs for administrator ownership)

	return nil
}
