package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestExportFormatAndKind(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/analytics/export?format=csv&kind=costs", nil)
	if got := exportFormat(req); got != "csv" {
		t.Fatalf("format=%q", got)
	}
	if got := exportKind(req); got != "costs" {
		t.Fatalf("kind=%q", got)
	}
	req = httptest.NewRequest(http.MethodGet, "/api/v1/analytics/export", nil)
	req.Header.Set("Accept", "text/csv")
	if got := exportFormat(req); got != "csv" {
		t.Fatalf("accept format=%q", got)
	}
	if got := exportKind(req); got != "all" {
		t.Fatalf("default kind=%q", got)
	}
}

func TestWriteExportJSON(t *testing.T) {
	rec := httptest.NewRecorder()
	bundle := exportBundle{
		Utilization: []map[string]any{{"day": "2026-11-13", "site_id": "default", "external_sn": "SN1", "idle_s": 1, "run_s": 2, "fail_s": 0, "samples": 1}},
		Costs:       []map[string]any{{"day": "2026-11-13", "site_id": "default", "material_cost": 1.5, "machine_cost": 2, "labor_cost": 0.5, "jobs": 1}},
		Failures:    []map[string]any{{"day": "2026-11-13", "site_id": "default", "code": "0700_8001", "count": 3}},
	}
	if err := writeExport(rec, "json", "all", bundle); err != nil {
		t.Fatal(err)
	}
	body := rec.Body.String()
	if !strings.Contains(body, `"utilization"`) || !strings.Contains(body, `"costs"`) || !strings.Contains(body, `"failures"`) {
		t.Fatalf("body=%s", body)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("content-type=%q", ct)
	}
}

func TestWriteExportCSVKind(t *testing.T) {
	rec := httptest.NewRecorder()
	bundle := exportBundle{
		Costs: []map[string]any{{"day": "2026-11-13", "site_id": "default", "material_cost": 1.5, "machine_cost": 2, "labor_cost": 0.5, "jobs": 1}},
	}
	if err := writeExport(rec, "csv", "costs", bundle); err != nil {
		t.Fatal(err)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "material_cost") || !strings.Contains(body, "1.5") {
		t.Fatalf("body=%s", body)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/csv") {
		t.Fatalf("content-type=%q", ct)
	}
}

func TestWriteExportCSVAllSections(t *testing.T) {
	rec := httptest.NewRecorder()
	bundle := exportBundle{
		Utilization: []map[string]any{{"day": "d", "site_id": "s", "external_sn": "sn", "idle_s": 1, "run_s": 2, "fail_s": 0, "samples": 1}},
		Failures:    []map[string]any{{"day": "d", "site_id": "s", "code": "x", "count": 9}},
	}
	if err := writeExport(rec, "csv", "all", bundle); err != nil {
		t.Fatal(err)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "utilization") || !strings.Contains(body, "failures") || !strings.Contains(body, ",x,9") {
		t.Fatalf("body=%s", body)
	}
}
