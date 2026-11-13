package db

import (
	"testing"
)

func TestHydratePrinterSnapshotUsesSnapshotExtras(t *testing.T) {
	raw := `{
		"external_sn":"SN1","name":"Bay","gcode_state":"RUNNING",
		"nozzle_temper":220.5,"bed_temper":55,
		"hms":["0700_8001_0001_0002"],
		"ams":[{"id":0,"trays":[{"slot":1,"type":"PLA","color":"00AE42FF","remain":80}]}],
		"task":{"external_task_id":"t1","subtask_name":"Benchy"},
		"tags":["circuit-a"],
		"bucket":"printing"
	}`
	got := HydratePrinterSnapshot("default:SN1", "default", "SN1", "Bay-1", "P1S", "PAUSE", true, 10, 60, raw)
	if got.Name != "Bay-1" || got.GcodeState != "PAUSE" || got.MCPercent != 10 {
		t.Fatalf("columns should win: %+v", got.PrinterSummary)
	}
	if got.NozzleTemper != 220.5 || got.BedTemper != 55 {
		t.Fatalf("temps %+v", got)
	}
	if len(got.HMS) != 1 || got.HMS[0] != "0700_8001_0001_0002" {
		t.Fatalf("hms %+v", got.HMS)
	}
	if len(got.AMS) != 1 || len(got.AMS[0].Trays) != 1 || got.AMS[0].Trays[0].Type != "PLA" {
		t.Fatalf("ams %+v", got.AMS)
	}
	if got.Task == nil || got.Task.SubtaskName != "Benchy" {
		t.Fatalf("task %+v", got.Task)
	}
	if len(got.Tags) != 1 || got.Tags[0] != "circuit-a" {
		t.Fatalf("tags %+v", got.Tags)
	}
	if got.Bucket != "printing" {
		t.Fatalf("bucket %q", got.Bucket)
	}
}

func TestHydratePrinterSnapshotEmptyRaw(t *testing.T) {
	got := HydratePrinterSnapshot("id", "default", "SN", "N", "M", "IDLE", true, 0, 0, "")
	if got.ExternalSN != "SN" || got.Name != "N" || len(got.AMS) != 0 {
		t.Fatalf("%+v", got)
	}
}
