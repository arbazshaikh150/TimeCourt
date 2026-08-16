package repository

import (
	"context"

	"github.com/arbazshaikh150/TimeCourt/internal/model"
	"gorm.io/gorm"
)

type FactVersionRepository interface {
	Get(ctx context.Context, factKey, subjectID string) (*model.FactVersion, error)
}

type GormFactVersionRepository struct{ db *gorm.DB }

func NewFactVersionRepository(db *gorm.DB) FactVersionRepository {
	return &GormFactVersionRepository{db: db}
}

func (r *GormFactVersionRepository) Get(ctx context.Context, factKey, subjectID string) (*model.FactVersion, error) {
	var factVersion model.FactVersion
	if err := r.db.WithContext(ctx).First(&factVersion, "fact_key = ? AND subject_id = ?", factKey, subjectID).Error; err != nil {
		return nil, err
	}
	return &factVersion, nil
}
