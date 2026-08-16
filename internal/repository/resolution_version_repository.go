package repository

import (
	"context"

	"github.com/arbazshaikh150/TimeCourt/internal/model"
	"gorm.io/gorm"
)

type ResolutionVersionRepository interface {
	Get(ctx context.Context, ruleKey string) (*model.ResolutionVersion, error)
}

type GormResolutionVersionRepository struct{ db *gorm.DB }

func NewResolutionVersionRepository(db *gorm.DB) ResolutionVersionRepository {
	return &GormResolutionVersionRepository{db: db}
}

func (r *GormResolutionVersionRepository) Get(ctx context.Context, ruleKey string) (*model.ResolutionVersion, error) {
	var resolutionVersion model.ResolutionVersion
	if err := r.db.WithContext(ctx).First(&resolutionVersion, "rule_key = ?", ruleKey).Error; err != nil {
		return nil, err
	}
	return &resolutionVersion, nil
}
