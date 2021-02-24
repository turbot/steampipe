package cmdconfig

import (
	"os"
	"strings"

	"github.com/spf13/viper"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/utils"
)

var globalViperInstance *viper.Viper

// InitViper :: initializes and configures an instance of viper
func InitViper(v *viper.Viper) {
	v.SetEnvPrefix("STEAMPIPE")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	// get this from the global instance
	cfgFile := viper.GetString(constants.ArgConfig)

	v.SetConfigFile(cfgFile)

	// If a config file is found, read it in.
	if _, err := os.Stat(cfgFile); err == nil {
		if err := v.ReadInConfig(); err != nil {
			utils.FailOnError(err)
		}
	}

	// set defaults
	v.Set(constants.ShowInteractiveOutputConfigKey, true)
}

// sets a global viper instance
func setConfig(v *viper.Viper) {
	globalViperInstance = v
}

// Viper :: fetches the global viper instance
func Viper() *viper.Viper {
	return globalViperInstance
}
