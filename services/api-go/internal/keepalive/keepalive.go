package keepalive

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

// Pinger defines the interface required to ping a database (e.g. Supabase PostgreSQL).
type Pinger interface {
	Ping(ctx context.Context) error
}

// Config holds configuration for the keep-alive background worker.
type Config struct {
	// ExternalURL is the public URL of the service (e.g. https://aupa-api.onrender.com).
	// Pinged periodically to prevent free-tier platforms like Render from sleeping.
	ExternalURL string

	// Interval is how often to run the ping loop (defaults to 10 minutes).
	Interval time.Duration

	// DB is the database pinger used to keep Supabase active.
	DB Pinger

	// Logger for structured logging.
	Logger *slog.Logger

	// Client is the HTTP client used for external requests.
	Client *http.Client
}

// Start launches a background worker that runs periodically to:
// 1. Ping the service's own public HTTP URL (resets Render's 15-minute inactivity timer).
// 2. Ping the PostgreSQL database (prevents Supabase from pausing after 7 days of inactivity).
func Start(ctx context.Context, cfg Config) {
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}
	if cfg.Client == nil {
		cfg.Client = &http.Client{Timeout: 15 * time.Second}
	}
	if cfg.Interval <= 0 {
		cfg.Interval = 10 * time.Minute
	}

	targetURL := strings.TrimSpace(cfg.ExternalURL)
	if targetURL != "" {
		targetURL = strings.TrimRight(targetURL, "/")
		if !strings.HasSuffix(targetURL, "/healthz") {
			targetURL += "/healthz"
		}
	}

	cfg.Logger.Info("keepalive worker started",
		"external_target", targetURL,
		"interval", cfg.Interval.String(),
		"has_db_pinger", cfg.DB != nil,
	)

	go func() {
		ticker := time.NewTicker(cfg.Interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				cfg.Logger.Info("keepalive worker stopped")
				return
			case <-ticker.C:
				// 1. External self-ping to keep Render awake
				if targetURL != "" {
					pingExternal(ctx, cfg.Client, targetURL, cfg.Logger)
				}

				// 2. Database ping to keep Supabase PostgreSQL active
				if cfg.DB != nil {
					pingDatabase(ctx, cfg.DB, cfg.Logger)
				}
			}
		}
	}()
}

func pingExternal(ctx context.Context, client *http.Client, targetURL string, logger *slog.Logger) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		logger.Warn("keepalive: failed to create external ping request", "url", targetURL, "error", err)
		return
	}
	req.Header.Set("User-Agent", "Aupa-KeepAlive/1.0")

	resp, err := client.Do(req)
	if err != nil {
		logger.Warn("keepalive: external ping request failed", "url", targetURL, "error", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		logger.Info("keepalive: external ping successful (Render kept awake)", "status", resp.StatusCode, "url", targetURL)
	} else {
		logger.Warn("keepalive: external ping returned unexpected status", "status", resp.StatusCode, "url", targetURL)
	}
}

func pingDatabase(ctx context.Context, db Pinger, logger *slog.Logger) {
	pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := db.Ping(pingCtx); err != nil {
		logger.Warn("keepalive: database ping failed", "error", err)
	} else {
		logger.Info("keepalive: database ping successful (Supabase kept active)")
	}
}
