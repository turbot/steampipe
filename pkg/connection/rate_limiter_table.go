package connection

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/turbot/steampipe/pkg/constants"
)

func (s *refreshConnectionState) ensureRateLimiterTable(ctx context.Context) error {
	// if the rate limiter table exists, nothing to do
	exists, err := rateLimiterTableExists(ctx, s.pool)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	// create rate limiter table

	// so we must load all plugins and retrieve their rate limiter defs
	// TODO KAI
	return nil
}

func rateLimiterTableExists(ctx context.Context, pool *pgxpool.Pool) (bool, error) {
	query := fmt.Sprintf(`SELECT EXISTS (
    SELECT FROM 
        pg_tables
    WHERE 
        schemaname = '%s' AND 
        tablename  = '%s'
    );`, constants.InternalSchema, constants.RateLimiterDefinitionTable)

	row := pool.QueryRow(ctx, query)
	var exists bool
	err := row.Scan(&exists)

	if err != nil {
		return false, err
	}
	return exists, nil
}

