package model

import "github.com/google/uuid"

type RuleDetail struct {
	RuleID            uuid.UUID `gorm:"type:uuid;primaryKey;not null" json:"rule_id"`
	FactKey           string    `gorm:"type:text;primaryKey;not null;index" json:"fact_key"`
	FactRequiredValue string    `gorm:"type:text;primaryKey;not null" json:"fact_required_value"`
}

func (RuleDetail) TableName() string {
	return "rule_details"
}
