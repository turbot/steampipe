package db_local

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"sort"
	"strings"

	"github.com/spf13/viper"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/db/db_common"
	"github.com/turbot/steampipe/pkg/steampipeconfig"
	"golang.org/x/exp/maps"
)

func setUserSearchPath(ctx context.Context, pool *pgxpool.Pool) ([]string, error) {
	var searchPath []string

	// is there a user search path in the config?
	// check ConfigKeyDatabaseSearchPath config (this is the value specified in the database config)
	if viper.IsSet(constants.ConfigKeyServerSearchPath) {
		searchPath = viper.GetStringSlice(constants.ConfigKeyServerSearchPath)
		// add 'internal' schema as last schema in the search path
		searchPath = append(searchPath, constants.InternalSchema)
	} else {
		// no config set - set user search path to default
		// - which is all the connection names, book-ended with public and internal
		searchPath = getDefaultSearchPath()
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

	query := fmt.Sprintf(`select usename from pg_user where pg_has_role(usename, '%s', 'member')`, constants.DatabaseUsersRole)
	rows, err := conn.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	// set the search path for all these roles
	var queries = []string{
		"lock table pg_user;",
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
			"alter user %s set search_path to %s;",
			db_common.PgEscapeName(user),
			strings.Join(escapedSearchPath, ","),
		))
	}

	log.Printf("[TRACE] user search path sql: %v", queries)
	_, err = executeSqlInTransaction(ctx, conn.Conn(), queries...)
	if err != nil {
		return nil, err
	}
	return searchPath, nil
}

// GetDefaultSearchPath builds default search path from the connection schemas, book-ended with public and internal
func getDefaultSearchPath() []string {
	searchPath := maps.Keys(steampipeconfig.GlobalConfig.Connections)
	sort.Strings(searchPath)
	// add the 'public' schema as the first schema in the search_path. This makes it
	// easier for users to build and work with their own tables, and since it's normally
	// empty, doesn't make using steampipe tables any more difficult.
	searchPath = append([]string{"public"}, searchPath...)
	// add 'internal' schema as last schema in the search path
	searchPath = append(searchPath, constants.InternalSchema)

	return searchPath
}
