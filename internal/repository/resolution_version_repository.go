package repository

import (
	"context"
	"fmt"

	"github.com/arbazshaikh150/TimeCourt/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TODO : Complete the implementation of the resolution and interpretor
// Deliver a complete trace in the response and store inside the decision table with proper versioning
// TODO : Idempotent based fetching is pending.
type ResolutionVersionRepository interface {
	Get(ctx context.Context, ruleKey string) (*model.ResolutionVersion, error)
}

type PgxResolutionVersionRepository struct{ db *pgxpool.Pool }

func NewResolutionVersionRepository(db *pgxpool.Pool) ResolutionVersionRepository {
	return &PgxResolutionVersionRepository{db: db}
}

func (r *PgxResolutionVersionRepository) Get(ctx context.Context, ruleKey string) (*model.ResolutionVersion, error) {
	// Based on the rulekey and the version that is needed --> i can fetch the
	result := &model.ResolutionVersion{}
	err := r.db.QueryRow(ctx, `
		SELECT rule_key , resolution_latest_version
		from resolution_versions
		where rule_key= $1
	`, ruleKey).Scan(&result.RuleKey, &result.ResolutionLatestVersion)

	if err != nil {
		fmt.Println("Error in fetching the latest version")
		return nil, err
	}

	return result, nil

}
