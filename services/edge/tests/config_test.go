package tests

import (
	"testing"

	"github.com/jjheffernan/BambuFarmer/services/edge/internal/config"
)

func TestUseFixtures(t *testing.T) {
	c := config.Config{FarmManagerURL: "", FixtureDir: "/fixtures/fm"}
	if !c.UseFixtures() {
		t.Fatal("expected fixtures when URL empty")
	}
	c.FarmManagerURL = "http://192.168.1.1:8888"
	if c.UseFixtures() {
		t.Fatal("live URL should win over fixtures")
	}
}

func TestConfigured(t *testing.T) {
	if (config.Config{}).Configured() {
		t.Fatal("empty should be false")
	}
	if !(config.Config{FarmManagerURL: "http://x"}).Configured() {
		t.Fatal("url should configure")
	}
}

func TestLoadDefaults(t *testing.T) {
	t.Setenv("SITE_ID", "")
	t.Setenv("FARM_MANAGER_URL", "")
	t.Setenv("FARM_MANAGER_TRANSPORT", "")
	t.Setenv("EDGE_FIXTURE_DIR", "")
	c := config.Load()
	if c.SiteID != "default" || c.FarmManagerTransport != "fm_native" {
		t.Fatalf("%+v", c)
	}
}
