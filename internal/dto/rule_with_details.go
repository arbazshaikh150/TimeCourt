package dto

import (
	"time"

	"github.com/google/uuid"
)

type RuleFetchDetails struct {
	RuleID      uuid.UUID `json:"rule_id"`
	TenantID    string    `json:"tenant_id"`
	RuleKey     string    `json:"rule_key"`
	RuleVersion int64     `json:"rule_version"`

	KnowledgeTime      time.Time `json:"knowledge_time"`
	EffectiveStartTime time.Time `json:"effective_start_time"`
	EffectiveEndTime   time.Time `json:"effective_end_time"`

	ResolutionVersion  int64 `json:"resolution_version"`
	InterpretorVersion int64 `json:"interpretor_version"`

	Source string `json:"source"`

	FactDetails []RuleFactDetails `json:"fact_details"`
}



type RuleFactDetails struct {
	FactKey           string `json:"fact_key"`
	RequiredFactValue string `json:"required_fact_value"`

	Facts []FactDetails `json:"facts"`
}



type FactDetails struct {
	FactInformationID uuid.UUID `json:"fact_information_id"`

	TenantID string `json:"tenant_id"`

	FactKey     string `json:"fact_key"`
	FactVersion int64  `json:"fact_version"`

	SubjectID string `json:"subject_id"`

	FactEffectiveStartTime time.Time `json:"fact_effective_start_time"`
	FactEffectiveEndTime   time.Time `json:"fact_effective_end_time"`

	KnowledgeTime time.Time `json:"knowledge_time"`

	FactValue  string `json:"fact_value"`
	Authority  string `json:"authority"`
	Confidence string `json:"confidence"`
	Source     string `json:"source"`
}
