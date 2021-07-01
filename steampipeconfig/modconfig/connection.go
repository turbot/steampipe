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
	ConnectionTypeAggregate = "aggregate"
)

// Connection is a struct representing the partially parsed connection
//
// (Partial as the connection config, which is plugin specific, is stored as raw HCL.
// This will be parsed by the plugin)
type Connection struct {
	// connection name
	Name string
	// Name of plugin
	Plugin string
	// Type - supported values: "aggregate"
	Type string
	// this is a list of names or wildcards which are resolved to connections
	// (only valid for "aggregate" type)
	ConnectionNames []string
	// a list of the resolved child connections
	// (only valid for "aggregate" type)
	Connections map[string]*Connection
	// unparsed HCL of plugin specific connection config
	Config string

	// options
	Options *options.Connection
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
// if this is an aggregate connection, there must be at least one child, and no duplicates
// if this is NOT an aggregate, there must be no children
func (c *Connection) Validate(connectionMap map[string]*Connection) []string {
	validConnectionTypes := []string{"", ConnectionTypeAggregate}
	if !helpers.StringSliceContains(validConnectionTypes, c.Type) {
		return []string{fmt.Sprintf("connection '%s' has invalid connection type '%s'", c.Name, c.Type)}
	}
	if c.Type == ConnectionTypeAggregate {
		return c.ValidateAggregateConnection(connectionMap)
	}
	// this is NOT an aggregate group - there should be no children
	var validationErrors []string

	if len(c.ConnectionNames) != 0 {
		validationErrors = append(validationErrors, fmt.Sprintf("connection '%s' has %d children, but is not of type 'aggregate'", c.Name, len(c.ConnectionNames)))
	}
	return validationErrors

}

func (c *Connection) ValidateAggregateConnection(connectionMap map[string]*Connection) []string {
	if len(c.Connections) == 0 {
		/// there should be at least one connection
		return []string{fmt.Sprintf("aggregate connection '%s' has no children", c.Name)}
	}

	var validationErrors []string

	// now ensure all child connections are loaded and use the same plugin as the parent connection
	for _, childConnection := range c.Connections {
		if childConnection.Plugin != c.Plugin {
			validationErrors = append(validationErrors,
				fmt.Sprintf("aggregate connection '%s' uses plugin %s but child connection '%s' uses plugin '%s'",
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
	log.Printf("[TRACE] Connection.PopulateChildren for aggregate connection %s", c.Name)
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
			// have we already added this connection
			if _, ok := c.Connections[name]; ok {
				continue
			}
			if match, _ := path.Match(childName, name); match {
				c.Connections[name] = connection
				log.Printf("[TRACE] Connection.PopulateChildren connection %s matches pattern %s", name, childName)
			}
		}
	}

	log.Printf("[TRACE] Connection.PopulateChildren complete: \n%v", c.Connections)

}
