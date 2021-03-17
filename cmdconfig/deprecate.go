package cmdconfig

import (
	"github.com/turbot/steampipe/constants"
)

// handle accessing deprecated options
func DatabasePort() int {
	port := Viper().GetInt(constants.ArgPort)
	if port == -1 {
		port = Viper().GetInt(constants.ArgPortDeprecated)
	}
	return port
}

func ListenAddress() string {
	listenAddress := Viper().GetString(constants.ArgListenAddress)
	if listenAddress == "" {
		listenAddress = Viper().GetString(constants.ArgListenAddressDeprecated)
	}
	return listenAddress
}
