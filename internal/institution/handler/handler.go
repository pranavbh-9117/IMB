// Package handler provides handler functionality for the IMB platform.
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

type InstitutionHandler struct {
	svc service.InstitutionService
}

func NewInstitutionHandler(svc service.InstitutionService) *InstitutionHandler {
	return &InstitutionHandler{svc: svc}
}

// Create godoc
// @Summary Create Institution
// @Description Creates a new institution. Accessible only by Super Admin.
// @Tags Institution
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.CreateRequest true "Institution Details"
// @Success 201 {object} response.SwaggerResponse[dto.InstitutionResponse] "Institution Created"
// @Failure 400 {object} response.SwaggerErrorResponse "Bad Request"
// @Failure 401 {object} response.SwaggerErrorResponse "Unauthorized"
// @Failure 403 {object} response.SwaggerErrorResponse "Forbidden"
// @Failure 409 {object} response.SwaggerErrorResponse "Conflict"
// @Router /institutions [post]
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
// @Summary Get Institution
// @Description Retrieves an institution by its UUID. Accessible only by Super Admin.
// @Tags Institution
// @Security BearerAuth
// @Produce json
// @Param id path string true "Institution ID"
// @Success 200 {object} response.SwaggerResponse[dto.InstitutionResponse] "Institution Retrieved"
// @Failure 400 {object} response.SwaggerErrorResponse "Bad Request"
// @Failure 401 {object} response.SwaggerErrorResponse "Unauthorized"
// @Failure 403 {object} response.SwaggerErrorResponse "Forbidden"
// @Failure 404 {object} response.SwaggerErrorResponse "Not Found"
// @Router /institutions/{id} [get]
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
// @Summary List Institutions
// @Description Retrieves a paginated list of institutions. Accessible only by Super Admin.
// @Tags Institution
// @Security BearerAuth
// @Produce json
// @Param offset query int false "Pagination offset" default(0)
// @Param limit query int false "Pagination limit" default(10)
// @Success 200 {object} response.SwaggerResponse[[]dto.InstitutionResponse] "Institutions Retrieved"
// @Failure 400 {object} response.SwaggerErrorResponse "Bad Request"
// @Failure 401 {object} response.SwaggerErrorResponse "Unauthorized"
// @Failure 403 {object} response.SwaggerErrorResponse "Forbidden"
// @Router /institutions [get]
func (h *InstitutionHandler) List(c *gin.Context) {
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

	institutions, err := h.svc.List(c.Request.Context(), offset, limit)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	res := make([]dto.InstitutionResponse, 0, len(institutions))
	for _, inst := range institutions {
		res = append(res, h.mapToResponse(&inst))
	}

	response.OK(c, "institutions retrieved successfully", res)
}

// Update godoc
// @Summary Update Institution
// @Description Partially updates an institution. Accessible only by Super Admin.
// @Tags Institution
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Institution ID"
// @Param request body dto.UpdateRequest true "Update Payload"
// @Success 200 {object} response.SwaggerResponse[dto.InstitutionResponse] "Institution Updated"
// @Failure 400 {object} response.SwaggerErrorResponse "Bad Request"
// @Failure 401 {object} response.SwaggerErrorResponse "Unauthorized"
// @Failure 403 {object} response.SwaggerErrorResponse "Forbidden"
// @Failure 404 {object} response.SwaggerErrorResponse "Not Found"
// @Router /institutions/{id} [patch]
func (h *InstitutionHandler) Update(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "invalid institution ID format")
		return
	}

	var req dto.UpdateInstitutionInput
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, validator.FormatBindingError(err))
		return
	}

	updated, err := h.svc.Update(c.Request.Context(), id, &req)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.OK(c, "institution updated successfully", h.mapToResponse(updated))
}

// Delete godoc
// @Summary Delete Institution
// @Description Soft-deletes an institution by its UUID. Accessible only by Super Admin.
// @Tags Institution
// @Security BearerAuth
// @Produce json
// @Param id path string true "Institution ID"
// @Success 200 {object} response.SwaggerResponse[any] "Institution Deleted"
// @Failure 400 {object} response.SwaggerErrorResponse "Bad Request"
// @Failure 401 {object} response.SwaggerErrorResponse "Unauthorized"
// @Failure 403 {object} response.SwaggerErrorResponse "Forbidden"
// @Failure 404 {object} response.SwaggerErrorResponse "Not Found"
// @Router /institutions/{id} [delete]
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
