package steampipeconfig

import (
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/turbot/go-kit/helpers"

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
	TerminalOptions          *options.Terminal
	GeneralOptions           *options.General
}

// ConfigMap :: create a config map to pass to viper
func (c *SteampipeConfig) ConfigMap() map[string]interface{} {
	res := map[string]interface{}{}

	// build flat config map with order or precedence (low to high): general, database, terminal
	// this means if (for example) 'search-path' is set in both database and terminal options,
	// the value from terminal options will have precedence
	// however, we also store all values scoped by their options type, so we will store:
	// 'database.search-path', 'terminal.search-path' AND 'search-path' (which will be equal to 'terminal.search-path')
	if c.GeneralOptions != nil {
		c.populateConfigMapForOptions(c.GeneralOptions, res)
	}
	if c.DatabaseOptions != nil {
		c.populateConfigMapForOptions(c.DatabaseOptions, res)
	}
	if c.TerminalOptions != nil {
		c.populateConfigMapForOptions(c.TerminalOptions, res)
	}

	return res
}

// populate the config map for a given options object
// NOTE: this mutates configMap
func (c *SteampipeConfig) populateConfigMapForOptions(o options.Options, configMap map[string]interface{}) {
	for k, v := range o.ConfigMap() {
		configMap[k] = v
		// also store a scoped version of the config property
		configMap[getScopedKey(o, k)] = v
	}
}

// generated a scoped key for the config property. For example if o is a database options object and k is 'search-path'
// the scoped key will be 'database.search-path'
func getScopedKey(o options.Options, k string) string {
	t := reflect.TypeOf(helpers.DereferencePointer(o)).Name()
	return fmt.Sprintf("%s.%s", strings.ToLower(t), k)
}

func (c *SteampipeConfig) SetOptions(opts options.Options) {
	switch o := opts.(type) {
	case *options.Connection:
		if c.DefaultConnectionOptions == nil {
			c.DefaultConnectionOptions = o
		} else {
			c.DefaultConnectionOptions.Merge(o)
		}
	case *options.Database:
		if c.DatabaseOptions == nil {
			c.DatabaseOptions = o
		} else {
			c.DatabaseOptions.Merge(o)
		}
	case *options.Terminal:
		if c.TerminalOptions == nil {
			c.TerminalOptions = o
		} else {
			c.TerminalOptions.Merge(o)
		}
	case *options.General:
		if c.GeneralOptions == nil {
			c.GeneralOptions = o
		} else {
			c.GeneralOptions.Merge(o)
		}
	}
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

func (c *SteampipeConfig) String() string {
	var connectionStrings []string
	for _, c := range c.Connections {
		connectionStrings = append(connectionStrings, c.String())
	}

	str := fmt.Sprintf(`
Connections: 
%s
----
DefaultConnectionOptions:
%s`, strings.Join(connectionStrings, "\n"), c.DefaultConnectionOptions.String())

	if c.DatabaseOptions != nil {
		str += fmt.Sprintf(`

DatabaseOptions:
%s`, c.DatabaseOptions.String())
	}
	if c.TerminalOptions != nil {
		str += fmt.Sprintf(`

TerminalOptions:
%s`, c.TerminalOptions.String())
	}
	if c.GeneralOptions != nil {
		str += fmt.Sprintf(`

GeneralOptions:
%s`, c.GeneralOptions.String())
	}

	return str
}
