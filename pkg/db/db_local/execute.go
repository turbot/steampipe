package db_local

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/turbot/steampipe/v2/pkg/constants"
	"github.com/turbot/steampipe/v2/pkg/db/db_common"
)

func executeSqlAsRoot(ctx context.Context, statements ...string) ([]pgconn.CommandTag, error) {
	log.Println("[DEBUG] executeSqlAsRoot start")
	defer log.Println("[DEBUG] executeSqlAsRoot end")

	rootClient, err := CreateLocalDbConnection(ctx, &CreateDbOptions{Username: constants.DatabaseSuperUser})
	if err != nil {
		return nil, err
	}
	return ExecuteSqlInTransaction(ctx, rootClient, statements...)
}

func ExecuteSqlInTransaction(ctx context.Context, conn *pgx.Conn, statements ...string) (results []pgconn.CommandTag, err error) {
	log.Println("[DEBUG] ExecuteSqlInTransaction start")
	defer log.Println("[DEBUG] ExecuteSqlInTransaction end")

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
	log.Println("[DEBUG] ExecuteSqlWithArgsInTransaction start")
	defer log.Println("[DEBUG] ExecuteSqlWithArgsInTransaction end")

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
