package cmdconfig

import (
	"fmt"
	"strings"

	"github.com/turbot/steampipe/steampipeconfig"

	"github.com/spf13/viper"
	"github.com/turbot/steampipe/cloud"
	"github.com/turbot/steampipe/constants"
)

func ValidateConnectionStringArgs() (*steampipeconfig.CloudMetadata, error) {
	workspaceDatabase := viper.GetString(constants.ArgWorkspaceDatabase)
	if workspaceDatabase == "local" {
		// local database - nothing to do here
		return nil, nil
	}
	connectionString := workspaceDatabase

	cloudMetadata := &steampipeconfig.CloudMetadata{}

	// so a backend was set - is it a connection string or a database name
	if !(strings.HasPrefix(workspaceDatabase, "postgresql://") || strings.HasPrefix(workspaceDatabase, "postgres://")) {
		// it must be a database name - verify the cloud token was provided
		cloudToken := viper.GetString(constants.ArgCloudToken)
		if cloudToken == "" {
			return nil, fmt.Errorf("cannot resolve workspace: required argument '--%s' missing", constants.ArgCloudToken)
		}

		// so we have a database and a token - build the connection string and set it in viper
		var err error
		if connectionString, err = cloud.GetConnectionString(workspaceDatabase, cloudToken); err != nil {
			return nil, err
		}
	}

	// now set the connection string in viper
	viper.Set(constants.ArgConnectionString, connectionString)

	return cloudMetadata, nil
}
