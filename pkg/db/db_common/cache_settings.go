package db_common

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
)

func ValidateClientCacheSettings(c Client) *modconfig.ErrorAndWarnings {
	errorsAndWarnings := modconfig.NewErrorsAndWarning(nil)
	errorsAndWarnings.Merge(ValidateClientCacheEnabled(c))
	errorsAndWarnings.Merge(ValidateClientCacheTtl(c))
	return errorsAndWarnings
}

func ValidateClientCacheEnabled(c Client) *modconfig.ErrorAndWarnings {
	errorsAndWarnings := modconfig.NewErrorsAndWarning(nil)

	if c.ServerSettings() == nil || !viper.IsSet(constants.ArgClientCacheEnabled) {
		// if there's no serverSettings, then this is a pre-21 server
		// behave as if there's no problem
		return errorsAndWarnings
	}

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

	serverMaxTtl := time.Duration(c.ServerSettings().CacheMaxTtl) * time.Second
	//nolint:golint,durationcheck //ArgCacheTtl is an int which is the number of TTL seconds
	clientTtl := viper.GetDuration(constants.ArgCacheTtl) * time.Second

	if serverMaxTtl < clientTtl {
		errorsAndWarnings.AddWarning(fmt.Sprintf("Server enforces maximum TTL at %d seconds. Setting TTL to %d seconds.", int(serverMaxTtl.Seconds()), int(serverMaxTtl.Seconds())))
	}
	return errorsAndWarnings
}
