package metaquery

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/viper"
	pconstants "github.com/turbot/pipe-fittings/v2/constants"
	"github.com/turbot/steampipe-plugin-sdk/v5/sperr"
	"github.com/turbot/steampipe/v2/pkg/db/db_common"
)

// controls the cache in the connected FDW
func cacheControl(ctx context.Context, input *HandlerInput) error {
	if len(input.args()) == 0 {
		return showCache(ctx, input)
	}

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
	command := strings.ToLower(input.args()[0])
	switch command {
	case pconstants.ArgOn:
		serverSettings := input.Client.ServerSettings()
		if serverSettings != nil && !serverSettings.CacheEnabled {
			fmt.Println("Caching is disabled on the server.")
		}
		viper.Set(pconstants.ArgClientCacheEnabled, true)
		return db_common.SetCacheEnabled(ctx, true, conn)
	case pconstants.ArgOff:
		viper.Set(pconstants.ArgClientCacheEnabled, false)
		return db_common.SetCacheEnabled(ctx, false, conn)
	case pconstants.ArgClear:
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
	if seconds <= 0 {
		return sperr.New("TTL must be greater than 0")
	}
	if can, whyCannotSet := db_common.CanSetCacheTtl(input.Client.ServerSettings(), seconds); !can {
		fmt.Println(whyCannotSet)
	}
	sessionResult := input.Client.AcquireSession(ctx)
	if sessionResult.Error != nil {
		return sessionResult.Error
	}
	defer func() {
		// we need to do this in a closure, otherwise the ctx will be evaluated immediately
		// and not in call-time
		sessionResult.Session.Close(false)
		viper.Set(pconstants.ArgCacheTtl, seconds)
	}()
	return db_common.SetCacheTtl(ctx, time.Duration(seconds)*time.Second, sessionResult.Session.Connection.Conn())
}

func showCache(_ context.Context, input *HandlerInput) error {
	if input.Client.ServerSettings() != nil && !input.Client.ServerSettings().CacheEnabled {
		fmt.Println("Caching is disabled on the server.")
		return nil
	}

	currentStatusString := "off"
	action := "on"

	if !viper.IsSet(pconstants.ArgClientCacheEnabled) || viper.GetBool(pconstants.ArgClientCacheEnabled) {
		currentStatusString = "on"
		action = "off"
	}

	fmt.Printf(
		`Caching is %s. To turn it %s, type %s`,
		pconstants.Bold(currentStatusString),
		pconstants.Bold(action),
		pconstants.Bold(fmt.Sprintf(".cache %s", action)),
	)

	// add an empty line here so that the rendering buffer can start from the next line
	fmt.Println()

	return nil
}

func showCacheTtl(ctx context.Context, input *HandlerInput) error {
	if viper.IsSet(pconstants.ArgCacheTtl) {
		ttl := getEffectiveCacheTtl(input.Client.ServerSettings(), viper.GetInt(pconstants.ArgCacheTtl))
		fmt.Println("Cache TTL is", ttl, "seconds.")
	} else if input.Client.ServerSettings() != nil {
		serverTtl := input.Client.ServerSettings().CacheMaxTtl
		fmt.Println("Cache TTL is", serverTtl, "seconds.")
	}
	errorsAndWarnings := db_common.ValidateClientCacheTtl(input.Client)
	errorsAndWarnings.ShowWarnings()
	// we don't know what the setting is
	return nil
}

// getEffectiveCacheTtl returns the lower of the server TTL and the clientTtl
func getEffectiveCacheTtl(serverSettings *db_common.ServerSettings, clientTtl int) int {
	if serverSettings != nil {
		return int(math.Min(float64(serverSettings.CacheMaxTtl), float64(clientTtl)))
	}
	return clientTtl
}
