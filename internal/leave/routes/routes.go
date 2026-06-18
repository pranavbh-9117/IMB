package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/pranavbh-9117/IMB/internal/leave/handler"
)

// Register mounts all leave routes under the provided router group.
// The router group must have the RequireAuth middleware applied prior to registration.
func Register(rg *gin.RouterGroup, h *handler.LeaveHandler) {
	rg.POST("", h.ApplyLeave)
	rg.GET("", h.ListLeaves)
	rg.PUT("/:id", h.ProcessLeave)
	rg.DELETE("/:id", h.CancelLeave)
}
