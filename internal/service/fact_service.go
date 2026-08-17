package service

import (
	"context"
	"fmt"

	"github.com/arbazshaikh150/TimeCourt/internal/model"
	"github.com/arbazshaikh150/TimeCourt/internal/repository"
	"github.com/google/uuid"
)

type FactService struct {
	repository        repository.FactInformationRepository
	idempotentService IdempotentService
}

func NewFactService(
	repository repository.FactInformationRepository,
	idempotentService IdempotentService,
) *FactService {
	return &FactService{
		repository:        repository,
		idempotentService: idempotentService,
	}
}

// Getting the result From the repository
func (s *FactService) Get(
	ctx context.Context,
	factInformationID uuid.UUID,
) (*model.FactInformation, error) {

	return s.repository.Get(
		ctx,
		factInformationID,
	)
}

func (s *FactService) Create(
	ctx context.Context,
	factInformation *model.FactInformation,
	idempotentKey uuid.UUID,
) error {
	id := uuid.New()
	factInformation.FactInformationID = id
	idempotentResult, err := s.idempotentService.Lock(ctx, idempotentKey, factInformation.Source, factInformation.TenantID)

	if err != nil {
		fmt.Printf("failed to acquire idempotency lock for key %s: %v\n", idempotentKey, err)
		return fmt.Errorf("acquire idempotency lock for key %s: %w", idempotentKey, err)
	}

	if !idempotentResult.Acquired {
		err := fmt.Errorf(
			"idempotency lock was not acquired for key %s (source=%s, status=%s)",
			idempotentKey,
			idempotentResult.Source,
			idempotentResult.Status,
		)
		fmt.Println(err)
		return err
	}

	return s.repository.Create(ctx, factInformation, idempotentKey)
}
