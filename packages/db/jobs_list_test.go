package db

import (
	"strings"
	"testing"
	"time"
)

func TestJobListWhereEmpty(t *testing.T) {
	where, args := jobListWhere("default", JobListFilter{})
	if where != ` WHERE site_id=$1` || len(args) != 1 || args[0] != "default" {
		t.Fatalf("%q %#v", where, args)
	}
}

func TestJobListWhereStateAndQuery(t *testing.T) {
	where, args := jobListWhere("default", JobListFilter{Query: `100%_plate`, State: "succeeded"})
	if !strings.Contains(where, `state=$2`) || strings.Count(where, `ILIKE $3 ESCAPE '\'`) != 4 {
		t.Fatal(where)
	}
	if args[1] != "succeeded" || args[2] != `%100\%\_plate%` {
		t.Fatalf("%#v", args)
	}
}

func TestJobListWhereIgnoresAllState(t *testing.T) {
	where, args := jobListWhere("default", JobListFilter{State: "all"})
	if strings.Contains(where, "state=") || len(args) != 1 {
		t.Fatalf("%q %#v", where, args)
	}
}

func TestJobListWhereSNAndTime(t *testing.T) {
	from := mustParseTime(t, "2026-01-01T00:00:00Z")
	to := mustParseTime(t, "2026-02-01T00:00:00Z")
	where, args := jobListWhere("default", JobListFilter{ExternalSN: "SN1", From: &from, To: &to})
	if !strings.Contains(where, `external_sn=$2`) || !strings.Contains(where, `>= $3`) || !strings.Contains(where, `< $4`) {
		t.Fatal(where)
	}
	if args[1] != "SN1" {
		t.Fatalf("%#v", args)
	}
}

func mustParseTime(t *testing.T, s string) time.Time {
	t.Helper()
	ts, err := time.Parse(time.RFC3339, s)
	if err != nil {
		t.Fatal(err)
	}
	return ts
}
