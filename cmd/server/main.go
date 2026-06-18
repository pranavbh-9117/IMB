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
	"os"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/pranavbh-9117/IMB/docs"
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
	// 1. Load configuration
	cfg, err := config.Load()
	if err != nil {
		os.Exit(1)
	}

	// 2. Initialize global logger
	logger.Init(cfg.App.Env)
	ctx := context.Background()

	// 3. Initialize database
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

	// 4. Run auto-migrations
	if err := migration.Run(db); err != nil {
		logger.Error(ctx, "Migration failed", "error", err)
		os.Exit(1)
	}
	logger.Info(ctx, "Database migration completed")

	// 5. Seed Super Admin
	if err := seed.Run(db, cfg.Seed); err != nil {
		logger.Error(ctx, "Seed failed", "error", err)
		os.Exit(1)
	}
	logger.Info(ctx, "Database seed completed")

	// 5. Instantiate Auth Module
	userRepo := authrepo.NewUserRepository(db)
	tokenRepo := authrepo.NewRefreshTokenRepository(db)
	authSvc := authservice.NewAuthService(userRepo, tokenRepo, cfg.JWT)
	authHandler := authhandler.NewAuthHandler(authSvc, cfg.JWT)

	// 6. Instantiate Institution Module
	institutionRepo := instrepo.NewInstitutionRepository(db)
	institutionSvc := instservice.NewInstitutionService(institutionRepo)
	institutionHandler := insthandler.NewInstitutionHandler(institutionSvc)

	// 7. Instantiate Leave Module
	leaveRepo := leaverepo.NewLeaveRepository(db)
	leaveSvc := leaveservice.NewLeaveService(leaveRepo)
	leaveHandler := leavehandler.NewLeaveHandler(leaveSvc)

	// 8. Instantiate User Module (Injected with LeaveInitializer)
	userManagementRepo := userrepo.NewUserRepository(db)
	userSvc := userservice.NewUserService(userManagementRepo, leaveSvc)
	userHandler := userhandler.NewUserHandler(userSvc)

	// Initialize Gin Router
	r := gin.Default()

	// Mount Request Logger
	r.Use(middleware.RequestLogger())

	// 8. API v1 Group
	v1 := r.Group("/api/v1")

	// 9. Register Swagger UI (Public)
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 10. Instantiate Middlewares
	authMiddleware := middleware.RequireAuth(cfg.JWT.Secret)
	superAdminMiddleware := middleware.RequireRoles(domain.RoleSuperAdmin)

	// 10. Register Auth Routes
	authGroup := v1.Group("/auth")
	authroutes.Register(authGroup, authHandler, authMiddleware)

	// 11. Register Institution Routes (Protected by Auth & Super Admin)
	instGroup := v1.Group("/institutions")
	instGroup.Use(authMiddleware, superAdminMiddleware)
	instroutes.Register(instGroup, institutionHandler)

	// 12. Register User Routes (Protected by Auth & Super/Institute Admin)
	userGroup := v1.Group("/users")
	userGroup.Use(authMiddleware, middleware.RequireRoles(domain.RoleSuperAdmin, domain.RoleInstituteAdmin))
	userroutes.Register(userGroup, userHandler)

	// 13. Register Leave Routes (Protected by Auth)
	leaveGroup := v1.Group("/leaves")
	leaveGroup.Use(authMiddleware)
	leaveroutes.Register(leaveGroup, leaveHandler)



	// 13. Start Server
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
