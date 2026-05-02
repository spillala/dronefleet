-- name: ListMissions :many
SELECT id, drone_id, status, started_at, completed_at, created_at, updated_at
FROM missions
ORDER BY id;

-- name: GetMission :one
SELECT id, drone_id, status, started_at, completed_at, created_at, updated_at
FROM missions
WHERE id = $1;

-- name: ListMissionWaypoints :many
SELECT mission_id, sequence, lat, lng, alt
FROM mission_waypoints
WHERE mission_id = $1
ORDER BY sequence;
