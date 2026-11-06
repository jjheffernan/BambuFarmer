package tests

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBaselineFilesPresent(t *testing.T) {
	root := filepath.Join("..")
	for _, rel := range []string{
		"schema/baseline.sql",
		"migrations/001_baseline.sql",
		"seed/lab.sql",
	} {
		p := filepath.Join(root, rel)
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatalf("%s: %v", rel, err)
		}
		if len(b) < 100 {
			t.Fatalf("%s too small", rel)
		}
	}
}
