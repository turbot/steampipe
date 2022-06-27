package steampipeconfig

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"

	"github.com/hashicorp/go-version"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/ociinstaller"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/pkg/steampipeconfig/options"
	"github.com/turbot/steampipe/pkg/utils"
)

// SteampipeConfig is a struct to hold Connection map and Steampipe options
type SteampipeConfig struct {
	// map of connection name to partially parsed connection config
	Connections map[string]*modconfig.Connection

	// Steampipe options
	DefaultConnectionOptions *options.Connection
	DatabaseOptions          *options.Database
	TerminalOptions          *options.Terminal
	GeneralOptions           *options.General
	commandName              string
}

func NewSteampipeConfig(commandName string) *SteampipeConfig {
	return &SteampipeConfig{
		Connections: make(map[string]*modconfig.Connection),
		commandName: commandName,
	}
}

func (c *SteampipeConfig) Validate() error {
	var validationErrors []string
	for _, connection := range c.Connections {

		// if the connection is an aggregator, populate the child connections
		// this resolves any wildcards in the connection list
		if connection.Type == modconfig.ConnectionTypeAggregator {
			connection.PopulateChildren(c.Connections)
		}
		validationErrors = append(validationErrors, connection.Validate(c.Connections)...)
	}
	if len(validationErrors) > 0 {
		return fmt.Errorf("config validation failed with %d %s: \n  - %s", len(validationErrors), utils.Pluralize("error", len(validationErrors)), strings.Join(validationErrors, "\n  - "))
	}
	return nil
}

// ConfigMap creates a config map to pass to viper
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
		// NOTE: do not load terminal options for check command
		// this is a short term workaround to handle the clashing 'output' argument
		// this will be refactored
		if c.commandName == "check" {
			return
		}
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
	if envStr, ok := os.LookupEnv(constants.EnvCacheEnabled); ok {
		if parsedEnv, err := types.ToBool(envStr); err == nil {
			c.DefaultConnectionOptions.Cache = &parsedEnv
		}
	}
	if c.DefaultConnectionOptions.Cache == nil {
		// if DefaultConnectionOptions.Cache value is NOT set, default it to true
		c.DefaultConnectionOptions.Cache = &defaultCacheEnabled
	}

	// if CacheTTLEnvVar is set, overwrite the value in DefaultConnectionOptions
	if ttlString, ok := os.LookupEnv(constants.EnvCacheTTL); ok {
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
	log.Printf("[TRACE] GetConnectionOptions")
	connection, ok := c.Connections[connectionName]
	if !ok {
		log.Printf("[TRACE] returning default %v", c.DefaultConnectionOptions)
		// if we can't find connection, jsy return defaults
		return c.DefaultConnectionOptions
	}
	// does the connection have connection options set - if not, return the default
	if connection.Options == nil {
		log.Printf("[TRACE] returning default %v", c.DefaultConnectionOptions)
		return c.DefaultConnectionOptions
	}
	// so there are connection options, ensure all fields are set
	log.Printf("[TRACE] connection defines options %v", connection.Options)

	// create a copy of the options to return
	result := &options.Connection{
		Cache:    c.DefaultConnectionOptions.Cache,
		CacheTTL: c.DefaultConnectionOptions.CacheTTL,
	}
	if connection.Options.Cache != nil {
		log.Printf("[TRACE] connection defines cache option %v", *connection.Options.Cache)
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

func (c *SteampipeConfig) ConnectionsForPlugin(pluginLongName string, pluginVersion *version.Version) []*modconfig.Connection {
	var res []*modconfig.Connection
	for _, con := range c.Connections {
		// extract stream from plugin
		ref := ociinstaller.NewSteampipeImageRef(con.Plugin)
		org, plugin, stream := ref.GetOrgNameAndStream()
		longName := fmt.Sprintf("%s/%s", org, plugin)
		if longName == pluginLongName {
			if stream == "latest" {
				res = append(res, con)
			} else {
				connectionPluginVersion, err := version.NewVersion(stream)
				if err != nil && connectionPluginVersion.LessThanOrEqual(pluginVersion) {
					res = append(res, con)
				}
			}
		}
	}
	return res
}

// ConnectionNames returns a flat list of connection names
func (c *SteampipeConfig) ConnectionNames() []string {
	res := make([]string, len(c.Connections))
	idx := 0
	for connectionName := range c.Connections {
		res[idx] = connectionName
		idx++
	}
	return res
}

func (c *SteampipeConfig) ConnectionList() []*modconfig.Connection {
	res := make([]*modconfig.Connection, len(c.Connections))
	idx := 0
	for _, c := range c.Connections {
		res[idx] = c
		idx++
	}
	return res
}
