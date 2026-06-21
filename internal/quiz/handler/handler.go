// Package handler provides HTTP handlers for the quiz module.
package handler

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/pranavbh-9117/IMB/internal/middleware"
	"github.com/pranavbh-9117/IMB/internal/quiz/dto"
	"github.com/pranavbh-9117/IMB/internal/quiz/service"
	"github.com/pranavbh-9117/IMB/pkg/apperror"
	"github.com/pranavbh-9117/IMB/pkg/response"
	"github.com/pranavbh-9117/IMB/pkg/validator"
)

type QuizHandler struct {
	svc service.QuizService
}


func NewQuizHandler(svc service.QuizService) *QuizHandler {
	return &QuizHandler{svc: svc}
}


func (h *QuizHandler) handleServiceError(c *gin.Context, err error) {
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

// CreateQuiz godoc
// @Summary Create Quiz
// @Description Creates a new quiz. Faculty only.
// @Tags Quiz
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.CreateQuizRequest true "Quiz Details"
// @Success 201 {object} response.SwaggerResponse[dto.QuizResponse] "Quiz Created"
// @Failure 400 {object} response.SwaggerErrorResponse "Bad Request"
// @Failure 401 {object} response.SwaggerErrorResponse "Unauthorized"
// @Failure 403 {object} response.SwaggerErrorResponse "Forbidden"
// @Failure 500 {object} response.SwaggerErrorResponse "Internal Server Error"
// @Router /quizzes [post]
func (h *QuizHandler) CreateQuiz(c *gin.Context) {
	var req dto.CreateQuizRequest
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

	res, err := h.svc.CreateQuiz(c.Request.Context(), *instID, userID, &req)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.Created(c, "quiz created successfully", res)
}

// GetQuiz godoc
// @Summary Get Quiz
// @Description Retrieves a quiz by its ID. Faculty can view their own quizzes. Students can view published quizzes for their institution.
// @Tags Quiz
// @Security BearerAuth
// @Produce json
// @Param id path string true "Quiz ID"
// @Success 200 {object} response.SwaggerResponse[dto.QuizResponse] "Quiz Retrieved"
// @Failure 400 {object} response.SwaggerErrorResponse "Bad Request"
// @Failure 401 {object} response.SwaggerErrorResponse "Unauthorized"
// @Failure 403 {object} response.SwaggerErrorResponse "Forbidden"
// @Failure 404 {object} response.SwaggerErrorResponse "Not Found"
// @Failure 500 {object} response.SwaggerErrorResponse "Internal Server Error"
// @Router /quizzes/{id} [get]
func (h *QuizHandler) GetQuiz(c *gin.Context) {
	idParam := c.Param("id")
	quizID, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "invalid quiz ID format")
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

	role, err := middleware.GetRole(c)
	if err != nil {
		response.Unauthorized(c, "role not found in token")
		return
	}

	res, err := h.svc.GetQuiz(c.Request.Context(), *instID, userID, role, quizID)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}


	if role == "STUDENT" {
		for i := range res.Questions {
			for j := range res.Questions[i].Options {
				res.Questions[i].Options[j].IsCorrect = false 
			}
		}
	}

	response.OK(c, "quiz retrieved successfully", res)
}

// UpdateQuiz godoc
// @Summary Update Quiz
// @Description Updates an existing draft quiz metadata. Faculty only. Published quizzes cannot be modified.
// @Tags Quiz
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Quiz ID"
// @Param request body dto.UpdateQuizRequest true "Update Details"
// @Success 200 {object} response.SwaggerResponse[any] "Quiz Updated"
// @Failure 400 {object} response.SwaggerErrorResponse "Bad Request"
// @Failure 401 {object} response.SwaggerErrorResponse "Unauthorized"
// @Failure 403 {object} response.SwaggerErrorResponse "Forbidden"
// @Failure 404 {object} response.SwaggerErrorResponse "Not Found"
// @Failure 409 {object} response.SwaggerErrorResponse "Conflict"
// @Failure 500 {object} response.SwaggerErrorResponse "Internal Server Error"
// @Router /quizzes/{id} [put]
func (h *QuizHandler) UpdateQuiz(c *gin.Context) {
	idParam := c.Param("id")
	quizID, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "invalid quiz ID format")
		return
	}

	var req dto.UpdateQuizRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, validator.FormatBindingError(err))
		return
	}

	userID, err := middleware.GetUserID(c)
	if err != nil {
		response.Unauthorized(c, "user ID not found in token")
		return
	}

	err = h.svc.UpdateQuiz(c.Request.Context(), userID, quizID, &req)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.OK(c, "quiz updated successfully", nil)
}

