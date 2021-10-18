package cmdconfig

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
	"github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/steampipeconfig"
)

// InitViper :: initializes and configures an instance of viper
func InitViper() {
	v := viper.GetViper()
	// set defaults
	v.Set(constants.ConfigKeyShowInteractiveOutput, true)

	if installDir, isSet := os.LookupEnv("STEAMPIPE_INSTALL_DIR"); isSet {
		v.SetDefault(constants.ArgInstallDir, installDir)
	} else {
		v.SetDefault(constants.ArgInstallDir, "~/.steampipe")
	}
}

// Viper :: fetches the global viper instance
func Viper() *viper.Viper {
	return viper.GetViper()
}

func SetViperDefaults(config *steampipeconfig.SteampipeConfig) {
	setBaseDefaults()
	if config != nil {
		overrideDefaultsFromConfig(config)
	}
	overrideDefaultsFromEnv()
}

// for keys which do not have a corresponding command flag, we need a separate defaulting mechanism
func setBaseDefaults() {
	defaults := map[string]interface{}{
		constants.ArgUpdateCheck: true,
		constants.ArgInstallDir:  constants.DefaultInstallDir,
	}

	for k, v := range defaults {
		viper.SetDefault(k, v)
	}
}

func overrideDefaultsFromConfig(config *steampipeconfig.SteampipeConfig) {
	for k, v := range config.ConfigMap() {
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
		constants.EnvUpdateCheck:                     {constants.ArgUpdateCheck, "bool"},
		constants.EnvInstallDir:                      {constants.ArgInstallDir, "string"},
		constants.EnvDatabaseBackendConnectionString: {constants.ArgConnectionString, "string"},
		constants.EnvServicePassword:                 {constants.ArgServicePassword, "string"},
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
