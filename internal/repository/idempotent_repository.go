package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/arbazshaikh150/TimeCourt/internal/dto"
	"github.com/arbazshaikh150/TimeCourt/internal/enums"
	"github.com/arbazshaikh150/TimeCourt/internal/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Implementing the idempotent repository Layer
type IdempotentRepository interface {
	Get(ctx context.Context, idempotentKey uuid.UUID) (*model.Idempotent, error)
	// Another function for acquiring lock
	// Return the source which acquires the lock
	Lock(ctx context.Context, idempotentKey uuid.UUID, source string) (*dto.IdempotentResult, error)
}

type PgIdempotentRepository struct {
	db *pgxpool.Pool
}

// using simple pgx layer for runtime database operation
func NewIdempotentRepository(db *pgxpool.Pool) IdempotentRepository {
	return &PgIdempotentRepository{
		db: db,
	}
}

// Implementing the necessary interface function
func (r *PgIdempotentRepository) Get(ctx context.Context, idempotentKey uuid.UUID) (*model.Idempotent, error) {
	var idempotent model.Idempotent

	// Using basic sql query
	err := r.db.QueryRow(
		ctx,
		`
		SELECT
			idempotent_key,
			tenant_id,
			source,
			request_digest,
			resource_id,
			resource_value,
			time,
			status
		FROM idempotency_records
		WHERE idempotent_key = $1
		`,
		idempotentKey,
	).Scan(
		&idempotent.IdempotentKey,
		&idempotent.TenantID,
		&idempotent.Source,
		&idempotent.RequestDigest,
		&idempotent.ResourceID,
		&idempotent.ResourceValueRef,
		&idempotent.CurrentTime,
		&idempotent.Status,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			fmt.Println("No rows found for the given idempotent key : %s", idempotentKey)
			return nil, err
		}

		return nil, err
	}

	return &idempotent, nil
}

func (r *PgIdempotentRepository) Lock(ctx context.Context, idempotentKey uuid.UUID, source string) (*dto.IdempotentResult, error) {
	var result dto.IdempotentResult

	// Try to create the idempotency record.
	// Successfully inserting it means this request acquired the lock.
	err := r.db.QueryRow(
		ctx,
		`
		INSERT INTO idempotency_records (
			idempotent_key,
			source,
			status
		)
		VALUES ($1, $2, $3)
		ON CONFLICT (idempotent_key) DO NOTHING
		RETURNING source, status
		`,
		idempotentKey,
		source,
		enums.IdempotentProcessing,
	).Scan(
		&result.Source,
		&result.Status,
	)

	if err == nil {
		result.Acquired = true
		return &result, nil
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	// The INSERT didn't happen because the row already exists.
	err = r.db.QueryRow(
		ctx,
		`
		SELECT source, status
		FROM idempotency_records
		WHERE idempotent_key = $1
		`,
		idempotentKey,
	).Scan(
		&result.Source,
		&result.Status,
	)

	if err != nil {
		return nil, err
	}

	result.Acquired = false

	return &result, nil
}
