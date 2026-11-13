package tests

import (
	"testing"

	"github.com/jjheffernan/BambuFarmer/packages/farm"
)

func TestKinds(t *testing.T) {
	for _, k := range []farm.Kind{farm.KindNative, farm.KindCDP, farm.KindFixture, farm.KindDirect, farm.KindHub, farm.KindNone} {
		if k == "" {
			t.Fatal("empty kind")
		}
	}
	if farm.KindNative != "fm_native" {
		t.Fatal(farm.KindNative)
	}
}

func TestSentinelErrors(t *testing.T) {
	for _, err := range []error{farm.ErrNotConfigured, farm.ErrNotImplemented, farm.ErrNotConnected, farm.ErrUnsupportedKind} {
		if err == nil || err.Error() == "" {
			t.Fatal("bad sentinel")
		}
	}
}
