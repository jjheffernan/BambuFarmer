package farmmanager

import (
	"fmt"
	"strconv"
	"strings"
)

// DefaultTestedMax is the newest Farm Manager Server release we have smoke-tested.
// Bump after a successful live FM smoke (see fixtures/fm/SMOKE_FM32.md).
// Override at runtime with FARM_MANAGER_TESTED_MAX.
const DefaultTestedMax = "03.02.00.00"

// Version status values for EdgeStatus.version_status.
const (
	VersionOK         = "ok"
	VersionBelowMin   = "below_min"
	VersionUnverified = "unverified" // newer than last smoke-tested release
	VersionUnknown    = "unknown"
	VersionNA         = "n_a"
)

// VersionReport diagnoses FM release vs min gate and tested ceiling.
type VersionReport struct {
	Got       string
	Min       string
	TestedMax string
	Status    string // ok | below_min | unverified | unknown | n_a
	OK        *bool  // compatible pointer for API (nil = unknown/n_a)
	Hint      string
}

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

// DiagnoseVersion classifies got against min and testedMax for operator UI/logs.
// Empty min → n_a (no gate). Empty/invalid got → unknown. got < min → below_min.
// got > testedMax → unverified (still compatible). Else ok.
func DiagnoseVersion(got, min, testedMax string) VersionReport {
	got = strings.TrimSpace(got)
	min = strings.TrimSpace(min)
	testedMax = strings.TrimSpace(testedMax)
	if testedMax == "" {
		testedMax = DefaultTestedMax
	}
	rep := VersionReport{Got: got, Min: min, TestedMax: testedMax}
	if min == "" {
		rep.Status = VersionNA
		rep.Hint = "FARM_MANAGER_MIN_VERSION unset — version gate disabled"
		return rep
	}
	if got == "" || !looksLikeRelease(got) {
		rep.Status = VersionUnknown
		rep.Hint = fmt.Sprintf("cannot parse FM release %q — expected dotted form like %s; check GET /captain after FM upgrade", got, DefaultTestedMax)
		return rep
	}
	if CompareRelease(got, min) < 0 {
		t := false
		rep.OK = &t
		rep.Status = VersionBelowMin
		rep.Hint = fmt.Sprintf("FM %s is below FARM_MANAGER_MIN_VERSION %s — upgrade Farm Manager Server (O&M / Studio-bypass paths assume ≥%s)", got, min, min)
		return rep
	}
	t := true
	rep.OK = &t
	if CompareRelease(got, testedMax) > 0 {
		rep.Status = VersionUnverified
		rep.Hint = fmt.Sprintf("FM %s is newer than last BF smoke-tested release %s — watch for API drift; run fixtures/fm/SMOKE_FM32.md then bump FARM_MANAGER_TESTED_MAX", got, testedMax)
		return rep
	}
	rep.Status = VersionOK
	rep.Hint = fmt.Sprintf("FM %s meets min %s (tested through %s)", got, min, testedMax)
	return rep
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
