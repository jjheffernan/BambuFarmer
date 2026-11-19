package tests

import (
	"testing"

	"github.com/jjheffernan/BambuFarmer/packages/events"
)

func TestStreamForType(t *testing.T) {
	cases := map[string]string{
		events.TypeJobFinished:  events.StreamJob,
		events.TypeJobFailed:    events.StreamJob,
		events.TypeAlertRaised:  events.StreamAlert,
		events.TypePrinterOffline: events.StreamPrinter,
		events.TypeSliceQueued:  events.StreamSlice,
		"studio_bridge.queued":  events.StreamStudioBridge,
		"unknown":               "",
	}
	for typ, want := range cases {
		if got := events.StreamForType(typ); got != want {
			t.Fatalf("%s: got %q want %q", typ, got, want)
		}
	}
}
