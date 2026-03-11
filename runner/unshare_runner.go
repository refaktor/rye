//go:build !wasm

package runner

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/refaktor/rye/security"
)

// peekUnsharePolicy does a quick, lightweight read of the .ryesec file in
// scriptDir to determine whether unshare isolation is requested by a local
// project policy. It intentionally skips the full security validation that
// LoadSecurityPolicy performs — the full validation still runs later during
// normal policy loading.
//
// Returns (enabled, config, found):
//   - found  = false if no .ryesec file exists or it cannot be parsed
//   - enabled = true  if the file requests namespace isolation
//   - config  = the requested namespace options when enabled is true
func peekUnsharePolicy(scriptDir string) (enabled bool, cfg UnshareConfig, found bool) {
	localPath := filepath.Join(scriptDir, security.LocalPolicyFilename)
	data, err := os.ReadFile(localPath)
	if err != nil {
		// File absent or unreadable — that is perfectly normal.
		return false, UnshareConfig{}, false
	}

	var pol security.SecurityPolicy
	if err := yaml.Unmarshal(data, &pol); err != nil {
		// Malformed file; let the full policy loader report the error later.
		return false, UnshareConfig{}, false
	}

	if !pol.Unshare.Enabled {
		return false, UnshareConfig{}, true
	}

	return true, UnshareConfig{
		Fs:  pol.Unshare.Fs,
		Net: pol.Unshare.Net,
		Pid: pol.Unshare.Pid,
		Uts: pol.Unshare.Uts,
	}, true
}
