package tests

import (
	"encoding/json"
	"testing"

	"github.com/jjheffernan/BambuFarmer/packages/api"
	"github.com/jjheffernan/BambuFarmer/services/core/internal/printers"
)

func TestIntFrom(t *testing.T) {
	if printers.IntFrom(float64(3)) != 3 {
		t.Fatal("float64")
	}
	if printers.IntFrom(7) != 7 {
		t.Fatal("int")
	}
	if printers.IntFrom(int64(9)) != 9 {
		t.Fatal("int64")
	}
	if printers.IntFrom(json.Number("11")) != 11 {
		t.Fatal("json.Number")
	}
	if printers.IntFrom("x") != 0 {
		t.Fatal("unknown")
	}
}

func TestNormalizeSerials(t *testing.T) {
	got := printers.NormalizeSerials(api.PrinterActionsRequest{
		Serials:     []string{" A ", "", "A", "B"},
		ExternalSNs: []string{"B", "C"},
	})
	if len(got) != 3 || got[0] != "A" || got[1] != "B" || got[2] != "C" {
		t.Fatalf("got %#v", got)
	}
}

func TestPublishControlValidation(t *testing.T) {
	if _, err := printers.PublishControl(t.Context(), nil, "default", "led", []string{"SN1"}); err == nil {
		t.Fatal("expected unsupported action")
	}
	if _, err := printers.PublishControl(t.Context(), nil, "default", api.ControlPause, nil); err == nil {
		t.Fatal("expected serials required")
	}
	if _, err := printers.PublishControl(t.Context(), nil, "default", api.ControlPause, []string{"SN1"}); err == nil {
		t.Fatal("expected redis not configured")
	}
}
