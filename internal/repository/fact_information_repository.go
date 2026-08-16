package repository

import (
	"context"

	"github.com/arbazshaikh150/TimeCourt/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FactInformationRepository interface {
	Get(ctx context.Context, factInformationID uuid.UUID) (*model.FactInformation, error)
}

type GormFactInformationRepository struct{ db *gorm.DB }

func NewFactInformationRepository(db *gorm.DB) FactInformationRepository {
	return &GormFactInformationRepository{db: db}
}

func (r *GormFactInformationRepository) Get(ctx context.Context, factInformationID uuid.UUID) (*model.FactInformation, error) {
	var factInformation model.FactInformation
	if err := r.db.WithContext(ctx).First(&factInformation, "fact_information_id = ?", factInformationID).Error; err != nil {
		return nil, err
	}
	return &factInformation, nil
}
