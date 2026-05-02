package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spillala/dronefleet/internal/app"
	"github.com/spillala/dronefleet/internal/config"
	"github.com/spillala/dronefleet/internal/postgres"
	"github.com/spillala/dronefleet/internal/repository"
	"github.com/spillala/dronefleet/internal/store"
)

func main() {
	// Structured JSON logger for production
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	// Load config from environment
	cfg := config.Load()

	// Plain logger passed into Server (your existing pattern)
	logger := log.New(os.Stdout, "", log.LstdFlags)

	repo, cleanup, err := buildStore(context.Background(), cfg)
	if err != nil {
		slog.Error("store initialization failed", "err", err)
		os.Exit(1)
	}
	defer cleanup()

	// Build server: config + OpenAPI handlers + repository.
	srv := app.NewServer(cfg, logger, repo)

	httpSrv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      srv.Routes(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		slog.Info("dronefleet starting",
			"port", cfg.Port,
			"version", cfg.Version,
			"env", cfg.Environment,
			"gitSHA", cfg.GitSHA,
			"database", databaseMode(cfg),
		)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "err", err)
			os.Exit(1)
		}
	}()

	<-stop
	slog.Info("shutdown signal received")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := httpSrv.Shutdown(ctx); err != nil {
		slog.Error("graceful shutdown failed", "err", err)
	} else {
		slog.Info("server stopped cleanly")
	}
}

func buildStore(ctx context.Context, cfg config.Config) (repository.Store, func(), error) {
	if cfg.DatabaseURL == "" {
		return store.New(), func() {}, nil
	}

	pool, err := postgres.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, nil, err
	}
	return postgres.New(pool), pool.Close, nil
}

func databaseMode(cfg config.Config) string {
	if cfg.DatabaseURL == "" {
		return "memory"
	}
	return "postgres"
}
