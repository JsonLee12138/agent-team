package governance

import "testing"

func TestEvaluateGate_DeclaredReferenceNotFound(t *testing.T) {
	t.Parallel()

	result := EvaluateGate(GateInput{
		TaskPacket: TaskPacket{
			TaskID:             "task-1",
			ModuleID:           "task",
			DeclaredReferences: []string{"req-1"},
		},
		Index: Index{Entries: []IndexEntry{{ID: "task-1"}}},
	})

	if result.Code != "declared_reference_not_found" {
		t.Fatalf("expected declared_reference_not_found, got %s", result.Code)
	}
	if result.Level != GateLevelBlocker {
		t.Fatalf("expected blocker level")
	}
}

func TestEvaluateGate_RuleOverrideConflict(t *testing.T) {
	t.Parallel()

	result := EvaluateGate(GateInput{
		TaskPacket: TaskPacket{TaskID: "task-1", ModuleID: "task"},
		Index:      Index{Entries: []IndexEntry{{ID: "task-1"}}},
		LoadedRules: RuleLoadResult{Conflicts: []RuleConflict{{
			Key:            "k1",
			HigherRuleID:   "p1",
			HigherRuleFrom: "public",
			LowerRuleID:    "m1",
			LowerRuleFrom:  "module",
		}}},
	})

	if result.Code != "rule_override_conflict" {
		t.Fatalf("expected rule_override_conflict, got %s", result.Code)
	}
}

func TestEvaluateGate_ArchivedBlocked(t *testing.T) {
	t.Parallel()

	result := EvaluateGate(GateInput{
		TaskPacket: TaskPacket{TaskID: "task-1", ModuleID: "task", UsesArchivedInput: true},
		Index:      Index{Entries: []IndexEntry{{ID: "task-1"}}},
	})

	if result.Code != "archived_blocked" {
		t.Fatalf("expected archived_blocked, got %s", result.Code)
	}
}

func TestEvaluateGate_Pass(t *testing.T) {
	t.Parallel()

	ticket := NewArchivedExceptionTicket("tk-1", "task-1", "owner-1", "debug", TimeNowUTC())
	result := EvaluateGate(GateInput{
		TaskPacket: TaskPacket{
			TaskID:             "task-1",
			ModuleID:           "task",
			Owner:              "owner-1",
			UsesArchivedInput:  true,
			DeclaredReferences: []string{"req-1"},
		},
		Index: Index{Entries: []IndexEntry{{ID: "task-1"}, {ID: "req-1"}}},
		LoadedRules: RuleLoadResult{
			Effective: []Rule{{ID: "p1", Key: "k1", Value: "v1"}},
		},
		ArchivedTicket: &ticket,
	})

	if result.Code != "ok" {
		t.Fatalf("expected ok, got %s", result.Code)
	}
}
