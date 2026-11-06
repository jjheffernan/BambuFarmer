package config

import (
	"os"
	"strings"
	"time"

	pkgconfig "github.com/jjheffernan/BambuFarmer/packages/config"
)

// Config is integrations process configuration.
type Config struct {
	HTTPAddr string

	SpoolmanURL   string
	SpoolmanToken string // optional; for reverse-proxy auth (X-API-Key / Bearer)

	ManyfoldURL          string
	ManyfoldToken        string // personal access token preferred
	ManyfoldClientID     string
	ManyfoldClientSecret string

	HTTPTimeout time.Duration
}

// Load reads env with defaults.
func Load() Config {
	return Config{
		HTTPAddr:             pkgconfig.Getenv(pkgconfig.EnvHTTPAddr, ":8080"),
		SpoolmanURL:          strings.TrimRight(os.Getenv("SPOOLMAN_URL"), "/"),
		SpoolmanToken:        os.Getenv("SPOOLMAN_API_TOKEN"),
		ManyfoldURL:          strings.TrimRight(os.Getenv("MANYFOLD_URL"), "/"),
		ManyfoldToken:        os.Getenv("MANYFOLD_TOKEN"),
		ManyfoldClientID:     os.Getenv("MANYFOLD_CLIENT_ID"),
		ManyfoldClientSecret: os.Getenv("MANYFOLD_CLIENT_SECRET"),
		HTTPTimeout:          15 * time.Second,
	}
}
