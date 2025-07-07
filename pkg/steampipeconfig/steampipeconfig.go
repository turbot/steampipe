package steampipeconfig

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/hashicorp/go-version"
	"github.com/turbot/go-kit/helpers"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/pipe-fittings/v2/constants"
	"github.com/turbot/pipe-fittings/v2/error_helpers"
	"github.com/turbot/pipe-fittings/v2/filepaths"
	"github.com/turbot/pipe-fittings/v2/modconfig"
	"github.com/turbot/pipe-fittings/v2/ociinstaller"
	poptions "github.com/turbot/pipe-fittings/v2/options"
	"github.com/turbot/pipe-fittings/v2/plugin"
	"github.com/turbot/pipe-fittings/v2/versionfile"
	"github.com/turbot/pipe-fittings/v2/workspace_profile"
	"github.com/turbot/steampipe-plugin-sdk/v5/sperr"
	"github.com/turbot/steampipe/v2/pkg/options"
)

// SteampipeConfig is a struct to hold Connection map and Steampipe options
type SteampipeConfig struct {
	// map of plugin configs, keyed by plugin image ref
	// (for each image ref we store an array of configs)
	Plugins map[string][]*plugin.Plugin
	// map of plugin configs, keyed by plugin instance
	PluginsInstances map[string]*plugin.Plugin
	// map of connection name to partially parsed connection config
	Connections map[string]*modconfig.SteampipeConnection

	// Steampipe options
	DatabaseOptions *options.Database
	GeneralOptions  *options.General
	PluginOptions   *options.Plugin
	// map of installed plugin versions, keyed by plugin image ref
	PluginVersions map[string]*versionfile.InstalledVersion
}

func NewSteampipeConfig(commandName string) *SteampipeConfig {
	return &SteampipeConfig{
		Connections:      make(map[string]*modconfig.SteampipeConnection),
		Plugins:          make(map[string][]*plugin.Plugin),
		PluginsInstances: make(map[string]*plugin.Plugin),
	}
}

