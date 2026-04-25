package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/spillala/dronefleet/internal/handlers"
	"github.com/spillala/dronefleet/internal/models"
	"github.com/spillala/dronefleet/internal/store"
)

// newTestServer builds a fresh router for each test — isolated state.
func newTestServer() http.Handler {
	s := store.New()
	h := handlers.New(s, "test")

	r := chi.NewRouter()
	r.Get("/health", h.Health)
	r.Get("/ready",  h.Ready)
	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/fleet",                           h.GetFleet)
		r.Get("/fleet/{droneID}",                 h.GetDrone)
		r.Post("/fleet/{droneID}/simulate-fault", h.SimulateFault)
		r.Post("/fleet/{droneID}/recover",        h.RecoverDrone)
		r.Post("/fleet/{droneID}/telemetry",      h.UpdateTelemetry)
		r.Get("/missions",                        h.GetMissions)
		r.Get("/missions/{missionID}",            h.GetMission)
	})
	return r
}

func get(srv http.Handler, path string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, req)
	return rr
}

func post(srv http.Handler, path string, body any) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	if body != nil {
		json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(http.MethodPost, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, req)
	return rr
}

// --- Health ---

func TestHealth(t *testing.T) {
	rr := get(newTestServer(), "/health")
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var resp models.HealthResponse
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp.Status != "ok" {
		t.Errorf("expected status=ok, got %q", resp.Status)
	}
}

func TestReady(t *testing.T) {
	rr := get(newTestServer(), "/ready")
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var resp models.ReadyResponse
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp.Status != "ready" {
		t.Errorf("expected status=ready, got %q", resp.Status)
	}
	if resp.DronesLoaded == 0 {
		t.Error("expected drones_loaded > 0")
	}
}

// --- Fleet ---

func TestGetFleet(t *testing.T) {
	rr := get(newTestServer(), "/api/v1/fleet")
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var resp models.FleetSummary
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp.Total != 4 {
		t.Errorf("expected 4 drones, got %d", resp.Total)
	}
	if len(resp.Drones) == 0 {
		t.Error("expected drones array to be non-empty")
	}
}

func TestGetFleetHasFaults(t *testing.T) {
	rr := get(newTestServer(), "/api/v1/fleet")
	var resp models.FleetSummary
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp.Faults < 1 {
		t.Errorf("expected at least 1 fault drone (drone-004), got %d", resp.Faults)
	}
}

func TestGetDrone(t *testing.T) {
	rr := get(newTestServer(), "/api/v1/fleet/drone-001")
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var resp models.Drone
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp.ID != "drone-001" {
		t.Errorf("expected drone-001, got %q", resp.ID)
	}
}

func TestGetDroneNotFound(t *testing.T) {
	rr := get(newTestServer(), "/api/v1/fleet/drone-999")
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

// --- Fault simulation ---

func TestSimulateFault(t *testing.T) {
	srv := newTestServer()
	rr := post(srv, "/api/v1/fleet/drone-001/simulate-fault", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var resp models.MessageResponse
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp.Drone == nil {
		t.Fatal("expected drone in response")
	}
	if resp.Drone.Status != models.StatusFault {
		t.Errorf("expected status=fault, got %q", resp.Drone.Status)
	}
}

func TestSimulateFaultNotFound(t *testing.T) {
	rr := post(newTestServer(), "/api/v1/fleet/drone-999/simulate-fault", nil)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestRecoverDrone(t *testing.T) {
	srv := newTestServer()
	// Put in fault first
	post(srv, "/api/v1/fleet/drone-001/simulate-fault", nil)
	// Then recover
	rr := post(srv, "/api/v1/fleet/drone-001/recover", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var resp models.MessageResponse
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp.Drone.Status != models.StatusIdle {
		t.Errorf("expected status=idle after recovery, got %q", resp.Drone.Status)
	}
}

func TestRecoverDroneNotFound(t *testing.T) {
	rr := post(newTestServer(), "/api/v1/fleet/drone-999/recover", nil)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

// --- Telemetry ---

func TestUpdateTelemetry(t *testing.T) {
	payload := models.TelemetryUpdate{
		Lat: 12.9730, Lng: 77.5960, Altitude: 52.5, Battery: 65,
	}
	rr := post(newTestServer(), "/api/v1/fleet/drone-002/telemetry", payload)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestUpdateTelemetryInvalidJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/fleet/drone-002/telemetry",
		bytes.NewBufferString("not-json"))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	newTestServer().ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

// --- Missions ---

func TestGetMissions(t *testing.T) {
	rr := get(newTestServer(), "/api/v1/missions")
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var resp models.MissionList
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp.Total == 0 {
		t.Error("expected at least one mission")
	}
}

func TestGetMission(t *testing.T) {
	rr := get(newTestServer(), "/api/v1/missions/mission-001")
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var m models.Mission
	json.NewDecoder(rr.Body).Decode(&m)
	if m.ID != "mission-001" {
		t.Errorf("expected mission-001, got %q", m.ID)
	}
}

func TestGetMissionNotFound(t *testing.T) {
	rr := get(newTestServer(), "/api/v1/missions/mission-999")
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}
