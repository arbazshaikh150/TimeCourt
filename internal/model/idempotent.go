package model

import (
	"time"

	"github.com/arbazshaikh150/TimeCourt/internal/enums"
	"github.com/google/uuid"
)

type Idempotent struct {
	IdempotentKey uuid.UUID `gorm:"type:uuid;primaryKey" json:"idempotent_key"`

	TenantID string `gorm:"type:text;index" json:"tenant_id"`
	Source   string `gorm:"type:text" json:"source"`

	RequestDigest *string `gorm:"type:text" json:"request_digest"`

	ResourceID       *uuid.UUID `gorm:"type:uuid;index" json:"resource_id"`
	ResourceValueRef uuid.UUID `gorm:"type:uuid" json:"resource_value"`

	LockedTime    time.Time              `gorm:"type:timestamptz;default:CURRENT_TIMESTAMP" json:"time"`
	CommittedTime *time.Time             `gorm:"type:timestamptz" json:"committed_time"`
	Status        enums.IdempotentStatus `gorm:"type:varchar(30);not null;index" json:"status"`
}

func (Idempotent) TableName() string {
	return "idempotency_records"
}
