package fail

import "testing"

func TestCode(t *testing.T) {
	if Code(nil) != "" || Code(map[string]any{"print_error": 0}) != "" {
		t.Fatal("empty")
	}
	if Code(map[string]any{"print_error": float64(42)}) != "42" {
		t.Fatal(Code(map[string]any{"print_error": float64(42)}))
	}
	if Code(map[string]any{"print_error": "0xABC"}) != "0xABC" {
		t.Fatal("string")
	}
}
