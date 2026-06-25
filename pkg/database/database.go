// Package database provides database functionality for the IMB platform.
package database

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/pranavbh-9117/IMB/pkg/config"
)



// Create Postgresql connection
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

	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	return db, nil
}

// PoolStats returns current connection pool statistics.
type PoolStats struct {
	OpenConnections int   `json:"open_connections"`
	InUse           int   `json:"in_use"`
	Idle            int   `json:"idle"`
	WaitCount       int64 `json:"wait_count"`
	MaxOpenConns    int   `json:"max_open_conns"`
}

// Check DB is Reachable and returns pool stats
func HealthCheck(db *gorm.DB) (PoolStats, error) {
	sqlDB, err := db.DB()
	if err != nil {
		return PoolStats{}, fmt.Errorf("database: failed to retrieve sql.DB for health check: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return PoolStats{}, fmt.Errorf("database: health check failed: %w", err)
	}

	stats := sqlDB.Stats()
	return PoolStats{
		OpenConnections: stats.OpenConnections,
		InUse:           stats.InUse,
		Idle:            stats.Idle,
		WaitCount:       stats.WaitCount,
		MaxOpenConns:    stats.MaxOpenConnections,
	}, nil
}

// Builds Connection Strings
func buildDSN(cfg config.DatabaseConfig) string {
	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		cfg.Host, cfg.User, cfg.Password, cfg.Name, cfg.Port, cfg.SSLMode,
	)
}
