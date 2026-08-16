package repository

import (
	"context"

	"github.com/arbazshaikh150/TimeCourt/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RuleTableRepository interface {
	Get(ctx context.Context, ruleID uuid.UUID) (*model.RuleTable, error)
}

type GormRuleTableRepository struct{ db *gorm.DB }

func NewRuleTableRepository(db *gorm.DB) RuleTableRepository {
	return &GormRuleTableRepository{db: db}
}

func (r *GormRuleTableRepository) Get(ctx context.Context, ruleID uuid.UUID) (*model.RuleTable, error) {
	var rule model.RuleTable
	if err := r.db.WithContext(ctx).First(&rule, "rule_id = ?", ruleID).Error; err != nil {
		return nil, err
	}
	return &rule, nil
}
