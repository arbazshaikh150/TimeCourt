package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Create a database will pooling
type Database struct {
	Pool *pgxpool.Pool
}

func NewPool(ctx context.Context, dsn string) (*Database, error) {
	// Connecting to the database
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to create pgx pool: %w", err)
	}

	// Now pool is acquired , i have to check the connection is succesful or not
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Database{
		Pool: pool,
	}, nil
}

func (p *Database) Close() {
	p.Pool.Close()
}
