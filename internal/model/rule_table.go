package model

import (
	"time"

	"github.com/google/uuid"
)

// TODO : Effective timestamp is pending
type RuleTable struct {
	RuleID   uuid.UUID `gorm:"type:uuid;primaryKey" json:"rule_id"`
	TenantID string    `gorm:"type:text;not null;index" json:"tenant_id"`

	RuleKey     string `gorm:"type:text;not null;index" json:"rule_key"`
	RuleVersion int64  `gorm:"not null" json:"rule_version"`

	KnowledgeTime      time.Time `gorm:"not null;index" json:"knowledge_time"`
	EffectiveStartTime time.Time `gorm:"not null" json:"effective_start_time"`
	EffectiveEndTime   time.Time `gorm:"not null" json:"effective_end_time"`

	ResolutionVersion  int64 `gorm:"not null" json:"resolution_version"`
	InterpretorVersion int64 `gorm:"not null" json:"interpretor_version"`

	Source string `gorm:"type:text;not null" json:"source"`
}

func (RuleTable) TableName() string {
	return "rules"
}
