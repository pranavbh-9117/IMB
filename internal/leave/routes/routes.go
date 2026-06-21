// Package routes provides routes functionality for the IMB platform.
package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/pranavbh-9117/IMB/internal/leave/handler"
)

// Leave routes
func Register(rg *gin.RouterGroup, h *handler.LeaveHandler) {
	rg.POST("", h.ApplyLeave)
	rg.GET("", h.ListLeaves)
	rg.PUT("/:id", h.ProcessLeave)
	rg.DELETE("/:id", h.CancelLeave)
}
