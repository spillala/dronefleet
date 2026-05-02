package store

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/spillala/dronefleet/internal/api"
	"github.com/spillala/dronefleet/internal/repository"
)

// Store is a thread-safe in-memory data store.
// It is kept for local development and tests; PostgreSQL is the production backend.
type Store struct {
	mu       sync.RWMutex
	drones   map[string]api.Drone
	missions map[string]api.Mission
}

// New creates and seeds a Store with sample data.
func New() *Store {
	s := &Store{
		drones:   make(map[string]api.Drone),
		missions: make(map[string]api.Mission),
	}
	s.seed()
	return s
}

func (s *Store) seed() {
	missionID := "mission-001"
	now := time.Now().UTC()

	s.drones = map[string]api.Drone{
		"drone-001": {Id: "drone-001", Name: "Alpha", Status: api.DroneStatusIdle, Battery: 92, Lat: 12.9716, Lng: 77.5946, Altitude: 0.0},
		"drone-002": {Id: "drone-002", Name: "Bravo", Status: api.DroneStatusActive, Battery: 67, Lat: 12.9720, Lng: 77.5950, Altitude: 45.0, MissionId: &missionID},
		"drone-003": {Id: "drone-003", Name: "Charlie", Status: api.DroneStatusCharging, Battery: 23, Lat: 12.9710, Lng: 77.5940, Altitude: 0.0},
		"drone-004": {Id: "drone-004", Name: "Delta", Status: api.DroneStatusFault, Battery: 55, Lat: 12.9725, Lng: 77.5960, Altitude: 0.0},
	}

	s.missions = map[string]api.Mission{
		"mission-001": {
			Id:      "mission-001",
			DroneId: "drone-002",
			Status:  api.MissionStatusActive,
			Waypoints: []api.Waypoint{
				{Lat: 12.9730, Lng: 77.5955, Alt: 50.0},
				{Lat: 12.9740, Lng: 77.5965, Alt: 50.0},
			},
			StartedAt: &now,
		},
	}
}

func (s *Store) Ping(context.Context) error {
	return nil
}

// --- Drone operations ---

// GetFleet returns all drones with a fleet summary.
func (s *Store) GetFleet(context.Context) (api.FleetSummary, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	drones := make([]api.Drone, 0, len(s.drones))
	active, faults := 0, 0
	for _, d := range s.drones {
		drones = append(drones, d)
		if d.Status == api.DroneStatusActive {
			active++
		}
		if d.Status == api.DroneStatusFault {
			faults++
		}
	}
	return api.FleetSummary{
		Total:  len(s.drones),
		Active: active,
		Faults: faults,
		Drones: drones,
	}, nil
}

// GetDrone returns a drone by ID, or an error if not found.
func (s *Store) GetDrone(_ context.Context, id string) (api.Drone, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	d, ok := s.drones[id]
	if !ok {
		return api.Drone{}, fmt.Errorf("%w: drone %s", repository.ErrNotFound, id)
	}
	return d, nil
}

// SimulateFault sets a drone to fault status with low battery.
func (s *Store) SimulateFault(_ context.Context, id string) (api.Drone, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	d, ok := s.drones[id]
	if !ok {
		return api.Drone{}, fmt.Errorf("%w: drone %s", repository.ErrNotFound, id)
	}
	d.Status = api.DroneStatusFault
	d.Battery = 10
	s.drones[id] = d
	return d, nil
}

// RecoverDrone sets a drone back to idle status.
func (s *Store) RecoverDrone(_ context.Context, id string) (api.Drone, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	d, ok := s.drones[id]
	if !ok {
		return api.Drone{}, fmt.Errorf("%w: drone %s", repository.ErrNotFound, id)
	}
	d.Status = api.DroneStatusIdle
	s.drones[id] = d
	return d, nil
}

// UpdateTelemetry updates live drone position and battery.
func (s *Store) UpdateTelemetry(_ context.Context, id string, t api.TelemetryUpdate) (api.Drone, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	d, ok := s.drones[id]
	if !ok {
		return api.Drone{}, fmt.Errorf("%w: drone %s", repository.ErrNotFound, id)
	}
	d.Lat = t.Lat
	d.Lng = t.Lng
	d.Altitude = t.Altitude
	d.Battery = t.Battery
	s.drones[id] = d
	return d, nil
}

// --- Mission operations ---

// GetMissions returns all missions.
func (s *Store) GetMissions(context.Context) (api.MissionList, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	missions := make([]api.Mission, 0, len(s.missions))
	for _, m := range s.missions {
		missions = append(missions, m)
	}
	return api.MissionList{
		Total:    len(s.missions),
		Missions: missions,
	}, nil
}

// GetMission returns a single mission by ID.
func (s *Store) GetMission(_ context.Context, id string) (api.Mission, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	m, ok := s.missions[id]
	if !ok {
		return api.Mission{}, fmt.Errorf("%w: mission %s", repository.ErrNotFound, id)
	}
	return m, nil
}
