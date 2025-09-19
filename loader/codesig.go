//go:build !tinygo

package loader

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/refaktor/rye/security"
)

func checkCodeSignature(content string) int {
	parts := strings.SplitN(content, ";ryesig ", 2)
	content = strings.TrimSpace(parts[0])
	if len(parts) != 2 {
		fmt.Println("\x1b[33m" + "No rye signature found. Exiting." + "\x1b[0m")
		return -1
	}

	signature := parts[1]
	sig := strings.TrimSpace(signature)
	bsig, err := hex.DecodeString(sig)
	if err != nil {
		fmt.Println("\x1b[33m" + "Invalid signature format: " + err.Error() + "\x1b[0m")
		return -2
	}

	// Verify the signature using the security package
	if security.VerifySignature([]byte(content), bsig) {
		return 1 // Signature is valid
	}

	fmt.Println("\x1b[33m" + "Rye signature is not valid with any trusted public key! Exiting." + "\x1b[0m")
	return -2
}
