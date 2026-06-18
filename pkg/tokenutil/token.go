package tokenutil

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
)

// GenerateRefreshToken produces a cryptographically random 256-bit (32-byte)
// token encoded as a URL-safe base64 string. The raw token is returned to the
// client and must never be persisted directly; store only its hash via
// HashRefreshToken.
func GenerateRefreshToken() (string, error) {
	bytes := make([]byte, 32)

	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("tokenutil: failed to generate random bytes: %w", err)
	}

	return base64.URLEncoding.EncodeToString(bytes), nil
}

// HashRefreshToken returns the SHA-256 hex digest of the provided raw token.
// This is the value that is stored in the database. The same function must be
// called on the incoming token before any database lookup.
func HashRefreshToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}
