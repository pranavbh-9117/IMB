// Package password provides password functionality for the IMB platform.
package password

import (
	"crypto/rand"
	"math/big"
)


const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"

// Generate Temporary password when user is created
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
