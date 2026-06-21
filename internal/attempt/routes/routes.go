// Package routes provides routing functionality for the attempt module.
package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/pranavbh-9117/IMB/internal/attempt/handler"
	"github.com/pranavbh-9117/IMB/internal/domain"
	"github.com/pranavbh-9117/IMB/internal/middleware"
)

// Register mounts all attempt routes. The router groups must have RequireAuth applied.
func Register(quizGroup *gin.RouterGroup, rootGroup *gin.RouterGroup, h *handler.AttemptHandler) {
	studentOnly := middleware.RequireRoles(domain.RoleStudent)
	facultyOnly := middleware.RequireRoles(domain.RoleFaculty)

	// POST /quizzes/:id/attempt (Student Only)
	quizGroup.POST("/:id/attempt", studentOnly, h.SubmitAttempt)

	// GET /quizzes/:id/results (Faculty Only)
	quizGroup.GET("/:id/results", facultyOnly, h.GetQuizResults)

	// GET /results (Student Only)
	rootGroup.GET("/results", studentOnly, h.GetStudentResults)
}
