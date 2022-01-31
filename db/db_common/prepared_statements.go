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

func CreatePreparedStatements(ctx context.Context, resourceMaps *modconfig.WorkspaceResourceMaps, session *DatabaseSession) (err error, warnings []string) {
	log.Printf("[TRACE] CreatePreparedStatements")

	utils.LogTime("db.CreatePreparedStatements start")
	defer utils.LogTime("db.CreatePreparedStatements end")

	// first get the SQL to create all prepared statements
	sqlMap := GetPreparedStatementsSQL(resourceMaps)
	if len(sqlMap) == 0 {
		return nil, nil
	}

	for name, sql := range sqlMap {
		if _, err := session.Connection.ExecContext(ctx, sql); err != nil {
			warnings = append(warnings, fmt.Sprintf("failed to create prepared statement for %s: %v", name, err))
		}
	}

	// return context error - this enables calling code to respond to cancellation
	return ctx.Err(), warnings
}

func GetPreparedStatementsSQL(resourceMaps *modconfig.WorkspaceResourceMaps) map[string]string {
	// make map of resource name to create SQL
	sqlMap := make(map[string]string)
	for _, query := range resourceMaps.Queries {
		// if the query does not have parameters, it is NOT a prepared statement
		if len(query.Params) == 0 {
			continue
		}

		// remove trailing semicolons from sql as this breaks the prepare statement
		rawSql := cleanPreparedStatementSQL(typehelpers.SafeString(query.SQL))
		preparedStatementName := query.GetPreparedStatementName()
		sqlMap[query.FullName] = fmt.Sprintf("PREPARE %s AS (\n%s\n)", preparedStatementName, rawSql)
	}

	for _, control := range resourceMaps.Controls {
		// only create prepared statements for controls with inline SQL
		if control.SQL == nil {
			continue
		}
		// if the control does not have parameters, it is NOT a prepared statement
		if len(control.Params) == 0 {
			continue
		}

		// remove trailing semicolons from sql as this breaks the prepare statement
		rawSql := strings.TrimRight(strings.TrimSpace(typehelpers.SafeString(control.SQL)), ";")
		preparedStatementName := control.GetPreparedStatementName()
		sqlMap[control.FullName] = fmt.Sprintf("PREPARE %s AS (\n%s\n)", preparedStatementName, rawSql)
	}

	return sqlMap
}

func cleanPreparedStatementSQL(query string) string {
	rawSql := strings.TrimRight(strings.TrimSpace(query), ";")
	return rawSql
}
