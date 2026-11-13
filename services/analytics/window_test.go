package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestParseDayWindowEmpty(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/x", nil)
	w, err := parseDayWindow(r)
	if err != nil || w.Set {
		t.Fatalf("empty: %#v err=%v", w, err)
	}
}

func TestParseDayWindowOK(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/x?from=2026-01-01&to=2026-01-31", nil)
	w, err := parseDayWindow(r)
	if err != nil || !w.Set {
		t.Fatal(err)
	}
	if w.From.Format("2006-01-02") != "2026-01-01" || w.To.Format("2006-01-02") != "2026-01-31" {
		t.Fatalf("got %v %v", w.From, w.To)
	}
}

func TestParseDayWindowRejects(t *testing.T) {
	cases := []string{
		"/x?from=2026-01-01",
		"/x?to=2026-01-01",
		"/x?from=nope&to=2026-01-01",
		"/x?from=2026-02-01&to=2026-01-01",
	}
	for _, u := range cases {
		r := httptest.NewRequest(http.MethodGet, u, nil)
		if _, err := parseDayWindow(r); err == nil {
			t.Fatalf("expected error for %s", u)
		}
	}
}

func TestGroupBy(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/x?group_by=site", nil)
	if groupBy(r) != "site" {
		t.Fatal(groupBy(r))
	}
	r = httptest.NewRequest(http.MethodGet, "/x", nil)
	if groupBy(r) != "printer" {
		t.Fatal(groupBy(r))
	}
}
