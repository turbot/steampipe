package initialisation

import (
	"context"
	"strings"

	"github.com/spf13/viper"
	"github.com/turbot/pipe-fittings/v2/constants"
	"github.com/turbot/pipe-fittings/v2/pipes"
	"github.com/turbot/pipe-fittings/v2/steampipeconfig"
	"github.com/turbot/steampipe/v2/pkg/error_helpers"
)

func getPipesMetadata(ctx context.Context) (*steampipeconfig.PipesMetadata, error) {
	workspaceDatabase := viper.GetString(constants.ArgWorkspaceDatabase)
	if workspaceDatabase == "local" {
		// local database - nothing to do here
		return nil, nil
	}
	connectionString := workspaceDatabase

	var pipesMetadata *steampipeconfig.PipesMetadata

	// so a backend was set - is it a connection string or a database name
	workspaceDatabaseIsConnectionString := strings.HasPrefix(workspaceDatabase, "postgresql://") || strings.HasPrefix(workspaceDatabase, "postgres://")
	if !workspaceDatabaseIsConnectionString {
		// it must be a database name - verify the cloud token was provided
		cloudToken := viper.GetString(constants.ArgPipesToken)
		if cloudToken == "" {
			return nil, error_helpers.MissingCloudTokenError
		}

		// so we have a database and a token - build the connection string and set it in viper
		var err error
		if pipesMetadata, err = pipes.GetPipesMetadata(ctx, workspaceDatabase, cloudToken); err != nil {
			return nil, err
		}
		// read connection string out of pipesMetadata
		connectionString = pipesMetadata.ConnectionString
	}

	// now set the connection string in viper
	viper.Set(constants.ArgConnectionString, connectionString)

	return pipesMetadata, nil
}