// DeleteQuiz godoc
// @Summary Delete Quiz
// @Description Deletes an existing quiz. Faculty only. Quizzes with attempts cannot be deleted.
// @Tags Quiz
// @Security BearerAuth
// @Produce json
// @Param id path string true "Quiz ID"
// @Success 200 {object} response.SwaggerResponse[any] "Quiz Deleted"
// @Failure 400 {object} response.SwaggerErrorResponse "Bad Request"
// @Failure 401 {object} response.SwaggerErrorResponse "Unauthorized"
// @Failure 403 {object} response.SwaggerErrorResponse "Forbidden"
// @Failure 404 {object} response.SwaggerErrorResponse "Not Found"
// @Failure 409 {object} response.SwaggerErrorResponse "Conflict"
// @Failure 500 {object} response.SwaggerErrorResponse "Internal Server Error"
// @Router /quizzes/{id} [delete]
func (h *QuizHandler) DeleteQuiz(c *gin.Context) {
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

	err = h.svc.DeleteQuiz(c.Request.Context(), userID, quizID)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.OK(c, "quiz deleted successfully", nil)
}

// PublishQuiz godoc
// @Summary Publish Quiz
// @Description Publishes a quiz, making it available to students. Faculty only.
// @Tags Quiz
// @Security BearerAuth
// @Produce json
// @Param id path string true "Quiz ID"
// @Success 200 {object} response.SwaggerResponse[any] "Quiz Published"
// @Failure 400 {object} response.SwaggerErrorResponse "Bad Request"
// @Failure 401 {object} response.SwaggerErrorResponse "Unauthorized"
// @Failure 403 {object} response.SwaggerErrorResponse "Forbidden"
// @Failure 404 {object} response.SwaggerErrorResponse "Not Found"
// @Failure 500 {object} response.SwaggerErrorResponse "Internal Server Error"
// @Router /quizzes/{id}/publish [put]
func (h *QuizHandler) PublishQuiz(c *gin.Context) {
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

	err = h.svc.PublishQuiz(c.Request.Context(), userID, quizID)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.OK(c, "quiz published successfully", nil)
}

// ListQuizzes godoc
// @Summary List Quizzes
// @Description Retrieves a list of quizzes. Faculty see their own quizzes. Students see published quizzes for their institution.
// @Tags Quiz
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.SwaggerResponse[[]dto.QuizResponse] "Quizzes Retrieved"
// @Failure 401 {object} response.SwaggerErrorResponse "Unauthorized"
// @Failure 403 {object} response.SwaggerErrorResponse "Forbidden"
// @Failure 500 {object} response.SwaggerErrorResponse "Internal Server Error"
// @Router /quizzes [get]
func (h *QuizHandler) ListQuizzes(c *gin.Context) {
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

	role, err := middleware.GetRole(c)
	if err != nil {
		response.Unauthorized(c, "role not found in token")
		return
	}

	res, err := h.svc.ListQuizzes(c.Request.Context(), *instID, userID, role)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.OK(c, "quizzes retrieved successfully", res)
}

// CreateQuestion godoc
// @Summary Create Question
// @Description Adds a new question with options to a draft quiz. Faculty only.
// @Tags Quiz
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Quiz ID"
// @Param request body dto.CreateQuestionRequest true "Question Details"
// @Success 201 {object} response.SwaggerResponse[dto.QuestionResponse] "Question Created"
// @Failure 400 {object} response.SwaggerErrorResponse "Bad Request"
// @Failure 401 {object} response.SwaggerErrorResponse "Unauthorized"
// @Failure 403 {object} response.SwaggerErrorResponse "Forbidden"
// @Failure 404 {object} response.SwaggerErrorResponse "Not Found"
// @Failure 409 {object} response.SwaggerErrorResponse "Conflict"
// @Failure 500 {object} response.SwaggerErrorResponse "Internal Server Error"
// @Router /quizzes/{id}/questions [post]
func (h *QuizHandler) CreateQuestion(c *gin.Context) {
	idParam := c.Param("id")
	quizID, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "invalid quiz ID format")
		return
	}

	var req dto.CreateQuestionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, validator.FormatBindingError(err))
		return
	}

	userID, err := middleware.GetUserID(c)
	if err != nil {
		response.Unauthorized(c, "user ID not found in token")
		return
	}

	res, err := h.svc.CreateQuestion(c.Request.Context(), userID, quizID, &req)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.Created(c, "question created successfully", res)
}
