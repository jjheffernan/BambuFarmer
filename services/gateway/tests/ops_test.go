package tests

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jjheffernan/BambuFarmer/services/gateway/ops"
)

func TestRequestUpdateWritesFile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("OPS_DIR", dir)
	st, err := ops.RequestUpdate("tester")
	if err != nil {
		t.Fatal(err)
	}
	if !st.Pending {
		t.Fatal("expected pending")
	}
	if _, err := os.Stat(filepath.Join(dir, "update.requested")); err != nil {
		t.Fatal(err)
	}
	st2 := ops.GetStatus()
	if !st2.Pending || st2.RequestedBy != "tester" {
		t.Fatalf("status=%+v", st2)
	}
}
