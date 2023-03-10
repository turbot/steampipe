package db_local

import (
	"context"
	"github.com/spf13/viper"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/db/db_client"
	"github.com/turbot/steampipe/pkg/schema"
	"github.com/turbot/steampipe/pkg/steampipeconfig"
	"github.com/turbot/steampipe/pkg/utils"
	"log"
)

// LocalDbClient wraps over DbClient
type LocalDbClient struct {
	db_client.DbClient
	invoker constants.Invoker
}

// GetLocalClient starts service if needed and creates a new LocalDbClient
func GetLocalClient(ctx context.Context, invoker constants.Invoker, onConnectionCallback db_client.DbConnectionCallback) (*LocalDbClient, error) {
	utils.LogTime("db.GetLocalClient start")
	defer utils.LogTime("db.GetLocalClient end")

	// start db if necessary
	if err := EnsureDBInstalled(ctx); err != nil {
		return nil, err
	}

	startResult := StartServices(ctx, viper.GetInt(constants.ArgDatabasePort), ListenTypeLocal, invoker)
	if startResult.Error != nil {
		return nil, startResult.Error
	}

	client, err := NewLocalClient(ctx, invoker, onConnectionCallback)
	if err != nil {
		ShutdownService(ctx, invoker)
	}
	return client, err
}

// NewLocalClient verifies that the local database instance is running and returns a LocalDbClient to interact with it
func NewLocalClient(ctx context.Context, invoker constants.Invoker, onConnectionCallback db_client.DbConnectionCallback) (*LocalDbClient, error) {
	utils.LogTime("db.NewLocalClient start")
	defer utils.LogTime("db.NewLocalClient end")

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
// close the connection to the database and shuts down the backend if we are the last connection
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

// SetUserSearchPath sets the search path for the all steampipe users of the db service
// do this by finding all users assigned to the role steampipe_users and set their search path
