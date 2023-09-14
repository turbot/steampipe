package steampipeconfig

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/steampipe-plugin-sdk/v5/sperr"
	"log"
	"os"
	"strings"

	"github.com/hashicorp/go-version"
	"github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/error_helpers"
	"github.com/turbot/steampipe/pkg/ociinstaller"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/pkg/steampipeconfig/options"
)

// SteampipeConfig is a struct to hold Connection map and Steampipe options
type SteampipeConfig struct {
	// map of plugin configs, keyed by plugin image ref
	Plugins map[string][]*modconfig.Plugin
	// map of plugin configs, keyed by plugin config label
	PluginsByLabel map[string]*modconfig.Plugin
	// map of connection name to partially parsed connection config
	Connections map[string]*modconfig.Connection

	// Steampipe options
	DefaultConnectionOptions *options.Connection
	DatabaseOptions          *options.Database
	DashboardOptions         *options.GlobalDashboard
	TerminalOptions          *options.Terminal
	GeneralOptions           *options.General
	PluginOptions            *options.Plugin
	// TODO remove this  in 0.22
	// it is only needed due to conflicts with output name in terminal options
	// https://github.com/turbot/steampipe/issues/2534
	commandName string
}

func NewSteampipeConfig(commandName string) *SteampipeConfig {
	return &SteampipeConfig{
		Connections:    make(map[string]*modconfig.Connection),
		Plugins:        make(map[string][]*modconfig.Plugin),
		PluginsByLabel: make(map[string]*modconfig.Plugin),
		commandName:    commandName,
	}
}

// Validate validates all connections
// connections with validation errors are removed
func (c *SteampipeConfig) Validate() (validationWarnings, validationErrors []string) {
	for connectionName, connection := range c.Connections {
		// if the connection is an aggregator, populate the child connections
		// this resolves any wildcards in the connection list
		if connection.Type == modconfig.ConnectionTypeAggregator {
			connection.PopulateChildren(c.Connections)
		}
		w, e := connection.Validate(c.Connections)
		validationWarnings = append(validationWarnings, w...)
		validationErrors = append(validationErrors, e...)
		// if this connection validation remove
		if len(e) > 0 {
			delete(c.Connections, connectionName)
		}
	}

	return
}

// ConfigMap creates a config map to pass to viper
func (c *SteampipeConfig) ConfigMap() map[string]interface{} {
	res := modconfig.ConfigMap{}

	// build flat config map with order or precedence (low to high): general, database, terminal
	// this means if (for example) 'search-path' is set in both database and terminal options,
	// the value from terminal options will have precedence
	// however, we also store all values scoped by their options type, so we will store:
	// 'database.search-path', 'terminal.search-path' AND 'search-path' (which will be equal to 'terminal.search-path')
	if c.GeneralOptions != nil {
		res.PopulateConfigMapForOptions(c.GeneralOptions)
	}
	if c.DatabaseOptions != nil {
		res.PopulateConfigMapForOptions(c.DatabaseOptions)
	}
	if c.DashboardOptions != nil {
		res.PopulateConfigMapForOptions(c.DashboardOptions)
	}
	if c.TerminalOptions != nil {
		res.PopulateConfigMapForOptions(c.TerminalOptions)
	}
	if c.PluginOptions != nil {
		res.PopulateConfigMapForOptions(c.PluginOptions)
	}

	return res
}

func (c *SteampipeConfig) SetOptions(opts options.Options) (errorsAndWarnings *error_helpers.ErrorAndWarnings) {
	errorsAndWarnings = error_helpers.NewErrorsAndWarning(nil)

	switch o := opts.(type) {
	case *options.Connection:
		// TODO: remove in 0.21 [https://github.com/turbot/steampipe/issues/3251]
		errorsAndWarnings.AddWarning(deprecationWarning("connection options"))
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
	case *options.GlobalDashboard:
		if c.DashboardOptions == nil {
			c.DashboardOptions = o
		} else {
			c.DashboardOptions.Merge(o)
		}
	case *options.Terminal:
		// TODO: remove in 0.21 [https://github.com/turbot/steampipe/issues/3251]
		errorsAndWarnings.AddWarning(deprecationWarning("terminal options"))

		// NOTE: ignore terminal options if current command is not query
		// this is a short term workaround to handle the clashing 'output' argument
		// this will be refactored
		// TODO: remove in 0.21 [https://github.com/turbot/steampipe/issues/3251]
		if c.commandName != constants.CmdNameQuery {
			break
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
		// TODO: remove in 0.22 [https://github.com/turbot/steampipe/issues/3251]
		if c.GeneralOptions.MaxParallel != nil {
			errorsAndWarnings.AddWarning(deprecationWarning(fmt.Sprintf("'%s' in %s", constants.Bold("max_parallel"), constants.Bold("general options"))))
		}
	case *options.Plugin:
		if c.PluginOptions == nil {
			c.PluginOptions = o
		} else {
			c.PluginOptions.Merge(o)
		}

		// TODO: remove in 0.21 [https://github.com/turbot/steampipe/issues/3251]
		if c.GeneralOptions.MaxParallel != nil {
			errorsAndWarnings.AddWarning(deprecationWarning(fmt.Sprintf("'%s' in %s", constants.Bold("max_parallel"), constants.Bold("general options"))))
		}
	}
	return errorsAndWarnings
}

func deprecationWarning(subject string) string {
	if subject == "terminal options" {
		return fmt.Sprintf("%s has been deprecated and will be removed in a future version of Steampipe.\nThese can now be set in a steampipe %s.", constants.Bold(subject), constants.Bold("workspace"))
	}
	return fmt.Sprintf("%s has been deprecated and will be removed in a future version of Steampipe.\nThis can now be set in a steampipe %s.", subject, constants.Bold("workspace"))
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

	// As connection options are alco loaded by the FDW, which does not have access to viper,
	// we must manually apply env var defaulting

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
		// if DefaultConnectionOptions.CacheTTL value is NOT set, default it to true
		c.DefaultConnectionOptions.CacheTTL = &defaultTTL
	}
}

