package db_local

import (
	"context"
	"fmt"
	"log"
	"os"

	psutils "github.com/shirou/gopsutil/process"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/constants/runtime"
	"github.com/turbot/steampipe/pkg/db/db_common"
	"github.com/turbot/steampipe/pkg/error_helpers"
	"github.com/turbot/steampipe/pkg/filepaths"
	"github.com/turbot/steampipe/pkg/pluginmanager"
	"github.com/turbot/steampipe/pkg/statushooks"
	"github.com/turbot/steampipe/pkg/utils"
)

// StopStatus is a pseudoEnum for service stop result
type StopStatus int

const (
	// start from 1 to prevent confusion with int zero-value
	ServiceStopped StopStatus = iota + 1
	ServiceNotRunning
	ServiceStopFailed
	ServiceStopTimedOut
)

// ShutdownService stops the database instance if the given 'invoker' matches
func ShutdownService(ctx context.Context, invoker constants.Invoker) {
	utils.LogTime("db_local.ShutdownService start")
	defer utils.LogTime("db_local.ShutdownService end")

	if error_helpers.IsContextCanceled(ctx) {
		ctx = context.Background()
	}

	status, _ := GetState()

	// if the service is not running or it was invoked by 'steampipe service',
	// then we don't shut it down
	if status == nil || status.Invoker == constants.InvokerService {
		return
	}

	// how many clients are connected
	// under a fresh context
	clientCounts, err := GetClientCount(context.Background())
	// if there are other clients connected
	// and if there's no error
	if err == nil && clientCounts.SteampipeClients > 0 {
		// there are other steampipe clients connected to the database
		// we don't need to stop the service
		// the last one to exit will shutdown the service
		log.Printf("[INFO] ShutdownService not closing database service - %d steampipe %s connected", clientCounts.SteampipeClients, utils.Pluralize("client", clientCounts.SteampipeClients))
		return
	}

	// we can shut down the database
	stopStatus, err := StopServices(ctx, false, invoker)
	if err != nil {
		error_helpers.ShowError(ctx, err)
	}
	if stopStatus == ServiceStopped {
		return
	}

	// shutdown failed - try to force stop
	_, err = StopServices(ctx, true, invoker)
	if err != nil {
		error_helpers.ShowError(ctx, err)
	}

}

type ClientCount struct {
	SteampipeClients     int
	PluginManagerClients int
	TotalClients         int
}

// GetClientCount returns the number of connections to the service from anyone other than
// _this_execution_ of steampipe
//
// We assume that any connections from this execution will eventually be closed
// - if there are any other external connections, we cannot shut down the database
//
// this is to handle cases where either a third party tool is connected to the database,
// or other Steampipe sessions are attached to an already running Steampipe service
// - we do not want the db service being closed underneath them
//
// note: we need the PgClientAppName check to handle the case where there may be one or more open DB connections
// from this instance at the time of shutdown - for example when a control run is cancelled
// If we do not exclude connections from this execution, the DB will not be shut down after a cancellation
func GetClientCount(ctx context.Context) (*ClientCount, error) {
	utils.LogTime("db_local.GetClientCount start")
	defer utils.LogTime(fmt.Sprintf("db_local.GetClientCount end"))

	rootClient, err := CreateLocalDbConnection(ctx, &CreateDbOptions{Username: constants.DatabaseSuperUser})
	if err != nil {
		return nil, err
	}
	defer rootClient.Close(ctx)

	query := `
SELECT 
  application_name,
  count(*)
FROM 
  pg_stat_activity 
WHERE
	-- get only the network client processes
  client_port IS NOT NULL 
	AND
	-- which are client backends
  backend_type=$1 
	AND
	-- which are not connections from this application
  application_name!=$2
GROUP BY application_name
`

	counts := &ClientCount{}

	log.Println("[INFO] ClientConnectionAppName: ", runtime.ClientConnectionAppName)
	rows, err := rootClient.Query(ctx, query, "client backend", runtime.ClientConnectionAppName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var appName string
		var count int

		if err := rows.Scan(&appName, &count); err != nil {
			return nil, err
		}
		log.Printf("[INFO] appName: %s, count: %d", appName, count)

		counts.TotalClients += count

		if db_common.IsClientAppName(appName) {
			counts.SteampipeClients += count
		}

		// plugin manager uses the service prefix
		if db_common.IsServiceAppName(appName) {
			counts.PluginManagerClients += count
		}
	}

	return counts, nil
}

// StopServices searches for and stops the running instance. Does nothing if an instance was not found
func StopServices(ctx context.Context, force bool, invoker constants.Invoker) (status StopStatus, e error) {
	log.Printf("[TRACE] StopDB invoker %s, force %v", invoker, force)
	utils.LogTime("db_local.StopDB start")

	defer func() {
		if e == nil {
			os.Remove(filepaths.RunningInfoFilePath())
		}
		utils.LogTime("db_local.StopDB end")
	}()

	log.Println("[INFO] shutting down plugin manager")
	// stop the plugin manager
	// this means it may be stopped even if we fail to stop the service - that is ok - we will restart it if needed
	pluginManagerStopError := pluginmanager.Stop()
	log.Println("[INFO] shut down plugin manager")

	// stop the DB Service
	log.Println("[INFO] stopping DB Service")
	stopResult, dbStopError := stopDBService(ctx, force)
	log.Println("[INFO] stopped DB Service")

	return stopResult, error_helpers.CombineErrors(dbStopError, pluginManagerStopError)
}

func stopDBService(ctx context.Context, force bool) (StopStatus, error) {
	if force {
		// check if we have a process from another install-dir
		statushooks.SetStatus(ctx, "Checking for running instances…")
		// do not use a context that can be cancelled
		anyStopped := killPostgresInstanceIfAny(context.Background())
		// plugin manager is already stopped at this point, but we have seen instances where stray plugin manager
		// processes were left behind even after force stop. So we kill any leftover plugin manager processes(if any).
		// Adding this step adds 1 process call(in the best case scenario) but confirms that no plugin manager processes
		// are leftover.
		anyPluginManagerStopped := killPluginManagerInstanceIfAny(context.Background())
		if anyStopped || anyPluginManagerStopped {
			return ServiceStopped, nil
		}
		return ServiceNotRunning, nil
	}

	dbState, err := GetState()
	if err != nil {
		return ServiceStopFailed, err
	}

	if dbState == nil {
		// we do not have a info file
		// assume that the service is not running
		return ServiceNotRunning, nil
	}

	// GetStatus has made sure that the process exists
	process, err := psutils.NewProcess(int32(dbState.Pid))
	if err != nil {
		return ServiceStopFailed, err
	}

	err = doThreeStepPostgresExit(ctx, process)
	if err != nil {
		// we couldn't stop it still.
		// timeout
		return ServiceStopTimedOut, err
	}

	return ServiceStopped, nil
}
