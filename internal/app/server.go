package app

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/spillala/dronefleet/internal/api"
	"github.com/spillala/dronefleet/internal/config"
	"github.com/spillala/dronefleet/internal/repository"
	"github.com/spillala/dronefleet/internal/service"
)

// Server owns the HTTP layer: config, repository, service, and routing.
type Server struct {
	cfg     config.Config
	logger  *log.Logger
	store   repository.Store
	service *service.Service
}

// NewServer wires the generated OpenAPI handler to the chosen repository.
func NewServer(cfg config.Config, logger *log.Logger, repo repository.Store) *Server {
	return &Server{
		cfg:     cfg,
		logger:  logger,
		store:   repo,
		service: service.New(cfg, logger, repo),
	}
}

// Routes builds and returns the full HTTP handler.
// Called by main.go and by tests via httptest.
func (s *Server) Routes() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))
	r.Use(s.loggingMiddleware)

	api.HandlerFromMux(s.service, r)
	r.Get("/healthz", s.service.GetHealth)
	r.Get("/readyz", s.service.GetReady)

	return r
}

// loggingMiddleware logs the request method, path, status, and duration.
func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(ww, r)
		s.logger.Printf(
			"method=%s path=%s status=%d duration=%dms remote=%s",
			r.Method, r.URL.Path, ww.Status(),
			time.Since(start).Milliseconds(), r.RemoteAddr,
		)
	})
}
