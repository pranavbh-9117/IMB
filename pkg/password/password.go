// Package password provides password functionality for the IMB platform.
package password

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// Hash plain password
func Hash(plain string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("password: failed to hash: %w", err)
	}

	return string(bytes), nil
}

// Compare Passwords 
func Compare(hash, plain string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain))
	return err == nil
}
