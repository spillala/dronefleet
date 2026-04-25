package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/spillala/dronefleet/internal/models"
	"github.com/spillala/dronefleet/internal/store"
)

// Handler holds dependencies for all HTTP handlers.
type Handler struct {
	store   *store.Store
	version string
}

// New creates a Handler with the given store.
func New(s *store.Store, version string) *Handler {
	return &Handler{store: s, version: version}
}

// --- helpers ---

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, models.ErrorResponse{Error: msg})
}

// --- Health ---

// Health handles GET /health — liveness probe.
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, models.HealthResponse{
		Status:  "ok",
		Service: "dronefleet",
		Version: h.version,
	})
}

// Ready handles GET /ready — readiness probe.
func (h *Handler) Ready(w http.ResponseWriter, r *http.Request) {
	fleet := h.store.GetFleet()
	writeJSON(w, http.StatusOK, models.ReadyResponse{
		Status:       "ready",
		DronesLoaded: fleet.Total,
	})
}

// --- Fleet ---

// GetFleet handles GET /api/v1/fleet.
func (h *Handler) GetFleet(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, h.store.GetFleet())
}

// GetDrone handles GET /api/v1/fleet/{droneID}.
func (h *Handler) GetDrone(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "droneID")
	drone, err := h.store.GetDrone(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, drone)
}

// SimulateFault handles POST /api/v1/fleet/{droneID}/simulate-fault.
// Used by tests and AI agents to trigger fault scenarios.
func (h *Handler) SimulateFault(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "droneID")
	drone, err := h.store.SimulateFault(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, models.MessageResponse{
		Message: "fault simulated on " + id,
		Drone:   drone,
	})
}

// RecoverDrone handles POST /api/v1/fleet/{droneID}/recover.
// Called by AI SRE Agent 2 after successful remediation.
func (h *Handler) RecoverDrone(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "droneID")
	drone, err := h.store.RecoverDrone(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, models.MessageResponse{
		Message: "drone " + id + " recovered",
		Drone:   drone,
	})
}

// UpdateTelemetry handles POST /api/v1/fleet/{droneID}/telemetry.
// Receives live position and battery data from drone or simulator.
func (h *Handler) UpdateTelemetry(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "droneID")

	var t models.TelemetryUpdate
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		writeError(w, http.StatusBadRequest, "invalid telemetry payload")
		return
	}

	_, err := h.store.UpdateTelemetry(id, t)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

// --- Missions ---

// GetMissions handles GET /api/v1/missions.
func (h *Handler) GetMissions(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, h.store.GetMissions())
}

// GetMission handles GET /api/v1/missions/{missionID}.
func (h *Handler) GetMission(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "missionID")
	mission, err := h.store.GetMission(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, mission)
}
