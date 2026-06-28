// Package handler provides HTTP handlers for the attempt module.
package handler

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/pranavbh-9117/IMB/internal/attempt/dto"
	"github.com/pranavbh-9117/IMB/internal/attempt/service"
	"github.com/pranavbh-9117/IMB/internal/middleware"
	"github.com/pranavbh-9117/IMB/pkg/apperror"
	"github.com/pranavbh-9117/IMB/pkg/response"
	"github.com/pranavbh-9117/IMB/pkg/validator"
)


type AttemptHandler struct {
	svc service.AttemptService
}


func NewAttemptHandler(svc service.AttemptService) *AttemptHandler {
	return &AttemptHandler{svc: svc}
}

func (h *AttemptHandler) handleServiceError(c *gin.Context, err error) {
	var appErr *apperror.AppError
	if errors.As(err, &appErr) {
		switch appErr.Code {
		case 400:
			response.BadRequest(c, appErr.Message)
		case 401:
			response.Unauthorized(c, appErr.Message)
		case 403:
			response.Forbidden(c, appErr.Message)
		case 404:
			response.NotFound(c, appErr.Message)
		case 409:
			response.Conflict(c, appErr.Message)
		default:
			response.InternalServerError(c)
		}
		return
	}
	response.InternalServerError(c)
}

// SubmitAttempt godoc
// @Summary Submit Quiz Attempt
// @Description Submits a quiz attempt. Students only. Quiz must be published and student must belong to the same institution. One attempt per student.
// @Tags Quiz Attempts
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Quiz ID"
// @Param request body dto.SubmitAttemptRequest true "Attempt Details"
// @Success 201 {object} response.SwaggerResponse[dto.SubmitResultResponse] "Attempt Submitted"
// @Failure 400 {object} response.SwaggerErrorResponse "Bad Request"
// @Failure 401 {object} response.SwaggerErrorResponse "Unauthorized"
// @Failure 403 {object} response.SwaggerErrorResponse "Forbidden"
// @Failure 404 {object} response.SwaggerErrorResponse "Not Found"
// @Failure 409 {object} response.SwaggerErrorResponse "Conflict"
// @Failure 500 {object} response.SwaggerErrorResponse "Internal Server Error"
// @Router /quizzes/{id}/attempt [post]
func (h *AttemptHandler) SubmitAttempt(c *gin.Context) {
	idParam := c.Param("id")
	quizID, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "invalid quiz ID format")
		return
	}

	var req dto.SubmitAttemptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, validator.FormatBindingError(err))
		return
	}

	instID, err := middleware.GetInstitutionID(c)
	if err != nil || instID == nil {
		response.Unauthorized(c, "institution ID not found in token")
		return
	}

	userID, err := middleware.GetUserID(c)
	if err != nil {
		response.Unauthorized(c, "user ID not found in token")
		return
	}

	res, err := h.svc.SubmitAttempt(c.Request.Context(), *instID, userID, quizID, &req)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.Created(c, "attempt submitted successfully", res)
}

// GetLeaderboard godoc
// @Summary Get Quiz Leaderboard
// @Description Retrieves the leaderboard rankings for a specific quiz.
// @Tags Quiz Attempts
// @Security BearerAuth
// @Produce json
// @Param id path string true "Quiz ID"
// @Success 200 {object} response.SwaggerResponse[dto.LeaderboardResponse] "Leaderboard Retrieved"
// @Failure 400 {object} response.SwaggerErrorResponse "Bad Request"
// @Failure 401 {object} response.SwaggerErrorResponse "Unauthorized"
// @Failure 403 {object} response.SwaggerErrorResponse "Forbidden"
// @Failure 404 {object} response.SwaggerErrorResponse "Not Found"
// @Failure 500 {object} response.SwaggerErrorResponse "Internal Server Error"
// @Router /quizzes/{id}/leaderboard [get]
func (h *AttemptHandler) GetLeaderboard(c *gin.Context) {
	idParam := c.Param("id")
	quizID, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "invalid quiz ID format")
		return
	}

	institutionID, err := middleware.GetInstitutionID(c)
	if err != nil || institutionID == nil {
		response.Unauthorized(c, "institution ID not found in token")
		return
	}

	res, err := h.svc.GetLeaderboard(c.Request.Context(), *institutionID, quizID)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.OK(c, "leaderboard retrieved successfully", res)
}

// GetStudentResults godoc
// @Summary Get Student Results
// @Description Retrieves all quiz results for the authenticated student.
// @Tags Quiz Attempts
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.SwaggerResponse[[]dto.StudentResultResponse] "Results Retrieved"
// @Failure 401 {object} response.SwaggerErrorResponse "Unauthorized"
// @Failure 403 {object} response.SwaggerErrorResponse "Forbidden"
// @Failure 500 {object} response.SwaggerErrorResponse "Internal Server Error"
// @Router /results [get]
func (h *AttemptHandler) GetStudentResults(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		response.Unauthorized(c, "user ID not found in token")
		return
	}

	res, err := h.svc.GetStudentResults(c.Request.Context(), userID)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.OK(c, "results retrieved successfully", res)
}

// GetQuizResults godoc
// @Summary Get Quiz Results
// @Description Retrieves performance data for a specific quiz. Faculty only. Can only view results for quizzes they created.
// @Tags Quiz Attempts
// @Security BearerAuth
// @Produce json
// @Param id path string true "Quiz ID"
// @Success 200 {object} response.SwaggerResponse[[]dto.FacultyResultResponse] "Quiz Results Retrieved"
// @Failure 401 {object} response.SwaggerErrorResponse "Unauthorized"
// @Failure 403 {object} response.SwaggerErrorResponse "Forbidden"
// @Failure 404 {object} response.SwaggerErrorResponse "Not Found"
// @Failure 500 {object} response.SwaggerErrorResponse "Internal Server Error"
// @Router /quizzes/{id}/results [get]
func (h *AttemptHandler) GetQuizResults(c *gin.Context) {
	idParam := c.Param("id")
	quizID, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "invalid quiz ID format")
		return
	}

	userID, err := middleware.GetUserID(c)
	if err != nil {
		response.Unauthorized(c, "user ID not found in token")
		return
	}

	institutionID, err := middleware.GetInstitutionID(c)
	if err != nil || institutionID == nil {
		response.Unauthorized(c, "institution ID not found in token")
		return
	}

	res, err := h.svc.GetQuizResults(c.Request.Context(), *institutionID, userID, quizID)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.OK(c, "quiz results retrieved successfully", res)
}
