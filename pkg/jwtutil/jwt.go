// Package jwtutil provides jwtutil functionality for the IMB platform.
package jwtutil

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Payload embedded in every token
type Claims struct {
	UserID        string `json:"user_id"`
	Role          string `json:"role"`
	InstitutionID string `json:"institution_id"`
	jwt.RegisteredClaims
}

// Generate JWT
func GenerateAccessToken(userID, role, institutionID string, expiry time.Duration, secret string) (string, error) {
	now := time.Now()

	claims := Claims{
		UserID:        userID,
		Role:          role,
		InstitutionID: institutionID,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(expiry)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("jwtutil: failed to sign access token: %w", err)
	}

	return signed, nil
}

// Validate JWT
func ValidateAccessToken(tokenString, secret string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("jwtutil: unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("jwtutil: failed to parse access token: %w", err)
	}

	if !token.Valid {
		return nil, errors.New("jwtutil: access token is invalid")
	}

	return claims, nil
}
