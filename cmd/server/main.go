// @title Institute Management Backend API
// @version 1.0
// @description API documentation for the Institute Management Platform.
// @host localhost:8080
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

package main

import (
	"context"
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/pranavbh-9117/IMB/docs"
	attempthandler "github.com/pranavbh-9117/IMB/internal/attempt/handler"
	attemptrepo "github.com/pranavbh-9117/IMB/internal/attempt/repository"
	attemptroutes "github.com/pranavbh-9117/IMB/internal/attempt/routes"
	attemptservice "github.com/pranavbh-9117/IMB/internal/attempt/service"
	authhandler "github.com/pranavbh-9117/IMB/internal/auth/handler"
	authrepo "github.com/pranavbh-9117/IMB/internal/auth/repository"
	authroutes "github.com/pranavbh-9117/IMB/internal/auth/routes"
	authservice "github.com/pranavbh-9117/IMB/internal/auth/service"
	"github.com/pranavbh-9117/IMB/internal/domain"
	insthandler "github.com/pranavbh-9117/IMB/internal/institution/handler"
	instrepo "github.com/pranavbh-9117/IMB/internal/institution/repository"
	instroutes "github.com/pranavbh-9117/IMB/internal/institution/routes"
	instservice "github.com/pranavbh-9117/IMB/internal/institution/service"
	leavehandler "github.com/pranavbh-9117/IMB/internal/leave/handler"
	leaverepo "github.com/pranavbh-9117/IMB/internal/leave/repository"
	leaveroutes "github.com/pranavbh-9117/IMB/internal/leave/routes"
	leaveservice "github.com/pranavbh-9117/IMB/internal/leave/service"
	"github.com/pranavbh-9117/IMB/internal/middleware"
	"github.com/pranavbh-9117/IMB/internal/migration"
	quizhandler "github.com/pranavbh-9117/IMB/internal/quiz/handler"
	quizrepo "github.com/pranavbh-9117/IMB/internal/quiz/repository"
	quizroutes "github.com/pranavbh-9117/IMB/internal/quiz/routes"
	quizservice "github.com/pranavbh-9117/IMB/internal/quiz/service"
	"github.com/pranavbh-9117/IMB/internal/seed"
	userhandler "github.com/pranavbh-9117/IMB/internal/user/handler"
	userrepo "github.com/pranavbh-9117/IMB/internal/user/repository"
	userroutes "github.com/pranavbh-9117/IMB/internal/user/routes"
	userservice "github.com/pranavbh-9117/IMB/internal/user/service"
	"github.com/pranavbh-9117/IMB/pkg/config"
	"github.com/pranavbh-9117/IMB/pkg/database"
	"github.com/pranavbh-9117/IMB/pkg/logger"
)

func main() {
	//Loading configurations
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	//Initializing logger
	logger.Init(cfg.App.Env)

	//Create context
	ctx := context.Background()

	//Initializing DB
	db, err := database.New(cfg.Database)
	if err != nil {
		logger.Error(ctx, "Failed to initialize database", "error", err)
		os.Exit(1)
	}

	if err := database.HealthCheck(db); err != nil {
		logger.Error(ctx, "Database health check failed", "error", err)
		os.Exit(1)
	}
	logger.Info(ctx, "Database connected successfully")

	//DB Migrations
	if err := migration.Run(db); err != nil {
		logger.Error(ctx, "Migration failed", "error", err)
		os.Exit(1)
	}
	logger.Info(ctx, "Database migration completed")

	//HardSeeding Super Admin
	if err := seed.Run(db, cfg.Seed); err != nil {
		logger.Error(ctx, "Seed failed", "error", err)
		os.Exit(1)
	}
	logger.Info(ctx, "Database seed completed")

	//Auth Module
	userRepo := authrepo.NewUserRepository(db)
	tokenRepo := authrepo.NewRefreshTokenRepository(db)
	authSvc := authservice.NewAuthService(userRepo, tokenRepo, cfg.JWT, cfg.OAuth)
	authHandler := authhandler.NewAuthHandler(authSvc, cfg.JWT)

	//Institution Module
	institutionRepo := instrepo.NewInstitutionRepository(db)
	institutionSvc := instservice.NewInstitutionService(institutionRepo)
	institutionHandler := insthandler.NewInstitutionHandler(institutionSvc)

	//Leave Module
	leaveRepo := leaverepo.NewLeaveRepository(db)
	leaveSvc := leaveservice.NewLeaveService(leaveRepo)
	leaveHandler := leavehandler.NewLeaveHandler(leaveSvc)

	//User Module
	userManagementRepo := userrepo.NewUserRepository(db)
	userSvc := userservice.NewUserService(userManagementRepo, leaveSvc)
	userHandler := userhandler.NewUserHandler(userSvc)

	//Quiz Module
	quizRepo := quizrepo.NewQuizRepository(db)
	quizSvc := quizservice.NewQuizService(quizRepo)
	quizHandler := quizhandler.NewQuizHandler(quizSvc)

	//Quiz Attempt Module
	attemptRepo := attemptrepo.NewAttemptRepository(db)
	attemptSvc := attemptservice.NewAttemptService(attemptRepo, quizSvc)
	attemptHandler := attempthandler.NewAttemptHandler(attemptSvc)

	//Gin Router
	r := gin.New()
	r.Use(gin.Recovery())

	//Request Logger
	r.Use(middleware.RequestLogger())

	//API v1 Group
	v1 := r.Group("/api/v1")

	//Swagger UI
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	//Middlewares
	authMiddleware := middleware.RequireAuth(cfg.JWT.Secret)
	superAdminMiddleware := middleware.RequireRoles(domain.RoleSuperAdmin)

	//Auth Routes
	authGroup := v1.Group("/auth")
	authroutes.Register(authGroup, authHandler, authMiddleware)

	//Institution Routes
	instGroup := v1.Group("/institutions")
	instGroup.Use(authMiddleware, superAdminMiddleware)
	instroutes.Register(instGroup, institutionHandler)

	//User Routes
	userGroup := v1.Group("/users")
	userGroup.Use(authMiddleware, middleware.RequireRoles(domain.RoleSuperAdmin, domain.RoleInstituteAdmin))
	userroutes.Register(userGroup, userHandler)

	//Register Leave Routes
	leaveGroup := v1.Group("/leaves")
	leaveGroup.Use(authMiddleware)
	leaveroutes.Register(leaveGroup, leaveHandler)

	//Quiz Routes
	quizGroup := v1.Group("/quizzes")
	quizGroup.Use(authMiddleware)
	quizroutes.Register(quizGroup, quizHandler)

	//Quiz Attempt Routes
	attemptRootGroup := v1.Group("")
	attemptRootGroup.Use(authMiddleware)
	attemptroutes.Register(quizGroup, attemptRootGroup, attemptHandler)

	//Server Setup
	port := cfg.App.Port
	if port == "" {
		port = "8080"
	}

	logger.Info(ctx, "Starting server", "port", port)
	if err := r.Run(":" + port); err != nil {
		logger.Error(ctx, "Failed to start server", "error", err)
		os.Exit(1)
	}
}
