package jobs

import (
	"testing"

	"github.com/jjheffernan/BambuFarmer/packages/db"
)

func TestPublishFilamentUsedNoop(t *testing.T) {
	if err := PublishFilamentUsed(t.Context(), nil, "default", "job", "9", 4); err != nil {
		t.Fatal(err)
	}
	if err := PublishFilamentUsed(t.Context(), nil, "default", "job", "", 4); err != nil {
		t.Fatal(err)
	}
}

func TestPublishJobCostedNoop(t *testing.T) {
	if err := PublishJobCosted(t.Context(), nil, "default", "job", "sn", 60, db.CostParts{Total: 1}, "USD"); err != nil {
		t.Fatal(err)
	}
}
