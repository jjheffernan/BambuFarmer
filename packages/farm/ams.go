package farm

import (
	"encoding/json"
	"strconv"
)

// AMSUnit is one AMS / AMS HT / external hub on a printer.
type AMSUnit struct {
	ID    int       `json:"id"`
	Trays []AMSTray `json:"trays"`
}

// AMSTray is one slot in an AMS unit.
type AMSTray struct {
	Slot   int    `json:"slot"`
	Type   string `json:"type,omitempty"`
	Color  string `json:"color,omitempty"`
	TagUID string `json:"tag_uid,omitempty"`
	Remain int    `json:"remain,omitempty"` // 0–100 when the printer reports it
}

// ParseAMSJSON normalizes a push_status "ams" object or array into AMSUnit slices.
func ParseAMSJSON(raw json.RawMessage) []AMSUnit {
	if len(raw) == 0 || string(raw) == "null" {
		return nil
	}
	// Nested: {"ams":[...],"ams_exist_bits":...}
	var wrap struct {
		AMS json.RawMessage `json:"ams"`
	}
	if err := json.Unmarshal(raw, &wrap); err == nil && len(wrap.AMS) > 0 {
		raw = wrap.AMS
	}
	var units []json.RawMessage
	if err := json.Unmarshal(raw, &units); err != nil {
		return nil
	}
	out := make([]AMSUnit, 0, len(units))
	for _, u := range units {
		unit, ok := parseAMSUnit(u)
		if !ok {
			continue
		}
		out = append(out, unit)
	}
	return out
}

func parseAMSUnit(raw json.RawMessage) (AMSUnit, bool) {
	var head struct {
		ID   any             `json:"id"`
		Tray json.RawMessage `json:"tray"`
	}
	if err := json.Unmarshal(raw, &head); err != nil {
		return AMSUnit{}, false
	}
	id := anyInt(head.ID)
	var trays []AMSTray
	if len(head.Tray) > 0 {
		var items []json.RawMessage
		if err := json.Unmarshal(head.Tray, &items); err == nil {
			for _, item := range items {
				if t, ok := parseAMSTray(item); ok {
					trays = append(trays, t)
				}
			}
		}
	}
	if trays == nil {
		trays = []AMSTray{}
	}
	return AMSUnit{ID: id, Trays: trays}, true
}

func parseAMSTray(raw json.RawMessage) (AMSTray, bool) {
	var t struct {
		ID        any    `json:"id"`
		TrayType  string `json:"tray_type"`
		Type      string `json:"type"`
		TrayColor string `json:"tray_color"`
		Color     string `json:"color"`
		TagUID    string `json:"tag_uid"`
		Remain    any    `json:"remain"`
		RemainPct any    `json:"remain_pct"`
	}
	if err := json.Unmarshal(raw, &t); err != nil {
		return AMSTray{}, false
	}
	typ := t.TrayType
	if typ == "" {
		typ = t.Type
	}
	color := t.TrayColor
	if color == "" {
		color = t.Color
	}
	remain := anyInt(t.Remain)
	if remain == 0 {
		remain = anyInt(t.RemainPct)
	}
	return AMSTray{
		Slot:   anyInt(t.ID),
		Type:   typ,
		Color:  color,
		TagUID: t.TagUID,
		Remain: remain,
	}, true
}

func anyInt(v any) int {
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	case int64:
		return int(n)
	case string:
		i, _ := strconv.Atoi(n)
		return i
	case json.Number:
		i, _ := n.Int64()
		return int(i)
	default:
		return 0
	}
}
