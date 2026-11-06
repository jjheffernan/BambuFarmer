package farm

import "testing"

func TestKinds(t *testing.T) {
	for _, k := range []Kind{KindNative, KindCDP, KindFixture, KindHub, KindNone} {
		if k == "" {
			t.Fatal("empty kind")
		}
	}
	if KindNative != "fm_native" {
		t.Fatal(KindNative)
	}
}

func TestSentinelErrors(t *testing.T) {
	for _, err := range []error{ErrNotConfigured, ErrNotImplemented, ErrNotConnected, ErrUnsupportedKind} {
		if err == nil || err.Error() == "" {
			t.Fatal("bad sentinel")
		}
	}
}
