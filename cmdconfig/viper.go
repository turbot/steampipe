package cmdconfig

import (
	"strings"

	"github.com/spf13/viper"
	"github.com/turbot/steampipe/constants"
)

var viperWrapper *viper.Viper

// InitViper :: initializes and configures an instance of viper
func InitViper() {
	viper.GetViper().SetEnvPrefix("STEAMPIPE")
	viper.GetViper().AutomaticEnv()
	viper.GetViper().SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	// set defaults
	viper.GetViper().Set(constants.ShowInteractiveOutputConfigKey, true)
}

// Viper :: fetches the global viper instance
func Viper() *viper.Viper {
	return viper.GetViper()
}
