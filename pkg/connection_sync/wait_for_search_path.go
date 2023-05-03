package connection_sync

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	typehelpers "github.com/turbot/go-kit/types"
	sdkplugin "github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/db/db_common"
	"github.com/turbot/steampipe/pkg/error_helpers"
	"github.com/turbot/steampipe/pkg/steampipeconfig"
)

// WaitForSearchPathHeadSchemas identifies the first connection in the search path for each plugin,
// and wait for these connections to be ready
// if any of the connections are in error state, return an error
// this is used to ensure unqualified queries and tables are resolved to the correct connection
func WaitForSearchPathHeadSchemas(ctx context.Context, client db_common.Client, searchPath []string) error {
	conn, err := client.AcquireConnection(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	connectionStateMap, err := steampipeconfig.LoadConnectionState(ctx, conn.Conn(), steampipeconfig.WithWaitForPending())
	if err != nil {
		return err
	}

	// build list of connections we must wait for
	requiredSchemas := getFirstSearchPathConnectionForPlugins(searchPath, connectionStateMap)
	return waitForSchemasToBeReady(ctx, conn, connectionStateMap, requiredSchemas)

}

func getFirstSearchPathConnectionForPlugins(searchPath []string, connectionStateMap steampipeconfig.ConnectionDataMap) []string {
	// build map of the connections which we must wait for:
	// for static plugins, just the first connection in the search path
	// for dynamic schemas all schemas in the search paths (as we do not know which schema may provide a given table)
	requiredSchemasMap := getFirstSearchPathConnectionMapForPlugins(searchPath, connectionStateMap)
	// convert this into a list
	var requiredSchemas []string
	for _, connections := range requiredSchemasMap {
		requiredSchemas = append(requiredSchemas, connections...)
	}
	return requiredSchemas
}

func waitForSchemasToBeReady(ctx context.Context, conn *pgxpool.Conn, connectionStateMap steampipeconfig.ConnectionDataMap, schemas []string) error {
	var loadingSchemas []string
	for _, connectionName := range schemas {
		connectionState, ok := connectionStateMap[connectionName]
		if !ok {
			// not expected but not impossible - state may have changed while we iterate
			continue
		}
		// is this connection still loading
		if !connectionState.Loaded() {
			loadingSchemas = append(loadingSchemas, connectionName)
		}
	}

	// are there any schemas still loading - if so wait
	if len(loadingSchemas) > 0 {
		// reload the connection state, waiting for the required schemas to be loaded
		connectionStateMap, err := steampipeconfig.LoadConnectionState(ctx, conn.Conn(), steampipeconfig.WithWaitUntilReady(loadingSchemas...))
		if err != nil {
			return err
		}
		// ensure all schemas we are waiting for are not in error state
		return checkConnectionErrors(schemas, connectionStateMap)
	}
	return nil
}

// if any of the given connections are in error state, return an error
func checkConnectionErrors(schemas []string, connectionStateMap steampipeconfig.ConnectionDataMap) error {
	var errors []error
	for _, connectionName := range schemas {
		connectionState, ok := connectionStateMap[connectionName]
		if !ok {
			// not expected but not impossible - state may have changed while we iterate
			continue
		}
		if connectionState.State == constants.ConnectionStateError {
			err := fmt.Errorf("connection '%s' failed to load: %s",
				connectionName, typehelpers.SafeString(connectionState.ConnectionError))
			errors = append(errors, err)
		}
	}
	return error_helpers.CombineErrors(errors...)
}

// getFirstSearchPathConnectionMapForPlugins builds map of plugin to the connections which must be loaded to ensure we can resolve unqualified queries
// for static plugins, just the first connection in the search path is included
// for dynamic schemas all search paths are included
func getFirstSearchPathConnectionMapForPlugins(searchPath []string, connectionStateMap steampipeconfig.ConnectionDataMap) map[string][]string {
	res := make(map[string][]string)
	for _, connectionName := range searchPath {
		// is this in the connection state map
		connectionState, ok := connectionStateMap[connectionName]
		if !ok {
			continue
		}

		// get the plugin
		plugin := connectionState.Plugin
		// if this is the first connection for this plugin, or this is a dynamic plugin, add to the result map
		if len(res[plugin]) == 0 || connectionState.SchemaMode == sdkplugin.SchemaModeDynamic {
			res[plugin] = append(res[plugin], connectionName)
		}
	}
	return res
}
