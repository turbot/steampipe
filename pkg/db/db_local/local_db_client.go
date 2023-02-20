package db_local

import (
	"context"
	"github.com/spf13/viper"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/db/db_client"
	"github.com/turbot/steampipe/pkg/schema"
	"github.com/turbot/steampipe/pkg/steampipeconfig"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/pkg/utils"
	"log"
)

// LocalDbClient wraps over DbClient
type LocalDbClient struct {
	db_client.DbClient
	invoker constants.Invoker
}

// GetLocalClient starts service if needed and creates a new LocalDbClient
func GetLocalClient(ctx context.Context, invoker constants.Invoker, onConnectionCallback db_client.DbConnectionCallback) (*LocalDbClient, *modconfig.ErrorAndWarnings) {
	utils.LogTime("db.GetLocalClient start")
	defer utils.LogTime("db.GetLocalClient end")

	// start db if necessary
	if err := EnsureDBInstalled(ctx); err != nil {
		return nil, modconfig.NewErrorsAndWarning(err)
	}

	startResult := StartServices(ctx, viper.GetInt(constants.ArgDatabasePort), ListenTypeLocal, invoker)
	if startResult.Error != nil {
		return nil, &startResult.ErrorAndWarnings
	}

	client, err := newLocalClient(ctx, invoker, onConnectionCallback)
	if err != nil {
		ShutdownService(ctx, invoker)
		startResult.Error = err
	}
	return client, &startResult.ErrorAndWarnings
}

// newLocalClient verifies that the local database instance is running and returns a LocalDbClient to interact with it
// (This FAILS if local service is not running - use GetLocalClient to start service first)
func newLocalClient(ctx context.Context, invoker constants.Invoker, onConnectionCallback db_client.DbConnectionCallback) (*LocalDbClient, error) {
	utils.LogTime("db.newLocalClient start")
	defer utils.LogTime("db.newLocalClient end")

	connString, err := getLocalSteampipeConnectionString(nil)
	if err != nil {
		return nil, err
	}
	dbClient, err := db_client.NewDbClient(ctx, connString, onConnectionCallback)
	if err != nil {
		log.Printf("[TRACE] error getting local client %s", err.Error())
		return nil, err
	}

	c := &LocalDbClient{DbClient: *dbClient, invoker: invoker}
	log.Printf("[TRACE] created local client %p", c)
	return c, nil
}

// Close implements Client
// close the connection to the database and shuts down the db service if we are the last connection
func (c *LocalDbClient) Close(ctx context.Context) error {
	if err := c.DbClient.Close(ctx); err != nil {
		return err
	}
	log.Printf("[TRACE] local client close complete")

	log.Printf("[TRACE] shutdown local service %v", c.invoker)
	ShutdownService(ctx, c.invoker)
	return nil
}

// GetSchemaFromDB for LocalDBClient optimises the schema extraction by extracting schema
// information for connections backed by distinct plugins and then fanning back out.
// NOTE: we can only do this optimisation for a LOCAL db connection as we have access to connection config
func (c *LocalDbClient) GetSchemaFromDB(ctx context.Context, schemas ...string) (*schema.Metadata, error) {
	// build a ConnectionSchemaMap object to identify the schemas to load
	connectionSchemaMap, err := steampipeconfig.NewConnectionSchemaMap()
	if err != nil {
		return nil, err
	}
	// get the unique schema - we use this to limit the schemas we load from the database
	schemas = connectionSchemaMap.UniqueSchemas()
	metadata, err := c.DbClient.GetSchemaFromDB(ctx, schemas...)

	// we now need to add in all other schemas which have the same schemas as those we have loaded
	for loadedSchema, otherSchemas := range connectionSchemaMap {
		// all 'otherSchema's have the same schema as loadedSchema
		exemplarSchema, ok := metadata.Schemas[loadedSchema]
		if !ok {
			// should can happen in the case of a dynamic plugin with no tables - use empty schema
			exemplarSchema = make(map[string]schema.TableSchema)
		}

		for _, s := range otherSchemas {
			metadata.Schemas[s] = exemplarSchema
		}
	}
	return metadata, nil

}

func (c *LocalDbClient) buildSchemasQuery(schemas []string) string {
	for idx, s := range schemas {
		schemas[idx] = fmt.Sprintf("'%s'", s)
	}

	// build the schemas filter clause
	schemaClause := ""
	if len(schemas) > 0 {
		schemaClause = fmt.Sprintf(`
    cols.table_schema in (%s)
	OR`, strings.Join(schemas, ","))
	}

	query := fmt.Sprintf(`
SELECT
    table_name,
    column_name,
    column_default,
    is_nullable,
    data_type,
	udt_name,
    table_schema,
    (COALESCE(pg_catalog.col_description(c.oid, cols.ordinal_position :: int),'')) as column_comment,
    (COALESCE(pg_catalog.obj_description(c.oid),'')) as table_comment
FROM
    information_schema.columns cols
LEFT JOIN
    pg_catalog.pg_namespace nsp ON nsp.nspname = cols.table_schema
LEFT JOIN
    pg_catalog.pg_class c ON c.relname = cols.table_name AND c.relnamespace = nsp.oid
WHERE %s
	LEFT(cols.table_schema,8) = 'pg_temp_'
`, schemaClause)
	return query
}

