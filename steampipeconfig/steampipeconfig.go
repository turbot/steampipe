package steampipeconfig

import (
	"log"
	"os"
	"strings"

	"github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/steampipeconfig/options"
)

// SteampipeConfig :: Connection map and Steampipe settings
type SteampipeConfig struct {
	// map of connection name to partially parsed connection config
	Connections map[string]*Connection

	// Steampipe options
	DefaultConnectionOptions *options.Connection
	DatabaseOptions          *options.Database
	ConsoleOptions           *options.Console
	GeneralOptions           *options.General
}

func newSteampipeConfig() *SteampipeConfig {
	return &SteampipeConfig{
		Connections: make(map[string]*Connection),
	}
}

// ConfigMap :: create a config map to pass to viper
func (c *SteampipeConfig) ConfigMap() map[string]interface{} {
	res := map[string]interface{}{}
	for _, o := range []options.Options{c.DatabaseOptions, c.ConsoleOptions, c.GeneralOptions} {
		for k, v := range o.ConfigMap() {
			res[k] = v
		}
	}
	return res
}

func (c *SteampipeConfig) SetOptions(opts options.Options) {
	switch o := opts.(type) {
	case *options.Connection:
		c.DefaultConnectionOptions = o
	case *options.Database:
		c.DatabaseOptions = o
	case *options.Console:
		c.ConsoleOptions = o
	case *options.General:
		c.GeneralOptions = o
	}
}

const CacheEnabledEnvVar = "STEAMPIPE_CACHE"
const CacheTTLEnvVar = "STEAMPIPE_CACHE_TTL"
const defaultTTL = 300

// if default connection options have been set, assign them to any connection which do not define specific options
func (c *SteampipeConfig) setDefaultConnectionOptions() {
	if c.DefaultConnectionOptions == nil {
		c.DefaultConnectionOptions = &options.Connection{}
	}
	if c.DefaultConnectionOptions.Cache == nil {
		// if not default is set in the connection config, try the env var
		// default to 'enabled'
		var cacheEnabled = true

		if envStr, ok := os.LookupEnv(CacheEnabledEnvVar); ok {
			log.Printf("[WARN] setDefaultConnectionOptions READING ENV")
			cacheEnabled = strings.ToUpper(envStr) == "TRUE"
		}
		c.DefaultConnectionOptions.Cache = &cacheEnabled
	}
	if c.DefaultConnectionOptions.CacheTTL == nil {
		// if not default is set in the connection config, try the env var
		// default to 'enabled'
		var ttlSecs = defaultTTL
		if ttlString, ok := os.LookupEnv(CacheTTLEnvVar); ok {
			if parsed, err := types.ToInt64(ttlString); err == nil {
				ttlSecs = int(parsed)
			}
		}
		c.DefaultConnectionOptions.CacheTTL = &ttlSecs
	}

	// now find any connection with no options set and set the defaults
	for _, connection := range c.Connections {
		if connection.Options == nil {
			connection.Options = c.DefaultConnectionOptions
		} else {
			// so there is a connection options - check all parameters are set and default missing ones
			if connection.Options.Cache == nil {
				connection.Options.Cache = c.DefaultConnectionOptions.Cache
			}
			if connection.Options.CacheTTL == nil {
				connection.Options.CacheTTL = c.DefaultConnectionOptions.CacheTTL
			}
		}
	}
}
