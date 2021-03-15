package cmdconfig

import (
	"strings"

	"github.com/spf13/viper"
	"github.com/turbot/steampipe/constants"
)

// InitViper :: initializes and configures an instance of viper
func InitViper() {
	v := viper.GetViper()
	v.SetEnvPrefix("STEAMPIPE")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	// set defaults
	v.Set(constants.ShowInteractiveOutputConfigKey, true)
}

// Viper :: fetches the global viper instance
func Viper() *viper.Viper {
	return viper.GetViper()
}
