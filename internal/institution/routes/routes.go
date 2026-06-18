package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/pranavbh-9117/IMB/internal/institution/handler"
)

// Register mounts all institution routes under the provided router group.
// The router group (e.g. /api/v1/institutions) should already have the required
// authentication and RBAC middlewares applied by the caller.
func Register(rg *gin.RouterGroup, h *handler.InstitutionHandler) {
	rg.POST("", h.Create)
	rg.GET("", h.List)
	rg.GET("/:id", h.GetByID)
	rg.PATCH("/:id", h.Update)
	rg.DELETE("/:id", h.Delete)
}
