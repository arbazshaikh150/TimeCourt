package repository

import (
	"context"

	"github.com/arbazshaikh150/TimeCourt/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RuleVersionRepository interface {
	Get(ctx context.Context, ruleKey string) (*model.RuleVersion, error)
}

type PgxRuleVersionRepository struct{ db *pgxpool.Pool }

func NewRuleVersionRepository(db *pgxpool.Pool) RuleVersionRepository {
	return &PgxRuleVersionRepository{db: db}
}

// TODO : IDEMPOTENT BASED FETCHING IS PENDING ( for successful queries retry )
func (r *PgxRuleVersionRepository) Get(
	ctx context.Context,
	ruleKey string,
) (*model.RuleVersion, error) {

	var ruleVersion model.RuleVersion

	err := r.db.QueryRow(
		ctx,
		`
		SELECT
			rule_key,
			rule_version_id
		FROM rule_versions
		WHERE rule_key = $1
		`,
		ruleKey,
	).Scan(
		&ruleVersion.RuleKey,
		&ruleVersion.RuleVersionID,
	)

	if err != nil {
		return nil, err
	}

	return &ruleVersion, nil
}

