package steampipeconfig

import (
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
