package db_local

import (
	"context"
	"github.com/jackc/pgx/v5"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/turbot/steampipe/pkg/constants"
)

func executeSqlAsRoot(ctx context.Context, statements ...string) ([]pgconn.CommandTag, error) {
	var results []pgconn.CommandTag
	rootClient, err := createLocalDbClient(ctx, &CreateDbOptions{Username: constants.DatabaseSuperUser})
	if err != nil {
		return nil, err
	}
	defer rootClient.Close(ctx)
	for _, statement := range statements {
		result, err := rootClient.Exec(ctx, statement)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	return results, nil
}

func executeSqlAsRootBatch(ctx context.Context, statements ...string) ([]pgconn.CommandTag, error) {
	var results []pgconn.CommandTag
	rootClient, err := createLocalDbClient(ctx, &CreateDbOptions{Username: constants.DatabaseSuperUser})
	if err != nil {
		return nil, err
	}
	defer rootClient.Close(ctx)
	batch := &pgx.Batch{}
	for _, s := range statements {
		batch.Queue(s)
	}
	batchResult := rootClient.SendBatch(ctx, batch)

	for range statements {
		result, err := batchResult.Exec()
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	batchResult.Close()

	return results, nil
}
