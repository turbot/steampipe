package modconfig

import (
	"fmt"
	"github.com/turbot/pipe-fittings/plugin"
	"log"
	"path"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/turbot/steampipe/pkg/constants"
	"golang.org/x/exp/maps"
)

const (
	ConnectionTypePlugin     = "plugin"
	ConnectionTypeAggregator = "aggregator"
	ImportSchemaEnabled      = "enabled"
	ImportSchemaDisabled     = "disabled"
)

var ValidImportSchemaValues = []string{ImportSchemaEnabled, ImportSchemaDisabled}

// Connection is a struct representing the partially parsed connection
//
// (Partial as the connection config, which is plugin specific, is stored as raw HCL.
// This will be parsed by the plugin)
// json tags needed as this is stored in the connection state file
type Connection struct {
	// connection name
	Name string `json:"name"`
	// name of plugin as mentioned in config - this may be an alias to a plugin image ref
	// OR the label of a plugin config
	PluginAlias string `json:"plugin_short_name"`
	// image ref plugin.
	// we resolve this after loading all plugin configs
	Plugin string `json:"plugin"`
	// the label of the plugin config we are using
	PluginInstance *string `json:"plugin_instance"`
	// Path to the installed plugin (if it exists)
	PluginPath *string
	// connection type - supported values: "aggregator"
	Type string `json:"type,omitempty"`
	// should a schema be created for this connection - supported values: "enabled", "disabled"
	ImportSchema string `json:"import_schema"`
	// list of names or wildcards which are resolved to connections
	// (only valid for "aggregator" type)
	ConnectionNames []string `json:"connections,omitempty"`
	// a map of the resolved child connections
	// (only valid for "aggregator" type)
	Connections map[string]*Connection `json:"-"`
	// a list of the names resolved child connections
	// (only valid for "aggregator" type)
	ResolvedConnectionNames []string `json:"resolved_connections,omitempty"`
	// unparsed HCL of plugin specific connection config
	Config string `json:"config,omitempty"`

	Error error

	DeclRange plugin.Range `json:"decl_range"`
}

func (c *Connection) GetDeclRange() plugin.Range {
	return c.DeclRange
}

func (c *Connection) GetName() string {
	return c.Name
}

func (c *Connection) GetDisplayName() string {
	if c.ImportDisabled() {
		return fmt.Sprintf("%s (disabled)", c.Name)
	}
	return c.Name
}

func NewConnection(block *hcl.Block) *Connection {
	return &Connection{
		Name:         block.Labels[0],
		DeclRange:    plugin.NewRange(hclhelpers.BlockRange(block)),
		ImportSchema: ImportSchemaEnabled,
		// default to plugin
		Type: ConnectionTypePlugin,
	}
}

func (c *Connection) ImportDisabled() bool {
	return c.ImportSchema == constants.ConnectionStateDisabled
}

func (c *Connection) Equals(other *Connection) bool {
	return c.Name == other.Name &&
		c.Plugin == other.Plugin &&
		c.Type == other.Type &&
		strings.Join(c.ConnectionNames, ",") == strings.Join(other.ConnectionNames, ",") &&
		c.Config == other.Config &&
		c.ImportSchema == other.ImportSchema

}

func (c *Connection) String() string {
	return fmt.Sprintf("\n----\nName: %s\nPlugin: %s\nConfig:\n%s\n", c.Name, c.Plugin, c.Config)
}

