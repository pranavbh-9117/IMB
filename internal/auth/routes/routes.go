package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/pranavbh-9117/IMB/internal/auth/handler"
)

// Register mounts all authentication routes under the provided router group.
// The caller is responsible for passing the correct group prefix
// (e.g. /api/v1/auth) and the authentication middleware required for protected
// auth routes (like change-password).
func Register(rg *gin.RouterGroup, h *handler.AuthHandler, authMiddleware gin.HandlerFunc) {
	rg.POST("/login", h.Login)
	rg.POST("/refresh", h.Refresh)
	rg.POST("/logout", h.Logout)

	protected := rg.Group("/")
	protected.Use(authMiddleware)
	protected.POST("/change-password", h.ChangePassword)
}
