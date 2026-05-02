package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds all application configuration loaded from environment variables.
type Config struct {
	// Core runtime settings.
	AppName     string
	Environment string
	Port        string
	LogLevel    string
	Version     string
	GitSHA      string
	BuildTime   string

	// Feature flags.
	FeatureCacheWarm bool

	// DroneFleet specific
	MaxFleetSize    int
	TelemetryBuffer int

	// PostgreSQL backend. Empty means use the in-memory store for local demos/tests.
	DatabaseURL string
}

// Load reads config from environment with sensible defaults.
func Load() Config {
	return Config{
		AppName:          getEnv("APP_NAME", "dronefleet"),
		Environment:      getEnv("APP_ENV", "dev"),
		Port:             getEnv("APP_PORT", "8080"),
		LogLevel:         getEnv("LOG_LEVEL", "info"),
		Version:          getEnv("APP_VERSION", "dev"),
		GitSHA:           getEnv("GIT_SHA", "local"),
		BuildTime:        getEnv("BUILD_TIME", "unknown"),
		FeatureCacheWarm: getEnv("FEATURE_CACHE_WARM", "true") == "true",
		MaxFleetSize:     getEnvInt("MAX_FLEET_SIZE", 50),
		TelemetryBuffer:  getEnvInt("TELEMETRY_BUFFER", 100),
		DatabaseURL:      getEnv("DATABASE_URL", ""),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		i, err := strconv.Atoi(v)
		if err == nil {
			return i
		}
		fmt.Fprintf(os.Stderr, "warn: invalid int for %s=%q, using default %d\n", key, v, fallback)
	}
	return fallback
}
