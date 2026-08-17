package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type FactVersionRepository interface {
	Get(ctx context.Context, factKey, subjectID string) (int64, error)
	Update(ctx context.Context, factKey, subjectID string) error
}

type PgxFactVersionRepository struct {
	db *pgxpool.Pool
}

func NewFactVersionRepository(db *pgxpool.Pool) FactVersionRepository {
	return &PgxFactVersionRepository{
		db: db,
	}
}

func (r *PgxFactVersionRepository) Get(ctx context.Context, factKey, subjectID string) (int64, error) {
	// Based on the subject id and the factkey it will return the factVersion
	var factVersion int64
	err := r.db.QueryRow(ctx,
		`
	SELECT latest_version
	FROM fact_versions
	WHERE fact_key = $1
		AND subject_id = $2
	`, factKey, subjectID).Scan(&factVersion)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, err // 0 --> No Version found
		}

		return 0, err
	}

	return factVersion, nil
}


// Updating the latest version 
// Issue : Race Condition --> have to fix this issue 
func (r *PgxFactVersionRepository) Update(
	ctx context.Context,
	factKey, subjectID string,
) error {
	_, err := r.db.Exec(
		ctx,
		`
		INSERT INTO fact_versions (
			fact_key,
			subject_id,
			latest_version
		)
		VALUES ($1, $2, 1)
		ON CONFLICT (fact_key, subject_id)
		DO UPDATE
		SET latest_version = fact_versions.latest_version + 1
		`,
		factKey,
		subjectID,
	)

	return err
}
