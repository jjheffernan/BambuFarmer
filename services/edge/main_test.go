package main

import (
	"os"
	"testing"
)

// Package main has no helpers to unit-test beyond compile smoke.
// Ensure the module builds under `go test`.
func TestMainPackageCompiles(t *testing.T) {
	if os.Getenv("HTTP_ADDR") == "never-set-this-in-ci-for-this-test" {
		t.Fatal("unreachable")
	}
}
