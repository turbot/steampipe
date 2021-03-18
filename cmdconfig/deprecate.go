package cmdconfig

import (
	"github.com/spf13/viper"
	"github.com/turbot/steampipe/constants"
)

// handle accessing deprecated options
func DatabasePort() int {
	if viper.IsSet(constants.ArgPort) {
		return Viper().GetInt(constants.ArgPort)
	}
	return Viper().GetInt(constants.ArgPortDeprecated)
}

func ListenAddress() string {
	if viper.IsSet(constants.ArgPort) {
		return Viper().GetString(constants.ArgListenAddress)
	}
	return Viper().GetString(constants.ArgListenAddressDeprecated)
}
