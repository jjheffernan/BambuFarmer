package tests

import (
	"testing"

	"github.com/jjheffernan/BambuFarmer/packages/farm"
	"github.com/jjheffernan/BambuFarmer/packages/printerstate"
	"github.com/jjheffernan/BambuFarmer/services/edge/internal/tracker"
)

func TestObserveHappyPath(t *testing.T) {
	tr := tracker.New("site-a")
	d := farm.DeviceReport{
		ExternalSN: "01P00A000000001",
		Model:      "C12",
		Name:       "P1S-1",
		Online:     true,
		GcodeState: "IDLE",
	}
	snap, from, to, forced, err := tr.Observe(d)
	if err != nil {
		t.Fatal(err)
	}
	if from != printerstate.Idle || to != printerstate.Idle || forced {
		t.Fatalf("from=%s to=%s forced=%v", from, to, forced)
	}
	if snap.Bucket != string(printerstate.BucketNotInTask) {
		t.Fatalf("bucket=%s", snap.Bucket)
	}

	d.GcodeState = "RUNNING"
	d.MCPercent = 42
	snap, from, to, forced, err = tr.Observe(d)
	if err != nil {
		t.Fatal(err)
	}
	if from != printerstate.Idle || to != printerstate.Running {
		t.Fatalf("from=%s to=%s", from, to)
	}
	if forced {
		t.Fatal("IDLE→RUNNING should be soft-allowed")
	}
	if snap.MCPercent != 42 || snap.GcodeState != "RUNNING" {
		t.Fatalf("%+v", snap)
	}
}

func TestForceSkip(t *testing.T) {
	tr := tracker.New("s")
	d := farm.DeviceReport{ExternalSN: "x", Online: true, GcodeState: "FINISH"}
	_, _, to, forced, err := tr.Observe(d)
	if err != nil {
		t.Fatal(err)
	}
	if to != printerstate.Finish || !forced {
		t.Fatalf("to=%s forced=%v", to, forced)
	}
}
