package db_common

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/error_helpers"
)

func ValidateClientCacheSettings(c Client) *error_helpers.ErrorAndWarnings {
	errorsAndWarnings := error_helpers.NewErrorsAndWarning(nil)
	errorsAndWarnings.Merge(ValidateClientCacheEnabled(c))
	errorsAndWarnings.Merge(ValidateClientCacheTtl(c))
	return errorsAndWarnings
}

func ValidateClientCacheEnabled(c Client) *error_helpers.ErrorAndWarnings {
	errorsAndWarnings := error_helpers.NewErrorsAndWarning(nil)

	if c.ServerSettings() == nil || !viper.IsSet(constants.ArgClientCacheEnabled) {
		// if there's no serverSettings, then this is a pre-21 server
		// behave as if there's no problem
		return errorsAndWarnings
	}

	if !c.ServerSettings().CacheEnabled && viper.GetBool(constants.ArgClientCacheEnabled) {
		errorsAndWarnings.AddWarning("client cache is enabled, but will have no effect since server cache is disabled")
	}

	// if there's no serverSettings, then this is a pre-21 server
	// nothing to check
	return errorsAndWarnings
}

func ValidateClientCacheTtl(c Client) *error_helpers.ErrorAndWarnings {
	errorsAndWarnings := error_helpers.NewErrorsAndWarning(nil)

	if c.ServerSettings() == nil || !viper.IsSet(constants.ArgCacheTtl) {
		// if there's no serverSettings, then this is a pre-21 server
		// behave as if there's no problem
		return errorsAndWarnings
	}

	serverMaxTtl := time.Duration(c.ServerSettings().CacheMaxTtl) * time.Second
	//nolint:golint,durationcheck //ArgCacheTtl is an int which is the number of TTL seconds
	clientTtl := viper.GetDuration(constants.ArgCacheTtl) * time.Second

	if serverMaxTtl < clientTtl {
		errorsAndWarnings.AddWarning(fmt.Sprintf("client cache TTL (%v) is higher than server cache TTL (%v) - server TTL is the effective TTL", clientTtl, serverMaxTtl))
	}
	return errorsAndWarnings
}
