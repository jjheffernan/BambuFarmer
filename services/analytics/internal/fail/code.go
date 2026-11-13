package fail

import (
	"fmt"
	"strconv"
)

// Code turns a printer payload print_error into a rollup key.
// Zero or missing means no failure code. Prefer Codes when HMS may be present.
func Code(payload map[string]any) string {
	codes := Codes(payload)
	if len(codes) == 0 {
		return ""
	}
	return codes[0]
}

// Codes returns distinct failure keys from print_error and hms[].
func Codes(payload map[string]any) []string {
	if payload == nil {
		return nil
	}
	out := make([]string, 0, 4)
	seen := map[string]struct{}{}
	add := func(s string) {
		if s == "" || s == "0" || s == "<nil>" {
			return
		}
		if _, ok := seen[s]; ok {
			return
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	add(printErrorCode(payload["print_error"]))
	switch v := payload["hms"].(type) {
	case []string:
		for _, s := range v {
			add(s)
		}
	case []any:
		for _, item := range v {
			switch x := item.(type) {
			case string:
				add(x)
			case map[string]any:
				add(hmsMapCode(x))
			}
		}
	}
	return out
}

func printErrorCode(v any) string {
	switch n := v.(type) {
	case float64:
		if n == 0 {
			return ""
		}
		return strconv.FormatInt(int64(n), 10)
	case int:
		if n == 0 {
			return ""
		}
		return strconv.Itoa(n)
	case int64:
		if n == 0 {
			return ""
		}
		return strconv.FormatInt(n, 10)
	case string:
		if n == "" || n == "0" {
			return ""
		}
		return n
	default:
		if n == nil {
			return ""
		}
		s := fmt.Sprint(n)
		if s == "" || s == "0" || s == "<nil>" {
			return ""
		}
		return s
	}
}

func hmsMapCode(m map[string]any) string {
	if s, ok := m["code"].(string); ok && s != "" && s != "0" {
		// already formatted key, or nested — prefer "code" string when attr absent
		if _, hasAttr := m["attr"]; !hasAttr {
			return s
		}
	}
	attr, okA := asU32(m["attr"])
	code, okC := asU32(m["code"])
	if !okA || !okC || (attr == 0 && code == 0) {
		return ""
	}
	return fmt.Sprintf("%04X_%04X_%04X_%04X",
		(attr>>16)&0xFFFF, attr&0xFFFF,
		(code>>16)&0xFFFF, code&0xFFFF)
}

func asU32(v any) (uint32, bool) {
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
	default:
		return 0, false
	}
}
