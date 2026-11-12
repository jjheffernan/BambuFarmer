package main

import (
	"strings"
	"testing"
)

func TestJobsPageUsesSSENotIntervalPoll(t *testing.T) {
	if !strings.Contains(layoutHTML, `eq .Active "jobs"`) {
		t.Fatal("jobs page section missing from layout")
	}
	if strings.Contains(layoutHTML, "setInterval(loadJobs") {
		t.Fatal("jobs page still polls with setInterval(loadJobs)")
	}
	if !strings.Contains(layoutHTML, "connectJobsSSE") {
		t.Fatal("expected connectJobsSSE on jobs page")
	}
	if !strings.Contains(layoutHTML, "EventSource") {
		t.Fatal("expected EventSource for jobs live updates")
	}
	if !strings.Contains(layoutHTML, "data-reprint") || !strings.Contains(layoutHTML, "/reprint") {
		t.Fatal("jobs page missing reprint control")
	}
	for _, ev := range []string{
		"job.queued",
		"job.assigned",
		"job.dispatched",
		"job.progress",
		"job.finished",
		"job.failed",
		"job.cancelled",
	} {
		if !strings.Contains(layoutHTML, ev) {
			t.Fatalf("missing SSE listener for %s", ev)
		}
	}
}
