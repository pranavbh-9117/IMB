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

// Validate JWT and extracts the payload
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

		c.Set(userIDKey, userID)
		c.Set(roleKey, domain.Role(claims.Role))
		c.Set(institutionIDKey, instID)

	
		c.Next()
	}
}
