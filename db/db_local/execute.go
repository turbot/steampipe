package db_local

import (
	"database/sql"

	"github.com/turbot/steampipe/constants"
)

func executeSqlAsRoot(statements ...string) ([]sql.Result, error) {
	var results []sql.Result
	rootClient, err := createLocalDbClient(&CreateDbOptions{Username: constants.DatabaseSuperUser})
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
