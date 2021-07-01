package modconfig

import (
	"fmt"
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
	// Child connections (only valid for "aggregate" type
	Connections []string
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
func (c *Connection) Validate() []string {
	var validationErrors []string

	validConnectionTypes := []string{"", ConnectionTypeAggregate}
	if !helpers.StringSliceContains(validConnectionTypes, c.Type) {
		return []string{fmt.Sprintf("connection '%s' has invalid connection type '%s'", c.Name, c.Type)}
	}
	if c.Type == ConnectionTypeAggregate {
		if len(c.Connections) == 0 {
			/// there should be at least one connection
			validationErrors = append(validationErrors, fmt.Sprintf("aggregate connection '%s' has no children", c.Name))
		} else {
			// check for duplicate entries
			if helpers.StringSliceHasDuplicates(c.Connections) {
				validationErrors = append(validationErrors, fmt.Sprintf("aggregate connection '%s' has duplicate children", c.Name))
			}
		}
	} else {
		// this is NOT an aggregate group - there should be no children
		if len(c.Connections) != 0 {
			validationErrors = append(validationErrors, fmt.Sprintf("connection '%s' has %d children, but is not of type 'aggregate'", c.Name, len(c.Connections)))
		}
	}
	return validationErrors
}
