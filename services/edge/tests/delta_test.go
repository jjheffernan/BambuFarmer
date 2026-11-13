package tests

import (
	"testing"

	"github.com/jjheffernan/BambuFarmer/packages/api"
	"github.com/jjheffernan/BambuFarmer/services/edge/app/delta"
)

func TestFilterChanged(t *testing.T) {
	f := delta.New()
	snap := api.PrinterSnapshot{
		PrinterSummary: api.PrinterSummary{ExternalSN: "sn1", GcodeState: "IDLE", Online: true},
	}
	if !f.Changed(snap) {
		t.Fatal("first sighting should change")
	}
	if f.Changed(snap) {
		t.Fatal("identical snapshot should not change")
	}
	snap.GcodeState = "RUNNING"
	snap.MCPercent = 10
	if !f.Changed(snap) {
		t.Fatal("state change should change")
	}
}

func TestFilterReset(t *testing.T) {
	f := delta.New()
	snap := api.PrinterSnapshot{PrinterSummary: api.PrinterSummary{ExternalSN: "a", GcodeState: "IDLE"}}
	_ = f.Changed(snap)
	f.Reset()
	if !f.Changed(snap) {
		t.Fatal("after reset should change again")
	}
}

func TestAlertRaised(t *testing.T) {
	f := delta.New()
	idle := api.PrinterSnapshot{PrinterSummary: api.PrinterSummary{ExternalSN: "sn1", Online: true, GcodeState: "IDLE"}}
	if f.AlertRaised(idle) {
		t.Fatal("idle should not alert")
	}
	paused := idle
	paused.GcodeState = "PAUSE"
	if !f.AlertRaised(paused) {
		t.Fatal("PAUSE should alert")
	}
	if f.AlertRaised(paused) {
		t.Fatal("identical attention should not re-alert")
	}
	paused.HMS = []string{"0700_8001_0001_0002"}
	if !f.AlertRaised(paused) {
		t.Fatal("new HMS should alert")
	}
	if f.AlertRaised(idle) {
		t.Fatal("return to idle should not raise")
	}
}

func TestAMSChanged(t *testing.T) {
	f := delta.New()
	empty := api.PrinterSnapshot{PrinterSummary: api.PrinterSummary{ExternalSN: "sn1"}}
	if f.AMSChanged(empty) {
		t.Fatal("first empty AMS should not emit")
	}
	with := empty
	with.AMS = []api.PrinterAMS{{ID: 0, Trays: []api.PrinterAMSTray{{Slot: 0, Type: "PLA"}}}}
	if !f.AMSChanged(with) {
		t.Fatal("first non-empty AMS should emit")
	}
	if f.AMSChanged(with) {
		t.Fatal("identical AMS should not emit")
	}
	with.AMS[0].Trays[0].Remain = 50
	if !f.AMSChanged(with) {
		t.Fatal("AMS tray change should emit")
	}
}
