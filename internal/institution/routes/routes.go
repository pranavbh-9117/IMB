// Package routes provides routes functionality for the IMB platform.
package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/pranavbh-9117/IMB/internal/institution/handler"
)

// Institution routes
func Register(rg *gin.RouterGroup, h *handler.InstitutionHandler) {
	rg.POST("", h.Create)
	rg.GET("", h.List)
	rg.GET("/:id", h.GetByID)
	rg.PATCH("/:id", h.Update)
	rg.DELETE("/:id", h.Delete)
}
