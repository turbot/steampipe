package db_local

import (
	"context"
	"fmt"
	"log"

	"github.com/spf13/viper"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/db/db_client"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/pkg/utils"
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

	listenAddresses := StartListenType(ListenTypeLocal).ToListenAddresses()
	port := viper.GetInt(constants.ArgDatabasePort)
	log.Println(fmt.Sprintf("[TRACE] GetLocalClient - listenAddresses=%s, port=%d", listenAddresses, port))
	// start db if necessary
	if err := EnsureDBInstalled(ctx); err != nil {
		return nil, modconfig.NewErrorsAndWarning(err)
	}

	startResult := StartServices(ctx, listenAddresses, port, invoker)
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
