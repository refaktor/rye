package security

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"

	"gopkg.in/yaml.v3"
)

// PolicySource indicates where the security policy came from
type PolicySource string

const (
	PolicySourceNone     PolicySource = "none"
	PolicySourceCLI      PolicySource = "cli"
	PolicySourceLocal    PolicySource = "local"    // .ryesec in script dir
	PolicySourceSystem   PolicySource = "system"   // /etc/rye/
	PolicySourceEmbedded PolicySource = "embedded" // Compiled into binary
)

// SecurityPolicy represents a complete security configuration
type SecurityPolicy struct {
	// Metadata
	Version     string       `yaml:"version"`
	Description string       `yaml:"description,omitempty"`
	Source      PolicySource `yaml:"-"` // Set at runtime, not from file

	// Seccomp configuration
	Seccomp struct {
		Enabled bool   `yaml:"enabled"`
		Profile string `yaml:"profile"` // "strict", "readonly"
		Action  string `yaml:"action"`  // "errno", "kill", "trap", "log"
	} `yaml:"seccomp"`

	// Landlock configuration
	Landlock struct {
		Enabled bool     `yaml:"enabled"`
		Profile string   `yaml:"profile"` // "readonly", "readexec", "custom"
		Paths   []string `yaml:"paths"`   // For custom profile: "/path:rw" format
	} `yaml:"landlock"`

	// Code signing configuration
	CodeSig struct {
		Enforced      bool     `yaml:"enforced"`
		PublicKeys    []string `yaml:"public_keys,omitempty"`      // Inline hex-encoded keys
		PublicKeysFile string  `yaml:"public_keys_file,omitempty"` // Path to file with keys (must be root-owned)
	} `yaml:"codesig"`

	// Policy enforcement
	Mandatory bool `yaml:"mandatory"` // If true, cannot be relaxed by CLI flags

	// Allowed paths for scripts (if empty, any script can run)
	AllowedScriptPaths []string `yaml:"allowed_script_paths,omitempty"`
}

// SystemPolicyPaths defines where to look for system-wide policies
var SystemPolicyPaths = []string{
	"/etc/rye/mandatory.yaml",
	"/etc/rye/security.yaml",
}

// LocalPolicyFilename is the name of local policy files
const LocalPolicyFilename = ".ryesec"

// LoadSecurityPolicy loads security policy with the following precedence:
// 1. Embedded policy (highest - compiled into binary)
// 2. System policy (/etc/rye/mandatory.yaml)
// 3. Local policy (.ryesec in script directory)
// 4. CLI flags (lowest)
func LoadSecurityPolicy(scriptDir string, cliPolicy *SecurityPolicy) (*SecurityPolicy, error) {
	// 1. Check for embedded policy first (highest priority)
	if embeddedPolicy := GetEmbeddedPolicy(); embeddedPolicy != nil {
		embeddedPolicy.Source = PolicySourceEmbedded
		return embeddedPolicy, nil
	}

	// 2. Check for system-wide mandatory policy
	for _, sysPath := range SystemPolicyPaths {
		if policy, err := loadPolicyFromFile(sysPath); err == nil {
			policy.Source = PolicySourceSystem
			return policy, nil
		}
	}

	// 3. Check for local .ryesec in script directory
	if scriptDir != "" {
		localPath := filepath.Join(scriptDir, LocalPolicyFilename)
		if policy, err := loadPolicyFromFile(localPath); err == nil {
			policy.Source = PolicySourceLocal
			return policy, nil
		}
	}

	// 4. Fall back to CLI policy
	if cliPolicy != nil {
		cliPolicy.Source = PolicySourceCLI
		return cliPolicy, nil
	}

	// No policy found
	return &SecurityPolicy{Source: PolicySourceNone}, nil
}

// loadPolicyFromFile loads a policy from a YAML file with security checks
func loadPolicyFromFile(filePath string) (*SecurityPolicy, error) {
	// Check if file exists
	info, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}

	// Security check: verify file ownership and permissions
	if err := verifyFileSecure(filePath, info); err != nil {
		return nil, fmt.Errorf("security policy file %s is insecure: %w", filePath, err)
	}

	// Read and parse the file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read policy file: %w", err)
	}

	var policy SecurityPolicy
	if err := yaml.Unmarshal(data, &policy); err != nil {
		return nil, fmt.Errorf("failed to parse policy file: %w", err)
	}

	return &policy, nil
}

// verifyFileSecure checks that a file has secure ownership and permissions
func verifyFileSecure(filePath string, info os.FileInfo) error {
	// Check permissions: must not be writable by group or others
	mode := info.Mode()
	if mode&0022 != 0 {
		return fmt.Errorf("file has insecure permissions %s (writable by group/others)", mode.String())
	}

	// On Unix systems, check if owned by root
	if runtime.GOOS != "windows" {
		stat, ok := info.Sys().(*syscall.Stat_t)
		if ok && stat.Uid != 0 {
			return fmt.Errorf("file must be owned by root (current uid: %d)", stat.Uid)
		}
	}

	return nil
}

