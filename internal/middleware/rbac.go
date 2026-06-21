// Package middleware provides middleware functionality for the IMB platform.
package middleware

import (
	"github.com/gin-gonic/gin"

	"github.com/pranavbh-9117/IMB/internal/domain"
	"github.com/pranavbh-9117/IMB/pkg/response"
)

// Check for Authorization
func RequireRoles(allowedRoles ...domain.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, err := GetRole(c)
		if err != nil {
			response.Unauthorized(c, "authentication required")
			c.Abort()
			return
		}

		for _, allowed := range allowedRoles {
			if role == allowed {
				c.Next()
				return
			}
		}

		response.Forbidden(c, "access denied: insufficient permissions")
		c.Abort()
	}
}
