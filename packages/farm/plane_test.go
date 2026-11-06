package farm_test

import (
	"testing"

	"github.com/jjheffernan/BambuFarmer/packages/farm"
)

func TestKindConstants(t *testing.T) {
	if farm.KindNative != "fm_native" {
		t.Fatal(farm.KindNative)
	}
	if farm.ErrNotImplemented == nil {
		t.Fatal("expected sentinel")
	}
}