// ApplySecurityPolicy applies the given security policy
func ApplySecurityPolicy(policy *SecurityPolicy) error {
	if policy == nil {
		return nil
	}

	// Log policy source
	if policy.Source != PolicySourceNone {
		fmt.Fprintf(os.Stderr, "\033[2;37mSecurity policy loaded from: %s\033[0m\n", policy.Source)
	}

	// Apply seccomp
	if policy.Seccomp.Enabled {
		seccompConfig := SeccompConfig{
			Enabled: true,
			Profile: policy.Seccomp.Profile,
			Action:  policy.Seccomp.Action,
		}
		if err := InitSeccomp(seccompConfig); err != nil {
			return fmt.Errorf("failed to apply seccomp policy: %w", err)
		}
	}

	// Apply landlock
	if policy.Landlock.Enabled {
		landlockConfig := LandlockConfig{
			Enabled: true,
			Profile: policy.Landlock.Profile,
			Paths:   policy.Landlock.Paths,
		}
		if err := InitLandlock(landlockConfig); err != nil {
			return fmt.Errorf("failed to apply landlock policy: %w", err)
		}
	}

	// Apply code signing
	if policy.CodeSig.Enforced {
		keysLoaded := false

		// First try inline keys
		if len(policy.CodeSig.PublicKeys) > 0 {
			if err := LoadPublicKeysFromStrings(policy.CodeSig.PublicKeys); err != nil {
				return fmt.Errorf("failed to load inline public keys: %w", err)
			}
			keysLoaded = true
		}

		// Then try keys file (can add additional keys or be the only source)
		if policy.CodeSig.PublicKeysFile != "" {
			if err := LoadPublicKeysFromFile(policy.CodeSig.PublicKeysFile); err != nil {
				return fmt.Errorf("failed to load public keys from file: %w", err)
			}
			keysLoaded = true
		}

		if !keysLoaded {
			return fmt.Errorf("code signing enforced but no public keys provided (use public_keys or public_keys_file)")
		}

		CurrentCodeSigEnabled = true
		os.Setenv("RYE_CODESIG_ENABLED", "1")
		fmt.Fprintf(os.Stderr, "\033[2;37mCode signing enabled with %d trusted key(s)\033[0m\n", len(TrustedPublicKeys))
	}

	return nil
}

// MergePolicies merges a base policy with overrides
// Only allows overrides to be MORE restrictive, not less
func MergePolicies(base, override *SecurityPolicy) *SecurityPolicy {
	if base == nil {
		return override
	}
	if override == nil {
		return base
	}

	result := *base // Copy base

	// If base is mandatory, don't allow any relaxation
	if base.Mandatory {
		// Can only ADD restrictions
		if override.Seccomp.Enabled && !result.Seccomp.Enabled {
			result.Seccomp = override.Seccomp
		}
		if override.Landlock.Enabled && !result.Landlock.Enabled {
			result.Landlock = override.Landlock
		}
		if override.CodeSig.Enforced && !result.CodeSig.Enforced {
			result.CodeSig = override.CodeSig
		}
	} else {
		// Non-mandatory: override takes precedence but we log it
		if override.Seccomp.Enabled {
			result.Seccomp = override.Seccomp
		}
		if override.Landlock.Enabled {
			result.Landlock = override.Landlock
		}
		if override.CodeSig.Enforced {
			result.CodeSig = override.CodeSig
		}
	}

	return &result
}

// ValidateScriptPath checks if a script is allowed to run under this policy
func (p *SecurityPolicy) ValidateScriptPath(scriptPath string) error {
	if len(p.AllowedScriptPaths) == 0 {
		return nil // No restrictions
	}

	absScript, err := filepath.Abs(scriptPath)
	if err != nil {
		return fmt.Errorf("cannot resolve script path: %w", err)
	}

	for _, allowed := range p.AllowedScriptPaths {
		absAllowed, err := filepath.Abs(allowed)
		if err != nil {
			continue
		}

		// Check if script is under allowed path
		if strings.HasPrefix(absScript, absAllowed) {
			return nil
		}
	}

	return fmt.Errorf("script %s is not in allowed paths", scriptPath)
}

// String returns a human-readable description of the policy
func (p *SecurityPolicy) String() string {
	var parts []string

	parts = append(parts, fmt.Sprintf("Source: %s", p.Source))

	if p.Seccomp.Enabled {
		parts = append(parts, fmt.Sprintf("Seccomp: %s (action: %s)", p.Seccomp.Profile, p.Seccomp.Action))
	}

	if p.Landlock.Enabled {
		parts = append(parts, fmt.Sprintf("Landlock: %s", p.Landlock.Profile))
		if len(p.Landlock.Paths) > 0 {
			parts = append(parts, fmt.Sprintf("  Paths: %v", p.Landlock.Paths))
		}
	}

	if p.CodeSig.Enforced {
		parts = append(parts, fmt.Sprintf("CodeSig: enforced (%d keys)", len(p.CodeSig.PublicKeys)))
	}

	if p.Mandatory {
		parts = append(parts, "Mandatory: yes (cannot be relaxed)")
	}

	return strings.Join(parts, "\n")
}
