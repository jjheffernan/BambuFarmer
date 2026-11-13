package jobs

import (
	"testing"

	"github.com/jjheffernan/BambuFarmer/packages/api"
	"github.com/jjheffernan/BambuFarmer/packages/db"
)

func TestWaitingOnPrevious(t *testing.T) {
	if WaitingOnPrevious("", db.JobQueued) {
		t.Fatal("no predecessor")
	}
	if !WaitingOnPrevious("prev", db.JobRunning) || !WaitingOnPrevious("prev", "") {
		t.Fatal("should wait")
	}
	if WaitingOnPrevious("prev", db.JobSucceeded) {
		t.Fatal("predecessor finished")
	}
}

func TestPickPrinterModelAndTarget(t *testing.T) {
	idle := []api.PrinterSummary{
		{ExternalSN: "A", Model: "P1S"},
		{ExternalSN: "B", Model: "X1C"},
	}
	used := map[string]bool{}
	if sn := PickPrinter(api.Job{TargetModel: "x1c"}, idle, idle, used, 0); sn != "B" {
		t.Fatalf("model: %s", sn)
	}
	if sn := PickPrinter(api.Job{ExternalSN: "A", TargetModel: "X1C"}, idle, idle, used, 0); sn != "" {
		t.Fatalf("targeted model mismatch: %s", sn)
	}
	used["B"] = true
	if sn := PickPrinter(api.Job{}, idle, idle, used, 0); sn != "A" {
		t.Fatalf("free: %s", sn)
	}
}

func TestPickPrinterCircuitMax(t *testing.T) {
	idle := []api.PrinterSummary{
		{ExternalSN: "B", Tags: []string{"circuit-a"}},
		{ExternalSN: "C", Tags: []string{"circuit-b"}},
	}
	known := []api.PrinterSummary{
		{ExternalSN: "A", Tags: []string{"circuit-a"}},
		{ExternalSN: "B", Tags: []string{"circuit-a"}},
		{ExternalSN: "C", Tags: []string{"circuit-b"}},
	}
	used := map[string]bool{"A": true}
	if sn := PickPrinter(api.Job{}, idle, known, used, 1); sn != "C" {
		t.Fatalf("circuit full: %s", sn)
	}
	if sn := PickPrinter(api.Job{}, idle, known, used, 2); sn != "B" {
		t.Fatalf("circuit room: %s", sn)
	}
}
