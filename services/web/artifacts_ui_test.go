package main

import (
	"strings"
	"testing"
)

func TestArtifactsPage(t *testing.T) {
	if !strings.Contains(layoutHTML, `href="/artifacts"`) || !strings.Contains(layoutHTML, `eq .Active "artifacts"`) {
		t.Fatal("expected Artifacts nav and page")
	}
	if !strings.Contains(layoutHTML, "/api/v1/artifacts") || !strings.Contains(layoutHTML, "/content") {
		t.Fatal("expected artifacts list/download API usage")
	}
	if !strings.Contains(layoutHTML, "artifact-form") || !strings.Contains(layoutHTML, "loadArtifacts") {
		t.Fatal("expected upload form and loader")
	}
}
