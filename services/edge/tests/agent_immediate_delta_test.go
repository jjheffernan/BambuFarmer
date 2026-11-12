package tests

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jjheffernan/BambuFarmer/packages/api"
	"github.com/jjheffernan/BambuFarmer/packages/farm"
	"github.com/jjheffernan/BambuFarmer/packages/printerstate"
	"github.com/jjheffernan/BambuFarmer/services/edge/app/agent"
	"github.com/jjheffernan/BambuFarmer/services/edge/app/bambu/fixtureplane"
	"github.com/jjheffernan/BambuFarmer/services/edge/app/config"
)

func TestStartPrintAdvancesFixtureBeforePoll(t *testing.T) {
	root := filepath.Join("..", "..", "fixtures", "fm")
	if _, err := os.Stat(root); err != nil {
		t.Skip("fixtures not found:", err)
	}
	ctx := context.Background()
	plane, err := fixtureplane.New(root, nil)
	if err != nil {
		t.Fatal(err)
	}
	a := agent.New(agent.Options{
		Config: config.Config{SiteID: "lab", PollInterval: time.Hour},
		Plane:  plane,
		Log:    slog.New(slog.NewTextHandler(io.Discard, nil)),
	})
	if err := a.Connect(ctx); err != nil {
		t.Fatal(err)
	}
	ref, err := plane.Upload3MF(ctx, "plate.gcode.3mf", bytes.NewReader([]byte("x")))
	if err != nil {
		t.Fatal(err)
	}
	devs, err := plane.ListDevices(ctx)
	if err != nil || len(devs) == 0 {
		t.Fatalf("devices: %v len=%d", err, len(devs))
	}
	sn := devs[0].ExternalSN
	if _, err := a.StartPrint(ctx, farm.PrintRequest{
		JobID: "job_1", ExternalSN: sn, FileRef: ref, Name: "plate",
	}); err != nil {
		t.Fatal(err)
	}
	after, err := plane.ListDevices(ctx)
	if err != nil {
		t.Fatal(err)
	}
	var got string
	for _, d := range after {
		if d.ExternalSN == sn {
			got = d.GcodeState
			break
		}
	}
	if got != string(printerstate.Running) {
		t.Fatalf("want RUNNING immediately after StartPrint, got %q", got)
	}
	if err := a.Control(ctx, sn, api.ControlPause); err != nil {
		t.Fatal(err)
	}
	afterPause, err := plane.ListDevices(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, d := range afterPause {
		if d.ExternalSN == sn && d.GcodeState != string(printerstate.Pause) {
			t.Fatalf("want PAUSE after Control, got %q", d.GcodeState)
		}
	}
}
