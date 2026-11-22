package db_common

import (
	"math"

	"github.com/spf13/viper"
	pconstants "github.com/turbot/pipe-fittings/v2/constants"
	"github.com/turbot/steampipe-plugin-sdk/v6/sperr"
	"github.com/turbot/steampipe/v2/pkg/constants"
)

// MaxDbConnections returns the size of the connection pool, as set by --max-parallel or
// STEAMPIPE_MAX_PARALLEL (default: constants.DefaultMaxConnections).
// The pool counts its connections in an int32, so a larger value is an error that names the setting and the maximum.
func MaxDbConnections() (int32, error) {
	maxParallel := constants.DefaultMaxConnections
	if viper.IsSet(pconstants.ArgMaxParallel) {
		maxParallel = viper.GetInt(pconstants.ArgMaxParallel)
	}
	// compare as int64 so the check also compiles where int is 32 bits
	v := int64(maxParallel)
	if v > math.MaxInt32 || v < math.MinInt32 {
		return 0, sperr.New("invalid value %d for the %s setting (--%s, %s or max_parallel): the maximum allowed value is %d",
			maxParallel, pconstants.ArgMaxParallel, pconstants.ArgMaxParallel, constants.EnvMaxParallel, int64(math.MaxInt32))
	}
	return int32(v), nil
}
