package service

import (
	"context"
	"fmt"

	"github.com/arbazshaikh150/TimeCourt/internal/dto"
	"github.com/arbazshaikh150/TimeCourt/internal/model"
	"github.com/arbazshaikh150/TimeCourt/internal/repository"
	"github.com/google/uuid"
)

// TODO : Idempotent Goroutine call is pending
type ResolutionService struct {
	repository        repository.ResolutionTableRepository
	idempotentService IdempotentService
}

func NewResolutionService(
	repository repository.ResolutionTableRepository,
	idempotentService IdempotentService,
) *ResolutionService {
	return &ResolutionService{
		repository:        repository,
		idempotentService: idempotentService,
	}
}

func (r *ResolutionService) Get(
	ctx context.Context,
	resolutionID uuid.UUID,
) (*model.ResolutionTable, error) {

	resolution, err := r.repository.Get(ctx, resolutionID)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to get resolution %s: %w",
			resolutionID,
			err,
		)
	}

	return resolution, nil
}

func (r *ResolutionService) Create(
	ctx context.Context,
	req *dto.CreateResolutionRequest,
	idempotentKey uuid.UUID,
) error {

	// 1. Acquire idempotency lock

	idempotentResult, err := r.idempotentService.Lock(
		ctx,
		idempotentKey,
		req.Source,
		req.TenantID,
	)

	if err != nil {
		return fmt.Errorf(
			"failed to acquire idempotency lock for key %s: %w",
			idempotentKey,
			err,
		)
	}

	if !idempotentResult.Acquired {
		return fmt.Errorf(
			"idempotency lock was not acquired for key %s (source=%s, status=%s)",
			idempotentKey,
			idempotentResult.Source,
			idempotentResult.Status,
		)
	}

	// 2. Generate Resolution ID

	resolutionID := uuid.New()

	// 3. Convert DTO -> ResolutionTable
	// Version is NOT set here.
	// Repository generates the version atomically using:
	// resolution_versions

	resolution := &model.ResolutionTable{
		ResolutionID:     resolutionID,
		RuleKey:          req.RuleKey,
		TenantID:         req.TenantID,
		ResolutionPolicy: req.ResolutionPolicy,
	}

	// 4. Repository handles the complete transaction
	//
	// Repository will:
	//
	//   - increment resolution version
	//   - insert resolution table
	//   - update idempotency record
	//   - commit

	err = r.repository.Create(
		ctx,
		resolution,
		idempotentKey,
	)

	if err != nil {
		return fmt.Errorf(
			"failed to create resolution %s: %w",
			resolutionID,
			err,
		)
	}

	return nil
}


// Resolution Service will be called only when there is an ambiguity 
// Else directly pass to interpretor service