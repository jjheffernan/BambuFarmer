package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

// dayWindow is an inclusive calendar-day range (UTC dates as YYYY-MM-DD).
type dayWindow struct {
	From time.Time
	To   time.Time
	Set  bool
}

func parseDayWindow(r *http.Request) (dayWindow, error) {
	fromS := strings.TrimSpace(r.URL.Query().Get("from"))
	toS := strings.TrimSpace(r.URL.Query().Get("to"))
	if fromS == "" && toS == "" {
		return dayWindow{}, nil
	}
	if fromS == "" || toS == "" {
		return dayWindow{}, fmt.Errorf("from and to are both required (YYYY-MM-DD)")
	}
	from, err := time.Parse("2006-01-02", fromS)
	if err != nil {
		return dayWindow{}, fmt.Errorf("from: want YYYY-MM-DD")
	}
	to, err := time.Parse("2006-01-02", toS)
	if err != nil {
		return dayWindow{}, fmt.Errorf("to: want YYYY-MM-DD")
	}
	if to.Before(from) {
		return dayWindow{}, fmt.Errorf("to must be on or after from")
	}
	return dayWindow{From: from, To: to, Set: true}, nil
}

func groupBy(r *http.Request) string {
	g := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("group_by")))
	switch g {
	case "site", "day", "printer":
		return g
	default:
		return "printer"
	}
}
