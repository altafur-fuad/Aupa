package config

import (
	"os"
	"testing"
	"time"
)

func TestConfigLoadDefaults(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error loading default config, got %v", err)
	}

	if cfg.Environment != "development" {
		t.Errorf("expected development, got %s", cfg.Environment)
	}
	if cfg.Port != "8080" {
		t.Errorf("expected 8080, got %s", cfg.Port)
	}
	if cfg.Currency != "BDT" {
		t.Errorf("expected BDT, got %s", cfg.Currency)
	}
	if cfg.AccessTokenTTL != 15*time.Minute {
		t.Errorf("expected 15m, got %v", cfg.AccessTokenTTL)
	}
}

func TestConfigProductionRequiresJWTSecret(t *testing.T) {
	_ = os.Setenv("AUPA_ENV", "production")
	_ = os.Setenv("AUPA_JWT_SECRET", "short")
	defer func() {
		_ = os.Unsetenv("AUPA_ENV")
		_ = os.Unsetenv("AUPA_JWT_SECRET")
	}()

	_, err := Load()
	if err == nil {
		t.Fatalf("expected error in production when JWT secret is weak, got nil")
	}
}
