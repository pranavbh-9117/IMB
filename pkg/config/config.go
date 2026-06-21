// Package config provides config functionality for the IMB platform.
package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// App configurations from .env
type Config struct {
	App      AppConfig
	Database DatabaseConfig
	JWT      JWTConfig
	Seed     SeedConfig
}

// Server configuration.
type AppConfig struct {
	Port string
	Env  string
}

// Database configuration
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

// JWT Configuration
type JWTConfig struct {
	Secret        string
	AccessExpiry  time.Duration
	RefreshExpiry time.Duration
}

// HardSeed Configuration
type SeedConfig struct {
	SuperAdminEmail    string
	SuperAdminPassword string
}

// Loads configuration  from .env
func Load() (*Config, error) {

	_ = godotenv.Load()

	accessExpiry, err := parseDuration("JWT_ACCESS_EXPIRY")
	if err != nil {
		return nil, err
	}

	refreshExpiry, err := parseDuration("JWT_REFRESH_EXPIRY")
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		App: AppConfig{
			Port: os.Getenv("APP_PORT"),
			Env:  os.Getenv("APP_ENV"),
		},
		Database: DatabaseConfig{
			Host:     os.Getenv("DB_HOST"),
			Port:     os.Getenv("DB_PORT"),
			User:     os.Getenv("DB_USER"),
			Password: os.Getenv("DB_PASSWORD"),
			Name:     os.Getenv("DB_NAME"),
			SSLMode:  os.Getenv("DB_SSLMODE"),
		},
		JWT: JWTConfig{
			Secret:        os.Getenv("JWT_SECRET"),
			AccessExpiry:  accessExpiry,
			RefreshExpiry: refreshExpiry,
		},
		Seed: SeedConfig{
			SuperAdminEmail:    os.Getenv("SUPER_ADMIN_EMAIL"),
			SuperAdminPassword: os.Getenv("SUPER_ADMIN_PASSWORD"),
		},
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Checks all the .env variables are loaded
func (c *Config) validate() error {
	required := []struct {
		key string
		val string
	}{
		{"APP_PORT", c.App.Port},
		{"APP_ENV", c.App.Env},
		{"DB_HOST", c.Database.Host},
		{"DB_PORT", c.Database.Port},
		{"DB_USER", c.Database.User},
		{"DB_PASSWORD", c.Database.Password},
		{"DB_NAME", c.Database.Name},
		{"DB_SSLMODE", c.Database.SSLMode},
		{"JWT_SECRET", c.JWT.Secret},
		{"SUPER_ADMIN_EMAIL", c.Seed.SuperAdminEmail},
		{"SUPER_ADMIN_PASSWORD", c.Seed.SuperAdminPassword},
	}

	var missing []string
	for _, r := range required {
		if strings.TrimSpace(r.val) == "" {
			missing = append(missing, r.key)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}

	return nil
}

// Parse JWT Duration defined in .env
func parseDuration(key string) (time.Duration, error) {
	val := os.Getenv(key)
	if strings.TrimSpace(val) == "" {
		return 0, fmt.Errorf("missing required environment variable: %s", key)
	}

	d, err := time.ParseDuration(val)
	if err != nil {
		return 0, fmt.Errorf(
			"invalid value for %s: %q — expected a Go duration string (e.g. \"15m\", \"168h\")",
			key, val,
		)
	}

	return d, nil
}
