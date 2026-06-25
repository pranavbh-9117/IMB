package app

import (
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/pranavbh-9117/IMB/docs"
	attempthandler "github.com/pranavbh-9117/IMB/internal/attempt/handler"
	attemptroutes "github.com/pranavbh-9117/IMB/internal/attempt/routes"
	authhandler "github.com/pranavbh-9117/IMB/internal/auth/handler"
	authroutes "github.com/pranavbh-9117/IMB/internal/auth/routes"
	"github.com/pranavbh-9117/IMB/internal/domain"
	"github.com/pranavbh-9117/IMB/internal/health"
	insthandler "github.com/pranavbh-9117/IMB/internal/institution/handler"
	instroutes "github.com/pranavbh-9117/IMB/internal/institution/routes"
	leavehandler "github.com/pranavbh-9117/IMB/internal/leave/handler"
	leaveroutes "github.com/pranavbh-9117/IMB/internal/leave/routes"
	"github.com/pranavbh-9117/IMB/internal/middleware"
	quizhandler "github.com/pranavbh-9117/IMB/internal/quiz/handler"
	quizroutes "github.com/pranavbh-9117/IMB/internal/quiz/routes"
	userhandler "github.com/pranavbh-9117/IMB/internal/user/handler"
	userroutes "github.com/pranavbh-9117/IMB/internal/user/routes"
)

func (a *App) setupRoutes(
	healthHandler *health.HealthHandler,
	authHandler *authhandler.AuthHandler,
	institutionHandler *insthandler.InstitutionHandler,
	leaveHandler *leavehandler.LeaveHandler,
	userHandler *userhandler.UserHandler,
	quizHandler *quizhandler.QuizHandler,
	attemptHandler *attempthandler.AttemptHandler,
) {
	r := a.router
	cfg := a.cfg

	// Request Logger
	r.Use(middleware.RequestLogger())

	// API v1 Group
	v1 := r.Group("/api/v1")

	// Swagger UI
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Health Check
	r.GET("/health", healthHandler.Health)

	// Middlewares
	authMiddleware := middleware.RequireAuth(cfg.JWT.Secret)
	superAdminMiddleware := middleware.RequireRoles(domain.RoleSuperAdmin)

	// Auth Routes
	authGroup := v1.Group("/auth")
	authroutes.Register(authGroup, authHandler, authMiddleware)

	// Institution Routes
	instGroup := v1.Group("/institutions")
	instGroup.Use(authMiddleware, superAdminMiddleware)
	instroutes.Register(instGroup, institutionHandler)

	// User Routes
	userGroup := v1.Group("/users")
	userGroup.Use(authMiddleware, middleware.RequireRoles(domain.RoleSuperAdmin, domain.RoleInstituteAdmin))
	userroutes.Register(userGroup, userHandler)

	// Register Leave Routes
	leaveGroup := v1.Group("/leaves")
	leaveGroup.Use(authMiddleware)
	leaveroutes.Register(leaveGroup, leaveHandler)

	// Quiz Routes
	quizGroup := v1.Group("/quizzes")
	quizGroup.Use(authMiddleware)
	quizroutes.Register(quizGroup, quizHandler)

	// Quiz Attempt Routes
	attemptRootGroup := v1.Group("")
	attemptRootGroup.Use(authMiddleware)
	attemptroutes.Register(quizGroup, attemptRootGroup, attemptHandler)
}
