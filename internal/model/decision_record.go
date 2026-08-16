package model

import (
	"time"

	"github.com/google/uuid"
)

type DecisionRecord struct {
	DecisionID       uuid.UUID `gorm:"type:uuid;primaryKey" json:"decision_id"`
	TenantID         string    `gorm:"type:text;not null;index" json:"tenant_id"`
	TimeOfCompletion time.Time `gorm:"not null" json:"time_of_completion"`
	TimeWhenToCheck  time.Time `gorm:"not null;index" json:"time_when_to_check"`
	Version          int64     `gorm:"not null" json:"version"`

	CreatedAt time.Time `gorm:"not null;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"not null;autoUpdateTime" json:"updated_at"`
}

func (DecisionRecord) TableName() string {
	return "decision_records"
}
