// Package handler provides handler functionality for the IMB platform.
package handler

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/pranavbh-9117/IMB/internal/domain"
	"github.com/pranavbh-9117/IMB/internal/middleware"
	"github.com/pranavbh-9117/IMB/internal/user/dto"
	"github.com/pranavbh-9117/IMB/internal/user/service"
	"github.com/pranavbh-9117/IMB/pkg/response"
	"github.com/pranavbh-9117/IMB/pkg/validator"
)

// UserHandler processes HTTP requests for user management.
type UserHandler struct {
	svc service.UserService
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(svc service.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

// Create godoc
// @Summary Create User
// @Description Creates a new user (Institute Admin, Faculty, or Student). Accessible by Super Admin and Institute Admin.
// @Tags User
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.CreateRequest true "User Details"
// @Success 201 {object} response.SwaggerResponse[dto.CreateResponse] "User Created with Temporary Password"
// @Failure 400 {object} response.SwaggerErrorResponse "Bad Request"
// @Failure 401 {object} response.SwaggerErrorResponse "Unauthorized"
// @Failure 403 {object} response.SwaggerErrorResponse "Forbidden"
// @Failure 409 {object} response.SwaggerErrorResponse "Conflict"
// @Router /users [post]
func (h *UserHandler) Create(c *gin.Context) {
	var req dto.CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, validator.FormatBindingError(err))
		return
	}

	requesterRole, err := middleware.GetRole(c)
	if err != nil {
		response.Unauthorized(c, "unauthorized")
		return
	}

	requesterInstID, err := middleware.GetInstitutionID(c)
	if err != nil {
		response.Unauthorized(c, "unauthorized")
		return
	}

	user := &domain.User{
		Name:  req.Name,
		Email: req.Email,
		Role:  domain.Role(req.Role),
	}

	if req.InstitutionID != nil {
		id, err := uuid.Parse(*req.InstitutionID)
		if err != nil {
			response.BadRequest(c, "invalid institution_id format")
			return
		}
		user.InstitutionID = &id
	}

	result, err := h.svc.Create(c.Request.Context(), requesterRole, requesterInstID, user)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	resp := dto.CreateResponse{
		User:              h.mapToResponse(result.User),
		TemporaryPassword: result.TempPassword,
	}

	response.Created(c, "user created successfully", resp)
}

// GetByID godoc
// @Summary Get User
// @Description Retrieves a user by their UUID.
// @Tags User
// @Security BearerAuth
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} response.SwaggerResponse[dto.UserResponse] "User Retrieved"
// @Failure 400 {object} response.SwaggerErrorResponse "Bad Request"
// @Failure 401 {object} response.SwaggerErrorResponse "Unauthorized"
// @Failure 403 {object} response.SwaggerErrorResponse "Forbidden"
// @Failure 404 {object} response.SwaggerErrorResponse "Not Found"
// @Router /users/{id} [get]
func (h *UserHandler) GetByID(c *gin.Context) {
	idParam := c.Param("id")
	targetID, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "invalid user ID format")
		return
	}

	requesterRole, err := middleware.GetRole(c)
	if err != nil {
		response.Unauthorized(c, "unauthorized")
		return
	}

	requesterInstID, err := middleware.GetInstitutionID(c)
	if err != nil {
		response.Unauthorized(c, "unauthorized")
		return
	}

	user, err := h.svc.GetByID(c.Request.Context(), requesterRole, requesterInstID, targetID)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.OK(c, "user retrieved successfully", h.mapToResponse(user))
}

// List godoc
// @Summary List Users
// @Description Retrieves a paginated list of users within the caller's authorized tenant scope.
// @Tags User
// @Security BearerAuth
// @Produce json
// @Param offset query int false "Pagination offset" default(0)
// @Param limit query int false "Pagination limit" default(10)
// @Success 200 {object} response.SwaggerResponse[[]dto.UserResponse] "Users Retrieved"
// @Failure 400 {object} response.SwaggerErrorResponse "Bad Request"
// @Failure 401 {object} response.SwaggerErrorResponse "Unauthorized"
// @Failure 403 {object} response.SwaggerErrorResponse "Forbidden"
// @Router /users [get]
func (h *UserHandler) List(c *gin.Context) {
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

	requesterRole, err := middleware.GetRole(c)
	if err != nil {
		response.Unauthorized(c, "unauthorized")
		return
	}

	requesterInstID, err := middleware.GetInstitutionID(c)
	if err != nil {
		response.Unauthorized(c, "unauthorized")
		return
	}

	users, err := h.svc.List(c.Request.Context(), requesterRole, requesterInstID, offset, limit)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	res := make([]dto.UserResponse, 0, len(users))
	for _, u := range users {
		res = append(res, h.mapToResponse(&u))
	}

	response.OK(c, "users retrieved successfully", res)
}

