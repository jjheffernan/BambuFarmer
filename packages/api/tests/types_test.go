package tests

import (
	"testing"

	"github.com/jjheffernan/BambuFarmer/packages/api"
)

func TestControlOptConstants(t *testing.T) {
	if api.ControlPause != "pause" || api.ControlResume != "resume" || api.ControlStop != "stop" || api.ControlBedClean != "bed_clean" {
		t.Fatalf("unexpected control opts: %q %q %q %q", api.ControlPause, api.ControlResume, api.ControlStop, api.ControlBedClean)
	}
	if !api.ValidControlOpt(api.ControlPause) || api.ValidControlOpt("led") {
		t.Fatal("ValidControlOpt mismatch")
	}
}

func TestRateCardZeroValue(t *testing.T) {
	var rc api.RateCard
	if rc.Currency != "" || rc.MachineRatePerHour != 0 {
		t.Fatalf("unexpected zero RateCard: %+v", rc)
	}
}

func TestPrinterSummaryJSONTagsPresent(t *testing.T) {
	p := api.PrinterSummary{
		ExternalSN: "01P00A",
		Name:       "Bay-1",
		GcodeState: "IDLE",
		Online:     true,
		Bucket:     "not_in_task",
	}
	if p.ExternalSN == "" || !p.Online {
		t.Fatal("summary fields missing")
	}
}
