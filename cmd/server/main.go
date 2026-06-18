package main

import (
	"log"

	"github.com/gin-gonic/gin"

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
	"github.com/pranavbh-9117/IMB/pkg/config"
	"github.com/pranavbh-9117/IMB/pkg/database"
	"github.com/pranavbh-9117/IMB/pkg/response"
	userhandler "github.com/pranavbh-9117/IMB/internal/user/handler"
	userrepo "github.com/pranavbh-9117/IMB/internal/user/repository"
	userroutes "github.com/pranavbh-9117/IMB/internal/user/routes"
	userservice "github.com/pranavbh-9117/IMB/internal/user/service"
)

func main() {
	// 1. Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// 2. Initialize database
	db, err := database.New(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	if err := database.HealthCheck(db); err != nil {
		log.Fatalf("Database health check failed: %v", err)
	}
	log.Println("Database connected successfully")

	// 3. Run auto-migrations
	if err := migration.Run(db); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}
	log.Println("Database migration completed")

	// 4. Seed Super Admin
	if err := seed.Run(db, cfg.Seed); err != nil {
		log.Fatalf("Seed failed: %v", err)
	}
	log.Println("Database seed completed")

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

	// 7. Initialize Gin Router
	r := gin.Default()

	// 8. API v1 Group
	v1 := r.Group("/api/v1")

	// 9. Instantiate Middlewares
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

	// 14. Register Protected Test Routes
	protected := v1.Group("/")
	protected.Use(authMiddleware)

	protected.GET("/protected", func(c *gin.Context) {
		response.OK(c, "you have accessed a protected route", nil)
	})

	protected.GET("/admin-only", middleware.RequireRoles(domain.RoleSuperAdmin), func(c *gin.Context) {
		response.OK(c, "welcome, super admin", nil)
	})

	protected.GET("/faculty-only", middleware.RequireRoles(domain.RoleFaculty), func(c *gin.Context) {
		response.OK(c, "welcome, faculty", nil)
	})

	protected.GET("/student-only", middleware.RequireRoles(domain.RoleStudent), func(c *gin.Context) {
		response.OK(c, "welcome, student", nil)
	})

	// 13. Start Server
	port := cfg.App.Port
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting server on port %s...", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}