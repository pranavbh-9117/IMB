// Package handler implements HTTP request handlers for the dashboard module.
package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	_ "github.com/pranavbh-9117/IMB/internal/dashboard/dto"
	"github.com/pranavbh-9117/IMB/internal/dashboard/service"
	"github.com/pranavbh-9117/IMB/internal/middleware"
	"github.com/pranavbh-9117/IMB/pkg/cache"
	"github.com/pranavbh-9117/IMB/pkg/logger"
	"github.com/pranavbh-9117/IMB/pkg/response"
)

// AdminDashboardHandler handles HTTP requests for admin dashboard endpoints.
type AdminDashboardHandler interface {
	GetAdminDashboard(c *gin.Context)
}

type adminDashboardHandler struct {
	svc   service.AdminDashboardService
	cache cache.CacheClient
	ttl   time.Duration
}

// NewAdminDashboardHandler initializes an AdminDashboardHandler with cache-aside capability.
func NewAdminDashboardHandler(svc service.AdminDashboardService, cacheClient cache.CacheClient, ttl time.Duration) AdminDashboardHandler {
	return &adminDashboardHandler{
		svc:   svc,
		cache: cacheClient,
		ttl:   ttl,
	}
}

type envelope struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

// GetAdminDashboard godoc
// @Summary Get Admin Dashboard
// @Description Retrieves aggregated dashboard metrics (student count, faculty count, quiz count, leave statistics) for the admin's institution.
// @Tags Dashboard
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.SwaggerResponse[dto.AdminDashboardData] "Dashboard Retrieved"
// @Header 200 {string} X-Cache "HIT or MISS"
// @Header 200 {string} X-Cache-TTL "Configured cache TTL in seconds"
// @Failure 401 {object} response.SwaggerErrorResponse "Unauthorized"
// @Failure 403 {object} response.SwaggerErrorResponse "Forbidden"
// @Failure 500 {object} response.SwaggerErrorResponse "Internal Server Error"
// @Router /admin/dashboard [get]
func (h *adminDashboardHandler) GetAdminDashboard(c *gin.Context) {
	instID, err := middleware.GetInstitutionID(c)
	if err != nil || instID == nil {
		response.Unauthorized(c, "unauthorized institution access")
		return
	}

	ctx := c.Request.Context()
	cacheKey := cache.AdminDashboardKey(*instID)
	ttlHeaderVal := fmt.Sprintf("%d", int(h.ttl.Seconds()))

	// Try reading from cache
	cachedBytes, getErr := h.cache.Get(ctx, cacheKey)
	if getErr == nil {
		c.Header("X-Cache", "HIT")
		c.Header("X-Cache-TTL", ttlHeaderVal)
		c.Data(http.StatusOK, "application/json", cachedBytes)
		return
	}

	if !errors.Is(getErr, cache.ErrCacheMiss) {
		logger.Warn(ctx, "admin dashboard: cache read failed", "error", getErr)
	}

	// Cache miss or read error -> fetch from database
	data, svcErr := h.svc.GetAdminDashboard(ctx, *instID)
	if svcErr != nil {
		logger.Error(ctx, "admin dashboard: aggregation failed", "error", svcErr)
		response.InternalServerError(c)
		return
	}

	payload := envelope{
		Success: true,
		Message: "Dashboard fetched successfully",
		Data:    data,
	}
	jsonBytes, marshErr := json.Marshal(payload)
	if marshErr != nil {
		logger.Error(ctx, "admin dashboard: serialization failed", "error", marshErr)
		response.InternalServerError(c)
		return
	}

	// Synchronous non-fatal cache write
	if setErr := h.cache.Set(ctx, cacheKey, jsonBytes, h.ttl); setErr != nil {
		logger.Warn(ctx, "admin dashboard: cache write failed", "error", setErr)
	}

	c.Header("X-Cache", "MISS")
	c.Header("X-Cache-TTL", ttlHeaderVal)
	c.Data(http.StatusOK, "application/json", jsonBytes)
}