func (c *LocalDbClient) RefreshConnectionAndSearchPaths(ctx context.Context, forceUpdateConnectionNames ...string) *steampipeconfig.RefreshConnectionResult {
	statushooks.SetStatus(ctx, "Refreshing connections")
	res := c.refreshConnections(ctx, forceUpdateConnectionNames...)
	if res.Error != nil {
		return res
	}

	statushooks.SetStatus(ctx, "Setting up functions")
	if err := refreshFunctions(ctx); err != nil {
		res.Error = err
		return res
	}

	statushooks.SetStatus(ctx, "Loading schema")
	// reload the foreign schemas, in case they have changed
	if err := c.LoadSchemaNames(ctx); err != nil {
		res.Error = err
		return res
	}

	cloneSchema := `-- Function: clone_foreign_schema(text, text)

-- DROP FUNCTION clone_foreign_schema(text, text);
-- SELECT * FROM clone_foreign_schema('aws', 'aws2')
CREATE OR REPLACE FUNCTION clone_foreign_schema(
    source_schema text,
    dest_schema text,
    plugin_name text)
    RETURNS text AS
$BODY$

DECLARE
    src_oid          oid;
    object           text;
    dest_table       text;
    table_sql      text;
    columns_sql      text;
    type_            text;
    column_          text;
    res              text;
BEGIN

    -- Check that source_schema exists
    SELECT oid INTO src_oid
    FROM pg_namespace
    WHERE nspname = source_schema;
    IF NOT FOUND
    THEN
        RAISE EXCEPTION 'source schema % does not exist!', source_schema;
        RETURN '';
    END IF;

-- Create schema
    EXECUTE 'DROP SCHEMA IF EXISTS ' ||  dest_schema || ' CASCADE';
    EXECUTE 'CREATE SCHEMA ' || dest_schema;
    EXECUTE 'COMMENT ON SCHEMA ' || dest_schema || 'IS  ''steampipe plugin: ' || plugin_name;
    EXECUTE 'GRANT USAGE ON SCHEMA ' || dest_schema || ' TO steampipe_users';
    EXECUTE 'ALTER DEFAULT PRIVILEGES IN SCHEMA ' || dest_schema || 'GRANT SELECT ON TABLES TO steampipe_users';

-- Create tables
    FOR object IN
        SELECT TABLE_NAME::text
        FROM information_schema.tables
        WHERE table_schema = source_schema
          AND table_type = 'FOREIGN'

        LOOP
            columns_sql := '';

            FOR column_, type_ IN
                SELECT column_name::text, data_type::text
                FROM information_schema.COLUMNS
                WHERE table_schema = source_schema
                  AND TABLE_NAME = object

                LOOP

                    IF columns_sql <> ''
                    THEN
                        columns_sql = columns_sql || ',';

                    END IF;
                    columns_sql = columns_sql || column_ || ' ' || type_;

                END LOOP;

            dest_table := '"' || dest_schema || '".' || quote_ident(object);
            table_sql :='CREATE FOREIGN TABLE ' || dest_table || ' (' || columns_sql || ') SERVER steampipe';
            EXECUTE table_sql;

            SELECT CONCAT(res, table_sql, ';') into res;
        END LOOP;
    RETURN res;
END

$BODY$
    LANGUAGE plpgsql VOLATILE
                     COST 100;
`
	_, err := executeSqlAsRoot(ctx, cloneSchema)

	statushooks.SetStatus(ctx, "Loading steampipe connections")
	// load the connection state and cache it!
	connectionMap, _, err := steampipeconfig.GetConnectionState(c.ForeignSchemaNames())
	if err != nil {
		res.Error = err
		return res
	}
	res.ConnectionMap = connectionMap
	// set user search path first - client may fall back to using it
	statushooks.SetStatus(ctx, "Setting up search path")

	// we need to send a muted ctx here since this function selects from the database
	// which by default puts up a "Loading" spinner. We don't want that here
	mutedCtx := statushooks.DisableStatusHooks(ctx)
	err = c.setUserSearchPath(mutedCtx)
	if err != nil {
		res.Error = err
		return res
	}

	if err := c.SetRequiredSessionSearchPath(ctx); err != nil {
		res.Error = err
		return res
	}

	// if there is an unprocessed db backup file, restore it now
	if err := restoreDBBackup(ctx); err != nil {
		res.Error = err
		return res
	}

	return res
}