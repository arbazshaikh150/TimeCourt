package repository

import (
	"context"

	"github.com/arbazshaikh150/TimeCourt/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DecisionRecordRepository interface {
	Get(ctx context.Context, decisionID uuid.UUID) (*model.DecisionRecord, error)
}

type GormDecisionRecordRepository struct{ db *gorm.DB }

func NewDecisionRecordRepository(db *gorm.DB) DecisionRecordRepository {
	return &GormDecisionRecordRepository{db: db}
}

func (r *GormDecisionRecordRepository) Get(ctx context.Context, decisionID uuid.UUID) (*model.DecisionRecord, error) {
	var record model.DecisionRecord
	if err := r.db.WithContext(ctx).First(&record, "decision_id = ?", decisionID).Error; err != nil {
		return nil, err
	}
	return &record, nil
}
