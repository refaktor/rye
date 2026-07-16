//go:build tinygo || no_baseio
// +build tinygo no_baseio

package loader

func checkCodeSignature(content string) int {
	return 1 // Signature is valid
}
