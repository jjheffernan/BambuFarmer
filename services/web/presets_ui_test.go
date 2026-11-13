package main

import (
	"strings"
	"testing"
)

func TestPresetsPageHasImportExport(t *testing.T) {
	if !strings.Contains(layoutHTML, `eq .Active "presets"`) {
		t.Fatal("presets page missing")
	}
	if !strings.Contains(layoutHTML, "/api/v1/presets/export") {
		t.Fatal("expected export link")
	}
	if !strings.Contains(layoutHTML, "/api/v1/presets/import") {
		t.Fatal("expected import POST")
	}
	if !strings.Contains(layoutHTML, `id="presets-import"`) || !strings.Contains(layoutHTML, `id="presets-file"`) {
		t.Fatal("expected import textarea and file input")
	}
	if !strings.Contains(layoutHTML, "loadPresets") {
		t.Fatal("expected reload after import")
	}
	if !strings.Contains(layoutHTML, "/revisions") || !strings.Contains(layoutHTML, "/restore") {
		t.Fatal("expected revision history/restore UI")
	}
}
