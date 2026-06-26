package app

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	attempthandler "github.com/pranavbh-9117/IMB/internal/attempt/handler"
	attemptrepo "github.com/pranavbh-9117/IMB/internal/attempt/repository"
	attemptservice "github.com/pranavbh-9117/IMB/internal/attempt/service"
	authhandler "github.com/pranavbh-9117/IMB/internal/auth/handler"
	authrepo "github.com/pranavbh-9117/IMB/internal/auth/repository"
	authservice "github.com/pranavbh-9117/IMB/internal/auth/service"
	"github.com/pranavbh-9117/IMB/internal/health"
	insthandler "github.com/pranavbh-9117/IMB/internal/institution/handler"
	instrepo "github.com/pranavbh-9117/IMB/internal/institution/repository"
	instservice "github.com/pranavbh-9117/IMB/internal/institution/service"
	leavehandler "github.com/pranavbh-9117/IMB/internal/leave/handler"
	leaverepo "github.com/pranavbh-9117/IMB/internal/leave/repository"
	leaveservice "github.com/pranavbh-9117/IMB/internal/leave/service"
	quizhandler "github.com/pranavbh-9117/IMB/internal/quiz/handler"
	quizrepo "github.com/pranavbh-9117/IMB/internal/quiz/repository"
	quizservice "github.com/pranavbh-9117/IMB/internal/quiz/service"
	"github.com/pranavbh-9117/IMB/internal/seed"
	userhandler "github.com/pranavbh-9117/IMB/internal/user/handler"
	userrepo "github.com/pranavbh-9117/IMB/internal/user/repository"
	userservice "github.com/pranavbh-9117/IMB/internal/user/service"
	"github.com/pranavbh-9117/IMB/internal/workerpool"
	"github.com/pranavbh-9117/IMB/pkg/config"
	"github.com/pranavbh-9117/IMB/pkg/database"
	"github.com/pranavbh-9117/IMB/pkg/email"
	"github.com/pranavbh-9117/IMB/pkg/logger"
)

type App struct {
	router *gin.Engine
	cfg    *config.Config
	db     *gorm.DB
	pool   workerpool.Pool
}

func NewApp() (*App, error) {
	// Loading configurations
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Initializing logger
	logger.Init(cfg.App.Env)

	ctx := context.Background()

	// Initializing DB
	db, err := database.New(cfg.Database)
	if err != nil {
		logger.Error(ctx, "Failed to initialize database", "error", err)
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	if _, err := database.HealthCheck(db); err != nil {
		logger.Error(ctx, "Database health check failed", "error", err)
		return nil, fmt.Errorf("database health check failed: %w", err)
	}
	logger.Info(ctx, "Database connected successfully")

	logger.Info(ctx, "Database connection pool initialized",
		"max_open", cfg.Database.MaxOpenConns,
		"max_idle", cfg.Database.MaxIdleConns,
		"conn_max_lifetime", cfg.Database.ConnMaxLifetime.String(),
		"conn_max_idle_time", cfg.Database.ConnMaxIdleTime.String(),
	)

	// HardSeeding Super Admin
	if err := seed.Run(db, cfg.Seed); err != nil {
		logger.Error(ctx, "Seed failed", "error", err)
		return nil, fmt.Errorf("seed failed: %w", err)
	}
	logger.Info(ctx, "Database seed completed")

	// Initialize Application Singleton Worker Pool
	wp, err := workerpool.New(workerpool.Options{
		WorkersCount: 10,
		QueueSize:    100,
		Logger:       slogAdapter{},
	})
	if err != nil {
		logger.Error(ctx, "Failed to initialize worker pool", "error", err)
		return nil, fmt.Errorf("failed to initialize worker pool: %w", err)
	}

	app := &App{
		router: gin.New(),
		cfg:    cfg,
		db:     db,
		pool:   wp,
	}

	app.router.Use(gin.Recovery())

	// Initialize dependencies and routes
	app.setupDependencies()

	return app, nil
}

func (a *App) Start() error {
	port := a.cfg.App.Port
	if port == "" {
		port = "8080"
	}

	logger.Info(context.Background(), "Starting server", "port", port)
	if err := a.router.Run(":" + port); err != nil {
		logger.Error(context.Background(), "Failed to start server", "error", err)
		return err
	}
	return nil
}

func (a *App) setupDependencies() {
	db := a.db
	cfg := a.cfg

	// Email Subsystem
	emailSvc := email.NewMailSender(cfg.SMTP)

	// Auth Module
	userRepo := authrepo.NewUserRepository(db)
	tokenRepo := authrepo.NewRefreshTokenRepository(db)
	resetRepo := authrepo.NewPasswordResetTokenRepository(db)
	authSvc := authservice.NewAuthService(userRepo, tokenRepo, resetRepo, cfg.JWT, cfg.OAuth, emailSvc)
	authHandler := authhandler.NewAuthHandler(authSvc, cfg.JWT)


	// Health Module
	healthHandler := health.NewHealthHandler(db)

	// Institution Module
	institutionRepo := instrepo.NewInstitutionRepository(db)
	institutionSvc := instservice.NewInstitutionService(institutionRepo)
	institutionHandler := insthandler.NewInstitutionHandler(institutionSvc)


	// Leave Module
	leaveRepo := leaverepo.NewLeaveRepository(db)
	leaveSvc := leaveservice.NewLeaveService(leaveRepo, emailSvc)
	leaveHandler := leavehandler.NewLeaveHandler(leaveSvc)

	// User Module
	userManagementRepo := userrepo.NewUserRepository(db)
	userSvc := userservice.NewUserService(userManagementRepo, leaveSvc)
	userHandler := userhandler.NewUserHandler(userSvc)

	// Quiz Module
	quizRepo := quizrepo.NewQuizRepository(db)
	quizSvc := quizservice.NewQuizService(quizRepo)
	quizHandler := quizhandler.NewQuizHandler(quizSvc)

	// Quiz Attempt Module
	attemptRepo := attemptrepo.NewAttemptRepository(db)
	attemptSvc := attemptservice.NewAttemptService(attemptRepo, quizSvc)
	attemptHandler := attempthandler.NewAttemptHandler(attemptSvc)

	// Setup routes
	a.setupRoutes(
		healthHandler,
		authHandler,
		institutionHandler,
		leaveHandler,
		userHandler,
		quizHandler,
		attemptHandler,
	)
}

type slogAdapter struct{}

func (slogAdapter) Info(ctx context.Context, msg string, args ...any) {
	logger.Info(ctx, msg, args...)
}

func (slogAdapter) Error(ctx context.Context, msg string, args ...any) {
	logger.Error(ctx, msg, args...)
}

