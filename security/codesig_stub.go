//go:build windows
// +build windows

// codesig_stub.go - Code signature verification for Rye scripts (Windows version)
//
// Code signing is configured through security policies (.ryesec, /etc/rye/*.yaml, or embedded).
// Public keys can be specified inline in the policy or loaded from a separate file.
//
// Note: On Windows, root ownership checks are not performed.
package security

import (
	"bufio"
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
)

// TrustedPublicKeys stores the list of trusted public keys
var TrustedPublicKeys []ed25519.PublicKey

// CurrentCodeSigEnabled indicates whether code signature verification is currently enabled
var CurrentCodeSigEnabled bool

// LoadPublicKeysFromFile loads trusted public keys from a file
// On Windows, we only check that the file is not world-writable (no root ownership check)
func LoadPublicKeysFromFile(filePath string) error {
	// Check file permissions
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("failed to stat public keys file: %w", err)
	}

	mode := fileInfo.Mode()
	if mode&0022 != 0 {
		return fmt.Errorf("public keys file %s has insecure permissions: %s", filePath, mode.String())
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

	return LoadPublicKeysFromStrings(keys)
}

// LoadPublicKeysFromStrings loads trusted public keys from a slice of hex-encoded strings
func LoadPublicKeysFromStrings(hexKeys []string) error {
	TrustedPublicKeys = nil

	for i, hexKey := range hexKeys {
		if strings.TrimSpace(hexKey) == "" {
			continue
		}

		pubKeyBytes, err := hex.DecodeString(hexKey)
		if err != nil {
			return fmt.Errorf("invalid public key format at index %d: %w", i, err)
		}

		if len(pubKeyBytes) != ed25519.PublicKeySize {
			return fmt.Errorf("invalid public key length at index %d: expected %d bytes, got %d",
				i, ed25519.PublicKeySize, len(pubKeyBytes))
		}

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
		return true
	}

	for _, pubKey := range TrustedPublicKeys {
		if ed25519.Verify(pubKey, content, signature) {
			return true
		}
	}

	return false
}
