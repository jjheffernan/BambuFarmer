package vp

import (
	"archive/zip"
	"bytes"
	"testing"
)

func TestHintsFromPlainJSON(t *testing.T) {
	h := HintsFrom3MF([]byte(`{"printer_model":"Bambu Lab P1S","filament_type":"PLA"}`))
	if h.TargetModel != "P1S" {
		t.Fatalf("model %q", h.TargetModel)
	}
	if len(h.Filaments) == 0 || h.Filaments[0] != "PLA" {
		t.Fatalf("filaments %#v", h.Filaments)
	}
}

func TestHintsFromZip(t *testing.T) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	w, err := zw.Create("Metadata/project_settings.config")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = w.Write([]byte(`{"printer_model":"X1C","filament_type":"PETG"}`))
	_ = zw.Close()
	h := HintsFrom3MF(buf.Bytes())
	if h.TargetModel != "X1C" {
		t.Fatalf("%+v", h)
	}
}

func TestMergeHints(t *testing.T) {
	if got := MergeHints("A", "B", "C"); got != "A" {
		t.Fatal(got)
	}
	if got := MergeHints("", "B", "C"); got != "B" {
		t.Fatal(got)
	}
	if got := MergeHints("", "", "C"); got != "C" {
		t.Fatal(got)
	}
}
