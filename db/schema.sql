CREATE TYPE drone_status AS ENUM ('active', 'idle', 'charging', 'fault');
CREATE TYPE mission_status AS ENUM ('planned', 'active', 'completed', 'aborted');

CREATE TABLE drones (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    status drone_status NOT NULL,
    battery INTEGER NOT NULL CHECK (battery BETWEEN 0 AND 100),
    lat DOUBLE PRECISION NOT NULL,
    lng DOUBLE PRECISION NOT NULL,
    altitude DOUBLE PRECISION NOT NULL,
    mission_id TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE missions (
    id TEXT PRIMARY KEY,
    drone_id TEXT NOT NULL REFERENCES drones(id),
    status mission_status NOT NULL,
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

ALTER TABLE drones
    ADD CONSTRAINT drones_mission_id_fkey
    FOREIGN KEY (mission_id) REFERENCES missions(id)
    DEFERRABLE INITIALLY DEFERRED;

CREATE TABLE mission_waypoints (
    mission_id TEXT NOT NULL REFERENCES missions(id) ON DELETE CASCADE,
    sequence INTEGER NOT NULL,
    lat DOUBLE PRECISION NOT NULL,
    lng DOUBLE PRECISION NOT NULL,
    alt DOUBLE PRECISION NOT NULL,
    PRIMARY KEY (mission_id, sequence)
);

CREATE TABLE telemetry_events (
    id BIGSERIAL PRIMARY KEY,
    drone_id TEXT NOT NULL REFERENCES drones(id) ON DELETE CASCADE,
    lat DOUBLE PRECISION NOT NULL,
    lng DOUBLE PRECISION NOT NULL,
    altitude DOUBLE PRECISION NOT NULL,
    battery INTEGER NOT NULL CHECK (battery BETWEEN 0 AND 100),
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX telemetry_events_drone_recorded_idx
    ON telemetry_events (drone_id, recorded_at DESC);

CREATE INDEX drones_status_idx ON drones (status);
CREATE INDEX missions_status_idx ON missions (status);
