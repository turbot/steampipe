package db_local

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/spf13/viper"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/db/db_common"
	"log"
	"strings"
)

func setUserSearchPath(ctx context.Context, conn *pgx.Conn, foreignSchemaNames []string) error {
	var searchPath []string

	// is there a user search path in the config?
	// check ConfigKeyDatabaseSearchPath config (this is the value specified in the database config)
	if viper.IsSet(constants.ConfigKeyDatabaseSearchPath) {
		searchPath = viper.GetStringSlice(constants.ConfigKeyDatabaseSearchPath)
		// add 'internal' schema as last schema in the search path
		searchPath = append(searchPath, constants.InternalSchema)
	} else {
		// no config set - set user search path to default
		// - which is all the connection names, book-ended with public and internal
		searchPath = db_common.GetDefaultSearchPath(ctx, foreignSchemaNames)
	}

	// escape the schema names
	escapedSearchPath := db_common.PgEscapeSearchPath(searchPath)

	log.Println("[TRACE] setting user search path to", searchPath)

	// get all roles which are a member of steampipe_users
	query := fmt.Sprintf(`select usename from pg_user where pg_has_role(usename, '%s', 'member')`, constants.DatabaseUsersRole)
	rows, err := conn.Query(ctx, query)
	if err != nil {
		return err
	}

	// set the search path for all these roles
	var queries = []string{
		"lock table pg_user;",
	}

	for rows.Next() {
		var user string
		if err := rows.Scan(&user); err != nil {
			return err
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
	_, err = executeSqlAsRoot(ctx, queries...)
	if err != nil {
		return err
	}
	return nil
}