func (c *SteampipeConfig) GetConnectionOptions(connectionName string) *options.Connection {
	log.Printf("[TRACE] GetConnectionOptions for %s", connectionName)
	connection, ok := c.Connections[connectionName]
	if !ok {
		log.Printf("[TRACE] connection %s not found - returning default \n%v", connectionName, c.DefaultConnectionOptions)
		// if we can't find connection, just return defaults
		return c.DefaultConnectionOptions
	}
	// does the connection have connection options set - if not, return the default
	if connection.Options == nil {
		log.Printf("[TRACE] connection %s has no options - returning default \n%v", connectionName, c.DefaultConnectionOptions)
		return c.DefaultConnectionOptions
	}
	// so there are connection options, ensure all fields are set
	log.Printf("[TRACE] connection %s defines options %v", connectionName, connection.Options)

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
	if c.DashboardOptions != nil {
		str += fmt.Sprintf(`

DashboardOptions:
%s`, c.DashboardOptions.String())
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
	if c.PluginOptions != nil {
		str += fmt.Sprintf(`

PluginOptions:
%s`, c.PluginOptions.String())
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

func (c *SteampipeConfig) addPlugin(plugin *modconfig.Plugin, block *hcl.Block) error {
	if _, exists := c.PluginsByLabel[plugin.Label]; exists {
		return sperr.New("duplicate plugin: '%s' in '%s'", plugin.Source, block.TypeRange.Filename)
	}
	// get the _display_ image ref to key the map
	imageRef := plugin.GetImageRef()
	// add to list of plugin configs for this image ref
	c.Plugins[imageRef] = append(c.Plugins[imageRef], plugin)
	c.PluginsByLabel[plugin.Label] = plugin
	return nil
}

// ensure we have a plugin config struct for all plugins mentioned in connection config,
// even if there is not an explicit HCL config for it
func (c *SteampipeConfig) initializePlugins() map[string]error {

	var failedConnections = make(map[string]error)
	for _, connection := range c.Connections {
		plugin, err := c.resolvePluginForConnection(connection)
		if err != nil {
			failedConnections[connection.Name] = err
		} else {
			// set the Plugin property on the connection
			connection.Plugin = plugin.GetImageRef()
			connection.PluginConfig = plugin
		}
	}
	return failedConnections
}

func (c *SteampipeConfig) resolvePluginForConnection(connection *modconfig.Connection) (*modconfig.Plugin, error) {
	// NOTE: must set connection plugin property

	/* resolution steps:
		1) is there a plugin config with a name which matches the connection 'plugin' field
	 	2) is there a SINGLE plugin config for the image ref resolved from the connection 'plugin' field
	       NOTE: if there is more than one config for the plugin this is an error
		3) create an empty config for the plugin with the name set to the image ref
	*/

	// NOTE: at this point, c.Plugin is NOT populated, only c.PluginAlias
	// we populate c.Plugin AFTER resolving te plugin
	// 1) does this connection 'plugin' field refer to the label of a plugin config block
	if p := c.PluginsByLabel[connection.PluginAlias]; p != nil {
		return p, nil
	}

	// ok so there is no name match - treat the connection PluginAlias as an image ref
	imageRef := ociinstaller.NewSteampipeImageRef(connection.PluginAlias).DisplayImageRef()
	// is there a default plugin config for this plugin - if so the label will be the image ref
	if p := c.PluginsByLabel[imageRef]; p != nil {
		return p, nil
	}

	// no default config - check if there is configured config for this plugin
	pluginsForImageRef := c.Plugins[imageRef]

	switch len(pluginsForImageRef) {
	case 0:
		// there is no plugin config for this connection - add one
		p := modconfig.PluginForConnection(connection)
		// now add to our map
		// (NOTE: it;s ok to pass an empty HCL block - it is only used for the duplicate config error
		// and we know we will not get that
		c.addPlugin(p, &hcl.Block{})
		return p, nil

	case 1:
		return pluginsForImageRef[0], nil

	default:
		// so there is more than one plugin config for the plugin, and the connection DOES NOT specify which one to use
		// this is an error
		return nil, sperr.New("connection '%s' specifies plugin '%s' but there are %d plugin configs defined so the correct config cannot be resolved", connection.Name, connection.PluginAlias, len(pluginsForImageRef))

	}
}
