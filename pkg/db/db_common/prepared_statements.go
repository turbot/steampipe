package db_common

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/jackc/pgx/v4"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/pkg/utils"
)

func CreatePreparedStatements(ctx context.Context, resourceMaps *modconfig.ModResources, conn *pgx.Conn) (error, map[string]error) {
	log.Printf("[TRACE] CreatePreparedStatements")

	utils.LogTime("db.CreatePreparedStatements start")
	defer utils.LogTime("db.CreatePreparedStatements end")

	// first get the SQL to create all prepared statements
	sqlMap := GetPreparedStatementsSQL(resourceMaps)
	if len(sqlMap) == 0 {
		return nil, nil
	}

	// map of prepared statement failures, keyed by query name
	failureMap := make(map[string]error)
	for name, sql := range sqlMap {
		if _, err := conn.Exec(ctx, sql); err != nil {
			failureMap[name] = err
		}
	}

	// return context error - this enables calling code to respond to cancellation
	return ctx.Err(), failureMap
}

func GetPreparedStatementsSQL(resourceMaps *modconfig.ModResources) map[string]string {
	// make map of resource name to create SQL
	sqlMap := make(map[string]string)
	for _, queryProvider := range resourceMaps.QueryProviders() {
		if createSQL := getPreparedStatementCreateSql(queryProvider); createSQL != nil {
			sqlMap[queryProvider.Name()] = *createSQL
		}
	}
	return sqlMap
}

func getPreparedStatementCreateSql(queryProvider modconfig.QueryProvider) *string {
	// the query is a prepared statement if it defines its own sql and has parameters or (positional) arguments
	if !modconfig.QueryProviderIsParameterised(queryProvider) {
		return nil
	}

	// if the query provider has params, is MUST define SQL

	// remove trailing semicolons from sql as this breaks the prepare statement
	rawSql := cleanPreparedStatementCreateSQL(typehelpers.SafeString(queryProvider.GetSQL()))
	preparedStatementName := queryProvider.GetPreparedStatementName()
	createSQL := fmt.Sprintf("PREPARE %s AS (\n%s\n)", preparedStatementName, rawSql)
	return &createSQL
}

func cleanPreparedStatementCreateSQL(query string) string {
	rawSql := strings.TrimRight(strings.TrimSpace(query), ";")
	return rawSql
}
