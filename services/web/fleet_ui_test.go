package main

import (
	"strings"
	"testing"
)

func TestFleetPageShowsAMSAndTaskHelpers(t *testing.T) {
	if !strings.Contains(layoutHTML, `eq .Active "fleet"`) {
		t.Fatal("fleet page missing")
	}
	if !strings.Contains(layoutHTML, "amsChips") || !strings.Contains(layoutHTML, "data-ams-chips") {
		t.Fatal("expected AMS chip renderer")
	}
	if !strings.Contains(layoutHTML, "taskLine") || !strings.Contains(layoutHTML, "subtask_name") {
		t.Fatal("expected task line renderer")
	}
	if !strings.Contains(layoutHTML, "printer-drawer") || !strings.Contains(layoutHTML, "openDrawer") {
		t.Fatal("expected printer detail drawer")
	}
	if !strings.Contains(layoutHTML, "/liveview") {
		t.Fatal("expected liveview in drawer")
	}
	if !strings.Contains(layoutHTML, "startLiveview") || !strings.Contains(layoutHTML, "LIVEVIEW_MS") {
		t.Fatal("expected liveview poll while drawer open")
	}
	if !strings.Contains(layoutHTML, "drawer-liveview") {
		t.Fatal("expected drawer-liveview img id")
	}
	if !strings.Contains(layoutHTML, "printer.ams_changed") {
		t.Fatal("expected ams_changed SSE listener")
	}
	if !strings.Contains(layoutHTML, "alert.raised") || !strings.Contains(layoutHTML, "alert-banner") {
		t.Fatal("expected alert.raised banner")
	}
	if !strings.Contains(layoutHTML, "tag-chips") || !strings.Contains(layoutHTML, "rebuildTagChips") {
		t.Fatal("expected fleet tag filter chips")
	}
	if !strings.Contains(layoutHTML, "fleet-q") || !strings.Contains(layoutHTML, "fleetQ") {
		t.Fatal("expected fleet search filter")
	}
}
