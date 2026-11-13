package farm

import "testing"

func TestFormatHMS(t *testing.T) {
	// attr=0x07008001 code=0x00010002 → 0700_8001_0001_0002
	got := FormatHMS(0x07008001, 0x00010002)
	if got != "0700_8001_0001_0002" {
		t.Fatal(got)
	}
}

func TestParseHMSJSON(t *testing.T) {
	attr := uint32(0x07008001) // 117473281
	code := uint32(0x00010002) // 65538
	raw := []byte(`[{"attr":117473281,"code":65538},{"attr":117473281,"code":65538},"0700_8001_0002_0003"]`)
	got := ParseHMSJSON(raw)
	want0 := FormatHMS(attr, code)
	if len(got) != 2 {
		t.Fatalf("%v", got)
	}
	if got[0] != want0 || got[1] != "0700_8001_0002_0003" {
		t.Fatalf("%v want %s", got, want0)
	}
	if len(ParseHMSJSON([]byte(`[]`))) != 0 {
		t.Fatal("empty")
	}
	if ParseHMSJSON(nil) != nil {
		t.Fatal("nil")
	}
}
