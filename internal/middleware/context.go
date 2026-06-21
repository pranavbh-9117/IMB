// Package middleware provides middleware functionality for the IMB platform.
package middleware

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/pranavbh-9117/IMB/internal/domain"
)


const (
	userIDKey        = "auth.user_id"
	roleKey          = "auth.role"
	institutionIDKey = "auth.institution_id"
)

//  Custom error indicates that a protected handler was mounted without the RequireAuth middleware.
var ErrContextValueMissing = errors.New("authentication value missing from context")

// Retrieve UserID form gin Context
func GetUserID(c *gin.Context) (uuid.UUID, error) {
	val, exists := c.Get(userIDKey)
	if !exists {
		return uuid.Nil, ErrContextValueMissing
	}

	id, ok := val.(uuid.UUID)
	if !ok {
		return uuid.Nil, errors.New("user ID in context is not a valid UUID")
	}

	return id, nil
}

// Retrieve UserRole from gin Context
func GetRole(c *gin.Context) (domain.Role, error) {
	val, exists := c.Get(roleKey)
	if !exists {
		return "", ErrContextValueMissing
	}

	role, ok := val.(domain.Role)
	if !ok {
		return "", errors.New("role in context is not a valid domain.Role")
	}

	return role, nil
}

// Retrieve InstitutionID from gin Context
func GetInstitutionID(c *gin.Context) (*uuid.UUID, error) {
	val, exists := c.Get(institutionIDKey)
	if !exists {
		return nil, ErrContextValueMissing
	}
	
	instID, ok := val.(*uuid.UUID)
	if !ok {
		return nil, errors.New("institution ID in context is not a valid *uuid.UUID")
	}

	return instID, nil
}
