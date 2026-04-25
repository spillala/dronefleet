package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/spillala/dronefleet/internal/handlers"
	"github.com/spillala/dronefleet/internal/store"
)

const version = "1.0.0"

func main() {
	// Structured logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Dependencies
	s := store.New()
	h := handlers.New(s, version)

	// Router
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))
	r.Use(loggingMiddleware)

	// Health & readiness probes (no auth, no timeout issues)
	r.Get("/health", h.Health)
	r.Get("/ready",  h.Ready)

	// API routes
	r.Route("/api/v1", func(r chi.Router) {
		// Fleet
		r.Get("/fleet",                              h.GetFleet)
		r.Get("/fleet/{droneID}",                    h.GetDrone)
		r.Post("/fleet/{droneID}/simulate-fault",    h.SimulateFault)
		r.Post("/fleet/{droneID}/recover",           h.RecoverDrone)
		r.Post("/fleet/{droneID}/telemetry",         h.UpdateTelemetry)

		// Missions
		r.Get("/missions",              h.GetMissions)
		r.Get("/missions/{missionID}", h.GetMission)
	})

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		slog.Info("dronefleet starting", "port", port, "version", version)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "err", err)
			os.Exit(1)
		}
	}()

	<-stop
	slog.Info("shutdown signal received")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("graceful shutdown failed", "err", err)
	} else {
		slog.Info("server stopped cleanly")
	}
}

// loggingMiddleware logs each request with method, path, status and duration.
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(ww, r)
		slog.Info("request",
			"method",   r.Method,
			"path",     r.URL.Path,
			"status",   ww.Status(),
			"duration", fmt.Sprintf("%dms", time.Since(start).Milliseconds()),
			"id",       middleware.GetReqID(r.Context()),
		)
	})
}
