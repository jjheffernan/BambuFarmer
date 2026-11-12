package db

import "testing"

func TestNudgeQueueIndex(t *testing.T) {
	if j, ok := NudgeQueueIndex(0, 3, true); ok || j != 0 {
		t.Fatalf("front earlier: %d %v", j, ok)
	}
	if j, ok := NudgeQueueIndex(2, 3, false); ok || j != 2 {
		t.Fatalf("back later: %d %v", j, ok)
	}
	if j, ok := NudgeQueueIndex(1, 3, true); !ok || j != 0 {
		t.Fatalf("earlier: %d %v", j, ok)
	}
	if j, ok := NudgeQueueIndex(1, 3, false); !ok || j != 2 {
		t.Fatalf("later: %d %v", j, ok)
	}
}
