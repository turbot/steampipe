package db_common

import (
	"context"
	"fmt"
	"log"
	"strings"

	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/utils"
)

func CreatePreparedStatements(ctx context.Context, resourceMaps *modconfig.WorkspaceResourceMaps, client Client) error {
	log.Printf("[TRACE] CreatePreparedStatements")

	utils.LogTime("db.CreatePreparedStatements start")
	defer utils.LogTime("db.CreatePreparedStatements end")

	for name, query := range resourceMaps.QueryMap {
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
			log.Printf("[TRACE] failed to create prepared statement for %s: %v", name, err)
			return fmt.Errorf("failed to create prepared statement for %s: %v", name, err)
		}
	}

	for name, control := range resourceMaps.ControlMap {
		// query map contains long and short names for controls - avoid dupes
		if !strings.HasPrefix(name, "control.") {
			continue
		}
		// only create prepared statements for controls with inline SQL
		if control.SQL == nil {
			continue
		}

		// remove trailing semicolons from sql as this breaks the prepare statement
		rawSql := strings.TrimRight(strings.TrimSpace(typehelpers.SafeString(control.SQL)), ";")
		sql := fmt.Sprintf("PREPARE %s AS (\n%s\n)", control.PreparedStatementName(), rawSql)
		// execute the query, passing 'true' to disable the spinner
		_, err := client.ExecuteSync(ctx, sql, true)
		if err != nil {
			log.Printf("[TRACE] failed to create prepared statement for %s: %v", name, err)
			return fmt.Errorf("failed to create prepared statements for %s: %v", name, err)
		}
	}

	// return context error - this enables calling code to respond to cancellation
	return ctx.Err()
}

// UpdatePreparedStatements first attempts to deallocate all prepared statements in workspace, then recreates them
func UpdatePreparedStatements(ctx context.Context, resourceMaps *modconfig.WorkspaceResourceMaps, client Client) error {
	log.Printf("[TRACE] UpdatePreparedStatements")

	utils.LogTime("db.UpdatePreparedStatements start")
	defer utils.LogTime("db.UpdatePreparedStatements end")

	var sql []string
	for name, query := range resourceMaps.QueryMap {
		// query map contains long and short names for queries - avoid dupes
		if !strings.HasPrefix(name, "query.") {
			continue
		}
		sql = append(sql, fmt.Sprintf("DEALLOCATE %s;", query.PreparedStatementName()))
	}
	for name, control := range resourceMaps.ControlMap {
		// query map contains long and short names for controls - avoid dupes
		if !strings.HasPrefix(name, "control.") {
			continue
		}
		// do not create prepared statements for controls which reference another query
		if control.Query != nil {
			continue
		}
		sql = append(sql, fmt.Sprintf("DEALLOCATE %s;", control.PreparedStatementName()))
	}

	// execute the query, passing 'true' to disable the spinner
	s := strings.Join(sql, "\n")
	_, err := client.ExecuteSync(ctx, s, true)
	if err != nil {
		log.Printf("[TRACE] failed to update prepared statements - deallocate returned error %v", err)
		return err
	}

	// now recreate them
	return CreatePreparedStatements(ctx, resourceMaps, client)

}
