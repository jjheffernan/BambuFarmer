package tests

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/jjheffernan/BambuFarmer/packages/api"
	"github.com/jjheffernan/BambuFarmer/packages/farm"
	"github.com/jjheffernan/BambuFarmer/packages/printerstate"
	"github.com/jjheffernan/BambuFarmer/services/edge/app/bambu/fixtureplane"
)

func TestFixtureUploadAndStartPrint(t *testing.T) {
	root := filepath.Join("..", "..", "fixtures", "fm")
	if _, err := os.Stat(root); err != nil {
		t.Skip("fixtures not found:", err)
	}
	plane, err := fixtureplane.New(root, nil)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	ref, err := plane.Upload3MF(ctx, "demo.gcode.3mf", bytes.NewReader([]byte("x")))
	if err != nil {
		t.Fatal(err)
	}
	if ref.ID == "" {
		t.Fatal("empty file ref")
	}
	job, err := plane.StartPrint(ctx, farm.PrintRequest{
		JobID:      "job-fixture",
		ExternalSN: "SN-FIXTURE",
		FileRef:    ref,
		Name:       "demo",
	})
	if err != nil {
		t.Fatal(err)
	}
	if job.TaskID == "" {
		t.Fatal("empty task id")
	}
}

func TestFixtureStartPrintRequiresFileID(t *testing.T) {
	root := filepath.Join("..", "..", "fixtures", "fm")
	if _, err := os.Stat(root); err != nil {
		t.Skip("fixtures not found:", err)
	}
	plane, err := fixtureplane.New(root, nil)
	if err != nil {
		t.Fatal(err)
	}
	_, err = plane.StartPrint(context.Background(), farm.PrintRequest{
		JobID:      "job-missing-file",
		ExternalSN: "SN-FIXTURE",
	})
	if err == nil {
		t.Fatal("expected error when FileRef.ID missing")
	}
}

func TestFixtureStartPrintAndControlAdvanceState(t *testing.T) {
	root := filepath.Join("..", "..", "fixtures", "fm")
	if _, err := os.Stat(root); err != nil {
		t.Skip("fixtures not found:", err)
	}
	plane, err := fixtureplane.New(root, nil)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	devs, err := plane.ListDevices(ctx)
	if err != nil || len(devs) == 0 {
		t.Fatalf("list devices: %v len=%d", err, len(devs))
	}
	sn := devs[0].ExternalSN
	if sn == "" {
		t.Fatal("empty SN in fixture")
	}

	ref, err := plane.Upload3MF(ctx, "plate.gcode.3mf", bytes.NewReader([]byte("plate")))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := plane.StartPrint(ctx, farm.PrintRequest{
		JobID: "job-state", ExternalSN: sn, FileRef: ref, Name: "plate",
	}); err != nil {
		t.Fatal(err)
	}

	afterStart, err := plane.ListDevices(ctx)
	if err != nil {
		t.Fatal(err)
	}
	got := findSN(afterStart, sn)
	if got.GcodeState != string(printerstate.Running) {
		t.Fatalf("after StartPrint want RUNNING got %q", got.GcodeState)
	}

	if err := plane.Control(ctx, sn, api.ControlPause); err != nil {
		t.Fatal(err)
	}
	afterPause, err := plane.ListDevices(ctx)
	if err != nil {
		t.Fatal(err)
	}
	got = findSN(afterPause, sn)
	if got.GcodeState != string(printerstate.Pause) {
		t.Fatalf("after pause want PAUSE got %q", got.GcodeState)
	}

	if err := plane.Control(ctx, sn, api.ControlStop); err != nil {
		t.Fatal(err)
	}
	afterStop, err := plane.ListDevices(ctx)
	if err != nil {
		t.Fatal(err)
	}
	got = findSN(afterStop, sn)
	if got.GcodeState != string(printerstate.Idle) {
		t.Fatalf("after stop want IDLE got %q", got.GcodeState)
	}
}

func findSN(devs []farm.DeviceReport, sn string) farm.DeviceReport {
	for _, d := range devs {
		if d.ExternalSN == sn {
			return d
		}
	}
	return farm.DeviceReport{}
}
