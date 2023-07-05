package metaquery

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/spf13/viper"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/db/db_common"
	"github.com/turbot/steampipe/sperr"
)

// controls the cache in the connected FDW
func cacheControl(ctx context.Context, input *HandlerInput) error {
	if len(input.args()) == 0 {
		return showCache(ctx, input)
	}

	command := input.args()[0]
	// just get the active session from the connection pool
	// and set the cache parameters on it.
	// NOTE: this works because the interactive client
	// always has only one active connection due to the way it works
	sessionResult := input.Client.AcquireSession(ctx)
	if sessionResult.Error != nil {
		return sessionResult.Error
	}
	defer func() {
		// we need to do this in a closure, otherwise the ctx will be evaluated immediately
		// and not in call-time
		sessionResult.Session.Close(false)
	}()

	conn := sessionResult.Session.Connection.Conn()
	switch command {
	case constants.ArgOn:
		serverSettings := input.Client.ServerSettings()
		if serverSettings != nil && !serverSettings.CacheEnabled {
			fmt.Println("Caching is disabled on the server")
		}
		viper.Set(constants.ArgClientCacheEnabled, true)
		return db_common.SetCacheEnabled(ctx, true, conn)
	case constants.ArgOff:
		viper.Set(constants.ArgClientCacheEnabled, false)
		return db_common.SetCacheEnabled(ctx, false, conn)
	case constants.ArgClear:
		return db_common.CacheClear(ctx, conn)
	}

	return fmt.Errorf("invalid command")
}

// sets the cache TTL
func cacheTTL(ctx context.Context, input *HandlerInput) error {
	if len(input.args()) == 0 {
		return showCacheTtl(ctx, input)
	}
	seconds, err := strconv.Atoi(input.args()[0])
	if err != nil {
		return sperr.WrapWithMessage(err, "valid value is the number of seconds")
	}
	if seconds < 0 {
		return sperr.New("TTL must be greater than 0")
	}
	if input.Client.ServerSettings() != nil {
		serverttl := time.Duration(input.Client.ServerSettings().CacheMaxTtl) * time.Second
		newttl := time.Duration(seconds) * time.Second
		if newttl > serverttl {
			fmt.Println("Server enforces maximum TTL to", serverttl.Seconds(), "seconds. Setting to", serverttl.Seconds(), "seconds.")
			return nil
		}
	}
	sessionResult := input.Client.AcquireSession(ctx)
	if sessionResult.Error != nil {
		return sessionResult.Error
	}
	defer func() {
		// we need to do this in a closure, otherwise the ctx will be evaluated immediately
		// and not in call-time
		sessionResult.Session.Close(false)
		viper.Set(constants.ArgCacheTtl, seconds)
	}()
	return db_common.SetCacheTtl(ctx, time.Duration(seconds)*time.Second, sessionResult.Session.Connection.Conn())
}

func showCache(_ context.Context, input *HandlerInput) error {
	if input.Client.ServerSettings() != nil && !input.Client.ServerSettings().CacheEnabled {
		fmt.Println("Caching is disabled on the server")
		return nil
	}

	currentStatusString := "off"
	action := "on"

	if !viper.IsSet(constants.ArgClientCacheEnabled) || viper.GetBool(constants.ArgClientCacheEnabled) {
		currentStatusString = "on"
		action = "off"
	}

	fmt.Printf(
		`Caching is %s. To turn it %s, type %s`,
		constants.Bold(currentStatusString),
		constants.Bold(action),
		constants.Bold(fmt.Sprintf(".cache %s", action)),
	)

	// add an empty line here so that the rendering buffer can start from the next line
	fmt.Println()

	return nil
}

func showCacheTtl(ctx context.Context, input *HandlerInput) error {
	if viper.IsSet(constants.ArgCacheTtl) {
		// if there is a client override, show that
		//nolint:golint,durationcheck // ArgCacheTtl is an int of the number of seconds
		ttl := viper.GetDuration(constants.ArgCacheTtl) * time.Second
		fmt.Println("Cache TTL is overridden to", ttl, "at the client")
	} else {
		if input.Client.ServerSettings() != nil {
			serverTtl := time.Duration(input.Client.ServerSettings().CacheMaxTtl) * time.Second
			fmt.Println("Cache TTL is set to", serverTtl, "on the server")
		}
	}
	ew := db_common.ValidateClientCacheTtl(input.Client)
	if ew != nil {
		ew.ShowWarnings()
	}
	// we don't know what the setting is
	return nil
}
