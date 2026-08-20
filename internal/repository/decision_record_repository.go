package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/arbazshaikh150/TimeCourt/internal/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DecisionRecordRepository interface {
	Create(ctx context.Context, record *model.DecisionRecord, details *model.DecisionDetails) error
	Get(ctx context.Context, decisionID uuid.UUID) (*model.DecisionRecord, error)
}

type PgxDecisionRecordRepository struct{ db *pgxpool.Pool }

func NewDecisionRecordRepository(db *pgxpool.Pool) DecisionRecordRepository {
	return &PgxDecisionRecordRepository{db: db}
}

// Create persists a decision and its explanation atomically. Both models use
// the same DecisionID, preserving the one-to-one decision record/detail link.
func (r *PgxDecisionRecordRepository) Create(
	ctx context.Context,
	record *model.DecisionRecord,
	details *model.DecisionDetails,
) error {
	if record == nil || details == nil {
		return fmt.Errorf("decision record and decision details are required")
	}
	if record.DecisionID != details.DecisionID {
		return fmt.Errorf("decision record and details must use the same decision ID")
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin decision transaction: %w", err)
	}
	defer tx.Rollback(ctx)
	now := time.Now().UTC()

	_, err = tx.Exec(ctx, `
		INSERT INTO decision_records (
			decision_id, tenant_id, time_of_completion, time_when_to_check, version,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $6)
	`, record.DecisionID, record.TenantID, record.TimeOfCompletion, record.TimeWhenToCheck, record.Version, now)
	if err != nil {
		return fmt.Errorf("create decision record: %w", err)
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO decision_details (
			decision_id, subject_id, rule_key, rule_used_version,
			resolution_policy_version, fact_required_data, fact_present_data,
			failed_facts, resolution_policy_ref, status, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $11)
	`,
		details.DecisionID,
		details.SubjectID,
		details.RuleKey,
		details.RuleUsedVersion,
		details.ResolutionPolicyVersion,
		[]byte(details.FactRequiredData),
		[]byte(details.FactPresentData),
		[]byte(details.FailedFacts),
		details.ResolutionPolicyRef,
		details.Status,
		now,
	)
	if err != nil {
		return fmt.Errorf("create decision details: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit decision transaction: %w", err)
	}
	return nil
}

func (r *PgxDecisionRecordRepository) Get(ctx context.Context, decisionID uuid.UUID) (*model.DecisionRecord, error) {
	var record model.DecisionRecord
	err := r.db.QueryRow(ctx, `
		SELECT decision_id, tenant_id, time_of_completion, time_when_to_check,
			version, created_at, updated_at
		FROM decision_records
		WHERE decision_id = $1
	`, decisionID).Scan(
		&record.DecisionID,
		&record.TenantID,
		&record.TimeOfCompletion,
		&record.TimeWhenToCheck,
		&record.Version,
		&record.CreatedAt,
		&record.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, err
		}
		return nil, fmt.Errorf("get decision record %s: %w", decisionID, err)
	}
	return &record, nil
}
