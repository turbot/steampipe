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

func showCache(_ context.Context, input *HandlerInput) error {
	if input.Client.ServerSettings() != nil && !input.Client.ServerSettings().CacheEnabled {
		fmt.Println("Caching is disabled on the server")
		return nil
	}

	currentStatusString := "off"
	action := "on"
	if viper.GetBool(constants.ArgClientCacheEnabled) {
		currentStatusString = "on"
		action = "off"
	}

	fmt.Printf(
		`Caching is %s. To turn it %s, type %s\n`,
		constants.Bold(currentStatusString),
		constants.Bold(action),
		constants.Bold(fmt.Sprintf(".cache %s", action)),
	)

	return nil
}

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
			fmt.Println("Cannot turn on cache - caching is disabled on the server")
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
	seconds, err := strconv.Atoi(input.args()[0])
	if err != nil {
		return sperr.WrapWithMessage(err, "valid value is the number of seconds")
	}
	if seconds < 0 {
		return sperr.New("ttl must be greater than 0")
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
