package service

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/spillala/dronefleet/internal/api"
	"github.com/spillala/dronefleet/internal/config"
	"github.com/spillala/dronefleet/internal/repository"
)

type Service struct {
	cfg    config.Config
	logger *log.Logger
	store  repository.Store
}

func New(cfg config.Config, logger *log.Logger, store repository.Store) *Service {
	return &Service{cfg: cfg, logger: logger, store: store}
}

func (s *Service) GetHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, api.HealthResponse{
		Status:  "ok",
		Service: s.cfg.AppName,
		Version: s.cfg.Version,
	})
}

func (s *Service) GetReady(w http.ResponseWriter, r *http.Request) {
	if err := s.store.Ping(r.Context()); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, api.ErrorResponse{Error: "database not ready"})
		return
	}

	fleet, err := s.store.GetFleet(r.Context())
	if err != nil {
		writeJSON(w, http.StatusServiceUnavailable, api.ErrorResponse{Error: "fleet data not ready"})
		return
	}

	loaded := fleet.Total
	writeJSON(w, http.StatusOK, api.ReadyResponse{
		Status:       "ready",
		Database:     databaseMode(s.cfg),
		DronesLoaded: &loaded,
	})
}

func (s *Service) GetVersion(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, api.VersionResponse{
		AppName:     s.cfg.AppName,
		Environment: s.cfg.Environment,
		Version:     s.cfg.Version,
		GitSha:      s.cfg.GitSHA,
		BuildTime:   s.cfg.BuildTime,
	})
}

func (s *Service) GetConfig(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, api.ConfigResponse{
		AppName:            s.cfg.AppName,
		Environment:        s.cfg.Environment,
		LogLevel:           s.cfg.LogLevel,
		FeatureCacheWarm:   s.cfg.FeatureCacheWarm,
		DatabaseConfigured: s.cfg.DatabaseURL != "",
	})
}

func (s *Service) WarmCache(w http.ResponseWriter, _ *http.Request) {
	if !s.cfg.FeatureCacheWarm {
		writeJSON(w, http.StatusForbidden, api.ErrorResponse{Error: "cache warm task disabled"})
		return
	}

	s.logger.Printf("task=cache-warm env=%s source=api", s.cfg.Environment)
	writeJSON(w, http.StatusAccepted, api.TaskResponse{
		Status:     "accepted",
		Task:       "cache-warm",
		ExecutedAt: time.Now().UTC(),
	})
}

func (s *Service) ListFleet(w http.ResponseWriter, r *http.Request) {
	fleet, err := s.store.GetFleet(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, api.ErrorResponse{Error: "failed to load fleet"})
		return
	}
	writeJSON(w, http.StatusOK, fleet)
}

func (s *Service) GetDrone(w http.ResponseWriter, r *http.Request, droneID api.DroneId) {
	drone, err := s.store.GetDrone(r.Context(), droneID)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, drone)
}

func (s *Service) SimulateDroneFault(w http.ResponseWriter, r *http.Request, droneID api.DroneId) {
	drone, err := s.store.SimulateFault(r.Context(), droneID)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, api.DroneMessageResponse{
		Message: "fault simulated on " + droneID,
		Drone:   drone,
	})
}

func (s *Service) RecoverDrone(w http.ResponseWriter, r *http.Request, droneID api.DroneId) {
	drone, err := s.store.RecoverDrone(r.Context(), droneID)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, api.DroneMessageResponse{
		Message: "drone " + droneID + " recovered",
		Drone:   drone,
	})
}

func (s *Service) UpdateDroneTelemetry(w http.ResponseWriter, r *http.Request, droneID api.DroneId) {
	var telemetry api.TelemetryUpdate
	if err := json.NewDecoder(r.Body).Decode(&telemetry); err != nil {
		writeJSON(w, http.StatusBadRequest, api.ErrorResponse{Error: "invalid telemetry payload"})
		return
	}
	if telemetry.Battery < 0 || telemetry.Battery > 100 {
		writeJSON(w, http.StatusBadRequest, api.ErrorResponse{Error: "battery must be between 0 and 100"})
		return
	}

	if _, err := s.store.UpdateTelemetry(r.Context(), droneID, telemetry); err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, api.StatusResponse{Status: "updated"})
}

func (s *Service) ListMissions(w http.ResponseWriter, r *http.Request) {
	missions, err := s.store.GetMissions(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, api.ErrorResponse{Error: "failed to load missions"})
		return
	}
	writeJSON(w, http.StatusOK, missions)
}

func (s *Service) GetMission(w http.ResponseWriter, r *http.Request, missionID api.MissionId) {
	mission, err := s.store.GetMission(r.Context(), missionID)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, mission)
}

func databaseMode(cfg config.Config) string {
	if cfg.DatabaseURL == "" {
		return "memory"
	}
	return "postgres"
}

func writeStoreError(w http.ResponseWriter, err error) {
	if errors.Is(err, repository.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, api.ErrorResponse{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusInternalServerError, api.ErrorResponse{Error: "internal server error"})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
