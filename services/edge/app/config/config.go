package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// Config is edge process configuration (12-factor env).
type Config struct {
	SiteID string

	FarmManagerURL       string
	FarmManagerTransport string // fm_native | cdp
	FarmManagerUser      string
	FarmManagerPassword  string
	TLSInsecure          bool // lab only: skip verify when using https:// FM URL
	CertDir              string
	CDPURL               string

	PollInterval   time.Duration
	HTTPAddr       string
	RedisURL       string
	FixtureDir     string // when set (or URL empty), poll redacted FM JSON fixtures
	ArtifactsDir   string // shared with core; object_key relative paths for Upload3MF
	DirectPrinters string // sn=name pairs when transport is direct_printer
	MQTTURL        string // tls://host:1883; empty skips the broker
}

// Load reads env with defaults suitable for local compose.
func Load() Config {
	return Config{
		SiteID:               getenv("SITE_ID", "default"),
		FarmManagerURL:       strings.TrimRight(os.Getenv("FARM_MANAGER_URL"), "/"),
		FarmManagerTransport: getenv("FARM_MANAGER_TRANSPORT", "fm_native"),
		FarmManagerUser:      os.Getenv("FARM_MANAGER_USER"),
		FarmManagerPassword:  os.Getenv("FARM_MANAGER_PASSWORD"),
		TLSInsecure:          envBool("FARM_MANAGER_TLS_INSECURE", false),
		CertDir:              os.Getenv("EDGE_CERT_DIR"),
		CDPURL:               os.Getenv("CDP_URL"),
		PollInterval:         envDuration("FARM_MANAGER_POLL_INTERVAL", 3*time.Second),
		HTTPAddr:             getenv("HTTP_ADDR", ":8080"),
		RedisURL:             os.Getenv("REDIS_URL"),
		FixtureDir:           os.Getenv("EDGE_FIXTURE_DIR"),
		ArtifactsDir:         getenv("ARTIFACTS_DIR", "/data/artifacts"),
		DirectPrinters:       os.Getenv("DIRECT_PRINTERS"),
		MQTTURL:              strings.TrimSpace(os.Getenv("FARM_MQTT_URL")),
	}
}

// UseFixtures is true when no FM URL is set and a fixture directory is configured.
// Live FARM_MANAGER_URL always wins over fixtures.
func (c Config) UseFixtures() bool {
	return !c.Configured() && c.FixtureDir != ""
}

// Configured reports whether a Farm Manager base URL is set.
func (c Config) Configured() bool {
	return c.FarmManagerURL != ""
}

func getenv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

func envBool(k string, d bool) bool {
	v := os.Getenv(k)
	if v == "" {
		return d
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return d
	}
	return b
}

func envDuration(k string, d time.Duration) time.Duration {
	v := os.Getenv(k)
	if v == "" {
		return d
	}
	parsed, err := time.ParseDuration(v)
	if err != nil {
		return d
	}
	return parsed
}
