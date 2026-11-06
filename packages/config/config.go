package config

import (
	"os"
	"strconv"
)

// Getenv returns env or fallback.
func Getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// GetenvInt returns env parsed as int or fallback.
func GetenvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

// Common keys used across services.
const (
	EnvServiceName = "SERVICE_NAME"
	EnvHTTPAddr    = "HTTP_ADDR"
	EnvDatabaseURL = "DATABASE_URL"
	EnvRedisURL    = "REDIS_URL"
	EnvAuthMode    = "AUTH_MODE"
)