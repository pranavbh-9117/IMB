// Package routes provides routes functionality for the IMB platform.
package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/pranavbh-9117/IMB/internal/user/handler"
)

// Register mounts all user routes under the provided router group.
// The router group must have authentication and appropriate RBAC middlewares
// (RequireRoles(SUPER_ADMIN, INSTITUTE_ADMIN)) applied prior to registration.
func Register(rg *gin.RouterGroup, h *handler.UserHandler) {
	rg.POST("", h.Create)
	rg.GET("", h.List)
	rg.GET("/:id", h.GetByID)
	rg.PUT("/:id", h.Update)
	rg.DELETE("/:id", h.Delete)
}
