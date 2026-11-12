package spoolman

import (
	"strconv"
	"strings"
)

// ParseSlots reads AMS_SLOT_MAP ("0=spool-a,1=spool-b").
// Empty input and bad pairs yield an empty map.
func ParseSlots(raw string) map[int]string {
	out := map[int]string{}
	for _, part := range strings.Split(raw, ",") {
		k, v, ok := strings.Cut(strings.TrimSpace(part), "=")
		if !ok {
			continue
		}
		n, err := strconv.Atoi(strings.TrimSpace(k))
		v = strings.TrimSpace(v)
		if err != nil || n < 0 || v == "" {
			continue
		}
		out[n] = v
	}
	return out
}

// SpoolFor returns the spool id for an AMS tray. A missing tray is not found.
func SpoolFor(slots map[int]string, tray int) (string, bool) {
	if len(slots) == 0 {
		return "", false
	}
	id, ok := slots[tray]
	if !ok || id == "" {
		return "", false
	}
	return id, true
}
