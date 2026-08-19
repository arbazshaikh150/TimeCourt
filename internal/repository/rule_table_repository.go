package repository

import (
	"context"

	"github.com/arbazshaikh150/TimeCourt/internal/enums"
	"github.com/arbazshaikh150/TimeCourt/internal/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RuleTableRepository interface {
	Get(ctx context.Context, ruleID uuid.UUID) (*model.RuleTable, error)
	Create(
		ctx context.Context,
		rule *model.RuleTable,
		ruleDetails []model.RuleDetail,
		idempotentKey uuid.UUID,
	) error
}

type PgxRuleTableRepository struct {
	db *pgxpool.Pool
}

func NewRuleTableRepository(db *pgxpool.Pool) RuleTableRepository {
	return &PgxRuleTableRepository{
		db: db,
	}
}

func (r *PgxRuleTableRepository) Get(
	ctx context.Context,
	ruleID uuid.UUID,
) (*model.RuleTable, error) {

	var rule model.RuleTable

	err := r.db.QueryRow(
		ctx,
		`
		SELECT
			rule_id,
			tenant_id,
			rule_key,
			rule_version,
			knowledge_time,
			effective_start_time,
			effective_end_time,
			resolution_version,
			interpretor_version,
			source
		FROM rules
		WHERE rule_id = $1
		`,
		ruleID,
	).Scan(
		&rule.RuleID,
		&rule.TenantID,
		&rule.RuleKey,
		&rule.RuleVersion,
		&rule.KnowledgeTime,
		&rule.EffectiveStartTime,
		&rule.EffectiveEndTime,
		&rule.ResolutionVersion,
		&rule.InterpretorVersion,
		&rule.Source,
	)

	if err != nil {
		return nil, err
	}

	return &rule, nil
}

// Creating a rule based on the facts
// Getting the list of the facts and there desired values for the given rule
// Adding them into the rule_details table and storing the metadata here
// First fetching the rule version and updating it and then storing the metadata inside the rules table
// after that for the array of the rules and required values storing it into the rule_details table
// lastly updating the idempotency record as completed ( have idempotency key and the source)

func (r *PgxRuleTableRepository) Create(
	ctx context.Context,
	rule *model.RuleTable,
	ruleDetails []model.RuleDetail,
	idempotentKey uuid.UUID,
) error {
	// Begin transaction.
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	//
	// 1. Increment the latest rule version.
	//
	// This also locks the corresponding row when it already
	// exists, making concurrent rule creation safe.
	//

	var latestVersion int64

	err = tx.QueryRow(
		ctx,
		`
		INSERT INTO rule_versions (
			rule_key,
			rule_version_id
		)
		VALUES ($1, 1)
		ON CONFLICT (rule_key)
		DO UPDATE
		SET rule_version_id = rule_versions.rule_version_id + 1
		RETURNING rule_version_id
		`,
		rule.RuleKey,
	).Scan(&latestVersion)

	if err != nil {
		return err
	}

	// Set the newly generated version on the rule.
	rule.RuleVersion = latestVersion

	//
	// 2. Insert rule metadata into rules table.
	//
	// I have a ruleid as the server generate uuid and then i can use that as a foreign key
	_, err = tx.Exec(
		ctx,
		`
		INSERT INTO rules (
			rule_id,
			tenant_id,
			rule_key,
			rule_version,
			knowledge_time,
			effective_start_time,
			effective_end_time,
			resolution_version,
			interpretor_version,
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
			$10
		)
		`,
		rule.RuleID,
		rule.TenantID,
		rule.RuleKey,
		rule.RuleVersion,
		rule.KnowledgeTime,
		rule.EffectiveStartTime,
		rule.EffectiveEndTime,
		rule.ResolutionVersion,
		rule.InterpretorVersion,
		rule.Source,
	)

	if err != nil {
		return err
	}

	//
	// 3. Insert all required facts into rule_details.
	//

	for _, detail := range ruleDetails {

		// Make sure every detail belongs to this rule.
		detail.RuleID = rule.RuleID

		_, err = tx.Exec(
			ctx,
			`
			INSERT INTO rule_details (
				rule_id,
				fact_key,
				fact_required_value
			)
			VALUES (
				$1,
				$2,
				$3
			)
			`,
			detail.RuleID,
			detail.FactKey,
			detail.FactRequiredValue,
		)

		if err != nil {
			return err
		}
	}

	//
	// 4. Mark idempotency record as successfully completed.
	//

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
		rule.RuleID,
		idempotentKey,
	)

	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}
