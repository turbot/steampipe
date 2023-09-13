package db_local

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/turbot/steampipe/pkg/db/db_common"
	"log"

	"github.com/spf13/viper"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/db/db_client"
	"github.com/turbot/steampipe/pkg/error_helpers"
	"github.com/turbot/steampipe/pkg/utils"
)

// LocalDbClient wraps over DbClient
type LocalDbClient struct {
	db_client.DbClient
	NotificationCache *db_common.NotificationListener
	invoker           constants.Invoker
}

// GetLocalClient starts service if needed and creates a new LocalDbClient
func GetLocalClient(ctx context.Context, invoker constants.Invoker, onConnectionCallback db_client.DbConnectionCallback, opts ...db_client.ClientOption) (*LocalDbClient, *error_helpers.ErrorAndWarnings) {
	utils.LogTime("db.GetLocalClient start")
	defer utils.LogTime("db.GetLocalClient end")

	listenAddresses := StartListenType(ListenTypeLocal).ToListenAddresses()
	port := viper.GetInt(constants.ArgDatabasePort)
	log.Println(fmt.Sprintf("[TRACE] GetLocalClient - listenAddresses=%s, port=%d", listenAddresses, port))
	// start db if necessary
	if err := EnsureDBInstalled(ctx); err != nil {
		return nil, error_helpers.NewErrorsAndWarning(err)
	}

	startResult := StartServices(ctx, listenAddresses, port, invoker)
	if startResult.Error != nil {
		return nil, &startResult.ErrorAndWarnings
	}

	client, err := newLocalClient(ctx, invoker, onConnectionCallback, opts...)
	if err != nil {
		ShutdownService(ctx, invoker)
		startResult.Error = err
	}

	return client, &startResult.ErrorAndWarnings
}

// newLocalClient verifies that the local database instance is running and returns a LocalDbClient to interact with it
// (This FAILS if local service is not running - use GetLocalClient to start service first)
func newLocalClient(ctx context.Context, invoker constants.Invoker, onConnectionCallback db_client.DbConnectionCallback, opts ...db_client.ClientOption) (*LocalDbClient, error) {
	utils.LogTime("db.newLocalClient start")
	defer utils.LogTime("db.newLocalClient end")

	connString, err := getLocalSteampipeConnectionString(nil)
	if err != nil {
		return nil, err
	}
	dbClient, err := db_client.NewDbClient(ctx, connString, onConnectionCallback, opts...)
	if err != nil {
		log.Printf("[TRACE] error getting local client %s", err.Error())
		return nil, err
	}

	client := &LocalDbClient{DbClient: *dbClient, invoker: invoker}
	log.Printf("[TRACE] created local client %p", client)

	// get a connection for the notification cache
	conn, err := client.AcquireManagementConnection(ctx)
	if err != nil {
		client.Close(ctx)
		return nil, err
	}
	// hijack from the pool  as we will be keeping open for the lifetime of this run
	// notification cache will manage the lifecycle of the connection
	notificationConnection := conn.Hijack()
	client.NotificationCache, err = db_common.NewNotificationListener(ctx, notificationConnection)
	if err != nil {
		client.Close(ctx)
		return nil, err
	}

	return client, nil
}

// Close implements Client
// close the connection to the database and shuts down the db service if we are the last connection
func (c *LocalDbClient) Close(ctx context.Context) error {
	if c.NotificationCache != nil {
		c.NotificationCache.Stop(ctx)
	}

	if err := c.DbClient.Close(ctx); err != nil {
		return err
	}
	log.Printf("[TRACE] local client close complete")

	log.Printf("[TRACE] shutdown local service %v", c.invoker)
	ShutdownService(ctx, c.invoker)
	return nil
}

func (c *LocalDbClient) RegisterNotificationListener(f func(notification *pgconn.Notification)) {
	c.NotificationCache.RegisterListener(f)
}
