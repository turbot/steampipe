package db_local

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/viper"
	"github.com/turbot/pipe-fittings/v2/modconfig"
	"github.com/turbot/steampipe/v2/pkg/constants"
	"github.com/turbot/steampipe/v2/pkg/db/db_common"
	"github.com/turbot/steampipe/v2/pkg/steampipeconfig"
)

func SetUserSearchPath(ctx context.Context, pool *pgxpool.Pool) ([]string, error) {
	var searchPath []string

	// is there a user search path in the config?
	// check ConfigKeyDatabaseSearchPath config (this is the value specified in the database config)
	if viper.IsSet(constants.ConfigKeyServerSearchPath) {
		searchPath = viper.GetStringSlice(constants.ConfigKeyServerSearchPath)
		// the Internal Schema should always go at the end
		searchPath = db_common.EnsureInternalSchemaSuffix(searchPath)
	} else {
		prefix := viper.GetStringSlice(constants.ConfigKeyServerSearchPathPrefix)
		// no config set - set user search path to default
		// - which is all the connection names, book-ended with public and internal
		searchPath = append(prefix, getDefaultSearchPath()...)
	}

	// escape the schema names
	escapedSearchPath := db_common.PgEscapeSearchPath(searchPath)

	log.Println("[TRACE] setting user search path to", searchPath)

	// get all roles which are a member of steampipe_users
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	query := fmt.Sprintf(`SELECT USENAME FROM pg_user WHERE pg_has_role(usename, '%s', 'member')`, constants.DatabaseUsersRole)
	rows, err := conn.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	// set the search path for all these roles
	var queries = []string{
		"LOCK TABLE pg_user IN SHARE ROW EXCLUSIVE MODE;",
	}

	for rows.Next() {
		var user string
		if err := rows.Scan(&user); err != nil {
			return nil, err
		}
		if user == "root" {
			continue
		}
		queries = append(queries, fmt.Sprintf(
			"ALTER USER %s SET SEARCH_PATH TO %s;",
			db_common.PgEscapeName(user),
			strings.Join(escapedSearchPath, ","),
		))
	}

	log.Printf("[TRACE] user search path sql: %v", queries)
	_, err = ExecuteSqlInTransaction(ctx, conn.Conn(), queries...)
	if err != nil {
		return nil, err
	}
	return searchPath, nil
}

// GetDefaultSearchPath builds default search path from the connection schemas, book-ended with public and internal
func getDefaultSearchPath() []string {
	// add all connections to the seatrch path (UNLESS ImportSchema is disabled)
	var searchPath []string

	// Check if GlobalConfig is initialized
	if steampipeconfig.GlobalConfig != nil {
		for connectionName, connection := range steampipeconfig.GlobalConfig.Connections {
			if connection.ImportSchema == modconfig.ImportSchemaEnabled {
				searchPath = append(searchPath, connectionName)
			}
		}
	}

	sort.Strings(searchPath)
	// add the 'public' schema as the first schema in the search_path. This makes it
	// easier for users to build and work with their own tables, and since it's normally
	// empty, doesn't make using steampipe tables any more difficult.
	searchPath = append([]string{"public"}, searchPath...)
	// add 'internal' schema as last schema in the search path
	searchPath = append(searchPath, constants.InternalSchema)

	return searchPath
}
