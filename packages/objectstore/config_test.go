package objectstore

import (
	"testing"
)

func TestConfigDisabled(t *testing.T) {
	t.Setenv("OBJECT_STORE_ENDPOINT", "")
	if FromEnv().Enabled() {
		t.Fatal("expected disabled")
	}
}

func TestConfigEnabled(t *testing.T) {
	t.Setenv("OBJECT_STORE_ENDPOINT", "http://minio:9000")
	t.Setenv("OBJECT_STORE_ACCESS_KEY", "minio")
	t.Setenv("OBJECT_STORE_SECRET_KEY", "secret")
	t.Setenv("OBJECT_STORE_BUCKET", "farm")
	cfg := FromEnv()
	if !cfg.Enabled() || cfg.Bucket != "farm" || cfg.Secure {
		t.Fatalf("%+v", cfg)
	}
}

func TestParseEndpoint(t *testing.T) {
	h, sec, err := parseEndpoint("http://minio:9000", false)
	if err != nil || h != "minio:9000" || sec {
		t.Fatalf("%s %v %v", h, sec, err)
	}
	h, sec, err = parseEndpoint("minio:9000", true)
	if err != nil || h != "minio:9000" || !sec {
		t.Fatalf("%s %v %v", h, sec, err)
	}
}

func TestSanitizeKey(t *testing.T) {
	if _, err := sanitizeKey("../x"); err == nil {
		t.Fatal("expected reject")
	}
	k, err := sanitizeKey("a/b.gcode.3mf")
	if err != nil || k != "a/b.gcode.3mf" {
		t.Fatal(k, err)
	}
}

func TestDialDisabled(t *testing.T) {
	c, err := Dial(Config{})
	if err != nil || c != nil {
		t.Fatalf("%v %v", c, err)
	}
}
