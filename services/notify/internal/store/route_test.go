package store

import "testing"

func TestTypeAllowed(t *testing.T) {
	if !typeAllowed([]string{"job.finished", "alert.raised"}, "job.finished") {
		t.Fatal("expected allow")
	}
	if typeAllowed([]string{"job.finished"}, "job.failed") {
		t.Fatal("expected deny")
	}
	if !typeAllowed([]string{"*"}, "anything") {
		t.Fatal("wildcard")
	}
}

func TestNormalizeTypes(t *testing.T) {
	got := normalizeTypes(nil)
	if len(got) != 3 {
		t.Fatalf("%v", got)
	}
	got = normalizeTypes([]string{" job.failed ", "job.failed", ""})
	if len(got) != 1 || got[0] != "job.failed" {
		t.Fatalf("%v", got)
	}
}
