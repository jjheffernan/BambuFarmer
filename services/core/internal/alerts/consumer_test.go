package alerts

import "testing"

func TestStringSlice(t *testing.T) {
	if got := stringSlice([]any{"a", 1, "b"}); len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Fatalf("%v", got)
	}
	if got := stringSlice([]string{"x"}); len(got) != 1 || got[0] != "x" {
		t.Fatalf("%v", got)
	}
	if stringSlice(nil) != nil {
		t.Fatal("nil")
	}
}
