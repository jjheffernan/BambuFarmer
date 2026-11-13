package direct

import (
	"bytes"
	"context"
	"testing"

	"github.com/jjheffernan/BambuFarmer/packages/api"
	"github.com/jjheffernan/BambuFarmer/packages/farm"
)

func TestEmptyPlaneIsReady(t *testing.T) {
	p := New("")
	if p.Kind() != farm.KindDirect {
		t.Fatal(p.Kind())
	}
	if err := p.Ready(context.Background()); err != nil {
		t.Fatal(err)
	}
	devs, err := p.ListDevices(context.Background())
	if err != nil || len(devs) != 0 {
		t.Fatalf("%v %+v", err, devs)
	}
}

func TestPrintAndPause(t *testing.T) {
	p := New("abc=Alpha")
	ref, err := p.Upload3MF(context.Background(), "plate.gcode.3mf", bytes.NewReader([]byte("plate")))
	if err != nil || ref.ID == "" {
		t.Fatal(err)
	}
	job, err := p.StartPrint(context.Background(), farm.PrintRequest{JobID: "j1", ExternalSN: "abc", FileRef: ref})
	if err != nil || job.TaskID != "direct-j1" {
		t.Fatalf("%+v %v", job, err)
	}
	devs, _ := p.ListDevices(context.Background())
	if len(devs) != 1 || devs[0].GcodeState != "RUNNING" || devs[0].Name != "Alpha" {
		t.Fatalf("%+v", devs)
	}
	if err := p.Control(context.Background(), "abc", api.ControlPause); err != nil {
		t.Fatal(err)
	}
	devs, _ = p.ListDevices(context.Background())
	if devs[0].GcodeState != "PAUSE" {
		t.Fatal(devs[0].GcodeState)
	}
	if err := p.Control(context.Background(), "missing", api.ControlStop); err == nil {
		t.Fatal("expected unknown printer")
	}
	jpeg, err := p.LiveviewJPEG(context.Background(), "abc")
	if err != nil || len(jpeg) < 2 || jpeg[0] != 0xff {
		t.Fatalf("%v %v", jpeg, err)
	}
}
