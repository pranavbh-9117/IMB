// Package routes provides routes functionality for the IMB platform.
package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/pranavbh-9117/IMB/internal/auth/handler"
)

// auth routes
func Register(rg *gin.RouterGroup, h *handler.AuthHandler, authMiddleware gin.HandlerFunc) {
	rg.POST("/login", h.Login)
	rg.POST("/forgot-password", h.ForgotPassword)
	rg.POST("/reset-password", h.ResetPassword)
	rg.POST("/refresh", h.Refresh)
	rg.POST("/logout", h.Logout)



	rg.GET("/google/login", h.GoogleLogin)
	rg.GET("/google/callback", h.GoogleCallback)

	protected := rg.Group("/")
	protected.Use(authMiddleware)
	protected.POST("/change-password", h.ChangePassword)
}
