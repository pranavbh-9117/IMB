// Package middleware provides middleware functionality for the IMB platform.
package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/pranavbh-9117/IMB/internal/domain"
	"github.com/pranavbh-9117/IMB/pkg/jwtutil"
	"github.com/pranavbh-9117/IMB/pkg/response"
)

// RequireAuth creates a Gin middleware that validates the JWT access token
// provided in the Authorization header. If the token is valid, it extracts the
// UserID, Role, and InstitutionID claims, parses them into appropriate Go
// types, and stores them in the request context for downstream handlers.
// If the token is missing, malformed, or invalid, the request is aborted with
// a 401 Unauthorized response.
func RequireAuth(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "authorization header is required")
			c.Abort()
			return
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			response.Unauthorized(c, "authorization header must use Bearer scheme")
			c.Abort()
			return
		}

		rawToken := strings.TrimPrefix(authHeader, "Bearer ")

		claims, err := jwtutil.ValidateAccessToken(rawToken, jwtSecret)
		if err != nil {
			response.Unauthorized(c, "invalid or expired access token")
			c.Abort()
			return
		}

		// Parse claims into proper types before storing in context
		userID, err := uuid.Parse(claims.UserID)
		if err != nil {
			response.Unauthorized(c, "malformed user ID in token")
			c.Abort()
			return
		}

		var instID *uuid.UUID
		if claims.InstitutionID != "" {
			parsedID, err := uuid.Parse(claims.InstitutionID)
			if err != nil {
				response.Unauthorized(c, "malformed institution ID in token")
				c.Abort()
				return
			}
			instID = &parsedID
		}

		// Store typed values in the context using unexported keys
		c.Set(userIDKey, userID)
		c.Set(roleKey, domain.Role(claims.Role))
		c.Set(institutionIDKey, instID)

		// Proceed to the next handler
		c.Next()
	}
}
