package initialisation

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
	"github.com/turbot/steampipe/pkg/cloud"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/steampipeconfig"
)

func getCloudMetadata() (*steampipeconfig.CloudMetadata, error) {
	workspaceDatabase := viper.GetString(constants.ArgWorkspaceDatabase)
	if workspaceDatabase == "local" {
		// local database - nothing to do here
		return nil, nil
	}
	connectionString := workspaceDatabase

	var cloudMetadata *steampipeconfig.CloudMetadata

	workspaceDatabaseIsConnectionString := strings.HasPrefix(workspaceDatabase, "postgresql://") || strings.HasPrefix(workspaceDatabase, "postgres://")
	fetchSnapshotWorkspace := (viper.IsSet(constants.ArgShare) || viper.IsSet(constants.ArgSnapshot)) && !viper.IsSet(constants.ArgWorkspace)

	// so a backend was set - is it a connection string or a database name
	if fetchSnapshotWorkspace || !workspaceDatabaseIsConnectionString {
		// it must be a database name - verify the cloud token was provided
		cloudToken := viper.GetString(constants.ArgCloudToken)
		if cloudToken == "" {
			return nil, fmt.Errorf("cannot resolve workspace: required argument '--%s' missing", constants.ArgCloudToken)
		}

		// so we have a database and a token - build the connection string and set it in viper
		var err error
		if cloudMetadata, err = cloud.GetCloudMetadata(workspaceDatabase, cloudToken); err != nil {
			return nil, err
		}
		if !workspaceDatabaseIsConnectionString {
			// read connection string out of cloudMetadata
			connectionString = cloudMetadata.ConnectionString
		}
		if fetchSnapshotWorkspace {
			viper.Set(constants.ArgWorkspace, cloudMetadata.WorkspaceSnapshot)
		}
	}

	// now set the connection string in viper
	viper.Set(constants.ArgConnectionString, connectionString)

	return cloudMetadata, nil
}
