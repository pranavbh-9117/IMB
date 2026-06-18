// Package password provides password functionality for the IMB platform.
package password

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// Hash generates a bcrypt hash of the provided plaintext password using the
// default cost factor. The returned string is safe to store directly in the
// database. An error is returned if hashing fails.
func Hash(plain string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("password: failed to hash: %w", err)
	}

	return string(bytes), nil
}

// Compare reports whether plain matches the stored bcrypt hash.
// Returns true only when the plaintext produces an identical hash.
// Returns false for any mismatch, including malformed hashes.
func Compare(hash, plain string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain))
	return err == nil
}
