package cmdconfig

import (
	"strings"

	"github.com/spf13/viper"
	"github.com/turbot/steampipe/constants"
)

var globalViperInstance *viper.Viper

// InitViper :: initializes and configures an instance of viper
func InitViper(v *viper.Viper) {
	v.SetEnvPrefix("STEAMPIPE")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

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
