package tests

import (
	"os"
	"testing"

	"github.com/jjheffernan/BambuFarmer/packages/config"
)

func TestGetenv(t *testing.T) {
	t.Setenv("BF_TEST_KEY", "hello")
	if config.Getenv("BF_TEST_KEY", "fallback") != "hello" {
		t.Fatal("expected env value")
	}
	if config.Getenv("BF_TEST_MISSING_XYZ", "fallback") != "fallback" {
		t.Fatal("expected fallback")
	}
}

func TestGetenvInt(t *testing.T) {
	t.Setenv("BF_TEST_INT", "42")
	if config.GetenvInt("BF_TEST_INT", 1) != 42 {
		t.Fatal("expected 42")
	}
	t.Setenv("BF_TEST_INT", "nope")
	if config.GetenvInt("BF_TEST_INT", 7) != 7 {
		t.Fatal("expected fallback on parse error")
	}
	if config.GetenvInt("BF_TEST_INT_MISSING", 9) != 9 {
		t.Fatal("expected fallback when unset")
	}
}

// TestForceFailCanary intentionally fails when FORCE_TEST_FAIL=1 so CI/local
// can prove the test harness reports failures. Normal runs skip.
func TestForceFailCanary(t *testing.T) {
	if os.Getenv("FORCE_TEST_FAIL") != "1" {
		t.Skip("set FORCE_TEST_FAIL=1 to probe harness failure detection")
	}
	t.Fatal("intentional force-fail canary — testing infrastructure probe")
}
