package normalize_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/jjheffernan/BambuFarmer/packages/printerstate"
	"github.com/jjheffernan/BambuFarmer/services/edge/internal/bambu/fixtureplane"
	"github.com/jjheffernan/BambuFarmer/services/edge/internal/tracker"
)

func TestFixturesViaControlPlane(t *testing.T) {
	root := filepath.Join("..", "..", "..", "..", "fixtures", "fm")
	if _, err := os.Stat(root); err != nil {
		t.Skip("fixtures not found:", err)
	}
	plane, err := fixtureplane.New(root, nil)
	if err != nil {
		t.Fatal(err)
	}
	tr := tracker.New("default")

	ctx := context.Background()
	devs, err := plane.ListDevices(ctx)
	if err != nil {
		t.Fatal(err)
	}
	snaps, err := tr.ObserveAll(devs)
	if err != nil {
		t.Fatal(err)
	}
	if len(snaps) < 1 {
		t.Fatal("expected devices")
	}
	devs2, err := plane.ListDevices(ctx)
	if err != nil {
		t.Fatal(err)
	}
	snaps2, err := tr.ObserveAll(devs2)
	if err != nil {
		t.Fatal(err)
	}
	_ = snaps
	if snaps2[0].GcodeState != string(printerstate.Idle) && snaps2[0].GcodeState != string(printerstate.Running) {
		b, _ := json.Marshal(snaps2[0])
		t.Fatalf("unexpected state: %s", b)
	}
}
