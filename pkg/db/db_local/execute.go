package db_local

import (
	"context"
	"database/sql"
	"log"

	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/db_common"
)

func executeSqlAsRoot(ctx context.Context, statements ...string) ([]sql.Result, error) {
	log.Println("[DEBUG] executeSqlAsRoot start")
	defer log.Println("[DEBUG] executeSqlAsRoot end")

	rootClient, err := CreateLocalDbConnectionPool(ctx, &CreateDbOptions{Username: constants.DatabaseSuperUser})
	if err != nil {
		return nil, err
	}
	conn, err := rootClient.Conn(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	return ExecuteSqlInTransaction(ctx, conn, statements...)
}

func ExecuteSqlInTransaction(ctx context.Context, conn *sql.Conn, statements ...string) (results []sql.Result, err error) {
	log.Println("[DEBUG] ExecuteSqlInTransaction start")
	defer log.Println("[DEBUG] ExecuteSqlInTransaction end")

	err = db_common.BeginFunc(ctx, conn, func(tx *sql.Tx) error {
		for _, statement := range statements {
			result, err := tx.ExecContext(ctx, statement)
			if err != nil {
				return err
			}
			results = append(results, result)
		}
		return nil
	})
	return results, err
}

func ExecuteSqlWithArgsInTransaction(ctx context.Context, conn *sql.Conn, queries ...db_common.QueryWithArgs) (results []sql.Result, err error) {
	log.Println("[DEBUG] ExecuteSqlWithArgsInTransaction start")
	defer log.Println("[DEBUG] ExecuteSqlWithArgsInTransaction end")

	err = db_common.BeginFunc(ctx, conn, func(tx *sql.Tx) error {
		for _, q := range queries {
			result, err := tx.ExecContext(ctx, q.Query, q.Args...)
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
