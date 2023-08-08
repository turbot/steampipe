package db_local

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/constants/runtime"
	"github.com/turbot/steampipe/pkg/db/db_common"
)

func executeSqlAsRoot(ctx context.Context, statements ...string) ([]pgconn.CommandTag, error) {
	rootClient, err := CreateLocalDbConnection(ctx, &CreateDbOptions{Username: constants.DatabaseSuperUser, AppName: runtime.ServiceConnectionAppName})
	if err != nil {
		return nil, err
	}
	return ExecuteSqlInTransaction(ctx, rootClient, statements...)
}

func ExecuteSqlInTransaction(ctx context.Context, conn *pgx.Conn, statements ...string) (results []pgconn.CommandTag, err error) {
	err = pgx.BeginFunc(ctx, conn, func(tx pgx.Tx) error {
		for _, statement := range statements {
			result, err := tx.Exec(ctx, statement)
			if err != nil {
				return err
			}
			results = append(results, result)
		}
		return nil
	})
	return results, err
}

func ExecuteSqlWithArgsInTransaction(ctx context.Context, conn *pgx.Conn, queries ...db_common.QueryWithArgs) (results []pgconn.CommandTag, err error) {
	err = pgx.BeginFunc(ctx, conn, func(tx pgx.Tx) error {
		for _, q := range queries {
			result, err := tx.Exec(ctx, q.Query, q.Args...)
			if err != nil {
				// set the results to nil - so that we don't return stuff in an error return
				results = nil
				return err
			}
			results = append(results, result)
		}
		return nil
	})
	return results, err
}
