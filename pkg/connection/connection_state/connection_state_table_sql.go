package connection_state

import (
	"fmt"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/db/db_common"
	"github.com/turbot/steampipe/pkg/steampipeconfig"
)

func GetConnectionStateTableCreateSql() string {
	return fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s.%s (
    			name TEXT PRIMARY KEY,
-- 			    connection_type TEXT,
-- 			    child_connections TEXT[],
    			state TEXT NOT NULL,
    			error TEXT NULL,
    			plugin TEXT NOT NULL,
    			schema_mode TEXT NOT NULL,
    			schema_hash TEXT NULL,
    			comments_set BOOL DEFAULT FALSE,
    			connection_mod_time TIMESTAMPTZ NOT NULL,
    			plugin_mod_time TIMESTAMPTZ NOT NULL
    			);`, constants.InternalSchema, constants.ConnectionStateTable)
}

func GetConnectionStateErrorSql(connectionName string, err error) db_common.QueryWithArgs {
	query := fmt.Sprintf(`UPDATE %s.%s
SET state = 'error',
	error = $1,
	connection_mod_time = now()
WHERE
	name = $2
	`,
		constants.InternalSchema, constants.ConnectionStateTable)
	args := []any{constants.ConnectionStateError, err.Error(), connectionName}
	return db_common.QueryWithArgs{query, args}
}

func GetIncompleteConnectionStateErrorSql(err error) db_common.QueryWithArgs {
	query := fmt.Sprintf(`UPDATE %s.%s
SET state = 'error',
	error = $1,
	connection_mod_time = now()
WHERE
	state <> 'ready'
	`,
		constants.InternalSchema, constants.ConnectionStateTable)
	args := []any{err.Error()}
	return db_common.QueryWithArgs{Query: query, Args: args}
}

func GetStartUpdateConnectionStateSql(c *steampipeconfig.ConnectionState) db_common.QueryWithArgs {
	// if state is updating, set comments to false
	commentsSet := c.State == constants.ConnectionStateReady
	// upsert
	query := fmt.Sprintf(`INSERT INTO %s.%s (name, 
		state,
		error,
		plugin,
		schema_mode,
		schema_hash,
		comments_set,
		connection_mod_time,
		plugin_mod_time)
VALUES($1,$2,$3,$4,$5,$6,$7,now(),$8) 
ON CONFLICT (name) 
DO 
   UPDATE SET 
 			  state = $2, 
			  error = $3,
			  plugin = $4,
			  schema_mode = $5,
			  schema_hash = $6,
			  comments_set = $7,
			  connection_mod_time = now(),
			  plugin_mod_time = $8
`, constants.InternalSchema, constants.ConnectionStateTable)
	args := []any{c.Connection.Name, c.State, c.ConnectionError, c.Plugin, c.SchemaMode, c.SchemaHash, commentsSet, c.PluginModTime}
	return db_common.QueryWithArgs{query, args}
}

func GetSetConnectionReadySql(connection *steampipeconfig.ConnectionState) db_common.QueryWithArgs {
	query := fmt.Sprintf(`UPDATE %s.%s 
    SET	state = $1, 
	 	connection_mod_time = now(),
	 	plugin_mod_time = $2
    WHERE 
        name = $3
`,
		constants.InternalSchema, constants.ConnectionStateTable,
	)
	args := []any{constants.ConnectionStateReady, connection.PluginModTime, connection.ConnectionName}
	return db_common.QueryWithArgs{query, args}
}

func GetDeleteConnectionStateSql(connectionName string) db_common.QueryWithArgs {
	query := fmt.Sprintf(`DELETE FROM %s.%s WHERE NAME=$1`, constants.InternalSchema, constants.ConnectionStateTable)
	args := []any{connectionName}
	return db_common.QueryWithArgs{query, args}
}

func GetSetConnectionDeletingSql(connectionName string) db_common.QueryWithArgs {
	query := fmt.Sprintf(`UPDATE %s.%s 
    SET	state = 'deleting', 
	 	connection_mod_time = now()
    WHERE 
        name = $1
`,
		constants.InternalSchema, constants.ConnectionStateTable,
	)
	args := []any{connectionName}
	return db_common.QueryWithArgs{query, args}
}

func GetSetConnectionStateCommentLoadedSql(connectionName string, commentsLoaded bool) db_common.QueryWithArgs {
	query := fmt.Sprintf(`UPDATE  %s.%s
SET comments_loaded = $1
WHERE NAME=$2`, constants.InternalSchema, constants.ConnectionStateTable)
	args := []any{commentsLoaded, connectionName}
	return db_common.QueryWithArgs{query, args}
}
