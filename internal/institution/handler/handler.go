package handler

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/pranavbh-9117/IMB/internal/domain"
	"github.com/pranavbh-9117/IMB/internal/institution/dto"
	"github.com/pranavbh-9117/IMB/internal/institution/service"
	"github.com/pranavbh-9117/IMB/pkg/response"
	"github.com/pranavbh-9117/IMB/pkg/validator"
)

// InstitutionHandler processes HTTP requests for institution management.
type InstitutionHandler struct {
	svc service.InstitutionService
}

// NewInstitutionHandler creates a new InstitutionHandler.
func NewInstitutionHandler(svc service.InstitutionService) *InstitutionHandler {
	return &InstitutionHandler{svc: svc}
}

// Create godoc
// POST /api/v1/institutions
func (h *InstitutionHandler) Create(c *gin.Context) {
	var req dto.CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, validator.FormatBindingError(err))
		return
	}

	inst := &domain.Institution{
		Name:    req.Name,
		Code:    req.Code,
		Address: req.Address,
		Phone:   req.Phone,
		Email:   req.Email,
	}

	if err := h.svc.Create(c.Request.Context(), inst); err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.Created(c, "institution created successfully", h.mapToResponse(inst))
}

// GetByID godoc
// GET /api/v1/institutions/:id
func (h *InstitutionHandler) GetByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "invalid institution ID format")
		return
	}

	inst, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.OK(c, "institution retrieved successfully", h.mapToResponse(inst))
}

// List godoc
// GET /api/v1/institutions
func (h *InstitutionHandler) List(c *gin.Context) {
	// Parse offset, default to 0
	offsetStr := c.DefaultQuery("offset", "0")
	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		response.BadRequest(c, "invalid offset parameter")
		return
	}

	// Parse limit, default to 10
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 0 {
		response.BadRequest(c, "invalid limit parameter")
		return
	}

	institutions, err := h.svc.List(c.Request.Context(), offset, limit)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	// Map domain slice to DTO slice
	res := make([]dto.InstitutionResponse, 0, len(institutions))
	for _, inst := range institutions {
		res = append(res, h.mapToResponse(&inst))
	}

	response.OK(c, "institutions retrieved successfully", res)
}

// Update godoc
// PATCH /api/v1/institutions/:id
func (h *InstitutionHandler) Update(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "invalid institution ID format")
		return
	}

	var req dto.UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, validator.FormatBindingError(err))
		return
	}

	// Fetch existing to apply partial updates
	existing, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	if req.Name != nil {
		existing.Name = *req.Name
	}
	if req.Address != nil {
		existing.Address = *req.Address
	}
	if req.Phone != nil {
		existing.Phone = *req.Phone
	}
	if req.Email != nil {
		existing.Email = *req.Email
	}

	if err := h.svc.Update(c.Request.Context(), id, existing); err != nil {
		h.handleServiceError(c, err)
		return
	}

	// Fetch the saved object to guarantee consistency in the response
	updated, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.OK(c, "institution updated successfully", h.mapToResponse(updated))
}

// Delete godoc
// DELETE /api/v1/institutions/:id
func (h *InstitutionHandler) Delete(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "invalid institution ID format")
		return
	}

	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.OK(c, "institution deleted successfully", nil)
}

// --- private helpers ---

func (h *InstitutionHandler) mapToResponse(inst *domain.Institution) dto.InstitutionResponse {
	return dto.InstitutionResponse{
		ID:        inst.ID.String(),
		Name:      inst.Name,
		Code:      inst.Code,
		Address:   inst.Address,
		Phone:     inst.Phone,
		Email:     inst.Email,
		IsActive:  inst.IsActive,
		CreatedAt: inst.CreatedAt,
	}
}

func (h *InstitutionHandler) handleServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidInput):
		response.BadRequest(c, err.Error())
	case errors.Is(err, service.ErrDuplicateCode):
		response.Conflict(c, err.Error())
	case errors.Is(err, service.ErrInstitutionNotFound):
		response.NotFound(c, err.Error())
	default:
		response.InternalServerError(c)
	}
}
