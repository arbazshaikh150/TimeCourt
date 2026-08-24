package service

import (
	"context"
	"fmt"
	"time"

	"github.com/arbazshaikh150/TimeCourt/internal/enums"
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
) (*model.FactInformation, error) {
	id := uuid.New()
	factInformation.FactInformationID = id
	idempotentResult, err := s.idempotentService.Lock(ctx, idempotentKey, factInformation.Source, factInformation.TenantID)

	if err != nil {
		fmt.Printf("failed to acquire idempotency lock for key %s: %v\n", idempotentKey, err)
		return nil, fmt.Errorf("acquire idempotency lock for key %s: %w", idempotentKey, err)
	}

	if !idempotentResult.Acquired {

		// 2 DATABASE CALL --> Will do optimize with joins if latency is high.
		if idempotentResult.Status == enums.IdempotentSuccess {
			idempotent, err := s.idempotentService.Get(ctx, idempotentKey)
			if err != nil {
				return nil, fmt.Errorf("get idempotency record for key %s: %w", idempotentKey, err)
			}

			fact, err := s.repository.Get(ctx, idempotent.ResourceValueRef)
			if err != nil {
				return nil, fmt.Errorf("get idempotent fact %s: %w", idempotent.ResourceValueRef, err)
			}

			return fact, nil
		}

		err := fmt.Errorf(
			"idempotency lock was not acquired for key %s (source=%s, status=%s)",
			idempotentKey,
			idempotentResult.Source,
			idempotentResult.Status,
		)
		fmt.Println(err)
		return nil, err
	}

	if err := s.repository.Create(ctx, factInformation, idempotentKey); err != nil {
		return nil, err
	}

	return factInformation, nil
}

// Fetching from the repostory and then sending the response
func (s *FactService) FindByEffectiveTime(
	ctx context.Context,
	factKey string,
	subjectID string,
	tenantID string,
	timeWhereToCheck time.Time,
	timeWhenToCheck time.Time,
) ([]*model.FactInformation, error) {
	fmt.Println("Fetching from the repository")
	return s.repository.FindByEffectiveTime(
		ctx,
		factKey,
		subjectID,
		tenantID,
		timeWhereToCheck,
		timeWhenToCheck,
	)
}
