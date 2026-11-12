package tests

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/jjheffernan/BambuFarmer/packages/farm"
	"github.com/jjheffernan/BambuFarmer/services/edge/app/bambu/fixtureplane"
)

func TestFixtureUploadAndStartPrint(t *testing.T) {
	root := filepath.Join("..", "..", "fixtures", "fm")
	if _, err := os.Stat(root); err != nil {
		t.Skip("fixtures not found:", err)
	}
	plane, err := fixtureplane.New(root, nil)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	ref, err := plane.Upload3MF(ctx, "demo.gcode.3mf", bytes.NewReader([]byte("x")))
	if err != nil {
		t.Fatal(err)
	}
	if ref.ID == "" {
		t.Fatal("empty file ref")
	}
	job, err := plane.StartPrint(ctx, farm.PrintRequest{
		JobID:      "job-fixture",
		ExternalSN: "SN-FIXTURE",
		FileRef:    ref,
		Name:       "demo",
	})
	if err != nil {
		t.Fatal(err)
	}
	if job.TaskID == "" {
		t.Fatal("empty task id")
	}
}
