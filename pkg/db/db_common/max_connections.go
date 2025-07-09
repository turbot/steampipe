package db_common

import (
	"github.com/spf13/viper"
	pconstants "github.com/turbot/pipe-fittings/v2/constants"
	"github.com/turbot/steampipe/v2/pkg/constants"
)

func MaxDbConnections() int {
	maxParallel := constants.DefaultMaxConnections
	if viper.IsSet(pconstants.ArgMaxParallel) {
		maxParallel = viper.GetInt(pconstants.ArgMaxParallel)
	}
	return maxParallel
}
