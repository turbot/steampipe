package db_common

import (
	"github.com/spf13/viper"
	"github.com/turbot/pipe-fittings/constants"
)

func MaxDbConnections() int {
	maxParallel := constants.DefaultMaxConnections
	if viper.IsSet(constants.ArgMaxParallel) {
		maxParallel = viper.GetInt(constants.ArgMaxParallel)
	}
	return maxParallel
}
