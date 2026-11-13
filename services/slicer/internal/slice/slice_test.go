package slice

import "testing"

func TestJobUnset(t *testing.T) {
	got, err := Job("  ", " plate ")
	if err != nil || got.Sliced || got.Name != "plate" || got.Reason != "ORCA_BIN unset" {
		t.Fatalf("%+v %v", got, err)
	}
	if _, err := Job("", ""); err == nil {
		t.Fatal("expected name")
	}
}

func TestJobBinPresentDoesNotSlice(t *testing.T) {
	got, err := Job("/usr/bin/true", "plate")
	if err != nil || got.Sliced || got.Reason == "ORCA_BIN unset" {
		t.Fatalf("%+v %v", got, err)
	}
}
