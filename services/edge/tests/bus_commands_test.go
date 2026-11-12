package tests

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/jjheffernan/BambuFarmer/packages/api"
	"github.com/jjheffernan/BambuFarmer/packages/events"
	"github.com/jjheffernan/BambuFarmer/packages/farm"
	"github.com/jjheffernan/BambuFarmer/services/edge/app/bus"
)

type fakeCommander struct {
	ctrlSN   string
	ctrlOpt  api.ControlOpt
	req      farm.PrintRequest
	ref      farm.ExternalJobRef
	err      error
	bound    string
	uploaded string
	upName   string
}

func (f *fakeCommander) Control(_ context.Context, sn string, opt api.ControlOpt) error {
	f.ctrlSN = sn
	f.ctrlOpt = opt
	return nil
}

func (f *fakeCommander) StartPrint(_ context.Context, req farm.PrintRequest) (farm.ExternalJobRef, error) {
	f.req = req
	if f.err != nil {
		return farm.ExternalJobRef{}, f.err
	}
	return f.ref, nil
}

func (f *fakeCommander) Upload3MF(_ context.Context, name string, r io.Reader) (farm.FileRef, error) {
	b, _ := io.ReadAll(r)
	f.upName = name
	f.uploaded = string(b)
	return farm.FileRef{ID: "fm-99", Name: name}, nil
}

func (f *fakeCommander) BindJob(sn, jobID string, ref farm.ExternalJobRef) {
	f.bound = sn + ":" + jobID + ":" + ref.TaskID
}

func TestHandleCommandStartPrint(t *testing.T) {
	cmd := &fakeCommander{ref: farm.ExternalJobRef{TaskID: "42", SubtaskID: "43"}}
	p := &bus.Publisher{SiteID: "default", Log: nil}
	env := events.NewEnvelope("command.start_print", "core", "default", map[string]any{
		"job_id":      "job-1",
		"external_sn": "SN1",
		"file_id":     "503",
		"task_name":   "plate",
	})
	if err := p.HandleCommand(context.Background(), cmd, env); err != nil {
		t.Fatal(err)
	}
	if cmd.req.JobID != "job-1" || cmd.req.ExternalSN != "SN1" || cmd.req.FileRef.ID != "503" {
		t.Fatalf("%+v", cmd.req)
	}
	if cmd.bound != "SN1:job-1:42" {
		t.Fatalf("bind=%q", cmd.bound)
	}
}

func TestHandleCommandStartPrintFailureACKs(t *testing.T) {
	cmd := &fakeCommander{err: errors.New("missing file")}
	p := &bus.Publisher{SiteID: "default"}
	env := events.NewEnvelope("command.start_print", "core", "default", map[string]any{
		"job_id": "job-1", "external_sn": "SN1", "file_id": "503",
	})
	if err := p.HandleCommand(context.Background(), cmd, env); err != nil {
		t.Fatalf("expected ACK (nil), got %v", err)
	}
}

func TestHandleCommandStartPrintUploadsArtifact(t *testing.T) {
	dir := t.TempDir()
	key := "plate.gcode.3mf"
	if err := os.WriteFile(filepath.Join(dir, key), []byte("3mf-bytes"), 0o644); err != nil {
		t.Fatal(err)
	}
	cmd := &fakeCommander{ref: farm.ExternalJobRef{TaskID: "7"}}
	p := &bus.Publisher{SiteID: "default", ArtifactsDir: dir}
	env := events.NewEnvelope("command.start_print", "core", "default", map[string]any{
		"job_id": "job-2", "external_sn": "SN2", "object_key": key, "filename": key,
	})
	if err := p.HandleCommand(context.Background(), cmd, env); err != nil {
		t.Fatal(err)
	}
	if cmd.uploaded != "3mf-bytes" {
		t.Fatalf("upload body=%q", cmd.uploaded)
	}
	if cmd.req.FileRef.ID != "fm-99" {
		t.Fatalf("expected FM file id after upload, got %+v", cmd.req.FileRef)
	}
}

func TestHandleCommandControl(t *testing.T) {
	cmd := &fakeCommander{}
	p := &bus.Publisher{SiteID: "default"}
	env := events.NewEnvelope("command.pause", "core", "default", map[string]any{
		"external_sn": "SN9",
	})
	if err := p.HandleCommand(context.Background(), cmd, env); err != nil {
		t.Fatal(err)
	}
	if cmd.ctrlSN != "SN9" || cmd.ctrlOpt != api.ControlPause {
		t.Fatalf("%s %s", cmd.ctrlSN, cmd.ctrlOpt)
	}
}
