package main

import (
	"strings"
	"testing"
)

func TestShellFlashAndNoHTMX(t *testing.T) {
	if !strings.Contains(layoutHTML, `id="bf-flash"`) || !strings.Contains(layoutHTML, "function bfFlash") {
		t.Fatal("expected shared bfFlash toast")
	}
	if strings.Contains(layoutHTML, "htmx.org") {
		t.Fatal("unused htmx CDN should be removed")
	}
	if !strings.Contains(layoutHTML, "bf-btn-danger") {
		t.Fatal("expected bf-btn-danger in layout")
	}
}
