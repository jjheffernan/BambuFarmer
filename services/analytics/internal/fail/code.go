package fail

import (
	"fmt"
	"strconv"
)

// Code turns a printer payload print_error into a rollup key.
// Zero or missing means no failure code.
func Code(payload map[string]any) string {
	if payload == nil {
		return ""
	}
	switch v := payload["print_error"].(type) {
	case float64:
		if v == 0 {
			return ""
		}
		return strconv.FormatInt(int64(v), 10)
	case int:
		if v == 0 {
			return ""
		}
		return strconv.Itoa(v)
	case int64:
		if v == 0 {
			return ""
		}
		return strconv.FormatInt(v, 10)
	case string:
		if v == "" || v == "0" {
			return ""
		}
		return v
	default:
		if v == nil {
			return ""
		}
		s := fmt.Sprint(v)
		if s == "" || s == "0" || s == "<nil>" {
			return ""
		}
		return s
	}
}
