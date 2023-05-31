package connection_state

import (
	"fmt"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/db/db_common"
	"github.com/turbot/steampipe/pkg/steampipeconfig"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
)

// GetConnectionStateTableCreateSql returns the sql to create the conneciton state table
func GetConnectionStateTableCreateSql() db_common.QueryWithArgs {
	query := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s.%s (
    			name TEXT PRIMARY KEY,
    			state TEXT NOT NULL,
 			    type TEXT NULL,
 			    import_schema TEXT NOT NULL,
    			error TEXT NULL,
    			plugin TEXT NOT NULL,
    			schema_mode TEXT NOT NULL,
    			schema_hash TEXT NULL,
    			comments_set BOOL DEFAULT FALSE,
    			connection_mod_time TIMESTAMPTZ NOT NULL,
    			plugin_mod_time TIMESTAMPTZ NOT NULL
    			);`, constants.InternalSchema, constants.ConnectionStateTable)
	return db_common.QueryWithArgs{Query: query}
}

// GetConnectionStateErrorSql returns the sql to set a connection to 'error'
func GetConnectionStateErrorSql(connectionName string, err error) db_common.QueryWithArgs {
	query := fmt.Sprintf(`UPDATE %s.%s
SET state = '%s',
	error = $1,
	connection_mod_time = now()
WHERE
	name = $2
	`,
		constants.InternalSchema, constants.ConnectionStateTable, constants.ConnectionStateError)
	args := []any{constants.ConnectionStateError, err.Error(), connectionName}
	return db_common.QueryWithArgs{query, args}
}

// GetIncompleteConnectionStateErrorSql returns the sql to set all incomplete connections to 'error' (unless they alre already in error)
func GetIncompleteConnectionStateErrorSql(err error) db_common.QueryWithArgs {
	query := fmt.Sprintf(`UPDATE %s.%s
SET state = '%s',
	error = $1,
	connection_mod_time = now()
WHERE
	state <> 'ready' 
AND state <> 'disabled' 
AND state <> 'error' 
	`,
		constants.InternalSchema, constants.ConnectionStateTable, constants.ConnectionStateError)
	args := []any{err.Error()}
	return db_common.QueryWithArgs{Query: query, Args: args}
}

// GetIncompleteConnectionStatePendingIncompleteSql returns the sql to set all incomplete connections to 'incomplete'
func GetIncompleteConnectionStatePendingIncompleteSql() db_common.QueryWithArgs {
	query := fmt.Sprintf(`UPDATE %s.%s
SET state = '%s',
	connection_mod_time = now(),
    error = null
WHERE
	state <> 'ready' 
AND state <> 'disabled' 
	`,
		constants.InternalSchema, constants.ConnectionStateTable, constants.ConnectionStatePendingIncomplete)

	return db_common.QueryWithArgs{Query: query}
}

// GetReadConnectionStatePendingSql returns the sql to set all ready connections to 'pending'
func GetReadConnectionStatePendingSql() db_common.QueryWithArgs {
	query := fmt.Sprintf(`UPDATE %s.%s
SET state = '%s',
	connection_mod_time = now(),
    error = null
WHERE
	state = 'ready' 
	`,
		constants.InternalSchema, constants.ConnectionStateTable, constants.ConnectionStatePending)

	return db_common.QueryWithArgs{Query: query}
}

// GetUpdateConnectionStateSql returns the sql to update the connection state in the able with the current properties
func GetUpdateConnectionStateSql(c *steampipeconfig.ConnectionState) db_common.QueryWithArgs {
	// upsert
	query := fmt.Sprintf(`INSERT INTO %s.%s (name, 
		state,
		type,
 		import_schema,
		error,
		plugin,
		schema_mode,
		schema_hash,
		comments_set,
		connection_mod_time,
		plugin_mod_time)
VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,now(),$10) 
ON CONFLICT (name) 
DO 
   UPDATE SET 
			  state = $2, 
 			  type = $3,
              import_schema = $4,		
 		      error = $5,
			  plugin = $6,
			  schema_mode = $7,
			  schema_hash = $8,
			  comments_set = $9,
			  connection_mod_time = now(),
			  plugin_mod_time = $10
`, constants.InternalSchema, constants.ConnectionStateTable)
	args := []any{
		c.ConnectionName,
		c.State,
		c.Type,
		c.ImportSchema,
		c.ConnectionError,
		c.Plugin,
		c.SchemaMode,
		c.SchemaHash,
		c.CommentsSet,
		c.PluginModTime}
	return db_common.QueryWithArgs{query, args}
}

func GetNewConnectionStateTableInsertSql(c *modconfig.Connection) db_common.QueryWithArgs {
	query := fmt.Sprintf(`INSERT INTO %s.%s (name, 
		state,
		type,
 		import_schema,
		error,
		plugin,
		schema_mode,
		schema_hash,
		comments_set,
		connection_mod_time,
		plugin_mod_time)
VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,now(),now()) 
`, constants.InternalSchema, constants.ConnectionStateTable)

	schemaMode := ""
	commentsSet := false
	schemaHash := ""
	args := []any{
		c.Name,
		constants.ConnectionStatePendingIncomplete,
		c.Type,
		c.ImportSchema,
		nil,
		c.Plugin,
		schemaMode,
		schemaHash,
		commentsSet,
	}

	return db_common.QueryWithArgs{
		Query: query,
		Args:  args,
	}

}
func GetSetConnectionStateSql(connectionName string, state string) db_common.QueryWithArgs {
	query := fmt.Sprintf(`UPDATE %s.%s 
    SET	state = '%s', 
	 	connection_mod_time = now()
    WHERE 
        name = $1
`,
		constants.InternalSchema, constants.ConnectionStateTable, state,
	)
	args := []any{connectionName}
	return db_common.QueryWithArgs{query, args}
}

func GetDeleteConnectionStateSql(connectionName string) db_common.QueryWithArgs {
	query := fmt.Sprintf(`DELETE FROM %s.%s WHERE NAME=$1`, constants.InternalSchema, constants.ConnectionStateTable)
	args := []any{connectionName}
	return db_common.QueryWithArgs{query, args}
}

func GetSetConnectionStateCommentLoadedSql(connectionName string, commentsLoaded bool) db_common.QueryWithArgs {
	query := fmt.Sprintf(`UPDATE  %s.%s
SET comments_set = $1
WHERE NAME=$2`, constants.InternalSchema, constants.ConnectionStateTable)
	args := []any{commentsLoaded, connectionName}
	return db_common.QueryWithArgs{query, args}
}
