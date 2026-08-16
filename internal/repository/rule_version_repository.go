package repository

import (
	"context"

	"github.com/arbazshaikh150/TimeCourt/internal/model"
	"gorm.io/gorm"
)

type RuleVersionRepository interface {
	Get(ctx context.Context, ruleKey string) (*model.RuleVersion, error)
}

type GormRuleVersionRepository struct{ db *gorm.DB }

func NewRuleVersionRepository(db *gorm.DB) RuleVersionRepository {
	return &GormRuleVersionRepository{db: db}
}

func (r *GormRuleVersionRepository) Get(ctx context.Context, ruleKey string) (*model.RuleVersion, error) {
	var ruleVersion model.RuleVersion
	if err := r.db.WithContext(ctx).First(&ruleVersion, "rule_key = ?", ruleKey).Error; err != nil {
		return nil, err
	}
	return &ruleVersion, nil
}
