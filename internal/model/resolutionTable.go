package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type ResolutionTable struct {
	ResolutionID uuid.UUID `gorm:"type:uuid;primaryKey" json:"resolution_id"`

	RuleKey  string `gorm:"type:text;not null;index" json:"rule_key"`
	Version  int64  `gorm:"not null" json:"version"`
	TenantID string `gorm:"type:text;not null;index" json:"tenant_id"`

	CurrentTime time.Time `gorm:"not null;autoCreateTime" json:"current_time"`

	// PostgreSQL JSONB
	ResolutionPolicy datatypes.JSON `gorm:"type:jsonb;not null" json:"resolution_policy"`
}

func (ResolutionTable) TableName() string {
	return "resolution_tables"
}
