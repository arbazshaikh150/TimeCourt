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
	Lock(ctx context.Context, idempotentKey uuid.UUID, source string, tenant_id string) (*dto.IdempotentResult, error)
	Commit(ctx context.Context, idempotentKey uuid.UUID, source string) error
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
			resource_value_ref,
			locked_time,
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
		&idempotent.LockedTime,
		&idempotent.Status,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			fmt.Printf("No rows found for the given idempotent key : %s\n", idempotentKey)
			return nil, err
		}

		return nil, err
	}

	return &idempotent, nil
}

func (r *PgIdempotentRepository) Lock(ctx context.Context, idempotentKey uuid.UUID, source string, tenant_id string) (*dto.IdempotentResult, error) {
	var result dto.IdempotentResult
	// Try to create the idempotency record.
	// Successfully inserting it means this request acquired the lock.
	err := r.db.QueryRow(
		ctx,
		`
		INSERT INTO idempotency_records (
			idempotent_key,
			source,
			status,
			tenant_id
		)
		VALUES ($1, $2, $3 , $4)
		ON CONFLICT (idempotent_key) DO NOTHING
		RETURNING source, status
		`,
		idempotentKey,
		source,
		enums.IdempotentProcessing,
		tenant_id,
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

// Adding the commit function
func (r *PgIdempotentRepository) Commit(
	ctx context.Context,
	idempotentKey uuid.UUID,
	source string,
) error {

	tag, err := r.db.Exec(
		ctx,
		`
		UPDATE idempotency_records
		SET
			status = $1,
			committed_time = CURRENT_TIMESTAMP
		WHERE idempotent_key = $2
			AND source = $3
			AND status = $4
		`,
		enums.IdempotentSuccess,
		idempotentKey,
		source,
		enums.IdempotentProcessing,
	)

	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf(
			"idempotency record cannot be committed: key=%s source=%s",
			idempotentKey,
			source,
		)
	}
	fmt.Println("Successfully Committed Into Idempotent Key")
	return nil
}
