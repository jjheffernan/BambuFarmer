package events

import (
	"testing"
	"time"
)

func TestCommandStream(t *testing.T) {
	if CommandStream("shop-a") != "bf.commands.edge.shop-a" {
		t.Fatal(CommandStream("shop-a"))
	}
}

func TestNewEnvelope(t *testing.T) {
	env := NewEnvelope(TypePrinterStateChanged, "edge", "default", map[string]any{"external_sn": "x"})
	if env.Type != TypePrinterStateChanged || env.Source != "edge" || env.SiteID != "default" {
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
	if StreamPrinter != "bf.events.printer" || StreamJob != "bf.events.job" {
		t.Fatalf("%s %s", StreamPrinter, StreamJob)
	}
}
