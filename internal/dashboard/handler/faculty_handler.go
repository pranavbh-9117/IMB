// Package handler implements HTTP request handlers for the dashboard module.
package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	_ "github.com/pranavbh-9117/IMB/internal/dashboard/dto"
	"github.com/pranavbh-9117/IMB/internal/dashboard/service"
	"github.com/pranavbh-9117/IMB/internal/middleware"
	"github.com/pranavbh-9117/IMB/pkg/logger"
	"github.com/pranavbh-9117/IMB/pkg/response"
)

// FacultyDashboardHandler handles HTTP requests for faculty dashboard endpoints.
type FacultyDashboardHandler interface {
	GetFacultyDashboard(c *gin.Context)
}

type facultyDashboardHandler struct {
	svc service.FacultyDashboardService
}

// NewFacultyDashboardHandler initializes a FacultyDashboardHandler.
func NewFacultyDashboardHandler(svc service.FacultyDashboardService) FacultyDashboardHandler {
	return &facultyDashboardHandler{svc: svc}
}

// GetFacultyDashboard godoc
// @Summary Get Faculty Dashboard
// @Description Retrieves aggregated dashboard metrics for the logged-in faculty member.
// @Tags Dashboard
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.SwaggerResponse[dto.FacultyDashboardData] "Dashboard Retrieved"
// @Failure 401 {object} response.SwaggerErrorResponse "Unauthorized"
// @Failure 403 {object} response.SwaggerErrorResponse "Forbidden"
// @Failure 500 {object} response.SwaggerErrorResponse "Internal Server Error"
// @Failure 504 {object} response.SwaggerErrorResponse "Gateway Timeout"
// @Router /faculty/dashboard [get]
func (h *facultyDashboardHandler) GetFacultyDashboard(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		response.Unauthorized(c, "unauthorized access")
		return
	}

	instID, err := middleware.GetInstitutionID(c)
	if err != nil || instID == nil {
		response.Unauthorized(c, "unauthorized institution access")
		return
	}

	ctx := c.Request.Context()
	data, err := h.svc.GetFacultyDashboard(ctx, userID, *instID)
	if err != nil {
		logger.Error(ctx, "faculty dashboard: aggregation failed", "error", err)
		if strings.Contains(err.Error(), "timeout") {
			c.JSON(http.StatusGatewayTimeout, gin.H{
				"success": false,
				"message": "gateway timeout fetching dashboard data",
				"data":    nil,
			})
			return
		}
		response.InternalServerError(c)
		return
	}

	response.OK(c, "Dashboard fetched successfully", data)
}
