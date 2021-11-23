package cmdconfig

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/db/db_common"
)

func ValidateConnectionStringArgs() error {
	workspaceDatabase := viper.GetString(constants.ArgWorkspaceDatabase)
	if workspaceDatabase == "" {
		// no database set - so no connection string
		return nil
	}
	connectionString := workspaceDatabase

	// so a backend was set - is it a connection string or a database name
	if !strings.HasPrefix(workspaceDatabase, "postgresql://") {
		// it must be a database name - verify the cloud token was provided
		cloudToken := viper.GetString(constants.ArgCloudToken)
		if cloudToken == "" {
			return fmt.Errorf("if %s is not a connection string, %s must be set", constants.EnvWorkspaceDatabase, constants.EnvCloudToken)
		}

		// so we have a database and a token - build the connection string and set it in viper
		var err error
		if connectionString, err = db_common.GetConnectionString(workspaceDatabase, cloudToken); err != nil {
			return err
		}
	}

	// TODO SSL MODE
	// now set the connection string in viper
	viper.Set(constants.ArgConnectionString, connectionString+"?sslmode=require")

	return nil
}

//// VerifyLocalDb ensures a local db is being used
//func VerifyLocalDb() {
//	err := ValidateConnectionStringArgs()
//	utils.FailOnError()
//	connectionString := viper.GetString(constants.ArgConnectionString)
//	localDb := connectionString == ""
//	return localDb
//
//}
