package bambu

import (
	"testing"

	"github.com/jjheffernan/BambuFarmer/packages/farm"
	"github.com/jjheffernan/BambuFarmer/services/edge/app/config"
)

func TestOpenDirectIgnoresFarmURL(t *testing.T) {
	p, err := Open(config.Config{
		FarmManagerTransport: "direct_printer",
		FarmManagerURL:       "http://farm.example",
		DirectPrinters:       "abc=Alpha",
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if p.Kind() != farm.KindDirect {
		t.Fatal(p.Kind())
	}
	devs, err := p.ListDevices(t.Context())
	if err != nil || len(devs) != 1 || devs[0].Name != "Alpha" {
		t.Fatalf("%v %+v", err, devs)
	}
}
