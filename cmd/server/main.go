package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"github.com/pranavbh-9117/IMB/internal/auth/handler"
	"github.com/pranavbh-9117/IMB/internal/auth/repository"
	"github.com/pranavbh-9117/IMB/internal/auth/routes"
	"github.com/pranavbh-9117/IMB/internal/auth/service"
	"github.com/pranavbh-9117/IMB/internal/domain"
	"github.com/pranavbh-9117/IMB/internal/middleware"
	"github.com/pranavbh-9117/IMB/internal/migration"
	"github.com/pranavbh-9117/IMB/internal/seed"
	"github.com/pranavbh-9117/IMB/pkg/config"
	"github.com/pranavbh-9117/IMB/pkg/database"
	"github.com/pranavbh-9117/IMB/pkg/response"
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

	// 5. Instantiate Auth Repositories
	userRepo := repository.NewUserRepository(db)
	tokenRepo := repository.NewRefreshTokenRepository(db)

	// 6. Instantiate Auth Service
	authSvc := service.NewAuthService(userRepo, tokenRepo, cfg.JWT)

	// 7. Instantiate Auth Handler
	authHandler := handler.NewAuthHandler(authSvc, cfg.JWT)

	// 8. Initialize Gin Router
	r := gin.Default()

	// 9. API v1 Group
	v1 := r.Group("/api/v1")

	// 10. Instantiate Auth Middleware
	authMiddleware := middleware.RequireAuth(cfg.JWT.Secret)

	// 11. Register Public Auth Routes
	authGroup := v1.Group("/auth")
	routes.Register(authGroup, authHandler, authMiddleware)

	// 12. Register Protected Test Routes
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