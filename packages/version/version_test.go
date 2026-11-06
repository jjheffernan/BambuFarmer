package version

import "testing"

func TestGet(t *testing.T) {
	info := Get()
	if info.Version == "" {
		t.Fatal("Version empty")
	}
	// Defaults for non-ldflags builds.
	if Version == "" || Commit == "" {
		t.Fatal("package vars empty")
	}
}