// Update godoc
// @Summary Update User
// @Description Partially updates a user's name or email. Roles cannot be changed.
// @Tags User
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param request body dto.UpdateRequest true "Update Payload"
// @Success 200 {object} response.SwaggerResponse[dto.UserResponse] "User Updated"
// @Failure 400 {object} response.SwaggerErrorResponse "Bad Request"
// @Failure 401 {object} response.SwaggerErrorResponse "Unauthorized"
// @Failure 403 {object} response.SwaggerErrorResponse "Forbidden"
// @Failure 404 {object} response.SwaggerErrorResponse "Not Found"
// @Failure 409 {object} response.SwaggerErrorResponse "Conflict"
// @Router /users/{id} [put]
func (h *UserHandler) Update(c *gin.Context) {
	idParam := c.Param("id")
	targetID, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "invalid user ID format")
		return
	}

	var req dto.UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, validator.FormatBindingError(err))
		return
	}

	// ADR-006: Role Immutability
	if req.Role != nil {
		response.BadRequest(c, service.ErrRoleImmutable.Error())
		return
	}

	requesterID, err := middleware.GetUserID(c)
	if err != nil {
		response.Unauthorized(c, "unauthorized")
		return
	}

	requesterRole, err := middleware.GetRole(c)
	if err != nil {
		response.Unauthorized(c, "unauthorized")
		return
	}

	requesterInstID, err := middleware.GetInstitutionID(c)
	if err != nil {
		response.Unauthorized(c, "unauthorized")
		return
	}

	updates := &domain.User{}
	if req.Name != nil {
		updates.Name = *req.Name
	}
	if req.Email != nil {
		updates.Email = *req.Email
	}

	if err := h.svc.Update(c.Request.Context(), requesterID, requesterRole, requesterInstID, targetID, updates); err != nil {
		h.handleServiceError(c, err)
		return
	}

	// Fetch updated state to return to client
	updated, err := h.svc.GetByID(c.Request.Context(), requesterRole, requesterInstID, targetID)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.OK(c, "user updated successfully", h.mapToResponse(updated))
}

// Delete godoc
// @Summary Delete User
// @Description Deactivates a user by UUID. Self-deletion and Super Admin deletion are blocked.
// @Tags User
// @Security BearerAuth
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} response.SwaggerResponse[any] "User Deleted"
// @Failure 400 {object} response.SwaggerErrorResponse "Bad Request"
// @Failure 401 {object} response.SwaggerErrorResponse "Unauthorized"
// @Failure 403 {object} response.SwaggerErrorResponse "Forbidden"
// @Failure 404 {object} response.SwaggerErrorResponse "Not Found"
// @Router /users/{id} [delete]
func (h *UserHandler) Delete(c *gin.Context) {
	idParam := c.Param("id")
	targetID, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "invalid user ID format")
		return
	}

	requesterID, err := middleware.GetUserID(c)
	if err != nil {
		response.Unauthorized(c, "unauthorized")
		return
	}

	requesterRole, err := middleware.GetRole(c)
	if err != nil {
		response.Unauthorized(c, "unauthorized")
		return
	}

	requesterInstID, err := middleware.GetInstitutionID(c)
	if err != nil {
		response.Unauthorized(c, "unauthorized")
		return
	}

	if err := h.svc.Delete(c.Request.Context(), requesterID, requesterRole, requesterInstID, targetID); err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.OK(c, "user deleted successfully", nil)
}

// --- private helpers ---

func (h *UserHandler) mapToResponse(user *domain.User) dto.UserResponse {
	resp := dto.UserResponse{
		ID:        user.ID.String(),
		Name:      user.Name,
		Email:     user.Email,
		Role:      string(user.Role),
		IsActive:  user.IsActive,
		CreatedAt: user.CreatedAt,
	}
	if user.InstitutionID != nil {
		idStr := user.InstitutionID.String()
		resp.InstitutionID = &idStr
	}
	return resp
}

func (h *UserHandler) handleServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidInput):
		response.BadRequest(c, err.Error())
	case errors.Is(err, service.ErrRoleImmutable):
		response.BadRequest(c, err.Error())
	case errors.Is(err, service.ErrSelfManagement):
		response.BadRequest(c, err.Error())
	case errors.Is(err, service.ErrLockoutPrevention):
		response.BadRequest(c, err.Error())
	case errors.Is(err, service.ErrDuplicateEmail):
		response.Conflict(c, err.Error())
	case errors.Is(err, service.ErrUserNotFound):
		response.NotFound(c, err.Error()) // Handled explicitly as 404 per ADR-009
	case errors.Is(err, service.ErrUnauthorized):
		response.Forbidden(c, err.Error())
	case errors.Is(err, service.ErrInvalidRole):
		response.Forbidden(c, err.Error())
	default:
		response.InternalServerError(c)
	}
}
