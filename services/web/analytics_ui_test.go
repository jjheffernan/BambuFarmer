package main

import (
	"strings"
	"testing"
)

func TestAnalyticsPageRendersRollupTableNotRawJSON(t *testing.T) {
	if !strings.Contains(layoutHTML, `eq .Active "analytics"`) {
		t.Fatal("analytics page section missing from layout")
	}
	if strings.Contains(layoutHTML, "analytics-out") {
		t.Fatal("analytics page still dumps raw JSON into #analytics-out")
	}
	if !strings.Contains(layoutHTML, "analytics-util-table") {
		t.Fatal("expected daily utilization table")
	}
	if !strings.Contains(layoutHTML, "analytics-counters") {
		t.Fatal("expected event counters summary")
	}
	if !strings.Contains(layoutHTML, "analytics-cost-table") {
		t.Fatal("expected daily costs table")
	}
	if !strings.Contains(layoutHTML, "analytics-fail-table") {
		t.Fatal("expected failure codes table")
	}
	if !strings.Contains(layoutHTML, "analytics-util-bars") {
		t.Fatal("expected utilization chart bars")
	}
	if !strings.Contains(layoutHTML, "analytics-cost-bars") {
		t.Fatal("expected cost chart bars")
	}
	if !strings.Contains(layoutHTML, "analytics-fail-bars") {
		t.Fatal("expected failure chart bars")
	}
	for _, path := range []string{
		"/api/v1/analytics/utilization",
		"/api/v1/analytics/counters",
		"/api/v1/analytics/costs",
		"/api/v1/analytics/failures",
		"/api/v1/analytics/export",
	} {
		if !strings.Contains(layoutHTML, path) {
			t.Fatalf("expected %s API fetch", path)
		}
	}
	if !strings.Contains(layoutHTML, "analytics-export-csv") {
		t.Fatal("expected CSV export link")
	}
}
