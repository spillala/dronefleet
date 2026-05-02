INSERT INTO drones (id, name, status, battery, lat, lng, altitude)
VALUES
    ('drone-001', 'Alpha', 'idle', 92, 12.9716, 77.5946, 0.0),
    ('drone-002', 'Bravo', 'active', 67, 12.9720, 77.5950, 45.0),
    ('drone-003', 'Charlie', 'charging', 23, 12.9710, 77.5940, 0.0),
    ('drone-004', 'Delta', 'fault', 55, 12.9725, 77.5960, 0.0)
ON CONFLICT (id) DO NOTHING;

INSERT INTO missions (id, drone_id, status, started_at)
VALUES ('mission-001', 'drone-002', 'active', now())
ON CONFLICT (id) DO NOTHING;

UPDATE drones
SET mission_id = 'mission-001'
WHERE id = 'drone-002';

INSERT INTO mission_waypoints (mission_id, sequence, lat, lng, alt)
VALUES
    ('mission-001', 1, 12.9730, 77.5955, 50.0),
    ('mission-001', 2, 12.9740, 77.5965, 50.0)
ON CONFLICT (mission_id, sequence) DO NOTHING;
