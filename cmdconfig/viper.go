package cmdconfig

import (
	"os"

	"github.com/spf13/viper"
	"github.com/turbot/steampipe/constants"
)

// InitViper :: initializes and configures an instance of viper
func InitViper(v *viper.Viper) {
	// set defaults
	v.Set(constants.ShowInteractiveOutputConfigKey, true)

	installDir, isSet := os.LookupEnv("STEAMPIPE_INSTALL_DIR")

	if isSet {
		v.SetDefault(constants.ArgInstallDir, installDir)
	} else {
		v.SetDefault(constants.ArgInstallDir, "~/.steampipe")
	}
}

// Viper :: fetches the global viper instance
func Viper() *viper.Viper {
	return viper.GetViper()
}
