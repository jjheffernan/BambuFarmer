package db

import (
	"strings"
	"testing"
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
