package db_common

import (
	"github.com/spf13/viper"
	constants2 "github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/steampipe/pkg/constants"
)

func MaxDbConnections() int {
	maxParallel := constants.DefaultMaxConnections
	if viper.IsSet(constants2.ArgMaxParallel) {
		maxParallel = viper.GetInt(constants2.ArgMaxParallel)
	}
	return maxParallel
}
