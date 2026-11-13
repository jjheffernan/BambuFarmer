package main

import (
	"strings"
	"testing"
	"time"
)

func TestUtilizationSQLDefault(t *testing.T) {
	sql, args := utilizationSQL(dayWindow{}, "printer")
	if !strings.Contains(sql, "FROM agg_utilization_daily") || len(args) != 0 {
		t.Fatalf("%s %#v", sql, args)
	}
	if !strings.Contains(sql, "LIMIT 30") {
		t.Fatal(sql)
	}
}

func TestUtilizationSQLWindowAndGroup(t *testing.T) {
	win := dayWindow{
		From: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		To:   time.Date(2026, 1, 7, 0, 0, 0, 0, time.UTC),
		Set:  true,
	}
	sql, args := utilizationSQL(win, "site")
	if !strings.Contains(sql, "GROUP BY site_id") || len(args) != 2 {
		t.Fatalf("%s %#v", sql, args)
	}
	sql, args = utilizationSQL(win, "day")
	if !strings.Contains(sql, "GROUP BY day") || args[0] != "2026-01-01" {
		t.Fatalf("%s %#v", sql, args)
	}
}
