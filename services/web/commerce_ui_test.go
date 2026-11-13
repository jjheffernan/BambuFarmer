package main

import (
	"strings"
	"testing"
)

func TestCommercePage(t *testing.T) {
	if !strings.Contains(layoutHTML, `href="/commerce"`) || !strings.Contains(layoutHTML, `eq .Active "commerce"`) {
		t.Fatal("expected Commerce nav and page")
	}
	if !strings.Contains(layoutHTML, "/api/v1/commerce/status") ||
		!strings.Contains(layoutHTML, "/api/v1/commerce/webhook") ||
		!strings.Contains(layoutHTML, "/api/v1/commerce/orders/status") {
		t.Fatal("expected commerce API usage")
	}
	if !strings.Contains(layoutHTML, "commerce-webhook-form") {
		t.Fatal("expected webhook test form")
	}
}
