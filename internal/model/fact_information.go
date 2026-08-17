package model

import (
	"time"

	"github.com/google/uuid"
)

type FactInformation struct {
	FactInformationID uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`

	TenantID string `gorm:"type:text;not null;index" json:"tenant_id"`

	FactKey     string `gorm:"type:text;not null;index" json:"fact_key"`
	FactVersion int64  `gorm:"not null" json:"fact_version"`

	SubjectID string `gorm:"type:text;not null;index" json:"subject_id"`

	FactEffectiveStartTime time.Time `gorm:"not null" json:"fact_effective_start_time"`
	FactEffectiveEndTime   time.Time `gorm:"not null" json:"fact_effective_end_time"`

	KnowledgeTime time.Time `gorm:"not null;index" json:"knowledge_time"`

	FactValue  string `gorm:"type:text;not null" json:"fact_value"`
	Authority  string `gorm:"type:text;not null" json:"authority"`
	Confidence string `gorm:"type:text" json:"confidence"`
	Source     string `gorm:"type:text;not null" json:"source"`

	EffectivePeriod string `gorm:"type:tstzrange;->" json:"effective_period"`
}

func (FactInformation) TableName() string {
	return "fact_information"
}
