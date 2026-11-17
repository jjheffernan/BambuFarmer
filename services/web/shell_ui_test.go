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
	if !strings.Contains(layoutHTML, `id="bf-nav-toggle"`) || !strings.Contains(layoutHTML, "bf-nav-collapsed") {
		t.Fatal("expected collapsible nav toggle")
	}
	if !strings.Contains(layoutHTML, "bf-nav-group") || !strings.Contains(layoutHTML, "bf-nav-label") {
		t.Fatal("expected nav groups")
	}
	if !strings.Contains(layoutHTML, `id="admin-jump"`) || !strings.Contains(layoutHTML, "#admin-studio") {
		t.Fatal("expected Admin jump links")
	}
	if !strings.Contains(layoutHTML, "function bfBindFile") || !strings.Contains(layoutHTML, "bf-file") {
		t.Fatal("expected shared bfBindFile drop zone")
	}
	if !strings.Contains(layoutHTML, "bf-disclose") || !strings.Contains(layoutHTML, "bf-row-detail") {
		t.Fatal("expected collapsible disclose / row detail markup")
	}
	if !strings.Contains(layoutHTML, "viewport-fit=cover") {
		t.Fatal("expected viewport-fit=cover for notched phones")
	}
	if !strings.Contains(layoutHTML, `href="/static/manifest.webmanifest"`) {
		t.Fatal("expected web app manifest link")
	}
	if !strings.Contains(layoutHTML, "/static/icons/apple-touch-icon.png") {
		t.Fatal("expected apple-touch-icon")
	}
	if !strings.Contains(layoutHTML, "serviceWorker.register('/static/sw.js'") {
		t.Fatal("expected service worker registration")
	}
	if !strings.Contains(layoutHTML, "apple-mobile-web-app-capable") {
		t.Fatal("expected apple-mobile-web-app-capable meta")
	}
}
