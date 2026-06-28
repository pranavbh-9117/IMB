// Package config provides config functionality for the IMB platform.
package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// App configurations from .env
type Config struct {
	App      AppConfig
	Database DatabaseConfig
	JWT      JWTConfig
	OAuth    OAuthConfig
	Seed     SeedConfig
	SMTP     SMTPConfig
	Redis    RedisConfig
	Cache    CacheConfig
	Storage  StorageConfig
	DailyStatsCron string
	ReminderCron   string
}

// Server configuration.
type AppConfig struct {
	Port string
	Env  string
}

// Database configuration
type DatabaseConfig struct {
	Host            string
	Port            string
	User            string
	Password        string
	Name            string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
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

// OAuth Configuration
type OAuthConfig struct {
	GoogleClientID     string
	GoogleClientSecret string
	GoogleCallbackURL  string
}

// SMTP Configuration
type SMTPConfig struct {
	Host         string
	Port         string
	Username     string
	Password     string
	From         string
	SendTimeout  time.Duration
	MaxAttempts  int
	InitialDelay time.Duration
	MaxDelay     time.Duration
	Multiplier   float64
}

// Redis Configuration
type RedisConfig struct {
	Addr     string
	Password string
	DB       int
	TLS      bool
}

// Cache Configuration
type CacheConfig struct {
	AdminDashboardTTL time.Duration
}

// Storage Configuration
type StorageConfig struct {
	UploadDir string
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
			Host:            os.Getenv("DB_HOST"),
			Port:            os.Getenv("DB_PORT"),
			User:            os.Getenv("DB_USER"),
			Password:        os.Getenv("DB_PASSWORD"),
			Name:            os.Getenv("DB_NAME"),
			SSLMode:         os.Getenv("DB_SSLMODE"),
			MaxOpenConns:    parseIntOrDefault("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    parseIntOrDefault("DB_MAX_IDLE_CONNS", 10),
			ConnMaxLifetime: parseDurationOrDefault("DB_CONN_MAX_LIFETIME", 5*time.Minute),
			ConnMaxIdleTime: parseDurationOrDefault("DB_CONN_MAX_IDLE_TIME", 2*time.Minute),
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
		OAuth: OAuthConfig{
			GoogleClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
			GoogleClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
			GoogleCallbackURL:  os.Getenv("GOOGLE_CALLBACK_URL"),
		},
		SMTP: SMTPConfig{
			Host:         os.Getenv("SMTP_HOST"),
			Port:         os.Getenv("SMTP_PORT"),
			Username:     os.Getenv("SMTP_USERNAME"),
			Password:     os.Getenv("SMTP_PASSWORD"),
			From:         os.Getenv("SMTP_FROM"),
			SendTimeout:  parseDurationOrDefault("SMTP_SEND_TIMEOUT", 30*time.Second),
			MaxAttempts:  parseIntOrDefault("SMTP_RETRY_MAX_ATTEMPTS", 3),
			InitialDelay: parseDurationOrDefault("SMTP_RETRY_INITIAL_DELAY", 500*time.Millisecond),
			MaxDelay:     parseDurationOrDefault("SMTP_RETRY_MAX_DELAY", 10*time.Second),
			Multiplier:   parseFloatOrDefault("SMTP_RETRY_MULTIPLIER", 2.0),
		},
		Redis: RedisConfig{
			Addr:     os.Getenv("REDIS_ADDR"),
			Password: os.Getenv("REDIS_PASSWORD"),
			DB:       parseIntOrDefault("REDIS_DB", 0),
			TLS:      parseBoolOrDefault("REDIS_TLS", false),
		},
		Cache: CacheConfig{
			AdminDashboardTTL: parseDurationOrDefault("ADMIN_DASHBOARD_CACHE_TTL", 5*time.Minute),
		},
		Storage: StorageConfig{
			UploadDir: stringOrDefault("UPLOAD_DIR", "./uploads"),
		},
		DailyStatsCron: stringOrDefault("DAILY_STATS_CRON", "0 0 * * *"),
		ReminderCron:   stringOrDefault("REMINDER_CRON", "0 8 * * *"),
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
		{"GOOGLE_CLIENT_ID", c.OAuth.GoogleClientID},
		{"GOOGLE_CLIENT_SECRET", c.OAuth.GoogleClientSecret},
		{"GOOGLE_CALLBACK_URL", c.OAuth.GoogleCallbackURL},
		{"SMTP_HOST", c.SMTP.Host},
		{"SMTP_PORT", c.SMTP.Port},
		{"SMTP_USERNAME", c.SMTP.Username},
		{"SMTP_PASSWORD", c.SMTP.Password},
		{"SMTP_FROM", c.SMTP.From},
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

// parseIntOrDefault parses an integer from env, falling back to defaultVal
func parseIntOrDefault(key string, defaultVal int) int {
	val := os.Getenv(key)
	if strings.TrimSpace(val) == "" {
		return defaultVal
	}
	parsed, err := strconv.Atoi(val)
	if err != nil {
		log.Printf("WARNING: invalid value for %s: %q — falling back to default %d", key, val, defaultVal)
		return defaultVal
	}
	return parsed
}

// parseDurationOrDefault parses a duration from env, falling back to defaultVal
func parseDurationOrDefault(key string, defaultVal time.Duration) time.Duration {
	val := os.Getenv(key)
	if strings.TrimSpace(val) == "" {
		return defaultVal
	}
	parsed, err := time.ParseDuration(val)
	if err != nil {
		log.Printf("WARNING: invalid value for %s: %q — falling back to default %s", key, val, defaultVal.String())
		return defaultVal
	}
	return parsed
}

// parseFloatOrDefault parses a float64 from env, falling back to defaultVal
func parseFloatOrDefault(key string, defaultVal float64) float64 {
	val := os.Getenv(key)
	if strings.TrimSpace(val) == "" {
		return defaultVal
	}
	parsed, err := strconv.ParseFloat(val, 64)
	if err != nil {
		log.Printf("WARNING: invalid value for %s: %q — falling back to default %f", key, val, defaultVal)
		return defaultVal
	}
	return parsed
}

// parseBoolOrDefault parses a boolean from env, falling back to defaultVal
func parseBoolOrDefault(key string, defaultVal bool) bool {
	val := os.Getenv(key)
	if strings.TrimSpace(val) == "" {
		return defaultVal
	}
	parsed, err := strconv.ParseBool(val)
	if err != nil {
		log.Printf("WARNING: invalid value for %s: %q — falling back to default %t", key, val, defaultVal)
		return defaultVal
	}
	return parsed
}

// stringOrDefault parses a string from env, falling back to defaultVal
func stringOrDefault(key string, defaultVal string) string {
	val := os.Getenv(key)
	if strings.TrimSpace(val) == "" {
		return defaultVal
	}
	return val
}

