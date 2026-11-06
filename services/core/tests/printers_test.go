package tests

import (
	"encoding/json"
	"testing"

	"github.com/jjheffernan/BambuFarmer/services/core/internal/printers"
)

func TestIntFrom(t *testing.T) {
	if printers.IntFrom(float64(3)) != 3 {
		t.Fatal("float64")
	}
	if printers.IntFrom(7) != 7 {
		t.Fatal("int")
	}
	if printers.IntFrom(int64(9)) != 9 {
		t.Fatal("int64")
	}
	if printers.IntFrom(json.Number("11")) != 11 {
		t.Fatal("json.Number")
	}
	if printers.IntFrom("x") != 0 {
		t.Fatal("unknown")
	}
}
