package repository

import (
	"context"
	"errors"

	"github.com/spillala/dronefleet/internal/api"
)

var ErrNotFound = errors.New("not found")

type Store interface {
	Ping(ctx context.Context) error
	GetFleet(ctx context.Context) (api.FleetSummary, error)
	GetDrone(ctx context.Context, id string) (api.Drone, error)
	SimulateFault(ctx context.Context, id string) (api.Drone, error)
	RecoverDrone(ctx context.Context, id string) (api.Drone, error)
	UpdateTelemetry(ctx context.Context, id string, telemetry api.TelemetryUpdate) (api.Drone, error)
	GetMissions(ctx context.Context) (api.MissionList, error)
	GetMission(ctx context.Context, id string) (api.Mission, error)
}
