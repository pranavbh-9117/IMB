package handler

import (
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/pranavbh-9117/IMB/internal/middleware"
	_ "github.com/pranavbh-9117/IMB/internal/quiz/dto"
	"github.com/pranavbh-9117/IMB/pkg/response"
	"github.com/pranavbh-9117/IMB/pkg/validator"
)

// UploadMaterials godoc
// @Summary Upload Quiz Materials
// @Description Uploads supplementary materials (PDF, DOCX, Images) for a draft quiz. Faculty only. Max 20MB per file, max 10 files per quiz.
// @Tags Quiz Materials
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param id path string true "Quiz ID"
// @Param materials formData file true "Files to upload"
// @Success 201 {object} response.SwaggerResponse[dto.UploadMaterialsResponse] "Materials Uploaded"
// @Failure 400 {object} response.SwaggerErrorResponse "Bad Request"
// @Failure 401 {object} response.SwaggerErrorResponse "Unauthorized"
// @Failure 403 {object} response.SwaggerErrorResponse "Forbidden"
// @Failure 409 {object} response.SwaggerErrorResponse "Conflict"
// @Failure 500 {object} response.SwaggerErrorResponse "Internal Server Error"
// @Router /quizzes/{id}/materials [post]
func (h *QuizHandler) UploadMaterials(c *gin.Context) {
	idParam := c.Param("id")
	quizID, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "invalid quiz ID format")
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 200<<20)

	form, err := c.MultipartForm()
	if err != nil {
		response.BadRequest(c, validator.FormatBindingError(err))
		return
	}

	files := form.File["materials"]
	if len(files) == 0 {
		response.BadRequest(c, "no materials provided in the request")
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

	res, err := h.materialSvc.UploadMaterials(c.Request.Context(), *instID, userID, quizID, files)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.Created(c, "materials uploaded successfully", res)
}

// ListMaterials godoc
// @Summary List Quiz Materials
// @Description Retrieves metadata for all materials attached to a quiz.
// @Tags Quiz Materials
// @Security BearerAuth
// @Produce json
// @Param id path string true "Quiz ID"
// @Success 200 {object} response.SwaggerResponse[[]dto.MaterialResponse] "Materials Retrieved"
// @Failure 400 {object} response.SwaggerErrorResponse "Bad Request"
// @Failure 401 {object} response.SwaggerErrorResponse "Unauthorized"
// @Failure 403 {object} response.SwaggerErrorResponse "Forbidden"
// @Failure 500 {object} response.SwaggerErrorResponse "Internal Server Error"
// @Router /quizzes/{id}/materials [get]
func (h *QuizHandler) ListMaterials(c *gin.Context) {
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

	res, err := h.materialSvc.ListMaterials(c.Request.Context(), *instID, userID, role, quizID)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.OK(c, "materials retrieved successfully", res)
}

// DownloadMaterial godoc
// @Summary Download Quiz Material
// @Description Downloads a specific quiz material file.
// @Tags Quiz Materials
// @Security BearerAuth
// @Produce application/octet-stream
// @Param id path string true "Quiz ID"
// @Param materialId path string true "Material ID"
// @Success 200 {file} file "File downloaded"
// @Failure 400 {object} response.SwaggerErrorResponse "Bad Request"
// @Failure 401 {object} response.SwaggerErrorResponse "Unauthorized"
// @Failure 403 {object} response.SwaggerErrorResponse "Forbidden"
// @Failure 404 {object} response.SwaggerErrorResponse "Not Found"
// @Failure 500 {object} response.SwaggerErrorResponse "Internal Server Error"
// @Router /quizzes/{id}/materials/{materialId} [get]
func (h *QuizHandler) DownloadMaterial(c *gin.Context) {
	idParam := c.Param("id")
	quizID, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "invalid quiz ID format")
		return
	}

	matIDParam := c.Param("materialId")
	materialID, err := uuid.Parse(matIDParam)
	if err != nil {
		response.BadRequest(c, "invalid material ID format")
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

	res, err := h.materialSvc.DownloadMaterial(c.Request.Context(), *instID, userID, role, quizID, materialID)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}
	defer res.Reader.Close()

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", res.Filename))
	c.Header("Content-Type", res.ContentType)
	c.Header("Content-Length", fmt.Sprintf("%d", res.Size))
	c.Status(http.StatusOK)

	_, _ = io.Copy(c.Writer, res.Reader)
}
