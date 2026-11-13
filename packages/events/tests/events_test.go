package tests

import (
	"testing"
	"time"

	"github.com/jjheffernan/BambuFarmer/packages/events"
)

func TestCommandStream(t *testing.T) {
	if events.CommandStream("shop-a") != "bf.commands.edge.shop-a" {
		t.Fatal(events.CommandStream("shop-a"))
	}
}

func TestJobCommandTypeNames(t *testing.T) {
	if events.TypeCommandStartPrint != "command.start_print" {
		t.Fatal(events.TypeCommandStartPrint)
	}
	if events.TypeJobAssigned != "job.assigned" {
		t.Fatal(events.TypeJobAssigned)
	}
}

func TestNewEnvelope(t *testing.T) {
	env := events.NewEnvelope(events.TypePrinterStateChanged, "edge", "default", map[string]any{"external_sn": "x"})
	if env.Type != events.TypePrinterStateChanged || env.Source != "edge" || env.SiteID != "default" {
		t.Fatalf("%+v", env)
	}
	if env.Payload["external_sn"] != "x" {
		t.Fatal(env.Payload)
	}
	if env.TS.IsZero() || time.Since(env.TS) > time.Minute {
		t.Fatalf("bad ts %v", env.TS)
	}
}

func TestStreamConstants(t *testing.T) {
	if events.StreamPrinter != "bf.events.printer" || events.StreamJob != "bf.events.job" {
		t.Fatalf("%s %s", events.StreamPrinter, events.StreamJob)
	}
	if events.StreamSlice != "bf.events.slice" || events.TypeSliceQueued != "slice.queued" {
		t.Fatalf("%s %s", events.StreamSlice, events.TypeSliceQueued)
	}
	if events.TypePrinterAMSChanged != "printer.ams_changed" {
		t.Fatal(events.TypePrinterAMSChanged)
	}
}
