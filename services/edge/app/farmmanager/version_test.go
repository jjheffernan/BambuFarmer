package farmmanager

import "testing"

func TestCompareRelease(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"03.00.00.00", "03.00.00.00", 0},
		{"03.02.00.00", "03.00.00.00", 1},
		{"02.04.00.00", "03.00.00.00", -1},
		{"3.2", "03.00.00.00", 1},
		{"03.00", "03.00.00.01", -1},
	}
	for _, tc := range cases {
		if got := CompareRelease(tc.a, tc.b); got != tc.want {
			t.Errorf("CompareRelease(%q,%q)=%d want %d", tc.a, tc.b, got, tc.want)
		}
	}
}

func TestCompatibleRelease(t *testing.T) {
	ok, known := CompatibleRelease("03.02.00.00", "03.00.00.00")
	if !ok || !known {
		t.Fatalf("expected compatible known, got ok=%v known=%v", ok, known)
	}
	ok, known = CompatibleRelease("02.04.00.00", "03.00.00.00")
	if ok || !known {
		t.Fatalf("expected incompatible known, got ok=%v known=%v", ok, known)
	}
	ok, known = CompatibleRelease("fixture built at 0", "03.00.00.00")
	if ok || known {
		t.Fatalf("expected unknown, got ok=%v known=%v", ok, known)
	}
	ok, known = CompatibleRelease("", "03.00.00.00")
	if ok || known {
		t.Fatalf("expected unknown empty, got ok=%v known=%v", ok, known)
	}
}

func TestPreferRelease(t *testing.T) {
	if got := PreferRelease("03.02.00.00", "build"); got != "03.02.00.00" {
		t.Fatalf("got %q", got)
	}
	if got := PreferRelease("build", "03.01.00.00"); got != "03.01.00.00" {
		t.Fatalf("got %q", got)
	}
}
