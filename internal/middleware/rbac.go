// Package middleware provides middleware functionality for the IMB platform.
package middleware

import (
	"github.com/gin-gonic/gin"

	"github.com/pranavbh-9117/IMB/internal/domain"
	"github.com/pranavbh-9117/IMB/pkg/response"
)

// RequireRoles creates a Gin middleware that restricts access to users possessing
// one of the specified roles. It must be mounted after RequireAuth in the router
// chain, as it relies on the user's role being present in the Gin context.
// If the user's role is not found (indicating RequireAuth was omitted) or does
// not match any of the allowed roles, the request is aborted securely.
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
