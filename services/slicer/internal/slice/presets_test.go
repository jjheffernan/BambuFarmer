package slice

import (
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMaterializePresetsWritesSettings(t *testing.T) {
	dir := t.TempDir()
	do := func(req *http.Request) (*http.Response, error) {
		var body string
		switch {
		case strings.HasSuffix(req.URL.Path, "/presets/print"):
			body = `[{"id":"pp1","process_json":{"layer_height":0.2},"printer_json":{"nozzle":0.4}}]`
		case strings.HasSuffix(req.URL.Path, "/presets/filament"):
			body = `[{"id":"fp1","name":"PLA","material":"PLA","density_g_cm3":1.24}]`
		default:
			t.Fatalf("unexpected %s", req.URL)
		}
		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(strings.NewReader(body)),
			Header:     make(http.Header),
		}, nil
	}
	files, err := MaterializePresets(context.Background(), "http://core:8080", dir, "pp1", "fp1", do)
	if err != nil {
		t.Fatal(err)
	}
	if files.PrinterJSON == "" || files.ProcessJSON == "" || files.FilamentJSON == "" {
		t.Fatalf("%+v", files)
	}
	b, err := os.ReadFile(files.ProcessJSON)
	if err != nil || !strings.Contains(string(b), "layer_height") {
		t.Fatalf("%s %v", b, err)
	}
	req := Request{}
	ApplyPresetFiles(&req, files)
	if req.PrinterJSON != files.PrinterJSON || req.ProcessJSON != files.ProcessJSON || req.FilamentJSON != files.FilamentJSON {
		t.Fatalf("%+v", req)
	}
}

func TestMaterializeRequiresCore(t *testing.T) {
	if _, err := MaterializePresets(context.Background(), "", t.TempDir(), "pp1", "", nil); err == nil {
		t.Fatal("expected CORE_URL error")
	}
	if _, err := MaterializePresets(context.Background(), "http://core", t.TempDir(), "missing", "", func(req *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`[]`)), Header: make(http.Header)}, nil
	}); err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("%v", err)
	}
}

func TestMaterializeNoop(t *testing.T) {
	files, err := MaterializePresets(context.Background(), "", filepath.Join(t.TempDir()), "", "", nil)
	if err != nil || files.PrinterJSON != "" {
		t.Fatalf("%+v %v", files, err)
	}
}
