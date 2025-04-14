//go:build !windows
// +build !windows

// codesig.go
package security

import (
	"bufio"
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
)

// CodeSigConfig holds configuration for code signature verification
type CodeSigConfig struct {
	Enabled      bool   // Whether code signature verification is enabled by flag
	PubKeys      string // Path to the file containing trusted public keys
	ScriptDir    string // Directory of the script being executed
	AutoEnforced bool   // Whether code signing is auto-enforced due to .codepks in script dir
}

// TrustedPublicKeys stores the list of trusted public keys loaded from .codepks file
var TrustedPublicKeys []ed25519.PublicKey

// CurrentCodeSigEnabled indicates whether code signature verification is currently enabled
var CurrentCodeSigEnabled bool

// LoadTrustedPublicKeys loads trusted public keys from the specified file
func LoadTrustedPublicKeys(filePath string) error {
	// Clear any existing keys
	TrustedPublicKeys = nil

	// Check file ownership and permissions
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("failed to stat public keys file: %w", err)
	}

	// Get file mode to check permissions
	mode := fileInfo.Mode()

	// Check if the file is writable by group or others
	if mode&0022 != 0 {
		return fmt.Errorf("public keys file %s has insecure permissions: %s - should not be writable by group or others", filePath, mode.String())
	}

	// On Unix systems, check if the file is owned by root
	if runtime.GOOS != "windows" {
		// Get file system info to check ownership
		stat, ok := fileInfo.Sys().(*syscall.Stat_t)
		if ok {
			// Check if owner is root (uid 0)
			if stat.Uid != 0 {
				return fmt.Errorf("public keys file %s is not owned by root (uid: %d)", filePath, stat.Uid)
			}
		}
	}

	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open public keys file: %w", err)
	}
	defer file.Close()

	// Read the file line by line
	scanner := bufio.NewScanner(file)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Decode the hex-encoded public key
		pubKeyBytes, err := hex.DecodeString(line)
		if err != nil {
			return fmt.Errorf("invalid public key format at line %d: %w", lineNum, err)
		}

		// Validate key length (Ed25519 public keys are 32 bytes)
		if len(pubKeyBytes) != ed25519.PublicKeySize {
			return fmt.Errorf("invalid public key length at line %d: expected %d bytes, got %d",
				lineNum, ed25519.PublicKeySize, len(pubKeyBytes))
		}

		// Add the key to the trusted keys list
		TrustedPublicKeys = append(TrustedPublicKeys, ed25519.PublicKey(pubKeyBytes))
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading public keys file: %w", err)
	}

	if len(TrustedPublicKeys) == 0 {
		return fmt.Errorf("no valid public keys found in %s", filePath)
	}

	return nil
}

// CheckForCodePksInDir checks if a .codepks file exists in the specified directory
func CheckForCodePksInDir(dir string) (string, bool) {

	if dir == "" {
		return "", false
	}

	codePksPath := filepath.Join(dir, ".codepks")
	if _, err := os.Stat(codePksPath); err == nil {
		return codePksPath, true
	}
	return "", false
}

// InitCodeSig initializes code signature verification with the given configuration
func InitCodeSig(config CodeSigConfig) error {
	// Check for .codepks in script directory for auto-enforcement
	codePksPath := ""
	autoEnforced := false

	if config.ScriptDir != "" {
		var found bool
		codePksPath, found = CheckForCodePksInDir(config.ScriptDir)
		if found {
			autoEnforced = true
			fmt.Fprintf(os.Stderr, "Found .codepks in script directory, auto-enforcing code signing\n")
		}
	}

	// Determine if code signing should be enabled
	shouldEnable := config.Enabled || autoEnforced
	CurrentCodeSigEnabled = shouldEnable

	if !shouldEnable {
		return nil
	}

	// Determine which .codepks file to use
	pubKeysPath := config.PubKeys
	if autoEnforced && codePksPath != "" {
		// If auto-enforced, use the .codepks from the script directory
		pubKeysPath = codePksPath
	}

	// Load trusted public keys
	err := LoadTrustedPublicKeys(pubKeysPath)
	if err != nil {
		return fmt.Errorf("failed to load trusted public keys: %w", err)
	}

	// Set environment variable for builtins to check
	os.Setenv("RYE_CODESIG_ENABLED", "1")

	return nil
}

// VerifySignature verifies a signature against the content using trusted public keys
func VerifySignature(content []byte, signature []byte) bool {
	if !CurrentCodeSigEnabled {
		return true // If code signing is not enabled, consider all signatures valid
	}

	// Try to verify with any of the trusted public keys
	for _, pubKey := range TrustedPublicKeys {
		if ed25519.Verify(pubKey, content, signature) {
			return true // Signature is valid
		}
	}

	return false // Signature is not valid with any trusted public key
}
