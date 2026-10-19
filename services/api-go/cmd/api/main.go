package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/altafur-fuad/Aupa/services/api-go/internal/config"
	"github.com/altafur-fuad/Aupa/services/api-go/internal/db/postgres"
	"github.com/altafur-fuad/Aupa/services/api-go/internal/httpapi"
	"github.com/altafur-fuad/Aupa/services/api-go/internal/keepalive"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	logger.Info("starting aupa api service",
		"env", cfg.Environment,
		"port", cfg.Port,
		"currency", cfg.Currency,
	)

	// Bind to port immediately so Render's port scanner detects the listening port without waiting
	addr := "0.0.0.0:" + cfg.Port
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		logger.Error("failed to bind port listener", "error", err, "addr", addr)
		os.Exit(1)
	}
	defer listener.Close()
	logger.Info("aupa api server bound to port successfully", "addr", addr)

	var repos httpapi.Repositories
	var dbPoolClosed func() = func() {}
	var dbInstance *postgres.DB
	if cfg.DatabaseURL != "" {
		// Use bounded 5s timeout context so database connection attempts never hang or block server startup
		dbCtx, dbCancel := context.WithTimeout(context.Background(), 5*time.Second)
		db, err := postgres.NewDB(dbCtx, cfg.DatabaseURL)
		dbCancel()
		if err != nil {
			logger.Error("failed to connect to PostgreSQL database, falling back to in-memory test doubles", "error", err)
		} else {
			dbInstance = db
			dbPoolClosed = db.Close
			logger.Info("connected to PostgreSQL database pool successfully")
			repos.UserRepo = postgres.NewUserRepository(db.Pool)
			repos.SessionRepo = postgres.NewSessionRepository(db.Pool)
			catRepo := postgres.NewCatalogRepository(db.Pool)
			repos.RestRepo = catRepo
			repos.MenuRepo = catRepo
			repos.OrderRepo = postgres.NewOrderRepository(db.Pool)
			repos.IdempRepo = postgres.NewIdempotencyRepository(db.Pool)
			repos.AttemptRepo = postgres.NewPaymentAttemptRepository(db.Pool)
			repos.AuditRepo = postgres.NewWebhookAuditRepository(db.Pool)
			repos.OutboxRepo = postgres.NewOutboxRepository(db.Pool)
		}
	} else {
		logger.Info("running in in-memory repository mode (set DATABASE_URL to connect PostgreSQL)")
	}
	defer dbPoolClosed()

	router := httpapi.NewRouter(cfg, repos, logger)

	server := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	// Server run context for graceful shutdown
	serverCtx, serverStopCtx := context.WithCancel(context.Background())

	// Start background keep-alive worker (keeps Render 24/7 awake and Supabase DB active)
	keepalive.Start(serverCtx, keepalive.Config{
		ExternalURL: cfg.KeepAliveURL,
		Interval:    cfg.KeepAliveInterval,
		DB:          dbInstance,
		Logger:      logger,
	})

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		<-sig
		logger.Info("shutdown signal received, commencing graceful shutdown...")

		shutdownCtx, cancel := context.WithTimeout(serverCtx, cfg.ShutdownTimeout)
		defer cancel()

		go func() {
			<-shutdownCtx.Done()
			if errors.Is(shutdownCtx.Err(), context.DeadlineExceeded) {
				logger.Error("graceful shutdown timed out.. forcing exit")
				os.Exit(1)
			}
		}()

		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error("server shutdown error", "error", err)
		}
		serverStopCtx()
	}()

	logger.Info("aupa api server listening", "addr", listener.Addr().String())
	if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("server listen failed", "error", err)
		os.Exit(1)
	}

	<-serverCtx.Done()
	logger.Info("aupa api server stopped cleanly")
}
