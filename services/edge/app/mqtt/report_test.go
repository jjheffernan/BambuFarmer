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

func TestApplyParsesAMSAndTask(t *testing.T) {
	got, err := Apply(farm.DeviceReport{}, "device/ABC/report", []byte(
		`{"print":{"command":"push_status","msg":0,"ams":{"ams":[{"id":0,"tray":[{"id":2,"tray_type":"PLA","tray_color":"00AE42FF"}]}]},"task_id":"t1","subtask_name":"Benchy"}}`))
	if err != nil {
		t.Fatal(err)
	}
	if len(got.AMS) != 1 || len(got.AMS[0].Trays) != 1 || got.AMS[0].Trays[0].Slot != 2 || got.AMS[0].Trays[0].Type != "PLA" {
		t.Fatalf("%+v", got.AMS)
	}
	if got.TaskID != "t1" || got.SubtaskName != "Benchy" {
		t.Fatalf("%+v", got)
	}
}

func TestApplyParsesHMS(t *testing.T) {
	got, err := Apply(farm.DeviceReport{}, "device/ABC/report", []byte(
		`{"print":{"command":"push_status","msg":0,"gcode_state":"FAILED","hms":[{"attr":117473281,"code":65538}]}}`))
	if err != nil {
		t.Fatal(err)
	}
	if len(got.HMS) != 1 || got.HMS[0] != farm.FormatHMS(0x07008001, 0x00010002) {
		t.Fatalf("%+v", got.HMS)
	}
	got, err = Apply(got, "device/ABC/report", []byte(`{"print":{"command":"push_status","msg":1,"hms":[]}}`))
	if err != nil {
		t.Fatal(err)
	}
	if len(got.HMS) != 0 {
		t.Fatalf("expected clear, got %+v", got.HMS)
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
