// Package handler provides handler functionality for the IMB platform.
package handler

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/pranavbh-9117/IMB/internal/domain"
	"github.com/pranavbh-9117/IMB/internal/leave/dto"
	"github.com/pranavbh-9117/IMB/internal/leave/repository"
	"github.com/pranavbh-9117/IMB/internal/leave/service"
	"github.com/pranavbh-9117/IMB/internal/middleware"
	"github.com/pranavbh-9117/IMB/pkg/response"
	"github.com/pranavbh-9117/IMB/pkg/validator"
)

// LeaveHandler processes HTTP requests for leave management.
type LeaveHandler struct {
	svc service.LeaveService
}

// NewLeaveHandler creates a new LeaveHandler.
func NewLeaveHandler(svc service.LeaveService) *LeaveHandler {
	return &LeaveHandler{svc: svc}
}

// ApplyLeave godoc
// @Summary Apply for Leave
// @Description Submits a new leave request. Allowed for Faculty and Students.
// @Tags Leave
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.ApplyLeaveRequest true "Leave Request Details"
// @Success 201 {object} response.SwaggerResponse[dto.LeaveResponse] "Leave Request Submitted"
// @Failure 400 {object} response.SwaggerErrorResponse "Bad Request"
// @Failure 401 {object} response.SwaggerErrorResponse "Unauthorized"
// @Failure 403 {object} response.SwaggerErrorResponse "Forbidden"
// @Failure 409 {object} response.SwaggerErrorResponse "Conflict (Overlap)"
// @Router /leaves [post]
func (h *LeaveHandler) ApplyLeave(c *gin.Context) {
	var req dto.ApplyLeaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, validator.FormatBindingError(err))
		return
	}

	userID, err := middleware.GetUserID(c)
	if err != nil {
		response.Unauthorized(c, "unauthorized")
		return
	}

	role, err := middleware.GetRole(c)
	if err != nil {
		response.Unauthorized(c, "unauthorized")
		return
	}

	instIDPtr, err := middleware.GetInstitutionID(c)
	if err != nil || instIDPtr == nil {
		response.Unauthorized(c, "unauthorized: missing institution")
		return
	}

	domainReq := &domain.LeaveRequest{
		StartDate: req.StartDate,
		EndDate:   req.EndDate,
		Reason:    req.Reason,
	}

	createdReq, err := h.svc.ApplyLeave(c.Request.Context(), userID, role, *instIDPtr, domainReq)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.Created(c, "leave request submitted successfully", h.mapToResponse(createdReq))
}

// ProcessLeave godoc
// @Summary Process Leave Request
// @Description Approves or rejects a leave request. Enforces hierarchical RBAC (Faculty approves Student, Inst Admin approves Faculty).
// @Tags Leave
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Leave Request ID"
// @Param request body dto.ProcessLeaveRequest true "Process Details"
// @Success 200 {object} response.SwaggerResponse[dto.LeaveResponse] "Leave Request Processed"
// @Failure 400 {object} response.SwaggerErrorResponse "Bad Request"
// @Failure 401 {object} response.SwaggerErrorResponse "Unauthorized"
// @Failure 403 {object} response.SwaggerErrorResponse "Forbidden"
// @Failure 404 {object} response.SwaggerErrorResponse "Not Found"
// @Router /leaves/{id} [put]
func (h *LeaveHandler) ProcessLeave(c *gin.Context) {
	idParam := c.Param("id")
	requestID, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "invalid leave request ID")
		return
	}

	var req dto.ProcessLeaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, validator.FormatBindingError(err))
		return
	}

	reviewerID, err := middleware.GetUserID(c)
	if err != nil {
		response.Unauthorized(c, "unauthorized")
		return
	}

	role, err := middleware.GetRole(c)
	if err != nil {
		response.Unauthorized(c, "unauthorized")
		return
	}

	instIDPtr, err := middleware.GetInstitutionID(c)
	if err != nil || instIDPtr == nil {
		response.Unauthorized(c, "unauthorized: missing institution")
		return
	}

	err = h.svc.ProcessLeaveApproval(c.Request.Context(), requestID, reviewerID, role, *instIDPtr, domain.LeaveStatus(req.Status), req.Note)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	updatedReq, err := h.svc.GetLeaveDetails(c.Request.Context(), reviewerID, role, *instIDPtr, requestID)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.OK(c, "leave request processed successfully", h.mapToResponse(updatedReq))
}

