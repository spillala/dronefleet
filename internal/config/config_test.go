package config

import "testing"

func TestLoadDefaults(t *testing.T) {
	clearEnv(t)

	cfg := Load()

	if cfg.AppName != "dronefleet" {
		t.Fatalf("expected default app name dronefleet, got %s", cfg.AppName)
	}
	if cfg.Port != "8080" {
		t.Fatalf("expected default port 8080, got %s", cfg.Port)
	}
	if cfg.MaxFleetSize != 50 {
		t.Fatalf("expected default MaxFleetSize 50, got %d", cfg.MaxFleetSize)
	}
	if cfg.TelemetryBuffer != 100 {
		t.Fatalf("expected default TelemetryBuffer 100, got %d", cfg.TelemetryBuffer)
	}
	if cfg.DatabaseURL != "" {
		t.Fatalf("expected default DatabaseURL to be empty, got %q", cfg.DatabaseURL)
	}
}

func TestLoadFromEnv(t *testing.T) {
	clearEnv(t)
	t.Setenv("APP_NAME", "dronefleet-test")
	t.Setenv("APP_PORT", "9090")
	t.Setenv("MAX_FLEET_SIZE", "25")
	t.Setenv("DATABASE_URL", "postgres://dronefleet:secret@localhost:5432/dronefleet")

	cfg := Load()

	if cfg.AppName != "dronefleet-test" {
		t.Fatalf("expected dronefleet-test, got %s", cfg.AppName)
	}
	if cfg.Port != "9090" {
		t.Fatalf("expected 9090, got %s", cfg.Port)
	}
	if cfg.MaxFleetSize != 25 {
		t.Fatalf("expected 25, got %d", cfg.MaxFleetSize)
	}
	if cfg.DatabaseURL == "" {
		t.Fatal("expected DatabaseURL to be loaded from env")
	}
}

func TestFeatureFlagDefault(t *testing.T) {
	clearEnv(t)

	cfg := Load()
	if !cfg.FeatureCacheWarm {
		t.Fatal("expected FeatureCacheWarm to default to true")
	}
}

func TestFeatureFlagDisabled(t *testing.T) {
	clearEnv(t)
	t.Setenv("FEATURE_CACHE_WARM", "false")

	cfg := Load()
	if cfg.FeatureCacheWarm {
		t.Fatal("expected FeatureCacheWarm to be false")
	}
}

func clearEnv(t *testing.T) {
	t.Helper()
	for _, key := range []string{
		"APP_NAME",
		"APP_ENV",
		"APP_PORT",
		"LOG_LEVEL",
		"APP_VERSION",
		"GIT_SHA",
		"BUILD_TIME",
		"FEATURE_CACHE_WARM",
		"MAX_FLEET_SIZE",
		"TELEMETRY_BUFFER",
		"DATABASE_URL",
	} {
		t.Setenv(key, "")
	}
}
