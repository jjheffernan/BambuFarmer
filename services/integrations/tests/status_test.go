package tests

import (
	"testing"

	"github.com/jjheffernan/BambuFarmer/services/integrations/internal/status"
)

func TestURL(t *testing.T) {
	off := status.URL("")
	if off["configured"] != false {
		t.Fatalf("%v", off)
	}
	on := status.URL("http://spoolman:7912")
	if on["configured"] != true || on["url"] != "http://spoolman:7912" {
		t.Fatalf("%v", on)
	}
}
