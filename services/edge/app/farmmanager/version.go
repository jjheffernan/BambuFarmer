package farmmanager

import (
	"strconv"
	"strings"
)

// CompareRelease compares dotted FM release strings (e.g. "03.02.00.00").
// Returns -1 if a < b, 0 if equal, 1 if a > b. Non-numeric segments sort as 0.
func CompareRelease(a, b string) int {
	as := splitRelease(a)
	bs := splitRelease(b)
	n := len(as)
	if len(bs) > n {
		n = len(bs)
	}
	for i := 0; i < n; i++ {
		av, bv := 0, 0
		if i < len(as) {
			av = as[i]
		}
		if i < len(bs) {
			bv = bs[i]
		}
		if av < bv {
			return -1
		}
		if av > bv {
			return 1
		}
	}
	return 0
}

// CompatibleRelease reports whether got meets or exceeds min (dotted release).
// Empty got → unknown (ok=false, known=false). Empty min → always compatible.
func CompatibleRelease(got, min string) (ok bool, known bool) {
	got = strings.TrimSpace(got)
	min = strings.TrimSpace(min)
	if min == "" {
		return true, got != ""
	}
	if got == "" || !looksLikeRelease(got) {
		return false, false
	}
	return CompareRelease(got, min) >= 0, true
}

// PreferRelease picks the best dotted release from captain release/version fields.
func PreferRelease(release, version string) string {
	if looksLikeRelease(release) {
		return strings.TrimSpace(release)
	}
	if looksLikeRelease(version) {
		return strings.TrimSpace(version)
	}
	if strings.TrimSpace(release) != "" {
		return strings.TrimSpace(release)
	}
	return strings.TrimSpace(version)
}

func looksLikeRelease(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" || !strings.Contains(s, ".") {
		return false
	}
	parts := strings.Split(s, ".")
	if len(parts) < 2 {
		return false
	}
	digits := 0
	for _, p := range parts {
		if _, err := strconv.Atoi(p); err == nil {
			digits++
		}
	}
	return digits >= 2
}

func splitRelease(s string) []int {
	parts := strings.Split(strings.TrimSpace(s), ".")
	out := make([]int, 0, len(parts))
	for _, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil {
			n = 0
		}
		out = append(out, n)
	}
	return out
}
