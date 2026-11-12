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
	if !strings.Contains(layoutHTML, "/api/v1/analytics/utilization") {
		t.Fatal("expected utilization API fetch")
	}
	if !strings.Contains(layoutHTML, "/api/v1/analytics/counters") {
		t.Fatal("expected counters API fetch")
	}
}
