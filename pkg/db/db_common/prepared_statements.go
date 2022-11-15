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

type PrepareStatementFailures struct {
	Failures map[string]error
	Error    error
}

func NewPrepareStatementFailures() *PrepareStatementFailures {
	return &PrepareStatementFailures{Failures: make(map[string]error)}
}

func CreatePreparedStatements(ctx context.Context, resourceMaps *modconfig.ResourceMaps, conn *pgx.Conn, combineSql bool) (error, *PrepareStatementFailures) {
	log.Printf("[TRACE] CreatePreparedStatements")

	utils.LogTime("db.CreatePreparedStatements start")
	defer utils.LogTime("db.CreatePreparedStatements end")

	// first get the SQL to create all prepared statements
	sqlMap := GetPreparedStatementsSQL(resourceMaps)
	if len(sqlMap) == 0 {
		return nil, nil
	}

	// map of prepared statement failures, keyed by query name
	failureMap := NewPrepareStatementFailures()

	// TODO KAI TEST TIMING FOR CLOUD
	for name, sql := range sqlMap {
		if _, err := conn.Prepare(ctx, name, sql); err != nil {
			failureMap.Failures[name] = err
		}
	}
	//
	//if combineSql {
	//	sql := strings.Join(maps.Values(sqlMap), ";\n")
	//	if _, err := conn.Exec(ctx, sql); err != nil {
	//		failureMap.Error = err
	//	}
	//} else {
	//	for name, sql := range sqlMap {
	//		if _, err := conn.Exec(ctx, sql); err != nil {
	//			failureMap.Failures[name] = err
	//		}
	//	}
	//}

	// return context error - this enables calling code to respond to cancellation
	return ctx.Err(), failureMap
}

func GetPreparedStatementsSQL(resourceMaps *modconfig.ResourceMaps) map[string]string {
	// TODO key by query name and have object with prepared statement name and sql
	// make map of resource name to create SQL
	sqlMap := make(map[string]string)
	for _, queryProvider := range resourceMaps.QueryProviders() {
		if !modconfig.QueryProviderIsParameterised(queryProvider) {
			continue
		}
		rawSql := cleanPreparedStatementCreateSQL(typehelpers.SafeString(queryProvider.GetSQL()))
		preparedStatementName := queryProvider.GetPreparedStatementName()

		sqlMap[preparedStatementName] = rawSql

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
