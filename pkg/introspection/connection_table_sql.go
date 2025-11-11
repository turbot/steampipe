package introspection

import (
	"fmt"

	"github.com/turbot/pipe-fittings/v2/modconfig"
	"github.com/turbot/steampipe/v2/pkg/constants"
	"github.com/turbot/steampipe/v2/pkg/db/db_common"
	"github.com/turbot/steampipe/v2/pkg/steampipeconfig"
	"golang.org/x/exp/maps"
)

func GetConnectionStateTableDropSql() []db_common.QueryWithArgs {
	queryFormat := `DROP TABLE IF EXISTS %s.%s;`
	return getConnectionStateQueries(queryFormat, nil)
}

func GetConnectionStateTableCreateSql() []db_common.QueryWithArgs {
	queryFormat := `CREATE TABLE IF NOT EXISTS %s.%s (
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
);`
	return getConnectionStateQueries(queryFormat, nil)
}

// GetConnectionStateTableGrantSql returns the sql to setup SELECT permission for the 'steampipe_users' role
func GetConnectionStateTableGrantSql() []db_common.QueryWithArgs {
	queryFormat := fmt.Sprintf(
		`GRANT SELECT ON TABLE %%s.%%s TO %s;`,
		constants.DatabaseUsersRole,
	)
	return getConnectionStateQueries(queryFormat, nil)
}

// GetConnectionStateErrorSql returns the sql to set a connection to 'error'
func GetConnectionStateErrorSql(connectionName string, err error) []db_common.QueryWithArgs {
	queryFormat := fmt.Sprintf(`UPDATE %%s.%%s
SET state = '%s',
	error = $1,
	connection_mod_time = now()
WHERE
	name = $2
	`, constants.ConnectionStateError)

	args := []any{err.Error(), connectionName}
	return getConnectionStateQueries(queryFormat, args)
}

// GetIncompleteConnectionStateErrorSql returns the sql to set all incomplete connections to 'error' (unless they alre already in error)
func GetIncompleteConnectionStateErrorSql(err error) []db_common.QueryWithArgs {
	queryFormat := fmt.Sprintf(`UPDATE %%s.%%s
SET state = '%s',
	error = $1,
	connection_mod_time = now()
WHERE
	state <> 'ready' 
AND state <> 'disabled' 
AND state <> 'error' 
	`,
		constants.ConnectionStateError)
	args := []any{err.Error()}
	return getConnectionStateQueries(queryFormat, args)
}

// GetUpsertConnectionStateSql returns the sql to update the connection state in the able with the current properties
func GetUpsertConnectionStateSql(c *steampipeconfig.ConnectionState) []db_common.QueryWithArgs {
	// upsert
	queryFormat := `INSERT INTO %s.%s (name, 
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
			  
`
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
	return getConnectionStateQueries(queryFormat, args)
}

func GetNewConnectionStateFromConnectionInsertSql(c *modconfig.SteampipeConnection) []db_common.QueryWithArgs {
	queryFormat := `INSERT INTO %s.%s (name, 
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
`
	schemaMode := ""
	commentsSet := false
	schemaHash := ""

	args := []any{
		c.Name,
		constants.ConnectionStatePendingIncomplete,
		c.Type,
		maps.Keys(c.Connections),
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

	return getConnectionStateQueries(queryFormat, args)
}

func GetSetConnectionStateSql(connectionName string, state string) []db_common.QueryWithArgs {
	queryFormat := `UPDATE %s.%s
    SET	state = $1,
	 	connection_mod_time = now()
    WHERE
        name = $2
`

	args := []any{state, connectionName}
	return getConnectionStateQueries(queryFormat, args)
}

func GetDeleteConnectionStateSql(connectionName string) []db_common.QueryWithArgs {
	queryFormat := `DELETE FROM %s.%s WHERE NAME=$1`
	args := []any{connectionName}
	return getConnectionStateQueries(queryFormat, args)
}

func GetSetConnectionStateCommentLoadedSql(connectionName string, commentsLoaded bool) []db_common.QueryWithArgs {
	queryFormat := `UPDATE  %s.%s
SET comments_set = $1
WHERE NAME=$2`
	args := []any{commentsLoaded, connectionName}
	return getConnectionStateQueries(queryFormat, args)
}

func getConnectionStateQueries(queryFormat string, args []any) []db_common.QueryWithArgs {
	query := fmt.Sprintf(queryFormat, constants.InternalSchema, constants.ConnectionTable)
	legacyQuery := fmt.Sprintf(queryFormat, constants.InternalSchema, constants.LegacyConnectionStateTable)
	return []db_common.QueryWithArgs{
		{Query: query, Args: args},
		{Query: legacyQuery, Args: args},
	}
}
