package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/arbazshaikh150/TimeCourt/internal/dto"
	"github.com/arbazshaikh150/TimeCourt/internal/enums"
	"github.com/arbazshaikh150/TimeCourt/internal/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RuleTableRepository interface {
	Get(ctx context.Context, ruleID uuid.UUID) (*model.RuleTable, error)
	Create(ctx context.Context, rule *model.RuleTable, ruleDetails []model.RuleDetail, idempotentKey uuid.UUID) error
	Find(ctx context.Context, ruleKey, subjectID string, timeWhenToCheck, timeWhereToCheck time.Time) (*dto.RuleFetchDetails, error)
}

type PgxRuleTableRepository struct{ db *pgxpool.Pool }

func NewRuleTableRepository(db *pgxpool.Pool) RuleTableRepository {
	return &PgxRuleTableRepository{db: db}
}

const getRuleQuery = `
	SELECT rule_id, tenant_id, rule_key, rule_version, knowledge_time,
		effective_start_time, effective_end_time, resolution_version,
		interpretor_version, source
	FROM rules
	WHERE rule_id = $1
`

func (r *PgxRuleTableRepository) Get(ctx context.Context, ruleID uuid.UUID) (*model.RuleTable, error) {
	var rule model.RuleTable
	if err := scanRule(r.db.QueryRow(ctx, getRuleQuery, ruleID), &rule); err != nil {
		return nil, fmt.Errorf("get rule %s: %w", ruleID, err)
	}
	return &rule, nil
}

func (r *PgxRuleTableRepository) Create(
	ctx context.Context,
	rule *model.RuleTable,
	ruleDetails []model.RuleDetail,
	idempotentKey uuid.UUID,
) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin rule creation transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := assignNextRuleVersion(ctx, tx, rule); err != nil {
		return fmt.Errorf("assign rule version: %w", err)
	}
	if err := insertRule(ctx, tx, rule); err != nil {
		return fmt.Errorf("insert rule %s: %w", rule.RuleID, err)
	}
	if err := insertRuleDetails(ctx, tx, rule.RuleID, ruleDetails); err != nil {
		return fmt.Errorf("insert rule details for %s: %w", rule.RuleID, err)
	}
	if err := markIdempotencyRecordComplete(ctx, tx, idempotentKey, rule.RuleID); err != nil {
		return fmt.Errorf("complete idempotency record %s: %w", idempotentKey, err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit rule creation transaction: %w", err)
	}
	return nil
}

type rowScanner interface{ Scan(dest ...any) error }

func scanRule(row rowScanner, rule *model.RuleTable) error {
	return row.Scan(
		&rule.RuleID, &rule.TenantID, &rule.RuleKey, &rule.RuleVersion,
		&rule.KnowledgeTime, &rule.EffectiveStartTime, &rule.EffectiveEndTime,
		&rule.ResolutionVersion, &rule.InterpretorVersion, &rule.Source,
	)
}

const nextRuleVersionQuery = `
	INSERT INTO rule_versions (rule_key, rule_version_id)
	VALUES ($1, 1)
	ON CONFLICT (rule_key)
	DO UPDATE SET rule_version_id = rule_versions.rule_version_id + 1
	RETURNING rule_version_id
`

func assignNextRuleVersion(ctx context.Context, tx pgx.Tx, rule *model.RuleTable) error {
	return tx.QueryRow(ctx, nextRuleVersionQuery, rule.RuleKey).Scan(&rule.RuleVersion)
}

const insertRuleQuery = `
	INSERT INTO rules (
		rule_id, tenant_id, rule_key, rule_version, knowledge_time,
		effective_start_time, effective_end_time, resolution_version,
		interpretor_version, source
	)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
`

func insertRule(ctx context.Context, tx pgx.Tx, rule *model.RuleTable) error {
	_, err := tx.Exec(ctx, insertRuleQuery,
		rule.RuleID, rule.TenantID, rule.RuleKey, rule.RuleVersion,
		rule.KnowledgeTime, rule.EffectiveStartTime, rule.EffectiveEndTime,
		rule.ResolutionVersion, rule.InterpretorVersion, rule.Source,
	)
	return err
}

const insertRuleDetailQuery = `
	INSERT INTO rule_details (rule_id, fact_key, fact_required_value)
	VALUES ($1, $2, $3)
`

func insertRuleDetails(ctx context.Context, tx pgx.Tx, ruleID uuid.UUID, details []model.RuleDetail) error {
	for _, detail := range details {
		if _, err := tx.Exec(ctx, insertRuleDetailQuery, ruleID, detail.FactKey, detail.FactRequiredValue); err != nil {
			return err
		}
	}
	return nil
}

const completeIdempotencyRecordQuery = `
	UPDATE idempotency_records
	SET status = $1, resource_value_ref = $2, committed_time = CURRENT_TIMESTAMP
	WHERE idempotent_key = $3
`

func markIdempotencyRecordComplete(ctx context.Context, tx pgx.Tx, idempotentKey, ruleID uuid.UUID) error {
	_, err := tx.Exec(ctx, completeIdempotencyRecordQuery, enums.IdempotentSuccess, ruleID, idempotentKey)
	return err
}
