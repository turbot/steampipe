package introspection

import (
	"fmt"

	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/db/db_common"
	"github.com/turbot/steampipe/pkg/steampipeconfig"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
)

// GetConnectionStateTableDropSql returns the sql to create the conneciton state table
func GetConnectionStateTableDropSql() db_common.QueryWithArgs {
	query := fmt.Sprintf(`DROP TABLE IF EXISTS %s.%s;`, constants.InternalSchema, constants.ConnectionStateTable)
	return db_common.QueryWithArgs{Query: query}
}

func GetConnectionStateTableCreateSql() db_common.QueryWithArgs {
	query := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s.%s (
	name TEXT PRIMARY KEY,
	state TEXT,
	type TEXT NULL,
	connections TEXT[] NULL,
	import_schema TEXT,
	error TEXT NULL,
	plugin TEXT,
	plugin_instance TEXT,
	schema_mode TEXT,
	schema_hash TEXT NULL,
	comments_set BOOL DEFAULT FALSE,
	connection_mod_time TIMESTAMPTZ,
	plugin_mod_time TIMESTAMPTZ
);`, constants.InternalSchema, constants.ConnectionStateTable)
	return db_common.QueryWithArgs{Query: query}
}

// GetConnectionStateTableGrantSql returns the sql to setup SELECT permission for the 'steampipe_users' role
func GetConnectionStateTableGrantSql() db_common.QueryWithArgs {
	return db_common.QueryWithArgs{Query: fmt.Sprintf(
		`GRANT SELECT ON TABLE %s.%s TO %s;`,
		constants.InternalSchema,
		constants.ConnectionStateTable,
		constants.DatabaseUsersRole,
	)}
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
	return db_common.QueryWithArgs{Query: query, Args: args}
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

// GetUpsertConnectionStateSql returns the sql to update the connection state in the able with the current properties
func GetUpsertConnectionStateSql(c *steampipeconfig.ConnectionState) db_common.QueryWithArgs {
	// upsert
	query := fmt.Sprintf(`INSERT INTO %s.%s (name, 
		state,
		type,
 		connections,
 		import_schema,
		error,
		plugin,
		plugin_instance,
		schema_mode,
		schema_hash,
		comments_set,
		connection_mod_time,
		plugin_mod_time)
VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,now(),$12) 
ON CONFLICT (name) 
DO 
   UPDATE SET 
			  state = $2, 
 			  type = $3,
              connections = $4,
       		  import_schema = $5,		
 		      error = $6,
			  plugin = $7,
			  plugin_instance = $8,
			  schema_mode = $9,
			  schema_hash = $10,
			  comments_set = $11,
			  connection_mod_time = now(),
			  plugin_mod_time = $12
			  
`, constants.InternalSchema, constants.ConnectionStateTable)
	args := []any{
		c.ConnectionName,
		c.State,
		c.Type,
		c.Connections,
		c.ImportSchema,
		c.ConnectionError,
		c.Plugin,
		c.PluginInstance,
		c.SchemaMode,
		c.SchemaHash,
		c.CommentsSet,
		c.PluginModTime,
	}
	return db_common.QueryWithArgs{Query: query, Args: args}
}

func GetNewConnectionStateFromConnectionInsertSql(c *modconfig.Connection) db_common.QueryWithArgs {
	query := fmt.Sprintf(`INSERT INTO %s.%s (name, 
		state,
		type,
	    connections,
 		import_schema,
		error,
		plugin,
		plugin_instance,
		schema_mode,
		schema_hash,
		comments_set,
		connection_mod_time,
		plugin_mod_time)
VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,now(),now()) 
`, constants.InternalSchema, constants.ConnectionStateTable)

	schemaMode := ""
	commentsSet := false
	schemaHash := ""
	args := []any{
		c.Name,
		constants.ConnectionStatePendingIncomplete,
		c.Type,
		c.Connections,
		c.ImportSchema,
		nil,
		c.Plugin,
		c.PluginInstance,
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
	return db_common.QueryWithArgs{Query: query, Args: args}
}

func GetDeleteConnectionStateSql(connectionName string) db_common.QueryWithArgs {
	query := fmt.Sprintf(`DELETE FROM %s.%s WHERE NAME=$1`, constants.InternalSchema, constants.ConnectionStateTable)
	args := []any{connectionName}
	return db_common.QueryWithArgs{Query: query, Args: args}
}

func GetSetConnectionStateCommentLoadedSql(connectionName string, commentsLoaded bool) db_common.QueryWithArgs {
	query := fmt.Sprintf(`UPDATE  %s.%s
SET comments_set = $1
WHERE NAME=$2`, constants.InternalSchema, constants.ConnectionStateTable)
	args := []any{commentsLoaded, connectionName}
	return db_common.QueryWithArgs{Query: query, Args: args}
}
