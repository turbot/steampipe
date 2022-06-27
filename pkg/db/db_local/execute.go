package db_local

import (
	"context"
	"database/sql"

	"github.com/turbot/steampipe/pkg/constants"
)

func executeSqlAsRoot(ctx context.Context, statements ...string) ([]sql.Result, error) {
	var results []sql.Result
	rootClient, err := createLocalDbClient(ctx, &CreateDbOptions{Username: constants.DatabaseSuperUser})
	if err != nil {
		return nil, err
	}
	defer rootClient.Close()
	for _, statement := range statements {
		result, err := rootClient.Exec(statement)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	return results, nil
}
