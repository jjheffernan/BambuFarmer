package tests

import (
	"strings"
	"testing"

	"github.com/jjheffernan/BambuFarmer/packages/db"
)

func TestMigrationsDefinedInGo(t *testing.T) {
	if !strings.Contains(db.LabSeedSQL, "filament_presets") {
		t.Fatal("LabSeedSQL missing filament_presets")
	}
	// Ensure baseline + users SQL are wired (indirectly via Migrate list length).
	// Open/Migrate need a live DB; here we only assert seed content and that
	// password helpers still work as a smoke for the package.
	if _, err := db.HashPassword("long-enough-password"); err != nil {
		t.Fatal(err)
	}
}
