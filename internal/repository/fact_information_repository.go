package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/arbazshaikh150/TimeCourt/internal/enums"
	"github.com/arbazshaikh150/TimeCourt/internal/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TODO : Time unit must be same for everyone
type FactInformationRepository interface {
	Get(ctx context.Context, factInformationID uuid.UUID) (*model.FactInformation, error)
	Create(
		ctx context.Context,
		factInformation *model.FactInformation,
		idempotentKey uuid.UUID,
	) error
	FindByEffectiveTime(
		ctx context.Context,
		factKey string,
		subjectID string,
		tenantId string,
		timeWhereToCheck time.Time,
		timeWhenToCheck time.Time,
	) ([]*model.FactInformation, error)
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
	idempotentKey uuid.UUID,
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
		tenant_id,
		latest_version
	)
	VALUES ($1, $2, $3, 1)
	ON CONFLICT (fact_key, subject_id , tenant_id)
	DO UPDATE
	SET latest_version = fact_versions.latest_version + 1
	RETURNING latest_version
	`,
		factInformation.FactKey,
		factInformation.SubjectID,
		factInformation.TenantID,
	).Scan(&latestVersion)

	if err != nil {
		fmt.Println("Insert error", err)
		return err
	}

	// Insert fact information.
	// Set the generated version on the fact.
	factInformation.FactVersion = latestVersion

	// Insert fact information.
	_, err = tx.Exec(
		ctx,
		`
	INSERT INTO fact_information (
		fact_information_id,
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
		source
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
		$12
	)
	`,
		factInformation.FactInformationID,
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
	)

	if err != nil {
		fmt.Println("Error in creating the fact : ", err)
		return err
	}

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
		factInformation.FactInformationID,
		idempotentKey,
	)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// Method for finding the overlapping fact for the given factkey + subjectid
// Given a date and then finding the fact that overlaps with it
// Require two parameter , What was my timewheretocheck and which time i have to check timewhentocheck
// Also (timewhentocheck <= knowledge time) and timewheretocheck must overlap with the effectivetimestamp
// It must use the proper indexing
func (r *pgxFactInformationRepository) FindByEffectiveTime(
	ctx context.Context,
	factKey string,
	subjectID string,
	tenantID string,
	timeWhereToCheck time.Time,
	timeWhenToCheck time.Time,
) ([]*model.FactInformation, error) {

	rows, err := r.db.Query(
		ctx,
		`
		SELECT
			fact_information_id,
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
		WHERE fact_key = $1
			AND subject_id = $2
			AND tenant_id = $5
			AND effective_period @> $3::timestamptz
			AND knowledge_time <= $4
		`,
		factKey,
		subjectID,
		timeWhereToCheck,
		timeWhenToCheck,
		tenantID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	facts := make([]*model.FactInformation, 0)

	for rows.Next() {
		factInformation := &model.FactInformation{}

		err := rows.Scan(
			&factInformation.FactInformationID,
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

		facts = append(facts, factInformation)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return facts, nil
}
