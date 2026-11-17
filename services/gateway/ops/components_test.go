package ops

import (
	"testing"
)

func TestNormalizeComponentsEmptyMeansConfigured(t *testing.T) {
	ids := NormalizeComponents(nil)
	if len(ids) == 0 {
		t.Fatal("expected defaults")
	}
	found := false
	for _, id := range ids {
		if id == "gateway" || id == "core" {
			found = true
		}
	}
	if !found {
		t.Fatalf("%v", ids)
	}
}

func TestNormalizeComponentsFilters(t *testing.T) {
	ids := NormalizeComponents([]string{"edge", "nope", "edge", "web"})
	if len(ids) != 2 || ids[0] != "edge" || ids[1] != "web" {
		t.Fatalf("%v", ids)
	}
}

func TestValidComponentIDs(t *testing.T) {
	m := ValidComponentIDs()
	if !m["studio-bridge"] || !m["gateway"] {
		t.Fatalf("%v", m)
	}
}
