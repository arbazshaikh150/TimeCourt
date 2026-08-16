package repository

import (
	"context"

	"github.com/arbazshaikh150/TimeCourt/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DecisionDetailsRepository interface {
	Get(ctx context.Context, decisionID uuid.UUID) (*model.DecisionDetails, error)
}

type GormDecisionDetailsRepository struct{ db *gorm.DB }

func NewDecisionDetailsRepository(db *gorm.DB) DecisionDetailsRepository {
	return &GormDecisionDetailsRepository{db: db}
}

func (r *GormDecisionDetailsRepository) Get(ctx context.Context, decisionID uuid.UUID) (*model.DecisionDetails, error) {
	var details model.DecisionDetails
	if err := r.db.WithContext(ctx).First(&details, "decision_id = ?", decisionID).Error; err != nil {
		return nil, err
	}
	return &details, nil
}