// ListLeaves godoc
// @Summary List Leaves
// @Description Retrieves a paginated list of leaves. Pass '?view=approvals' to see leaves pending your review.
// @Tags Leave
// @Security BearerAuth
// @Produce json
// @Param offset query int false "Pagination offset" default(0)
// @Param limit query int false "Pagination limit" default(10)
// @Param view query string false "View mode (approvals or empty)"
// @Param status query string false "Filter by status (pending, approved, rejected, cancelled)"
// @Success 200 {object} response.SwaggerResponse[[]dto.LeaveResponse] "Leaves Retrieved"
// @Failure 400 {object} response.SwaggerErrorResponse "Bad Request"
// @Failure 401 {object} response.SwaggerErrorResponse "Unauthorized"
// @Router /leaves [get]
func (h *LeaveHandler) ListLeaves(c *gin.Context) {
	offsetStr := c.DefaultQuery("offset", "0")
	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		response.BadRequest(c, "invalid offset parameter")
		return
	}

	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 0 {
		response.BadRequest(c, "invalid limit parameter")
		return
	}

	userID, err := middleware.GetUserID(c)
	if err != nil {
		response.Unauthorized(c, "unauthorized")
		return
	}

	role, err := middleware.GetRole(c)
	if err != nil {
		response.Unauthorized(c, "unauthorized")
		return
	}

	instIDPtr, err := middleware.GetInstitutionID(c)
	if err != nil || instIDPtr == nil {
		response.Unauthorized(c, "unauthorized: missing institution")
		return
	}

	var filter repository.RequestFilter

	// Allow approvers to query their subordinates' leaves by passing view=approvals
	view := c.Query("view")
	if view == "approvals" {
		// Service layer ensures Faculty see Student leaves, Institute Admins see Faculty leaves
	} else {
		// Default to viewing own leaves
		filter.UserID = &userID
	}

	if status := c.Query("status"); status != "" {
		st := domain.LeaveStatus(status)
		filter.Status = &st
	}

	leaves, err := h.svc.ListLeaves(c.Request.Context(), userID, role, *instIDPtr, filter, offset, limit)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	res := make([]dto.LeaveResponse, 0, len(leaves))
	for _, l := range leaves {
		res = append(res, h.mapToResponse(&l))
	}

	response.OK(c, "leave requests retrieved successfully", res)
}

// CancelLeave godoc
// @Summary Cancel Leave
// @Description Cancels a pending leave request belonging to the caller.
// @Tags Leave
// @Security BearerAuth
// @Produce json
// @Param id path string true "Leave Request ID"
// @Success 200 {object} response.SwaggerResponse[any] "Leave Cancelled"
// @Failure 400 {object} response.SwaggerErrorResponse "Bad Request"
// @Failure 401 {object} response.SwaggerErrorResponse "Unauthorized"
// @Failure 403 {object} response.SwaggerErrorResponse "Forbidden"
// @Failure 404 {object} response.SwaggerErrorResponse "Not Found"
// @Router /leaves/{id} [delete]
func (h *LeaveHandler) CancelLeave(c *gin.Context) {
	idParam := c.Param("id")
	requestID, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "invalid leave request ID")
		return
	}

	userID, err := middleware.GetUserID(c)
	if err != nil {
		response.Unauthorized(c, "unauthorized")
		return
	}

	if err := h.svc.CancelLeave(c.Request.Context(), userID, requestID); err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.OK(c, "leave request cancelled successfully", nil)
}

// --- Private Helpers ---

func (h *LeaveHandler) mapToResponse(req *domain.LeaveRequest) dto.LeaveResponse {
	res := dto.LeaveResponse{
		ID:            req.ID.String(),
		UserID:        req.UserID.String(),
		InstitutionID: req.InstitutionID.String(),
		StartDate:     req.StartDate,
		EndDate:       req.EndDate,
		Reason:        req.Reason,
		Status:        string(req.Status),
		ReviewNote:    req.ReviewNote,
		CreatedAt:     req.CreatedAt,
	}

	if req.ReviewedBy != nil {
		idStr := req.ReviewedBy.String()
		res.ReviewedBy = &idStr
	}
	if req.ReviewedAt != nil {
		res.ReviewedAt = req.ReviewedAt
	}

	return res
}

func (h *LeaveHandler) handleServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidInput):
		response.BadRequest(c, err.Error())
	case errors.Is(err, service.ErrInsufficientBalance):
		response.BadRequest(c, err.Error())
	case errors.Is(err, service.ErrLeaveNotPending):
		response.BadRequest(c, err.Error())
	case errors.Is(err, service.ErrOverlap):
		response.Conflict(c, err.Error())
	case errors.Is(err, service.ErrRequestNotFound):
		response.NotFound(c, err.Error())
	case errors.Is(err, service.ErrBalanceNotFound):
		response.NotFound(c, err.Error())
	case errors.Is(err, service.ErrUnauthorized):
		response.Forbidden(c, err.Error())
	default:
		response.InternalServerError(c)
	}
}
