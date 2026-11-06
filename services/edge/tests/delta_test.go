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
