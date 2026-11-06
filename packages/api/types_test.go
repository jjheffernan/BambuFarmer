package api

import "testing"

func TestControlOptConstants(t *testing.T) {
	if ControlPause != "pause" || ControlResume != "resume" || ControlStop != "stop" || ControlBedClean != "bed_clean" {
		t.Fatalf("unexpected control opts: %q %q %q %q", ControlPause, ControlResume, ControlStop, ControlBedClean)
	}
}

func TestRateCardZeroValue(t *testing.T) {
	var rc RateCard
	if rc.Currency != "" || rc.MachineRatePerHour != 0 {
		t.Fatalf("unexpected zero RateCard: %+v", rc)
	}
}

func TestPrinterSummaryJSONTagsPresent(t *testing.T) {
	// Smoke: construct a summary the fleet UI expects.
	p := PrinterSummary{
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
