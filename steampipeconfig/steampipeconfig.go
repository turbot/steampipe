package steampipeconfig

import (
	"os"

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
	ConsoleOptions           *options.Terminal
	GeneralOptions           *options.General
	// array of options interfaces useful to build  ConfigMap
	Options []options.Options
}

func newSteampipeConfig() *SteampipeConfig {
	return &SteampipeConfig{
		Connections: make(map[string]*Connection),
	}
}

// ConfigMap :: create a config map to pass to viper
func (c *SteampipeConfig) ConfigMap() map[string]interface{} {
	res := map[string]interface{}{}
	for _, o := range c.Options {
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
	case *options.Terminal:
		c.ConsoleOptions = o
	case *options.General:
		c.GeneralOptions = o
	}
	c.Options = append(c.Options, opts)
}

const CacheEnabledEnvVar = "STEAMPIPE_CACHE"
const CacheTTLEnvVar = "STEAMPIPE_CACHE_TTL"

var defaultCacheEnabled = true
var defaultTTL = 300

// if default connection options have been set, assign them to any connection which do not define specific options
func (c *SteampipeConfig) setDefaultConnectionOptions() {
	if c.DefaultConnectionOptions == nil {
		c.DefaultConnectionOptions = &options.Connection{}
	}

	// precedence for the default is (high to low):
	// env var
	// default connection config
	// base default

	// if CacheEnabledEnvVar is set, overwrite the value in DefaultConnectionOptions
	if envStr, ok := os.LookupEnv(CacheEnabledEnvVar); ok {
		if parsedEnv, err := types.ToBool(envStr); err == nil {
			c.DefaultConnectionOptions.Cache = &parsedEnv
		}
	}
	if c.DefaultConnectionOptions.Cache == nil {
		// if DefaultConnectionOptions.Cache value is NOT set, default it to true
		c.DefaultConnectionOptions.Cache = &defaultCacheEnabled
	}

	// if CacheTTLEnvVar is set, overwrite the value in DefaultConnectionOptions
	if ttlString, ok := os.LookupEnv(CacheTTLEnvVar); ok {
		if parsed, err := types.ToInt64(ttlString); err == nil {
			ttl := int(parsed)
			c.DefaultConnectionOptions.CacheTTL = &ttl
		}
	}

	if c.DefaultConnectionOptions.CacheTTL == nil {
		// if DefaultConnectionOptions.Cache value is NOT set, default it to true
		c.DefaultConnectionOptions.CacheTTL = &defaultTTL
	}
}

func (c *SteampipeConfig) GetConnectionOptions(connectionName string) *options.Connection {
	connection, ok := c.Connections[connectionName]
	if !ok {
		// if we can't find connection, jsy return defaults
		return c.DefaultConnectionOptions
	}
	// does the connection have connection options set - if not, return the default
	if connection.Options == nil {
		return c.DefaultConnectionOptions
	}
	// so there are connection options, ensure all fields are set

	// create a copy of the options to return
	result := &options.Connection{
		Cache:    c.DefaultConnectionOptions.Cache,
		CacheTTL: c.DefaultConnectionOptions.CacheTTL,
	}
	if connection.Options.Cache != nil {
		result.Cache = connection.Options.Cache
	}
	if connection.Options.CacheTTL != nil {
		result.CacheTTL = connection.Options.CacheTTL
	}
	return result
}
