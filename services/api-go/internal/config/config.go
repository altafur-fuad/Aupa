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
		Environment:     getFirstEnv("development", "AUPA_ENV", "ENVIRONMENT", "APP_ENV"),
		Port:            getFirstEnv("8080", "PORT"),
		ReadTimeout:     getFirstEnvDuration(10*time.Second, "AUPA_READ_TIMEOUT", "READ_TIMEOUT"),
		WriteTimeout:    getFirstEnvDuration(15*time.Second, "AUPA_WRITE_TIMEOUT", "WRITE_TIMEOUT"),
		IdleTimeout:     getFirstEnvDuration(60*time.Second, "AUPA_IDLE_TIMEOUT", "IDLE_TIMEOUT"),
		ShutdownTimeout: getFirstEnvDuration(10*time.Second, "AUPA_SHUTDOWN_TIMEOUT", "SHUTDOWN_TIMEOUT"),

		JWTSecret:       getFirstEnv("aupa-dev-jwt-secret-do-not-use-in-production-min-32-bytes", "AUPA_JWT_SECRET", "JWT_SECRET", "JWT_SIGNING_KEY"),
		JWTIssuer:       getFirstEnv("aupa.com.bd", "AUPA_JWT_ISSUER", "JWT_ISSUER"),
		JWTAudience:     getFirstEnv("aupa-client", "AUPA_JWT_AUDIENCE", "JWT_AUDIENCE"),
		AccessTokenTTL:  getFirstEnvDuration(15*time.Minute, "AUPA_ACCESS_TOKEN_TTL", "ACCESS_TOKEN_TTL"),
		RefreshTokenTTL: getFirstEnvDuration(30*24*time.Hour, "AUPA_REFRESH_TOKEN_TTL", "REFRESH_TOKEN_TTL"),

		Currency:      getFirstEnv("BDT", "AUPA_CURRENCY", "CURRENCY"),
		DefaultZoneID: getFirstEnv("zone_dhaka_central", "AUPA_DEFAULT_ZONE_ID", "DEFAULT_ZONE_ID"),
		DatabaseURL:   getFirstEnv("", "DATABASE_URL"),

		KeepAliveURL:      getFirstEnv("", "KEEP_ALIVE_URL", "RENDER_EXTERNAL_URL", "PUBLIC_URL"),
		KeepAliveInterval: getFirstEnvDuration(10*time.Minute, "AUPA_KEEP_ALIVE_INTERVAL", "KEEP_ALIVE_INTERVAL"),
	}

	if cfg.Environment == "production" {
		if cfg.JWTSecret == "aupa-dev-jwt-secret-do-not-use-in-production-min-32-bytes" || len(cfg.JWTSecret) < 32 {
			return nil, fmt.Errorf("in production, AUPA_JWT_SECRET must be explicitly set and at least 32 characters long")
		}
	}

	return cfg, nil
}

func getFirstEnv(defaultVal string, keys ...string) string {
	for _, key := range keys {
		if val := os.Getenv(key); val != "" {
			return val
		}
	}
	return defaultVal
}

func getFirstEnvDuration(defaultVal time.Duration, keys ...string) time.Duration {
	for _, key := range keys {
		if valStr := os.Getenv(key); valStr != "" {
			if dur, err := time.ParseDuration(valStr); err == nil {
				return dur
			}
		}
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
