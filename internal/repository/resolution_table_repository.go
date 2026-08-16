package repository

import (
	"context"

	"github.com/arbazshaikh150/TimeCourt/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ResolutionTableRepository interface {
	Get(ctx context.Context, resolutionID uuid.UUID) (*model.ResolutionTable, error)
}

type GormResolutionTableRepository struct{ db *gorm.DB }

func NewResolutionTableRepository(db *gorm.DB) ResolutionTableRepository {
	return &GormResolutionTableRepository{db: db}
}

func (r *GormResolutionTableRepository) Get(ctx context.Context, resolutionID uuid.UUID) (*model.ResolutionTable, error) {
	var resolution model.ResolutionTable
	if err := r.db.WithContext(ctx).First(&resolution, "resolution_id = ?", resolutionID).Error; err != nil {
		return nil, err
	}
	return &resolution, nil
}
