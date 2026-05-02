-- name: ListDrones :many
SELECT id, name, status, battery, lat, lng, altitude, mission_id, created_at, updated_at
FROM drones
ORDER BY id;

-- name: GetDrone :one
SELECT id, name, status, battery, lat, lng, altitude, mission_id, created_at, updated_at
FROM drones
WHERE id = $1;

-- name: CountDronesByStatus :many
SELECT status, count(*) AS total
FROM drones
GROUP BY status;

-- name: SimulateDroneFault :one
UPDATE drones
SET status = 'fault',
    battery = 10,
    updated_at = now()
WHERE id = $1
RETURNING id, name, status, battery, lat, lng, altitude, mission_id, created_at, updated_at;

-- name: RecoverDrone :one
UPDATE drones
SET status = 'idle',
    updated_at = now()
WHERE id = $1
RETURNING id, name, status, battery, lat, lng, altitude, mission_id, created_at, updated_at;

-- name: UpdateDroneTelemetry :one
UPDATE drones
SET lat = $2,
    lng = $3,
    altitude = $4,
    battery = $5,
    updated_at = now()
WHERE id = $1
RETURNING id, name, status, battery, lat, lng, altitude, mission_id, created_at, updated_at;

-- name: InsertTelemetryEvent :one
INSERT INTO telemetry_events (drone_id, lat, lng, altitude, battery)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, drone_id, lat, lng, altitude, battery, recorded_at;
