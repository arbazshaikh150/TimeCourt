package model

import "github.com/google/uuid"

type DecisionVersion struct {
	DecisionID uuid.UUID `gorm:"type:uuid;primaryKey" json:"decision_id"`
	Version    int64     `gorm:"primaryKey" json:"version"`
}

func (DecisionVersion) TableName() string {
	return "decision_versions"
}