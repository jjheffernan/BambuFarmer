package farm

import (
	"testing"
)

func TestParseAMSJSONNested(t *testing.T) {
	raw := []byte(`{"ams":[{"id":"0","tray":[{"id":"1","tray_type":"PLA","tray_color":"00AE42FF","tag_uid":"ABC","remain":80}]}]}`)
	got := ParseAMSJSON(raw)
	if len(got) != 1 || got[0].ID != 0 || len(got[0].Trays) != 1 {
		t.Fatalf("%+v", got)
	}
	tr := got[0].Trays[0]
	if tr.Slot != 1 || tr.Type != "PLA" || tr.Color != "00AE42FF" || tr.TagUID != "ABC" || tr.Remain != 80 {
		t.Fatalf("%+v", tr)
	}
}

func TestParseAMSJSONArray(t *testing.T) {
	raw := []byte(`[{"id":0,"tray":[{"id":0,"type":"PETG","color":"FF0000FF"}]}]`)
	got := ParseAMSJSON(raw)
	if len(got) != 1 || got[0].Trays[0].Type != "PETG" {
		t.Fatalf("%+v", got)
	}
}

func TestParseAMSJSONEmpty(t *testing.T) {
	if ParseAMSJSON(nil) != nil {
		t.Fatal("nil")
	}
	if len(ParseAMSJSON([]byte(`[]`))) != 0 {
		t.Fatal("empty")
	}
}
