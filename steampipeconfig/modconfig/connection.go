package modconfig

import (
	"fmt"
	"log"
	"path"
	"reflect"

	"github.com/turbot/go-kit/helpers"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/steampipe/steampipeconfig/options"
)

const (
	ConnectionTypeAggregator = "aggregator"
)

// Connection is a struct representing the partially parsed connection
//
// (Partial as the connection config, which is plugin specific, is stored as raw HCL.
// This will be parsed by the plugin)
type Connection struct {
	// connection name
	Name string
	// The name of plugin as mentioned in config
	PluginShortName string
	// The fully qualified name of the plugin. derived from the short name
	Plugin string
	// Type - supported values: "aggregator"
	Type string
	// this is a list of names or wildcards which are resolved to connections
	// (only valid for "aggregator" type)
	ConnectionNames []string
	// a list of the resolved child connections
	// (only valid for "aggregator" type)
	Connections map[string]*Connection
	// unparsed HCL of plugin specific connection config
	Config string

	// options
	Options   *options.Connection
	DeclRange hcl.Range
}

func NewConnection(block *hcl.Block) *Connection {
	return &Connection{
		Name:      block.Labels[0],
		DeclRange: block.TypeRange,
	}
}

// Equals
func (c *Connection) Equals(other *Connection) bool {
	connectionOptionsEqual := (c.Options == nil) == (other.Options == nil)
	if c.Options != nil {
		connectionOptionsEqual = c.Options.Equals(other.Options)
	}
	return c.Name == other.Name &&
		connectionOptionsEqual &&
		reflect.DeepEqual(c.Config, other.Config)
}

// SetOptions sets the options on the connection
// verify the options object is a valid options type (only options.Connection currently supported)
func (c *Connection) SetOptions(opts options.Options, block *hcl.Block) hcl.Diagnostics {
	var diags hcl.Diagnostics
	switch o := opts.(type) {
	case *options.Connection:
		c.Options = o
	default:
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("invalid nested option type %s - only 'connection' options blocks are supported for Connections", reflect.TypeOf(o).Name()),
			Subject:  &block.DefRange,
		})
	}
	return diags
}

func (c *Connection) String() string {
	return fmt.Sprintf("\n----\nName: %s\nPlugin: %s\nConfig:\n%s\nOptions:\n%s\n", c.Name, c.Plugin, c.Config, c.Options.String())
}

// Validate verifies the Type property is valid,
// if this is an aggregator connection, there must be at least one child, and no duplicates
// if this is NOT an aggregator, there must be no children
func (c *Connection) Validate(connectionMap map[string]*Connection) []string {
	validConnectionTypes := []string{"", ConnectionTypeAggregator}
	if !helpers.StringSliceContains(validConnectionTypes, c.Type) {
		return []string{fmt.Sprintf("connection '%s' has invalid connection type '%s'", c.Name, c.Type)}
	}
	if c.Type == ConnectionTypeAggregator {
		return c.ValidateAggregatorConnection(connectionMap)
	}
	// this is NOT an aggregator group - there should be no children
	var validationErrors []string

	if len(c.ConnectionNames) != 0 {
		validationErrors = append(validationErrors, fmt.Sprintf("connection '%s' has %d children, but is not of type 'aggregator'", c.Name, len(c.ConnectionNames)))
	}
	return validationErrors

}

func (c *Connection) ValidateAggregatorConnection(connectionMap map[string]*Connection) []string {
	if len(c.Connections) == 0 {
		/// there should be at least one connection
		return []string{fmt.Sprintf("aggregator connection '%s' has no children", c.Name)}
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
	return validationErrors
}

func (c *Connection) PopulateChildren(connectionMap map[string]*Connection) {
	log.Printf("[TRACE] Connection.PopulateChildren for aggregator connection %s", c.Name)
	c.Connections = make(map[string]*Connection)
	for _, childName := range c.ConnectionNames {
		// if this resolves as an existing connection, populate it
		if childConnection, ok := connectionMap[childName]; ok {
			log.Printf("[TRACE] Connection.PopulateChildren found matching connection %s", childName)
			c.Connections[childName] = childConnection
			continue
		}
		log.Printf("[TRACE] Connection.PopulateChildren no connection matches %s - treating as a wildcard", childName)
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
			if match, _ := path.Match(childName, name); match {
				// verify that this connection is of a compatible type
				if connection.Plugin == c.Plugin {
					c.Connections[name] = connection
					log.Printf("[TRACE] connection '%s' matches pattern '%s'", name, childName)
				}
			}
		}
	}
}
