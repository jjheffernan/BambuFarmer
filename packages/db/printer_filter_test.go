package db

import (
	"testing"

	"github.com/jjheffernan/BambuFarmer/packages/api"
)

func TestFilterPrinterSnapshots(t *testing.T) {
	items := []api.PrinterSnapshot{
		{PrinterSummary: api.PrinterSummary{ExternalSN: "A", Name: "Alpha", Model: "X1C", GcodeState: "RUNNING", Online: true, Tags: []string{"circuit-a"}}},
		{PrinterSummary: api.PrinterSummary{ExternalSN: "B", Name: "Beta", Model: "P1S", GcodeState: "IDLE", Online: false, Tags: []string{"circuit-b"}}},
		{PrinterSummary: api.PrinterSummary{ExternalSN: "C", Name: "Gamma", Model: "X1C", GcodeState: "PAUSE", Online: true, Tags: []string{"circuit-a", "lab"}}},
	}
	got := FilterPrinterSnapshots(items, PrinterListFilter{Tag: "circuit-a"})
	if len(got) != 2 {
		t.Fatalf("tag: %d", len(got))
	}
	got = FilterPrinterSnapshots(items, PrinterListFilter{State: "idle"})
	if len(got) != 1 || got[0].ExternalSN != "B" {
		t.Fatalf("state: %+v", got)
	}
	got = FilterPrinterSnapshots(items, PrinterListFilter{Model: "x1", Online: "true"})
	if len(got) != 2 {
		t.Fatalf("model+online: %d", len(got))
	}
	got = FilterPrinterSnapshots(items, PrinterListFilter{Query: "beta"})
	if len(got) != 1 || got[0].ExternalSN != "B" {
		t.Fatalf("query: %+v", got)
	}
}
