package vp

import (
	"os"
	"strings"
)

// Config is one virtual printer identity for Studio/Orca LAN receive.
type Config struct {
	Enabled    bool
	Serial     string
	Name       string
	Model      string
	AccessCode string
	BindAddrs  []string // e.g. :3000,:3002
	FTPSAddr   string   // e.g. :990
	MQTTAddr   string   // e.g. :8883
	PASVStart  int
	PASVEnd    int
	CertDir    string
}

// LoadConfig reads STUDIO_BRIDGE_VP_* env (disabled unless STUDIO_BRIDGE_VP=1/true).
func LoadConfig() Config {
	c := Config{
		Enabled:    envBool("STUDIO_BRIDGE_VP", false),
		Serial:     getenv("STUDIO_BRIDGE_VP_SN", "BFVP00000000000"),
		Name:       getenv("STUDIO_BRIDGE_VP_NAME", "BambuFarmer VP"),
		Model:      getenv("STUDIO_BRIDGE_VP_MODEL", "BL-P001"),
		AccessCode: getenv("STUDIO_BRIDGE_VP_ACCESS_CODE", "12345678"),
		FTPSAddr:   getenv("STUDIO_BRIDGE_FTPS_ADDR", ":990"),
		MQTTAddr:   getenv("STUDIO_BRIDGE_MQTT_ADDR", ":8883"),
		PASVStart:  envInt("STUDIO_BRIDGE_PASV_START", 50000),
		PASVEnd:    envInt("STUDIO_BRIDGE_PASV_END", 50009),
		CertDir:    getenv("STUDIO_BRIDGE_CERT_DIR", "/data/studio-bridge/certs"),
	}
	raw := getenv("STUDIO_BRIDGE_BIND_ADDRS", ":3000,:3002")
	for _, p := range strings.Split(raw, ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			c.BindAddrs = append(c.BindAddrs, p)
		}
	}
	return c
}

// StatusMap is JSON-friendly status for GET /status.
func (c Config) StatusMap() map[string]any {
	return map[string]any{
		"enabled":     c.Enabled,
		"serial":      c.Serial,
		"name":        c.Name,
		"model":       c.Model,
		"bind_addrs":  c.BindAddrs,
		"ftps_addr":   c.FTPSAddr,
		"mqtt_addr":   c.MQTTAddr,
		"pasv_range":  []int{c.PASVStart, c.PASVEnd},
		"cert_dir":    c.CertDir,
		"access_code": maskCode(c.AccessCode),
	}
}

func maskCode(s string) string {
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

func envBool(k string, d bool) bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv(k)))
	if v == "" {
		return d
	}
	return v == "1" || v == "true" || v == "yes" || v == "on"
}

func envInt(k string, d int) int {
	v := strings.TrimSpace(os.Getenv(k))
	if v == "" {
		return d
	}
	n := 0
	for _, ch := range v {
		if ch < '0' || ch > '9' {
			return d
		}
		n = n*10 + int(ch-'0')
	}
	return n
}
