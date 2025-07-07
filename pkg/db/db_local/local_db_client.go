package db_local

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/spf13/viper"
	pconstants "github.com/turbot/pipe-fittings/v2/constants"
	"github.com/turbot/pipe-fittings/v2/error_helpers"
	"github.com/turbot/pipe-fittings/v2/utils"
	"github.com/turbot/steampipe/v2/pkg/constants"
	"github.com/turbot/steampipe/v2/pkg/db/db_client"
	"github.com/turbot/steampipe/v2/pkg/db/db_common"
	pb "github.com/turbot/steampipe/v2/pkg/pluginmanager_service/grpc/proto"
)

// LocalDbClient wraps over DbClient
type LocalDbClient struct {
	db_client.DbClient
	notificationListener *db_common.NotificationListener
	invoker              constants.Invoker
}

// GetLocalClient starts service if needed and creates a new LocalDbClient
func GetLocalClient(ctx context.Context, invoker constants.Invoker, opts ...db_client.ClientOption) (*LocalDbClient, error_helpers.ErrorAndWarnings) {
	utils.LogTime("db.GetLocalClient start")
	defer utils.LogTime("db.GetLocalClient end")

	log.Printf("[INFO] GetLocalClient")
	defer log.Printf("[INFO] GetLocalClient complete")

	listenAddresses := StartListenType(ListenTypeLocal).ToListenAddresses()
	port := viper.GetInt(pconstants.ArgDatabasePort)
	log.Println(fmt.Sprintf("[TRACE] GetLocalClient - listenAddresses=%s, port=%d", listenAddresses, port))
	// start db if necessary
	if err := EnsureDBInstalled(ctx); err != nil {
		return nil, error_helpers.NewErrorsAndWarning(err)
	}

	log.Printf("[INFO] StartServices")
	startResult := StartServices(ctx, listenAddresses, port, invoker)
	if startResult.Error != nil {
		return nil, startResult.ErrorAndWarnings
	}

	log.Printf("[INFO] newLocalClient")
	client, err := newLocalClient(ctx, invoker, opts...)
	if err != nil {
		ShutdownService(ctx, invoker)
		startResult.Error = err
	}

	// after creating the client, refresh connections
	// NOTE: we cannot do this until after creating the client to ensure we do not miss notifications
	if startResult.Status == ServiceStarted {
		// ask the plugin manager to refresh connections
		// this is executed asyncronously by the plugin manager
		// we ignore this error, since RefreshConnections is async and all errors will flow through
		// the notification system
		// we do not expect any I/O errors on this since the PluginManager is running in the same box
		_, _ = startResult.PluginManager.RefreshConnections(&pb.RefreshConnectionsRequest{})
	}

	return client, startResult.ErrorAndWarnings
}

// newLocalClient verifies that the local database instance is running and returns a LocalDbClient to interact with it
// (This FAILS if local service is not running - use GetLocalClient to start service first)
func newLocalClient(ctx context.Context, invoker constants.Invoker, opts ...db_client.ClientOption) (*LocalDbClient, error) {
	utils.LogTime("db.newLocalClient start")
	defer utils.LogTime("db.newLocalClient end")

	connString, err := getLocalSteampipeConnectionString(nil)
	if err != nil {
		return nil, err
	}
	dbClient, err := db_client.NewDbClient(ctx, connString, opts...)
	if err != nil {
		log.Printf("[TRACE] error getting local client %s", err.Error())
		return nil, err
	}

	client := &LocalDbClient{DbClient: *dbClient, invoker: invoker}
	log.Printf("[INFO] created local client %p", client)

	if err := client.initNotificationListener(ctx); err != nil {
		client.Close(ctx)
		return nil, err
	}

	return client, nil
}

func (c *LocalDbClient) initNotificationListener(ctx context.Context) error {
	// get a connection for the notification cache
	conn, err := c.AcquireManagementConnection(ctx)
	if err != nil {
		c.Close(ctx)
		return err
	}
	// hijack from the pool  as we will be keeping open for the lifetime of this run
	// notification cache will manage the lifecycle of the connection
	notificationConnection := conn.Hijack()
	listener, err := db_common.NewNotificationListener(ctx, notificationConnection)
	if err != nil {
		return err
	}
	c.notificationListener = listener

	return nil
}

// Close implements Client
// close the connection to the database and shuts down the db service if we are the last connection
func (c *LocalDbClient) Close(ctx context.Context) error {
	if c.notificationListener != nil {
		c.notificationListener.Stop(ctx)
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
	c.notificationListener.RegisterListener(f)
}
