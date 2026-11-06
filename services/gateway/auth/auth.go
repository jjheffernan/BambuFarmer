package auth

// NormalizeMode maps legacy aliases to canonical AUTH_MODE values.
func NormalizeMode(mode string) string {
	if mode == "local_insecure" {
		return "lan"
	}
	return mode
}

// LanWarning returns a boot/UI warning for insecure LAN mode.
func LanWarning(mode string) string {
	if mode == "lan" {
		return "insecure LAN mode — do not expose publicly"
	}
	return ""
}
