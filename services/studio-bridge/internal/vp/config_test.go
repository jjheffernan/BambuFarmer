package vp

import "testing"

func TestMaskCode(t *testing.T) {
	if got := maskCode("12345678"); got != "1******8" {
		t.Fatalf("%q", got)
	}
}

func TestLoadConfigDisabledByDefault(t *testing.T) {
	t.Setenv("STUDIO_BRIDGE_VP", "")
	c := LoadConfig()
	if c.Enabled {
		t.Fatal("expected disabled")
	}
	t.Setenv("STUDIO_BRIDGE_VP", "1")
	t.Setenv("STUDIO_BRIDGE_VP_SN", "SNTEST")
	c = LoadConfig()
	if !c.Enabled || c.Serial != "SNTEST" {
		t.Fatalf("%+v", c)
	}
}
