//go:build !windows
// +build !windows

// codesig.go - Code signature verification for Rye scripts
//
// Code signing is configured through security policies (.ryesec, /etc/rye/*.yaml, or embedded).
// Public keys can be specified inline in the policy or loaded from a separate file.
package security

import (
	"bufio"
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"os"
	"runtime"
	"strings"
	"syscall"
)

// TrustedPublicKeys stores the list of trusted public keys
var TrustedPublicKeys []ed25519.PublicKey

// CurrentCodeSigEnabled indicates whether code signature verification is currently enabled
var CurrentCodeSigEnabled bool

// LoadPublicKeysFromFile loads trusted public keys from a file
// The file must be owned by root and not writable by group/others
func LoadPublicKeysFromFile(filePath string) error {
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

	// On Unix systems, check if owned by root
	if runtime.GOOS != "windows" {
		stat, ok := fileInfo.Sys().(*syscall.Stat_t)
		if ok {
			if stat.Uid != 0 {
				return fmt.Errorf("public keys file %s must be owned by root (current uid: %d)", filePath, stat.Uid)
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
	var keys []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		keys = append(keys, line)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading public keys file: %w", err)
	}

	if len(keys) == 0 {
		return fmt.Errorf("no valid public keys found in %s", filePath)
	}

	// Use the common function to parse the keys
	return LoadPublicKeysFromStrings(keys)
}

// LoadPublicKeysFromStrings loads trusted public keys from a slice of hex-encoded strings
func LoadPublicKeysFromStrings(hexKeys []string) error {
	// Clear any existing keys
	TrustedPublicKeys = nil

	for i, hexKey := range hexKeys {
		// Skip empty strings
		if strings.TrimSpace(hexKey) == "" {
			continue
		}

		// Decode the hex-encoded public key
		pubKeyBytes, err := hex.DecodeString(hexKey)
		if err != nil {
			return fmt.Errorf("invalid public key format at index %d: %w", i, err)
		}

		// Validate key length (Ed25519 public keys are 32 bytes)
		if len(pubKeyBytes) != ed25519.PublicKeySize {
			return fmt.Errorf("invalid public key length at index %d: expected %d bytes, got %d",
				i, ed25519.PublicKeySize, len(pubKeyBytes))
		}

		// Add the key to the trusted keys list
		TrustedPublicKeys = append(TrustedPublicKeys, ed25519.PublicKey(pubKeyBytes))
	}

	if len(TrustedPublicKeys) == 0 {
		return fmt.Errorf("no valid public keys provided")
	}

	return nil
}

// VerifySignature verifies a signature against the content using trusted public keys
func VerifySignature(content []byte, signature []byte) bool {
	if !CurrentCodeSigEnabled {
		fmt.Println("codesig not enabled")
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