// Validate verifies the Type property is valid,
// if this is an aggregator connection, there must be at least one child, and no duplicates
// if this is NOT an aggregator, there must be no children
func (c *Connection) Validate(map[string]*Connection) (warnings []string, errors []string) {
	validConnectionTypes := []string{ConnectionTypePlugin, ConnectionTypeAggregator}
	if !helpers.StringSliceContains(validConnectionTypes, c.Type) {
		return nil, []string{fmt.Sprintf("connection '%s' has invalid connection type '%s'", c.Name, c.Type)}
	}

	if c.Type == ConnectionTypeAggregator {
		return c.ValidateAggregatorConnection()
	}

	// this is NOT an aggregator group - there should be no children
	var validationErrors []string

	if len(c.ConnectionNames) != 0 {
		validationErrors = append(validationErrors, fmt.Sprintf("connection '%s' has %d children, but is not of type 'aggregator'", c.Name, len(c.ConnectionNames)))
	}
	validImportSchemaValues := utils.SliceToLookup(ValidImportSchemaValues)
	if _, isValid := validImportSchemaValues[c.ImportSchema]; !isValid {
		validationErrors = append(validationErrors, fmt.Sprintf("invalid value '%s'for import_schema, must be one of ['%s']", c.ImportSchema, strings.Join(ValidImportSchemaValues, "','")))
	}

	return nil, validationErrors

}

func (c *Connection) ValidateAggregatorConnection() (warnings, errors []string) {
	if len(c.Connections) == 0 {
		/// there should be at least one connection - raise as warning
		return []string{c.GetEmptyAggregatorError()}, nil
	}

	var validationErrors []string

	// now ensure all child connections are loaded and use the same plugin as the parent connection
	for _, childConnection := range c.Connections {
		if childConnection.Plugin != c.Plugin {
			validationErrors = append(validationErrors,
				fmt.Sprintf("aggregator connection '%s' uses plugin %s but child connection '%s' uses plugin '%s'",
					c.Name,
					c.Plugin,
					childConnection.Name,
					childConnection.Plugin,
				))
		}

	}
	return nil, validationErrors
}

func (c *Connection) GetEmptyAggregatorError() string {
	patterns := c.ConnectionNames
	if len(patterns) == 0 {
		return fmt.Sprintf("aggregator '%s' defines no child connections", c.Name)
	}
	if len(patterns) == 1 {
		return fmt.Sprintf("aggregator '%s' with pattern '%s' matches no connections",
			c.Name,
			patterns[0])
	}
	return fmt.Sprintf("aggregator '%s' with patterns ['%s'] matches no connections",
		c.Name,
		strings.Join(patterns, "','"))
}

func (c *Connection) PopulateChildren(connectionMap map[string]*Connection) []string {
	log.Printf("[TRACE] Connection.PopulateChildren for aggregator connection %s", c.Name)
	c.Connections = make(map[string]*Connection)
	var failures []string
	for _, childPattern := range c.ConnectionNames {
		// if this resolves as an existing connection, populate it
		if childConnection, ok := connectionMap[childPattern]; ok {
			// verify this child connection has the same plugin instance
			if childConnection.PluginInstance != c.PluginInstance {
				msg := fmt.Sprintf("aggregator connection %s specifies child connection %s but it has a different plugin instance",
					c.Name, childPattern)
				log.Println("[WARN]", msg)
				failures = append(failures, msg)
			} else {
				log.Printf("[TRACE] Connection.PopulateChildren found matching connection %s", childPattern)
				c.Connections[childPattern] = childConnection
			}
			continue
		}

		log.Printf("[TRACE] Connection.PopulateChildren no connection matches %s - treating as a wildcard", childPattern)
		// otherwise treat the connection name as a wildcard and see what matches
		for name, connection := range connectionMap {
			// if this is an aggregator connection, skip (this will also avoid us adding ourselves)
			if connection.Type == ConnectionTypeAggregator {
				continue
			}
			// have we already added this connection
			if _, ok := c.Connections[name]; ok {
				continue
			}
			if match, _ := path.Match(childPattern, name); match {
				// verify that this connection is the same plugin instance
				if connection.PluginInstance == c.PluginInstance {
					c.Connections[name] = connection
					log.Printf("[TRACE] connection '%s' matches pattern '%s'", name, childPattern)
				}
			}
		}
	}
	c.ResolvedConnectionNames = maps.Keys(c.Connections)
	return failures
}

// GetResolveConnectionNames return the names of all child connections
// (will only be non-empty for aggregator connections)
func (c *Connection) GetResolveConnectionNames() []string {
	res := make([]string, len(c.Connections))
	idx := 0
	for k := range c.Connections {
		res[idx] = k
		idx++
	}
	return res
}
