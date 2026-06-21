// Package routes provides routing functionality for the quiz module.
package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/pranavbh-9117/IMB/internal/domain"
	"github.com/pranavbh-9117/IMB/internal/middleware"
	"github.com/pranavbh-9117/IMB/internal/quiz/handler"
)

// Quiz Routes
func Register(rg *gin.RouterGroup, h *handler.QuizHandler) {
	facultyOnly := middleware.RequireRoles(domain.RoleFaculty)
	facultyOrStudent := middleware.RequireRoles(domain.RoleFaculty, domain.RoleStudent)

	// GET /quizzes
	rg.GET("", facultyOrStudent, h.ListQuizzes)

	// GET /quizzes/:id
	rg.GET("/:id", facultyOrStudent, h.GetQuiz)

	// POST /quizzes
	rg.POST("", facultyOnly, h.CreateQuiz)

	// PUT /quizzes/:id
	rg.PUT("/:id", facultyOnly, h.UpdateQuiz)

	// DELETE /quizzes/:id
	rg.DELETE("/:id", facultyOnly, h.DeleteQuiz)

	// PUT /quizzes/:id/publish
	rg.PUT("/:id/publish", facultyOnly, h.PublishQuiz)

	// POST /quizzes/:id/questions
	rg.POST("/:id/questions", facultyOnly, h.CreateQuestion)
}
