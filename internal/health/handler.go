package health

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/pranavbh-9117/IMB/pkg/database"
)

type HealthHandler struct {
	db *gorm.DB
}

func NewHealthHandler(db *gorm.DB) *HealthHandler {
	return &HealthHandler{db: db}
}

type HealthResponse struct {
	Status string             `json:"status"`
	DB     database.PoolStats `json:"db"`
}

func (h *HealthHandler) Health(c *gin.Context) {
	stats, err := database.HealthCheck(h.db)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "error", "error": err.Error()})
		return
	}

	response := HealthResponse{
		Status: "ok",
		DB:     stats,
	}

	c.JSON(http.StatusOK, response)
}