// Validate validates all connections
// connections with validation errors are removed
func (c *SteampipeConfig) Validate() (validationWarnings, validationErrors []string) {
	for connectionName, connection := range c.Connections {
		// if the connection is an aggregator, populate the child connections
		// this resolves any wildcards in the connection list
		if connection.Type == modconfig.ConnectionTypeAggregator {
			aggregatorFailures := connection.PopulateChildren(c.Connections)
			validationWarnings = append(validationWarnings, aggregatorFailures...)
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
	res := workspace_profile.ConfigMap{}

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
	if c.PluginOptions != nil {
		res.PopulateConfigMapForOptions(c.PluginOptions)
	}

	return res
}

func (c *SteampipeConfig) SetOptions(opts poptions.Options) (errorsAndWarnings error_helpers.ErrorAndWarnings) {
	errorsAndWarnings = error_helpers.NewErrorsAndWarning(nil)

	switch o := opts.(type) {
	case *options.Database:
		if c.DatabaseOptions == nil {
			c.DatabaseOptions = o
		} else {
			c.DatabaseOptions.Merge(o)
		}
	case *options.General:
		if c.GeneralOptions == nil {
			c.GeneralOptions = o
		} else {
			c.GeneralOptions.Merge(o)
		}
	case *options.Plugin:
		if c.PluginOptions == nil {
			c.PluginOptions = o
		} else {
			c.PluginOptions.Merge(o)
		}
	}
	return errorsAndWarnings
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
`, strings.Join(connectionStrings, "\n"))

	if c.DatabaseOptions != nil {
		str += fmt.Sprintf(`

DatabaseOptions:
%s`, c.DatabaseOptions.String())
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

func (c *SteampipeConfig) ConnectionsForPlugin(pluginLongName string, pluginVersion *version.Version) []*modconfig.SteampipeConnection {
	var res []*modconfig.SteampipeConnection
	for _, con := range c.Connections {
		// extract constraint from plugin
		ref := ociinstaller.NewImageRef(con.Plugin)
		org, plugin, constraint := ref.GetOrgNameAndStream()
		longName := fmt.Sprintf("%s/%s", org, plugin)
		if longName == pluginLongName {
			if constraint == "latest" {
				res = append(res, con)
			} else {
				connectionPluginVersion, err := version.NewVersion(constraint)
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

func (c *SteampipeConfig) ConnectionList() []*modconfig.SteampipeConnection {
	res := make([]*modconfig.SteampipeConnection, len(c.Connections))
	idx := 0
	for _, c := range c.Connections {
		res[idx] = c
		idx++
	}
	return res
}

// add a plugin config to PluginsInstances and Plugins
// NOTE: this returns an error if we already have a config with the same label
func (c *SteampipeConfig) addPlugin(plugin *plugin.Plugin) error {
	if existingPlugin, exists := c.PluginsInstances[plugin.Instance]; exists {
		return duplicatePluginError(existingPlugin, plugin)
	}

	// get the image ref to key the map
	imageRef := plugin.Plugin

	pluginVersion, ok := c.PluginVersions[imageRef]
	if !ok {
		// just log it
		log.Printf("[WARN] addPlugin called for plugin '%s' which is not installed", imageRef)
		return nil
	}
	//  populate the version from the plugin version file data
	plugin.Version = pluginVersion.Version

	// add to list of plugin configs for this image ref
	c.Plugins[imageRef] = append(c.Plugins[imageRef], plugin)
	c.PluginsInstances[plugin.Instance] = plugin

	return nil
}

func duplicatePluginError(existingPlugin, newPlugin *plugin.Plugin) error {
	return sperr.New("duplicate plugin instance: '%s'\n\t(%s:%d)\n\t(%s:%d)",
		existingPlugin.Instance, *existingPlugin.FileName, *existingPlugin.StartLineNumber,
		*newPlugin.FileName, *newPlugin.StartLineNumber)
}

// ensure we have a plugin config struct for all plugins mentioned in connection config,
// even if there is not an explicit HCL config for it
// NOTE: this populates the  Plugin and PluginInstance field of the connections
func (c *SteampipeConfig) initializePlugins() {
	for _, connection := range c.Connections {
		plugin, err := c.resolvePluginInstanceForConnection(connection)
		if err != nil {
			log.Printf("[WARN] cannot resolve plugin for connection '%s': %s", connection.Name, err.Error())
			connection.Error = err
			continue
		}
		// if plugin is nil, but there is no error, it must be referring to a plugin which has no instance config
		// and is not installed - set the plugin error
		if plugin == nil {
			// set the Plugin to the image ref of the plugin
			connection.Plugin = ociinstaller.NewImageRef(connection.PluginAlias).DisplayImageRef()
			connection.Error = fmt.Errorf(constants.ConnectionErrorPluginNotInstalled)
			log.Printf("[INFO] connection '%s' requires plugin '%s' which is not loaded and has no instance config", connection.Name, connection.PluginAlias)
			continue
		}
		// set the PluginAlias on the connection

		// set the PluginAlias and Plugin property on the connection
		pluginImageRef := plugin.Plugin
		connection.PluginAlias = plugin.Alias
		connection.Plugin = pluginImageRef
		if pluginPath, _ := filepaths.GetPluginPath(pluginImageRef, plugin.Alias); pluginPath != "" {
			// plugin is installed - set the instance and the plugin path
			connection.PluginInstance = &plugin.Instance
			connection.PluginPath = &pluginPath
		} else {
			// set the plugin error
			connection.Error = fmt.Errorf(constants.ConnectionErrorPluginNotInstalled)
			// leave instance unset
			log.Printf("[INFO] connection '%s' requires plugin '%s' - this is not installed", connection.Name, plugin.Alias)
		}

	}

}

/*
	 find a plugin instance which satisfies the Plugin field of the connection
	  resolution steps:
		1) if PluginInstance is already set, the connection must have a HCL reference to a plugin block
	 		- just validate the block exists
		2) handle local???
		3) have we already created a default plugin config for this plugin
		4) is there a SINGLE plugin config for the image ref resolved from the connection 'plugin' field
	       NOTE: if there is more than one config for the plugin this is an error
		5) create a default config for the plugin (with the label set to the image ref)
*/
func (c *SteampipeConfig) resolvePluginInstanceForConnection(connection *modconfig.SteampipeConnection) (*plugin.Plugin, error) {
	// NOTE: at this point, c.Plugin is NOT populated, only either c.PluginAlias or c.PluginInstance
	// we populate c.Plugin AFTER resolving the plugin

	// if PluginInstance is already set, the connection must have a HCL reference to a plugin block
	// find the block
	if connection.PluginInstance != nil {
		p := c.PluginsInstances[*connection.PluginInstance]
		if p == nil {
			return nil, fmt.Errorf("connection '%s' specifies 'plugin=\"plugin.%s\"' but 'plugin.%s' does not exist. (%s:%d)",
				connection.Name,
				typehelpers.SafeString(connection.PluginInstance),
				typehelpers.SafeString(connection.PluginInstance),
				connection.DeclRange.Filename,
				connection.DeclRange.Start.Line,
			)
		}
		return p, nil
	}

	// resolve the image ref (this handles the special case of locally developed plugins in the plugins/local folder)
	imageRef := plugin.ResolvePluginImageRef(connection.PluginAlias)

	// verify the plugin is installed - if not return nil
	if _, ok := c.PluginVersions[imageRef]; !ok {
		// tactical - check if the plugin binary exists
		pluginBinaryPath := filepaths.PluginBinaryPath(imageRef, connection.PluginAlias)
		if _, err := os.Stat(pluginBinaryPath); err != nil {
			log.Printf("[INFO] plugin '%s' is not installed", imageRef)
			return nil, nil
		}

		// so the plugin binary exists but it does not exist in the versions.json
		// this is probably because it has been built locally - add a version entry with version set to 'local'
		c.PluginVersions[imageRef] = &versionfile.InstalledVersion{
			Version: "local",
		}
	}

	// how many plugin instances are there for this image ref?
	pluginsForImageRef := c.Plugins[imageRef]

	switch len(pluginsForImageRef) {
	case 0:
		// there is no plugin instance for this connection - add an implicit plugin instance
		p := plugin.NewImplicitPlugin(connection.PluginAlias, imageRef)

		// now add to our map
		if err := c.addPlugin(p); err != nil {
			// log the error but do not return it - we
			return nil, err
		}
		return p, nil

	case 1:
		// ok we can resolve
		return pluginsForImageRef[0], nil

	default:
		// so there is more than one plugin config for the plugin, and the connection DOES NOT specify which one to use
		// this is an error
		var strs = make([]string, len(pluginsForImageRef))
		for i, p := range pluginsForImageRef {
			strs[i] = fmt.Sprintf("\t%s (%s:%d)", p.Instance, *p.FileName, *p.StartLineNumber)
		}
		return nil, sperr.New("connection '%s' specifies 'plugin=\"%s\"' but the correct instance cannot be uniquely resolved. There are %d plugin instances matching that configuration:\n%s", connection.Name, connection.PluginAlias, len(pluginsForImageRef), strings.Join(strs, "\n"))
	}
}

// GetNonSearchPathConnections returns a list of connection names that are not in the provided search path
func (c *SteampipeConfig) GetNonSearchPathConnections(searchPath []string) []string {
	var res []string
	//convert searchPath to map for easy lookup
	searchPathLookup := helpers.SliceToLookup(searchPath)

	for connectionName := range c.Connections {
		if _, inSearchPath := searchPathLookup[connectionName]; !inSearchPath {
			res = append(res, connectionName)
		}
	}
	return res
}
