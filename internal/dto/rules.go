package dto

import "time"

type CreateRuleRequest struct {
	TenantID           string              `json:"tenant_id"`
	RuleKey            string              `json:"rule_key"`
	KnowledgeTime      time.Time           `json:"knowledge_time"`
	EffectiveStartTime time.Time           `json:"effective_start_time"`
	EffectiveEndTime   time.Time           `json:"effective_end_time"`
	ResolutionVersion  int64               `json:"resolution_version"`
	InterpretorVersion int64               `json:"interpretor_version"`
	Source             string              `json:"source"`
	FactsKeys          []FactRequiredValue `json:"facts_keys"`
}

type FactRequiredValue struct {
	FactKey           string `json:"fact_key"`
	FactRequiredValue string `json:"fact_required_value"`
}
