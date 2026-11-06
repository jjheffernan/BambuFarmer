package status

// URL reports whether an optional integration base URL is configured.
func URL(u string) map[string]any {
	if u == "" {
		return map[string]any{"configured": false}
	}
	return map[string]any{"configured": true, "url": u}
}
