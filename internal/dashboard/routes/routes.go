// Package routes provides routing for the dashboard module.
package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/pranavbh-9117/IMB/internal/dashboard/handler"
)

// Register registers admin dashboard routes.
func Register(rg *gin.RouterGroup, h handler.AdminDashboardHandler) {
	rg.GET("/dashboard", h.GetAdminDashboard)
}
