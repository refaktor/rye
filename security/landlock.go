//go:build linux && landlock
// +build linux,landlock

// To build with landlock support:
// go build -tags landlock

package security

import (
	"fmt"
	"os"
	"strings"

	"github.com/landlock-lsm/go-landlock/landlock"
)

// LandlockConfig holds the configuration for landlock filesystem access control
type LandlockConfig struct {
	Enabled bool
	Profile string
	Paths   []string
}

// CurrentLandlockProfile stores the active landlock profile
var CurrentLandlockProfile string

// isValidLandlockProfile checks if the specified profile is valid
func isValidLandlockProfile(profile string) bool {
	validProfiles := []string{"readonly", "readexec", "custom"}
	for _, p := range validProfiles {
		if profile == p {
			return true
		}
	}
	return false
}

// parseCustomAccessRights parses custom access rights from path specifications
// Format: path:permissions where permissions can be r (read), w (write), x (execute)
// Example: /home/user/data:rw,/tmp:rx
func parseCustomAccessRights(paths []string) ([]landlock.Rule, error) {
	var rules []landlock.Rule

	for _, pathSpec := range paths {
		if !strings.Contains(pathSpec, ":") {
			// Default to read-only if no permissions specified
			rules = append(rules, landlock.RODirs(pathSpec))
			continue
		}

		parts := strings.Split(pathSpec, ":")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid path specification: %s", pathSpec)
		}

		path := parts[0]
		perms := parts[1]

		// Check if the path exists and is a directory or file
		info, err := os.Stat(path)
		if err != nil {
			return nil, fmt.Errorf("failed to stat path %s: %w", path, err)
		}

		isDir := info.IsDir()

		// Parse permissions
		hasRead := strings.Contains(perms, "r")
		hasWrite := strings.Contains(perms, "w")
		// hasExec is not used directly as the landlock library doesn't expose execute permissions
		// for individual files in the same way. The execute permission is handled at the ABI level.
		_ = strings.Contains(perms, "x")

		// Apply appropriate rules based on permissions and path type
		if isDir {
			if hasRead && hasWrite {
				rules = append(rules, landlock.RWDirs(path))
			} else if hasRead {
				rules = append(rules, landlock.RODirs(path))
			}
		} else {
			if hasRead && hasWrite {
				rules = append(rules, landlock.RWFiles(path))
			} else if hasRead {
				rules = append(rules, landlock.ROFiles(path))
			}
		}

		// Note: The landlock library doesn't directly expose execute permissions
		// for individual files. The execute permission is handled at the ABI level.
	}

	return rules, nil
}

// InitLandlock initializes the landlock filesystem access control for Rye
func InitLandlock(config LandlockConfig) error {
	// Skip if landlock is disabled
	if !config.Enabled {
		return nil
	}

	// Check if the profile is valid
	if !isValidLandlockProfile(config.Profile) {
		return fmt.Errorf("invalid landlock profile: %s (valid profiles: readonly, readexec, custom)", config.Profile)
	}

	// Prepare rules based on profile
	var rules []landlock.Rule

	switch config.Profile {
	case "readonly":
		// If paths are specified, use them; otherwise, use current directory
		paths := config.Paths
		if len(paths) == 0 {
			wd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get working directory: %w", err)
			}
			paths = []string{wd}
		}

		for _, path := range paths {
			// Check if the path exists and is a directory or file
			info, err := os.Stat(path)
			if err != nil {
				return fmt.Errorf("failed to stat path %s: %w", path, err)
			}

			if info.IsDir() {
				rules = append(rules, landlock.RODirs(path))
			} else {
				rules = append(rules, landlock.ROFiles(path))
			}
		}

	case "readexec":
		// Similar to readonly but with execute permissions
		// Note: The landlock library doesn't directly expose execute permissions
		// We'll use the same RODirs/ROFiles as readonly, and the kernel will
		// handle execute permissions based on the file permissions
		paths := config.Paths
		if len(paths) == 0 {
			wd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get working directory: %w", err)
			}
			paths = []string{wd}
		}

		for _, path := range paths {
			// Check if the path exists and is a directory or file
			info, err := os.Stat(path)
			if err != nil {
				return fmt.Errorf("failed to stat path %s: %w", path, err)
			}

			if info.IsDir() {
				rules = append(rules, landlock.RODirs(path))
			} else {
				rules = append(rules, landlock.ROFiles(path))
			}
		}

	case "custom":
		// Parse custom access rights
		customRules, err := parseCustomAccessRights(config.Paths)
		if err != nil {
			return fmt.Errorf("failed to parse custom access rights: %w", err)
		}
		rules = append(rules, customRules...)
	}

	// Apply landlock restrictions
	err := landlock.V1.BestEffort().RestrictPaths(rules...)
	if err != nil {
		return fmt.Errorf("failed to apply landlock restrictions: %w", err)
	}

	// Set the global CurrentLandlockProfile variable
	CurrentLandlockProfile = config.Profile

	// Set an environment variable that can be checked by builtins
	os.Setenv("RYE_LANDLOCK_PROFILE", config.Profile)

	fmt.Printf("\033[2;37mInitializing landlock with profile: %s\033[0m\n", config.Profile)
	return nil
}
