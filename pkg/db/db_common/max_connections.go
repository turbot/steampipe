package db_common

import (
	"github.com/spf13/viper"
	"github.com/turbot/steampipe/pkg/constants"
)

func MaxDbConnections() int {
	maxParallel := constants_steampipe.DefaultMaxConnections
	if viper.IsSet(constants.ArgMaxParallel) {
		maxParallel = viper.GetInt(constants.ArgMaxParallel)
	}
	return maxParallel
}
