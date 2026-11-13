package main

import (
	"strings"
	"testing"
)

func TestCommercePage(t *testing.T) {
	if !strings.Contains(layoutHTML, `href="/commerce"`) || !strings.Contains(layoutHTML, `eq .Active "commerce"`) {
		t.Fatal("expected Commerce nav and page")
	}
	if !strings.Contains(layoutHTML, "/api/v1/commerce/status") ||
		!strings.Contains(layoutHTML, "/api/v1/commerce/webhook") ||
		!strings.Contains(layoutHTML, "/api/v1/commerce/orders/status") {
		t.Fatal("expected commerce API usage")
	}
	if !strings.Contains(layoutHTML, "commerce-webhook-form") {
		t.Fatal("expected webhook test form")
	}
	if !strings.Contains(layoutHTML, `id="commerce-artifact-pick"`) {
		t.Fatal("expected commerce artifact picker")
	}
	if !strings.Contains(layoutHTML, `id="commerce-spool-pick"`) || !strings.Contains(layoutHTML, "/api/v1/filament/spools") {
		t.Fatal("expected commerce spool picker")
	}
	if !strings.Contains(layoutHTML, "commerce-copy-job") {
		t.Fatal("expected commerce webhook copy job id")
	}
}
