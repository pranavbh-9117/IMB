// Package routes provides routes functionality for the IMB platform.
package routes

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/pranavbh-9117/IMB/internal/auth/handler"
	"github.com/pranavbh-9117/IMB/internal/middleware"
)

// auth routes
func Register(rg *gin.RouterGroup, h *handler.AuthHandler, authMiddleware gin.HandlerFunc) {
	ctx := context.Background()

	loginStore := middleware.NewRateLimitStore(ctx, time.Minute)
	forgotIPStore := middleware.NewRateLimitStore(ctx, time.Minute)
	forgotEmailStore := middleware.NewRateLimitStore(ctx, time.Minute)

	loginLimiter := middleware.RateLimit(middleware.RateLimitConfig{
		Store:     loginStore,
		Limit:     5,
		Window:    time.Minute,
		Extractor: middleware.ByIP(),
	})

	forgotIPLimiter := middleware.RateLimit(middleware.RateLimitConfig{
		Store:     forgotIPStore,
		Limit:     30,
		Window:    time.Minute,
		Extractor: middleware.ByIP(),
	})

	forgotEmailLimiter := middleware.RateLimit(middleware.RateLimitConfig{
		Store:     forgotEmailStore,
		Limit:     2,
		Window:    time.Minute,
		Extractor: middleware.ByBodyField("email"),
	})

	rg.POST("/login", loginLimiter, h.Login)
	rg.POST("/forgot-password", forgotIPLimiter, forgotEmailLimiter, h.ForgotPassword)
	rg.POST("/reset-password", h.ResetPassword)
	rg.POST("/refresh", h.Refresh)
	rg.POST("/logout", h.Logout)

	rg.GET("/google/login", h.GoogleLogin)
	rg.GET("/google/callback", h.GoogleCallback)

	protected := rg.Group("/")
	protected.Use(authMiddleware)
	protected.POST("/change-password", h.ChangePassword)
}
