package repository

import (
	"context"

	"github.com/arbazshaikh150/TimeCourt/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RuleDetailRepository interface {
	Get(ctx context.Context, ruleID uuid.UUID) (*model.RuleDetail, error)
}

type GormRuleDetailRepository struct{ db *gorm.DB }

func NewRuleDetailRepository(db *gorm.DB) RuleDetailRepository {
	return &GormRuleDetailRepository{db: db}
}

func (r *GormRuleDetailRepository) Get(ctx context.Context, ruleID uuid.UUID) (*model.RuleDetail, error) {
	var detail model.RuleDetail
	if err := r.db.WithContext(ctx).First(&detail, "rule_id = ?", ruleID).Error; err != nil {
		return nil, err
	}
	return &detail, nil
}
