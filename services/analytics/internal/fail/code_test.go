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

func TestCodesIncludesHMS(t *testing.T) {
	got := Codes(map[string]any{
		"print_error": float64(7),
		"hms": []any{
			"0700_8001_0001_0002",
			map[string]any{"attr": float64(117473281), "code": float64(65538)},
			"0700_8001_0001_0002",
		},
	})
	if len(got) != 2 || got[0] != "7" || got[1] != "0700_8001_0001_0002" {
		t.Fatalf("%v", got)
	}
	if len(Codes(map[string]any{"hms": []string{"A", "B"}})) != 2 {
		t.Fatal("string slice")
	}
}
