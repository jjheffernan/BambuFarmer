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
}
