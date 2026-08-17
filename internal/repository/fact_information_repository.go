package repository

import (
	"context"

	"github.com/arbazshaikh150/TimeCourt/internal/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type FactInformationRepository interface {
	Get(ctx context.Context, factInformationID uuid.UUID) (*model.FactInformation, error)
}

type pgxFactInformationRepository struct{ db *pgxpool.Pool }

func NewFactInformationRepository(db *pgxpool.Pool) FactInformationRepository {
	return &pgxFactInformationRepository{db: db}
}

func (r *pgxFactInformationRepository) Get(ctx context.Context, factInformationID uuid.UUID) (*model.FactInformation, error) {
	var factInformation model.FactInformation

	err := r.db.QueryRow(
		ctx,
		`
		SELECT
			fact_information_id,
			fact_id,
			tenant_id,
			fact_key,
			fact_version,
			subject_id,
			fact_effective_start_time,
			fact_effective_end_time,
			knowledge_time,
			fact_value,
			authority,
			confidence,
			source,
			effective_period::text
		FROM fact_information
		WHERE fact_information_id = $1
		`,
		factInformationID,
	).Scan(
		&factInformation.FactInformationID,
		&factInformation.FactID,
		&factInformation.TenantID,
		&factInformation.FactKey,
		&factInformation.FactVersion,
		&factInformation.SubjectID,
		&factInformation.FactEffectiveStartTime,
		&factInformation.FactEffectiveEndTime,
		&factInformation.KnowledgeTime,
		&factInformation.FactValue,
		&factInformation.Authority,
		&factInformation.Confidence,
		&factInformation.Source,
		&factInformation.EffectivePeriod,
	)

	if err != nil {
		return nil, err
	}

	return &factInformation, nil
}

// Adding inside the database
// It is an transaction operation
// First i have to lock the factVersion row and increment the version count
// Then add inside the fact Information table in one transaction
// TODO : IDEMPOTENCY SHOULD ALSO BE TAKEN CARE IN SERVICE
func (r *pgxFactInformationRepository) Create(
	ctx context.Context,
	factInformation *model.FactInformation,
) error {
	// Transaction begins
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var latestVersion int64
	// Incrementing the value and acquiring the lock
	err = tx.QueryRow(
		ctx,
		`
	INSERT INTO fact_versions (
		fact_key,
		subject_id,
		latest_version
	)
	VALUES ($1, 1)
	ON CONFLICT (fact_key, subject_id)
	DO UPDATE
	SET latest_version = fact_versions.latest_version + 1
	RETURNING latest_version
	`,
		factInformation.FactKey,
		factInformation.SubjectID,
	).Scan(&latestVersion)

	if err != nil {
		return err
	}

	// Insert fact information.
	_, err = tx.Exec(
		ctx,
		`
		INSERT INTO fact_information (
			fact_information_id,
			fact_id,
			tenant_id,
			fact_key,
			fact_version,
			subject_id,
			fact_effective_start_time,
			fact_effective_end_time,
			knowledge_time,
			fact_value,
			authority,
			confidence,
			source,
			effective_period
		)
		VALUES (
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			$7,
			$8,
			$9,
			$10,
			$11,
			$12,
			$13,
			$14
		)
		`,
		factInformation.FactInformationID,
		factInformation.FactID,
		factInformation.TenantID,
		factInformation.FactKey,
		factInformation.FactVersion,
		factInformation.SubjectID,
		factInformation.FactEffectiveStartTime,
		factInformation.FactEffectiveEndTime,
		factInformation.KnowledgeTime,
		factInformation.FactValue,
		factInformation.Authority,
		factInformation.Confidence,
		factInformation.Source,
		factInformation.EffectivePeriod,
	)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}
