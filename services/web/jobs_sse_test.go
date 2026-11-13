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
	if !strings.Contains(layoutHTML, "data-cancel") || !strings.Contains(layoutHTML, "/cancel") {
		t.Fatal("jobs page missing cancel control")
	}
	if !strings.Contains(layoutHTML, "data-dispatch") || !strings.Contains(layoutHTML, "/dispatch") {
		t.Fatal("jobs page missing dispatch control")
	}
	if !strings.Contains(layoutHTML, "Printer SN to dispatch") {
		t.Fatal("expected dispatch to prompt for SN when job has none")
	}
	if !strings.Contains(layoutHTML, `id="job-artifact-pick"`) || !strings.Contains(layoutHTML, "loadJobCreatePickers") {
		t.Fatal("jobs page missing artifact picker")
	}
	if !strings.Contains(layoutHTML, `id="job-sn-pick"`) ||
		!strings.Contains(layoutHTML, `id="job-after-pick"`) ||
		!strings.Contains(layoutHTML, `id="job-spool-pick"`) {
		t.Fatal("jobs page missing printer/after/spool pickers")
	}
	if !strings.Contains(layoutHTML, "data-history") || !strings.Contains(layoutHTML, "/events") {
		t.Fatal("jobs page missing history control")
	}
	if !strings.Contains(layoutHTML, "data-detail") || !strings.Contains(layoutHTML, "job-detail-panel") {
		t.Fatal("jobs page missing detail panel")
	}
	if !strings.Contains(layoutHTML, "job-events-panel") {
		t.Fatal("jobs page missing events panel")
	}
	if !strings.Contains(layoutHTML, `id="jobs-q"`) || !strings.Contains(layoutHTML, `id="jobs-state"`) {
		t.Fatal("jobs page missing search or state filter")
	}
	if !strings.Contains(layoutHTML, `id="job-spool"`) || !strings.Contains(layoutHTML, `id="job-model"`) || !strings.Contains(layoutHTML, "/queue") {
		t.Fatal("jobs page missing queue reorder or model target")
	}
	if !strings.Contains(layoutHTML, `id="job-queue-pos"`) || !strings.Contains(layoutHTML, "queue_pos") {
		t.Fatal("jobs page missing optional queue_pos on create")
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
