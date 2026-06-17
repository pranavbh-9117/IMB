package database

import (
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/pranavbh-9117/IMB/pkg/config"
)

// Pool configuration constants
const (
	maxOpenConns    = 25
	maxIdleConns    = 10
	connMaxLifetime = 5 * time.Minute
	connMaxIdleTime = 2 * time.Minute
)

// New opens a PostgreSQL connection using the provided DatabaseConfig,
// applies connection pool settings, and returns a *gorm.DB instance.
// An error is returned if the connection cannot be established or the
// pool cannot be configured.
func New(cfg config.DatabaseConfig) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(buildDSN(cfg)), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("database: failed to open connection: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("database: failed to retrieve sql.DB: %w", err)
	}

	sqlDB.SetMaxOpenConns(maxOpenConns)
	sqlDB.SetMaxIdleConns(maxIdleConns)
	sqlDB.SetConnMaxLifetime(connMaxLifetime)
	sqlDB.SetConnMaxIdleTime(connMaxIdleTime)

	return db, nil
}

// HealthCheck verifies the database connection is alive by sending a
// lightweight ping to the server. Returns a wrapped error if unreachable.
func HealthCheck(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("database: failed to retrieve sql.DB for health check: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("database: health check failed: %w", err)
	}

	return nil
}

// buildDSN constructs a PostgreSQL DSN string from the provided config fields.
func buildDSN(cfg config.DatabaseConfig) string {
	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		cfg.Host, cfg.User, cfg.Password, cfg.Name, cfg.Port, cfg.SSLMode,
	)
}
