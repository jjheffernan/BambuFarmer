package main

import "testing"

func TestGetenv(t *testing.T) {
	t.Setenv("BF_AN_TEST", "1")
	if getenv("BF_AN_TEST", "0") != "1" {
		t.Fatal("env")
	}
	if getenv("BF_AN_MISSING", "0") != "0" {
		t.Fatal("fallback")
	}
}
