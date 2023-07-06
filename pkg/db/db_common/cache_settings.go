package db_common

import (
	"fmt"

	"github.com/spf13/viper"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
)

func ValidateClientCacheSettings(c Client) *modconfig.ErrorAndWarnings {
	cacheEnabledResult := ValidateClientCacheEnabled(c)
	cacheTtlResult := ValidateClientCacheTtl(c)

	return modconfig.EmptyErrorsAndWarning().Merge(cacheEnabledResult).Merge(cacheTtlResult)
}

func ValidateClientCacheEnabled(c Client) *modconfig.ErrorAndWarnings {

	if c.ServerSettings() == nil || !viper.IsSet(constants.ArgClientCacheEnabled) {
		// if there's no serverSettings, then this is a pre-21 server
		// behave as if there's no problem
		return modconfig.EmptyErrorsAndWarning()
	}

	errorsAndWarnings := modconfig.NewErrorsAndWarning(nil)
	if !c.ServerSettings().CacheEnabled && viper.GetBool(constants.ArgClientCacheEnabled) {
		errorsAndWarnings.AddWarning("Caching is disabled on the server.")
	}
	// if there's no serverSettings, then this is a pre-21 server
	// nothing to check
	return errorsAndWarnings
}

func ValidateClientCacheTtl(c Client) *modconfig.ErrorAndWarnings {
	errorsAndWarnings := modconfig.NewErrorsAndWarning(nil)

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
		return false, fmt.Sprintf("Server enforces maximum TTL of %d seconds. TTL set to %d seconds.", serverMaxTtl, serverMaxTtl)
	}
	return true, ""
}
