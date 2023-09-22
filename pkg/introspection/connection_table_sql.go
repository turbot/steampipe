package introspection

import (
	"fmt"

	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/db/db_common"
	"github.com/turbot/steampipe/pkg/steampipeconfig"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
)

// GetLegacyConnectionStateTableDropSql returns the sql to drop the legacy connection state table
func GetLegacyConnectionStateTableDropSql() db_common.QueryWithArgs {
	query := fmt.Sprintf(`DROP TABLE IF EXISTS %s.%s;`, constants.InternalSchema, constants.LegacyConnectionStateTable)
	return db_common.QueryWithArgs{Query: query}
}

func GetConnectionStateTableDropSql() db_common.QueryWithArgs {
	query := fmt.Sprintf(`DROP TABLE IF EXISTS %s.%s;`, constants.InternalSchema, constants.ConnectionTable)
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
	plugin_instance TEXT NULL,
	schema_mode TEXT,
	schema_hash TEXT NULL,
	comments_set BOOL DEFAULT FALSE,
	connection_mod_time TIMESTAMPTZ,
	plugin_mod_time TIMESTAMPTZ,
	file_name TEXT, 
	start_line_number INTEGER, 
	end_line_number INTEGER
);`, constants.InternalSchema, constants.ConnectionTable)
	return db_common.QueryWithArgs{Query: query}
}

// GetConnectionStateTableGrantSql returns the sql to setup SELECT permission for the 'steampipe_users' role
func GetConnectionStateTableGrantSql() db_common.QueryWithArgs {
	return db_common.QueryWithArgs{Query: fmt.Sprintf(
		`GRANT SELECT ON TABLE %s.%s TO %s;`,
		constants.InternalSchema,
		constants.ConnectionTable,
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
		constants.InternalSchema, constants.ConnectionTable, constants.ConnectionStateError)
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
		constants.InternalSchema, constants.ConnectionTable, constants.ConnectionStateError)
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
		plugin_mod_time,
	    file_name,
	    start_line_number,
	    end_line_number)
VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,now(),$12,$13,$14,$15) 
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
			  plugin_mod_time = $12,
			  file_name = $13,
	    	  start_line_number = $14,
	     	  end_line_number = $15
			  
`, constants.InternalSchema, constants.ConnectionTable)
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
		c.FileName,
		c.StartLineNumber,
		c.EndLineNumber,
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
		plugin_mod_time,
		file_name,
	    start_line_number,
	    end_line_number)
VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,now(),now(),$12,$13,$14) 
`, constants.InternalSchema, constants.ConnectionTable)

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
		c.DeclRange.Filename,
		c.DeclRange.Start.Line,
		c.DeclRange.End.Line,
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
		constants.InternalSchema, constants.ConnectionTable, state,
	)
	args := []any{connectionName}
	return db_common.QueryWithArgs{Query: query, Args: args}
}

func GetDeleteConnectionStateSql(connectionName string) db_common.QueryWithArgs {
	query := fmt.Sprintf(`DELETE FROM %s.%s WHERE NAME=$1`, constants.InternalSchema, constants.ConnectionTable)
	args := []any{connectionName}
	return db_common.QueryWithArgs{Query: query, Args: args}
}

func GetSetConnectionStateCommentLoadedSql(connectionName string, commentsLoaded bool) db_common.QueryWithArgs {
	query := fmt.Sprintf(`UPDATE  %s.%s
SET comments_set = $1
WHERE NAME=$2`, constants.InternalSchema, constants.ConnectionTable)
	args := []any{commentsLoaded, connectionName}
	return db_common.QueryWithArgs{Query: query, Args: args}
}
