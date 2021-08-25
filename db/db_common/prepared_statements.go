package db_common

import (
	"context"
	"fmt"
	"strings"

	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/utils"
)

func CreatePreparedStatements(ctx context.Context, queryMap map[string]*modconfig.Query, client Client) error {
	utils.LogTime("db.CreatePreparedStatements start")
	defer utils.LogTime("db.CreatePreparedStatements end")

	var sql []string
	for name, query := range queryMap {
		// query map contains long and short names for queries - avoid dupes
		if !strings.HasPrefix(name, "query.") {
			continue
		}
		sql = append(sql, fmt.Sprintf("PREPARE %s AS (\n%s\n)", query.ShortName, typehelpers.SafeString(query.SQL)))
	}
	// execute the query, passing 'true' to disable the spinner
	_, err := client.ExecuteSync(ctx, strings.Join(sql, "\n"), true)
	if err != nil {
		return fmt.Errorf("failed to create prepared statements tables: %v", err)
	}

	// return context error - this enables calling code to respond to cancellation
	return ctx.Err()
}

// UpdatePreparedStatements first attempts to deallocate all prepared statements in workspace, then recreates them
func UpdatePreparedStatements(ctx context.Context, queryMap map[string]*modconfig.Query, client Client) error {
	utils.LogTime("db.UpdatePreparedStatements start")
	defer utils.LogTime("db.UpdatePreparedStatements end")

	for name, query := range queryMap {
		// query map contains long and short names for queries - avoid dupes
		if !strings.HasPrefix(name, "query.") {
			continue
		}
		sql := fmt.Sprintf("DEALLOCATE %s ", query.ShortName)
		// execute the query, passing 'true' to disable the spinner
		// ignore errors
		client.ExecuteSync(ctx, sql, true)
	}

	// now recreate them
	return CreatePreparedStatements(ctx, queryMap, client)

}
