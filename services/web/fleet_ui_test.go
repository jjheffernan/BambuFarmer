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
	if !strings.Contains(layoutHTML, `id="edge-status"`) || !strings.Contains(layoutHTML, "showEdgeStatus") {
		t.Fatal("expected fleet edge status strip")
	}
	if !strings.Contains(layoutHTML, `id="bulk-msg"`) || !strings.Contains(layoutHTML, "accepted") {
		t.Fatal("expected bulk action accepted/skipped feedback")
	}
	if !strings.Contains(layoutHTML, `id="fleet-online"`) || !strings.Contains(layoutHTML, `id="fleet-model"`) {
		t.Fatal("expected fleet online/model filters")
	}
	if !strings.Contains(layoutHTML, `id="fleet-state"`) || !strings.Contains(layoutHTML, "fleetState") {
		t.Fatal("expected fleet gcode_state filter")
	}
	if !strings.Contains(layoutHTML, "hmsChips") || !strings.Contains(layoutHTML, "data-hms-chips") {
		t.Fatal("expected fleet HMS chips on cards")
	}
	if !strings.Contains(layoutHTML, "/api/v1/edge/control/") || !strings.Contains(layoutHTML, "bfJSON(API + '/api/v1/edge/control/") {
		t.Fatal("expected per-card control via bfJSON")
	}
	if !strings.Contains(layoutHTML, "alert-banner-dismiss") {
		t.Fatal("expected alert banner dismiss")
	}
	if !strings.Contains(layoutHTML, "print_error") || !strings.Contains(layoutHTML, "layer_num") {
		t.Fatal("expected fleet print_error and layer progress")
	}
	if !strings.Contains(layoutHTML, "offline") {
		t.Fatal("expected fleet offline badge")
	}
	if !strings.Contains(layoutHTML, "data-alert-sn") || !strings.Contains(layoutHTML, "renderAlertBanner") {
		t.Fatal("expected alert banner Open → drawer")
	}
	if !strings.Contains(layoutHTML, `.get('sn')`) || !strings.Contains(layoutHTML, "drawer-copy-sn") {
		t.Fatal("expected fleet ?sn= deep-link and Copy SN")
	}
	if !strings.Contains(layoutHTML, "bf-btn-danger") || !strings.Contains(layoutHTML, `opt === 'stop'`) {
		t.Fatal("expected Stop confirm + danger styling")
	}
}
