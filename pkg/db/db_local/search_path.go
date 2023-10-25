package db_local

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/turbot/steampipe/pkg/constants"
	"log"
	"sort"
	"strings"

	"github.com/spf13/viper"
	"github.com/turbot/pipe-fittings/db_common"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/steampipe/pkg/steampipe_config_local"
)

func SetUserSearchPath(ctx context.Context, pool *sql.DB) ([]string, error) {
	var searchPath []string

	// is there a user search path in the config?
	// check ConfigKeyDatabaseSearchPath config (this is the value specified in the database config)
	if viper.IsSet(constants_steampipe.ConfigKeyServerSearchPath) {

		searchPath = viper.GetStringSlice(constants_steampipe.ConfigKeyServerSearchPath)
		// the Internal Schema should always go at the end
		searchPath = db_common.EnsureInternalSchemaSuffix(searchPath)
	} else {
		// no config set - set user search path to default
		// - which is all the connection names, book-ended with public and internal
		searchPath = getDefaultSearchPath()
	}

	// escape the schema names
	escapedSearchPath := db_common.PgEscapeSearchPath(searchPath)

	log.Println("[TRACE] setting user search path to", searchPath)

	// get all roles which are a member of steampipe_users
	conn, err := pool.Conn(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	query := fmt.Sprintf(`SELECT USENAME FROM pg_user WHERE pg_has_role(usename, '%s', 'member')`, constants_steampipe.DatabaseUsersRole)
	rows, err := conn.QueryContext(ctx, query)
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
	_, err = ExecuteSqlInTransaction(ctx, conn, queries...)
	if err != nil {
		return nil, err
	}
	return searchPath, nil
}

// GetDefaultSearchPath builds default search path from the connection schemas, book-ended with public and internal
func getDefaultSearchPath() []string {
	// add all connections to the seatrch path (UNLESS ImportSchema is disabled)
	var searchPath []string
	for connectionName, connection := range steampipe_config_local.GlobalConfig.Connections {
		if connection.ImportSchema == modconfig.ImportSchemaEnabled {
			searchPath = append(searchPath, connectionName)
		}
	}

	sort.Strings(searchPath)
	// add the 'public' schema as the first schema in the search_path. This makes it
	// easier for users to build and work with their own tables, and since it's normally
	// empty, doesn't make using steampipe tables any more difficult.
	searchPath = append([]string{"public"}, searchPath...)
	// add 'internal' schema as last schema in the search path
	searchPath = append(searchPath, constants_steampipe.InternalSchema)

	return searchPath
}
