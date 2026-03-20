package governance

import "testing"

func TestLoadRules_PriorityAndConflict(t *testing.T) {
	t.Parallel()

	result := LoadRules(
		[]Rule{{ID: "p1", Key: "k1", Value: "public"}},
		[]Rule{{ID: "m1", Key: "k1", Value: "module"}},
		[]Rule{{ID: "t1", Key: "k2", Value: "task"}},
	)

	if len(result.Effective) != 2 {
		t.Fatalf("expected 2 effective rules, got %d", len(result.Effective))
	}
	if result.Effective[0].ID != "p1" {
		t.Fatalf("expected public rule to win priority")
	}
	if len(result.Conflicts) != 1 {
		t.Fatalf("expected 1 conflict, got %d", len(result.Conflicts))
	}
	if result.Conflicts[0].Key != "k1" {
		t.Fatalf("expected conflict on key k1")
	}
}
