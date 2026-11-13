package main

import (
	"strings"
	"testing"
)

func TestFilamentAndLibraryPages(t *testing.T) {
	if !strings.Contains(layoutHTML, `href="/filament"`) || !strings.Contains(layoutHTML, `eq .Active "filament"`) {
		t.Fatal("expected Filament nav and page")
	}
	if !strings.Contains(layoutHTML, "/api/v1/filament/spools") || !strings.Contains(layoutHTML, "/api/v1/filament/links") {
		t.Fatal("expected filament API calls")
	}
	if !strings.Contains(layoutHTML, `href="/library"`) || !strings.Contains(layoutHTML, `eq .Active "library"`) {
		t.Fatal("expected Library nav and page")
	}
	if !strings.Contains(layoutHTML, "/api/v1/library/models") ||
		!strings.Contains(layoutHTML, "/api/v1/library/import") ||
		!strings.Contains(layoutHTML, "/api/v1/library/export") {
		t.Fatal("expected library API calls")
	}
	if !strings.Contains(layoutHTML, `id="library-export-pick"`) {
		t.Fatal("expected library export artifact picker")
	}
	if !strings.Contains(layoutHTML, "/api/v1/integrations/status") {
		t.Fatal("expected Filament/Library to probe integrations status")
	}
	if !strings.Contains(layoutHTML, "lowIDs") || !strings.Contains(layoutHTML, "text-bf-danger") {
		t.Fatal("expected filament low-stock row highlight")
	}
}
