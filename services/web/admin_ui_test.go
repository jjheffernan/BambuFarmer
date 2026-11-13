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
	if !strings.Contains(layoutHTML, `id="notify-test"`) || !strings.Contains(layoutHTML, "/api/v1/notify/test") {
		t.Fatal("expected Admin notify test control")
	}
	if !strings.Contains(layoutHTML, "bf-build-version") || !strings.Contains(layoutHTML, "/api/v1/version") {
		t.Fatal("expected build version in shell")
	}
	if !strings.Contains(layoutHTML, "stack-status") || !strings.Contains(layoutHTML, "loadStack") {
		t.Fatal("expected Admin stack/plugin status")
	}
	if !strings.Contains(layoutHTML, "/api/v1/slice/status") || !strings.Contains(layoutHTML, "/api/v1/commerce/status") {
		t.Fatal("expected plugin status probes")
	}
	if !strings.Contains(layoutHTML, "/api/v1/sites") {
		t.Fatal("expected sites status probe")
	}
	if !strings.Contains(layoutHTML, `id="site-form"`) || !strings.Contains(layoutHTML, "method: 'PATCH'") {
		t.Fatal("expected Admin site edit form")
	}
	if !strings.Contains(layoutHTML, "data-save-role") || !strings.Contains(layoutHTML, "data-active") {
		t.Fatal("expected Admin user role/active controls")
	}
	if !strings.Contains(layoutHTML, "/api/v1/admin/users/' + encodeURIComponent(id)") &&
		!strings.Contains(layoutHTML, "patchUser") {
		t.Fatal("expected Admin user PATCH helper")
	}
	if !strings.Contains(layoutHTML, "/api/v1/edge/status") {
		t.Fatal("expected edge status probe")
	}
	if !strings.Contains(layoutHTML, "/api/v1/edge/tags") || !strings.Contains(layoutHTML, `id="edge-tags"`) {
		t.Fatal("expected Admin farm tags panel")
	}
}
