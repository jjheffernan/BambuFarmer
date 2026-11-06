package tests

import (
	"testing"

	"github.com/jjheffernan/BambuFarmer/packages/version"
)

func TestGet(t *testing.T) {
	info := version.Get()
	if info.Version == "" {
		t.Fatal("Version empty")
	}
	if version.Version == "" || version.Commit == "" {
		t.Fatal("package vars empty")
	}
}
