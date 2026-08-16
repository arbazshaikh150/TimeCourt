package repository

import (
	"context"

	"github.com/arbazshaikh150/TimeCourt/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DecisionVersionRepository interface {
	Get(ctx context.Context, decisionID uuid.UUID, version int64) (*model.DecisionVersion, error)
}

type GormDecisionVersionRepository struct{ db *gorm.DB }

func NewDecisionVersionRepository(db *gorm.DB) DecisionVersionRepository {
	return &GormDecisionVersionRepository{db: db}
}

func (r *GormDecisionVersionRepository) Get(ctx context.Context, decisionID uuid.UUID, version int64) (*model.DecisionVersion, error) {
	var decisionVersion model.DecisionVersion
	if err := r.db.WithContext(ctx).First(&decisionVersion, "decision_id = ? AND version = ?", decisionID, version).Error; err != nil {
		return nil, err
	}
	return &decisionVersion, nil
}
