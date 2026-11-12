package config

import (
	"os"

	pkgconfig "github.com/jjheffernan/BambuFarmer/packages/config"
)

// Config is core process configuration (12-factor env).
type Config struct {
	SiteID       string
	HTTPAddr     string
	DatabaseURL  string
	RedisURL     string
	Hostname     string // consumer group member id hint
	ArtifactsDir string // local .gcode.3mf staging (shared volume with edge)
}

// Load reads env with defaults suitable for compose.
func Load() Config {
	return Config{
		SiteID:       pkgconfig.Getenv("SITE_ID", "default"),
		HTTPAddr:     pkgconfig.Getenv(pkgconfig.EnvHTTPAddr, ":8080"),
		DatabaseURL:  os.Getenv(pkgconfig.EnvDatabaseURL),
		RedisURL:     os.Getenv(pkgconfig.EnvRedisURL),
		Hostname:     pkgconfig.Getenv("HOSTNAME", "1"),
		ArtifactsDir: pkgconfig.Getenv("ARTIFACTS_DIR", "/data/artifacts"),
	}
}
