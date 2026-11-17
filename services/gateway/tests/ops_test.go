package tests

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/jjheffernan/BambuFarmer/services/gateway/ops"
)

func TestRequestUpdateWritesFile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("OPS_DIR", dir)
	st, err := ops.RequestUpdate("tester", []string{"edge", "web"}, false, false)
	if err != nil {
		t.Fatal(err)
	}
	if !st.Pending {
		t.Fatal("expected pending")
	}
	raw, err := os.ReadFile(filepath.Join(dir, "update.requested"))
	if err != nil {
		t.Fatal(err)
	}
	var req ops.Request
	if err := json.Unmarshal(raw, &req); err != nil {
		t.Fatal(err)
	}
	if req.By != "tester" || len(req.Components) != 2 || req.Components[0] != "edge" {
		t.Fatalf("%+v", req)
	}
	st2 := ops.GetStatus(context.Background())
	if !st2.Pending || st2.RequestedBy != "tester" {
		t.Fatalf("status=%+v", st2)
	}
	if len(st2.Components) == 0 {
		t.Fatal("expected component catalog in status")
	}
}

func TestRequestUpdateRejectsUnknown(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("OPS_DIR", dir)
	if _, err := ops.RequestUpdate("x", []string{"not-a-service"}, false, false); err == nil {
		t.Fatal("expected error")
	}
}
