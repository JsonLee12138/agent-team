package governance

func LoadRules(publicRules, moduleRules, taskRules []Rule) RuleLoadResult {
	return loadRulesByPriority([]namedRuleSet{
		{name: "public", rules: publicRules},
		{name: "module", rules: moduleRules},
		{name: "task", rules: taskRules},
	})
}

type namedRuleSet struct {
	name  string
	rules []Rule
}

type ruleWithSource struct {
	rule   Rule
	source string
}

func loadRulesByPriority(sets []namedRuleSet) RuleLoadResult {
	effective := make([]Rule, 0)
	seen := make(map[string]ruleWithSource)
	conflicts := make([]RuleConflict, 0)

	for _, set := range sets {
		for _, rule := range set.rules {
			existing, ok := seen[rule.Key]
			if !ok {
				seen[rule.Key] = ruleWithSource{rule: rule, source: set.name}
				effective = append(effective, rule)
				continue
			}
			if existing.rule.Value != rule.Value {
				conflicts = append(conflicts, RuleConflict{
					Key:            rule.Key,
					HigherRuleID:   existing.rule.ID,
					HigherRuleFrom: existing.source,
					LowerRuleID:    rule.ID,
					LowerRuleFrom:  set.name,
				})
			}
		}
	}

	return RuleLoadResult{Effective: effective, Conflicts: conflicts}
}
