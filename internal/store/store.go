package store

import (
	"fmt"
	"sync"
	"time"

	"github.com/spillala/dronefleet/internal/models"
)

// Store is a thread-safe in-memory data store.
// Replace with PostgreSQL for production use.
type Store struct {
	mu       sync.RWMutex
	drones   map[string]*models.Drone
	missions map[string]*models.Mission
}

// New creates and seeds a Store with sample data.
func New() *Store {
	s := &Store{
		drones:   make(map[string]*models.Drone),
		missions: make(map[string]*models.Mission),
	}
	s.seed()
	return s
}

func (s *Store) seed() {
	s.drones = map[string]*models.Drone{
		"drone-001": {ID: "drone-001", Name: "Alpha",   Status: models.StatusIdle,     Battery: 92, Lat: 12.9716, Lng: 77.5946, Altitude: 0.0},
		"drone-002": {ID: "drone-002", Name: "Bravo",   Status: models.StatusActive,   Battery: 67, Lat: 12.9720, Lng: 77.5950, Altitude: 45.0, MissionID: "mission-001"},
		"drone-003": {ID: "drone-003", Name: "Charlie", Status: models.StatusCharging, Battery: 23, Lat: 12.9710, Lng: 77.5940, Altitude: 0.0},
		"drone-004": {ID: "drone-004", Name: "Delta",   Status: models.StatusFault,    Battery: 55, Lat: 12.9725, Lng: 77.5960, Altitude: 0.0},
	}

	s.missions = map[string]*models.Mission{
		"mission-001": {
			ID:      "mission-001",
			DroneID: "drone-002",
			Status:  models.MissionActive,
			Waypoints: []models.Waypoint{
				{Lat: 12.9730, Lng: 77.5955, Alt: 50.0},
				{Lat: 12.9740, Lng: 77.5965, Alt: 50.0},
			},
			StartedAt: time.Now().Unix(),
		},
	}
}

// --- Drone operations ---

// GetFleet returns all drones with a fleet summary.
func (s *Store) GetFleet() *models.FleetSummary {
	s.mu.RLock()
	defer s.mu.RUnlock()

	drones := make([]*models.Drone, 0, len(s.drones))
	active, faults := 0, 0
	for _, d := range s.drones {
		drones = append(drones, d)
		if d.Status == models.StatusActive {
			active++
		}
		if d.Status == models.StatusFault {
			faults++
		}
	}
	return &models.FleetSummary{
		Total:  len(s.drones),
		Active: active,
		Faults: faults,
		Drones: drones,
	}
}

// GetDrone returns a drone by ID, or an error if not found.
func (s *Store) GetDrone(id string) (*models.Drone, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	d, ok := s.drones[id]
	if !ok {
		return nil, fmt.Errorf("drone %s not found", id)
	}
	return d, nil
}

// SimulateFault sets a drone to fault status with low battery.
func (s *Store) SimulateFault(id string) (*models.Drone, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	d, ok := s.drones[id]
	if !ok {
		return nil, fmt.Errorf("drone %s not found", id)
	}
	d.Status = models.StatusFault
	d.Battery = 10
	return d, nil
}

// RecoverDrone sets a drone back to idle status.
func (s *Store) RecoverDrone(id string) (*models.Drone, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	d, ok := s.drones[id]
	if !ok {
		return nil, fmt.Errorf("drone %s not found", id)
	}
	d.Status = models.StatusIdle
	return d, nil
}

// UpdateTelemetry updates live drone position and battery.
func (s *Store) UpdateTelemetry(id string, t models.TelemetryUpdate) (*models.Drone, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	d, ok := s.drones[id]
	if !ok {
		return nil, fmt.Errorf("drone %s not found", id)
	}
	d.Lat = t.Lat
	d.Lng = t.Lng
	d.Altitude = t.Altitude
	d.Battery = t.Battery
	return d, nil
}

// --- Mission operations ---

// GetMissions returns all missions.
func (s *Store) GetMissions() *models.MissionList {
	s.mu.RLock()
	defer s.mu.RUnlock()

	missions := make([]*models.Mission, 0, len(s.missions))
	for _, m := range s.missions {
		missions = append(missions, m)
	}
	return &models.MissionList{
		Total:    len(s.missions),
		Missions: missions,
	}
}

// GetMission returns a single mission by ID.
func (s *Store) GetMission(id string) (*models.Mission, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	m, ok := s.missions[id]
	if !ok {
		return nil, fmt.Errorf("mission %s not found", id)
	}
	return m, nil
}
