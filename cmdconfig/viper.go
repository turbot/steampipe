package cmdconfig

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
	"github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/filepaths"
)

// Viper fetches the global viper instance
func Viper() *viper.Viper {
	return viper.GetViper()
}

func SetViperDefaults(configMap map[string]interface{}) {
	setBaseDefaults()
	if configMap != nil {
		overrideDefaultsFromConfig(configMap)
	}
	overrideDefaultsFromEnv()
}

// for keys which do not have a corresponding command flag, we need a separate defaulting mechanism
func setBaseDefaults() {
	defaults := map[string]interface{}{
		constants.ArgUpdateCheck: true,
		constants.ArgInstallDir:  filepaths.DefaultInstallDir,
	}

	for k, v := range defaults {
		viper.SetDefault(k, v)
	}
}

func overrideDefaultsFromConfig(configMap map[string]interface{}) {
	for k, v := range configMap {
		viper.SetDefault(k, v)
	}
}

type envMapping struct {
	configVar string
	// "string", "int", "bool"
	varType string
}

func overrideDefaultsFromEnv() {
	// a map of known environment variables to map to viper keys
	envMappings := map[string]envMapping{
		constants.EnvUpdateCheck:       {constants.ArgUpdateCheck, "bool"},
		constants.EnvInstallDir:        {constants.ArgInstallDir, "string"},
		constants.EnvWorkspaceChDir:    {constants.ArgWorkspaceChDir, "string"},
		constants.EnvWorkspaceDatabase: {constants.ArgWorkspaceDatabase, "string"},
		constants.EnvCloudHost:         {constants.ArgCloudHost, "string"},
		constants.EnvCloudToken:        {constants.ArgCloudToken, "string"},
		constants.EnvServicePassword:   {constants.ArgServicePassword, "string"},
		constants.EnvCheckDisplayWidth: {constants.ArgCheckDisplayWidth, "int"},
		constants.EnvMaxParallel:       {constants.ArgMaxParallel, "int"},
	}
	for k, v := range envMappings {
		if val, ok := os.LookupEnv(k); ok {
			switch v.varType {
			case "string":
				viper.SetDefault(v.configVar, val)
			case "bool":
				if boolVal, err := types.ToBool(val); err == nil {
					viper.SetDefault(v.configVar, boolVal)
				}
			case "int":
				if intVal, err := types.ToInt64(val); err == nil {
					viper.SetDefault(v.configVar, intVal)
				}
			default:
				// must be an invalid value in the map above
				panic(fmt.Sprintf("invalid env var mapping type: %s", v.varType))
			}
		}
	}
}
