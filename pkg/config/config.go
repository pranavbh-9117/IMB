package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all application configuration loaded from environment variables.
type Config struct {
	App      AppConfig
	Database DatabaseConfig
	JWT      JWTConfig
	Seed     SeedConfig
}

// AppConfig holds HTTP server configuration.
type AppConfig struct {
	Port string
}

// DatabaseConfig holds PostgreSQL connection parameters.
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

// JWTConfig holds token signing and expiry configuration.
type JWTConfig struct {
	Secret        string
	AccessExpiry  time.Duration
	RefreshExpiry time.Duration
}

// SeedConfig holds credentials used to bootstrap the Super Admin account.
type SeedConfig struct {
	SuperAdminEmail    string
	SuperAdminPassword string
}

// Load reads the .env file if present, then populates a Config from environment
// variables. If .env is absent (e.g. in a containerised environment), OS-level
// variables are used directly without error. All variables are required; an
// error is returned if any are missing or malformed.
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

// validate checks that all required string fields are non-empty. Duration fields
// are validated upstream in parseDuration before the struct is populated.
func (c *Config) validate() error {
	required := []struct {
		key string
		val string
	}{
		{"APP_PORT", c.App.Port},
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

// parseDuration reads a Go duration string from the named environment variable
// and parses it. Returns an error if the variable is absent or the value is not
// a valid Go duration string (e.g. "15m", "168h").
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
