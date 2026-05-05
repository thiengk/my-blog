package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all configuration for the application.
// Values are read from environment variables.
type Config struct {
	// Server
	Port        string
	Environment string
	CORSOrigins []string

	// Database (PostgreSQL / Neon)
	DatabaseURL     string
	DBMaxConns      int32
	DBMinConns      int32
	DBMaxConnLife   time.Duration

	// Redis (Upstash)
	RedisURL string

	// Rate Limiting
	RateLimitPublic     int64         // max requests per window for public endpoints
	RateLimitNewsletter int64         // max requests per window for newsletter endpoints
	RateLimitWindow     time.Duration // sliding window size
}

// Load reads configuration from environment variables.
// Returns an error if required variables are missing.
func Load() (*Config, error) {
	cfg := &Config{
		Port:        getEnv("PORT", "8080"),
		Environment: getEnv("ENVIRONMENT", "development"),
		CORSOrigins: getEnvSlice("CORS_ORIGINS", []string{"http://localhost:4321"}),

		DatabaseURL:   getEnv("DATABASE_URL", ""),
		DBMaxConns:    int32(getEnvInt("DB_MAX_CONNS", 10)),
		DBMinConns:    int32(getEnvInt("DB_MIN_CONNS", 2)),
		DBMaxConnLife: getEnvDuration("DB_MAX_CONN_LIFE", 30*time.Minute),

		RedisURL: getEnv("REDIS_URL", ""),

		RateLimitPublic:     int64(getEnvInt("RATE_LIMIT_PUBLIC", 100)),
		RateLimitNewsletter: int64(getEnvInt("RATE_LIMIT_NEWSLETTER", 10)),
		RateLimitWindow:     getEnvDuration("RATE_LIMIT_WINDOW", 1*time.Minute),
	}

	// Validate required fields in production
	if cfg.Environment == "production" {
		if cfg.DatabaseURL == "" {
			return nil, fmt.Errorf("DATABASE_URL is required in production")
		}
		if cfg.RedisURL == "" {
			return nil, fmt.Errorf("REDIS_URL is required in production")
		}
	}

	return cfg, nil
}

// getEnv returns the value of an environment variable or a default value.
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// getEnvInt returns the integer value of an environment variable or a default.
func getEnvInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

// getEnvDuration returns the duration value of an environment variable or a default.
// Expects format like "30s", "5m", "1h".
func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value, exists := os.LookupEnv(key); exists {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

// getEnvSlice returns a string slice from a comma-separated environment variable.
func getEnvSlice(key string, defaultValue []string) []string {
	if value, exists := os.LookupEnv(key); exists {
		parts := strings.Split(value, ",")
		result := make([]string, 0, len(parts))
		for _, part := range parts {
			trimmed := strings.TrimSpace(part)
			if trimmed != "" {
				result = append(result, trimmed)
			}
		}
		return result
	}
	return defaultValue
}
