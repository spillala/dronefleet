package models

// DroneStatus represents the operational state of a drone.
type DroneStatus string

const (
	StatusActive   DroneStatus = "active"
	StatusIdle     DroneStatus = "idle"
	StatusCharging DroneStatus = "charging"
	StatusFault    DroneStatus = "fault"
)

// Drone represents a single drone in the fleet.
type Drone struct {
	ID        string      `json:"id"`
	Name      string      `json:"name"`
	Status    DroneStatus `json:"status"`
	Battery   int         `json:"battery"`   // 0–100 percent
	Lat       float64     `json:"lat"`
	Lng       float64     `json:"lng"`
	Altitude  float64     `json:"altitude"`
	MissionID string      `json:"mission_id,omitempty"`
}

// MissionStatus represents the state of a mission.
type MissionStatus string

const (
	MissionPlanned   MissionStatus = "planned"
	MissionActive    MissionStatus = "active"
	MissionCompleted MissionStatus = "completed"
	MissionAborted   MissionStatus = "aborted"
)

// Waypoint is a single GPS coordinate in a mission route.
type Waypoint struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
	Alt float64 `json:"alt"`
}

// Mission represents a drone flight mission.
type Mission struct {
	ID        string        `json:"id"`
	DroneID   string        `json:"drone_id"`
	Status    MissionStatus `json:"status"`
	Waypoints []Waypoint    `json:"waypoints"`
	StartedAt int64         `json:"started_at,omitempty"` // Unix timestamp
}

// FleetSummary is returned by GET /api/v1/fleet.
type FleetSummary struct {
	Total  int      `json:"total"`
	Active int      `json:"active"`
	Faults int      `json:"faults"`
	Drones []*Drone `json:"drones"`
}

// MissionList is returned by GET /api/v1/missions.
type MissionList struct {
	Total    int        `json:"total"`
	Missions []*Mission `json:"missions"`
}

// TelemetryUpdate holds incoming telemetry data.
type TelemetryUpdate struct {
	Lat      float64 `json:"lat"`
	Lng      float64 `json:"lng"`
	Altitude float64 `json:"altitude"`
	Battery  int     `json:"battery"`
}

// HealthResponse is returned by /health.
type HealthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
	Version string `json:"version"`
}

// ReadyResponse is returned by /ready.
type ReadyResponse struct {
	Status       string `json:"status"`
	DronesLoaded int    `json:"drones_loaded"`
}

// MessageResponse wraps a simple message + payload.
type MessageResponse struct {
	Message string `json:"message"`
	Drone   *Drone `json:"drone,omitempty"`
}

// ErrorResponse is returned on errors.
type ErrorResponse struct {
	Error string `json:"error"`
}
