package tests

import (
	"testing"

	"github.com/jjheffernan/BambuFarmer/services/core/internal/config"
)

func TestLoadDefaults(t *testing.T) {
	t.Setenv("SITE_ID", "")
	t.Setenv("HTTP_ADDR", "")
	t.Setenv("DATABASE_URL", "")
	t.Setenv("REDIS_URL", "")
	t.Setenv("HOSTNAME", "")
	c := config.Load()
	if c.SiteID != "default" || c.HTTPAddr != ":8080" || c.Hostname != "1" {
		t.Fatalf("%+v", c)
	}
}

func TestLoadOverrides(t *testing.T) {
	t.Setenv("SITE_ID", "shop-a")
	t.Setenv("HTTP_ADDR", ":9090")
	t.Setenv("DATABASE_URL", "postgres://x")
	c := config.Load()
	if c.SiteID != "shop-a" || c.HTTPAddr != ":9090" || c.DatabaseURL != "postgres://x" {
		t.Fatalf("%+v", c)
	}
}
