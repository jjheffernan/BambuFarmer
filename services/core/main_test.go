package main

import "testing"

func TestIntFrom(t *testing.T) {
	if intFrom(float64(3)) != 3 {
		t.Fatal("float64")
	}
	if intFrom(7) != 7 {
		t.Fatal("int")
	}
	if intFrom(int64(9)) != 9 {
		t.Fatal("int64")
	}
	if intFrom("x") != 0 {
		t.Fatal("unknown")
	}
}

func TestGetenv(t *testing.T) {
	t.Setenv("BF_CORE_TEST", "yes")
	if getenv("BF_CORE_TEST", "no") != "yes" {
		t.Fatal("env")
	}
	if getenv("BF_CORE_MISSING", "no") != "no" {
		t.Fatal("fallback")
	}
}
