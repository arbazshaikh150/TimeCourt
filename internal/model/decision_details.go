package model

import (
	"time"

	"github.com/arbazshaikh150/TimeCourt/internal/enums"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type DecisionDetails struct {
	DecisionID uuid.UUID `gorm:"type:uuid;primaryKey" json:"decision_id"`

	SubjectID string `gorm:"type:text;not null;index" json:"subject_id"`

	RuleKey                 string `gorm:"type:text;not null" json:"rule_key"`
	RuleUsedVersion         int64  `gorm:"not null" json:"rule_used_version"`
	ResolutionPolicyVersion int64  `gorm:"not null" json:"resolution_policy_version"`

	// JSONB data
	FactRequiredData datatypes.JSON `gorm:"type:jsonb;not null" json:"fact_required_data"`
	FactPresentData  datatypes.JSON `gorm:"type:jsonb;not null" json:"fact_present_data"`
	FailedFacts      datatypes.JSON `gorm:"type:jsonb" json:"failed_facts"`

	ResolutionPolicyRef uuid.UUID `gorm:"type:uuid;not null;index" json:"resolution_policy_ref"`

	Status enums.DecisionOutcome `gorm:"type:varchar(30);not null;index" json:"status"`

	CreatedAt time.Time `gorm:"not null;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"not null;autoUpdateTime" json:"updated_at"`
}

func (DecisionDetails) TableName() string {
	return "decision_details"
}
