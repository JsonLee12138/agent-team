package governance

func DeclaredReferenceNotFound(taskID, moduleID, ref string) GateResult {
	return GateResult{
		Code:    "declared_reference_not_found",
		Level:   GateLevelBlocker,
		Message: "declared reference not found in index",
		Context: map[string]string{
			"task_id":   taskID,
			"module_id": moduleID,
			"ref":       ref,
		},
		NextAction: "add object to index before workflow transition",
	}
}

func RuleOverrideConflict(taskID, moduleID string, conflict RuleConflict) GateResult {
	return GateResult{
		Code:    "rule_override_conflict",
		Level:   GateLevelBlocker,
		Message: "lower-priority rule overrides higher-priority rule",
		Context: map[string]string{
			"task_id":          taskID,
			"module_id":        moduleID,
			"rule_key":         conflict.Key,
			"higher_rule_id":   conflict.HigherRuleID,
			"higher_rule_from": conflict.HigherRuleFrom,
			"lower_rule_id":    conflict.LowerRuleID,
			"lower_rule_from":  conflict.LowerRuleFrom,
		},
		NextAction: "remove lower-priority override or align values",
	}
}

func ArchivedBlocked(taskID, moduleID string) GateResult {
	return GateResult{
		Code:    "archived_blocked",
		Level:   GateLevelBlocker,
		Message: "archived input is blocked without valid exception ticket",
		Context: map[string]string{
			"task_id":   taskID,
			"module_id": moduleID,
		},
		NextAction: "owner must issue one-time read-only archived exception ticket",
	}
}
