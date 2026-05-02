package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spillala/dronefleet/internal/api"
	"github.com/spillala/dronefleet/internal/db"
	"github.com/spillala/dronefleet/internal/repository"
)

type Store struct {
	pool    *pgxpool.Pool
	queries *db.Queries
}

func New(pool *pgxpool.Pool) *Store {
	return &Store{
		pool:    pool,
		queries: db.New(pool),
	}
}

func Connect(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return pool, nil
}

func (s *Store) Ping(ctx context.Context) error {
	return s.pool.Ping(ctx)
}

func (s *Store) GetFleet(ctx context.Context) (api.FleetSummary, error) {
	drones, err := s.queries.ListDrones(ctx)
	if err != nil {
		return api.FleetSummary{}, err
	}

	items := make([]api.Drone, 0, len(drones))
	active, faults := 0, 0
	for _, drone := range drones {
		item := convertDrone(drone)
		items = append(items, item)
		if item.Status == api.DroneStatusActive {
			active++
		}
		if item.Status == api.DroneStatusFault {
			faults++
		}
	}

	return api.FleetSummary{
		Total:  len(items),
		Active: active,
		Faults: faults,
		Drones: items,
	}, nil
}

func (s *Store) GetDrone(ctx context.Context, id string) (api.Drone, error) {
	drone, err := s.queries.GetDrone(ctx, id)
	if err != nil {
		return api.Drone{}, mapDBError(err, "drone", id)
	}
	return convertDrone(drone), nil
}

func (s *Store) SimulateFault(ctx context.Context, id string) (api.Drone, error) {
	drone, err := s.queries.SimulateDroneFault(ctx, id)
	if err != nil {
		return api.Drone{}, mapDBError(err, "drone", id)
	}
	return convertDrone(drone), nil
}

func (s *Store) RecoverDrone(ctx context.Context, id string) (api.Drone, error) {
	drone, err := s.queries.RecoverDrone(ctx, id)
	if err != nil {
		return api.Drone{}, mapDBError(err, "drone", id)
	}
	return convertDrone(drone), nil
}

func (s *Store) UpdateTelemetry(ctx context.Context, id string, telemetry api.TelemetryUpdate) (api.Drone, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return api.Drone{}, err
	}
	defer tx.Rollback(ctx)

	q := s.queries.WithTx(tx)
	drone, err := q.UpdateDroneTelemetry(ctx, db.UpdateDroneTelemetryParams{
		ID:       id,
		Lat:      telemetry.Lat,
		Lng:      telemetry.Lng,
		Altitude: telemetry.Altitude,
		Battery:  int32(telemetry.Battery),
	})
	if err != nil {
		return api.Drone{}, mapDBError(err, "drone", id)
	}

	_, err = q.InsertTelemetryEvent(ctx, db.InsertTelemetryEventParams{
		DroneID:  id,
		Lat:      telemetry.Lat,
		Lng:      telemetry.Lng,
		Altitude: telemetry.Altitude,
		Battery:  int32(telemetry.Battery),
	})
	if err != nil {
		return api.Drone{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return api.Drone{}, err
	}
	return convertDrone(drone), nil
}

func (s *Store) GetMissions(ctx context.Context) (api.MissionList, error) {
	rows, err := s.queries.ListMissions(ctx)
	if err != nil {
		return api.MissionList{}, err
	}

	missions := make([]api.Mission, 0, len(rows))
	for _, row := range rows {
		mission, err := s.convertMission(ctx, row)
		if err != nil {
			return api.MissionList{}, err
		}
		missions = append(missions, mission)
	}
	return api.MissionList{Total: len(missions), Missions: missions}, nil
}

func (s *Store) GetMission(ctx context.Context, id string) (api.Mission, error) {
	row, err := s.queries.GetMission(ctx, id)
	if err != nil {
		return api.Mission{}, mapDBError(err, "mission", id)
	}
	return s.convertMission(ctx, row)
}

func (s *Store) convertMission(ctx context.Context, mission db.Mission) (api.Mission, error) {
	waypointRows, err := s.queries.ListMissionWaypoints(ctx, mission.ID)
	if err != nil {
		return api.Mission{}, err
	}

	waypoints := make([]api.Waypoint, 0, len(waypointRows))
	for _, waypoint := range waypointRows {
		waypoints = append(waypoints, api.Waypoint{
			Lat: waypoint.Lat,
			Lng: waypoint.Lng,
			Alt: waypoint.Alt,
		})
	}

	result := api.Mission{
		Id:        mission.ID,
		DroneId:   mission.DroneID,
		Status:    api.MissionStatus(mission.Status),
		Waypoints: waypoints,
	}
	if mission.StartedAt.Valid {
		result.StartedAt = &mission.StartedAt.Time
	}
	if mission.CompletedAt.Valid {
		result.CompletedAt = &mission.CompletedAt.Time
	}
	return result, nil
}

func convertDrone(drone db.Drone) api.Drone {
	result := api.Drone{
		Id:       drone.ID,
		Name:     drone.Name,
		Status:   api.DroneStatus(drone.Status),
		Battery:  int(drone.Battery),
		Lat:      drone.Lat,
		Lng:      drone.Lng,
		Altitude: drone.Altitude,
	}
	if drone.MissionID.Valid {
		result.MissionId = &drone.MissionID.String
	}
	return result
}

func mapDBError(err error, entity, id string) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("%w: %s %s", repository.ErrNotFound, entity, id)
	}
	return err
}
