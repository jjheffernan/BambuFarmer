package main

import (
	"strings"
	"testing"
)

func TestSlicePage(t *testing.T) {
	if !strings.Contains(layoutHTML, `href="/slice"`) || !strings.Contains(layoutHTML, `eq .Active "slice"`) {
		t.Fatal("expected Slice nav and page")
	}
	if !strings.Contains(layoutHTML, "/api/v1/slice/status") || !strings.Contains(layoutHTML, "/api/v1/slice/jobs") {
		t.Fatal("expected slice API usage")
	}
	if !strings.Contains(layoutHTML, "slice-form") || !strings.Contains(layoutHTML, "ORCA_BIN") {
		t.Fatal("expected slice form and ORCA note")
	}
	if !strings.Contains(layoutHTML, `id="slice-artifact"`) || !strings.Contains(layoutHTML, "/api/v1/artifacts") {
		t.Fatal("expected staged artifact picker on Slice page")
	}
}
