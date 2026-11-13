package farm

import (
	"encoding/json"
	"fmt"
)

// FormatHMS turns printer attr/code integers into the common XXXX_XXXX_XXXX_XXXX key.
func FormatHMS(attr, code uint32) string {
	return fmt.Sprintf("%04X_%04X_%04X_%04X",
		(attr>>16)&0xFFFF, attr&0xFFFF,
		(code>>16)&0xFFFF, code&0xFFFF)
}

// ParseHMSJSON reads a push_status "hms" array of {attr,code} objects or preformatted strings.
func ParseHMSJSON(raw json.RawMessage) []string {
	if len(raw) == 0 || string(raw) == "null" {
		return nil
	}
	var items []json.RawMessage
	if err := json.Unmarshal(raw, &items); err != nil {
		return nil
	}
	out := make([]string, 0, len(items))
	seen := map[string]struct{}{}
	for _, item := range items {
		code := hmsItemCode(item)
		if code == "" {
			continue
		}
		if _, ok := seen[code]; ok {
			continue
		}
		seen[code] = struct{}{}
		out = append(out, code)
	}
	return out
}

func hmsItemCode(raw json.RawMessage) string {
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		if s == "" || s == "0" {
			return ""
		}
		return s
	}
	var e struct {
		Attr any `json:"attr"`
		Code any `json:"code"`
	}
	if err := json.Unmarshal(raw, &e); err != nil {
		return ""
	}
	attr, okA := asUint32(e.Attr)
	code, okC := asUint32(e.Code)
	if !okA || !okC {
		return ""
	}
	if attr == 0 && code == 0 {
		return ""
	}
	return FormatHMS(attr, code)
}

func asUint32(v any) (uint32, bool) {
	switch n := v.(type) {
	case float64:
		if n < 0 {
			return 0, false
		}
		return uint32(n), true
	case int:
		if n < 0 {
			return 0, false
		}
		return uint32(n), true
	case int64:
		if n < 0 {
			return 0, false
		}
		return uint32(n), true
	case json.Number:
		i, err := n.Int64()
		if err != nil || i < 0 {
			return 0, false
		}
		return uint32(i), true
	default:
		return 0, false
	}
}
