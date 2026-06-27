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

// RegisterFaculty registers faculty dashboard routes.
func RegisterFaculty(rg *gin.RouterGroup, h handler.FacultyDashboardHandler) {
	rg.GET("/dashboard", h.GetFacultyDashboard)
}
