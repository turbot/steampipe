package db_local

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/turbot/steampipe/pkg/constants"
)

func executeSqlAsRoot(ctx context.Context, statements ...string) ([]pgconn.CommandTag, error) {
	rootClient, err := CreateLocalDbConnection(ctx, &CreateDbOptions{Username: constants.DatabaseSuperUser})
	if err != nil {
		return nil, err
	}
	return executeSqlInTransaction(ctx, rootClient, statements...)
}

func executeSqlInTransaction(ctx context.Context, conn *pgx.Conn, statements ...string) (results []pgconn.CommandTag, err error) {

	tx, err := conn.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		}
	}()

	for _, statement := range statements {
		result, err := tx.Exec(ctx, statement)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return results, nil
}
