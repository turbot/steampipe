package cmdconfig

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
	"github.com/turbot/steampipe/cloud"
	"github.com/turbot/steampipe/constants"
)

func ValidateConnectionStringArgs() error {
	workspaceDatabase := viper.GetString(constants.ArgWorkspaceDatabase)
	if workspaceDatabase == "local" {
		// local database - nothing to do here
		return nil
	}
	connectionString := workspaceDatabase

	// so a backend was set - is it a connection string or a database name
	if !strings.HasPrefix(workspaceDatabase, "postgresql://") {
		// it must be a database name - verify the cloud token was provided
		cloudToken := viper.GetString(constants.ArgCloudToken)
		if cloudToken == "" {
			return fmt.Errorf("cannot resolve workspace: required argument '--%s' missing", constants.ArgCloudToken)
		}

		// so we have a database and a token - build the connection string and set it in viper
		var err error
		if connectionString, err = cloud.GetConnectionString(workspaceDatabase, cloudToken); err != nil {
			return err
		}
	}

	// now set the connection string in viper
	viper.Set(constants.ArgConnectionString, connectionString)

	return nil
}
