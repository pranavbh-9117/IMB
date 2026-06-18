package password

import (
	"crypto/rand"
	"math/big"
)

// charset defines the characters allowed in the generated temporary password.
const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"

// GenerateTemp generates a cryptographically secure random string of the
// specified length. It uses crypto/rand to prevent predictability.
func GenerateTemp(length int) (string, error) {
	if length <= 0 {
		return "", nil
	}

	result := make([]byte, length)
	charsetLength := big.NewInt(int64(len(charset)))

	for i := range result {
		num, err := rand.Int(rand.Reader, charsetLength)
		if err != nil {
			return "", err
		}
		result[i] = charset[num.Int64()]
	}

	return string(result), nil
}
