package proxy

import (
	"os"
	"strconv"
	"strings"
)

// Config for LAN-only transparent TCP proxy to a real printer (not FM-bound).
type Config struct {
	Enabled    bool
	TargetIP   string
	AccessCode string // documented for operators; TLS is end-to-end to printer
	Serial     string
	Name       string
	Model      string
	ListenHost string // empty = all interfaces
	Ports      []int  // local listen ports forwarded to same port on TargetIP
	PASVStart  int
	PASVEnd    int
}

// Load reads STUDIO_BRIDGE_MODE=proxy + STUDIO_BRIDGE_PROXY_*.
func Load() Config {
	mode := strings.ToLower(strings.TrimSpace(os.Getenv("STUDIO_BRIDGE_MODE")))
	c := Config{
		Enabled:    mode == "proxy",
		TargetIP:   strings.TrimSpace(os.Getenv("STUDIO_BRIDGE_PROXY_TARGET")),
		AccessCode: getenv("STUDIO_BRIDGE_PROXY_ACCESS_CODE", ""),
		Serial:     getenv("STUDIO_BRIDGE_PROXY_SN", "BFPROXY000000000"),
		Name:       getenv("STUDIO_BRIDGE_PROXY_NAME", "LAN Proxy"),
		Model:      getenv("STUDIO_BRIDGE_PROXY_MODEL", "BL-P001"),
		ListenHost: strings.TrimSpace(os.Getenv("STUDIO_BRIDGE_PROXY_LISTEN")),
		PASVStart:  envInt("STUDIO_BRIDGE_PASV_START", 50000),
		PASVEnd:    envInt("STUDIO_BRIDGE_PASV_END", 50100),
	}
	raw := getenv("STUDIO_BRIDGE_PROXY_PORTS", "3000,3002,8883,990,6000,322")
	for _, p := range strings.Split(raw, ",") {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		n, err := strconv.Atoi(p)
		if err != nil || n <= 0 {
			continue
		}
		c.Ports = append(c.Ports, n)
	}
	return c
}

// Ready is true when proxy can start.
func (c Config) Ready() bool {
	return c.Enabled && c.TargetIP != "" && len(c.Ports) > 0
}

// StatusMap for GET /status.
func (c Config) StatusMap() map[string]any {
	return map[string]any{
		"enabled":     c.Enabled,
		"ready":       c.Ready(),
		"target_ip":   c.TargetIP,
		"serial":      c.Serial,
		"name":        c.Name,
		"model":       c.Model,
		"ports":       c.Ports,
		"pasv_range":  []int{c.PASVStart, c.PASVEnd},
		"access_code": mask(c.AccessCode),
		"note":        "LAN/direct_printer only — do not point at FM-bound printers (no :8883). Transparent TCP relay; slicer uses printer access code.",
	}
}

func mask(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	if len(s) <= 2 {
		return "**"
	}
	return s[:1] + strings.Repeat("*", len(s)-2) + s[len(s)-1:]
}

func getenv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

func envInt(k string, d int) int {
	v := strings.TrimSpace(os.Getenv(k))
	if v == "" {
		return d
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return d
	}
	return n
}
