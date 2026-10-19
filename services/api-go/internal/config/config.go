package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config represents runtime configuration loaded from environment variables.
type Config struct {
	Environment     string
	Port            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration

	// JWT and Security
	JWTSecret       string
	JWTIssuer       string
	JWTAudience     string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration

	// Currency & Localization
	Currency      string
	DefaultZoneID string

	// Database
	DatabaseURL string

	// Keep-Alive (Prevents Render sleep and Supabase pause)
	KeepAliveURL      string
	KeepAliveInterval time.Duration
}

// Load loads configuration from environment variables with safe defaults.
func Load() (*Config, error) {
	cfg := &Config{
		Environment:     getEnv("AUPA_ENV", "development"),
		Port:            getEnv("PORT", "8080"),
		ReadTimeout:     getEnvDuration("AUPA_READ_TIMEOUT", 10*time.Second),
		WriteTimeout:    getEnvDuration("AUPA_WRITE_TIMEOUT", 15*time.Second),
		IdleTimeout:     getEnvDuration("AUPA_IDLE_TIMEOUT", 60*time.Second),
		ShutdownTimeout: getEnvDuration("AUPA_SHUTDOWN_TIMEOUT", 10*time.Second),

		JWTSecret:       getEnv("AUPA_JWT_SECRET", "aupa-dev-jwt-secret-do-not-use-in-production-min-32-bytes"),
		JWTIssuer:       getEnv("AUPA_JWT_ISSUER", "aupa.com.bd"),
		JWTAudience:     getEnv("AUPA_JWT_AUDIENCE", "aupa-client"),
		AccessTokenTTL:  getEnvDuration("AUPA_ACCESS_TOKEN_TTL", 15*time.Minute),
		RefreshTokenTTL: getEnvDuration("AUPA_REFRESH_TOKEN_TTL", 30*24*time.Hour),

		Currency:      getEnv("AUPA_CURRENCY", "BDT"),
		DefaultZoneID: getEnv("AUPA_DEFAULT_ZONE_ID", "zone_dhaka_central"),
		DatabaseURL:   getEnv("DATABASE_URL", ""),

		KeepAliveURL:      getFirstEnv("KEEP_ALIVE_URL", "RENDER_EXTERNAL_URL", "PUBLIC_URL"),
		KeepAliveInterval: getEnvDuration("AUPA_KEEP_ALIVE_INTERVAL", 10*time.Minute),
	}

	if cfg.Environment == "production" {
		if cfg.JWTSecret == "aupa-dev-jwt-secret-do-not-use-in-production-min-32-bytes" || len(cfg.JWTSecret) < 32 {
			return nil, fmt.Errorf("in production, AUPA_JWT_SECRET must be explicitly set and at least 32 characters long")
		}
	}

	return cfg, nil
}

func getFirstEnv(keys ...string) string {
	for _, key := range keys {
		if val := os.Getenv(key); val != "" {
			return val
		}
	}
	return ""
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvDuration(key string, defaultVal time.Duration) time.Duration {
	valStr := os.Getenv(key)
	if valStr == "" {
		return defaultVal
	}
	dur, err := time.ParseDuration(valStr)
	if err != nil {
		return defaultVal
	}
	return dur
}

func getEnvInt(key string, defaultVal int) int {
	valStr := os.Getenv(key)
	if valStr == "" {
		return defaultVal
	}
	val, err := strconv.Atoi(valStr)
	if err != nil {
		return defaultVal
	}
	return val
}
