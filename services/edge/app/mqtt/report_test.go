package mqtt

import (
	"testing"

	"github.com/jjheffernan/BambuFarmer/packages/farm"
)

func TestApplyFullReplacesTelemetry(t *testing.T) {
	base := farm.DeviceReport{ExternalSN: "old", Name: "Alpha", GcodeState: "IDLE", MCPercent: 90}
	got, err := Apply(base, "device/ABC/report", []byte(`{"print":{"command":"push_status","msg":0,"gcode_state":"RUNNING","mc_percent":12,"nozzle_temper":200}}`))
	if err != nil {
		t.Fatal(err)
	}
	if got.ExternalSN != "ABC" || got.Name != "Alpha" || got.GcodeState != "RUNNING" || got.MCPercent != 12 || got.NozzleTemper != 200 || !got.Online {
		t.Fatalf("%+v", got)
	}
}

func TestApplyDeltaKeepsMissingFields(t *testing.T) {
	base := farm.DeviceReport{ExternalSN: "ABC", GcodeState: "RUNNING", MCPercent: 12, NozzleTemper: 200}
	got, err := Apply(base, "/device/ABC/report/", []byte(`{"print":{"command":"push_status","msg":1,"mc_percent":40}}`))
	if err != nil {
		t.Fatal(err)
	}
	if got.GcodeState != "RUNNING" || got.NozzleTemper != 200 || got.MCPercent != 40 {
		t.Fatalf("%+v", got)
	}
}

func TestSkipOtherCommands(t *testing.T) {
	_, err := Apply(farm.DeviceReport{}, "device/ABC/report", []byte(`{"print":{"command":"push_status","msg":0}}`))
	if err != nil {
		t.Fatal(err)
	}
	if _, err = Apply(farm.DeviceReport{}, "device/ABC/report", []byte(`{"print":{"command":"gcode_line"}}`)); err == nil {
		t.Fatal("expected skip")
	}
	if _, err = Apply(farm.DeviceReport{}, "other", []byte(`{}`)); err == nil {
		t.Fatal("expected bad topic")
	}
}
