package objectstore

import (
	"os"
	"strings"
)

// Config is S3-compatible object storage (MinIO or cloud S3).
type Config struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	Region    string
	Secure    bool // https when true; inferred from endpoint if unset via env
}

// FromEnv reads OBJECT_STORE_* . Empty Endpoint means disabled (local ARTIFACTS_DIR only).
func FromEnv() Config {
	ep := strings.TrimSpace(os.Getenv("OBJECT_STORE_ENDPOINT"))
	secure := strings.HasPrefix(strings.ToLower(ep), "https://")
	if v := strings.TrimSpace(os.Getenv("OBJECT_STORE_SECURE")); v == "1" || strings.EqualFold(v, "true") {
		secure = true
	}
	if v := strings.TrimSpace(os.Getenv("OBJECT_STORE_SECURE")); v == "0" || strings.EqualFold(v, "false") {
		secure = false
	}
	return Config{
		Endpoint:  ep,
		AccessKey: first(os.Getenv("OBJECT_STORE_ACCESS_KEY"), os.Getenv("MINIO_ROOT_USER")),
		SecretKey: first(os.Getenv("OBJECT_STORE_SECRET_KEY"), os.Getenv("MINIO_ROOT_PASSWORD")),
		Bucket:    first(os.Getenv("OBJECT_STORE_BUCKET"), "bambufarmer"),
		Region:    first(os.Getenv("OBJECT_STORE_REGION"), "us-east-1"),
		Secure:    secure,
	}
}

// Enabled is true when an endpoint is configured.
func (c Config) Enabled() bool {
	return strings.TrimSpace(c.Endpoint) != ""
}

func first(a, b string) string {
	a = strings.TrimSpace(a)
	if a != "" {
		return a
	}
	return strings.TrimSpace(b)
}
