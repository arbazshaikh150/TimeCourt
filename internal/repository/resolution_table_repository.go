package repository

import (
	"context"
	"fmt"

	"github.com/arbazshaikh150/TimeCourt/internal/enums"
	"github.com/arbazshaikh150/TimeCourt/internal/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ResolutionTableRepository interface {
	Create(
		ctx context.Context,
		resolution *model.ResolutionTable,
		idempotentKey uuid.UUID,
	) error

	Get(
		ctx context.Context,
		resolutionID uuid.UUID,
	) (*model.ResolutionTable, error)
}

type PgxResolutionTableRepository struct {
	db *pgxpool.Pool
}

func NewResolutionTableRepository(db *pgxpool.Pool) ResolutionTableRepository {
	return &PgxResolutionTableRepository{
		db: db,
	}
}

func (r *PgxResolutionTableRepository) Create(
	ctx context.Context,
	resolution *model.ResolutionTable,
	idempotentKey uuid.UUID,
) error {

	// 1. Begin transaction

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin resolution transaction: %w", err)
	}

	defer tx.Rollback(ctx)

	// 2. Generate the next resolution version
	//
	// The INSERT ... ON CONFLICT performs the increment
	// atomically.

	var latestVersion int64

	err = tx.QueryRow(
		ctx,
		`
		INSERT INTO resolution_versions (
			rule_key,
			resolution_latest_version
		)
		VALUES ($1, 1)
		ON CONFLICT (rule_key)
		DO UPDATE
		SET resolution_latest_version =
			resolution_versions.resolution_latest_version + 1
		RETURNING resolution_latest_version
		`,
		resolution.RuleKey,
	).Scan(&latestVersion)

	if err != nil {
		return fmt.Errorf(
			"failed to generate resolution version for rule_key %q: %w",
			resolution.RuleKey,
			err,
		)
	}

	// 3. Set generated version on resolution

	resolution.Version = latestVersion

	// 4. Insert resolution table row

	_, err = tx.Exec(
		ctx,
		`
		INSERT INTO resolution_tables (
			resolution_id,
			rule_key,
			version,
			tenant_id,
			current_time,
			resolution_policy
		)
		VALUES (
			$1,
			$2,
			$3,
			$4,
			COALESCE($5, CURRENT_TIMESTAMP),
			$6
		)
		`,
		resolution.ResolutionID,
		resolution.RuleKey,
		resolution.Version,
		resolution.TenantID,
		resolution.CurrentTime,
		resolution.ResolutionPolicy,
	)

	if err != nil {
		return fmt.Errorf(
			"failed to insert resolution for rule_key %q version %d: %w",
			resolution.RuleKey,
			resolution.Version,
			err,
		)
	}

	// 5. Mark idempotency record as successful

	_, err = tx.Exec(
		ctx,
		`
		UPDATE idempotency_records
		SET
			status = $1,
			resource_value_ref = $2,
			committed_time = CURRENT_TIMESTAMP
		WHERE idempotent_key = $3
		`,
		enums.IdempotentSuccess,
		resolution.ResolutionID,
		idempotentKey,
	)

	if err != nil {
		return fmt.Errorf(
			"failed to update idempotency record %q: %w",
			idempotentKey,
			err,
		)
	}

	// 6. Commit

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf(
			"failed to commit resolution transaction: %w",
			err,
		)
	}

	return nil
}

func (r *PgxResolutionTableRepository) Get(
	ctx context.Context,
	resolutionID uuid.UUID,
) (*model.ResolutionTable, error) {

	var resolution model.ResolutionTable

	err := r.db.QueryRow(ctx, `
		SELECT
			resolution_id,
			rule_key,
			version,
			tenant_id,
			current_time,
			resolution_policy
		FROM resolution_tables
		WHERE resolution_id = $1
	`, resolutionID).Scan(
		&resolution.ResolutionID,
		&resolution.RuleKey,
		&resolution.Version,
		&resolution.TenantID,
		&resolution.CurrentTime,
		&resolution.ResolutionPolicy,
	)

	if err != nil {
		return nil, fmt.Errorf(
			"failed to get resolution table %q: %w",
			resolutionID,
			err,
		)
	}

	return &resolution, nil
}
