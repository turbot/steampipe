package db_common

import (
	"fmt"
	"github.com/turbot/steampipe/pkg/db/steampipe_db_common"

	"github.com/spf13/viper"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/error_helpers"
)

func ValidateClientCacheSettings(c steampipe_db_common.Client) *error_helpers.ErrorAndWarnings {
	cacheEnabledResult := ValidateClientCacheEnabled(c)
	cacheTtlResult := ValidateClientCacheTtl(c)

	return error_helpers.EmptyErrorsAndWarning().Merge(cacheEnabledResult).Merge(cacheTtlResult)
}

func ValidateClientCacheEnabled(c steampipe_db_common.Client) *error_helpers.ErrorAndWarnings {
	errorsAndWarnings := error_helpers.EmptyErrorsAndWarning()
	if c.ServerSettings() == nil || !viper.IsSet(constants.ArgClientCacheEnabled) {
		// if there's no serverSettings, then this is a pre-21 server
		// behave as if there's no problem
		return errorsAndWarnings
	}

	if !c.ServerSettings().CacheEnabled && viper.GetBool(constants.ArgClientCacheEnabled) {
		errorsAndWarnings.AddWarning("Caching is disabled on the server.")
	}
	return errorsAndWarnings
}

func ValidateClientCacheTtl(c steampipe_db_common.Client) *error_helpers.ErrorAndWarnings {
	errorsAndWarnings := error_helpers.EmptyErrorsAndWarning()

	if c.ServerSettings() == nil || !viper.IsSet(constants.ArgCacheTtl) {
		// if there's no serverSettings, then this is a pre-21 server
		// behave as if there's no problem
		return errorsAndWarnings
	}

	clientTtl := viper.GetInt(constants.ArgCacheTtl)
	if can, whyCannotSet := CanSetCacheTtl(c.ServerSettings(), clientTtl); !can {
		errorsAndWarnings.AddWarning(whyCannotSet)
	}
	return errorsAndWarnings
}

func CanSetCacheTtl(ss *ServerSettings, newTtl int) (bool, string) {
	if ss == nil {
		// nothing to enforce
		return true, ""
	}
	serverMaxTtl := ss.CacheMaxTtl
	if newTtl > serverMaxTtl {
		return false, fmt.Sprintf("Server enforces maximum TTL of %d seconds. Cannot set TTL to %d seconds. TTL set to %d seconds.", serverMaxTtl, newTtl, serverMaxTtl)
	}
	return true, ""
}
