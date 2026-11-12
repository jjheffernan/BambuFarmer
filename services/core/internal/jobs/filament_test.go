package jobs

import (
	"testing"
)

func TestPublishFilamentUsedNoop(t *testing.T) {
	if err := PublishFilamentUsed(t.Context(), nil, "default", "job", "9", 4); err != nil {
		t.Fatal(err)
	}
	if err := PublishFilamentUsed(t.Context(), nil, "default", "job", "", 4); err != nil {
		t.Fatal(err)
	}
}
