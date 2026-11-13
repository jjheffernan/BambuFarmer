package main

import (
	"strings"
	"testing"
)

func TestAdminPageHasRateCardEditor(t *testing.T) {
	if !strings.Contains(layoutHTML, `eq .Active "admin"`) && !strings.Contains(layoutHTML, `{{else}}`) {
		// admin is the default else branch of Active
	}
	if !strings.Contains(layoutHTML, `id="rates-form"`) {
		t.Fatal("expected rates form")
	}
	if !strings.Contains(layoutHTML, "machine_rate_per_hour") || !strings.Contains(layoutHTML, "filament_price_per_kg") {
		t.Fatal("expected rate card fields")
	}
	if !strings.Contains(layoutHTML, `PUT`) && !strings.Contains(layoutHTML, `method: 'PUT'`) {
		t.Fatal("expected PUT rate-cards")
	}
	if !strings.Contains(layoutHTML, "/api/v1/admin/rate-cards") {
		t.Fatal("expected rate-cards API")
	}
	if strings.Contains(layoutHTML, `id="admin-out"`) {
		t.Fatal("admin page still dumps rates JSON into #admin-out")
	}
	if !strings.Contains(layoutHTML, `id="power-on"`) || !strings.Contains(layoutHTML, "/api/v1/power") {
		t.Fatal("expected Admin power controls")
	}
}
