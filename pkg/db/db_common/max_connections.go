package db_common

import (
	"github.com/spf13/viper"
	pconstants "github.com/turbot/pipe-fittings/v2/constants"
	"github.com/turbot/steampipe/v2/pkg/constants"
)

// MaxDbConnections returns the size of the connection pool, as set by --max-parallel or
// STEAMPIPE_MAX_PARALLEL (default: constants.DefaultMaxConnections).
func MaxDbConnections() (int32, error) {
	maxParallel := constants.DefaultMaxConnections
	if viper.IsSet(pconstants.ArgMaxParallel) {
		maxParallel = viper.GetInt(pconstants.ArgMaxParallel)
	}
	return int32(maxParallel), nil //nolint:gosec // G115: STEAMPIPE_MAX_PARALLEL is user-configurable with no upper bound, see #5052
}
