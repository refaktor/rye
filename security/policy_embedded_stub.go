//go:build !embed_security
// +build !embed_security

package security

// GetEmbeddedPolicy returns nil when no policy is embedded
// This stub is used when building without the embed_security tag
func GetEmbeddedPolicy() *SecurityPolicy {
	return nil
}

// IsEmbeddedBuild returns false for non-embedded builds
func IsEmbeddedBuild() bool {
	return false
}
