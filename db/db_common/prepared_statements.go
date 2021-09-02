package db_common

import (
	"context"
	"fmt"
	"strings"

	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/utils"
)

func CreatePreparedStatements(ctx context.Context, workspace WorkspaceResourceProvider, client Client) error {
	queryMap := workspace.GetQueryMap()
	controlMap := workspace.GetControlMap()

	utils.LogTime("db.CreatePreparedStatements start")
	defer utils.LogTime("db.CreatePreparedStatements end")

	for name, query := range queryMap {
		// query map contains long and short names for queries - avoid dupes
		if !strings.HasPrefix(name, "query.") {
			continue
		}
		// remove trailing semicolons from sql as this breaks the prepare statement
		rawSql := strings.TrimRight(strings.TrimSpace(typehelpers.SafeString(query.SQL)), ";")
		sql := fmt.Sprintf("PREPARE %s AS (\n%s\n)", query.PreparedStatementName(), rawSql)
		// execute the query, passing 'true' to disable the spinner
		_, err := client.ExecuteSync(ctx, sql, true)
		if err != nil {
			return fmt.Errorf("failed to create prepared statements table %s: %v", name, err)
		}
	}

	for name, control := range controlMap {
		// query map contains long and short names for controls - avoid dupes
		if !strings.HasPrefix(name, "control.") ||
			// do not create prepared statements for controls which reference another query
			control.Query != nil {
			continue
		}
		if control.SQL == nil {
			continue
		}

		// remove trailing semicolons from sql as this breaks the prepare statement
		rawSql := strings.TrimRight(strings.TrimSpace(typehelpers.SafeString(control.SQL)), ";")
		sql := fmt.Sprintf("PREPARE %s AS (\n%s\n)", control.PreparedStatementName(), rawSql)
		// execute the query, passing 'true' to disable the spinner
		_, err := client.ExecuteSync(ctx, sql, true)
		if err != nil {
			return fmt.Errorf("failed to create prepared statements table %s: %v", name, err)
		}
	}

	// return context error - this enables calling code to respond to cancellation
	return ctx.Err()
}

// UpdatePreparedStatements first attempts to deallocate all prepared statements in workspace, then recreates them
func UpdatePreparedStatements(ctx context.Context, workspace WorkspaceResourceProvider, client Client) error {
	queryMap := workspace.GetQueryMap()
	controlMap := workspace.GetControlMap()
	utils.LogTime("db.UpdatePreparedStatements start")
	defer utils.LogTime("db.UpdatePreparedStatements end")

	for name, query := range queryMap {
		// query map contains long and short names for queries - avoid dupes
		if !strings.HasPrefix(name, "query.") {
			continue
		}
		sql := fmt.Sprintf("DEALLOCATE %s ", query.PreparedStatementName())
		// execute the query, passing 'true' to disable the spinner
		// ignore errors
		client.ExecuteSync(ctx, sql, true)
	}
	for name, control := range controlMap {
		// query map contains long and short names for controls - avoid dupes
		if !strings.HasPrefix(name, "control.") ||
			// do not create prepared statements for controls which reference another query
			control.Query != nil {
			continue
		}
		if control.Query != nil {
			continue
		}
		sql := fmt.Sprintf("DEALLOCATE %s ", control.PreparedStatementName())
		// execute the query, passing 'true' to disable the spinner
		// ignore errors
		client.ExecuteSync(ctx, sql, true)
	}

	// now recreate them
	return CreatePreparedStatements(ctx, workspace, client)

}
