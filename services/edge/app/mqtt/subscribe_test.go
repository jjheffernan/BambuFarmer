package mqtt

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/jjheffernan/BambuFarmer/packages/farm"
)

func TestOpenNeedsURLAndCerts(t *testing.T) {
	dir := t.TempDir()
	if _, err := Open("", dir, false); err != farm.ErrNotConfigured {
		t.Fatal(err)
	}
	if _, err := Open("tls://farm:1883", dir, false); err != farm.ErrNotConfigured {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "ca.pem"), []byte("nope"), 0o644); err != nil {
		t.Fatal(err)
	}
	src, err := Open("tls://farm:1883", dir, false)
	if err != nil || src.URL != "tls://farm:1883" {
		t.Fatal(err)
	}
}

func TestFoldDelta(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	in := make(chan Message, 2)
	out := Fold(ctx, in)
	in <- Message{Topic: "device/ABC/report", Payload: []byte(`{"print":{"command":"push_status","msg":0,"gcode_state":"RUNNING","mc_percent":1}}`)}
	in <- Message{Topic: "device/ABC/report", Payload: []byte(`{"print":{"command":"gcode_line"}}`)}
	in <- Message{Topic: "device/ABC/report", Payload: []byte(`{"print":{"command":"push_status","msg":1,"mc_percent":9}}`)}
	first := <-out
	second := <-out
	if first.GcodeState != "RUNNING" || first.MCPercent != 1 {
		t.Fatalf("%+v", first)
	}
	if second.GcodeState != "RUNNING" || second.MCPercent != 9 {
		t.Fatalf("%+v", second)
	}
}
