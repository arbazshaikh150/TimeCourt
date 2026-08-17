package service

import (
	"context"

	"github.com/arbazshaikh150/TimeCourt/internal/dto"
	"github.com/arbazshaikh150/TimeCourt/internal/model"
	"github.com/arbazshaikh150/TimeCourt/internal/repository"
	"github.com/google/uuid"
)

type IdempotentService struct {
	repository repository.IdempotentRepository
}

func NewIdempotentService(
	repository repository.IdempotentRepository,
) *IdempotentService {
	return &IdempotentService{
		repository: repository,
	}
}

func (s *IdempotentService) Get(
	ctx context.Context,
	idempotentKey uuid.UUID,
) (*model.Idempotent, error) {

	return s.repository.Get(
		ctx,
		idempotentKey,
	)
}

func (s *IdempotentService) Lock(
	ctx context.Context,
	idempotentKey uuid.UUID,
	source string,
	tenant_id string,
) (*dto.IdempotentResult, error) {

	return s.repository.Lock(
		ctx,
		idempotentKey,
		source,
		tenant_id,
	)
}

func (s *IdempotentService) Commit(
	ctx context.Context,
	idempotentKey uuid.UUID,
	source string,
) error {

	return s.repository.Commit(
		ctx,
		idempotentKey,
		source,
	)
}
