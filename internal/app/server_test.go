package app

import (
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/spillala/dronefleet/internal/config"
	"github.com/spillala/dronefleet/internal/store"
)

// newTestServer builds a fresh server for each test with isolated state.
// Same helper pattern as your original tests.
func newTestServer(cfg config.Config) *Server {
	return NewServer(cfg, log.New(io.Discard, "", 0), store.New())
}

// --- Platform endpoints ---

func TestVersionEndpoint(t *testing.T) {
	srv := newTestServer(config.Config{
		AppName:     "dronefleet",
		Environment: "dev",
		Version:     "1.2.3",
		GitSHA:      "abc123",
		BuildTime:   "2026-03-14T10:00:00Z",
	})

	req := httptest.NewRequest(http.MethodGet, "/version", nil)
	rr := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
	body := rr.Body.String()
	for _, expected := range []string{"dronefleet", "1.2.3", "abc123"} {
		if !strings.Contains(body, expected) {
			t.Fatalf("expected response to contain %q, body=%s", expected, body)
		}
	}
}

func TestCacheWarmDisabled(t *testing.T) {
	srv := newTestServer(config.Config{FeatureCacheWarm: false})

	req := httptest.NewRequest(http.MethodPost, "/tasks/cache-warm", nil)
	rr := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d", rr.Code)
	}
}

// --- New tests for drone fleet endpoints, same style as your originals ---

func TestHealthEndpoint(t *testing.T) {
	srv := newTestServer(config.Config{})
	rr := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/health", nil))

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "ok") {
		t.Fatal("expected body to contain 'ok'")
	}
}

func TestHealthzEndpoint(t *testing.T) {
	// Your original /healthz path still works
	srv := newTestServer(config.Config{})
	rr := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestReadyEndpoint(t *testing.T) {
	srv := newTestServer(config.Config{})
	rr := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/ready", nil))

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestFleetEndpoint(t *testing.T) {
	srv := newTestServer(config.Config{})
	rr := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/api/v1/fleet", nil))

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	body := rr.Body.String()
	for _, expected := range []string{"drones", "total", "active"} {
		if !strings.Contains(body, expected) {
			t.Fatalf("expected body to contain %q, got: %s", expected, body)
		}
	}
}

func TestDroneNotFound(t *testing.T) {
	srv := newTestServer(config.Config{})
	rr := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/api/v1/fleet/drone-999", nil))

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestSimulateFaultAndRecover(t *testing.T) {
	srv := newTestServer(config.Config{})

	// Simulate fault
	rr := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/api/v1/fleet/drone-001/simulate-fault", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("simulate-fault: expected 200, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "fault") {
		t.Fatal("expected response to contain 'fault'")
	}

	// Recover
	rr = httptest.NewRecorder()
	srv.Routes().ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/api/v1/fleet/drone-001/recover", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("recover: expected 200, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "idle") {
		t.Fatal("expected response to contain 'idle' after recovery")
	}
}

func TestMissionsEndpoint(t *testing.T) {
	srv := newTestServer(config.Config{})
	rr := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/api/v1/missions", nil))

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "missions") {
		t.Fatal("expected body to contain 'missions'")
	}
}

func TestCacheWarmEnabled(t *testing.T) {
	srv := newTestServer(config.Config{FeatureCacheWarm: true})

	req := httptest.NewRequest(http.MethodPost, "/tasks/cache-warm", nil)
	rr := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "accepted") {
		t.Fatal("expected body to contain 'accepted'")
	}
}
